// Package telemetry declares the closed product-event vocabulary (D-20),
// its per-event metadata key allowlists (D-21), and every Prometheus
// collector family ROADMAP Phase 1 criterion 4 names. It must not import
// database/sql or anything under github.com/openmtg/edh-go/server, so that
// `go test ./pkg/... -race` — the only target CI runs — gates it on every
// pull request.
package telemetry

import "strings"

// MaxMetadataValueBytes bounds the size, in raw UTF-8 bytes, of any single
// metadata value accepted by ValidateEvent. Counted with len over the
// value's bytes — no case folding, no Unicode normalization.
const MaxMetadataValueBytes = 128

// EventSpec describes one entry in the closed product-event vocabulary.
type EventSpec struct {
	// Authoritative marks one of the six server-owned events. A client
	// attempting to submit one of these through TrackProductEvent is
	// rejected with RejectionClientAuthoritative.
	Authoritative bool
	// Keys is this event's own metadata key allowlist. Per D-21 this set is
	// declared per event, never shared as a global union across events, so
	// a key permitted on one event is rejected on another that has no
	// business carrying it.
	Keys map[string]struct{}
}

func keys(names ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		out[n] = struct{}{}
	}
	return out
}

// Vocabulary is the closed, exactly-15-event product-event vocabulary,
// transcribed from the "Product Event Vocabulary" table in
// docs/product/2026-07-23-deck-to-game-activation-prd.md. Fields the PRD
// table names that already have a dedicated product_events column
// (session ID, source, role, game ID, duration) are carried on that column,
// not duplicated into metadata; only fields with no dedicated column are
// allowlisted metadata keys here. See
// docs/analytics/product-event-vocabulary.md for the full per-key
// provenance.
//
// "campaign source" is rendered as the four allowlisted attribution keys
// utm_source, utm_medium, utm_campaign, referrer_host, per this plan's task
// action.
var Vocabulary = map[string]EventSpec{
	// --- 9 client events ---
	"landing_primary_cta": {Keys: keys("utm_source", "utm_medium", "utm_campaign", "referrer_host")},
	"quick_start_viewed":  {Keys: keys()},
	"deck_import_started": {Keys: keys()}, // source type carried on the source column
	"game_create_started": {Keys: keys()},
	// share_method has no PRD row of its own: REQ-A4 requires recording
	// which share mechanism succeeded (native share vs. clipboard fallback)
	// on invite_copied. Provenance: REQ-A4.
	"invite_copied":         {Keys: keys("share_method")},
	"invite_viewed":         {Keys: keys("utm_source", "utm_medium", "utm_campaign", "referrer_host")},
	"join_started":          {Keys: keys()},
	"board_ready":           {Keys: keys()}, // elapsed ms carried on the duration_ms column
	"account_claim_started": {Keys: keys()},

	// --- 6 server-authoritative events (Authoritative: true) ---
	"deck_import_succeeded": {Authoritative: true, Keys: keys("card_count", "unresolved_count")},
	"deck_import_failed":    {Authoritative: true, Keys: keys("reason")},
	"guest_session_created": {Authoritative: true, Keys: keys()},
	"game_created":          {Authoritative: true, Keys: keys()},
	"player_joined":         {Authoritative: true, Keys: keys()},
	"account_claimed":       {Authoritative: true, Keys: keys()},
}

// EventNames returns every event name in Vocabulary.
func EventNames() []string {
	names := make([]string, 0, len(Vocabulary))
	for name := range Vocabulary {
		names = append(names, name)
	}
	return names
}

// Rejection is a bounded, closed set of reasons ValidateEvent may refuse an
// event. Its String() form is used verbatim as the Prometheus `reason`
// label value on vedh_product_events_dropped_total — bounded specifically
// so an attacker-chosen event name or key can never become a label.
type Rejection int

const (
	// RejectionNone means the event was accepted.
	RejectionNone Rejection = iota
	// RejectionUnknownEvent means the event name is not in Vocabulary.
	RejectionUnknownEvent
	// RejectionUnknownKey means a metadata key is not in this event's own
	// allowlist.
	RejectionUnknownKey
	// RejectionOversizedValue means a metadata value exceeds
	// MaxMetadataValueBytes.
	RejectionOversizedValue
	// RejectionClientAuthoritative means a client attempted to submit one
	// of the six server-owned events through the client-facing mutation.
	RejectionClientAuthoritative
	// RejectionMissingSession means the session ID was empty or
	// whitespace-only. product_events.session_id is NOT NULL and a
	// whitespace-only string satisfies GraphQL's String! while satisfying
	// nothing else: an event with no session cannot join any funnel.
	RejectionMissingSession
	// RejectionWriteError means the event passed validation but the insert
	// into product_events failed. ValidateEvent never returns this value —
	// it is used by server/product_events.go's write path so that
	// "write_error" is a member of this same bounded reason set rather
	// than a second, unbounded string literal reaching the Prometheus
	// `reason` label.
	RejectionWriteError
)

// String returns the bounded reason string used as the Prometheus `reason`
// label value.
func (r Rejection) String() string {
	switch r {
	case RejectionNone:
		return "none"
	case RejectionUnknownEvent:
		return "unknown_event"
	case RejectionUnknownKey:
		return "unknown_key"
	case RejectionOversizedValue:
		return "oversized_value"
	case RejectionClientAuthoritative:
		return "client_authoritative"
	case RejectionMissingSession:
		return "missing_session"
	case RejectionWriteError:
		return "write_error"
	default:
		return "unknown_rejection"
	}
}

// ValidateEvent decides whether an event may be written. Comparison is raw
// UTF-8 byte equality on event names and metadata keys — no case folding,
// no Unicode normalization. Session-ID emptiness is the one exception: it
// is a whitespace trim of the check, not a normalization of the retained
// value.
//
// The sessionID and clientSubmitted checks are here, in the CI-gated
// pkg/telemetry package, rather than re-implemented in server/, so the
// rule is enforced on every pull request rather than only when a live
// database is available to run server/ tests.
func ValidateEvent(name string, sessionID string, metadata map[string]string, clientSubmitted bool) Rejection {
	if strings.TrimSpace(sessionID) == "" {
		return RejectionMissingSession
	}

	spec, ok := Vocabulary[name]
	if !ok {
		return RejectionUnknownEvent
	}

	if clientSubmitted && spec.Authoritative {
		return RejectionClientAuthoritative
	}

	for k, v := range metadata {
		if _, allowed := spec.Keys[k]; !allowed {
			return RejectionUnknownKey
		}
		if len(v) > MaxMetadataValueBytes {
			return RejectionOversizedValue
		}
	}

	return RejectionNone
}

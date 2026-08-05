package telemetry

import (
	"sort"
	"strings"
	"testing"
)

// TestValidateEvent_ConstructionSmoke is a minimal smoke test written by
// this plan's task 1. Task 3 owns the full TestValidateEvent_Rejections
// table and the closed-vocabulary / forbidden-key structural tests; this
// test only proves ValidateEvent accepts a known-good event with a
// non-blank session ID before those land.
func TestValidateEvent_ConstructionSmoke(t *testing.T) {
	got := ValidateEvent("quick_start_viewed", "session-1", nil, false)
	if got != RejectionNone {
		t.Fatalf("ValidateEvent() = %v, want RejectionNone", got)
	}
}

// d20EventNames is the closed, exactly-15-event set D-20 names, transcribed
// verbatim from .planning/phases/01-measured-deck-import-foundation's
// 01-CONTEXT.md (and repeated in this plan's "New event names" artifact
// list). TestVocabulary_IsClosedAtFifteen fails the instant a 16th event is
// added or one of these names is renamed.
var d20EventNames = []string{
	"landing_primary_cta",
	"quick_start_viewed",
	"deck_import_started",
	"game_create_started",
	"invite_copied",
	"invite_viewed",
	"join_started",
	"board_ready",
	"account_claim_started",
	"deck_import_succeeded",
	"deck_import_failed",
	"guest_session_created",
	"game_created",
	"player_joined",
	"account_claimed",
}

// TestVocabulary_IsClosedAtFifteen asserts len(Vocabulary) == 15 and that
// the sorted key set equals the D-20 list literal-for-literal.
func TestVocabulary_IsClosedAtFifteen(t *testing.T) {
	if len(Vocabulary) != 15 {
		t.Fatalf("len(Vocabulary) = %d, want 15", len(Vocabulary))
	}

	got := EventNames()
	sort.Strings(got)
	want := append([]string(nil), d20EventNames...)
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("EventNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("EventNames()[%d] = %q, want %q (sorted sets differ: got %v, want %v)", i, got[i], want[i], got, want)
		}
	}
}

// forbiddenKeySubstrings is the structural form of the privacy constraint:
// no allowlisted metadata key may name a decklist, a URL, a credential, a
// token, a clipboard value, or an IP address, even indirectly. Substring
// matching rather than exact names is deliberate — it catches a future
// deck_source_url as well as a bare url.
var forbiddenKeySubstrings = []string{
	"deck", "list", "url", "uri", "link", "password",
	"passwd", "secret", "token", "jwt", "auth", "clipboard", "ip", "addr", "cards", "hand",
}

// TestVocabulary_NoForbiddenKeys iterates every event and every allowlisted
// key and fails when a lowercased key contains any forbidden substring.
// This goes red the moment a future contributor adds a leaking key.
func TestVocabulary_NoForbiddenKeys(t *testing.T) {
	for event, spec := range Vocabulary {
		for key := range spec.Keys {
			lowered := strings.ToLower(key)
			for _, bad := range forbiddenKeySubstrings {
				if strings.Contains(lowered, bad) {
					t.Errorf("event %q allowlists key %q, which contains forbidden substring %q", event, key, bad)
				}
			}
		}
	}
}

// authoritativeEventNames is exactly the six events the PRD's "Product
// Event Vocabulary" table labels Server.
var authoritativeEventNames = map[string]struct{}{
	"deck_import_succeeded": {},
	"deck_import_failed":    {},
	"guest_session_created": {},
	"game_created":          {},
	"player_joined":         {},
	"account_claimed":       {},
}

// TestVocabulary_AuthoritativeSetMatchesSpec asserts that exactly the six
// events the PRD labels Server carry Authoritative: true.
func TestVocabulary_AuthoritativeSetMatchesSpec(t *testing.T) {
	gotCount := 0
	for name, spec := range Vocabulary {
		_, wantAuthoritative := authoritativeEventNames[name]
		if spec.Authoritative != wantAuthoritative {
			t.Errorf("Vocabulary[%q].Authoritative = %v, want %v", name, spec.Authoritative, wantAuthoritative)
		}
		if spec.Authoritative {
			gotCount++
		}
	}
	if gotCount != len(authoritativeEventNames) {
		t.Fatalf("authoritative event count = %d, want %d", gotCount, len(authoritativeEventNames))
	}
}

// TestValidateEvent_Rejections is a table test covering every case
// ValidateEvent must distinguish, asserting the returned Rejection by
// value — not merely "not None" — so a refusal cannot pass under the wrong
// reason and then arrive at Prometheus mislabelled.
func TestValidateEvent_Rejections(t *testing.T) {
	maxValue := strings.Repeat("v", MaxMetadataValueBytes)
	overValue := strings.Repeat("v", MaxMetadataValueBytes+1)

	cases := []struct {
		name            string
		eventName       string
		sessionID       string
		metadata        map[string]string
		clientSubmitted bool
		want            Rejection
	}{
		{
			name:      "unknown event name",
			eventName: "not_a_real_event",
			sessionID: "session-1",
			want:      RejectionUnknownEvent,
		},
		{
			// Proves D-21's per-event allowlist is not silently collapsing
			// into a global union: share_method is allowlisted on
			// invite_copied but not on quick_start_viewed.
			name:      "key absent from this event's own allowlist but present on another event's",
			eventName: "quick_start_viewed",
			sessionID: "session-1",
			metadata:  map[string]string{"share_method": "clipboard"},
			want:      RejectionUnknownKey,
		},
		{
			name:      "value one byte over the limit",
			eventName: "landing_primary_cta",
			sessionID: "session-1",
			metadata:  map[string]string{"utm_source": overValue},
			want:      RejectionOversizedValue,
		},
		{
			name:      "value exactly at the limit is accepted",
			eventName: "landing_primary_cta",
			sessionID: "session-1",
			metadata:  map[string]string{"utm_source": maxValue},
			want:      RejectionNone,
		},
		{
			name:            "client submission of a server-owned event",
			eventName:       "game_created",
			sessionID:       "session-1",
			clientSubmitted: true,
			want:            RejectionClientAuthoritative,
		},
		{
			name:      "event name differing only in letter case is unknown",
			eventName: "Game_Created",
			sessionID: "session-1",
			want:      RejectionUnknownEvent,
		},
		{
			name:      "empty metadata map is accepted",
			eventName: "quick_start_viewed",
			sessionID: "session-1",
			metadata:  map[string]string{},
			want:      RejectionNone,
		},
		{
			name:      "one allowlisted key is accepted",
			eventName: "landing_primary_cta",
			sessionID: "session-1",
			metadata:  map[string]string{"utm_source": "google"},
			want:      RejectionNone,
		},
		{
			name:      "three allowlisted keys on the same event are accepted",
			eventName: "landing_primary_cta",
			sessionID: "session-1",
			metadata: map[string]string{
				"utm_source":    "google",
				"utm_medium":    "cpc",
				"utm_campaign":  "spring",
				"referrer_host": "google.com",
			},
			want: RejectionNone,
		},
		{
			name:      "empty session ID",
			eventName: "quick_start_viewed",
			sessionID: "",
			want:      RejectionMissingSession,
		},
		{
			name:      "whitespace-only session ID",
			eventName: "quick_start_viewed",
			sessionID: "   ",
			want:      RejectionMissingSession,
		},
		{
			name:      "single non-space character session ID is accepted",
			eventName: "quick_start_viewed",
			sessionID: "x",
			want:      RejectionNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateEvent(tc.eventName, tc.sessionID, tc.metadata, tc.clientSubmitted)
			if got != tc.want {
				t.Fatalf("ValidateEvent(%q, %q, %v, %v) = %v, want %v",
					tc.eventName, tc.sessionID, tc.metadata, tc.clientSubmitted, got, tc.want)
			}
		})
	}
}

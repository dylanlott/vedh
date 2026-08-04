// Package deckimport provides a deterministic, dependency-free grammar for
// turning a pasted or fetched decklist into a normalized set of entries. It
// must never import database/sql or any github.com/openmtg/edh-go/server
// path, so that `go test ./pkg/... -race` — the only target CI runs — gates
// it on every pull request.
package deckimport

// SourceType identifies where a deck's text originated. It is a named type,
// deliberately unlike the untyped string constants in
// server/gamelog_helpers.go, because a named type is what makes the
// WithLabelValues compile-time guarantee reviewable: a caller cannot pass an
// arbitrary string to a Prometheus label expecting a SourceType.
//
// The set is complete at four values here (D-24) even though only
// SourcePlainText is produced by this plan's detector: these values land in
// Prometheus labels and historical product_events rows, both of which are
// costly to rename, so the enum is fixed once rather than grown per source.
type SourceType string

const (
	// SourceMoxfield identifies a deck fetched from or shaped like a
	// Moxfield export. Detection lands in a later plan in this phase.
	SourceMoxfield SourceType = "moxfield"
	// SourceArchidekt identifies a deck fetched from or shaped like an
	// Archidekt export. Detection lands in a later plan in this phase.
	SourceArchidekt SourceType = "archidekt"
	// SourcePlainText identifies a plain pasted decklist with no
	// provider-specific shape. This is the only value DetectPasteSource
	// returns in this plan.
	SourcePlainText SourceType = "plain_text"
	// SourceUnknown identifies text whose origin cannot be determined.
	SourceUnknown SourceType = "unknown"
)

// AllSourceTypes returns every declared SourceType value. It exists so tests
// and the telemetry label-allowlist walk can assert a Prometheus `source`
// label value is always a member of the enum, never an arbitrary string.
func AllSourceTypes() []SourceType {
	return []SourceType{SourceMoxfield, SourceArchidekt, SourcePlainText, SourceUnknown}
}

// DetectPasteSource returns the SourceType of pasted text. In this plan it
// returns SourcePlainText for any non-empty input and SourceUnknown for
// empty input; provider-shape heuristics (Moxfield/Archidekt export
// detection) are added in plan 01-02.
func DetectPasteSource(text string) SourceType {
	if text == "" {
		return SourceUnknown
	}
	return SourcePlainText
}

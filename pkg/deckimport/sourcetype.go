// Package deckimport provides a deterministic, dependency-free grammar for
// turning a pasted or fetched decklist into a normalized set of entries. It
// must never import database/sql or any github.com/openmtg/edh-go/server
// path, so that `go test ./pkg/... -race` — the only target CI runs — gates
// it on every pull request.
package deckimport

import "strings"

// SourceType identifies where a deck's text originated. It is a named type,
// deliberately unlike the untyped string constants in
// server/gamelog_helpers.go, because a named type is what makes the
// WithLabelValues compile-time guarantee reviewable: a caller cannot pass an
// arbitrary string to a Prometheus label expecting a SourceType.
//
// The set is complete at four values (D-24): these values land in
// Prometheus labels and historical product_events rows, both of which are
// costly to rename, so the enum is fixed once rather than grown per source.
type SourceType string

const (
	// SourceMoxfield identifies a deck fetched from or shaped like a
	// Moxfield export.
	SourceMoxfield SourceType = "moxfield"
	// SourceArchidekt identifies a deck fetched from or shaped like an
	// Archidekt export.
	SourceArchidekt SourceType = "archidekt"
	// SourcePlainText identifies a plain pasted decklist with no
	// provider-specific shape.
	SourcePlainText SourceType = "plain_text"
	// SourceUnknown identifies text whose origin cannot be determined,
	// including text carrying markers from more than one provider.
	SourceUnknown SourceType = "unknown"
)

// AllSourceTypes returns every declared SourceType value. It exists so tests
// and the telemetry label-allowlist walk can assert a Prometheus `source`
// label value is always a member of the enum, never an arbitrary string.
func AllSourceTypes() []SourceType {
	return []SourceType{SourceMoxfield, SourceArchidekt, SourcePlainText, SourceUnknown}
}

// DetectPasteSource returns the SourceType of pasted text, using heuristics
// over line features only. This is D-23/D-24's paste-shape detector:
// telemetry only. Its result feeds nothing but ParsedDeck.Source — per the
// anti-pattern this phase's research called out by name, detection must
// never influence what a line parses to, and ParseWithSource exists
// specifically so a test can force every enum value here and prove no other
// field of the result ever changes.
//
// Detection runs over an independent read of the raw text using the same
// line-splitting and header/comment recognition Parse uses, but computes no
// ParsedEntry and consults no ParsedDeck: the two computations are
// structurally separate code paths that happen to agree on what a "line"
// and a "header" are.
func DetectPasteSource(text string) SourceType {
	if strings.TrimSpace(text) == "" {
		return SourceUnknown
	}

	var entryLines, moxfieldLines, archidektLines int

	for _, raw := range splitLines(text) {
		line := strings.TrimSpace(raw)
		if line == "" || isCommentLine(line) {
			continue
		}

		if matchSectionHeader(line) != headerNone {
			// Archidekt marker: a section header carrying a
			// parenthesised count, e.g. "Sideboard (12)" — a line
			// feature Moxfield's plain-text export does not use.
			if hasParenthesizedCount(line) {
				archidektLines++
			}
			continue
		}

		entryLines++

		// Moxfield marker: an uppercase "(SET) NUMBER" printing suffix,
		// Moxfield's plain-text export convention.
		if hasUppercaseSetAndCollector(line) {
			moxfieldLines++
		}
		// Archidekt marker: a trailing "[Category]" or a trailing
		// backtick-delimited "`Category`" — Archidekt's export
		// convention for tags/categories.
		if hasBracketOrBacktickCategory(line) {
			archidektLines++
		}
	}

	switch {
	case entryLines == 0:
		// No line was recognizable as an entry at all.
		return SourceUnknown
	case moxfieldLines > 0 && archidektLines > 0:
		// Markers from more than one provider: the funnel comparison
		// this enum exists for would rather have an honest "unknown"
		// than a guess.
		return SourceUnknown
	case moxfieldLines*2 > entryLines:
		// A majority of entry lines carry the Moxfield marker.
		return SourceMoxfield
	case archidektLines > 0:
		return SourceArchidekt
	default:
		return SourcePlainText
	}
}

// hasParenthesizedCount reports whether line, once trimmed, ends in a
// "(<digits>)" group — the count-suffix line feature a dropped-section
// header carries in Archidekt exports (e.g. "Sideboard (12)").
func hasParenthesizedCount(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasSuffix(trimmed, ")") {
		return false
	}
	idx := strings.LastIndexByte(trimmed, '(')
	if idx == -1 {
		return false
	}
	return isAllDigits(trimmed[idx+1 : len(trimmed)-1])
}

// hasUppercaseSetAndCollector reports whether line ends in a trailing
// "(SET) NUMBER" printing suffix whose SET is written in uppercase — the
// line feature Moxfield's plain-text export uses for every entry, and
// Archidekt conventionally does not (Archidekt's own convention lower-cases
// or leaves the set code as typed).
func hasUppercaseSetAndCollector(line string) bool {
	trimmed := strings.TrimRight(line, " \t")

	lastSpace := strings.LastIndexAny(trimmed, " \t")
	var token, before string
	if lastSpace == -1 {
		token, before = trimmed, ""
	} else {
		token = trimmed[lastSpace+1:]
		before = strings.TrimRight(trimmed[:lastSpace], " \t")
	}
	if !isAlnumToken(token) || !strings.HasSuffix(before, ")") {
		return false
	}

	idx := strings.LastIndexByte(before, '(')
	if idx == -1 {
		return false
	}
	setCode := before[idx+1 : len(before)-1]
	return setCode != "" && isUpperAlnum(setCode)
}

// hasBracketOrBacktickCategory reports whether line ends in a trailing
// "[Category]" or a trailing backtick-delimited "`Category`" — Archidekt's
// export convention for a card's tag/category.
func hasBracketOrBacktickCategory(line string) bool {
	trimmed := strings.TrimRight(line, " \t")

	if strings.HasSuffix(trimmed, "]") {
		idx := strings.LastIndexByte(trimmed, '[')
		return idx != -1 && trimmed[idx+1:len(trimmed)-1] != ""
	}
	if strings.HasSuffix(trimmed, "`") {
		rest := trimmed[:len(trimmed)-1]
		idx := strings.LastIndexByte(rest, '`')
		return idx != -1 && rest[idx+1:] != ""
	}
	return false
}

// isUpperAlnum reports whether s is non-empty, alphanumeric, and contains
// no lowercase ASCII letters (digits-only strings, like a numeric-only set
// code, also count as "uppercase" here since they contain no lowercase).
func isUpperAlnum(s string) bool {
	if !isAlnumToken(s) {
		return false
	}
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return false
		}
	}
	return true
}

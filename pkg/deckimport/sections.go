package deckimport

import (
	"fmt"
	"strings"
)

// sectionHeaderKind identifies which of the five recognized header words a
// line matched, or headerNone when it matched none of them.
type sectionHeaderKind int

const (
	headerNone sectionHeaderKind = iota
	headerCommander
	headerSideboard
	headerMaybeboard
	headerDeck
	headerCompanion
)

// headerWords maps each recognized header word, lower-cased, to its kind.
// Matching is case-insensitive on the header word only.
var headerWords = map[string]sectionHeaderKind{
	"commander":  headerCommander,
	"sideboard":  headerSideboard,
	"maybeboard": headerMaybeboard,
	"deck":       headerDeck,
	"companion":  headerCompanion,
}

// matchSectionHeader reports whether line, once an optional trailing
// "(<count>)" is stripped, is exactly one of the five header words
// (case-insensitive on the word, anchored to the whole line). Requiring the
// header word to constitute the *whole* line — apart from that optional
// count suffix — is what keeps "1 Commander's Sphere" an entry: it has
// content beyond "Commander" plus an optional count, so it never matches.
// isAllDigits (scanner.go) is reused here rather than redefined.
func matchSectionHeader(line string) sectionHeaderKind {
	trimmed := strings.TrimSpace(line)

	if idx := strings.LastIndexByte(trimmed, '('); idx != -1 && strings.HasSuffix(trimmed, ")") {
		inner := trimmed[idx+1 : len(trimmed)-1]
		if isAllDigits(inner) {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
	}

	return headerWords[strings.ToLower(trimmed)]
}

// sectionTracker holds the running section state a document's header lines
// establish, plus the bookkeeping D-11 requires for a dropped
// Sideboard/Maybeboard section. It lives for exactly one Parse call and
// carries no package-level state, so it introduces no concurrency hazard.
type sectionTracker struct {
	current  Section
	dropping bool

	// The currently-open dropping section's header row, held so its one
	// summary warning can be attributed to that row when the section
	// closes — which is also how that header row is accounted for in the
	// "no nonblank row disappears" invariant.
	dropHeaderLine int
	dropHeaderRaw  string
	dropCount      int
}

// noteHeader records a header line: it first closes out any dropping
// section already open (emitting that section's summary warning), then
// opens whatever the new header calls for.
func (t *sectionTracker) noteHeader(deck *ParsedDeck, kind sectionHeaderKind, sourceLine int, raw string) {
	t.closeDropSection(deck)

	switch kind {
	case headerCommander:
		t.current = SectionCommander
	case headerSideboard:
		t.current = SectionSideboard
		t.dropping = true
		t.dropHeaderLine = sourceLine
		t.dropHeaderRaw = raw
		t.dropCount = 0
	case headerMaybeboard:
		t.current = SectionMaybeboard
		t.dropping = true
		t.dropHeaderLine = sourceLine
		t.dropHeaderRaw = raw
		t.dropCount = 0
	case headerDeck, headerCompanion:
		t.current = SectionMain
	}
}

// closeDropSection, called when a new header appears or at end of input,
// appends the one summary warning D-11 requires for a just-finished
// Sideboard/Maybeboard section, citing that section's header row. This is
// deliberately how the header row itself is accounted for: the row becomes
// the summary Warning, and each entry beneath it became a DroppedSectionRow
// as it was scanned — together they cover every nonblank line the section
// contained.
func (t *sectionTracker) closeDropSection(deck *ParsedDeck) {
	if !t.dropping {
		return
	}

	word := "sideboard"
	if t.current == SectionMaybeboard {
		word = "maybeboard"
	}

	deck.NonBlankLines++
	deck.Warnings = append(deck.Warnings, Warning{
		SourceLine: t.dropHeaderLine,
		RawLine:    t.dropHeaderRaw,
		Message:    fmt.Sprintf("%d %s card(s) ignored.", t.dropCount, word),
	})

	t.dropping = false
}

// dropRow accounts for one row that fell inside an open Sideboard/Maybeboard
// section: it counts toward NonBlankLines and DroppedSectionRows, but never
// becomes an Entry.
func (t *sectionTracker) dropRow(deck *ParsedDeck) {
	deck.NonBlankLines++
	deck.DroppedSectionRows++
	t.dropCount++
}

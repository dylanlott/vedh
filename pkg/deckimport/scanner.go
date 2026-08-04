package deckimport

import (
	"strings"
)

// MaxDecklistBytes bounds the size of text accepted by Parse. Enforced
// before scanning so a caller cannot force unbounded memory use through a
// public, unauthenticated mutation.
const MaxDecklistBytes = 256 * 1024

// MaxDecklistLines bounds the number of lines accepted by Parse, enforced
// for the same reason as MaxDecklistBytes.
const MaxDecklistLines = 2000

// Parse turns pasted decklist text into a ParsedDeck. This plan's grammar is
// deliberately minimal — the <digits><single space><name> shape plus
// blank-line skipping — because this task proves the whole architecture
// end to end, not the full parser grammar; the complete six-syntax grammar
// (1x, comma-separated, quoted names, printing metadata, sections) is added
// across later plans in this phase.
//
// Parse never fails the whole deck for a malformed line: a line whose
// quantity is missing or non-positive becomes a Warning, never a
// BlockingError. The two size limits above are the only conditions that
// produce a BlockingError, because reading unbounded input is the one
// failure mode a per-line warning cannot bound.
//
// Parse logs nothing. Warnings and blocking errors are return values.
func Parse(text string) ParsedDeck {
	deck := ParsedDeck{
		Source: DetectPasteSource(text),
	}

	if len(text) > MaxDecklistBytes {
		deck.BlockingErrors = append(deck.BlockingErrors, BlockingError{
			Message: "Decklist is too large to import.",
		})
		return deck
	}

	lines := strings.Split(text, "\n")
	if len(lines) > MaxDecklistLines {
		deck.BlockingErrors = append(deck.BlockingErrors, BlockingError{
			Message: "Decklist has too many lines to import.",
		})
		return deck
	}

	for i, raw := range lines {
		sourceLine := i + 1

		// Trim a trailing carriage return (Windows pastes) only; leading
		// whitespace is meaningful to the blank-line check below but not
		// otherwise trimmed here, matching the minimal grammar this task
		// implements.
		line := strings.TrimSuffix(raw, "\r")

		if strings.TrimSpace(line) == "" {
			// Blank lines are the one thing allowed to vanish silently.
			continue
		}

		deck.NonBlankLines++

		qty, name, ok := parseQuantityAndName(line)
		if !ok {
			deck.Warnings = append(deck.Warnings, Warning{
				SourceLine: sourceLine,
				RawLine:    raw,
				Message:    "Could not read a quantity for this line; it was skipped.",
			})
			continue
		}

		deck.Entries = append(deck.Entries, ParsedEntry{
			Quantity:   qty,
			Name:       name,
			Section:    SectionMain,
			SourceLine: sourceLine,
			RawLine:    raw,
		})
	}

	return deck
}

// parseQuantityAndName implements the <digits><single space><name> shape:
// consume leading digits, consume exactly one run of spaces, and take the
// whole remainder as the name with internal commas preserved verbatim. It
// reports ok=false when no leading digits are present or the resulting
// quantity is non-positive.
func parseQuantityAndName(line string) (quantity int, name string, ok bool) {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, "", false
	}

	digits := line[:i]
	rest := line[i:]

	spaceCount := 0
	for spaceCount < len(rest) && rest[spaceCount] == ' ' {
		spaceCount++
	}
	if spaceCount == 0 {
		return 0, "", false
	}
	remainder := strings.TrimLeft(rest[spaceCount:], " ")
	if remainder == "" {
		return 0, "", false
	}

	qty := 0
	for _, d := range digits {
		qty = qty*10 + int(d-'0')
	}
	if qty <= 0 {
		return 0, "", false
	}

	return qty, remainder, true
}

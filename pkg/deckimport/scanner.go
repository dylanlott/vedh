package deckimport

import (
	"fmt"
	"strings"
)

// MaxDecklistBytes bounds the size of text accepted by Parse, counted in
// bytes (Go strings are byte sequences, so len(text) is already a byte
// count). Enforced before scanning so a caller cannot force unbounded
// memory use through a public, unauthenticated mutation.
const MaxDecklistBytes = 256 * 1024

// MaxDecklistLines bounds the number of lines accepted by Parse, enforced
// for the same reason as MaxDecklistBytes.
const MaxDecklistLines = 2000

// Parse turns pasted decklist text into a ParsedDeck. It implements the
// complete canonical decklist grammar: all six required syntaxes
// (`1 Sol Ring`, `1x Sol Ring`, `1,Sol Ring`, `1, Sol Ring`, quoted names,
// and bare names with an internal comma), double-faced card names written
// with " // ", D-07 printing metadata (set code, collector number,
// category), and D-11/D-12 section handling.
//
// Name comparison and lower-casing anywhere in this package use ASCII
// lower-casing only — two names that render alike but differ in Unicode
// code points are treated as different names; there is no Unicode
// normalization step. MaxDecklistBytes and MaxDecklistLines are both
// counted in bytes/lines of the raw input, not runes.
//
// Parse never fails the whole deck for a malformed line: a line whose
// quantity is missing, non-positive, or otherwise unreadable becomes a
// Warning, never a BlockingError. The two size limits above are the only
// conditions that produce a BlockingError on their own, because reading
// unbounded input is the one failure mode a per-line warning cannot bound.
// Parse also asserts its own "no nonblank row disappears" accounting
// invariant (AssertAccounting) before returning, so a bug that drops a row
// becomes a visible BlockingError rather than a silently short deck.
//
// Parse is pure: it reads no clock, no random source, and no database, so
// calling it twice on the same text always returns a deeply equal result.
// It holds no package-level mutable state, so it is safe to call
// concurrently from multiple goroutines.
//
// Parse logs nothing. Warnings and blocking errors are return values.
func Parse(text string) ParsedDeck {
	return ParseWithSource(text, DetectPasteSource(text))
}

// ParseWithSource is Parse with the ParsedDeck.Source field forced to a
// caller-supplied value instead of computed by DetectPasteSource. It exists
// so tests (and, later, Phase 2's URL-host detector) can prove D-23's
// structural guarantee: forcing Source to any of the four SourceType values
// never changes any other field of the result. Production code should call
// Parse; ParseWithSource is the seam that makes the "detection cannot
// influence parsing" property something a test can assert directly rather
// than something a reviewer has to trust.
func ParseWithSource(text string, source SourceType) ParsedDeck {
	deck := ParsedDeck{Source: source}

	// Rule 1: enforce the two input bounds before scanning at all. A
	// breach emits exactly one BlockingError and returns; nothing below
	// this point runs, so there's nothing left to account for.
	if len(text) > MaxDecklistBytes {
		deck.BlockingErrors = append(deck.BlockingErrors, BlockingError{
			Message: "Decklist is too large to import.",
		})
		return deck
	}

	lines := splitLines(text)
	if len(lines) > MaxDecklistLines {
		deck.BlockingErrors = append(deck.BlockingErrors, BlockingError{
			Message: "Decklist has too many lines to import.",
		})
		return deck
	}

	for i, raw := range lines {
		sourceLine := i + 1
		line := raw

		// Rule 2: blank lines (empty or whitespace-only) are the one
		// thing permitted to vanish without being counted anywhere —
		// splitLines has already stripped any line terminator, so this
		// is a pure content check.
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Rule 3: comment and section-header check, anchored at the
		// start of the line only. The anchor is load-bearing (Pitfall
		// 2): "//" also separates the two faces of a double-faced card,
		// and test/decklists/jarad.csv ships several of those unquoted.
		// Checking only line-initial "//" means a double slash anywhere
		// else on the line is never mistaken for a comment. Header
		// words are recognized here too, so they are never mistaken for
		// an entry — real section handling (dropping, D-11's summary
		// warning, D-12's commander preselection) is added in a later
		// plan; this task only guarantees a header line never becomes
		// an Entry.
		if isCommentLine(line) || isSectionHeaderWord(line) {
			continue
		}

		deck.NonBlankLines++

		// Rule 4: quantity extraction.
		qty, rest, warnMsg, ok := extractQuantity(line)
		if !ok {
			deck.Warnings = append(deck.Warnings, Warning{
				SourceLine: sourceLine,
				RawLine:    raw,
				Message:    warnMsg,
			})
			continue
		}

		// Rules 5-7: quoted-name check, else trailing-annotation
		// extraction on the unquoted remainder, else the name is
		// everything left, trimmed.
		name, setCode, collectorNumber, category := resolveNameAndMetadata(rest)

		// Rule 8: emit the entry with its SourceLine and RawLine.
		deck.Entries = append(deck.Entries, ParsedEntry{
			Quantity:        qty,
			Name:            name,
			SetCode:         setCode,
			CollectorNumber: collectorNumber,
			Category:        category,
			Section:         SectionMain,
			SourceLine:      sourceLine,
			RawLine:         raw,
		})
	}

	AssertAccounting(&deck)

	return deck
}

// AssertAccounting checks the "no nonblank row disappears" invariant —
// len(Entries)+len(Warnings)+len(BlockingErrors)+DroppedSectionRows must
// equal NonBlankLines — and, on mismatch, appends a BlockingError
// describing the discrepancy instead of returning a plausible-looking but
// silently short deck. It is exported so a test can drive it directly
// against a deliberately inconsistent ParsedDeck, and Parse/ParseWithSource
// call it on every normal return path (never on the early-return bounds
// checks above, where NonBlankLines is trivially zero).
func AssertAccounting(deck *ParsedDeck) {
	accounted := len(deck.Entries) + len(deck.Warnings) + len(deck.BlockingErrors) + deck.DroppedSectionRows
	if accounted != deck.NonBlankLines {
		deck.BlockingErrors = append(deck.BlockingErrors, BlockingError{
			Message: fmt.Sprintf(
				"internal accounting error: %d accounted row(s) (entries+warnings+errors+dropped) does not match %d nonblank input line(s)",
				accounted, deck.NonBlankLines,
			),
		})
	}
}

// splitLines splits text into lines, treating a lone "\r", a lone "\n", and
// the pair "\r\n" as identical line terminators, and never including the
// terminator itself in the returned line. This is what makes a Windows
// paste ("\r\n"), a classic-Mac paste ("\r"), and a Unix paste ("\n") all
// behave the same way, with no separate TrimSuffix(line, "\r") step needed
// afterward.
func splitLines(text string) []string {
	var lines []string
	start := 0
	i := 0
	for i < len(text) {
		switch text[i] {
		case '\n':
			lines = append(lines, text[start:i])
			i++
			start = i
		case '\r':
			lines = append(lines, text[start:i])
			i++
			if i < len(text) && text[i] == '\n' {
				i++
			}
			start = i
		default:
			i++
		}
	}
	lines = append(lines, text[start:])
	return lines
}

// isCommentLine reports whether line is a comment: line-initial "//" or a
// line-initial "# " (hash followed by a space). Both checks are anchored at
// index 0 — no leading whitespace is tolerated — because the anchor is what
// keeps a double-faced card's " // " face separator from ever being read as
// a comment.
func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "# ")
}

// sectionHeaderWords are the five words 01-CONTEXT.md's D-11/D-12 assign
// section meaning to. isSectionHeaderWord below is deliberately the only
// thing this task builds on top of them: recognizing a header line well
// enough that it never becomes an Entry. Real section semantics — dropping
// Sideboard/Maybeboard rows with a summary warning, preselecting Commander
// candidates — are added by the next plan in this phase.
var sectionHeaderWords = map[string]bool{
	"commander":  true,
	"sideboard":  true,
	"maybeboard": true,
	"deck":       true,
	"companion":  true,
}

// isSectionHeaderWord reports whether line, once an optional trailing
// "(<count>)" is stripped, is exactly one of the five section header words
// (case-insensitive on the word, anchored to the whole line). Requiring the
// header word to constitute the *whole* line — apart from that optional
// count suffix — is what keeps "1 Commander's Sphere" an entry: it has
// content beyond "Commander" plus an optional count, so it never matches.
func isSectionHeaderWord(line string) bool {
	trimmed := strings.TrimSpace(line)

	if idx := strings.LastIndexByte(trimmed, '('); idx != -1 && strings.HasSuffix(trimmed, ")") {
		inner := trimmed[idx+1 : len(trimmed)-1]
		if inner != "" && isAllDigits(inner) {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
	}

	return sectionHeaderWords[strings.ToLower(trimmed)]
}

// isAllDigits reports whether s is non-empty and consists only of ASCII
// digits, used to recognize a header's optional "(N)" count suffix.
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// extractQuantity implements rule 4. It consumes an optional leading "-"
// (only to recognize and reject a negative-quantity attempt, never to
// produce a negative Quantity), then leading ASCII digits, then at most one
// "x"/"X", then at most one comma, then any run of spaces and tabs. This
// single ordered consumption is what makes "1 ", "1x ", "1,", and "1, " all
// equivalent prefixes to the name that follows.
//
// ok is false whenever the line cannot become an Entry: warnMsg then holds
// the player-facing reason (too many digits, non-positive quantity, or a
// quantity with nothing after it), and the caller must record a Warning,
// never a BlockingError, and never a whole-deck failure.
//
// When the line has no leading digits (and no leading "-" either), the
// quantity defaults to 1 and the entire trimmed line becomes rest — this is
// what makes a bare "Sol Ring" line, and every card name in the corpus that
// happens to start with a letter, work with no special-casing.
func extractQuantity(line string) (quantity int, rest string, warnMsg string, ok bool) {
	i := 0
	neg := false
	if i < len(line) && line[i] == '-' {
		neg = true
		i++
	}

	digitStart := i
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	digitCount := i - digitStart

	if digitCount == 0 {
		// No digit run at all: not a quantity attempt. Even a lone
		// leading "-" with no digits behind it is just the start of a
		// name (no test relies on this; it is the conservative choice).
		return 1, strings.TrimSpace(line), "", true
	}

	if digitCount > 4 {
		// More than four digits: reject without ever parsing the value,
		// so an absurdly long digit run can never risk overflow.
		return 0, "", "This quantity has too many digits to be valid; the line was skipped.", false
	}

	if i < len(line) && (line[i] == 'x' || line[i] == 'X') {
		i++
	}
	if i < len(line) && line[i] == ',' {
		i++
	}
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}

	rest = line[i:]
	if strings.TrimSpace(rest) == "" {
		return 0, "", "This line has a quantity but no card name; it was skipped.", false
	}

	value := 0
	for _, d := range line[digitStart : digitStart+digitCount] {
		value = value*10 + int(d-'0')
	}
	if neg {
		value = -value
	}
	if value <= 0 {
		return 0, "", fmt.Sprintf("Quantity %d is not a valid card count; the line was skipped.", value), false
	}

	return value, rest, "", true
}

// resolveNameAndMetadata implements rules 5-7 on rest, the text remaining
// after extractQuantity has consumed the quantity and its one separator.
func resolveNameAndMetadata(rest string) (name, setCode, collectorNumber, category string) {
	// Rule 5: quoted-name check. This is the only place CSV quoting
	// semantics apply, and it is what makes the quoted Atraxa form work.
	if strings.HasPrefix(rest, `"`) {
		if closeIdx := strings.IndexByte(rest[1:], '"'); closeIdx != -1 {
			return rest[1 : 1+closeIdx], "", "", ""
		}
		// Malformed: no closing quote. Take the remainder verbatim
		// rather than dropping it — a row must never silently vanish.
		return rest[1:], "", "", ""
	}

	// Rules 6-7: trailing-annotation extraction on the unquoted
	// remainder, then the name is everything left, trimmed.
	return extractTrailingAnnotations(rest)
}

// extractTrailingAnnotations implements rule 6, stripping D-07 printing
// metadata from the *unquoted* remainder, right to left, in this order:
// hash tags, a trailing backtick-delimited category, a trailing bracketed
// category, a trailing bare alphanumeric collector number that follows a
// parenthesised group, and a trailing parenthesised set group. What remains
// is rule 7's name, trimmed. Set code letter case is preserved exactly as
// written: the corpus asserts both "C21" and "c21" round-trip unchanged,
// because the printing-disambiguation lookup planned for later in this
// phase lower-cases at query time, not at parse time.
func extractTrailingAnnotations(s string) (name, setCode, collectorNumber, category string) {
	s = strings.TrimRight(s, " \t")

	// Hash tags: one or more trailing "#tag" tokens, each preceded by a
	// token boundary (start of string or whitespace).
	var tags []string
	for {
		trimmed := strings.TrimRight(s, " \t")
		idx := strings.LastIndexByte(trimmed, '#')
		if idx == -1 {
			break
		}
		if idx > 0 && trimmed[idx-1] != ' ' && trimmed[idx-1] != '\t' {
			break
		}
		tag := trimmed[idx+1:]
		if tag == "" || strings.ContainsAny(tag, " \t") {
			break
		}
		tags = append([]string{tag}, tags...) // prepend: keep left-to-right order
		s = trimmed[:idx]
	}
	if len(tags) > 0 {
		category = strings.Join(tags, ",")
	}
	s = strings.TrimRight(s, " \t")

	// Trailing backtick-delimited category, e.g. the Archidekt
	// "`Maybeboard`" convention.
	if strings.HasSuffix(s, "`") {
		rest := s[:len(s)-1]
		if idx := strings.LastIndexByte(rest, '`'); idx != -1 {
			if inner := rest[idx+1:]; inner != "" {
				category = inner
				s = strings.TrimRight(rest[:idx], " \t")
			}
		}
	}

	// Trailing bracketed category, e.g. "[Ramp]".
	if strings.HasSuffix(s, "]") {
		if idx := strings.LastIndexByte(s, '['); idx != -1 {
			if inner := s[idx+1 : len(s)-1]; inner != "" {
				category = inner
				s = strings.TrimRight(s[:idx], " \t")
			}
		}
	}

	// Trailing bare alphanumeric collector number, but only when it
	// follows a parenthesised group — otherwise a card name that happens
	// to end in a bare word (there are none in Magic, but nothing here
	// should assume that) would lose a word it should keep.
	if trimmed := strings.TrimRight(s, " \t"); trimmed != "" {
		lastSpace := strings.LastIndexAny(trimmed, " \t")
		var token, before string
		if lastSpace == -1 {
			token, before = trimmed, ""
		} else {
			token = trimmed[lastSpace+1:]
			before = strings.TrimRight(trimmed[:lastSpace], " \t")
		}
		if isAlnumToken(token) && strings.HasSuffix(before, ")") {
			collectorNumber = token
			s = before
		}
	}

	// Trailing parenthesised set group, e.g. "(C21)" or "(c21)".
	if trimmed := strings.TrimRight(s, " \t"); strings.HasSuffix(trimmed, ")") {
		if idx := strings.LastIndexByte(trimmed, '('); idx != -1 {
			if inner := trimmed[idx+1 : len(trimmed)-1]; inner != "" && !strings.ContainsAny(inner, " \t") {
				setCode = inner
				s = strings.TrimRight(trimmed[:idx], " \t")
			}
		}
	}

	name = strings.TrimSpace(s)
	return name, setCode, collectorNumber, category
}

// isAlnumToken reports whether s is non-empty and consists only of ASCII
// letters and digits, used to recognize a bare collector number that may
// itself contain letters (e.g. a promo suffix like "263a").
func isAlnumToken(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}

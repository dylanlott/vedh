package deckimport

// Section identifies which part of a decklist an entry belongs to.
type Section string

const (
	// SectionMain is the default section for an entry with no header.
	SectionMain Section = "main"
	// SectionCommander holds entries that follow a "Commander" header;
	// D-12 preselects these as commander candidates.
	SectionCommander Section = "commander"
	// SectionSideboard holds entries that follow a "Sideboard" header.
	SectionSideboard Section = "sideboard"
	// SectionMaybeboard holds entries that follow a "Maybeboard" header.
	SectionMaybeboard Section = "maybeboard"
)

// ParsedEntry is one resolved line of a decklist, before card-name lookup.
type ParsedEntry struct {
	// Quantity is the number of copies requested.
	Quantity int
	// Name is the card name exactly as parsed, commas preserved verbatim.
	Name string
	// SetCode is the D-07 printing set code, "" when absent.
	SetCode string
	// CollectorNumber is the D-07 printing collector number, "" when absent.
	CollectorNumber string
	// Category is the D-07 free-text category annotation, "" when absent.
	Category string
	// Section is which part of the decklist this entry belongs to.
	Section Section
	// SourceLine is the 1-based input line number this entry came from.
	SourceLine int
	// RawLine is the original, unmodified input line.
	RawLine string
}

// Warning describes a non-blocking issue found while parsing one line. A
// Warning never causes the whole deck to fail; it is a return value the
// caller may choose to surface.
type Warning struct {
	// SourceLine is the 1-based input line number this warning applies to.
	SourceLine int
	// RawLine is the original, unmodified input line.
	RawLine string
	// Message is a human-readable, product-language description.
	Message string
}

// BlockingError describes an issue severe enough that the deck as a whole
// cannot be imported. Unlike Warning, a non-empty BlockingErrors slice on a
// ParsedDeck should set CanContinue to false in the caller.
type BlockingError struct {
	// SourceLine is the 1-based input line number this error applies to,
	// or 0 when the error applies to the whole input rather than one line.
	SourceLine int
	// RawLine is the original, unmodified input line, "" when the error
	// applies to the whole input.
	RawLine string
	// Message is a human-readable, product-language description.
	Message string
}

// ParsedDeck is the complete, normalized result of parsing one decklist. It
// is the single canonical result type: previewDeck and final library
// creation both consume the same ParsedDeck, so a deck is never parsed
// twice under divergent rules.
type ParsedDeck struct {
	// Entries holds every successfully parsed line.
	Entries []ParsedEntry
	// Warnings holds every non-blocking issue found while parsing.
	Warnings []Warning
	// BlockingErrors holds every issue severe enough to block import.
	BlockingErrors []BlockingError
	// DroppedSectionRows counts rows silently dropped because they were a
	// sideboard/maybeboard section header or belonged to one (D-11).
	DroppedSectionRows int
	// NonBlankLines counts every input line that was not empty after
	// trimming. len(Entries)+len(Warnings)+len(BlockingErrors)+DroppedSectionRows
	// must equal NonBlankLines: no nonblank input row may disappear
	// without becoming an entry, a warning, or an error.
	NonBlankLines int
	// Source is the detected origin of the input text.
	Source SourceType
}

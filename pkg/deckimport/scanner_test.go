package deckimport

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// scannerCase is one row of the required-syntax corpus. wantEntries and
// wantWarnings are always 0 or 1 for these single-line inputs: exactly one
// of them is 1, except for the comment/blank rows where both are 0 (the
// only rows permitted to vanish with no accounting).
type scannerCase struct {
	name                string
	input               string
	wantEntries         int
	wantWarnings        int
	wantQuantity        int
	wantName            string
	wantSetCode         string
	wantCollectorNumber string
	wantCategory        string
}

// requiredSyntaxCorpus is the non-negotiable minimum from
// 01-RESEARCH.md's Validation Architecture and 01-CONTEXT.md's Specific
// Ideas: every one of the six required syntaxes, the comma-truncation
// regression, the double-slash distinction, and D-07 printing metadata.
var requiredSyntaxCorpus = []scannerCase{
	{name: "digit space name", input: "1 Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{name: "digit x space name", input: "1x Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{name: "digit X space name", input: "1X Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{name: "digit comma name", input: "1,Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{name: "digit comma space name", input: "1, Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{
		name:        "quoted comma name",
		input:       `1,"Atraxa, Praetors' Voice"`,
		wantEntries: 1, wantQuantity: 1, wantName: "Atraxa, Praetors' Voice",
	},
	{
		name:        "unquoted comma name after space",
		input:       "1 Atraxa, Praetors' Voice",
		wantEntries: 1, wantQuantity: 1, wantName: "Atraxa, Praetors' Voice",
	},
	{
		name:        "comma truncation regression",
		input:       "1,Atraxa, Praetors' Voice",
		wantEntries: 1, wantQuantity: 1, wantName: "Atraxa, Praetors' Voice",
	},
	{name: "quantity 4", input: "4 Lightning Bolt", wantEntries: 1, wantQuantity: 4, wantName: "Lightning Bolt"},
	{name: "quantity 10", input: "10 Forest", wantEntries: 1, wantQuantity: 10, wantName: "Forest"},
	{name: "no leading quantity", input: "Sol Ring", wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring"},
	{
		name:        "printing metadata set and collector",
		input:       "1 Sol Ring (C21) 263",
		wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring", wantSetCode: "C21", wantCollectorNumber: "263",
	},
	{
		name:        "printing metadata set collector and bracket category",
		input:       "1x Sol Ring (c21) 263 [Ramp]",
		wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring",
		wantSetCode: "c21", wantCollectorNumber: "263", wantCategory: "Ramp",
	},
	{
		name:        "backtick category",
		input:       "1 Sol Ring `Maybeboard`",
		wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring", wantCategory: "Maybeboard",
	},
	{
		name:        "hash tag categories",
		input:       "1 Sol Ring #ramp #artifact",
		wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring", wantCategory: "ramp,artifact",
	},
	{
		name:        "unquoted double-faced name",
		input:       "1,Bala Ged Recovery // Bala Ged Sanctuary",
		wantEntries: 1, wantQuantity: 1, wantName: "Bala Ged Recovery // Bala Ged Sanctuary",
	},
	{
		name:        "quoted double-faced name with internal comma",
		input:       `1,"Agadeem's Awakening // Agadeem, the Undercrypt"`,
		wantEntries: 1, wantQuantity: 1, wantName: "Agadeem's Awakening // Agadeem, the Undercrypt",
	},
	{name: "double slash comment", input: "// Commander", wantEntries: 0, wantWarnings: 0},
	{name: "empty line", input: "", wantEntries: 0, wantWarnings: 0},
	{name: "whitespace only line", input: "   ", wantEntries: 0, wantWarnings: 0},
	{
		name:        "trailing carriage return",
		input:       "1 Sol Ring\r",
		wantEntries: 1, wantQuantity: 1, wantName: "Sol Ring",
	},
	{name: "negative quantity", input: "-1 Sol Ring", wantEntries: 0, wantWarnings: 1},
	{name: "zero quantity", input: "0 Sol Ring", wantEntries: 0, wantWarnings: 1},
	{name: "quantity too many digits", input: "99999 Sol Ring", wantEntries: 0, wantWarnings: 1},
	{name: "quantity with only whitespace name", input: "1    ", wantEntries: 0, wantWarnings: 1},
}

func TestScanner_RequiredSyntaxes(t *testing.T) {
	for _, tc := range requiredSyntaxCorpus {
		t.Run(tc.name, func(t *testing.T) {
			deck := Parse(tc.input)

			if got := len(deck.Entries); got != tc.wantEntries {
				t.Fatalf("Parse(%q).Entries has %d entries, want %d (deck=%+v)", tc.input, got, tc.wantEntries, deck)
			}
			if got := len(deck.Warnings); got != tc.wantWarnings {
				t.Fatalf("Parse(%q).Warnings has %d warnings, want %d (deck=%+v)", tc.input, got, tc.wantWarnings, deck)
			}
			if len(deck.BlockingErrors) != 0 {
				t.Fatalf("Parse(%q).BlockingErrors = %+v, want none", tc.input, deck.BlockingErrors)
			}

			if tc.wantEntries == 0 {
				return
			}
			e := deck.Entries[0]
			if e.Quantity != tc.wantQuantity {
				t.Errorf("Quantity = %d, want %d", e.Quantity, tc.wantQuantity)
			}
			if e.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", e.Name, tc.wantName)
			}
			if e.SetCode != tc.wantSetCode {
				t.Errorf("SetCode = %q, want %q", e.SetCode, tc.wantSetCode)
			}
			if e.CollectorNumber != tc.wantCollectorNumber {
				t.Errorf("CollectorNumber = %q, want %q", e.CollectorNumber, tc.wantCollectorNumber)
			}
			if e.Category != tc.wantCategory {
				t.Errorf("Category = %q, want %q", e.Category, tc.wantCategory)
			}
		})
	}
}

// TestScanner_NoRowDisappears asserts the accounting invariant over every
// row of the required-syntax corpus joined into one document, and over
// both of the repo's own test/decklists/*.csv fixtures.
func TestScanner_NoRowDisappears(t *testing.T) {
	var lines []string
	for _, tc := range requiredSyntaxCorpus {
		lines = append(lines, tc.input)
	}
	assertNoRowDisappears(t, "required-syntax corpus", strings.Join(lines, "\n"))

	for _, fixture := range []string{
		"../../test/decklists/jarad.csv",
		"../../test/decklists/kykar.csv",
	} {
		text, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("reading %s: %v", fixture, err)
		}
		assertNoRowDisappears(t, fixture, string(text))
	}
}

func assertNoRowDisappears(t *testing.T, label, text string) {
	t.Helper()
	deck := Parse(text)
	accounted := len(deck.Entries) + len(deck.Warnings) + len(deck.BlockingErrors) + deck.DroppedSectionRows
	if accounted != deck.NonBlankLines {
		t.Errorf(
			"%s: accounted %d (entries=%d warnings=%d errors=%d dropped=%d) != NonBlankLines %d",
			label, accounted, len(deck.Entries), len(deck.Warnings), len(deck.BlockingErrors), deck.DroppedSectionRows, deck.NonBlankLines,
		)
	}
	if len(deck.BlockingErrors) > 0 {
		t.Errorf("%s: unexpected BlockingErrors from the corpus itself: %+v", label, deck.BlockingErrors)
	}
}

// TestScanner_AccountingInvariantCatchesInconsistency drives AssertAccounting
// directly against a deliberately inconsistent ParsedDeck — constructed
// without going through Parse — and asserts it appends exactly the one
// blocking error the invariant promises.
func TestScanner_AccountingInvariantCatchesInconsistency(t *testing.T) {
	deck := ParsedDeck{NonBlankLines: 5} // nothing accounts for any of the 5 rows

	AssertAccounting(&deck)

	if len(deck.BlockingErrors) != 1 {
		t.Fatalf("BlockingErrors = %+v, want exactly one", deck.BlockingErrors)
	}
	if deck.BlockingErrors[0].Message == "" {
		t.Error("accounting BlockingError has an empty Message")
	}

	// A deck that IS consistent must never trip the invariant.
	consistent := ParsedDeck{
		NonBlankLines: 1,
		Entries:       []ParsedEntry{{Quantity: 1, Name: "Sol Ring"}},
	}
	AssertAccounting(&consistent)
	if len(consistent.BlockingErrors) != 0 {
		t.Fatalf("BlockingErrors = %+v, want none for a consistent deck", consistent.BlockingErrors)
	}
}

// TestScanner_IsPure asserts Parse reads no clock, no random source, and no
// database: two calls on the same text return a deeply equal result.
func TestScanner_IsPure(t *testing.T) {
	text := "1 Sol Ring\n2 Lightning Bolt\n1,\"Atraxa, Praetors' Voice\"\n"
	a := Parse(text)
	b := Parse(text)
	assertDecksEqual(t, "repeat Parse call", a, b)
}

// TestScanner_ConcurrentParseIsSafe parses the fixture corpus from several
// goroutines concurrently; run under `go test -race` this proves the
// package holds no mutable package-level state.
func TestScanner_ConcurrentParseIsSafe(t *testing.T) {
	texts := []string{
		"1 Sol Ring\n2 Lightning Bolt\n",
		"1,Sol Ring\n1, Sol Ring\n",
		"1,\"Atraxa, Praetors' Voice\"\n1 Atraxa, Praetors' Voice\n",
		"Sideboard\n1 Sol Ring\n1 Lightning Bolt\n",
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			text := texts[n%len(texts)]
			_ = Parse(text)
		}(i)
	}
	wg.Wait()
}

// TestScanner_InputBounds asserts MaxDecklistBytes and MaxDecklistLines are
// each enforced exactly at their boundary, in both directions, with no
// panic either way.
func TestScanner_InputBounds(t *testing.T) {
	t.Run("over byte limit", func(t *testing.T) {
		text := strings.Repeat("a", MaxDecklistBytes+1)
		deck := Parse(text)
		if len(deck.Entries) != 0 {
			t.Fatalf("Entries = %+v, want none", deck.Entries)
		}
		if len(deck.BlockingErrors) != 1 {
			t.Fatalf("BlockingErrors = %+v, want exactly one", deck.BlockingErrors)
		}
	})

	t.Run("under byte limit", func(t *testing.T) {
		name := strings.Repeat("A", MaxDecklistBytes-1-len("1 "))
		text := "1 " + name
		if len(text) != MaxDecklistBytes-1 {
			t.Fatalf("constructed text is %d bytes, want %d", len(text), MaxDecklistBytes-1)
		}
		deck := Parse(text)
		if len(deck.BlockingErrors) != 0 {
			t.Fatalf("BlockingErrors = %+v, want none", deck.BlockingErrors)
		}
		if len(deck.Entries) != 1 || deck.Entries[0].Name != name {
			t.Fatalf("Entries = %+v, want one entry named the padded name", deck.Entries)
		}
	})

	t.Run("over line limit", func(t *testing.T) {
		lines := make([]string, MaxDecklistLines+1)
		for i := range lines {
			lines[i] = "1 Sol Ring"
		}
		deck := Parse(strings.Join(lines, "\n"))
		if len(deck.Entries) != 0 {
			t.Fatalf("Entries = %+v, want none", deck.Entries)
		}
		if len(deck.BlockingErrors) != 1 {
			t.Fatalf("BlockingErrors = %+v, want exactly one", deck.BlockingErrors)
		}
	})

	t.Run("under line limit", func(t *testing.T) {
		lines := make([]string, MaxDecklistLines-1)
		for i := range lines {
			lines[i] = "1 Sol Ring"
		}
		deck := Parse(strings.Join(lines, "\n"))
		if len(deck.BlockingErrors) != 0 {
			t.Fatalf("BlockingErrors = %+v, want none", deck.BlockingErrors)
		}
		if len(deck.Entries) != MaxDecklistLines-1 {
			t.Fatalf("Entries has %d entries, want %d", len(deck.Entries), MaxDecklistLines-1)
		}
	})
}

// assertDecksEqual compares two ParsedDeck values field by field, since
// slices of structs are not comparable with ==.
func assertDecksEqual(t *testing.T, label string, a, b ParsedDeck) {
	t.Helper()
	if a.Source != b.Source {
		t.Errorf("%s: Source differs: %q vs %q", label, a.Source, b.Source)
	}
	if a.NonBlankLines != b.NonBlankLines {
		t.Errorf("%s: NonBlankLines differs: %d vs %d", label, a.NonBlankLines, b.NonBlankLines)
	}
	if a.DroppedSectionRows != b.DroppedSectionRows {
		t.Errorf("%s: DroppedSectionRows differs: %d vs %d", label, a.DroppedSectionRows, b.DroppedSectionRows)
	}
	if len(a.Entries) != len(b.Entries) {
		t.Fatalf("%s: Entries length differs: %d vs %d", label, len(a.Entries), len(b.Entries))
	}
	for i := range a.Entries {
		if a.Entries[i] != b.Entries[i] {
			t.Errorf("%s: Entries[%d] differs: %+v vs %+v", label, i, a.Entries[i], b.Entries[i])
		}
	}
	if len(a.Warnings) != len(b.Warnings) {
		t.Fatalf("%s: Warnings length differs: %d vs %d", label, len(a.Warnings), len(b.Warnings))
	}
	for i := range a.Warnings {
		if a.Warnings[i] != b.Warnings[i] {
			t.Errorf("%s: Warnings[%d] differs: %+v vs %+v", label, i, a.Warnings[i], b.Warnings[i])
		}
	}
	if len(a.BlockingErrors) != len(b.BlockingErrors) {
		t.Fatalf("%s: BlockingErrors length differs: %d vs %d", label, len(a.BlockingErrors), len(b.BlockingErrors))
	}
	for i := range a.BlockingErrors {
		if a.BlockingErrors[i] != b.BlockingErrors[i] {
			t.Errorf("%s: BlockingErrors[%d] differs: %+v vs %+v", label, i, a.BlockingErrors[i], b.BlockingErrors[i])
		}
	}
}

// testdataPath is a small helper the golden and sourcetype tests share.
func testdataPath(elems ...string) string {
	return filepath.Join(append([]string{"testdata"}, elems...)...)
}

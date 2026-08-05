package deckimport

import (
	"strings"
	"testing"
)

// TestSections_DropAndPreselect is the test 01-VALIDATION.md's D-11/D-12 row
// names: a Sideboard/Maybeboard header drops its rows with a visible count,
// a Commander header preselects candidates, and a name that merely begins
// with a header word is never mistaken for one.
func TestSections_DropAndPreselect(t *testing.T) {
	t.Run("sideboard header drops twelve rows with one summary warning", func(t *testing.T) {
		var lines []string
		lines = append(lines, "Sideboard")
		for i := 0; i < 12; i++ {
			lines = append(lines, "1 Sol Ring")
		}
		deck := Parse(strings.Join(lines, "\n"))

		if len(deck.Entries) != 0 {
			t.Fatalf("Entries = %+v, want none", deck.Entries)
		}
		if deck.DroppedSectionRows != 12 {
			t.Fatalf("DroppedSectionRows = %d, want 12", deck.DroppedSectionRows)
		}
		if len(deck.Warnings) != 1 {
			t.Fatalf("Warnings = %+v, want exactly one", deck.Warnings)
		}
		if !strings.Contains(deck.Warnings[0].Message, "12") {
			t.Errorf("Warnings[0].Message = %q, want it to state the dropped count", deck.Warnings[0].Message)
		}
	})

	t.Run("maybeboard header drops rows the same way", func(t *testing.T) {
		var lines []string
		lines = append(lines, "Maybeboard")
		for i := 0; i < 5; i++ {
			lines = append(lines, "1 Sol Ring")
		}
		deck := Parse(strings.Join(lines, "\n"))

		if len(deck.Entries) != 0 {
			t.Fatalf("Entries = %+v, want none", deck.Entries)
		}
		if deck.DroppedSectionRows != 5 {
			t.Fatalf("DroppedSectionRows = %d, want 5", deck.DroppedSectionRows)
		}
		if len(deck.Warnings) != 1 || !strings.Contains(deck.Warnings[0].Message, "5") {
			t.Fatalf("Warnings = %+v, want one warning stating 5", deck.Warnings)
		}
	})

	t.Run("header with parenthesised count suffix is recognized", func(t *testing.T) {
		deck := Parse("Sideboard (2)\n1 Sol Ring\n1 Lightning Bolt\n")
		if deck.DroppedSectionRows != 2 {
			t.Fatalf("DroppedSectionRows = %d, want 2", deck.DroppedSectionRows)
		}
		if len(deck.Entries) != 0 {
			t.Fatalf("Entries = %+v, want none", deck.Entries)
		}
	})

	t.Run("commander header preselects candidates and leaves them in entries", func(t *testing.T) {
		deck := Parse("Commander\n1 Atraxa, Praetors' Voice\n1 Sol Ring\n")
		if len(deck.Entries) != 2 {
			t.Fatalf("Entries = %+v, want 2", deck.Entries)
		}
		for _, e := range deck.Entries {
			if e.Section != SectionCommander {
				t.Errorf("Entries[%q].Section = %q, want %q", e.Name, e.Section, SectionCommander)
			}
		}
	})

	t.Run("deck and companion headers are recognized and produce no entry", func(t *testing.T) {
		deck := Parse("Deck\n1 Sol Ring\nCompanion\n1 Lutri, the Spellchaser\n")
		if len(deck.Entries) != 2 {
			t.Fatalf("Entries = %+v, want 2", deck.Entries)
		}
		for _, e := range deck.Entries {
			if e.Section != SectionMain {
				t.Errorf("Entries[%q].Section = %q, want %q", e.Name, e.Section, SectionMain)
			}
		}
	})

	t.Run("header matching is case-insensitive on the header word", func(t *testing.T) {
		deck := Parse("SIDEBOARD\n1 Sol Ring\n")
		if deck.DroppedSectionRows != 1 {
			t.Fatalf("DroppedSectionRows = %d, want 1", deck.DroppedSectionRows)
		}
	})

	t.Run("a card name beginning with a header word is an entry, not a header", func(t *testing.T) {
		deck := Parse("1 Commander's Sphere\n")
		if len(deck.Entries) != 1 {
			t.Fatalf("Entries = %+v, want exactly one", deck.Entries)
		}
		if deck.Entries[0].Name != "Commander's Sphere" {
			t.Errorf("Name = %q, want %q", deck.Entries[0].Name, "Commander's Sphere")
		}
		if deck.Entries[0].Section != SectionMain {
			t.Errorf("Section = %q, want %q", deck.Entries[0].Section, SectionMain)
		}
	})

	t.Run("dropped rows still count toward NonBlankLines", func(t *testing.T) {
		deck := Parse("Sideboard\n1 Sol Ring\n1 Lightning Bolt\n1 Forest\n")
		accounted := len(deck.Entries) + len(deck.Warnings) + len(deck.BlockingErrors) + deck.DroppedSectionRows
		if accounted != deck.NonBlankLines {
			t.Fatalf("accounted %d != NonBlankLines %d (deck=%+v)", accounted, deck.NonBlankLines, deck)
		}
		if len(deck.BlockingErrors) != 0 {
			t.Fatalf("BlockingErrors = %+v, want none", deck.BlockingErrors)
		}
	})
}

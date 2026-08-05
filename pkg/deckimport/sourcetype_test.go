package deckimport

import (
	"os"
	"strings"
	"testing"
)

func allSourceTypeSet() map[SourceType]bool {
	set := make(map[SourceType]bool)
	for _, s := range AllSourceTypes() {
		set[s] = true
	}
	return set
}

// TestSourceDetection_ReturnsOnlyEnumMembers asserts DetectPasteSource never
// returns a value outside AllSourceTypes(), for the fixture corpus and for
// adversarial input.
func TestSourceDetection_ReturnsOnlyEnumMembers(t *testing.T) {
	valid := allSourceTypeSet()

	inputs := []string{
		"",
		"\n",
		"   \n\n  ",
		"Sol Ring",
		"1 Sol Ring\n2 Lightning Bolt\n",
		"1 Sol Ring (C21) 263\n1 Lightning Bolt (M20) 150\n",
		"1 Sol Ring [Ramp]\n1 Lightning Bolt `Burn`\n",
	}
	for _, fixture := range fixturePaths {
		text, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("reading %s: %v", fixture, err)
		}
		inputs = append(inputs, string(text))
	}

	for _, in := range inputs {
		got := DetectPasteSource(in)
		if !valid[got] {
			t.Errorf("DetectPasteSource(%.30q...) = %q, not a member of AllSourceTypes()", in, got)
		}
	}
}

// TestSourceDetection_ShapesDetectDistinctly asserts a Moxfield-shaped
// export, an Archidekt-shaped export, a bare plain-text list, and a
// document mixing both providers' markers each detect as the SourceType
// the paste-shape heuristics are documented to key on.
func TestSourceDetection_ShapesDetectDistinctly(t *testing.T) {
	moxfieldShaped := buildLines(func(add func(string)) {
		add("1 Sol Ring (C21) 263")
		add("1 Lightning Bolt (M20) 150")
		add("1 Command Tower (C21) 269")
		add("1 Forest (C21) 273")
	})
	archidektBracket := buildLines(func(add func(string)) {
		add("1 Sol Ring [Ramp]")
		add("1 Lightning Bolt [Removal]")
		add("1 Forest [Land]")
	})
	archidektHeaderCount := buildLines(func(add func(string)) {
		add("1 Sol Ring")
		add("1 Lightning Bolt")
		add("Sideboard (2)")
		add("1 Forest")
		add("1 Island")
	})
	plainText := buildLines(func(add func(string)) {
		add("1 Sol Ring")
		add("1 Lightning Bolt")
		add("1 Forest")
	})
	mixedProviders := buildLines(func(add func(string)) {
		add("1 Sol Ring (C21) 263")
		add("1 Lightning Bolt [Removal]")
	})

	cases := []struct {
		name string
		text string
		want SourceType
	}{
		{"moxfield-shaped uppercase set+collector majority", moxfieldShaped, SourceMoxfield},
		{"archidekt-shaped bracket category", archidektBracket, SourceArchidekt},
		{"archidekt-shaped header count suffix", archidektHeaderCount, SourceArchidekt},
		{"bare plain text, no markers", plainText, SourcePlainText},
		{"mixed provider markers", mixedProviders, SourceUnknown},
		{"empty text", "", SourceUnknown},
		{"single blank line", "\n", SourceUnknown},
		{"single word, no quantity", "Sol", SourcePlainText},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectPasteSource(tc.text); got != tc.want {
				t.Errorf("DetectPasteSource(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}

// TestSourceDetection_DoesNotAffectParse is D-23's structural proof: for
// every fixture, parsing normally and parsing with the detected source
// forcibly overridden to each of the four enum values must produce deeply
// equal ParsedDecks apart from the Source field.
func TestSourceDetection_DoesNotAffectParse(t *testing.T) {
	for _, fixture := range fixturePaths {
		text, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("reading %s: %v", fixture, err)
		}
		t.Run(fixture, func(t *testing.T) {
			baseline := Parse(string(text))
			for _, forced := range AllSourceTypes() {
				forcedDeck := ParseWithSource(string(text), forced)
				if forcedDeck.Source != forced {
					t.Fatalf("ParseWithSource source = %q, want %q", forcedDeck.Source, forced)
				}
				assertDecksEqualIgnoringSource(t, fixture+" forced to "+string(forced), baseline, forcedDeck)
			}
		})
	}
}

// assertDecksEqualIgnoringSource is assertDecksEqual without the Source
// comparison — used specifically to prove D-23's isolation property.
func assertDecksEqualIgnoringSource(t *testing.T, label string, a, b ParsedDeck) {
	t.Helper()
	a.Source, b.Source = "", ""
	assertDecksEqual(t, label, a, b)
}

// buildLines is a small test helper for assembling a multi-line document
// without repeating strings.Join(lines, "\n") everywhere.
func buildLines(fn func(add func(string))) string {
	var lines []string
	fn(func(s string) { lines = append(lines, s) })
	return strings.Join(lines, "\n")
}

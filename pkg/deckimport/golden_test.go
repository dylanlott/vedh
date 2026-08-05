package deckimport

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"
)

// updateGolden regenerates the checked-in *.expected.json files from the
// current parser output. It must never be combined with CI=true — that
// combination would let a CI run silently rewrite the golden files a
// regression test exists to protect.
var updateGolden = flag.Bool("update", false, "regenerate testdata/*.expected.json from the current parser output")

// fixturePaths are the three whole-file golden fixtures this task ships:
// Moxfield-shaped, Archidekt-shaped, and generic plain-text. sourcetype_test.go
// reuses this list for its own corpus-wide assertions.
var fixturePaths = []string{
	testdataPath("moxfield_export.txt"),
	testdataPath("archidekt_export.txt"),
	testdataPath("plain_text_export.txt"),
}

// TestGolden_Fixtures parses each fixture and compares it against its
// checked-in expected ParsedDeck, additionally asserting the accounting
// invariant holds for real, whole-file input rather than only the synthetic
// corpus.
func TestGolden_Fixtures(t *testing.T) {
	if *updateGolden && os.Getenv("CI") != "" {
		t.Fatal("-update must never run in CI (CI is set); regenerate the golden files locally and commit the result")
	}

	for _, fixture := range fixturePaths {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			text, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("reading %s: %v", fixture, err)
			}
			deck := Parse(string(text))

			accounted := len(deck.Entries) + len(deck.Warnings) + len(deck.BlockingErrors) + deck.DroppedSectionRows
			if accounted != deck.NonBlankLines {
				t.Errorf("%s: accounted %d (entries=%d warnings=%d errors=%d dropped=%d) != NonBlankLines %d",
					fixture, accounted, len(deck.Entries), len(deck.Warnings), len(deck.BlockingErrors), deck.DroppedSectionRows, deck.NonBlankLines)
			}
			if len(deck.BlockingErrors) != 0 {
				t.Errorf("%s: unexpected BlockingErrors: %+v", fixture, deck.BlockingErrors)
			}

			expectedPath := strings.TrimSuffix(fixture, ".txt") + ".expected.json"

			if *updateGolden {
				data, err := json.MarshalIndent(deck, "", "  ")
				if err != nil {
					t.Fatalf("marshaling %s: %v", fixture, err)
				}
				data = append(data, '\n')
				if err := os.WriteFile(expectedPath, data, 0o644); err != nil {
					t.Fatalf("writing %s: %v", expectedPath, err)
				}
				return
			}

			wantData, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("reading %s: %v (run `go test ./pkg/deckimport -run TestGolden_Fixtures -update` to generate it)", expectedPath, err)
			}
			var want ParsedDeck
			if err := json.Unmarshal(wantData, &want); err != nil {
				t.Fatalf("unmarshaling %s: %v", expectedPath, err)
			}

			assertDecksEqual(t, fixture, deck, want)
		})
	}
}

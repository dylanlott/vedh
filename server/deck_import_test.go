package server

import (
	"context"
	"database/sql"
	"math"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/openmtg/edh-go/persistence"
	"github.com/openmtg/edh-go/pkg/deckimport"
)

// TestTracer_PreviewDeckEmitsMeasuredEvent proves the whole Phase 1
// architecture with one thin, production-quality path: a pasted
// "1 Sol Ring" travels through the GraphQL contract, pkg/deckimport, the
// existing cards lookup, and back out as a DeckPreview, while the same
// request writes a deck_import_succeeded row and moves a Prometheus
// counter.
func TestTracer_PreviewDeckEmitsMeasuredEvent(t *testing.T) {
	s := testAPI(t)
	sessionID := "tracer-session-" + t.Name()

	t.Cleanup(func() {
		if _, err := s.db.Exec(`DELETE FROM product_events WHERE session_id = $1`, sessionID); err != nil {
			t.Logf("cleanup: delete product_events: %v", err)
		}
	})

	before := testutil.ToFloat64(collectors.DeckImportCounter(deckimport.SourcePlainText, "success"))

	text := "1 Sol Ring"
	preview, err := s.PreviewDeck(context.Background(), InputDeckImport{
		Text:      &text,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview == nil {
		t.Fatal("PreviewDeck() returned nil preview")
	}
	if preview.CardCount != 1 {
		t.Fatalf("CardCount = %d, want 1", preview.CardCount)
	}
	if preview.SourceType != string(deckimport.SourcePlainText) {
		t.Fatalf("SourceType = %q, want %q", preview.SourceType, deckimport.SourcePlainText)
	}
	if !preview.CanContinue {
		t.Fatal("CanContinue = false, want true")
	}
	if len(preview.Entries) != 1 || preview.Entries[0].Name != "Sol Ring" || !preview.Entries[0].Resolved {
		t.Fatalf("Entries = %+v, want one resolved entry named Sol Ring", preview.Entries)
	}

	var rowCount int
	if err := s.db.QueryRow(
		`SELECT count(*) FROM product_events WHERE session_id = $1 AND event_name = 'deck_import_succeeded'`,
		sessionID,
	).Scan(&rowCount); err != nil {
		t.Fatalf("query product_events: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("deck_import_succeeded rows for session = %d, want 1", rowCount)
	}

	after := testutil.ToFloat64(collectors.DeckImportCounter(deckimport.SourcePlainText, "success"))
	if after != before+1 {
		t.Fatalf("vedh_deck_import_total{source=plain_text,outcome=success} delta = %v, want 1", after-before)
	}
}

// cardNameSearchIndexes are the four indexes this migration creates on top
// of the card_names table itself; assertCardNameSearchObjectsPresent checks
// all of them plus card_names in one pass.
var cardNameSearchIndexes = []string{
	"card_names_trgm_gist",
	"cards_name_lower_idx",
	"cards_facename_lower_idx",
	"cards_name_set_num_idx",
}

// assertCardNameSearchObjectsPresent asserts that the card_names table and
// its four indexes are present (present == true) or absent (present ==
// false) in the database reachable through db.
func assertCardNameSearchObjectsPresent(t *testing.T, db *sql.DB, present bool) {
	t.Helper()

	wantCount := 0
	if present {
		wantCount = 1
	}

	var tableCount int
	if err := db.QueryRow(
		`SELECT count(*) FROM pg_tables WHERE tablename = 'card_names'`,
	).Scan(&tableCount); err != nil {
		t.Fatalf("query pg_tables: %v", err)
	}
	if tableCount != wantCount {
		t.Fatalf("card_names table present = %v, want %v", tableCount == 1, present)
	}

	for _, idx := range cardNameSearchIndexes {
		var idxCount int
		if err := db.QueryRow(
			`SELECT count(*) FROM pg_indexes WHERE indexname = $1`, idx,
		).Scan(&idxCount); err != nil {
			t.Fatalf("query pg_indexes for %s: %v", idx, err)
		}
		if idxCount != wantCount {
			t.Fatalf("index %s present = %v, want %v", idx, idxCount == 1, present)
		}
	}
}

// assertTrgmExtensionInstalled asserts that pg_trgm is installed in the
// database reachable through db. The down migration deliberately never
// drops this extension (it is a shared, database-wide resource), so this
// assertion is expected to hold both before and after the down migration.
func assertTrgmExtensionInstalled(t *testing.T, db *sql.DB) {
	t.Helper()

	var count int
	if err := db.QueryRow(
		`SELECT count(*) FROM pg_extension WHERE extname = 'pg_trgm'`,
	).Scan(&count); err != nil {
		t.Fatalf("query pg_extension: %v", err)
	}
	if count != 1 {
		t.Fatal("pg_trgm extension not installed, want installed")
	}
}

// TestMigrations_CardNameSearch proves the name-search migration set (D-09
// prerequisite indexes, the card_names projection, and its trigram index)
// both exists in the already-migrated shared test database and survives a
// full up, down, up cycle in scratch databases for both migration
// directories -- reusing plan 01-01's withScratchMigrationDB by its exact
// name rather than declaring a second scratch-database helper. It never
// calls MigrateDown against the URL TestMain migrated: that would drop the
// seeded cards and the imported MTGJSON snapshot out from under every other
// test in this package.
func TestMigrations_CardNameSearch(t *testing.T) {
	s := testAPI(t)

	assertCardNameSearchObjectsPresent(t, s.db, true)
	assertTrgmExtensionInstalled(t, s.db)

	var count int
	if err := s.db.QueryRow(
		`SELECT count(*) FROM card_names WHERE name_lower = lower($1)`, "Sol Ring",
	).Scan(&count); err != nil {
		t.Fatalf("query card_names: %v", err)
	}
	if count == 0 {
		t.Fatal("card_names has no row for a name present in the seeded/imported cards, want at least one")
	}

	for _, dir := range []string{"../persistence/migrations/", "../persistence/migrations_test/"} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			withScratchMigrationDB(t, dir, func(t *testing.T, scratchURL string) {
				scratchDB, err := sql.Open("postgres", scratchURL)
				if err != nil {
					t.Fatalf("open scratch db: %v", err)
				}
				defer scratchDB.Close()

				// Up (via NewPostgres, which migrates and hands back an
				// open *sql.DB; closed immediately here, same as
				// TestMigrations_ProductEvents, so the connection this step
				// opens does not block withScratchMigrationDB's cleanup).
				upDB, err := persistence.NewPostgres(dir, scratchURL)
				if err != nil {
					t.Fatalf("up: %v", err)
				}
				if err := upDB.Close(); err != nil {
					t.Fatalf("close db after up: %v", err)
				}
				assertCardNameSearchObjectsPresent(t, scratchDB, true)
				assertTrgmExtensionInstalled(t, scratchDB)

				// Down.
				if err := persistence.MigrateDown(dir, scratchURL); err != nil {
					t.Fatalf("down: %v", err)
				}
				assertCardNameSearchObjectsPresent(t, scratchDB, false)
				// The one thing this migration's down deliberately does not
				// reverse: pg_trgm is a shared, database-wide resource.
				assertTrgmExtensionInstalled(t, scratchDB)

				// Up again.
				upAgainDB, err := persistence.NewPostgres(dir, scratchURL)
				if err != nil {
					t.Fatalf("up (again): %v", err)
				}
				if err := upAgainDB.Close(); err != nil {
					t.Fatalf("close db after second up: %v", err)
				}
				assertCardNameSearchObjectsPresent(t, scratchDB, true)
			})
		})
	}
}

// printingTestCard names one seeded printing of "Test Printing Card" used
// by the D-09/D-10 resolution tests below.
type printingTestCard struct {
	id      string
	setcode string
	number  string
}

// printingTestCards are two printings of the same name, deliberately
// ordered here with the alphabetically-later set code first, so a test
// relying on the deterministic lower(setcode) ASC ordering documented on
// nameOnlyFallbackQuery cannot pass by accident of insertion or slice order.
var printingTestCards = []printingTestCard{
	{id: "testprint-0004-xyz", setcode: "XYZ", number: "99"},
	{id: "testprint-0004-abc", setcode: "ABC", number: "1"},
}

// seedPrintingTestCards inserts printingTestCards as two printings of one
// card name ("Test Printing Card", chosen to never collide with a real
// MTGJSON name) and removes them (and any card_names projection row for
// that name) in t.Cleanup.
func seedPrintingTestCards(t *testing.T, s *graphQLServer) {
	t.Helper()

	for _, c := range printingTestCards {
		if _, err := s.db.Exec(
			`INSERT INTO cards (id, name, setcode, number, uuid) VALUES ($1, 'Test Printing Card', $2, $3, $1)
			ON CONFLICT (id) DO UPDATE SET setcode = EXCLUDED.setcode, number = EXCLUDED.number`,
			c.id, c.setcode, c.number,
		); err != nil {
			t.Fatalf("seed printing test card %s: %v", c.id, err)
		}
	}

	t.Cleanup(func() {
		for _, c := range printingTestCards {
			if _, err := s.db.Exec(`DELETE FROM cards WHERE id = $1`, c.id); err != nil {
				t.Logf("cleanup: delete card %s: %v", c.id, err)
			}
		}
		if _, err := s.db.Exec(
			`DELETE FROM card_names WHERE name_lower = lower('Test Printing Card')`,
		); err != nil {
			t.Logf("cleanup: delete card_names row: %v", err)
		}
	})
}

// previewCardText pastes one line naming Test Printing Card, optionally
// with a set code / collector number suffix in the "(SET) NUMBER" form the
// scanner recognizes.
func previewCardTextLine(setCode, number string) string {
	line := "1 Test Printing Card"
	if setCode != "" {
		line += " (" + setCode + ")"
		if number != "" {
			line += " " + number
		}
	}
	return line
}

// TestDeckImport_PrintingDisambiguation proves D-09's stage-one exact
// resolution: an entry naming a set code present in the snapshot selects
// that exact printing, and the upper-case and lower-case spellings of the
// same set code select the same row.
func TestDeckImport_PrintingDisambiguation(t *testing.T) {
	s := testAPI(t)
	seedPrintingTestCards(t, s)

	for _, setCode := range []string{"XYZ", "xyz"} {
		t.Run(setCode, func(t *testing.T) {
			text := previewCardTextLine(setCode, "99")
			preview, err := s.PreviewDeck(context.Background(), InputDeckImport{
				Text:      &text,
				SessionID: "printing-disambiguation-" + t.Name(),
			})
			if err != nil {
				t.Fatalf("PreviewDeck() error = %v", err)
			}
			if len(preview.Entries) != 1 {
				t.Fatalf("Entries = %+v, want exactly one", preview.Entries)
			}
			entry := preview.Entries[0]
			if !entry.Resolved || entry.Card == nil {
				t.Fatalf("entry not resolved: %+v", entry)
			}
			if entry.Card.ID != "testprint-0004-xyz" {
				t.Fatalf("Card.ID = %q, want %q (the XYZ printing, regardless of the set code's letter case)", entry.Card.ID, "testprint-0004-xyz")
			}
		})
	}
}

// TestDeckImport_MissingPrintingWarns proves D-10: an entry naming a set
// code absent from the snapshot still resolves (to some other printing of
// the same name), carries a warning naming the requested set code, keeps
// the parsed set code on the DeckPreviewEntry, and is never reported as
// unresolved.
func TestDeckImport_MissingPrintingWarns(t *testing.T) {
	s := testAPI(t)
	seedPrintingTestCards(t, s)

	text := previewCardTextLine("ZZZ", "42")
	preview, err := s.PreviewDeck(context.Background(), InputDeckImport{
		Text:      &text,
		SessionID: "missing-printing-" + t.Name(),
	})
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if len(preview.Entries) != 1 {
		t.Fatalf("Entries = %+v, want exactly one", preview.Entries)
	}
	entry := preview.Entries[0]

	if !entry.Resolved || entry.Card == nil {
		t.Fatalf("entry not resolved: %+v", entry)
	}
	if entry.SetCode == nil || *entry.SetCode != "ZZZ" {
		t.Fatalf("entry.SetCode = %v, want the parsed set code %q preserved", entry.SetCode, "ZZZ")
	}
	if len(preview.Unresolved) != 0 {
		t.Fatalf("Unresolved = %+v, want empty: a missing printing must not be reported as unresolved", preview.Unresolved)
	}

	found := false
	for _, w := range preview.Warnings {
		if strings.Contains(w, "ZZZ") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Warnings = %v, want one naming the requested set code %q", preview.Warnings, "ZZZ")
	}
}

// TestDeckImport_NameOnlyResolutionIsDeterministic proves that the
// name-only fallback's printing choice is the same across repeated calls,
// per nameOnlyFallbackQuery's documented lower(setcode)/number/id ordering
// — not decided by row arrival order.
func TestDeckImport_NameOnlyResolutionIsDeterministic(t *testing.T) {
	s := testAPI(t)
	seedPrintingTestCards(t, s)

	text := previewCardTextLine("", "")
	var gotIDs []string
	for i := 0; i < 3; i++ {
		preview, err := s.PreviewDeck(context.Background(), InputDeckImport{
			Text:      &text,
			SessionID: "name-only-deterministic-" + t.Name(),
		})
		if err != nil {
			t.Fatalf("PreviewDeck() call %d error = %v", i, err)
		}
		if len(preview.Entries) != 1 || preview.Entries[0].Card == nil {
			t.Fatalf("call %d: Entries = %+v, want one resolved entry", i, preview.Entries)
		}
		gotIDs = append(gotIDs, preview.Entries[0].Card.ID)
	}

	for i, id := range gotIDs {
		if id != gotIDs[0] {
			t.Fatalf("call %d resolved to Card.ID = %q, want %q (same as call 0): the choice must be deterministic", i, id, gotIDs[0])
		}
	}
	// The documented ordering key (lower(setcode) ASC) puts "ABC" ahead of
	// "XYZ", so the deterministic choice is the ABC printing.
	if gotIDs[0] != "testprint-0004-abc" {
		t.Fatalf("resolved Card.ID = %q, want %q per the documented lower(setcode) ASC ordering", gotIDs[0], "testprint-0004-abc")
	}
}

// TestDeckImport_LowConfidenceCutoffBoundary proves isLowConfidence's
// boundary condition: a score exactly at lowConfidenceCutoff is confident,
// and a score one float64 step below it is low-confidence.
func TestDeckImport_LowConfidenceCutoffBoundary(t *testing.T) {
	if isLowConfidence(lowConfidenceCutoff) {
		t.Fatalf("isLowConfidence(cutoff) = true, want false (a score exactly at the cutoff is confident)")
	}
	belowCutoff := math.Nextafter(lowConfidenceCutoff, 0)
	if !isLowConfidence(belowCutoff) {
		t.Fatalf("isLowConfidence(one step below cutoff) = false, want true")
	}
}

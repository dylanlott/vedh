package server

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

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

package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/openmtg/edh-go/pkg/telemetry"
)

// fakeEventWriter is a DB-free EventWriter used to exercise the never-fail
// write path without a live Postgres.
type fakeEventWriter struct {
	err error
}

var _ EventWriter = (*fakeEventWriter)(nil)

func (f *fakeEventWriter) Insert(ctx context.Context, e ProductEvent) error {
	return f.err
}

// TestProductEvents_WriteFailureIsNonFatal proves D-17: a product-event
// write that fails never fails, rolls back, or delays the caller's
// action — it increments vedh_product_events_dropped_total{reason=
// "write_error"} instead. This test substitutes a DB-free fake EventWriter
// through recordProductEventWith, so it needs no live database.
func TestProductEvents_WriteFailureIsNonFatal(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	before := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionWriteError))

	writer := &fakeEventWriter{err: errors.New("simulated write failure")}
	recordProductEventWith(context.Background(), logger, writer, ProductEvent{
		Name:      "deck_import_succeeded",
		SessionID: "test-session-write-failure",
	}, false)

	after := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionWriteError))
	if after != before+1 {
		t.Fatalf("vedh_product_events_dropped_total{reason=write_error} delta = %v, want 1", after-before)
	}
}

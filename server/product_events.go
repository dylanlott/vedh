package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/openmtg/edh-go/pkg/telemetry"
)

// ProductEvent mirrors the product_events table columns (see
// persistence/migrations/20260804120000_product_events.up.sql). Pointer
// fields are nullable columns; Metadata is the allowlisted key/value
// payload persisted as the metadata JSONB column.
type ProductEvent struct {
	Name       string
	SessionID  string
	UserID     *string
	GameID     *string
	Role       *string
	Source     *string
	Outcome    *string
	DurationMs *int
	Metadata   map[string]string
	// OccurredAt lets a caller pin an explicit timestamp — used by
	// TestProductEvents_OccurredAtTieOrdering to construct a
	// byte-identical tie. The zero value means "use the current time".
	OccurredAt time.Time
}

// EventWriter persists a ProductEvent. Declared as an interface — mirroring
// server/gamelog.go's EventLog and its compile-time assertion below — so a
// DB-free fake can exercise the never-fail write path in
// TestProductEvents_WriteFailureIsNonFatal without a live database.
type EventWriter interface {
	Insert(ctx context.Context, e ProductEvent) error
}

// pgEventWriter is the production EventWriter, backed by Postgres.
type pgEventWriter struct {
	db *sql.DB
}

// Ensure pgEventWriter fulfills EventWriter.
var _ EventWriter = (*pgEventWriter)(nil)

func (w *pgEventWriter) Insert(ctx context.Context, e ProductEvent) error {
	return insertProductEvent(ctx, w.db, e)
}

// insertProductEvent writes one row using the untargeted
// ON CONFLICT DO NOTHING form, which works against the partial unique
// index product_events_authoritative_once without naming it, and against
// any future unique index the same way.
func insertProductEvent(ctx context.Context, db *sql.DB, e ProductEvent) error {
	occurredAt := e.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	// A nil metadata map must marshal to the JSON empty object, never to
	// the JSON literal null: encoding/json marshals a nil map to "null",
	// which would violate product_events.metadata's NOT NULL constraint
	// semantics in spirit even though the column would technically accept
	// a JSON null value.
	metadata := e.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal product event metadata: %w", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO product_events
			(event_name, session_id, user_id, game_id, role, source, outcome, duration_ms, metadata, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10)
		ON CONFLICT DO NOTHING;`,
		e.Name, e.SessionID, e.UserID, e.GameID, e.Role, e.Source, e.Outcome, e.DurationMs, string(metadataJSON), occurredAt,
	)
	return err
}

// recordProductEvent is the never-fail entry point every emit site in this
// codebase calls, mirroring server/gamelog_helpers.go's logEvent: it guards
// on a nil server/db, validates through telemetry.ValidateEvent — which
// enforces both the closed vocabulary and the blank-session refusal — and
// on any rejection or write error increments
// vedh_product_events_dropped_total with the bounded reason and returns
// without touching the database. The caller's action always succeeds.
func (s *graphQLServer) recordProductEvent(ctx context.Context, e ProductEvent, clientSubmitted bool) {
	if s == nil || s.db == nil {
		return
	}
	recordProductEventWith(ctx, s.loggerFor(ctx), &pgEventWriter{db: s.db}, e, clientSubmitted)
}

// recordProductEventWith is recordProductEvent's write path with the writer
// and logger injected, so TestProductEvents_WriteFailureIsNonFatal can
// substitute a DB-free fake EventWriter without needing s.db to be
// non-nil.
func recordProductEventWith(ctx context.Context, logger *slog.Logger, w EventWriter, e ProductEvent, clientSubmitted bool) {
	rejection := telemetry.ValidateEvent(e.Name, e.SessionID, e.Metadata, clientSubmitted)
	if rejection != telemetry.RejectionNone {
		// Exactly one drop increment per refused event: return on the
		// first refusal rather than accumulating reasons, so the counter
		// delta is always one and an operator reading a spike knows how
		// many events it represents. Label the bounded reason, never the
		// offending event name or key.
		collectors.RecordProductEventDropped(rejection)
		logger.Warn("product event dropped", "event", e.Name, "reason", rejection.String())
		return
	}

	if err := w.Insert(ctx, e); err != nil {
		collectors.RecordProductEventDropped(telemetry.RejectionWriteError)
		logger.Warn("product event write failed", "err", err, "event", e.Name)
		return
	}

	// outcome is taken only from a server-owned event's own Outcome field
	// — never forwarded from clientSubmitted input — closing the
	// cardinality hole a client-controlled InputProductEvent.outcome
	// string would otherwise open on this label.
	outcome := "recorded"
	if !clientSubmitted && e.Outcome != nil {
		outcome = *e.Outcome
	}
	collectors.RecordProductEventWritten(outcome)
}

// TrackProductEvent is the client-facing product-event mutation. It must
// never call requireAuth — guests are the entire point of this funnel —
// and it always returns true, nil: D-19 means nothing awaits the result
// and D-22 makes the counter the only observability surface anyone reads.
// The authenticated user, if any, is attached server-side; any user
// identifier present in the input is ignored. It passes clientSubmitted:
// true, so a client attempting one of the six server-owned events is
// rejected with RejectionClientAuthoritative.
//
// publicQueries in server/authz.go is not an operation-level gate — it is
// consulted exactly once, at server/users.go:115, against a hardcoded
// literal that is not one of its members — so this resolver is public by
// virtue of not calling requireAuth, and must not be added to that map,
// where it would be dead code that reads like a security control.
func (s *graphQLServer) TrackProductEvent(ctx context.Context, input InputProductEvent) (bool, error) {
	// A limited call still returns true, nil, consistent with D-19/D-22:
	// nothing awaits trackProductEvent's result and the rate-limit
	// counter (incremented inside allowRequest, on both outcomes) is the
	// only surface anyone reads.
	clientKey := clientKeyFor(ctx, input.SessionID)
	if !s.allowRequest(ctx, ratelimit.SurfaceProductEvent, clientKey) {
		return true, nil
	}

	metadata := map[string]string{}
	for _, kv := range input.Metadata {
		if kv == nil {
			continue
		}
		metadata[kv.Key] = kv.Value
	}

	var userID *string
	if user, ok := authFromContext(ctx); ok && user != nil && user.ID != "" {
		id := user.ID
		userID = &id
	}

	s.recordProductEvent(ctx, ProductEvent{
		Name:       input.Name,
		SessionID:  input.SessionID,
		UserID:     userID,
		GameID:     input.GameID,
		Role:       input.Role,
		Source:     input.Source,
		Outcome:    input.Outcome,
		DurationMs: input.DurationMs,
		Metadata:   metadata,
	}, true)

	return true, nil
}

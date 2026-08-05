package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/openmtg/edh-go/persistence"
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

// --- shared helpers for the tests below -----------------------------------

// uniqueSessionID returns a session ID unique to the calling test, so tests
// neither see each other's rows nor depend on execution order.
func uniqueSessionID(t *testing.T) string {
	return t.Name() + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// cleanupProductEvents deletes every row for one session, following
// server/games_test.go's t.Cleanup convention.
func cleanupProductEvents(t *testing.T, db *sql.DB, sessionID string) {
	t.Helper()
	if _, err := db.Exec(`DELETE FROM product_events WHERE session_id = $1`, sessionID); err != nil {
		t.Logf("cleanup: delete product_events for session %s: %v", sessionID, err)
	}
}

// countProductEvents returns the number of rows for one (session, event
// name) pair.
func countProductEvents(t *testing.T, db *sql.DB, sessionID, eventName string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(
		`SELECT count(*) FROM product_events WHERE session_id = $1 AND event_name = $2`,
		sessionID, eventName,
	).Scan(&count); err != nil {
		t.Fatalf("count product_events: %v", err)
	}
	return count
}

// allDropReasons is every Rejection value that can appear as the reason
// label on vedh_product_events_dropped_total. RejectionNone is excluded: it
// means "accepted" and is never used as a label value.
var allDropReasons = []telemetry.Rejection{
	telemetry.RejectionUnknownEvent,
	telemetry.RejectionUnknownKey,
	telemetry.RejectionOversizedValue,
	telemetry.RejectionClientAuthoritative,
	telemetry.RejectionMissingSession,
	telemetry.RejectionWriteError,
}

// sumDroppedCounters totals vedh_product_events_dropped_total across every
// bounded reason, for tests that must prove nothing was dropped for any
// reason, not merely for one.
func sumDroppedCounters() float64 {
	total := 0.0
	for _, r := range allDropReasons {
		total += testutil.ToFloat64(collectors.ProductEventsDroppedCounter(r))
	}
	return total
}

// --- Terminology reminder for the tests below -----------------------------
//
// EventSpec.Authoritative marks the six server-owned names a client may not
// submit. The migration's dedup predicate names four *deduplicated* names
// (game_created, player_joined, guest_session_created, account_claimed).
// deck_import_succeeded and deck_import_failed are server-owned but
// deliberately outside the predicate, because a player may legitimately
// import several times. "Authoritative" and "deduplicated" are two
// different things; product_events_authoritative_once and
// TestProductEvents_AuthoritativeDedup keep the older word for continuity
// with 01-VALIDATION.md and 01-RESEARCH.md, and cover only the deduplicated
// four, never the six.

// TestProductEvents_AuthoritativeDedup proves the PostgreSQL 14 partial
// unique index product_events_authoritative_once actually deduplicates,
// rather than merely being present in the migration's SQL text.
func TestProductEvents_AuthoritativeDedup(t *testing.T) {
	s := testAPI(t)

	t.Run("both keys NULL collapse to one row", func(t *testing.T) {
		sessionID := uniqueSessionID(t)
		t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

		// This is the case the COALESCE wrappers exist for: on PostgreSQL 14
		// two NULLs are distinct in a unique index, so without them this
		// subtest is the one that fails.
		ev := ProductEvent{Name: "guest_session_created", SessionID: sessionID}
		s.recordProductEvent(context.Background(), ev, false)
		s.recordProductEvent(context.Background(), ev, false)

		if got := countProductEvents(t, s.db, sessionID, "guest_session_created"); got != 1 {
			t.Fatalf("rows = %d, want 1", got)
		}
	})

	t.Run("repeatable server event control is not deduplicated", func(t *testing.T) {
		sessionID := uniqueSessionID(t)
		t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

		// Deduplicating this event would be a bug, not a feature: a player
		// may legitimately import several times. This control goes red if
		// someone widens the predicate to "all authoritative events" by
		// conflating the two senses of "authoritative" above.
		ev := ProductEvent{Name: "deck_import_succeeded", SessionID: sessionID}
		s.recordProductEvent(context.Background(), ev, false)
		s.recordProductEvent(context.Background(), ev, false)

		if got := countProductEvents(t, s.db, sessionID, "deck_import_succeeded"); got != 2 {
			t.Fatalf("rows = %d, want 2", got)
		}
	})

	t.Run("deduplicated replay collapses without a write-error drop", func(t *testing.T) {
		sessionID := uniqueSessionID(t)
		t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

		gameID := "game-" + sessionID
		userID := "user-" + sessionID
		ev := ProductEvent{Name: "game_created", SessionID: sessionID, GameID: &gameID, UserID: &userID}

		s.recordProductEvent(context.Background(), ev, false)

		before := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionWriteError))
		s.recordProductEvent(context.Background(), ev, false)
		after := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionWriteError))

		// ON CONFLICT DO NOTHING is a successful no-op, not a write error;
		// counting it as a drop would make a healthy retry look like a
		// measurement gap.
		if after != before {
			t.Fatalf("write_error drop delta = %v, want 0", after-before)
		}
		if got := countProductEvents(t, s.db, sessionID, "game_created"); got != 1 {
			t.Fatalf("rows = %d, want 1", got)
		}
	})

	t.Run("client replay of a repeatable event and of a server-owned one", func(t *testing.T) {
		// Case 1: a repeated client-submitted board_ready — a client event
		// outside the dedup predicate — persists twice.
		boardSession := uniqueSessionID(t) + "-board"
		t.Cleanup(func() { cleanupProductEvents(t, s.db, boardSession) })
		boardInput := InputProductEvent{Name: "board_ready", SessionID: boardSession}
		if _, err := s.TrackProductEvent(context.Background(), boardInput); err != nil {
			t.Fatalf("TrackProductEvent (board_ready #1) error = %v", err)
		}
		if _, err := s.TrackProductEvent(context.Background(), boardInput); err != nil {
			t.Fatalf("TrackProductEvent (board_ready #2) error = %v", err)
		}
		if got := countProductEvents(t, s.db, boardSession, "board_ready"); got != 2 {
			t.Fatalf("board_ready rows = %d, want 2", got)
		}

		// Case 2: a client-submitted game_created with no prior server-side
		// row is refused outright — zero rows, one client_authoritative
		// drop. With no original, the refusal counts zero activations, not
		// one.
		noOriginalSession := uniqueSessionID(t) + "-no-original"
		t.Cleanup(func() { cleanupProductEvents(t, s.db, noOriginalSession) })
		before := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionClientAuthoritative))
		gameCreatedInput := InputProductEvent{Name: "game_created", SessionID: noOriginalSession}
		if _, err := s.TrackProductEvent(context.Background(), gameCreatedInput); err != nil {
			t.Fatalf("TrackProductEvent (game_created, no original) error = %v", err)
		}
		after := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionClientAuthoritative))
		if after != before+1 {
			t.Fatalf("client_authoritative drop delta = %v, want 1", after-before)
		}
		if got := countProductEvents(t, s.db, noOriginalSession, "game_created"); got != 0 {
			t.Fatalf("game_created rows = %d, want 0", got)
		}

		// Case 3: the server writes game_created first; an identical client
		// replay is refused and the server-side original is left
		// untouched — exactly one surviving row, plus one more
		// client_authoritative drop. This is the realistic production
		// shape: a client replaying an event the server already wrote, and
		// it is what makes truth #9's closing funnel clause testable
		// rather than rhetorical.
		replaySession := uniqueSessionID(t) + "-replay"
		t.Cleanup(func() { cleanupProductEvents(t, s.db, replaySession) })
		s.recordProductEvent(context.Background(), ProductEvent{Name: "game_created", SessionID: replaySession}, false)

		before = testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionClientAuthoritative))
		replayInput := InputProductEvent{Name: "game_created", SessionID: replaySession}
		if _, err := s.TrackProductEvent(context.Background(), replayInput); err != nil {
			t.Fatalf("TrackProductEvent (game_created replay) error = %v", err)
		}
		after = testutil.ToFloat64(collectors.ProductEventsDroppedCounter(telemetry.RejectionClientAuthoritative))
		if after != before+1 {
			t.Fatalf("client_authoritative drop delta = %v, want 1", after-before)
		}
		if got := countProductEvents(t, s.db, replaySession, "game_created"); got != 1 {
			t.Fatalf("game_created rows = %d, want 1 (the server-side original, unchanged)", got)
		}
	})
}

// TestProductEvents_ConcurrentInsert proves the in-flight-request case: two
// requests can reach the insert between each other's index check, and only
// the database can arbitrate — which is why the constraint lives in an
// index rather than in a read-then-write guard.
func TestProductEvents_ConcurrentInsert(t *testing.T) {
	s := testAPI(t)
	sessionID := uniqueSessionID(t)
	t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

	gameID := "game-" + sessionID
	userID := "user-" + sessionID
	ev := ProductEvent{Name: "game_created", SessionID: sessionID, GameID: &gameID, UserID: &userID}

	before := sumDroppedCounters()

	var wg sync.WaitGroup
	ready := make(chan struct{})
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-ready
			// recordProductEvent returns nothing: "the loser is silent" is
			// the assertion below, not a return value here.
			s.recordProductEvent(context.Background(), ev, false)
		}()
	}
	close(ready)
	wg.Wait()

	after := sumDroppedCounters()
	if after != before {
		t.Fatalf("dropped counter total delta = %v, want 0", after-before)
	}
	if got := countProductEvents(t, s.db, sessionID, "game_created"); got != 1 {
		t.Fatalf("rows = %d, want 1", got)
	}
}

// TestProductEvents_OccurredAtTieOrdering proves that a byte-identical
// occurred_at tie is ordered deterministically by ascending id, and that a
// min(occurred_at)-per-session funnel query cannot be changed by that tie.
// An occurred_at collision is not hypothetical: D-19's one-mutation-per-
// event delivery can land two events inside the same clock tick.
func TestProductEvents_OccurredAtTieOrdering(t *testing.T) {
	s := testAPI(t)
	sessionID := uniqueSessionID(t)
	t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })

	tie := time.Now().UTC().Truncate(time.Microsecond)
	// deck_import_started is outside the dedup predicate, so both rows
	// persist even though the tuple is otherwise identical.
	s.recordProductEvent(context.Background(), ProductEvent{
		Name: "deck_import_started", SessionID: sessionID, OccurredAt: tie,
	}, false)
	s.recordProductEvent(context.Background(), ProductEvent{
		Name: "deck_import_started", SessionID: sessionID, OccurredAt: tie,
	}, false)

	queryOrder := func() []int64 {
		rows, err := s.db.Query(
			`SELECT id FROM product_events WHERE session_id = $1 ORDER BY occurred_at, id`, sessionID)
		if err != nil {
			t.Fatalf("query ordered ids: %v", err)
		}
		defer rows.Close()
		var ids []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatalf("scan id: %v", err)
			}
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate ordered ids: %v", err)
		}
		return ids
	}

	first := queryOrder()
	if len(first) != 2 {
		t.Fatalf("rows = %d, want 2", len(first))
	}
	if first[0] >= first[1] {
		t.Fatalf("ids not ascending for a tied occurred_at: %v", first)
	}

	second := queryOrder()
	if first[0] != second[0] || first[1] != second[1] {
		t.Fatalf("order not stable across repeated queries: %v then %v", first, second)
	}

	var sessionRowCount int
	if err := s.db.QueryRow(
		`SELECT count(*) FROM (
			SELECT session_id, min(occurred_at) FROM product_events WHERE session_id = $1 GROUP BY session_id
		) t`,
		sessionID,
	).Scan(&sessionRowCount); err != nil {
		t.Fatalf("query min(occurred_at) per session: %v", err)
	}
	if sessionRowCount != 1 {
		t.Fatalf("min(occurred_at)-per-session rows = %d, want 1 (a tie must not change a funnel result)", sessionRowCount)
	}
}

// TestProductEvents_Allowlist is the behavioural half of the privacy rule,
// distinct from pkg/telemetry's pure ValidateEvent unit test: it proves the
// refusal reaches neither the table nor the wrong counter.
func TestProductEvents_Allowlist(t *testing.T) {
	s := testAPI(t)

	cases := []struct {
		name            string
		event           ProductEvent
		clientSubmitted bool
		wantReason      telemetry.Rejection
	}{
		{
			name:       "unknown event name",
			event:      ProductEvent{Name: "not_a_real_event"},
			wantReason: telemetry.RejectionUnknownEvent,
		},
		{
			name: "key absent from this event's own allowlist",
			event: ProductEvent{
				Name:     "quick_start_viewed",
				Metadata: map[string]string{"share_method": "clipboard"},
			},
			wantReason: telemetry.RejectionUnknownKey,
		},
		{
			name: "metadata value one byte over the limit",
			event: ProductEvent{
				Name:     "landing_primary_cta",
				Metadata: map[string]string{"utm_source": strings.Repeat("a", telemetry.MaxMetadataValueBytes+1)},
			},
			wantReason: telemetry.RejectionOversizedValue,
		},
		{
			name:            "client submission of a server-owned event",
			event:           ProductEvent{Name: "game_created"},
			clientSubmitted: true,
			wantReason:      telemetry.RejectionClientAuthoritative,
		},
		{
			name:       "whitespace-only session ID",
			event:      ProductEvent{Name: "quick_start_viewed", SessionID: "   "},
			wantReason: telemetry.RejectionMissingSession,
		},
		{
			name:       "accepted: known event with empty metadata",
			event:      ProductEvent{Name: "quick_start_viewed"},
			wantReason: telemetry.RejectionNone,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sessionID := tc.event.SessionID
			// Every case except the missing-session one needs a real,
			// non-blank session ID unique to this subtest — otherwise the
			// session check in ValidateEvent would fire first and mask the
			// case this subtest means to exercise.
			if tc.wantReason != telemetry.RejectionMissingSession {
				sessionID = uniqueSessionID(t)
				tc.event.SessionID = sessionID
				t.Cleanup(func() { cleanupProductEvents(t, s.db, sessionID) })
			}

			before := map[telemetry.Rejection]float64{}
			for _, r := range allDropReasons {
				before[r] = testutil.ToFloat64(collectors.ProductEventsDroppedCounter(r))
			}

			s.recordProductEvent(context.Background(), tc.event, tc.clientSubmitted)

			for _, r := range allDropReasons {
				after := testutil.ToFloat64(collectors.ProductEventsDroppedCounter(r))
				wantDelta := 0.0
				if r == tc.wantReason {
					wantDelta = 1.0
				}
				if after-before[r] != wantDelta {
					t.Fatalf("dropped[%s] delta = %v, want %v", r.String(), after-before[r], wantDelta)
				}
			}

			if tc.wantReason != telemetry.RejectionNone {
				if got := countProductEvents(t, s.db, sessionID, tc.event.Name); got != 0 {
					t.Fatalf("rows = %d, want 0 (a refused event must write nothing)", got)
				}
				return
			}

			var metadataJSON string
			if err := s.db.QueryRow(
				`SELECT metadata::text FROM product_events WHERE session_id = $1 AND event_name = $2`,
				sessionID, tc.event.Name,
			).Scan(&metadataJSON); err != nil {
				t.Fatalf("query metadata: %v", err)
			}
			if metadataJSON != "{}" {
				t.Fatalf("metadata = %q, want the empty JSON object", metadataJSON)
			}
		})
	}
}

// withScratchMigrationDB creates a scratch database derived from
// DATABASE_URL, hands its URL to fn, and drops the database in t.Cleanup —
// tolerating an already-dropped database so a failed run does not leave the
// cleanup itself failing. It is a named, reusable, within-package helper
// rather than inline mechanics: plan 01-04 task 1's
// TestMigrations_CardNameSearch is written to reuse this exact helper by
// this exact name for its own migration pair.
//
// One scratch database per call, not one shared across the up/down/up
// cycle of two directories: golang-migrate's version table is per-database
// and the two migration directories carry overlapping versions, so sharing
// would corrupt the second run's state.
func withScratchMigrationDB(t *testing.T, migrationsDir string, fn func(t *testing.T, scratchURL string)) {
	t.Helper()

	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		baseURL = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}

	adminDB, err := sql.Open("postgres", baseURL)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}

	scratchName := fmt.Sprintf("edhgo_mig_%d", time.Now().UnixNano())
	// CREATE DATABASE cannot run inside a transaction; ExecContext against
	// a *sql.DB issues this as one standalone statement, never wrapped in
	// an implicit transaction.
	if _, err := adminDB.ExecContext(context.Background(),
		fmt.Sprintf("CREATE DATABASE %s;", pq.QuoteIdentifier(scratchName))); err != nil {
		adminDB.Close()
		t.Fatalf("create scratch database %s: %v", scratchName, err)
	}

	t.Cleanup(func() {
		defer adminDB.Close()
		// A migration step run against the scratch database (NewPostgres,
		// MigrateDown) can leave its connection's backend lingering just
		// long enough to make an immediate DROP DATABASE race against
		// Postgres reclaiming it, even after the caller has closed its
		// *sql.DB. Terminate any remaining backend on the scratch database
		// first so the drop is not racing a connection this same test
		// opened and closed.
		if _, err := adminDB.ExecContext(context.Background(),
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid();`,
			scratchName,
		); err != nil {
			t.Logf("cleanup: terminate backends on scratch database %s: %v", scratchName, err)
		}

		var dropErr error
		for attempt := 0; attempt < 3; attempt++ {
			_, dropErr = adminDB.ExecContext(context.Background(),
				fmt.Sprintf("DROP DATABASE IF EXISTS %s;", pq.QuoteIdentifier(scratchName)))
			if dropErr == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if dropErr != nil {
			t.Logf("cleanup: drop scratch database %s: %v", scratchName, dropErr)
		}
	})

	scratchURL := *parsed
	scratchURL.Path = "/" + scratchName
	fn(t, scratchURL.String())
}

// assertProductEventsObjectsPresent asserts that the product_events table
// and all three of its indexes are present (present == true) or absent
// (present == false) in the database at dbURL.
func assertProductEventsObjectsPresent(t *testing.T, dbURL string, present bool) {
	t.Helper()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open scratch db: %v", err)
	}
	defer db.Close()

	wantCount := 0
	if present {
		wantCount = 1
	}

	var tableCount int
	if err := db.QueryRow(
		`SELECT count(*) FROM pg_tables WHERE tablename = 'product_events'`,
	).Scan(&tableCount); err != nil {
		t.Fatalf("query pg_tables: %v", err)
	}
	if tableCount != wantCount {
		t.Fatalf("product_events table present = %v, want %v", tableCount == 1, present)
	}

	for _, idx := range []string{
		"product_events_name_time_idx",
		"product_events_session_time_idx",
		"product_events_authoritative_once",
	} {
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

// TestMigrations_ProductEvents is the up/down/up proof the migration-parity
// constraint requires. It runs entirely against scratch databases created
// and dropped by withScratchMigrationDB, and never calls MigrateDown
// against the URL TestMain migrated — MigrateDown there would drop the
// seeded cards and the imported MTGJSON snapshot out from under every other
// test in the package.
func TestMigrations_ProductEvents(t *testing.T) {
	for _, dir := range []string{"../persistence/migrations/", "../persistence/migrations_test/"} {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			withScratchMigrationDB(t, dir, func(t *testing.T, scratchURL string) {
				// NewPostgres returns an open *sql.DB that its caller
				// normally keeps using; here each step's connection must
				// be closed immediately, or the pool it leaves open blocks
				// withScratchMigrationDB's cleanup from dropping the
				// scratch database.
				upDB, err := persistence.NewPostgres(dir, scratchURL)
				if err != nil {
					t.Fatalf("up: %v", err)
				}
				if err := upDB.Close(); err != nil {
					t.Fatalf("close db after up: %v", err)
				}
				assertProductEventsObjectsPresent(t, scratchURL, true)

				if err := persistence.MigrateDown(dir, scratchURL); err != nil {
					t.Fatalf("down: %v", err)
				}
				assertProductEventsObjectsPresent(t, scratchURL, false)

				upAgainDB, err := persistence.NewPostgres(dir, scratchURL)
				if err != nil {
					t.Fatalf("up (again): %v", err)
				}
				if err := upAgainDB.Close(); err != nil {
					t.Fatalf("close db after second up: %v", err)
				}
				assertProductEventsObjectsPresent(t, scratchURL, true)
			})
		})
	}
}

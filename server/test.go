package server

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"testing"
)

// testAPIDefaultDSN is a test-only local fallback used when DATABASE_URL is
// unset. Production startup requires DATABASE_URL explicitly.
const testAPIDefaultDSN = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"

func testAPI(t *testing.T) *graphQLServer {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}

	// WR-02 fix (code review, phase 01): read DATABASE_URL with a
	// documented, safe local fallback so an environment whose test
	// Postgres instance lives at a different host/port/credentials/DB
	// name is not silently unreachable.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = testAPIDefaultDSN
	}

	cfg := Conf{
		PostgresURL: dsn,
		DefaultPort: 8080,
		// fix(01-06): testAPI builds Conf as a literal rather than through
		// envconfig.Process (main.go's production path), so the `default`
		// tags on DeckImportRatePerMinute/DeckImportRateBurst are never
		// applied here -- an unset field would be Go's zero value, which
		// would make NewGraphQLServer construct a burst-0 registry that
		// denies every single PreviewDeck/TrackProductEvent call this
		// package's many existing tests make. Set high enough that no
		// existing test's call volume ever approaches it; only
		// server/ratelimit_test.go constructs its own small registry
		// directly to exercise the limited path.
		DeckImportRatePerMinute:   100000,
		DeckImportRateBurst:       100000,
		GuestSessionRatePerMinute: 100000,
		GuestSessionRateBurst:     100000,
		// GuestCreationEnabled mirrors the same fix(01-06) reasoning above:
		// this Conf literal bypasses envconfig.Process, so
		// GuestCreationEnabled's `default:"true"` tag (server/graphql.go)
		// never applies here -- an unset field would be Go's zero value
		// (false), which would make every guest-creation test in this
		// package fail the kill-switch check by default. Set true so this
		// package's tests exercise the production default; only
		// server/guest_users_test.go's TestGuestUsers_KillSwitch overrides
		// it back to false to exercise the switch itself.
		GuestCreationEnabled: true,
	}
	// TestMain has already migrated and seeded this run's scratch database.
	// Opening it directly here avoids re-running the migration machinery for
	// every test case.
	appDB, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		t.Fatalf("postgres is reachable but opening the connection pool failed (likely a real regression, not an absent dependency): %s", err)
	}
	if err := appDB.Ping(); err != nil {
		_ = appDB.Close()
		t.Fatalf("ping scratch test database: %s", err)
	}
	// The server suite historically reused one database without clearing rows,
	// so fixed game IDs and usernames leaked between otherwise independent
	// tests. Keep the immutable card fixture, but reset application-owned rows
	// before constructing each test server. No server tests run in parallel.
	if _, err := appDB.Exec(`
		TRUNCATE TABLE product_events, gamelog, games, users
		RESTART IDENTITY CASCADE
	`); err != nil {
		_ = appDB.Close()
		t.Fatalf("reset scratch test data: %s", err)
	}
	// fix(01-05): every testAPI call opens a brand-new *sql.DB connection
	// pool that nothing ever closed. That was survivable while the
	// package had few enough Test* functions to stay under Postgres's
	// max_connections (100, unchanged from this container's default),
	// but this plan's own added tests were the ones that tipped a full
	// `go test ./server` run over that ceiling ("pq: sorry, too many
	// clients already") -- confirmed by running the identical suite
	// against the pre-01-05 tree, where it does not occur. Closing here
	// is safe: this package's test logger discards all output (line 17
	// above), so a background goroutine (e.g. go s.publishGame(...))
	// that outlives its owning test and touches a closed pool afterward
	// logs into the void rather than panicking or failing a completed
	// test.
	t.Cleanup(func() {
		_ = appDB.Close()
	})
	s, err := NewGraphQLServer(appDB, cfg, logger)
	if err != nil {
		t.Errorf("failed to create new test server: %+v", err)
	}
	return s
}

func authCtx(username string) context.Context {
	return withAuth(context.Background(), &AuthUser{
		ID:       username,
		Username: username,
	})
}

func authCtxWithID(id string, username string) context.Context {
	return withAuth(context.Background(), &AuthUser{
		ID:       id,
		Username: username,
	})
}

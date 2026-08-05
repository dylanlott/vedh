package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/openmtg/edh-go/persistence"
)

// testAPIDefaultDSN is the local fallback DSN used when DATABASE_URL is
// unset, mirroring production's own envconfig default for PostgresURL
// (server/graphql.go's Conf) so a differently-configured environment is
// not silently unreachable.
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
		DeckImportRatePerMinute: 100000,
		DeckImportRateBurst:     100000,
	}
	// WR-02 fix (code review, phase 01): only "Postgres is not reachable at
	// all" (a plain TCP dial to the configured host:port fails) is treated
	// as a legitimate environment-dependent skip -- this package's own
	// tests are documented as NOT part of the CI-gated target
	// (.github/workflows/test.yml's "Test" job runs only `make test-unit`,
	// i.e. `go test ./pkg/... -race`; `make test-api` / `go test
	// ./server/...` requires a local Postgres and is a deliberately
	// separate, not-CI-gated target per README.md's "Backend integration
	// tests" section), so a developer machine with no Postgres running at
	// all is a supported, unremarkable state.
	//
	// Once that reachability check has passed, Postgres IS present and
	// listening -- a subsequent failure to reset migrations or open a
	// connection pool indicates something is actually broken (bad
	// credentials, a corrupted migration set, a permissions problem), not
	// an absent dependency, and silently downgrading that to a skip is
	// exactly the failure mode this fix closes: it previously let three
	// real regressions during this phase pass as skipped rather than
	// failed. t.Fatalf stops this test immediately with a failure so it
	// cannot be mistaken for "environment not set up".
	if err := ensurePostgresReachable(cfg.PostgresURL); err != nil {
		t.Skipf("postgres unavailable: %s", err)
	}
	if err := persistence.ForceCleanMigrations("../persistence/migrations_test/", cfg.PostgresURL); err != nil {
		t.Fatalf("postgres is reachable but resetting test migrations failed (likely a real regression, not an absent dependency): %s", err)
	}
	appDB, err := persistence.NewPostgres("../persistence/migrations_test/", cfg.PostgresURL)
	if err != nil {
		t.Fatalf("postgres is reachable but opening the connection pool failed (likely a real regression, not an absent dependency): %s", err)
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

func ensurePostgresReachable(dbURL string) error {
	parsed, err := url.Parse(dbURL)
	if err != nil {
		return err
	}
	host := parsed.Hostname()
	if host == "" {
		host = "localhost"
	}
	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
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

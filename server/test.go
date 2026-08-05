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

func testAPI(t *testing.T) *graphQLServer {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}

	cfg := Conf{
		PostgresURL: "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3",
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
	if err := ensurePostgresReachable(cfg.PostgresURL); err != nil {
		t.Skipf("postgres unavailable: %s", err)
	}
	if err := persistence.ForceCleanMigrations("../persistence/migrations_test/", cfg.PostgresURL); err != nil {
		t.Skipf("failed to reset test migrations: %s", err)
	}
	appDB, err := persistence.NewPostgres("../persistence/migrations_test/", cfg.PostgresURL)
	if err != nil {
		t.Skipf("postgres unavailable: %s", err)
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

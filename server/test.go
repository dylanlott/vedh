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
	}
	if err := ensurePostgresReachable(cfg.PostgresURL); err != nil {
		t.Skipf("postgres unavailable: %s", err)
	}
	appDB, err := persistence.NewPostgres("../persistence/migrations_test/", cfg.PostgresURL)
	if err != nil {
		t.Skipf("postgres unavailable: %s", err)
	}
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

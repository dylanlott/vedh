package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/openmtg/edh-go/persistence"
)

func TestMain(m *testing.M) {
	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		baseURL = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"
	}

	dbURL, cleanup, err := createServerTestDatabase(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
		_ = cleanup()
		fmt.Fprintf(os.Stderr, "test setup failed: set scratch DATABASE_URL: %v\n", err)
		os.Exit(1)
	}

	appDB, err := persistence.NewPostgres("../persistence/migrations_test/", dbURL)
	if err != nil {
		_ = cleanup()
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := appDB.Close(); err != nil {
		_ = cleanup()
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := seedServerTestCards(dbURL); err != nil {
		_ = cleanup()
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if code == 0 {
		fmt.Printf("all api tests passed ✅\n")
	}
	if err := cleanup(); err != nil {
		fmt.Fprintf(os.Stderr, "test cleanup failed: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func createServerTestDatabase(baseURL string) (string, func() error, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return "", nil, fmt.Errorf("DATABASE_URL must use postgres or postgresql")
	}
	if parsed.Hostname() == "" || parsed.Path == "" || parsed.Path == "/" {
		return "", nil, fmt.Errorf("DATABASE_URL must include a host and administration database")
	}

	adminDB, err := sql.Open("postgres", baseURL)
	if err != nil {
		return "", nil, fmt.Errorf("open administration database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		adminDB.Close()
		return "", nil, fmt.Errorf("connect to administration database: %w", err)
	}

	scratchName := fmt.Sprintf("edhgo_test_%d_%d", os.Getpid(), time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(scratchName)); err != nil {
		adminDB.Close()
		return "", nil, fmt.Errorf("create scratch database %s: %w", scratchName, err)
	}

	scratchURL := *parsed
	scratchURL.Path = "/" + scratchName

	cleanup := func() error {
		defer adminDB.Close()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		if _, err := adminDB.ExecContext(cleanupCtx,
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`,
			scratchName,
		); err != nil {
			return fmt.Errorf("terminate scratch database connections: %w", err)
		}
		if _, err := adminDB.ExecContext(cleanupCtx, "DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(scratchName)); err != nil {
			return fmt.Errorf("drop scratch database: %w", err)
		}
		return nil
	}

	return scratchURL.String(), cleanup, nil
}

func seedServerTestCards(dbURL string) error {
	rawFixture, err := os.ReadFile("testdata/deck_providers/archidekt_deck_2026-08-05.json")
	if err != nil {
		return fmt.Errorf("read committed Archidekt fixture: %w", err)
	}
	var fixture archidektDeckResponse
	if err := json.Unmarshal(rawFixture, &fixture); err != nil {
		return fmt.Errorf("decode committed Archidekt fixture: %w", err)
	}

	names := map[string]struct{}{
		"Gavi, Nest Warden": {},
	}
	for _, row := range fixture.Cards {
		if row.Card.OracleCard.Name != "" {
			names[row.Card.OracleCard.Name] = struct{}{}
		}
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("open scratch database for fixture seed: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin fixture seed: %w", err)
	}
	defer tx.Rollback()
	for name := range names {
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
		id := "testfixture-" + digest[:32]
		if _, err := tx.Exec(
			`INSERT INTO cards (id, name, uuid) VALUES ($1, $2, $1) ON CONFLICT DO NOTHING`,
			id, name,
		); err != nil {
			return fmt.Errorf("seed card %q: %w", name, err)
		}
		if _, err := tx.Exec(
			`INSERT INTO card_names (name_lower, display) VALUES (lower($1), $1) ON CONFLICT DO NOTHING`,
			name,
		); err != nil {
			return fmt.Errorf("seed card name %q: %w", name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit fixture seed: %w", err)
	}
	return nil
}

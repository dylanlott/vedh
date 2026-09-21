package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/openmtg/edh-go/persistence"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}
	db, err := persistence.NewPostgres("persistence/migrations_test", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare test database: %v\n", err)
		os.Exit(1)
	}
	if err := seedActivationCards(context.Background(), db); err != nil {
		fmt.Fprintf(os.Stderr, "seed activation cards: %v\n", err)
		_ = db.Close()
		os.Exit(1)
	}
	if err := db.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close test database: %v\n", err)
		os.Exit(1)
	}
}

func seedActivationCards(ctx context.Context, db *sql.DB) error {
	fixtures := []struct {
		id   string
		name string
	}{
		{"e2e-kykar", "Kykar, Wind's Fury"},
		{"e2e-island", "Island"},
		{"e2e-jarad", "Jarad, Golgari Lich Lord"},
		{"e2e-swamp", "Swamp"},
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, fixture := range fixtures {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO cards (id, name, uuid)
			VALUES ($1, $2, $1)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		`, fixture.id, fixture.name); err != nil {
			return fmt.Errorf("insert %q: %w", fixture.name, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO card_names (name_lower, display)
			VALUES (lower($1), $1)
			ON CONFLICT (name_lower) DO UPDATE SET display = EXCLUDED.display
		`, fixture.name); err != nil {
			return fmt.Errorf("project %q: %w", fixture.name, err)
		}
	}
	return tx.Commit()
}

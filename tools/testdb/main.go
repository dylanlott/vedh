package main

import (
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
	if err := db.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close test database: %v\n", err)
		os.Exit(1)
	}
}

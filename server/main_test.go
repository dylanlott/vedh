package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/openmtg/edh-go/persistence"
)

func TestMain(m *testing.M) {
	code := m.Run()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"
	}
	if err := persistence.MigrateDown("../persistence/migrations_test/", dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "test cleanup failed: %v\n", err)
		if code == 0 {
			code = 1
		}
	}

	os.Exit(code)
}

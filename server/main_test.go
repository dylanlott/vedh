package server

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/openmtg/edh-go/persistence"
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"
	}
	if err := persistence.ForceCleanMigrations("../persistence/migrations_test/", dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := persistence.MigrateDown("../persistence/migrations_test/", dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	appDB, err := persistence.NewPostgres("../persistence/migrations_test/", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := appDB.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}
	if err := importAllPrintingsForTests(dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "test setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if code == 0 {
		fmt.Printf("all api tests passed ✅")
	}

	// if err := persistence.MigrateDown("../persistence/migrations_test/", dbURL); err != nil {
	// 	fmt.Fprintf(os.Stderr, "test cleanup failed: %v\n", err)
	// 	if code == 0 {
	// 		code = 1
	// 	}
	// }

	os.Exit(code)
}

func importAllPrintingsForTests(dbURL string) error {
	jsonPath := os.Getenv("ALL_PRINTINGS_JSON_PATH")
	if jsonPath == "" {
		jsonPath = "../All Printings.json"
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err := persistence.ImportAllPrintingsJSON(dbURL, persistence.AllPrintingsImportOptions{
		JSONPath: jsonPath,
		Logger:   logger,
	})
	if err != nil {
		return fmt.Errorf("import all printings json: %w", err)
	}
	return nil
}

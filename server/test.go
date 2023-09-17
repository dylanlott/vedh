package server

import (
	"testing"

	"github.com/openmtg/edh-go/persistence"
)

func testAPI(t *testing.T) *graphQLServer {
	cfg := Conf{
		PostgresURL: "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable",
		DefaultPort: 8080,
	}
	appDB, err := persistence.NewPostgres("../persistence/migrations/", cfg.PostgresURL)
	if err != nil {
		t.Errorf("failed to get migrated app instance: %s", err)
	}
	s, err := NewGraphQLServer(appDB, cfg)
	if err != nil {
		t.Errorf("failed to create new test server: %+v", err)
	}
	return s
}

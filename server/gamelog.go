package server

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Ensure that pgLogger fulfills EventLog
var _ EventLog = (*pgLogger)(nil)

// ErrEmpty is returned if an Event's payload is nil.
var ErrEmpty = fmt.Errorf("must provide an event payload")

// pgLogger logs Events into Postgres.
type pgLogger struct {
	db *sql.DB
}

// EventLog declares an interface for an append-only Event log.
type EventLog interface {
	Add(ctx context.Context, event Event) error
}

// Event represents a change to boardstate
type Event struct {
	Payload map[string]interface{}
}

// Make the Event struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (e Event) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan fulfills the sql.Scanner interface. This method decodes a JSON encoded
// value into the struct fields.
func (e *Event) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("gamelog failed to scan: type assertion failed: %T", b)
	}
	return json.Unmarshal(b, &e)
}

// Add adds an event to the game log. It is an append only action and has
// no concept of the previous or next Event that will be processed.
// * Add can be called asynchronously.
// * We always maintain a server-centric view of events in the gamelog.
func (g *pgLogger) Add(ctx context.Context, event Event) error {
	if event.Payload == nil {
		return ErrEmpty
	}
	query := `INSERT INTO gamelog (payload) VALUES($1);`
	_, err := g.db.Exec(query, event)
	if err != nil {
		return fmt.Errorf("failed to add event to gamelog: %w", err)
	}
	return nil
}

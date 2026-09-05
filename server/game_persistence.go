package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// mutateGame serializes changes to one persisted game across every server
// process.  Reading the JSON payload and later writing a replacement payload
// must happen under the same row lock; otherwise concurrent board, turn, and
// join mutations silently discard one another's work.
func (s *graphQLServer) mutateGame(ctx context.Context, gameID string, mutate func(context.Context, *Game) (*Game, error)) (*Game, error) {
	if gameID == "" {
		return nil, errors.New("game ID is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin game mutation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	current, err := loadGameByIDTx(ctx, tx, gameID)
	if err != nil {
		return nil, err
	}
	collector := &mutationEventCollector{}
	next, err := mutate(withMutationEventCollector(ctx, collector), current)
	if err != nil {
		return nil, err
	}
	if next == nil || next.ID != gameID {
		return nil, errors.New("game mutation returned an invalid game")
	}
	if err := updateGameTx(ctx, tx, next); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit game mutation: %w", err)
	}
	// Emit events only after the game row has committed. If the mutation or
	// commit fails, the collector is discarded and no event can describe state
	// that was never persisted. Event-write failures remain non-fatal to the
	// already-committed game, matching logEvent's best-effort contract.
	for _, event := range collector.events {
		s.logEvent(ctx, event)
	}
	return next, nil
}

// createOrLoadGame creates candidate exactly once.  A retry by an existing
// participant returns the already-persisted game, while a caller who was not
// already in it is never allowed to replace its payload.
func (s *graphQLServer) createOrLoadGame(ctx context.Context, candidate *Game, caller *AuthUser) (*Game, bool, error) {
	if candidate == nil || candidate.ID == "" {
		return nil, false, errors.New("game ID is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, fmt.Errorf("begin game creation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, err := loadGameByIDTx(ctx, tx, candidate.ID)
	if err == nil {
		if !isUserInGame(existing, caller) {
			return nil, false, errors.New("forbidden: game ID already belongs to another game")
		}
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit existing game read: %w", err)
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	payload, err := json.Marshal(candidate)
	if err != nil {
		return nil, false, fmt.Errorf("marshal game: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO games (id, payload)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (id) DO NOTHING`, candidate.ID, string(payload))
	if err != nil {
		return nil, false, fmt.Errorf("insert game: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("inspect game insert: %w", err)
	}
	if inserted == 1 {
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit game creation: %w", err)
		}
		return candidate, true, nil
	}

	// A concurrent create won the unique-ID race. At READ COMMITTED the next
	// query sees that committed row, locks it, and applies the same safe retry
	// rule as the existing-row branch above.
	existing, err = loadGameByIDTx(ctx, tx, candidate.ID)
	if err != nil {
		return nil, false, err
	}
	if !isUserInGame(existing, caller) {
		return nil, false, errors.New("forbidden: game ID already belongs to another game")
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit concurrent game read: %w", err)
	}
	return existing, false, nil
}

func loadGameByIDTx(ctx context.Context, tx *sql.Tx, gameID string) (*Game, error) {
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM games WHERE id = $1 FOR UPDATE`, gameID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	game := &Game{}
	if err := json.Unmarshal(payload, game); err != nil {
		return nil, fmt.Errorf("decode game %s: %w", gameID, err)
	}
	ensureGameDefaults(game)
	return game, nil
}

func updateGameTx(ctx context.Context, tx *sql.Tx, game *Game) error {
	payload, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("marshal game: %w", err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE games SET payload = $2::jsonb WHERE id = $1`, game.ID, string(payload))
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("game %s disappeared during mutation", game.ID)
	}
	return nil
}

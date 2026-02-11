package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type rawLogPayload struct {
	Type    string          `json:"type"`
	Actor   *string         `json:"actor,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func (s *graphQLServer) GameLogs(ctx context.Context, gameID string, offset int, limit int) ([]*GameLogEvent, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if !isUserInGame(game, authUser) {
		return nil, fmt.Errorf("forbidden: not a participant in this game")
	}
	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(
		`SELECT id, game_id, eventtime, payload
		 FROM gamelog
		 WHERE game_id = $1
		 ORDER BY eventtime ASC, id ASC
		 LIMIT $2 OFFSET $3`,
		gameID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query gamelog: %w", err)
	}
	defer rows.Close()

	var events []*GameLogEvent
	for rows.Next() {
		var (
			id       int64
			gid      string
			ts       sql.NullTime
			payloadB []byte
		)
		if err := rows.Scan(&id, &gid, &ts, &payloadB); err != nil {
			return nil, fmt.Errorf("failed to scan gamelog: %w", err)
		}
		decoded := rawLogPayload{}
		if len(payloadB) > 0 {
			if err := json.Unmarshal(payloadB, &decoded); err != nil {
				return nil, fmt.Errorf("failed to parse gamelog payload: %w", err)
			}
		}
		payloadStr := ""
		if len(decoded.Payload) > 0 {
			payloadStr = string(decoded.Payload)
		}
		eventTime := ts.Time
		if !ts.Valid {
			eventTime = sql.NullTime{}.Time
		}
		event := &GameLogEvent{
			ID:        fmt.Sprintf("%d", id),
			GameID:    gid,
			EventTime: eventTime,
			Type:      decoded.Type,
			Actor:     decoded.Actor,
		}
		if payloadStr != "" {
			event.Payload = &payloadStr
		}
		if event.Type == "" {
			event.Type = "UNKNOWN"
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate gamelog rows: %w", err)
	}
	return events, nil
}

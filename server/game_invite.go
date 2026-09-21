package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/openmtg/edh-go/pkg/ratelimit"
)

const gameInviteCapacity = 4

// GameInvite returns the deliberately minimal public view of a table. Keep
// this projection separate from Game: adding a field to Game must never widen
// the anonymous invite surface by accident.
func (s *graphQLServer) GameInvite(ctx context.Context, gameID string, sessionID string) (*GameInvite, error) {
	gameID = strings.TrimSpace(gameID)
	sessionID = strings.TrimSpace(sessionID)
	if gameID == "" || sessionID == "" {
		return nil, errors.New("invite link is incomplete")
	}
	if !s.allowRequest(ctx, ratelimit.SurfaceInviteLookup, clientKeyFor(ctx, sessionID)) {
		return nil, errors.New("too many invite checks; try again shortly")
	}

	game, err := s.loadGameByID(gameID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		s.loggerFor(ctx).Warn("public invite lookup failed", "err", err)
		return nil, errors.New("this invite cannot be checked right now")
	}
	ensureGameDefaults(game)

	displayNames := make([]string, 0, len(game.Players))
	for _, player := range game.Players {
		if player == nil {
			continue
		}
		name := player.Username
		if player.DisplayName != nil && strings.TrimSpace(*player.DisplayName) != "" {
			name = strings.TrimSpace(*player.DisplayName)
		}
		displayNames = append(displayNames, name)
	}

	return &GameInvite{
		ID:                 game.ID,
		Format:             formatFromRules(game.Rules).ID,
		Status:             game.Status,
		PlayerDisplayNames: displayNames,
		PlayerCount:        len(game.Players),
		Capacity:           gameInviteCapacity,
		CreatedAt:          game.CreatedAt,
	}, nil
}

package server

import (
	"context"
	"errors"
)

func (s *graphQLServer) ClaimWin(ctx context.Context, gameID string, condition *string) (*Game, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	ensureGameDefaults(game)
	if game.Status == GameStatusFinished {
		return nil, errors.New("game already finished")
	}
	if !isUserInGame(game, authUser) {
		return nil, errors.New("forbidden: not a participant in this game")
	}
	if game.Turn == nil {
		return nil, errors.New("game has no turn state")
	}
	if game.Turn.Priority != authUser.Username {
		return nil, errors.New("forbidden: only priority player can claim a win")
	}

	sequence := buildClaimSequence(game, authUser.Username)
	if len(sequence) == 0 {
		return nil, errors.New("no eligible players for win claim")
	}
	game.PendingWinClaim = &PendingWinClaim{
		ClaimedBy: authUser.Username,
		Condition: condition,
		Remaining: sequence,
	}

	s.logEvent(ctx, Event{
		GameID: game.ID,
		Type:   EventTypeWinClaimed,
		Actor:  authUser.Username,
		Payload: map[string]interface{}{
			"claimedBy": authUser.Username,
			"condition": condition,
			"remaining": sequence,
		},
	})

	go s.publishGame(game.ID, game)
	if err := s.upsertGame(game); err != nil {
		return game, err
	}
	return game, nil
}

package server

import (
	"context"
	"strings"
)

func alivePlayerNames(game *Game) []string {
	if game == nil {
		return nil
	}
	var alive []string
	for _, p := range game.Players {
		if p == nil {
			continue
		}
		if p.Boardstate == nil {
			alive = append(alive, p.Username)
			continue
		}
		if p.Boardstate.Life > 0 {
			alive = append(alive, p.Username)
		}
	}
	return alive
}

func winnerIDsForUsers(game *Game, users []string) []string {
	var winners []string
	for _, name := range users {
		for _, p := range game.Players {
			if p == nil {
				continue
			}
			if p.Username == name {
				if p.ID != "" {
					winners = append(winners, p.ID)
				} else {
					winners = append(winners, name)
				}
				break
			}
		}
	}
	return winners
}

func ensureGameDefaults(game *Game) {
	if game == nil {
		return
	}
	if game.Status == "" {
		game.Status = GameStatusInProgress
	}
	ensureFormatRules(game, formatFromRules(game.Rules))
}

func finalizeGame(game *Game, result GameResult, winnerNames []string, condition *string) {
	if game == nil {
		return
	}
	game.Status = GameStatusFinished
	game.Result = &result
	game.WinnerIDs = winnerIDsForUsers(game, winnerNames)
	game.WinCondition = condition
	game.PendingWinClaim = nil
}

func buildClaimSequence(game *Game, claimer string) []string {
	if game == nil {
		return nil
	}
	alive := alivePlayerNames(game)
	if len(alive) == 0 {
		return nil
	}
	claimerIdx := -1
	for i, name := range alive {
		if name == claimer {
			claimerIdx = i
			break
		}
	}
	if claimerIdx == -1 {
		return nil
	}
	var sequence []string
	for i := 1; i <= len(alive); i++ {
		sequence = append(sequence, alive[(claimerIdx+i)%len(alive)])
	}
	return sequence
}

func claimMatchesPrioritySequence(claim *PendingWinClaim, toPlayer string) bool {
	if claim == nil || len(claim.Remaining) == 0 {
		return false
	}
	return strings.EqualFold(claim.Remaining[0], toPlayer)
}

func (s *graphQLServer) cancelPendingWinClaim(ctx context.Context, game *Game, actor string, reason string) {
	if game == nil || game.PendingWinClaim == nil {
		return
	}
	claim := game.PendingWinClaim
	game.PendingWinClaim = nil
	s.logEvent(ctx, Event{
		GameID: game.ID,
		Type:   EventTypeWinClaimCancelled,
		Actor:  actor,
		Payload: map[string]interface{}{
			"claimedBy": claim.ClaimedBy,
			"reason":    reason,
		},
	})
}

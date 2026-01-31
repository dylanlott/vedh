package server

import (
	"context"
	"encoding/json"
)

const (
	EventTypeLifeChanged        = "LIFE_CHANGED"
	EventTypeCardMoved          = "CARD_MOVED"
	EventTypeCardTapped         = "CARD_TAPPED"
	EventTypeLibraryShuffled    = "LIBRARY_SHUFFLED"
	EventTypeStackPushed        = "STACK_PUSHED"
	EventTypeStackResolved      = "STACK_RESOLVED"
	EventTypeTurnAdvanced       = "TURN_ADVANCED"
	EventTypePriorityPassed     = "PRIORITY_PASSED"
	EventTypeWinClaimed         = "WIN_CLAIMED"
	EventTypeWinClaimCancelled  = "WIN_CLAIM_CANCELLED"
	EventTypeGameFinished       = "GAME_FINISHED"
	EventTypeBoardstateSnapshot = "BOARDSTATE_SNAPSHOT"
)

func (s *graphQLServer) logEvent(ctx context.Context, event Event) {
	if s == nil || s.db == nil {
		return
	}
	g := &pgLogger{db: s.db}
	if err := g.Add(ctx, event); err != nil {
		s.loggerFor(ctx).Warn("failed to write gamelog event", "err", err, "game_id", event.GameID, "type", event.Type)
	}
}

func payloadFrom(v any) map[string]interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{}
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

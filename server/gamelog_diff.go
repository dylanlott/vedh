package server

import (
	"context"
)

func (s *graphQLServer) logBoardstateChanges(ctx context.Context, gameID string, actor string, prev *BoardState, next *BoardState) {
	if prev == nil || next == nil {
		return
	}
	if prev.Life != next.Life {
		s.logEvent(ctx, Event{
			GameID: gameID,
			Type:   EventTypeLifeChanged,
			Actor:  actor,
			Payload: map[string]interface{}{
				"user":  next.User,
				"from":  prev.Life,
				"to":    next.Life,
				"delta": next.Life - prev.Life,
			},
		})
	}

	if libraryShuffled(prev.Library, next.Library) {
		s.logEvent(ctx, Event{
			GameID: gameID,
			Type:   EventTypeLibraryShuffled,
			Actor:  actor,
			Payload: map[string]interface{}{
				"user": next.User,
			},
		})
	}

	prevMap := mapCardsByID(prev)
	nextMap := mapCardsByID(next)

	for id, nextRef := range nextMap {
		prevRef, ok := prevMap[id]
		if !ok {
			s.logEvent(ctx, Event{
				GameID: gameID,
				Type:   EventTypeCardMoved,
				Actor:  actor,
				Payload: map[string]interface{}{
					"cardId":   id,
					"cardName": nextRef.Card.Name,
					"fromZone": "UNKNOWN",
					"toZone":   nextRef.Zone,
					"toUser":   next.User,
				},
			})
			continue
		}
		if prevRef.Zone != nextRef.Zone {
			s.logEvent(ctx, Event{
				GameID: gameID,
				Type:   EventTypeCardMoved,
				Actor:  actor,
				Payload: map[string]interface{}{
					"cardId":   id,
					"cardName": nextRef.Card.Name,
					"fromZone": prevRef.Zone,
					"toZone":   nextRef.Zone,
					"fromUser": prev.User,
					"toUser":   next.User,
				},
			})
		}
		if prevRef.Zone == "Battlefield" && nextRef.Zone == "Battlefield" {
			prevTapped := prevRef.Card.Tapped != nil && *prevRef.Card.Tapped
			nextTapped := nextRef.Card.Tapped != nil && *nextRef.Card.Tapped
			if prevTapped != nextTapped {
				s.logEvent(ctx, Event{
					GameID: gameID,
					Type:   EventTypeCardTapped,
					Actor:  actor,
					Payload: map[string]interface{}{
						"cardId":   id,
						"cardName": nextRef.Card.Name,
						"user":     next.User,
						"tapped":   nextTapped,
					},
				})
			}
		}
	}

	for id, prevRef := range prevMap {
		if _, ok := nextMap[id]; ok {
			continue
		}
		s.logEvent(ctx, Event{
			GameID: gameID,
			Type:   EventTypeCardMoved,
			Actor:  actor,
			Payload: map[string]interface{}{
				"cardId":   id,
				"cardName": prevRef.Card.Name,
				"fromZone": prevRef.Zone,
				"toZone":   "UNKNOWN",
				"fromUser": prev.User,
			},
		})
	}
}

func (s *graphQLServer) logGameChanges(ctx context.Context, gameID string, actor string, prev *Game, next *Game) {
	if prev == nil || next == nil {
		return
	}
	if prev.Turn != nil && next.Turn != nil {
		if prev.Turn.Player != next.Turn.Player || prev.Turn.Phase != next.Turn.Phase || prev.Turn.Number != next.Turn.Number {
			s.logEvent(ctx, Event{
				GameID: gameID,
				Type:   EventTypeTurnAdvanced,
				Actor:  actor,
				Payload: map[string]interface{}{
					"from": map[string]interface{}{
						"player": prev.Turn.Player,
						"phase":  prev.Turn.Phase,
						"number": prev.Turn.Number,
					},
					"to": map[string]interface{}{
						"player": next.Turn.Player,
						"phase":  next.Turn.Phase,
						"number": next.Turn.Number,
					},
				},
			})
		}
	}

	prevStack := prev.Stack
	nextStack := next.Stack
	added := diffStack(prevStack, nextStack)
	for _, card := range added {
		if card == nil {
			continue
		}
		owner := ""
		if card.CurrentZone != nil {
			owner = *card.CurrentZone
		}
		s.logEvent(ctx, Event{
			GameID: gameID,
			Type:   EventTypeStackPushed,
			Actor:  actor,
			Payload: map[string]interface{}{
				"cardId":   card.ID,
				"cardName": card.Name,
				"owner":    owner,
			},
		})
	}

	if len(nextStack) < len(prevStack) {
		prevCounts := countStack(prevStack)
		nextCounts := countStack(nextStack)
		for _, card := range prevStack {
			key := stackKey(card)
			if key == "" {
				continue
			}
			if prevCounts[key] > nextCounts[key] {
				prevCounts[key]--
				owner := ""
				if card != nil && card.CurrentZone != nil {
					owner = *card.CurrentZone
				}
				s.logEvent(ctx, Event{
					GameID: gameID,
					Type:   EventTypeStackResolved,
					Actor:  actor,
					Payload: map[string]interface{}{
						"cardId":   card.ID,
						"cardName": card.Name,
						"owner":    owner,
					},
				})
			}
		}
	}
}

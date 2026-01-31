package server

import (
	"encoding/json"
)

var boardZones = []string{
	"Commander",
	"Battlefield",
	"Hand",
	"Graveyard",
	"Exiled",
	"Library",
	"Revealed",
	"Controlled",
}

type cardZoneRef struct {
	Zone string
	Card *Card
}

func cloneBoardState(bs *BoardState) *BoardState {
	if bs == nil {
		return nil
	}
	b, err := json.Marshal(bs)
	if err != nil {
		return bs
	}
	out := &BoardState{}
	if err := json.Unmarshal(b, out); err != nil {
		return bs
	}
	return out
}

func mapCardsByID(bs *BoardState) map[string]cardZoneRef {
	out := make(map[string]cardZoneRef)
	if bs == nil {
		return out
	}
	for _, zone := range boardZones {
		var cards []*Card
		switch zone {
		case "Commander":
			cards = bs.Commander
		case "Battlefield":
			cards = bs.Battlefield
		case "Hand":
			cards = bs.Hand
		case "Graveyard":
			cards = bs.Graveyard
		case "Exiled":
			cards = bs.Exiled
		case "Library":
			cards = bs.Library
		case "Revealed":
			cards = bs.Revealed
		case "Controlled":
			cards = bs.Controlled
		}
		for _, card := range cards {
			if card == nil || card.ID == "" {
				continue
			}
			out[card.ID] = cardZoneRef{Zone: zone, Card: card}
		}
	}
	return out
}

func libraryShuffled(prev []*Card, next []*Card) bool {
	if len(prev) < 2 || len(prev) != len(next) {
		return false
	}
	prevIDs := make(map[string]int, len(prev))
	nextIDs := make(map[string]int, len(next))
	orderSame := true
	for i, card := range prev {
		if card == nil || card.ID == "" {
			return false
		}
		prevIDs[card.ID]++
		if next[i] == nil || next[i].ID == "" || next[i].ID != card.ID {
			orderSame = false
		}
	}
	for _, card := range next {
		if card == nil || card.ID == "" {
			return false
		}
		nextIDs[card.ID]++
	}
	if orderSame {
		return false
	}
	if len(prevIDs) != len(nextIDs) {
		return false
	}
	for id, count := range prevIDs {
		if nextIDs[id] != count {
			return false
		}
	}
	return true
}

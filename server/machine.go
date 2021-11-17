package server

import (
	"fmt"

	"github.com/looplab/fsm"
)

type GameState struct {
	f *fsm.FSM
}

// New returns an FSM with the default phase transitions assigned.
// * It does not currently have any concept of a Player or turn passing.
// That logic must be added later. It only loops through one player's turn
// phases as an exercise in FSM construction
func New() *GameState {
	f := fsm.NewFSM("setup",
		fsm.Events{
			{Name: "untap", Src: []string{"setup", "untap"}, Dst: "upkeep"},
			{Name: "upkeep", Src: []string{"untap", "upkeep"}, Dst: "draw"},
			{Name: "draw", Src: []string{"upkeep", "draw"}, Dst: "main phase 1"},
			// main phase 1
			{Name: "main phase 1", Src: []string{"draw", "main phase 1", "exit combat"}, Dst: "enter combat"},
			// combat
			{Name: "enter combat", Src: []string{"main phase 1", "exit combat"}, Dst: "declare attackers"},
			{Name: "declare attackers", Src: []string{"enter combat"}, Dst: "declare blockers"},
			{Name: "declare blockers", Src: []string{"declare attackers"}, Dst: "resolve combat"},
			{Name: "resolve combat", Src: []string{"declare blockers"}, Dst: "exit combat"},
			{Name: "exit combat", Src: []string{"resolve combat"}, Dst: "main phase 2"},
			// main phase 2
			{Name: "main phase 2", Src: []string{"main phase 2", "end step"}, Dst: "end step"},
			// can't remember if end step happens before or after discard. I think it's _after_, though.
			{Name: "end step", Src: []string{"main phase 2", "end step"}, Dst: "discard"},
			{Name: "discard", Src: []string{"end step"}, Dst: "untap"},
		},
		fsm.Callbacks{})
	return &GameState{
		f: f,
	}
}

func (gs *GameState) LockIn() error {
	// check that each player has a valid deck and commander setup
	// turn must be established by here
	// decks and opening hands must be valid and chosen.
	// hands should be drawn after turn order is established.
	return fmt.Errorf("not impl")
}

func (gs *GameState) PassPriority() error {
	return fmt.Errorf("not impl")
}

package games

import (
	"fmt"
	"sync"
	"time"
)

// GameService defines a CRUD interface for Games.
type GameService interface {
	Create(game Game) (Game, error)
	Update(game Game) (Game, error)
	List() ([]Game, error)
	Get(id string) (Game, error)
}

type Boardstate interface {
	Battlefield()
	Hand()
	Library()
	Graveyard()
	Revealed()
	Exiled()
}

type Card interface{
	ID() string
	Name() string
	Data() map[string]interface{}
}

// Game declares the interface to our main resource: games
type Game interface {
	// All games must have a unique ID
	ID() string
	// Games are made up of player Boardstates in a specific order.
	// Turn order is described by the order of Boardstates returned by Players.
	Players() ([]Boardstate, error)
	// Games have a Pub/Sub model built into them for realtime
	PubSub
}

// PubSub declares a generic pub/sub interface for any type
type PubSub interface {
	Subscribe() (<-chan Game, error)
	Publish(game Game) error
}

// GQL is the eventual target for our GraphQL API to fulfill.
type GQL struct{} // TODO

// FullGame is an inmemory game store for testing and validation purposes
type FullGame struct {
	id string
	players []Boardstate
	createdAt time.Time
}

type MemStore struct {
	sync.Mutex

	games map[string]Game
}

func (m *MemStore) Create(game Game) (Game, error) {
	return nil, fmt.Errorf("not impl")
}
func (m *MemStore) Update(game Game) (Game, error) {
	return nil, fmt.Errorf("not impl")
}
func (m *MemStore) List() ([]Game, error) {
	m.Lock()
	defer m.Unlock()

	list := []Game{}
	for _, v := range m.games {
		list = append(list, v)
	}

	return list, nil
}
func (m *MemStore) Get(id string) (Game, error) {
	return nil, fmt.Errorf("not impl")
}

//
// FullGame
// Defines the FullGame interface which has a Pub Sub interface on it
//

// All games must have a unique ID
func (i *FullGame) ID() string {
	panic("not implemented") // TODO: Implement
}

// Games are made up of player Boardstates in a specific order.
// Turn order is described by the order of Boardstates returned by Players.
func (i *FullGame) Players() ([]Boardstate, error) {
	panic("not implemented") // TODO: Implement
}

// Subscribe emits any changes to the FullGame
func (i *FullGame) Subscribe() (<-chan Game, error) {
	panic("not implemented") // TODO: Implement
}

// Publish should be called every time a FullGame is updated.
func (i *FullGame) Publish(game Game) error {
	panic("not implemented") // TODO: Implement
}

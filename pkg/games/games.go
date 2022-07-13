package games

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

var buffer int64 = 1000 // size of event buffer per subscriber

// Games are managed by the GameService. Once a Game is acquired,
// all updates and state management is done through the Game API.

// GameService defines a CRUD interface for Games.
type GameService interface {
	Create(game Game) (Game, error)
	Join(gameID string, player Player) (Game, error)
	List() ([]Game, error)
	Get(id string) (Game, error)
}

// JSON is a shortcut type for dealing with JSON
type JSON map[string]interface{}

//
// PLAYER
//

// Player is essentially a wrapper around an arbitrary JSON
// object.
type Player interface {
	ID() string
	Boardstate() (JSON, error)
	Sync(JSON) error
	AttachGame(game Game) error
}

type inMemPlayer struct {
	sync.Mutex
	Game  Game
	state JSON
}

// NewPlayer returns a new inMemPlayer that fulfills the
// Player interface.
// We need to publish Game events when Player's update their
// individuals states. What is the best way to do that?
// This is my first approach, but we could do different stuff.
func NewPlayer(id string) (*inMemPlayer, error) {
	if id == "" {
		id = uuid.New().String()
	}
	p := &inMemPlayer{
		Game: nil, // NB: Game is set in AttachGame
		state: map[string]interface{}{
			"id": id,
		},
	}
	return p, nil
}

func (p *inMemPlayer) ID() string {
	if p.state == nil {
		p.state = make(JSON)
	}

	if val, ok := p.state["id"]; ok {
		return fmt.Sprintf("%v", val)
	}

	return "<ErrNoId>"
}

// Boardstate returns the current boardstate for the Player.
func (p *inMemPlayer) Boardstate() (JSON, error) {
	if p.state == nil {
		p.state = make(JSON)
	}
	return p.state, nil
}

// Sync
func (p *inMemPlayer) Sync(json JSON) error {
	p.Lock()
	p.state = json
	p.Unlock()
	if err := p.Game.Publish(p.Game); err != nil {
		return fmt.Errorf("failed to publish game event %w", err)
	}
	return nil
}

// AttachGame is necessary to attach a Game to a Player.
// This is meant to give them access to the Game's PubSub interface
// at player#Sync() call. I can't tell how I feel about this design yet.
// Will review this decision later.
func (p *inMemPlayer) AttachGame(game Game) error {
	p.Game = game
	return nil
}

//
// GAME
//

// Game declares the interface to our main resource: games
type Game interface {
	// All games must have a unique ID
	ID() string

	// Players returns a list of players in the Game.
	// * Games are made up of player Boardstates in a specific order.
	// * Turn order is described by the order of Boardstates returned by Players.
	Players() ([]Player, error)

	// Sync allows a Game to update n number of Players and updates
	// its internal view of those Players to match.
	// * If a Player can't be found by ID in the Game, it will error.
	// * Sync emits an update through PubSub every time a Player is successfully
	// updated.
	// * A failed Sync will not emit any events.
	// Sync(players ...Player) ([]Player, error)

	// Join will join a Player to the Game that matches `gameID`
	Join(player Player) (Game, error)

	// Get returns a single Player that matches `playerID` from the Game.
	// If no player exists with that ID it will return an error.
	Get(playerID string) (Player, error)
	// Games have a Pub/Sub model built into them for realtime notifications
	// of their entire view.
	PubSub
}

// TODO: make GQL a collection of FullGames served by GQL.
type GQL struct{} // TODO

// FullGame is an inmemory game store for testing and validation purposes
type FullGame struct {
	sync.Mutex

	id        string
	players   []Player
	createdAt time.Time

	subs []*Subscriber
}

// Subscriber relates a unique ID to a channel that messages are sent over.
// A subscriber receives Game messages any time a Player or Game is updated.
type Subscriber struct {
	id string
	ch chan Game
}

// MemStore fulfills the GameService interface and creates a API for
// managing multiple Games.
type MemStore struct {
	sync.Mutex

	games map[string]Game
}

// NewFullGame creates a new *FullGame
func (m *MemStore) NewFullGame(id string, players []Player) (*FullGame, error) {
	// we assign a random ID if one is not set.
	if id == "" {
		id = uuid.New().String()
	}

	if players == nil {
		players = []Player{}
	}

	g := &FullGame{
		id:        id,
		players:   players,
		createdAt: time.Now(),
		subs:      []*Subscriber{},
	}

	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	m.games[g.ID()] = g
	log.Printf("created game %s", id)

	return g, nil
}

// Join adds a Player to a game.
func (m *MemStore) Join(gameID string, player Player) (Game, error) {
	if gameID == "" {
		return nil, fmt.Errorf("ErrNoGameID")
	}

	m.Lock()
	defer m.Unlock()

	if game, ok := m.games[gameID]; !ok {
		return nil, fmt.Errorf("ErrGameNoExist")
	} else {
		return game.Join(player)
	}
}

// List returns a list of Games that are in the store currently.
func (m *MemStore) List() ([]Game, error) {
	m.Lock()
	defer m.Unlock()

	// TODO: return only active games.
	// TODO: sort and filter this list before returning
	list := []Game{}
	for _, v := range m.games {
		list = append(list, v)
	}

	return list, nil
}

// Get returns the Game that matches `id` or an error.
func (m *MemStore) Get(id string) (Game, error) {
	if game, ok := m.games[id]; ok {
		return game, nil
	}

	return nil, fmt.Errorf("ErrNoExist: %s", id)
}

// ID returns the FullGame's unique ID.
func (f *FullGame) ID() string {
	return f.id
}

// Games are made up of player Boardstates in a specific order.
// Turn order is described by the order of Boardstates returned by Players.
func (f *FullGame) Players() ([]Player, error) {
	return f.players, nil
}

// Get returns a single Player from the full game.
func (f *FullGame) Get(playerID string) (Player, error) {
	for _, player := range f.players {
		if player.ID() == playerID {
			return player, nil
		}
	}
	return nil, fmt.Errorf("ErrPlayerNotExist: id %s not found", playerID)
}

// Join will add a Player to a Game or return an error.
func (f *FullGame) Join(player Player) (Game, error) {
	if err := player.AttachGame(f); err != nil {
		return nil, fmt.Errorf("failed to attach game to player: %w", err)
	}

	log.Printf("player: %v", player)

	f.Lock()
	f.players = append(f.players, player)
	f.Unlock()

	f.Publish(f)

	return f, nil
}

//
//PUBSUB
//

// PubSub declares a generic pub/sub interface for any type
type PubSub interface {
	Subscribe() (chan Game, error)
	Publish(game Game) error
}

// Subscribe emits any changes to the FullGame
func (f *FullGame) Subscribe() (chan Game, error) {
	f.Lock()
	defer f.Unlock()

	id := uuid.New().String()
	s := &Subscriber{
		id: id,
		ch: make(chan Game, buffer),
	}
	f.subs = append(f.subs, s)

	log.Printf("subscribed %s", s.id)

	return s.ch, nil
}

// Publish should be called every time a FullGame is updated.
func (f *FullGame) Publish(game Game) error {
	for _, v := range f.subs {
		v.ch <- game
		log.Printf("published game event %v", game)
	}
	return nil
}

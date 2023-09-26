package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// BoardObserver wraps a UserID to a BoardState channel emiter.
type BoardObserver struct {
	// The UserID of the user subscribing to BoardUpdates
	UserID string
	// the channel that BoardUpdates are passed down
	Channel chan *BoardState
}

// FullBoardstate binds a set of observers to a game ID and user ID
type FullBoardstate struct {
	sync.Mutex
	// Game ID of the Boardstate in play
	GameID string
	// User ID of the Boardstate being observed
	UserID string
	// Observers keeps a map of UserID to BoardObservers
	Observers map[string]*BoardObserver
}

func (s *graphQLServer) BoardstateUpdated(ctx context.Context,
	obsID string,
	userID string,
) (<-chan *BoardState, error) {
	ch, err := s.registerObserver(ctx, obsID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to register listener: %w", err)
	}

	return ch, nil
}

func (s *graphQLServer) UpdateBoardState(
	ctx context.Context,
	input InputBoardState,
) (*BoardState, error) {
	if input.User == "" {
		return nil, fmt.Errorf("invalid user")
	}

	bs, err := boardStateFromInput(input)
	if err != nil {
		return nil, fmt.Errorf("invalid boardstate: %w", err)
	}

	game, err := s.GetGame(ctx, bs.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game from the database: %w", err)
	}

	// update matching username's boardstate
	for index, player := range game.Players {
		if player.Username == bs.User {
			game.Players[index].Boardstate = bs
			// break early
			break
		}
	}

	go s.publishBoardstate(bs)

	if err := s.upsertGame(game); err != nil {
		return nil, fmt.Errorf("failed to update player %s boardstate %w", bs.User, err)
	}

	return bs, nil
}

// Boardstates queries Redis for different boardstates per player or game
func (s *graphQLServer) Boardstates(ctx context.Context, gameID string, username *string) ([]*BoardState, error) {
	// get the game from the database
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	// send only given users boardstate
	if username != nil {
		for _, u := range game.Players {
			if u.Username == *username {
				return []*BoardState{u.Boardstate}, nil
			}
		}
	}
	// if username is not provided, send all boardstates
	var list []*BoardState
	for _, u := range game.Players {
		list = append(list, u.Boardstate)
	}
	return list, nil
}

// converts an InputBoardState to a native BoardState type or returns an error
// if its an invalid BoardState.
func boardStateFromInput(bs InputBoardState) (*BoardState, error) {
	data, err := json.Marshal(bs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input game: %s", err)
	}
	new := &BoardState{}
	err = json.Unmarshal(data, &new)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %s", err)
	}

	return new, nil
}

// publishes a boardstate update in a threadsafe function.
// * if the userID is not found, it logs an error and returns immediately.
// * if a boardstate for a given userID doesnt' exist, it logs an error and
// returns.
// * it acquires a lock on the *FullBoardState and this must be respected
// or else we'll run into race conditions as well.
func (s *graphQLServer) publishBoardstate(bs *BoardState) {
	log.Printf("boardstate published: %v", bs)
	s.mutex.Lock()
	fbs, ok := s.boards[bs.User]
	if !ok {
		log.Printf("pubishBoardState error: could not find boardstate: %s", bs.User)
		return
	}
	obs := fbs.Observers
	s.mutex.Unlock()

	fbs.Mutex.Lock()
	for _, v := range obs {
		v.Channel <- bs
	}
	fbs.Mutex.Unlock()
}

// registerObserver will add an observer with ID obsID to the map of observers
// for userID's BoardState. It returns a channel of BoardState updates or an
// error.
func (s *graphQLServer) registerObserver(ctx context.Context, obsID string, userID string) (chan *BoardState, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// locate if a boardstate exists for that userID already
	fbs, ok := s.boards[userID]
	if !ok {
		// create a fullboardstate since this userID has no current subscribers
		full := &FullBoardstate{
			UserID:    userID,
			Observers: map[string]*BoardObserver{},
		}
		// create and assign observer with obsID to that BoardStates's observers
		obs := &BoardObserver{
			UserID:  obsID,
			Channel: make(chan *BoardState),
		}

		// map observers by ID to the full board state
		full.Observers[obsID] = obs

		// map the fullboardstate by observed boardstate's userID
		s.boards[userID] = full

		// clean up after ourselves
		go func() {
			<-ctx.Done()
			full.Mutex.Lock()
			log.Printf("cleaning up observer [%s] of boardstate [%s]", obsID, userID)
			delete(full.Observers, obsID)
			full.Mutex.Unlock()
		}()

		log.Printf("registered observer [%s] to user [%s] boardstate", obsID, userID)
		return obs.Channel, nil
	}
	// if the fullboardstate exists, then create a new observer
	// and assign it to the fullboardstate
	obs := &BoardObserver{
		UserID:  obsID,
		Channel: make(chan *BoardState),
	}
	if fbs.Observers == nil {
		log.Printf("observers was empty, making a new boardstate observers map")
		fbs.Observers = make(map[string]*BoardObserver)
	}
	fbs.Observers[obsID] = obs
	log.Printf("registered observer [%s] to user [%s] boardstate", obsID, userID)
	log.Printf("list of observers: %+v", fbs.Observers)
	return obs.Channel, nil
}

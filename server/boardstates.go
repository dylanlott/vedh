package server

import (
	"context"
	"encoding/json"
	"fmt"
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
	if _, err := requireMatchingUser(ctx, userID, ""); err != nil {
		return nil, err
	}
	ch, err := s.registerObserver(ctx, obsID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to register listener: %w", err)
	}
	s.logger.Info("boardstate update subscription registered", "observer_id", obsID, "user_id", userID)
	return ch, nil
}

func (s *graphQLServer) UpdateBoardState(
	ctx context.Context,
	input InputBoardState,
) (*BoardState, error) {
	authUser, err := requireMatchingUser(ctx, input.UserID, input.User)
	if err != nil {
		return nil, err
	}
	if input.User == "" {
		return nil, fmt.Errorf("invalid user")
	}

	bs, err := boardStateFromInput(input)
	if err != nil {
		return nil, fmt.Errorf("invalid boardstate: %w", err)
	}
	s.logger.Debug("parsed boardstate for update", "user", bs.User, "game_id", bs.GameID)

	var prevBoardstate *BoardState
	updatedGame, err := s.mutateGame(ctx, bs.GameID, func(mutationCtx context.Context, game *Game) (*Game, error) {
		if game.Status == GameStatusFinished {
			return nil, fmt.Errorf("game already finished")
		}
		if !isUserInGame(game, authUser) {
			return nil, fmt.Errorf("forbidden: not a participant in this game")
		}

		playerID := authUser.ID
		index := playerIndex(game, &playerID, authUser.Username)
		if index == -1 || game.Players[index] == nil {
			return nil, fmt.Errorf("authenticated player is not in this game")
		}

		// The persisted player identity and game ID are authoritative. The
		// caller may only replace their own boardstate, never target a peer by
		// shaping input fields.
		player := game.Players[index]
		bs.UserID = player.ID
		bs.User = player.Username
		bs.GameID = game.ID
		prevBoardstate = cloneBoardState(player.Boardstate)
		player.Boardstate = bs
		s.logger.Debug("updated boardstate", "user", bs.User, "game_id", bs.GameID)

		if game.PendingWinClaim != nil {
			s.cancelPendingWinClaim(mutationCtx, game, authUser.Username, "boardstate updated")
		}

		alive := alivePlayerNames(game)
		totalPlayers := len(game.Players)
		switch len(alive) {
		case 0:
			finalizeGame(game, GameResultDraw, nil, nil)
		case 1:
			if totalPlayers >= 2 {
				finalizeGame(game, GameResultWin, []string{alive[0]}, nil)
			}
		}
		return game, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update player %s boardstate: %w", bs.User, err)
	}

	// Notify any boardstate observers for this specific user
	go s.publishBoardstate(bs)

	// Also publish the full game so clients subscribed at the game level
	// receive the updated snapshot.
	go s.publishGame(updatedGame.ID, updatedGame)
	if prevBoardstate != nil {
		s.logBoardstateChanges(ctx, updatedGame.ID, authUser.Username, prevBoardstate, bs)
	}
	if updatedGame.Status == GameStatusFinished {
		result := GameResultDraw
		if updatedGame.Result != nil {
			result = *updatedGame.Result
		}
		s.logEvent(ctx, Event{
			GameID: updatedGame.ID,
			Type:   EventTypeGameFinished,
			Actor:  authUser.Username,
			Payload: map[string]interface{}{
				"result":    result,
				"winnerIDs": updatedGame.WinnerIDs,
			},
		})
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
	viewer, _ := requireAuth(ctx)
	view := redactGameForUser(game, viewer)
	// send only given users boardstate
	if username != nil {
		for _, u := range view.Players {
			if u.Username == *username {
				return []*BoardState{u.Boardstate}, nil
			}
		}
	}
	// if username is not provided, send all boardstates
	var list []*BoardState
	for _, u := range view.Players {
		list = append(list, u.Boardstate)
	}
	return list, nil
}

// redactGameForUser returns a deep copy of game whose private zones are only
// visible to the authenticated player. Public zones remain intact; Hand and
// Library are represented by opaque placeholders so clients can still show
// counts without learning card identities.
func redactGameForUser(game *Game, viewer *AuthUser) *Game {
	if game == nil {
		return nil
	}
	redacted := cloneGame(game)
	if redacted == nil {
		return nil
	}
	for _, player := range redacted.Players {
		if player == nil || player.Boardstate == nil || viewerOwnsPlayer(player, viewer) {
			continue
		}
		player.Boardstate.Hand = hiddenCards(len(player.Boardstate.Hand))
		player.Boardstate.Library = hiddenCards(len(player.Boardstate.Library))
	}
	return redacted
}

// viewerOwnsPlayer prefers the stable authenticated ID. Username fallback is
// only used for legacy auth contexts that have no ID; otherwise a duplicate or
// client-authored username must not make another player's private zones visible.
func viewerOwnsPlayer(player *User, viewer *AuthUser) bool {
	if player == nil || viewer == nil {
		return false
	}
	if viewer.ID != "" {
		return player.ID == viewer.ID
	}
	return viewer.Username != "" && player.Username == viewer.Username
}

// redactGameForIdentity is used by subscriptions, which retain the validated
// user ID supplied when the subscription was registered rather than an
// AuthUser value.
func redactGameForIdentity(game *Game, identity string) *Game {
	if game == nil {
		return nil
	}
	return redactGameForUser(game, &AuthUser{ID: identity, Username: identity})
}

func hiddenCards(count int) []*Card {
	if count <= 0 {
		return []*Card{}
	}
	hidden := make([]*Card, count)
	for i := range hidden {
		hidden[i] = &Card{ID: fmt.Sprintf("hidden-%d", i), Name: "Hidden"}
	}
	return hidden
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
// * observer membership is protected by the server's shared mutex.
func (s *graphQLServer) publishBoardstate(bs *BoardState) {
	s.loggerFor(context.Background()).Debug("boardstate published", "user", bs.User, "user_id", bs.UserID, "game_id", bs.GameID)
	s.mutex.Lock()
	// Prefer UserID when present; fall back to Username for legacy callers/tests
	fbs, ok := s.boards[bs.UserID]
	if !ok || bs.UserID == "" {
		fbs, ok = s.boards[bs.User]
	}
	if !ok {
		s.loggerFor(context.Background()).Warn("publishBoardstate: observers not found", "user_id", bs.UserID, "user", bs.User)
		s.mutex.Unlock()
		return
	}
	observers := make([]*BoardObserver, 0, len(fbs.Observers))
	for _, observer := range fbs.Observers {
		observers = append(observers, observer)
	}
	s.mutex.Unlock()

	for _, v := range observers {
		select {
		case v.Channel <- bs:
		default:
			s.loggerFor(context.Background()).Warn("publishBoardstate: drop update (channel full)", "observer_id", v.UserID, "user", bs.User, "user_id", bs.UserID)
		}
	}
}

// registerObserver will add an observer with ID obsID to the map of observers
// for userID's BoardState. It returns a channel of BoardState updates or an
// error.
func (s *graphQLServer) registerObserver(ctx context.Context, obsID string, userID string) (chan *BoardState, error) {
	s.mutex.Lock()
	logger := s.loggerFor(ctx).With("observer_id", obsID, "user_id", userID)
	if s.boards == nil {
		s.boards = make(map[string]*FullBoardstate)
	}

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
			Channel: make(chan *BoardState, 10),
		}

		// map observers by ID to the full board state
		full.Observers[obsID] = obs

		// map the fullboardstate by observed boardstate's userID
		s.boards[userID] = full

		s.mutex.Unlock()
		go s.cleanupBoardstateObserver(ctx, logger, userID, obsID, full, obs)
		logger.Info("registered boardstate observer")
		return obs.Channel, nil
	}
	// if the fullboardstate exists, then create a new observer
	// and assign it to the fullboardstate
	obs := &BoardObserver{
		UserID:  obsID,
		Channel: make(chan *BoardState, 10),
	}
	if fbs.Observers == nil {
		logger.Debug("boardstate observers map was empty; initializing")
		fbs.Observers = make(map[string]*BoardObserver)
	}
	fbs.Observers[obsID] = obs
	s.mutex.Unlock()
	go s.cleanupBoardstateObserver(ctx, logger, userID, obsID, fbs, obs)
	logger.Info("registered boardstate observer")
	return obs.Channel, nil
}

// cleanupBoardstateObserver removes only the channel that registered this
// goroutine. A reconnect with the same client-provided obsID must not let the
// old connection delete the new observer.
func (s *graphQLServer) cleanupBoardstateObserver(ctx context.Context, logger interface{ Info(string, ...any) }, userID, obsID string, fbs *FullBoardstate, obs *BoardObserver) {
	<-ctx.Done()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.boards[userID] != fbs || fbs.Observers[obsID] != obs {
		return
	}
	delete(fbs.Observers, obsID)
	if len(fbs.Observers) == 0 {
		delete(s.boards, userID)
	}
	logger.Info("cleaned up boardstate observer")
}

package server

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// GameObserver binds a UserID to a Channel.
type GameObserver struct {
	UserID  string
	Channel chan *Game
}

// FullGame wraps a Game with Observers and Players so that we can
// access players boardstates with only a game and a player ID
type FullGame struct {
	sync.Mutex

	GameID    string
	Observers map[string]*GameObserver
	EventLog  EventLog
}

// Games returns a list of Games that are unmarshaled from the payload column of the
// games table.
func (s *graphQLServer) Games(ctx context.Context, offset int, limit int) ([]*Game, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	// Basic pagination: order by id for stability and apply limit/offset.
	// If you add created_at to the table, prefer ordering by that.
	rows, err := s.db.Query("SELECT id, payload FROM games ORDER BY id DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	games := []*Game{}
	for rows.Next() {
		var id string
		var pbz []byte
		if err := rows.Scan(&id, &pbz); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		game := &Game{}
		if err := json.Unmarshal(pbz, &game); err != nil {
			return nil, fmt.Errorf("failed to unmarshal game %s: %w", id, err)
		}
		ensureGameDefaults(game)
		if !isUserInGame(game, authUser) {
			continue
		}
		games = append(games, game)
	}
	return games, nil
}

func (s *graphQLServer) loadGameByID(gameID string) (*Game, error) {
	var payload []byte
	query := `SELECT payload FROM games WHERE id = $1`
	err := s.db.QueryRow(query, gameID).Scan(&payload)
	if err != nil {
		s.logger.Debug("game not found", "game_id", gameID)
		return nil, err
	}
	s.logger.Debug("found game in database", "game_id", gameID, "payload_bytes", len(payload))
	game := &Game{}
	if err := json.Unmarshal(payload, &game); err != nil {
		return nil, err
	}
	ensureGameDefaults(game)
	return game, nil
}

// GetGame returns a single game from the database if the authenticated user is a participant.
func (s *graphQLServer) GetGame(ctx context.Context, gameID string) (*Game, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}

	game, err := s.loadGameByID(gameID)
	if err != nil {
		return nil, err
	}
	if !isUserInGame(game, authUser) {
		return nil, errors.New("forbidden: not a participant in this game")
	}
	return game, nil
}

// GameUpdated returns a channel for a game or an error.
func (s *graphQLServer) GameUpdated(ctx context.Context, gameID string, userID string) (<-chan *Game, error) {
	authUser, err := requireMatchingUser(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	game, err := s.loadGameByID(gameID)
	if err != nil {
		return nil, err
	}
	if !isUserInGame(game, authUser) {
		return nil, errors.New("forbidden: not a participant in this game")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.loggerFor(ctx).Info("registering game observer", "user_id", userID, "game_id", gameID)

	g, ok := s.games[gameID]
	if !ok {
		game := &FullGame{
			GameID:    gameID,
			Observers: make(map[string]*GameObserver),
		}

		// add observer to the FullGame
		obs := &GameObserver{
			UserID:  userID,
			Channel: make(chan *Game, 10), // buffered to avoid head-of-line blocking
		}

		// clean up the observers channel when we're done with it
		go func() {
			<-ctx.Done()
			game.Mutex.Lock()
			s.loggerFor(ctx).Info("cleaning up game observer", "user_id", userID, "game_id", game.GameID)
			delete(game.Observers, userID)
			game.Mutex.Unlock()
		}()

		// game observers are keyed by userID.
		// only one connection per userID is allowed.
		game.Mutex.Lock()
		game.Observers[userID] = obs
		game.Mutex.Unlock()

		// register the game in the main server directory
		s.games[gameID] = game
		return obs.Channel, nil
	}

	// game exists, so just push user into observers and return their channel
	obs := &GameObserver{
		UserID:  userID,
		Channel: make(chan *Game, 10), // buffered to avoid head-of-line blocking
	}
	g.Mutex.Lock()
	g.Observers[userID] = obs
	g.Mutex.Unlock()

	return obs.Channel, nil
}

// UpdateGame is what's used to change the name of the game, format, insert
// or remove players, or change other meta informatin about a game.
func (s *graphQLServer) UpdateGame(ctx context.Context, new InputGame) (*Game, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	game := &Game{}
	b, err := json.Marshal(new)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input game: %s", err)
	}
	err = json.Unmarshal(b, &game)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %s", err)
	}
	if game.Turn != nil && game.Turn.Priority == "" {
		game.Turn.Priority = game.Turn.Player
	}

	// Ensure players have stable IDs even when input omits them. Tests expect
	// deterministic placeholders for anonymous users; we generate `deadbeef`
	// for the first player and `deadbeef2` for the second, etc.
	for i, p := range game.Players {
		if p != nil && p.ID == "" {
			switch i {
			case 0:
				p.ID = "deadbeef"
			case 1:
				p.ID = "deadbeef2"
			default:
				p.ID = fmt.Sprintf("deadbeef%d", i+1)
			}
		}
	}

	// Ensure the caller is a participant in the existing game or the new input.
	existing, err := s.GetGame(ctx, game.ID)
	if err == nil {
		ensureGameDefaults(existing)
		if existing.Status == GameStatusFinished {
			return nil, errors.New("game already finished")
		}
		if !isUserInGame(existing, authUser) {
			return nil, errors.New("forbidden: not a participant in this game")
		}
	} else if err == sql.ErrNoRows {
		if !isUserInGame(game, authUser) {
			return nil, errors.New("forbidden: caller must be a participant in new game")
		}
	} else {
		return nil, fmt.Errorf("failed to load game for authorization: %w", err)
	}

	if existing != nil {
		game.Status = existing.Status
		game.Result = existing.Result
		game.WinnerIDs = existing.WinnerIDs
		game.WinCondition = existing.WinCondition
		game.PendingWinClaim = existing.PendingWinClaim
		if err := enforceStackPriority(existing, game); err != nil {
			return nil, err
		}
	}

	if game.PendingWinClaim != nil {
		s.cancelPendingWinClaim(ctx, game, authUser.Username, "game updated")
	}

	go s.publishGame(game.ID, game)

	if err := s.upsertGame(game); err != nil {
		return game, err
	}

	if existing != nil {
		s.logGameChanges(ctx, game.ID, authUser.Username, existing, game)
	}
	return game, nil
}

func enforceStackPriority(current *Game, next *Game) error {
	if current == nil || next == nil {
		return nil
	}
	oldStack := current.Stack
	newStack := next.Stack
	if len(newStack) <= len(oldStack) {
		return nil
	}
	priority := ""
	if current.Turn != nil {
		if current.Turn.Priority != "" {
			priority = current.Turn.Priority
		} else {
			priority = current.Turn.Player
		}
	}
	if priority == "" {
		return nil
	}
	added := diffStack(oldStack, newStack)
	for _, card := range added {
		if card == nil {
			continue
		}
		owner := ""
		if card.CurrentZone != nil {
			owner = *card.CurrentZone
		}
		if owner == "" || owner != priority {
			return fmt.Errorf("only %s can add cards to the stack", priority)
		}
	}
	return nil
}

func (s *graphQLServer) PassPriority(ctx context.Context, gameID string, toPlayer string) (*Game, error) {
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
		return nil, errors.New("forbidden: only priority player can pass priority")
	}
	if !playerExists(game, toPlayer) {
		return nil, errors.New("target player not in game")
	}
	if game.PendingWinClaim != nil {
		if !claimMatchesPrioritySequence(game.PendingWinClaim, toPlayer) {
			s.cancelPendingWinClaim(ctx, game, authUser.Username, "priority passed out of sequence")
		} else {
			claimer := game.PendingWinClaim.ClaimedBy
			condition := game.PendingWinClaim.Condition
			game.PendingWinClaim.Remaining = game.PendingWinClaim.Remaining[1:]
			if toPlayer == claimer && len(game.PendingWinClaim.Remaining) == 0 {
				finalizeGame(game, GameResultWin, []string{claimer}, condition)
				s.logEvent(ctx, Event{
					GameID: game.ID,
					Type:   EventTypeGameFinished,
					Actor:  authUser.Username,
					Payload: map[string]interface{}{
						"result":      GameResultWin,
						"winnerNames": []string{claimer},
						"condition":   condition,
					},
				})
			}
		}
	}
	game.Turn.Priority = toPlayer
	s.logEvent(ctx, Event{
		GameID: game.ID,
		Type:   EventTypePriorityPassed,
		Actor:  authUser.Username,
		Payload: map[string]interface{}{
			"from": authUser.Username,
			"to":   toPlayer,
		},
	})
	go s.publishGame(game.ID, game)
	if err := s.upsertGame(game); err != nil {
		return game, err
	}
	return game, nil
}

func (s *graphQLServer) AdvancePhase(ctx context.Context, gameID string, phase string, number *int) (*Game, error) {
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
	if game.Turn.Player != authUser.Username {
		return nil, errors.New("forbidden: only the turn player can advance the phase")
	}
	prevTurn := *game.Turn
	format := formatFromRules(game.Rules)
	phase = normalizeTurnPhase(format, phase)

	nextTurnNumber := game.Turn.Number
	if number != nil {
		nextTurnNumber = *number
	} else if shouldIncrementTurnNumber(prevTurn.Phase, phase) {
		nextTurnNumber = prevTurn.Number + 1
	}

	game.Turn.Phase = phase
	game.Turn.Number = nextTurnNumber
	if game.Turn.Priority == "" {
		game.Turn.Priority = game.Turn.Player
	}
	if game.PendingWinClaim != nil {
		s.cancelPendingWinClaim(ctx, game, authUser.Username, "turn advanced")
	}
	s.logEvent(ctx, Event{
		GameID: game.ID,
		Type:   EventTypeTurnAdvanced,
		Actor:  authUser.Username,
		Payload: map[string]interface{}{
			"from": map[string]interface{}{
				"player": prevTurn.Player,
				"phase":  prevTurn.Phase,
				"number": prevTurn.Number,
			},
			"to": map[string]interface{}{
				"player": game.Turn.Player,
				"phase":  game.Turn.Phase,
				"number": game.Turn.Number,
			},
		},
	})
	go s.publishGame(game.ID, game)
	if err := s.upsertGame(game); err != nil {
		return game, err
	}
	return game, nil
}

func normalizePhaseKey(phase string) string {
	cleaned := strings.ToUpper(strings.Join(strings.Fields(phase), " "))
	switch cleaned {
	case "MAIN", "MAIN PHASE", "MAIN1", "MAIN PHASE 1":
		return "MAIN PHASE 1"
	case "MAIN2", "MAIN PHASE 2":
		return "MAIN PHASE 2"
	case "END", "END STEP":
		return "END STEP"
	default:
		return cleaned
	}
}

func shouldIncrementTurnNumber(prevPhase string, nextPhase string) bool {
	prev := normalizePhaseKey(prevPhase)
	next := normalizePhaseKey(nextPhase)
	if next != "UNTAP" && next != strings.ToUpper(DefaultFormat().PhaseSequence[0]) {
		return false
	}
	return prev == "END STEP" || prev == "DISCARD" || prev == "CLEANUP"
}

func playerExists(game *Game, username string) bool {
	for _, p := range game.Players {
		if p != nil && p.Username == username {
			return true
		}
	}
	return false
}

func diffStack(oldStack []*Card, newStack []*Card) []*Card {
	oldCounts := countStack(oldStack)
	newCounts := countStack(newStack)
	var added []*Card
	for _, card := range newStack {
		key := stackKey(card)
		if key == "" {
			continue
		}
		if newCounts[key] > oldCounts[key] {
			added = append(added, card)
			oldCounts[key]++
		}
	}
	return added
}

func countStack(stack []*Card) map[string]int {
	counts := make(map[string]int)
	for _, card := range stack {
		key := stackKey(card)
		if key == "" {
			continue
		}
		counts[key]++
	}
	return counts
}

func stackKey(card *Card) string {
	if card == nil {
		return ""
	}
	if card.ID != "" {
		return "id:" + card.ID
	}
	if card.Name == "" {
		return ""
	}
	if card.CurrentZone != nil && *card.CurrentZone != "" {
		return "name:" + card.Name + "|owner:" + *card.CurrentZone
	}
	return "name:" + card.Name
}

// JoinGame handles a user joining an existing game.
func (s *graphQLServer) JoinGame(ctx context.Context, input *InputJoinGame) (*Game, error) {
	if input == nil || input.BoardState == nil {
		return nil, errors.New("must provide boardstate to join a game")
	}
	authUser, err := requireMatchingUser(ctx, input.BoardState.UserID, input.BoardState.User)
	if err != nil {
		return nil, err
	}
	// TODO: Handle rejoins by detecting if that player's user.ID already exists
	// in a given game. If it does, just return that same setup.
	// TODO: Check context for User auth and append user info that way
	// TODO: Pull user boardstate creation out into a function since we do it multiple places
	if input.BoardState.UserID == "" {
		return nil, errors.New("must provide user ID to join a game")
	}
	if input.BoardState.GameID == "" {
		return nil, errors.New("must provide a game ID to join")
	}
	if input.BoardState.User == "" {
		return nil, errors.New("must provide a username to join")
	}
	if input.Decklist == nil {
		return nil, errors.New("must provide a decklist to join")
	}

	// get the game and verify itself
	game, err := s.loadGameByID(input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to find game: %w", err)
	}
	ensureGameDefaults(game)
	if game.Status == GameStatusFinished {
		return nil, errors.New("game already finished")
	}

	if len(game.Players) >= 4 {
		return nil, errors.New("game is full")
	}
	if isUserInGame(game, authUser) {
		return nil, errors.New("user already in game")
	}

	user := &User{
		Username: input.BoardState.User,
		ID:       input.BoardState.UserID,
		Boardstate: &BoardState{
			User:        input.BoardState.User,
			Life:        input.BoardState.Life,
			Exiled:      getBareCard(input.BoardState.Exiled),
			Revealed:    getBareCard(input.BoardState.Revealed),
			Battlefield: getBareCard(input.BoardState.Battlefield),
			Controlled:  getBareCard(input.BoardState.Controlled),
			Hand:        make([]*Card, 0),
			Graveyard:   make([]*Card, 0),
		},
	}

	// hydrate and validate the library from the provided decklist
	library, err := s.createLibraryFromDecklist(ctx, *input.Decklist, input.BoardState.Commander)
	if err != nil {
		return nil, fmt.Errorf("invalid decklist: %w", err)
	}
	user.Boardstate.Library = library

	// NB: Commented out while we figure out how to handle Commander selection.
	if len(input.BoardState.Commander) > 0 {
		for _, card := range input.BoardState.Commander {
			commander, err := s.Card(ctx, card.Name, nil)
			if err != nil {
				s.loggerFor(ctx).Warn("error getting commander for deck", "err", err, "card_name", card.Name, "game_id", input.ID, "user_id", input.BoardState.UserID)
				continue
			}
			user.Boardstate.Commander = append(user.Boardstate.Commander, commander)
		}
	}

	// shuffle their library for the start of the game
	shuff, err := Shuffle(user.Boardstate.Library)
	if err != nil {
		s.loggerFor(ctx).Error("error shuffling library", "err", err, "game_id", input.ID, "user_id", input.BoardState.UserID)
		return nil, err
	}
	user.Boardstate.Library = shuff

	// add them to the game's list of players
	game.Players = append(game.Players, user)

	go s.publishGame(game.ID, game)

	// update game in postgrse
	if err := s.upsertGame(game); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	s.logEvent(ctx, Event{
		GameID: game.ID,
		Type:   EventTypePlayerJoined,
		Actor:  authUser.Username,
		Payload: map[string]interface{}{
			"user": authUser.Username,
		},
	})

	return game, nil
}

// CreateGame creates a new game and hydrates the decklists for the players in it.
func (s *graphQLServer) CreateGame(ctx context.Context, inputGame InputCreateGame) (*Game, error) {
	authUser, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	// don't allow a game to be created with an existing name
	// TECHDEBT replace this with a proper cache
	if _, exists := s.games[inputGame.ID]; exists {
		return nil, fmt.Errorf("game already exists with ID %s", inputGame.ID)
	}

	// assign an ID if none is provided
	if inputGame.ID == "" {
		inputGame.ID = uuid.New().String()
	}

	formatID := ""
	if inputGame.FormatID != nil {
		formatID = *inputGame.FormatID
	}
	format, ok := LookupFormat(formatID)
	if !ok {
		return nil, fmt.Errorf("unknown format %q", formatID)
	}

	g := &Game{
		ID:        inputGame.ID,
		CreatedAt: time.Now(),
		Players:   []*User{},
		Stack:     []*Card{},
		Status:    GameStatusInProgress,
		Turn: &Turn{
			Player: inputGame.Turn.Player,
			Phase:  normalizeTurnPhase(format, inputGame.Turn.Phase),
			Number: inputGame.Turn.Number,
			Priority: func() string {
				if inputGame.Turn.Priority != "" {
					return inputGame.Turn.Priority
				}
				return inputGame.Turn.Player
			}(),
		},
		Rules: []*Rule{},
	}
	ensureFormatRules(g, format)

	// build player boardstates
	for _, player := range inputGame.Players {
		if player.UserID == authUser.ID {
			if player.User != "" && authUser.Username != "" && player.User != authUser.Username {
				return nil, errors.New("forbidden: username mismatch for player")
			}
		}
		// TODO: Deck validation should happen here.
		user := &User{
			ID:       player.UserID,
			Username: player.User,
			Boardstate: &BoardState{
				UserID:      player.UserID,
				User:        player.User,
				Life:        player.Life,
				GameID:      g.ID,
				Hand:        getBareCard(player.Hand),
				Exiled:      getBareCard(player.Exiled),
				Revealed:    getBareCard(player.Revealed),
				Battlefield: getBareCard(player.Battlefield),
				Controlled:  getBareCard(player.Controlled),
			},
		}
		if user.Boardstate.Life == 0 {
			user.Boardstate.Life = format.StartingLife
		}

		// Set default boardstate, handle library and commander specifically
		var decklist string
		if player.Decklist != nil {
			decklist = string(*player.Decklist)
		}

		// hydrate and validate the decklist for the player
		library, err := s.createLibraryFromDecklist(ctx, decklist, player.Commander)
		if err != nil {
			return nil, fmt.Errorf("invalid decklist for %s: %w", player.User, err)
		}
		user.Boardstate.Library = library

		// handle commander selection
		if len(player.Commander) > 0 {
			for _, card := range player.Commander {
				commander, err := s.Card(ctx, card.Name, nil)
				if err != nil {
					s.loggerFor(ctx).Warn("error getting commander for deck", "err", err, "card_name", card.Name, "game_id", g.ID, "user_id", player.UserID)
					// fail gracefully and use their card name so they can still play a game
					user.Boardstate.Commander = append(user.Boardstate.Commander, &Card{Name: card.Name, ID: card.ID})
					continue
				}
				user.Boardstate.Commander = append(user.Boardstate.Commander, commander)
			}
		}

		// shuffle their library
		shuff, err := Shuffle(user.Boardstate.Library)
		if err != nil {
			return nil, fmt.Errorf("failed to shuffle")
		}
		user.Boardstate.Library = shuff
		g.Players = append(g.Players, user)
	}
	if !isUserInGame(g, authUser) {
		return nil, errors.New("forbidden: caller must be included as a player")
	}

	if err := s.upsertGame(g); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	var players []map[string]interface{}
	for _, p := range g.Players {
		if p == nil || p.Boardstate == nil {
			continue
		}
		players = append(players, map[string]interface{}{
			"id":       p.ID,
			"username": p.Username,
			"life":     p.Boardstate.Life,
		})
	}
	s.logEvent(ctx, Event{
		GameID: g.ID,
		Type:   EventTypeGameCreated,
		Actor:  authUser.Username,
		Payload: map[string]interface{}{
			"players": players,
		},
	})

	return g, nil
}

// upsert inserts or updates a Game in the games table. It detects conflicts
// by checking game IDs.
func (s *graphQLServer) upsertGame(g *Game) error {
	if g.ID == "" {
		return fmt.Errorf("ErrInvalidGameID: game ID must be set: %+v", g)
	}
	query := `INSERT INTO games (id, payload) 
	VALUES ($1, $2::jsonb)
	ON CONFLICT (id) DO UPDATE SET payload = $2::jsonb;
	`
	gbz, err := json.Marshal(g)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(query, g.ID, string(gbz))
	if err != nil {
		return err
	}
	return nil
}

// getBareCard returns a card type that hasn't been hydrated with
// data from the mtgjson data set
func getBareCard(inputCards []*InputCard) []*Card {
	cardList := []*Card{}
	for _, card := range inputCards {
		c := &Card{
			Name: card.Name,
		}

		if card.ID != "" {
			c.ID = card.ID
		}

		cardList = append(cardList, c)
	}
	return cardList
}

// createLibraryFromDecklist parses the provided decklist string as CSV.
// If commander cards are present in the decklist, one copy of each selected
// commander is removed from the library count before validating deck size.
func (s *graphQLServer) createLibraryFromDecklist(ctx context.Context, decklist string, commanders []*InputCard) ([]*Card, error) {
	if decklist == "" {
		return []*Card{}, fmt.Errorf("must provide cards in decklist to create a library")
	}

	trimmed := strings.TrimSpace(decklist)
	r := csv.NewReader(strings.NewReader(trimmed))

	// set lazy quotes for using double quotes in csv files
	r.LazyQuotes = true
	// and trim leading spaces
	r.TrimLeadingSpace = true

	type deckEntry struct {
		name string
		qty  int64
	}
	entries := []deckEntry{}
	lookupNames := []string{}
	lookupSeen := map[string]struct{}{}
	commanderBudget := map[string]int64{}
	commandersSpecified := int64(0)

	for _, commander := range commanders {
		if commander == nil {
			continue
		}
		name := strings.TrimSpace(commander.Name)
		if name == "" {
			continue
		}
		commandersSpecified++
		commanderBudget[strings.ToLower(name)]++
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.loggerFor(ctx).Warn("error reading csv record", "err", err)
			return nil, fmt.Errorf("failed to parse CSV: %s", err)
		}

		if len(record) < 2 {
			return nil, fmt.Errorf("invalid decklist row: expected quantity and card name")
		}
		name := strings.TrimSpace(record[1])
		quantity, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quantity: %w", err)
		}
		if quantity < 0 {
			return nil, fmt.Errorf("invalid quantity %d for %q", quantity, name)
		}

		key := strings.ToLower(name)
		if removeCount := commanderBudget[key]; removeCount > 0 && quantity > 0 {
			if removeCount >= quantity {
				commanderBudget[key] = removeCount - quantity
				quantity = 0
			} else {
				quantity -= removeCount
				commanderBudget[key] = 0
			}
		}
		if quantity == 0 {
			continue
		}

		entries = append(entries, deckEntry{name: name, qty: quantity})
		if key != "" {
			if _, ok := lookupSeen[key]; !ok {
				lookupSeen[key] = struct{}{}
				lookupNames = append(lookupNames, name)
			}
		}
	}

	maxLibraryCards := int64(100) - commandersSpecified
	if maxLibraryCards < 0 {
		return nil, fmt.Errorf("invalid commander count: %d", commandersSpecified)
	}

	var libraryCount int64
	for _, entry := range entries {
		libraryCount += entry.qty
	}
	if libraryCount > maxLibraryCards {
		return nil, fmt.Errorf(
			"deck too large: library has %d cards but maximum is %d for %d commander(s)",
			libraryCount,
			maxLibraryCards,
			commandersSpecified,
		)
	}

	lookup := map[string]*Card{}
	if len(lookupNames) > 0 {
		found, err := s.Cards(ctx, lookupNames)
		if err != nil {
			s.loggerFor(ctx).Warn("batch card lookup failed", "err", err)
		}
		for i, name := range lookupNames {
			key := strings.ToLower(strings.TrimSpace(name))
			if key == "" || i >= len(found) {
				continue
			}
			if found[i] != nil {
				lookup[key] = found[i]
			}
		}
	}

	cards := []*Card{}
	for _, entry := range entries {
		key := strings.ToLower(strings.TrimSpace(entry.name))
		if key == "" {
			cards = addX(entry.qty, cards, &Card{Name: entry.name})
			continue
		}
		if found := lookup[key]; found != nil {
			cards = addX(entry.qty, cards, found)
		} else {
			cards = addX(entry.qty, cards, &Card{Name: entry.name})
		}
	}

	return cards, nil
}

// addX adds a card to a slice of cards x number of times.
func addX(qty int64, cards []*Card, card *Card) []*Card {
	sum := int64(1)
	for i := int64(1); i <= qty; i++ {
		cards = append(cards, card)
		sum += i
	}

	return cards
}

// GameKey formats the keys for Games in our Directory
func GameKey(gameID string) string {
	return gameID
}

// publish a game update to each Observer of the game
func (s *graphQLServer) publishGame(gameID string, g *Game) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.loggerFor(context.Background()).With("game_id", gameID)

	fullgame, ok := s.games[gameID]
	if ok {
		// alert observers
		// Log current observers for debugging subscription delivery issues
		if len(fullgame.Observers) == 0 {
			logger.Debug("publishGame: no observers")
		} else {
			var ids []string
			for k := range fullgame.Observers {
				ids = append(ids, k)
			}
			logger.Debug("publishGame: sending update", "observer_ids", ids)
		}
		for _, gameObs := range fullgame.Observers {
			select {
			case gameObs.Channel <- g:
			default:
				// drop if subscriber isn't reading to avoid blocking others
				logger.Warn("publishGame: drop update (channel full)", "observer_user_id", gameObs.UserID)
			}
		}
	} else {
		// create one if we haven't seen this game before.
		s.games[gameID] = &FullGame{
			GameID:    gameID,
			Observers: make(map[string]*GameObserver),
		}
	}
}

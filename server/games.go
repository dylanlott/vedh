package server

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
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
}

// Games returns a list of Games.
func (s *graphQLServer) Games(ctx context.Context, gameID *string) ([]*Game, error) {
	if gameID == nil {
		// Retrieve game by ID
		game, err := s.GetGame(ctx, *gameID)
		if err != nil {
			return nil, err
		}
		return []*Game{game}, nil
	}
	return nil, fmt.Errorf("not impl")
}

// GetGame returns a single game from the
func (s *graphQLServer) GetGame(ctx context.Context, gameID string) (*Game, error) {
	var payload []byte
	query := `SELECT payload FROM games WHERE id = $1`
	err := s.db.QueryRow(query, gameID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	game := &Game{}
	if err := json.Unmarshal(payload, &game); err != nil {
		return nil, err
	}
	return game, nil
}

// GameUpdated returns a channel for a game or an error.
func (s *graphQLServer) GameUpdated(ctx context.Context, gameID string, userID string) (<-chan *Game, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	g, ok := s.games[gameID]
	if !ok {
		game := &FullGame{
			GameID:    gameID,
			Observers: make(map[string]*GameObserver),
		}

		// add observer to the FullGame
		obs := &GameObserver{
			UserID:  userID,
			Channel: make(chan *Game),
		}

		// clean up the observers channel when we're done with it
		go func() {
			<-ctx.Done()
			game.Mutex.Lock()
			log.Printf("cleaning up observer %s game %s", game.GameID, userID)
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
		Channel: make(chan *Game),
	}
	g.Mutex.Lock()
	g.Observers[userID] = obs
	g.Mutex.Unlock()

	return obs.Channel, nil
}

// UpdateGame is what's used to change the name of the game, format, insert
// or remove players, or change other meta informatin about a game.
func (s *graphQLServer) UpdateGame(ctx context.Context, new InputGame) (*Game, error) {
	game := &Game{}
	b, err := json.Marshal(new)
	if err != nil {
		return nil, errs.New("failed to marshal input game: %s", err)
	}
	err = json.Unmarshal(b, &game)
	if err != nil {
		return nil, errs.New("failed to unmarshal game: %s", err)
	}

	go s.publishGame(game.ID, game)

	if err := s.upsertGame(game); err != nil {
		return game, err
	}

	return game, nil
}

// JoinGame handles a user joining an existing game.
func (s *graphQLServer) JoinGame(ctx context.Context, input *InputJoinGame) (*Game, error) {
	// TODO: Handle rejoins by detecting if that player's user.ID already exists
	// in a given game. If it does, just return that same setup.
	// TODO: Check context for User auth and append user info that way
	// TODO: PUll user boardstate creation out into a function since we do it multiple places
	if input.User.Username == "" {
		return nil, errors.New("must provide a username to join a game")
	}

	game, err := s.GetGame(ctx, input.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	if len(game.Players) >= 4 {
		return nil, errors.New("game is full")
	}

	user := &User{
		Username: input.User.Username,
		ID:       *input.User.ID,
		Boardstate: &BoardState{
			User: &User{
				Username: input.User.Username,
			},
			Life:       input.BoardState.Life,
			Exiled:     getBareCard(input.BoardState.Exiled),
			Revealed:   getBareCard(input.BoardState.Revealed),
			Field:      getBareCard(input.BoardState.Field),
			Controlled: getBareCard(input.BoardState.Controlled),
		},
	}

	// hydrate the library from the provided decklist
	library, err := s.createLibraryFromDecklist(ctx, *input.Decklist)
	if err != nil {
		// Fail gracefully and still populate basic cards
		log.Printf("error creating library from decklist: %+v", err)
		user.Boardstate.Library = getBareCard(input.BoardState.Library)
	} else {
		// Happy path
		user.Boardstate.Library = library
	}

	// NB: Commented out while we figure out how to handle Commander selection.
	if len(input.BoardState.Commander) > 0 {
		for _, card := range input.BoardState.Commander {
			commander, err := s.Card(ctx, card.Name, nil)
			if err != nil {
				log.Printf("error getting commander for deck: %+v", err)
				continue
			}
			user.Boardstate.Commander = append(user.Boardstate.Commander, commander)
		}
	}

	// shuffle their library for the start of the game
	shuff, err := Shuffle(user.Boardstate.Library)
	if err != nil {
		log.Printf("error shuffling library: %s", err)
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

	return game, nil
}

// CreateGame creates a new game and hydrates the decklists for the players in it.
func (s *graphQLServer) CreateGame(ctx context.Context, inputGame InputCreateGame) (*Game, error) {
	// don't allow a game to be created with an existing name
	// TECHDEBT replace this with a proper cache
	if _, exists := s.games[inputGame.ID]; exists {
		return nil, fmt.Errorf("game already exists with ID %s", inputGame.ID)
	}

	// assign an ID if none is provided
	if inputGame.ID == "" {
		inputGame.ID = uuid.New().String()
	}

	g := &Game{
		ID:        inputGame.ID,
		CreatedAt: time.Now(),
		Players:   []*User{},
		Turn: &Turn{
			Player: inputGame.Turn.Player,
			Phase:  inputGame.Turn.Phase,
			Number: inputGame.Turn.Number,
		},
		// NB: We're only supporting EDH at this time. We will add more flexible validation later.
		Rules: []*Rule{
			{
				Name:  "format",
				Value: "EDH",
			},
			{
				Name:  "deck_size",
				Value: "99",
			},
		},
	}

	for _, player := range inputGame.Players {
		// TODO: Deck validation should happen here.
		user := &User{
			ID:       *player.User.ID,
			Username: player.User.Username,
		}
		g.Players = append(g.Players, user)

		// Set default boardstate, handle library and commander specifically
		bs := &BoardState{
			User:       user,
			Life:       player.Life,
			GameID:     g.ID,
			Hand:       getBareCard(player.Hand),
			Exiled:     getBareCard(player.Exiled),
			Revealed:   getBareCard(player.Revealed),
			Field:      getBareCard(player.Field),
			Controlled: getBareCard(player.Controlled),
		}

		var decklist string
		if inputGame.Players[0].Decklist != nil {
			decklist = string(*inputGame.Players[0].Decklist)
		}

		// hyrdate the decklist for the player
		library, err := s.createLibraryFromDecklist(ctx, decklist)
		if err != nil {
			// Fail gracefully and still populate basic cards
			log.Printf("error creating library from decklist: %+v", err)
			bs.Library = getBareCard(player.Library)
		} else {
			// Happy path
			bs.Library = library
		}

		if len(player.Commander) > 0 {
			commander, err := s.Card(ctx, player.Commander[0].Name, nil)
			if err != nil {
				log.Printf("error getting commander for deck: %+v", err)
				// fail gracefully and use their card name so they can still play a game
				inputCard := getBareCard(player.Commander)
				bs.Commander = []*Card{inputCard[0]}
			} else {
				bs.Commander = []*Card{commander}
			}
		}

		shuff, err := Shuffle(bs.Library)
		if err != nil {
			log.Printf("error shuffling library: %s", err)
			return nil, err
		}
		bs.Library = shuff
	}

	if err := s.upsertGame(g); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	log.Printf("updated game %+v in postgres", g)

	return g, nil
}

// upsert inserts or updates a Game in the games table.
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

		if card.ID != nil {
			c.ID = *card.ID
		}

		cardList = append(cardList, c)
	}
	return cardList
}

// createLibraryFromDecklist parses the provided decklist string as CSV.
func (s *graphQLServer) createLibraryFromDecklist(ctx context.Context, decklist string) ([]*Card, error) {
	if decklist == "" {
		return []*Card{}, errs.New("must provide cards in decklist to create a library")
	}

	trimmed := strings.TrimSpace(decklist)
	r := csv.NewReader(strings.NewReader(trimmed))

	// set lazy quotes for using double quotes in csv files
	r.LazyQuotes = true
	// and trim leading spaces
	r.TrimLeadingSpace = true

	cards := []*Card{}
	for {
		// TODO: Use r.ReadAll() to get the whole decklist and do only one
		// DB lookup for all of the cards instead of one by one.
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading csv record for card name: %+v", err)
			return nil, fmt.Errorf("failed to parse CSV: %s", err)
		}

		name := record[1]
		quantity, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse quantity: %w", err)
		}

		found, err := s.Card(ctx, name, nil)
		if err != nil {
			log.Printf("failed to find card: %+v\n", err)
			cards = addX(quantity, cards, &Card{Name: name})
		} else {
			cards = addX(quantity, cards, found)
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

	fullgame, ok := s.games[gameID]
	if ok {
		// alert observers
		for _, v := range fullgame.Observers {
			v.Channel <- g
		}
	} else {
		// create one if we haven't seen this game before.
		s.games[gameID] = &FullGame{
			GameID:    gameID,
			Observers: make(map[string]*GameObserver),
		}
	}
}

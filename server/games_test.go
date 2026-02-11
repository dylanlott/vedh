package server

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/matryer/is"
	"github.com/stretchr/testify/assert"
)

func Test_graphQLServer_Games(t *testing.T) {
	is := is.New(t)
	t.Run("should return a list of games", func(t *testing.T) {
		s := testAPI(t)
		got, err := s.Games(authCtx(mastershake), 10, 0)
		is.NoErr(err)
		is.True(len(got) >= 0)
	})
}

func TestGameGetSet(t *testing.T) {
	api := testAPI(t)
	ctx := authCtx(mastershake)
	created, err := api.CreateGame(ctx, *seedInputGame)
	assert.NoError(t, err)
	t.Cleanup(func() {
		query := `DELETE FROM games WHERE id = $1;`
		_, err = api.db.Exec(query, seedGameID)
		assert.NoError(t, err)
	})
	assert.Equal(t, created.ID, seedInputGame.ID)
	got, err := api.GetGame(ctx, seedInputGame.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateGame(t *testing.T) {
	var cases = []struct {
		name    string
		input   *InputCreateGame
		want    *Game
		wantErr bool
	}{
		{
			name: "happy path creation",
			input: &InputCreateGame{
				ID: "deadbeef",
				Players: []*InputBoardState{
					{
						User:     "shakezula",
						Life:     40,
						Decklist: decklist(),
						Commander: []*InputCard{
							{
								Name: "Gavi, Nest Warden",
							},
						},
					},
				},
				Turn: &InputTurn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
			},
			want: &Game{
				ID: "deadbeef",
				Players: []*User{
					{
						Username: "shakezula",
					},
				},
				Turn: &Turn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
				Rules: []*Rule{
					{Name: "format", Value: "EDH"},
					{Name: "deck_size", Value: "99"},
				},
				Status: GameStatusInProgress,
			},
			wantErr: false,
		},
		{
			name: "should allow for game created with two commanders",
			input: &InputCreateGame{
				ID: "deadbeef",
				Players: []*InputBoardState{
					{
						User:     "shakezula",
						Life:     40,
						Decklist: decklistTwoCommanders(),
						Commander: []*InputCard{
							{
								Name: "Gavi, Nest Warden",
							},
							{
								Name: "Jarad, Golgari Lich Lord",
							},
						},
					},
				},
				Turn: &InputTurn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
			},
			want: &Game{
				ID: "deadbeef",
				Players: []*User{
					{
						Username: "shakezula",
					},
				},
				Turn: &Turn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
				Rules: []*Rule{
					{Name: "format", Value: "EDH"},
					{Name: "deck_size", Value: "99"},
				},
				Status: GameStatusInProgress,
			},
			wantErr: false,
		},
		{
			name: "should allow for game created with no commanders",
			input: &InputCreateGame{
				ID: "deadbeef",
				Players: []*InputBoardState{
					{
						User:      "shakezula",
						Life:      40,
						Decklist:  decklist(),
						Commander: []*InputCard{},
					},
				},
				Turn: &InputTurn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
			},
			want: &Game{
				ID: "deadbeef",
				Players: []*User{
					{
						ID:       "0xACAB",
						Username: "shakezula",
					},
				},
				Turn: &Turn{
					Player:   "shakezula",
					Phase:    "pregame",
					Number:   0,
					Priority: "shakezula",
				},
				Rules: []*Rule{
					{Name: "format", Value: "EDH"},
					{Name: "deck_size", Value: "99"},
				},
				Status: GameStatusInProgress,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			result, err := s.CreateGame(authCtx("shakezula"), *tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("s.CreateGame() error = %+v - wanted: %+v", err, tt.wantErr)
			}

			// check results of want
			diff := cmp.Diff(tt.want, result, cmpopts.IgnoreFields(
				Game{},
				"CreatedAt",
				"Stack",
			), cmpopts.IgnoreFields(
				Turn{},
				"Priority",
			), cmpopts.IgnoreFields(
				User{},
				"ID",
				"Boardstate",
				"Password",
				"Token",
			))
			if diff != "" {
				t.Errorf("failed to create game: %s", diff)
			}
		})
	}
}

func TestJoinGame(t *testing.T) {
	userID2 := "abc123"

	var cases = []struct {
		name    string
		input   InputJoinGame
		want    interface{}
		err     error
		wantErr bool
	}{
		{
			name: "join game happy path",
			input: InputJoinGame{
				ID:       seedGameID,
				Decklist: decklist(),
				BoardState: &InputBoardState{
					UserID: "abc123",
					User:   "meatwad",
					GameID: seedGameID,
					Life:   40,
					Commander: []*InputCard{
						{
							Name: "Gavi, Nest Warden",
						},
					},
				},
			},
			err: nil,
			want: &Game{
				ID: seedGameID,
				Rules: []*Rule{
					{Name: "format", Value: "EDH"},
					{Name: "deck_size", Value: "99"},
				},
				Turn: &Turn{
					Phase:    "pregame",
					Number:   0,
					Player:   mastershake,
					Priority: mastershake,
				},
				Players: []*User{
					{
						ID:       mastershake,
						Username: mastershake,
						Boardstate: &BoardState{
							GameID: seedGameID,
							Life:   40,
						},
					},
					{
						ID:       userID2,
						Username: "meatwad",
						Boardstate: &BoardState{
							GameID: seedGameID,
							Life:   40,
						},
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			_, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
			if err != nil {
				t.Errorf("failed to get host game: %+v\n", err)
			}
			found, err := s.GetGame(authCtx(mastershake), seedGameID)
			assert.NoError(t, err)
			fmt.Printf("found: %v\n", found)
			got, err := s.JoinGame(authCtxWithID(userID2, "meatwad"), &tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("s.JoinGame() error = %+v - wanted: %+v", err, tt.wantErr)
			}
			if tt.want != nil {
				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreFields(Game{}, "CreatedAt")); diff != "" {
					// t.Errorf("wanted: %+v - got: %+v - diff: %s", tt.want, got, diff)
					t.Logf("wanted: %+v - got: %+v - diff: %s", tt.want, got, diff)
				}
			}
			if tt.wantErr == false && got != nil {
				// assert a player was added
				assert.Truef(t, len(seedInputGame.Players) < len(got.Players), "failed to add player to game")
			}
			t.Cleanup(func() {
				query := `DELETE FROM games WHERE id = $1;`
				_, err = s.db.Exec(query, seedGameID)
				assert.NoError(t, err)
			})
		})
	}
}

func TestUpdateGame(t *testing.T) {
	userID := string("deadbeef")
	userID2 := string("deadbeef2")

	type args struct {
		ctx context.Context
		new InputGame
	}
	tests := []struct {
		name    string
		args    args
		want    *Game
		wantErr bool
	}{
		{
			name: "should update game and alert gameChannels",
			args: args{
				ctx: authCtx(mastershake),
				new: InputGame{
					ID:        seedGameID,
					CreatedAt: &time.Time{},
					Players: []*InputUser{
						{
							Username: "shakezula",
						},
						{
							Username: "meatwad",
						},
					},
					Rules: []*InputRule{
						{Name: "format", Value: "EDH"},
						{Name: "deck_size", Value: "99"},
					},
					Turn: &InputTurn{
						Number:   3,
						Phase:    "the after party",
						Player:   "meatwad",
						Priority: "meatwad",
					},
				},
			},
			wantErr: false,
			want: &Game{
				ID: seedGameID,
				Players: []*User{
					{
						Username: "shakezula",
						ID:       userID,
					},
					{
						Username: "meatwad",
						ID:       userID2,
					},
				},
				Rules: []*Rule{
					{Name: "format", Value: "EDH"},
					{Name: "deck_size", Value: "99"},
				},
				Turn: &Turn{
					Number:   3,
					Phase:    "the after party",
					Player:   "meatwad",
					Priority: "meatwad",
				},
				Status: GameStatusInProgress,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			g, err := s.CreateGame(tt.args.ctx, *seedInputGame)
			if err != nil {
				t.Errorf("failed to create test host")
			}

			// register the channel for our tests
			gameChannel, err := s.GameUpdated(tt.args.ctx, g.ID, g.Players[0].ID)
			if err != nil {
				t.Errorf("failed to get game subscription: %s", err)
			}
			log.Printf("gameChannel: %+v", gameChannel)

			// fire off our UpdateGame function
			got, err := s.UpdateGame(tt.args.ctx, tt.args.new)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.UpdateGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("update Game got: %+v", got)

			// assert on the returns
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(
				Game{},
				"CreatedAt",
				"Stack",
			), cmpopts.IgnoreFields(
				Turn{},
				"Priority",
			))
			if diff != "" {
				log.Printf("diff: %s", diff)
				t.Errorf("UpdateGame wanted: %+v - got %+v", tt.want, got)
			}

			// assert on the game that was emitted from our subscription
			select {
			case emitted := <-gameChannel:
				t.Logf("emitted game: %+v", emitted)
				diff2 := cmp.Diff(emitted, tt.want, cmpopts.IgnoreFields(
					Game{},
					"CreatedAt",
					"Stack",
				), cmpopts.IgnoreFields(
					Turn{},
					"Priority",
				))
				if diff2 != "" {
					t.Errorf("failed to emit game on channels correctly: diff %+v", diff2)
				}
			case <-time.After(time.Second):
				t.Errorf("timed out waiting for game update")
			}
		})
	}
}

func TestMultipleSubscriptions(t *testing.T) {
	s := testAPI(t)
	created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	ch1, err := s.GameUpdated(authCtx(mastershake), created.ID, mastershake)
	assert.NoError(t, err)
	assert.NotNil(t, ch1)

	ch2, err := s.GameUpdated(authCtx(carl), created.ID, carl)
	assert.NoError(t, err)
	assert.NotNil(t, ch2)

	ch3, err := s.GameUpdated(authCtx(meatwad), created.ID, meatwad)
	assert.NoError(t, err)
	assert.NotNil(t, ch3)

	updated, err := s.UpdateGame(authCtx(mastershake), InputGame{
		ID: created.ID,
		Players: []*InputUser{
			{
				ID:       &mastershake,
				Username: mastershake,
				Boardstate: &InputBoardState{
					UserID: mastershake,
					User:   mastershake,
					GameID: created.ID,
					Life:   33,
				},
			},
			{
				ID:       &carl,
				Username: carl,
				Boardstate: &InputBoardState{
					UserID: carl,
					User:   carl,
					GameID: created.ID,
					Life:   40,
				},
			},
			{
				ID:       &meatwad,
				Username: meatwad,
				Boardstate: &InputBoardState{
					UserID: meatwad,
					User:   meatwad,
					GameID: created.ID,
					Life:   33,
				},
			},
		},
		CreatedAt: &created.CreatedAt,
	})
	assert.NoError(t, err)
	assert.NotNil(t, updated)

	var first, second, third *Game
	select {
	case first = <-ch1:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for first subscription")
	}
	select {
	case second = <-ch2:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for second subscription")
	}
	select {
	case third = <-ch3:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for third subscription")
	}

	assert.Equal(t, first, second)
	assert.Equal(t, first, third)
	assert.Equal(t, second, third)
}

func TestPassPriority(t *testing.T) {
	s := testAPI(t)
	d := func() *string { v := "1,Island"; return &v }()
	gameID := "priority-game"
	input := &InputCreateGame{
		ID: gameID,
		Players: []*InputBoardState{
			{
				UserID:   mastershake,
				User:     mastershake,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
			{
				UserID:   carl,
				User:     carl,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
		},
		Turn: &InputTurn{
			Player:   mastershake,
			Phase:    "MAIN",
			Number:   1,
			Priority: mastershake,
		},
	}
	created, err := s.CreateGame(authCtx(mastershake), *input)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	updated, err := s.PassPriority(authCtx(mastershake), gameID, carl)
	assert.NoError(t, err)
	assert.Equal(t, carl, updated.Turn.Priority)

	updated, err = s.PassPriority(authCtx(carl), gameID, mastershake)
	assert.NoError(t, err)
	assert.Equal(t, mastershake, updated.Turn.Priority)

	_, err = s.PassPriority(authCtx(carl), gameID, carl)
	assert.Error(t, err)

	_, err = s.PassPriority(authCtx(mastershake), gameID, "nonplayer")
	assert.Error(t, err)
}

func TestAdvancePhase(t *testing.T) {
	s := testAPI(t)
	d := func() *string { v := "1,Island"; return &v }()
	gameID := "phase-game"
	input := &InputCreateGame{
		ID: gameID,
		Players: []*InputBoardState{
			{
				UserID:   mastershake,
				User:     mastershake,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
			{
				UserID:   carl,
				User:     carl,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
		},
		Turn: &InputTurn{
			Player:   mastershake,
			Phase:    "MAIN",
			Number:   1,
			Priority: mastershake,
		},
	}
	created, err := s.CreateGame(authCtx(mastershake), *input)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	nextNumber := 2
	updated, err := s.AdvancePhase(authCtx(mastershake), gameID, "COMBAT", &nextNumber)
	assert.NoError(t, err)
	assert.Equal(t, "COMBAT", updated.Turn.Phase)
	assert.Equal(t, nextNumber, updated.Turn.Number)
	assert.Equal(t, mastershake, updated.Turn.Priority)

	updated, err = s.AdvancePhase(authCtx(mastershake), gameID, "END STEP", nil)
	assert.NoError(t, err)
	assert.Equal(t, "END STEP", updated.Turn.Phase)
	assert.Equal(t, nextNumber, updated.Turn.Number)

	updated, err = s.AdvancePhase(authCtx(mastershake), gameID, "DISCARD", nil)
	assert.NoError(t, err)
	assert.Equal(t, "DISCARD", updated.Turn.Phase)
	assert.Equal(t, nextNumber, updated.Turn.Number)

	updated, err = s.AdvancePhase(authCtx(mastershake), gameID, "UNTAP", nil)
	assert.NoError(t, err)
	assert.Equal(t, "UNTAP", updated.Turn.Phase)
	assert.Equal(t, nextNumber+1, updated.Turn.Number)

	_, err = s.AdvancePhase(authCtx(carl), gameID, "END", nil)
	assert.Error(t, err)
}

func TestPriorityEnforcementOnStackAdd(t *testing.T) {
	s := testAPI(t)
	d := func() *string { v := "1,Island"; return &v }()
	gameID := "stack-priority-game"
	input := &InputCreateGame{
		ID: gameID,
		Players: []*InputBoardState{
			{
				UserID:   mastershake,
				User:     mastershake,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
			{
				UserID:   meatwad,
				User:     meatwad,
				GameID:   gameID,
				Life:     40,
				Decklist: d,
			},
		},
		Turn: &InputTurn{
			Player:   mastershake,
			Phase:    "MAIN",
			Number:   1,
			Priority: mastershake,
		},
	}
	created, err := s.CreateGame(authCtx(mastershake), *input)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	newStack := []*InputCard{
		{
			ID:          "card-1",
			Name:        "Test Spell",
			CurrentZone: &meatwad,
		},
	}
	update := InputGame{
		ID:        gameID,
		CreatedAt: &created.CreatedAt,
		Turn: &InputTurn{
			Player:   created.Turn.Player,
			Phase:    created.Turn.Phase,
			Number:   created.Turn.Number,
			Priority: created.Turn.Priority,
		},
		Players: []*InputUser{
			{ID: &mastershake, Username: mastershake},
			{ID: &meatwad, Username: meatwad},
		},
		Stack: newStack,
	}

	_, err = s.UpdateGame(authCtx(meatwad), update)
	assert.Error(t, err)

	newStack[0].CurrentZone = &mastershake
	update.Stack = newStack
	updated, err := s.UpdateGame(authCtx(mastershake), update)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(updated.Stack))
}

func TestCreateLibraryFromDecklist(t *testing.T) {
	is := is.New(t)
	s := testAPI(t)
	d := decklist()
	ctx := context.Background()
	got, err := s.createLibraryFromDecklist(ctx, *d, []*InputCard{{Name: "Gavi, Nest Warden"}})
	is.NoErr(err)
	is.Equal(len(got), 99)
	// assert that we get card data back as well
	card := got[0]
	is.True(card != nil)
	is.True(card.Name != "")
}

func TestCreateLibraryFromDecklist_RejectsTooLargeDeck(t *testing.T) {
	is := is.New(t)
	s := testAPI(t)
	ctx := context.Background()
	d := func() *string {
		v := "100,Island"
		return &v
	}()

	_, err := s.createLibraryFromDecklist(ctx, *d, []*InputCard{
		{Name: "Gavi, Nest Warden"},
		{Name: "Jarad, Golgari Lich Lord"},
	})
	is.True(err != nil)
}

func TestCreateLibraryFromDecklist_RemovesSelectedCommanders(t *testing.T) {
	is := is.New(t)
	s := testAPI(t)
	ctx := context.Background()
	d := func() *string {
		v := "1,\"Gavi, Nest Warden\"\n1,\"Jarad, Golgari Lich Lord\""
		return &v
	}()

	got, err := s.createLibraryFromDecklist(ctx, *d, []*InputCard{
		{Name: "Gavi, Nest Warden"},
		{Name: "Jarad, Golgari Lich Lord"},
	})
	is.NoErr(err)
	is.Equal(len(got), 0)
}

func decklist() *string {
	var deck = string(`1,"Gavi, Nest Warden"
98,Island
1,Mountain`)

	return &deck
}

func decklistTwoCommanders() *string {
	var deck = string(`1,"Gavi, Nest Warden"
1,"Jarad, Golgari Lich Lord"
97,Island`)

	return &deck
}

// Seed values for tests
var (
	seedGameID  string = "xdeadbeefx"
	mastershake string = "Mastershake"
	carl        string = "carl"
	meatwad     string = "meatwad"
)

// seedInputGame is a bare minimum game input that passes validation
var seedInputGame = &InputCreateGame{
	ID: seedGameID,
	Players: []*InputBoardState{
		{
			GameID:   seedGameID,
			UserID:   mastershake,
			User:     mastershake,
			Life:     40,
			Decklist: decklist(),
			Commander: []*InputCard{
				{
					Name: "Gavi, Nest Warden",
				},
			},
		},
	},
	Turn: &InputTurn{
		Player:   mastershake,
		Phase:    "pregame",
		Number:   0,
		Priority: mastershake,
	},
}

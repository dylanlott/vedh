package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/matryer/is"
	"github.com/openmtg/edh-go/pkg/deckimport"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

func TestCreateGame_GenericDuelFormatRoundTrips(t *testing.T) {
	s := testAPI(t)
	deck := func() *string { v := "60,Island"; return &v }()
	formatID := "GENERIC_DUEL"
	game, err := s.CreateGame(authCtx("shakezula"), InputCreateGame{
		ID:       "format-generic-duel",
		FormatID: &formatID,
		Turn:     &InputTurn{Player: "shakezula", Phase: "invalid", Number: 1, Priority: "shakezula"},
		Players:  []*InputBoardState{{UserID: "0xACAB", User: "shakezula", GameID: "format-generic-duel", Life: 0, Decklist: deck}},
	})
	assert.NoError(t, err)
	assert.Equal(t, "GENERIC_DUEL", findRuleValue(game.Rules, "format"))
	assert.Equal(t, "draw", game.Turn.Phase)
	assert.Equal(t, 20, game.Players[0].Boardstate.Life)
	assert.Equal(t, "60", findRuleValue(game.Rules, "deck_size"))
	assert.Equal(t, "20", findRuleValue(game.Rules, "starting_life"))
}

func TestGameFormatDefaults(t *testing.T) {
	s := testAPI(t)
	deck := func() *string { v := "1,Island"; return &v }()
	cases := []struct {
		name       string
		formatID   *string
		wantFormat string
		wantDeck   int
		wantLife   int
		wantPhase  string
	}{
		{name: "default format", formatID: nil, wantFormat: "EDH", wantDeck: 99, wantLife: 40, wantPhase: "pregame"},
		{name: "edh format", formatID: ptrString("EDH"), wantFormat: "EDH", wantDeck: 99, wantLife: 40, wantPhase: "pregame"},
		{name: "generic duel format", formatID: ptrString("GENERIC_DUEL"), wantFormat: "GENERIC_DUEL", wantDeck: 60, wantLife: 20, wantPhase: "draw"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gameID := "fmt-" + tt.wantFormat + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			game, err := s.CreateGame(authCtx("shakezula"), InputCreateGame{
				ID:       gameID,
				FormatID: tt.formatID,
				Turn:     &InputTurn{Player: "shakezula", Phase: "", Number: 1, Priority: "shakezula"},
				Players:  []*InputBoardState{{UserID: "0xACAB", User: "shakezula", GameID: gameID, Life: 0, Decklist: deck}},
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.wantFormat, findRuleValue(game.Rules, "format"))
			assert.Equal(t, strconv.Itoa(tt.wantDeck), findRuleValue(game.Rules, "deck_size"))
			assert.Equal(t, strconv.Itoa(tt.wantLife), findRuleValue(game.Rules, "starting_life"))
			assert.Equal(t, tt.wantLife, game.Players[0].Boardstate.Life)
			assert.Equal(t, tt.wantPhase, game.Turn.Phase)
		})
	}
}

func TestGames_CreateWritesGameCreated(t *testing.T) {
	s := testAPI(t)
	sessionID := uniqueSessionID(t)
	gameID := "game-created-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	input := telemetryCreateGameInput(gameID)
	setCreateGameSessionID(t, &input, sessionID)
	cleanupGameCreateTelemetry(t, s, gameID, sessionID)

	created, err := s.CreateGame(authCtxWithID(mastershake, mastershake), input)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	var persisted bool
	assert.NoError(t, s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM games WHERE id = $1)`, gameID).Scan(&persisted))
	assert.True(t, persisted, "game must be persisted before its conversion row is observable")

	var gotUserID, gotGameID string
	err = s.db.QueryRow(`
		SELECT user_id, game_id
		FROM product_events
		WHERE session_id = $1 AND event_name = 'game_created'
	`, sessionID).Scan(&gotUserID, &gotGameID)
	assert.NoError(t, err)
	assert.Equal(t, mastershake, gotUserID)
	assert.Equal(t, gameID, gotGameID)
	assert.Equal(t, 1, countProductEvents(t, s.db, sessionID, "game_created"))
}

func TestGames_CreateGameCreatedDedup(t *testing.T) {
	s := testAPI(t)
	sessionID := uniqueSessionID(t)
	gameID := "game-created-dedup-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	input := telemetryCreateGameInput(gameID)
	setCreateGameSessionID(t, &input, sessionID)
	cleanupGameCreateTelemetry(t, s, gameID, sessionID)

	for attempt := 0; attempt < 2; attempt++ {
		created, err := s.CreateGame(authCtxWithID(mastershake, mastershake), input)
		assert.NoError(t, err)
		assert.NotNil(t, created)
	}

	assert.Equal(t, 1, countProductEvents(t, s.db, sessionID, "game_created"))
}

func TestGames_CreateWithoutSessionWritesNoEvent(t *testing.T) {
	for _, tt := range []struct {
		name      string
		sessionID *string
	}{
		{name: "absent"},
		{name: "whitespace", sessionID: ptrString("  \t\n  ")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			lookupSessionID := uniqueSessionID(t)
			gameID := "game-created-no-session-" + tt.name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			input := telemetryCreateGameInput(gameID)
			if tt.sessionID != nil {
				setCreateGameSessionID(t, &input, *tt.sessionID)
			}
			cleanupGameCreateTelemetry(t, s, gameID, lookupSessionID)

			beforeDrops := sumDroppedCounters()
			created, err := s.CreateGame(authCtxWithID(mastershake, mastershake), input)
			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, beforeDrops, sumDroppedCounters(), "blank session must skip the event rather than record a telemetry rejection")

			var count int
			assert.NoError(t, s.db.QueryRow(`SELECT count(*) FROM product_events WHERE game_id = $1 AND event_name = 'game_created'`, gameID).Scan(&count))
			assert.Zero(t, count)
		})
	}
}

func TestGames_CreateFailureWritesNoEvent(t *testing.T) {
	s := testAPI(t)
	sessionID := uniqueSessionID(t)
	gameID := "game-created-failure-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	input := telemetryCreateGameInput(gameID)
	unknownFormat := "NOT_A_FORMAT"
	input.FormatID = &unknownFormat
	setCreateGameSessionID(t, &input, sessionID)
	cleanupGameCreateTelemetry(t, s, gameID, sessionID)

	beforeFailures := metricCounterValue(t, "vedh_game_create_total", "failure")
	created, err := s.CreateGame(authCtxWithID(mastershake, mastershake), input)
	assert.Error(t, err)
	assert.Nil(t, created)
	assert.Equal(t, 0, countProductEvents(t, s.db, sessionID, "game_created"))
	assert.Greater(t, metricCounterValue(t, "vedh_game_create_total", "failure"), beforeFailures)
}

func TestGames_CreateAndGuestMetricsHaveBoundedLabels(t *testing.T) {
	s := testAPI(t)
	_, _ = s.GuestSession(context.Background(), nil, "")

	sessionID := uniqueSessionID(t)
	gameID := "game-created-metrics-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	input := telemetryCreateGameInput(gameID)
	setCreateGameSessionID(t, &input, sessionID)
	cleanupGameCreateTelemetry(t, s, gameID, sessionID)
	created, err := s.CreateGame(authCtxWithID(mastershake, mastershake), input)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	for _, familyName := range []string{
		"vedh_guest_session_total",
		"vedh_guest_session_duration_seconds",
		"vedh_game_create_total",
		"vedh_game_create_duration_seconds",
	} {
		family := gatheredMetricFamily(t, familyName)
		assert.NotEmpty(t, family.Metric, "%s must carry at least one observation", familyName)
		for _, metric := range family.Metric {
			labels := make([]string, 0, len(metric.Label))
			for _, label := range metric.Label {
				labels = append(labels, label.GetName())
			}
			assert.ElementsMatch(t, []string{"outcome"}, labels, "%s must not expose a caller-derived identifier label", familyName)
		}
	}
}

func TestGames_CreateStoresDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName *string
	}{
		{name: "typed display name", displayName: ptrString("Dylan")},
		{name: "no display name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			userID := "create-display-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			username := "Generated Sliver " + strconv.FormatInt(time.Now().UnixNano(), 10)
			insertGameTestUser(t, s, userID, username, tt.displayName)

			gameID := "create-display-game-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			input := displayNameCreateGameInput(gameID, userID, username)
			cleanupGameTestRows(t, s, gameID, userID)

			created, err := s.CreateGame(authCtxWithID(userID, username), input)
			assert.NoError(t, err)
			if assert.Len(t, created.Players, 1) {
				assert.Equal(t, tt.displayName, created.Players[0].DisplayName)
				assert.Equal(t, username, created.Players[0].Username)
			}

			loaded, err := s.GetGame(authCtxWithID(userID, username), gameID)
			assert.NoError(t, err)
			if assert.Len(t, loaded.Players, 1) {
				assert.Equal(t, tt.displayName, loaded.Players[0].DisplayName)
				assert.Equal(t, username, loaded.Players[0].Username)
			}
		})
	}
}

func TestGames_JoinStoresDisplayName(t *testing.T) {
	s := testAPI(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	hostID, hostUsername := "join-host-"+suffix, "Host Sliver "+suffix
	joinerID, joinerUsername := "join-player-"+suffix, "Joining Sliver "+suffix
	hostDisplay, joinerDisplay := "Host Dylan", "Guest Dylan"
	insertGameTestUser(t, s, hostID, hostUsername, &hostDisplay)
	insertGameTestUser(t, s, joinerID, joinerUsername, &joinerDisplay)

	gameID := "join-display-game-" + suffix
	cleanupGameTestRows(t, s, gameID, hostID, joinerID)
	created, err := s.CreateGame(authCtxWithID(hostID, hostUsername), displayNameCreateGameInput(gameID, hostID, hostUsername))
	assert.NoError(t, err)
	assert.Equal(t, &hostDisplay, created.Players[0].DisplayName)

	deck := "1,Island"
	joined, err := s.JoinGame(authCtxWithID(joinerID, joinerUsername), &InputJoinGame{
		ID:       gameID,
		Decklist: &deck,
		BoardState: &InputBoardState{
			UserID: joinerID,
			User:   joinerUsername,
			GameID: gameID,
			Life:   40,
		},
	})
	assert.NoError(t, err)
	if assert.Len(t, joined.Players, 2) {
		assert.Equal(t, &hostDisplay, joined.Players[0].DisplayName, "joining must not disturb the host's stored label")
		assert.Equal(t, &joinerDisplay, joined.Players[1].DisplayName)
	}
}

func TestGames_LegacyPayloadHasNoDisplayName(t *testing.T) {
	s := testAPI(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	userID, username := "legacy-player-"+suffix, "Legacy Sliver "+suffix
	gameID := "legacy-display-game-" + suffix
	cleanupGameTestRows(t, s, gameID)

	legacy := &Game{
		ID:        gameID,
		CreatedAt: time.Now(),
		Players:   []*User{{ID: userID, Username: username}},
		Stack:     []*Card{},
		Status:    GameStatusInProgress,
		Turn:      &Turn{Player: username, Phase: "pregame", Priority: username},
		Rules:     []*Rule{},
	}
	payload, err := json.Marshal(legacy)
	assert.NoError(t, err)
	assert.NotContains(t, string(payload), "DisplayName", "nil display names must stay absent from legacy-compatible JSON")
	_, err = s.db.Exec(`INSERT INTO games (id, payload) VALUES ($1, $2::jsonb)`, gameID, string(payload))
	assert.NoError(t, err)

	loaded, err := s.GetGame(authCtxWithID(userID, username), gameID)
	assert.NoError(t, err)
	if assert.Len(t, loaded.Players, 1) {
		assert.Nil(t, loaded.Players[0].DisplayName)
		assert.Equal(t, username, loaded.Players[0].Username)
	}
}

func TestGames_DisplayNameIsNotAnIdentityKey(t *testing.T) {
	s := testAPI(t)
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	hostID, hostUsername := "identity-host-"+suffix, "Brave Sliver "+suffix
	joinerID, joinerUsername := "identity-joiner-"+suffix, "Clever Wizard "+suffix
	sharedDisplayName := "Dylan"
	insertGameTestUser(t, s, hostID, hostUsername, &sharedDisplayName)
	insertGameTestUser(t, s, joinerID, joinerUsername, &sharedDisplayName)

	gameID := "identity-display-game-" + suffix
	cleanupGameTestRows(t, s, gameID, hostID, joinerID)
	_, err := s.CreateGame(authCtxWithID(hostID, hostUsername), displayNameCreateGameInput(gameID, hostID, hostUsername))
	assert.NoError(t, err)
	deck := "1,Island"
	_, err = s.JoinGame(authCtxWithID(joinerID, joinerUsername), &InputJoinGame{
		ID:       gameID,
		Decklist: &deck,
		BoardState: &InputBoardState{
			UserID: joinerID,
			User:   joinerUsername,
			GameID: gameID,
			Life:   40,
		},
	})
	assert.NoError(t, err)

	for _, identity := range []struct {
		id       string
		username string
	}{
		{id: hostID, username: hostUsername},
		{id: joinerID, username: joinerUsername},
	} {
		loaded, err := s.GetGame(authCtxWithID(identity.id, identity.username), gameID)
		assert.NoError(t, err)
		var matched *User
		for _, player := range loaded.Players {
			if player.Username == identity.username {
				matched = player
			}
		}
		if assert.NotNil(t, matched, "identity lookup must continue matching Username") {
			assert.Equal(t, identity.id, matched.ID)
			assert.Equal(t, &sharedDisplayName, matched.DisplayName)
		}
	}
}

func displayNameCreateGameInput(gameID, userID, username string) InputCreateGame {
	formatID := "EDH"
	deck := "1,Island"
	return InputCreateGame{
		ID:       gameID,
		FormatID: &formatID,
		Players: []*InputBoardState{{
			UserID:   userID,
			User:     username,
			GameID:   gameID,
			Life:     40,
			Decklist: &deck,
		}},
		Turn: &InputTurn{Player: username, Phase: "pregame", Priority: username},
	}
}

func insertGameTestUser(t *testing.T, s *graphQLServer, userID, username string, displayName *string) {
	t.Helper()
	_, err := s.db.Exec(
		`INSERT INTO users (uuid, username, password, is_guest, display_name) VALUES ($1, $2, $3, true, $4)`,
		userID,
		username,
		"unused-test-password",
		displayName,
	)
	assert.NoError(t, err)
}

func cleanupGameTestRows(t *testing.T, s *graphQLServer, gameID string, userIDs ...string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID)
		for _, userID := range userIDs {
			_, _ = s.db.Exec(`DELETE FROM users WHERE uuid = $1`, userID)
		}
	})
}

func telemetryCreateGameInput(gameID string) InputCreateGame {
	formatID := "EDH"
	deck := "1,Island"
	return InputCreateGame{
		ID:       gameID,
		FormatID: &formatID,
		Players: []*InputBoardState{{
			UserID:   mastershake,
			User:     mastershake,
			GameID:   gameID,
			Life:     40,
			Decklist: &deck,
		}},
		Turn: &InputTurn{
			Player:   mastershake,
			Phase:    "pregame",
			Number:   0,
			Priority: mastershake,
		},
	}
}

// setCreateGameSessionID keeps the RED test commit buildable before gqlgen
// creates InputCreateGame.SessionID. Its first assertion is the schema gate;
// once generated, it sets the optional pointer exactly as GraphQL decoding does.
func setCreateGameSessionID(t *testing.T, input *InputCreateGame, sessionID string) {
	t.Helper()
	field := reflect.ValueOf(input).Elem().FieldByName("SessionID")
	if !field.IsValid() {
		t.Fatal("InputCreateGame.SessionID is missing; add it to schema.graphql and run make generate")
	}
	if field.Type() != reflect.TypeOf((*string)(nil)) {
		t.Fatalf("InputCreateGame.SessionID type = %v, want *string", field.Type())
	}
	value := sessionID
	field.Set(reflect.ValueOf(&value))
}

func cleanupGameCreateTelemetry(t *testing.T, s *graphQLServer, gameID, sessionID string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = s.db.Exec(`DELETE FROM product_events WHERE session_id = $1 OR game_id = $2`, sessionID, gameID)
		_, _ = s.db.Exec(`DELETE FROM games WHERE id = $1`, gameID)
	})
}

func gatheredMetricFamily(t *testing.T, name string) *dto.MetricFamily {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather Prometheus metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	t.Fatalf("metric family %s was not observed", name)
	return nil
}

func metricCounterValue(t *testing.T, familyName, outcome string) float64 {
	t.Helper()
	family := gatheredMetricFamily(t, familyName)
	for _, metric := range family.Metric {
		for _, label := range metric.Label {
			if label.GetName() == "outcome" && label.GetValue() == outcome {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
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
					{Name: "starting_life", Value: "40"},
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
					{Name: "starting_life", Value: "40"},
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
					{Name: "starting_life", Value: "40"},
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

// TestJoinGame_DefaultsLifeWhenOmitted is a regression test for CR-01: a
// joining player who omits (or zero-values) BoardState.Life must start at
// the game's format StartingLife, not at 0. Before the fix, the joining
// player's Boardstate.Life was set verbatim to the zero value, which
// game_finish.go's alivePlayerNames treats as "already eliminated" --
// letting the very next UpdateBoardState call from anyone finish the game
// as a win for the other side.
func TestJoinGame_DefaultsLifeWhenOmitted(t *testing.T) {
	userID2 := "abc123"
	s := testAPI(t)

	_, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
	assert.NoError(t, err)
	t.Cleanup(func() {
		query := `DELETE FROM games WHERE id = $1;`
		_, err := s.db.Exec(query, seedGameID)
		assert.NoError(t, err)
	})

	got, err := s.JoinGame(authCtxWithID(userID2, "meatwad"), &InputJoinGame{
		ID:       seedGameID,
		Decklist: decklist(),
		BoardState: &InputBoardState{
			UserID: userID2,
			User:   "meatwad",
			GameID: seedGameID,
			// Life deliberately omitted (zero value) to simulate a caller
			// that never set it.
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, got)

	var joined *User
	for _, p := range got.Players {
		if p != nil && p.ID == userID2 {
			joined = p
		}
	}
	if assert.NotNil(t, joined, "joining player not found in game.Players") {
		format := formatFromRules(got.Rules)
		assert.Equal(t, format.StartingLife, joined.Boardstate.Life, "joining player with omitted Life must default to format.StartingLife, not 0")
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

	deck := decklist()
	_, err = s.JoinGame(authCtxWithID(carl, carl), &InputJoinGame{
		ID:       created.ID,
		Decklist: deck,
		BoardState: &InputBoardState{
			UserID: carl,
			User:   carl,
			GameID: created.ID,
			Life:   40,
		},
	})
	assert.NoError(t, err)

	_, err = s.JoinGame(authCtxWithID(meatwad, meatwad), &InputJoinGame{
		ID:       created.ID,
		Decklist: deck,
		BoardState: &InputBoardState{
			UserID: meatwad,
			User:   meatwad,
			GameID: created.ID,
			Life:   40,
		},
	})
	assert.NoError(t, err)

	ch2, err := s.GameUpdated(authCtx(carl), created.ID, carl)
	assert.NoError(t, err)
	assert.NotNil(t, ch2)

	ch3, err := s.GameUpdated(authCtx(meatwad), created.ID, meatwad)
	assert.NoError(t, err)
	assert.NotNil(t, ch3)

	// ch1 may have buffered join-game updates from when only mastershake was subscribed.
	// Drain any queued events so all subscribers compare the same post-update payload.
drainCh1:
	for {
		select {
		case <-ch1:
		default:
			break drainCh1
		}
	}

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

func TestGetGame_RejectsNonParticipant(t *testing.T) {
	s := testAPI(t)
	created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	_, err = s.GetGame(authCtx(carl), created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestGameUpdated_RejectsNonParticipant(t *testing.T) {
	s := testAPI(t)
	created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	_, err = s.GameUpdated(authCtx(carl), created.ID, carl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
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
	parsed := deckimport.Parse(*d)
	got, err := s.createLibraryFromDecklist(ctx, &parsed, []*InputCard{{Name: "Gavi, Nest Warden"}})
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

	parsed := deckimport.Parse(*d)
	_, err := s.createLibraryFromDecklist(ctx, &parsed, []*InputCard{
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

	parsed := deckimport.Parse(*d)
	got, err := s.createLibraryFromDecklist(ctx, &parsed, []*InputCard{
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
func ptrString(v string) *string { return &v }

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

package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestBusinessMetricsRecordSignupAndLogin(t *testing.T) {
	s := testAPI(t)
	resetBusinessMetricsForTests()

	username := fmt.Sprintf("metrics-user-%d", time.Now().UnixNano())
	password := "secret123"
	if _, err := s.Signup(authCtx(username), username, password); err != nil {
		t.Fatalf("signup: %v", err)
	}
	if got := testutil.ToFloat64(vedhSignupsTotal.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected vedh signup success counter 1, got %v", got)
	}

	if _, err := s.Login(authCtx(username), username, password); err != nil {
		t.Fatalf("login: %v", err)
	}
	if got := testutil.ToFloat64(vedhLoginAttemptsTotal.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected vedh login success counter 1, got %v", got)
	}
	if got := testutil.ToFloat64(vedhUsersTotal); got != 1 {
		t.Fatalf("expected vedh total users gauge 1, got %v", got)
	}
	if got := testutil.ToFloat64(vedhDailyActiveUsersTotal); got != 1 {
		t.Fatalf("expected vedh daily active users gauge 1, got %v", got)
	}
	if got := testutil.ToFloat64(vedhWeeklyActiveUsersTotal); got != 1 {
		t.Fatalf("expected vedh weekly active users gauge 1, got %v", got)
	}
	if got := testutil.ToFloat64(vedhMonthlyActiveUsersTotal); got != 1 {
		t.Fatalf("expected vedh monthly active users gauge 1, got %v", got)
	}
}

func TestBusinessMetricsRefreshActiveUserWindowsFromPersistedActivity(t *testing.T) {
	s := testAPI(t)
	resetBusinessMetricsForTests()
	if err := refreshVedhUserActivityMetrics(s.db); err != nil {
		t.Fatalf("refresh baseline active user metrics: %v", err)
	}
	baseTotal := testutil.ToFloat64(vedhUsersTotal)
	baseDaily := testutil.ToFloat64(vedhDailyActiveUsersTotal)
	baseWeekly := testutil.ToFloat64(vedhWeeklyActiveUsersTotal)
	baseMonthly := testutil.ToFloat64(vedhMonthlyActiveUsersTotal)

	now := time.Now().UTC()
	passwordHash, err := hashPassword("secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	seedUser := func(id string, username string, activeAgo time.Duration) {
		activeAt := now.Add(-activeAgo)
		if _, err := s.db.Exec(
			`INSERT INTO "users" (uuid, username, password, timestamp, last_login_at, last_active_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			id,
			username,
			passwordHash,
			now.Add(-30*24*time.Hour),
			activeAt,
			activeAt,
		); err != nil {
			t.Fatalf("seed user %s: %v", username, err)
		}
	}

	seedUser(fmt.Sprintf("daily-%d", now.UnixNano()), fmt.Sprintf("daily_%d", now.UnixNano()), 2*time.Hour)
	seedUser(fmt.Sprintf("weekly-%d", now.UnixNano()+1), fmt.Sprintf("weekly_%d", now.UnixNano()+1), 48*time.Hour)
	seedUser(fmt.Sprintf("monthly-%d", now.UnixNano()+2), fmt.Sprintf("monthly_%d", now.UnixNano()+2), 20*24*time.Hour)

	if err := refreshVedhUserActivityMetrics(s.db); err != nil {
		t.Fatalf("refresh active user metrics: %v", err)
	}

	if got := testutil.ToFloat64(vedhUsersTotal); got != baseTotal+3 {
		t.Fatalf("expected vedh total users gauge %.0f, got %v", baseTotal+3, got)
	}
	if got := testutil.ToFloat64(vedhDailyActiveUsersTotal); got != baseDaily+1 {
		t.Fatalf("expected vedh daily active users gauge %.0f, got %v", baseDaily+1, got)
	}
	if got := testutil.ToFloat64(vedhWeeklyActiveUsersTotal); got != baseWeekly+2 {
		t.Fatalf("expected vedh weekly active users gauge %.0f, got %v", baseWeekly+2, got)
	}
	if got := testutil.ToFloat64(vedhMonthlyActiveUsersTotal); got != baseMonthly+3 {
		t.Fatalf("expected vedh monthly active users gauge %.0f, got %v", baseMonthly+3, got)
	}
}

func TestBusinessMetricsRecordGameCreateAndJoin(t *testing.T) {
	s := testAPI(t)
	resetBusinessMetricsForTests()

	ownerID := fmt.Sprintf("owner-%d", time.Now().UnixNano())
	ownerName := fmt.Sprintf("owner_%d", time.Now().UnixNano())
	joinerID := fmt.Sprintf("joiner-%d", time.Now().UnixNano())
	joinerName := fmt.Sprintf("joiner_%d", time.Now().UnixNano())
	gameID := fmt.Sprintf("metrics-game-%d", time.Now().UnixNano())
	formatID := "EDH"

	game, err := s.CreateGame(authCtxWithID(ownerID, ownerName), InputCreateGame{
		ID:       gameID,
		FormatID: &formatID,
		Turn:     &InputTurn{Player: ownerName, Phase: "pregame", Number: 0, Priority: ownerName},
		Players: []*InputBoardState{{
			UserID:   ownerID,
			User:     ownerName,
			GameID:   gameID,
			Life:     40,
			Decklist: decklist(),
			Commander: []*InputCard{{
				Name: "Gavi, Nest Warden",
			}},
		}},
	})
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	if game == nil {
		t.Fatalf("expected created game")
	}
	if got := testutil.ToFloat64(vedhGamesCreatedTotal.WithLabelValues("edh")); got != 1 {
		t.Fatalf("expected vedh games created counter 1, got %v", got)
	}

	joined, err := s.JoinGame(authCtxWithID(joinerID, joinerName), &InputJoinGame{
		ID:       gameID,
		Decklist: decklist(),
		BoardState: &InputBoardState{
			UserID:   joinerID,
			User:     joinerName,
			GameID:   gameID,
			Life:     40,
			Decklist: decklist(),
			Commander: []*InputCard{{
				Name: "Gavi, Nest Warden",
			}},
		},
	})
	if err != nil {
		t.Fatalf("join game: %v", err)
	}
	if joined == nil {
		t.Fatalf("expected joined game")
	}
	if got := testutil.ToFloat64(vedhGameJoinAttemptsTotal.WithLabelValues("success")); got != 1 {
		t.Fatalf("expected vedh join success counter 1, got %v", got)
	}
}

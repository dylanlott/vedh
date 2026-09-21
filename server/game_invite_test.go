package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameInvite_PublicProjection(t *testing.T) {
	s := testAPI(t)
	insertGameTestUser(t, s, "invite-host-id", "invite-host", ptrString("Dylan's pod"))
	created, err := s.CreateGame(
		authCtxWithID("invite-host-id", "invite-host"),
		displayNameCreateGameInput("public-invite", "invite-host-id", "invite-host"),
	)
	require.NoError(t, err)

	invite, err := s.GameInvite(context.Background(), created.ID, "invite-session")
	require.NoError(t, err)
	require.NotNil(t, invite)
	assert.Equal(t, created.ID, invite.ID)
	assert.Equal(t, "EDH", invite.Format)
	assert.Equal(t, GameStatusInProgress, invite.Status)
	assert.Equal(t, []string{"Dylan's pod"}, invite.PlayerDisplayNames)
	assert.Equal(t, 1, invite.PlayerCount)
	assert.Equal(t, gameInviteCapacity, invite.Capacity)
	assert.Equal(t, created.CreatedAt, invite.CreatedAt)

	// This exact field set is the privacy boundary. A new field on Game can
	// never flow through this independently generated type.
	inviteType := reflect.TypeOf(GameInvite{})
	fields := make([]string, 0, inviteType.NumField())
	for i := 0; i < inviteType.NumField(); i++ {
		fields = append(fields, inviteType.Field(i).Name)
	}
	assert.Equal(t, []string{
		"ID", "Format", "Status", "PlayerDisplayNames", "PlayerCount", "Capacity", "CreatedAt",
	}, fields)
}

func TestGameInvite_NotFoundIsNil(t *testing.T) {
	s := testAPI(t)
	invite, err := s.GameInvite(context.Background(), "missing-game", "invite-session")
	require.NoError(t, err)
	assert.Nil(t, invite)
}

func TestGameInvite_ReturnsFinishedAndFullStates(t *testing.T) {
	t.Run("finished", func(t *testing.T) {
		s := testAPI(t)
		created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
		require.NoError(t, err)
		_, err = s.mutateGame(context.Background(), created.ID, func(_ context.Context, game *Game) (*Game, error) {
			game.Status = GameStatusFinished
			return game, nil
		})
		require.NoError(t, err)

		invite, err := s.GameInvite(context.Background(), created.ID, "finished-session")
		require.NoError(t, err)
		require.NotNil(t, invite)
		assert.Equal(t, GameStatusFinished, invite.Status)
	})

	t.Run("full", func(t *testing.T) {
		s := testAPI(t)
		created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
		require.NoError(t, err)
		for _, username := range []string{carl, meatwad, "frylock"} {
			deck := decklist()
			_, err = s.JoinGame(authCtxWithID(username, username), &InputJoinGame{
				ID:         created.ID,
				Decklist:   deck,
				BoardState: &InputBoardState{GameID: created.ID, UserID: username, User: username, Life: 40},
			})
			require.NoError(t, err)
		}

		invite, err := s.GameInvite(context.Background(), created.ID, "full-session")
		require.NoError(t, err)
		require.NotNil(t, invite)
		assert.Equal(t, gameInviteCapacity, invite.PlayerCount)
		assert.Equal(t, gameInviteCapacity, invite.Capacity)
	})
}

func TestGameInvite_RateLimited(t *testing.T) {
	s := testAPI(t)
	s.limiter = ratelimit.NewRegistry(1, 1)

	_, err := s.GameInvite(context.Background(), "missing-game", "same-session")
	require.NoError(t, err)
	_, err = s.GameInvite(context.Background(), "missing-game", "same-session")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many invite checks")
}

func TestJoinGame_RecordsAuthoritativePlayerJoinedAfterCommit(t *testing.T) {
	s := testAPI(t)
	created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
	require.NoError(t, err)
	sessionID := "invite-join-session"
	deck := decklist()
	_, err = s.JoinGame(authCtxWithID(carl, carl), &InputJoinGame{
		ID:         created.ID,
		SessionID:  &sessionID,
		Decklist:   deck,
		BoardState: &InputBoardState{GameID: created.ID, UserID: carl, User: carl, Life: 40},
	})
	require.NoError(t, err)

	var gameID, userID, role, source, outcome string
	err = s.db.QueryRow(`
		SELECT game_id, user_id, role, source, outcome
		FROM product_events
		WHERE event_name = 'player_joined' AND session_id = $1`, sessionID,
	).Scan(&gameID, &userID, &role, &source, &outcome)
	require.NoError(t, err)
	assert.Equal(t, created.ID, gameID)
	assert.Equal(t, carl, userID)
	assert.Equal(t, "invitee", role)
	assert.Equal(t, "invite", source)
	assert.Equal(t, "success", outcome)
	assert.Equal(t, 1, countProductEvents(t, s.db, sessionID, "player_joined"))
}

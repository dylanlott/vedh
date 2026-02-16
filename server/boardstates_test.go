package server

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBoardState(t *testing.T) {
	type args struct {
		ctx   context.Context
		input InputBoardState
	}
	tests := []struct {
		name    string
		args    args
		want    *BoardState
		wantErr bool
	}{
		{
			name: "should update boardstate and notify listeners",
			args: args{
				ctx: authCtx(mastershake),
				input: InputBoardState{
					GameID: seedGameID,
					UserID: mastershake,
					User:   mastershake,
					Life:   38,
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: false,
			want: &BoardState{
				GameID: seedGameID,
				UserID: mastershake,
				User:   mastershake,
				Life:   38,
				Commander: []*Card{
					{Name: "Gavi, Nest Warden"},
				},
			},
		},
		{
			name: "should return an error if game does not exist",
			args: args{
				ctx: authCtx("shakezula"),
				input: InputBoardState{
					GameID: "doesnotexist",
					UserID: "shakezula",
					User:   "shakezula",
					Life:   38,
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// assemble
			s := testAPI(t)
			_, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
			if err != nil {
				t.Errorf("failed to create dummy game: %s", err)
			}

			// register listener
			bch, err := s.BoardstateUpdated(
				tt.args.ctx,
				"TestUpdateBoardStateObserver",
				tt.args.input.UserID,
			)
			assert.Equal(t, err, nil)
			assert.NotNil(t, bch)

			time.Sleep(time.Millisecond * 500)

			got, err := s.UpdateBoardState(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.UpdateBoardState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("graphQLServer.UpdateBoardState() = %v, want %v", got, tt.want)
			}

			if tt.wantErr == false {
				select {
				case boardstate := <-bch:
					if !reflect.DeepEqual(boardstate, tt.want) {
						diff := cmp.Diff(boardstate, tt.want)
						t.Logf("%s", diff)
						t.Errorf("UpdateBoardState failed to notify listeners")
					} else {
						t.Logf("successfully received boardstate from update: %+v", boardstate)
					}
				case <-time.After(time.Second):
					t.Errorf("timed out waiting for boardstate update")
				}
			}
		})
	}
}

func TestUpdateBoardState_AutoFinishRules(t *testing.T) {
	t.Run("single-player game does not auto-finish when one player is alive", func(t *testing.T) {
		s := testAPI(t)
		created, err := s.CreateGame(authCtx(mastershake), *seedInputGame)
		assert.NoError(t, err)
		assert.NotNil(t, created)

		_, err = s.UpdateBoardState(authCtx(mastershake), InputBoardState{
			GameID: created.ID,
			UserID: mastershake,
			User:   mastershake,
			Life:   39,
		})
		assert.NoError(t, err)

		got, err := s.GetGame(authCtx(mastershake), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, GameStatusInProgress, got.Status)
		assert.Nil(t, got.Result)
	})

	t.Run("multiplayer game auto-finishes when only one player remains alive", func(t *testing.T) {
		s := testAPI(t)
		input := InputCreateGame{
			ID: "xdeadbeefx-autofinish-rules",
			Players: []*InputBoardState{
				{
					GameID:   "xdeadbeefx-autofinish-rules",
					UserID:   mastershake,
					User:     mastershake,
					Life:     40,
					Decklist: decklist(),
				},
				{
					GameID:   "xdeadbeefx-autofinish-rules",
					UserID:   carl,
					User:     carl,
					Life:     0,
					Decklist: decklist(),
				},
			},
			Turn: &InputTurn{
				Player:   mastershake,
				Phase:    "pregame",
				Number:   0,
				Priority: mastershake,
			},
		}

		created, err := s.CreateGame(authCtx(mastershake), input)
		assert.NoError(t, err)
		assert.NotNil(t, created)

		_, err = s.UpdateBoardState(authCtx(mastershake), InputBoardState{
			GameID: created.ID,
			UserID: mastershake,
			User:   mastershake,
			Life:   39,
		})
		assert.NoError(t, err)

		got, err := s.GetGame(authCtx(mastershake), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, GameStatusFinished, got.Status)
		if assert.NotNil(t, got.Result) {
			assert.Equal(t, GameResultWin, *got.Result)
		}
	})
}

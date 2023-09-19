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
	userID := string("0xACAB")

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
				ctx: context.Background(),
				input: InputBoardState{
					GameID: seedGameID,
					User: &InputUser{
						Username: "shakezula",
						ID:       &userID,
					},
					Life: 38, // test life edits
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: false,
			want: &BoardState{
				GameID: seedGameID,
				User:   "shakezula",
				Life:   38,
				Commander: []*Card{
					{Name: "Gavi, Nest Warden"},
				},
			},
		},
		{
			name: "should return an error if game does not exist",
			args: args{
				ctx: context.Background(),
				input: InputBoardState{
					GameID: "doesnotexist",
					User: &InputUser{
						Username: "shakezula",
						ID:       &userID,
					},
					Life: 38, // test life edits
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
			s := testAPI(t)
			_, err := s.CreateGame(tt.args.ctx, *seedInputGame)
			if err != nil {
				t.Errorf("failed to create dummy game: %s", err)
			}

			bch, err := s.BoardstateUpdated(
				tt.args.ctx,
				"TestUpdateBoardStateObserver",
				*tt.args.input.User.ID,
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

			// assert on boardstate notifications if no error is expected
			if tt.wantErr == false {
				boardstate := <-bch
				if !reflect.DeepEqual(boardstate, tt.want) {
					diff := cmp.Diff(boardstate, tt.want)
					t.Logf("%s", diff)
					t.Errorf("UpdateBoardState failed to notify listeners")
				} else {
					t.Logf("successfully received boardstate from update: %+v", boardstate)
				}
			}
		})
	}
}

// TECHDEBT write a test for multiple observers for one game
// func TestMultipleObservers(t *testing.T) {
// 	s := testAPI(t)
// 	ctx := context.Background()
// 	ctx, done := context.WithCancel(ctx)
// 	created, err := s.CreateGame(ctx, *seedInputGame)
// 	assert.NoError(t, err)

// 	// NB: these should both be observing the same user from *seedInputGame
// 	ch1, err := s.BoardstateUpdated(ctx, "testobs1", seedUserID)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, ch1)
// 	ch2, err := s.BoardstateUpdated(ctx, "testobs2", seedUserID)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, ch2)

// 	created.Turn.Number = 3
// }

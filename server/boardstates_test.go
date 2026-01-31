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
				ctx: authCtx("shakezula"),
				input: InputBoardState{
					GameID: seedGameID,
					UserID: "shakezula",
					User:   "shakezula",
					Life:   38,
					Commander: []*InputCard{
						{Name: "Gavi, Nest Warden"},
					},
				},
			},
			wantErr: false,
			want: &BoardState{
				GameID: seedGameID,
				UserID: "shakezula",
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

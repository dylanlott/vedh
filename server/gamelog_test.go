package server

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func Test_pgLogger_Add(t *testing.T) {
	s := testAPI(t)
	type fields struct {
		db *sql.DB
		s  *graphQLServer
	}
	type args struct {
		ctx   context.Context
		event Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should error if nil payload is provided",
			fields: fields{
				db: s.db,
				s:  s,
			},
			args: args{
				ctx: context.Background(),
				event: Event{
					Payload: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "should log a game event and handle a list of cards",
			fields: fields{
				db: s.db,
				s:  s,
			},
			args: args{
				ctx: context.Background(),
				event: Event{
					Payload: map[string]interface{}{
						"boardstate": BoardState{
							User: "shakezula",
							Life: 40,
						},
						"timestamp": time.Now(),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &pgLogger{
				db: tt.fields.db,
			}
			if err := g.Add(tt.args.ctx, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("pgLogger.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

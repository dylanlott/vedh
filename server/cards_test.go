package server

import (
	"context"
	"reflect"
	"testing"
)

func Test_graphQLServer_Search(t *testing.T) {
	type args struct {
		ctx           context.Context
		name          string
		colors        []*string
		colorIdentity []*string
		keywords      []*string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Card
		wantErr bool
	}{
		{
			name: "should return a list of cards with a similar",
			args: args{
				ctx:  context.TODO(),
				name: "Jarad, Golgari Lich Lord",
			},
			want: []*Card{
				{Name: "Jarad, Golgari Lich Lord"},
				{Name: "Jarad, Golgari Lich Lord"},
				{Name: "Jarad, Golgari Lich Lord"},
				{Name: "Jarad, Golgari Lich Lord"},
				{Name: "Jarad, Golgari Lich Lord"},
			},
		},
		{
			name: "should handle apostrophes",
			args: args{
				ctx:  context.TODO(),
				name: "Kykar, Wind's Fury",
			},
			want: []*Card{
				{Name: "Kykar, Wind's Fury"},
				{Name: "Kykar, Wind's Fury"},
				{Name: "Kykar, Wind's Fury"},
			},
		},
		{
			name: "should handle rough / tumble style syntax",
			args: args{
				ctx:  context.TODO(),
				name: "Rough // Tumble",
			},
			want: []*Card{
				{Name: "Rough // Tumble"},
				{Name: "Rough // Tumble"},
				{Name: "Rough // Tumble"},
				{Name: "Rough // Tumble"},
				{Name: "Rough // Tumble"},
				{Name: "Rough // Tumble"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			got, err := s.Search(tt.args.ctx, &tt.args.name, tt.args.colors, tt.args.colorIdentity, tt.args.keywords)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// compare lengths of returned results as a rough heuristic for success.
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("graphQLServer.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

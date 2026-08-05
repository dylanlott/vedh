package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			if len(got) < len(tt.want) {
				t.Errorf("graphQLServer.Search() = %v, want at least %v", got, tt.want)
			}
		})
	}
}

func Test_graphQLServer_Card(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
		id   *string
	}
	tests := []struct {
		name    string
		args    args
		want    *Card
		wantErr bool
	}{
		{
			name: "should return a card",
			args: args{
				ctx:  context.Background(),
				name: "Kykar, Wind's Fury",
			},
			wantErr: false,
			want: &Card{
				Name: "Kykar, Wind's Fury",
				ID:   "31511",
			},
		},
		{
			name: "should handle // syntax",
			args: args{
				ctx:  context.Background(),
				name: "Rough // Tumble",
			},
			wantErr: false,
			want: &Card{
				Name: "Rough // Tumble",
				ID:   "43091",
			},
		},
		{
			name: "should handle mdfc cards",
			args: args{
				ctx:  context.Background(),
				name: "Beyeen Veil",
			},
			wantErr: false,
			want: &Card{
				Name: "Beyeen Veil // Beyeen Coast",
				ID:   "62364",
			},
		},
		{
			name: "should handle apostrophes in card names",
			args: args{
				ctx:  context.Background(),
				name: "Cho-Manno's Blessing",
			},
			wantErr: false,
			want: &Card{
				Name: "Cho-Manno's Blessing",
				ID:   "37396",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			got, err := s.Card(tt.args.ctx, tt.args.name, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.Card() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(Card{},
				"ID", "Colors", "Cmc", "UUID", "Power", "Toughness", "Subtypes",
				"Supertypes", "Types", "Text", "Tcgid", "ScryfallID")); diff != "" {
				t.Logf("%s", diff)
				t.Fail()
			}
		})
	}
}

func Test_graphQLServer_Cards(t *testing.T) {
	type args struct {
		ctx  context.Context
		list []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Card
		wantErr bool
	}{
		{
			name: "should return a card",
			args: args{
				ctx: context.Background(),
				list: []string{
					"Kykar, Wind's Fury",
					"Jarad, Golgari Lich Lord",
				},
			},
			want: []*Card{
				{Name: "Kykar, Wind's Fury"},
				{Name: "Jarad, Golgari Lich Lord"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testAPI(t)
			got, err := s.Cards(tt.args.ctx, tt.args.list)
			if (err != nil) != tt.wantErr {
				t.Errorf("graphQLServer.Cards() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(Card{},
				"ID", "Colors", "Cmc", "UUID", "Power", "Toughness", "Subtypes",
				"Supertypes", "Types", "Text", "Tcgid", "ScryfallID")); diff != "" {
				t.Logf("%s", diff)
				t.Fail()
			}
		})
	}
}

// TestCards_LowerCasedNeedleMatchesMixedCaseName proves the WHERE clause
// change is strictly wider than the previous case-sensitive comparison: a
// fully lower-cased needle now matches a stored name whose letter case
// differs, which the old `name = ANY($1)` comparison could never do.
func TestCards_LowerCasedNeedleMatchesMixedCaseName(t *testing.T) {
	s := testAPI(t)

	got, err := s.Cards(context.Background(), []string{"kykar, wind's fury"})
	if err != nil {
		t.Fatalf("graphQLServer.Cards() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Cards() returned %d results, want 1", len(got))
	}
	if got[0].Name != "Kykar, Wind's Fury" {
		t.Fatalf("Cards() = %+v, want a resolved card named %q", got[0], "Kykar, Wind's Fury")
	}
}

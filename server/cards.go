package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/zeebo/errs"
)

// pgCard is a model used to reflect our internal Postgres model of a card.
// We must serialize it to a *Card type provided by our GraphQL model before
// returning it.
type PGCard struct {
	Id                     string
	Artist                 string
	Asciiname              string
	Availability           string
	Bordercolor            string
	CardKingdomId          string
	ColorIdentity          string
	Colors                 string
	ConvertedManaCost      string
	FaceConvertedManaCost  string
	FaceManaValue          string
	FlavorName             string
	FlavorText             string
	Keywords               string
	MtgJsonv4Id            string
	Name                   string
	Number                 string
	OriginalText           string
	OriginalType           string
	Power                  string
	ScryfallID             string
	ScryfallIllustrationId string
	ScryfallOracleId       string
	SetCode                string
	Side                   string
	Subtypes               string
	Supertypes             string
	TcgplayerProductId     string
	Text                   string
	Toughness              string
	Type                   string
	Types                  string
	Uuid                   string
}

// Card returns a single most-recently entered card by ID from the database that
// exactly matches the provided `name`.
// ! This does not currently respect the `id` parameter when passed.
func (s *graphQLServer) Card(ctx context.Context, name string, id *string) (*Card, error) {
	// This will grab the single most recently inserted card that matches the name provided.
	row := s.db.QueryRow(`SELECT name, id, colors, convertedmanacost, types,
		power, toughness, text, subtypes, supertypes, tcgplayerproductid,
		scryfallid, uuid 
		FROM cards 
		WHERE name = $1 
		OR facename = $1
		ORDER BY id ASC 
		LIMIT 1;`, name)
	if row.Err() != nil {
		return nil, fmt.Errorf("failed to query card %s: %w", name, row.Err())
	}
	c := &PGCard{}
	err := row.Scan(&c.Name, &c.Id, &c.Colors, &c.ConvertedManaCost, &c.Types,
		&c.Power, &c.Toughness, &c.Text, &c.Subtypes, &c.Supertypes,
		&c.TcgplayerProductId, &c.ScryfallID, &c.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to scan card %s: %w", name, err)
	}
	return &Card{
		Name:       c.Name,
		ID:         c.Id,
		Colors:     &c.Colors,
		Cmc:        &c.ConvertedManaCost,
		Types:      &c.Types,
		Power:      &c.Power,
		Toughness:  &c.Toughness,
		Text:       &c.Text,
		Subtypes:   &c.Subtypes,
		Supertypes: &c.Supertypes,
		Tcgid:      &c.TcgplayerProductId,
		ScryfallID: &c.ScryfallID,
		UUID:       &c.Uuid,
	}, nil
}

// Cards optimizes for batched lookups or parameterized searches.
func (s *graphQLServer) Cards(ctx context.Context, list []string) ([]*Card, error) {
	combined := []error{}
	cards := []*Card{}
	for _, card := range list {
		c, err := s.Card(ctx, card, nil)
		if err != nil {
			combined = append(combined, err)
		}
		cards = append(cards, c)
	}
	return cards, errs.Combine(combined...)
}

// Search will search for card names in the database.
func (s *graphQLServer) Search(
	ctx context.Context,
	name *string,
	colors []*string,
	colorIdentity []*string,
	keywords []*string,
) ([]*Card, error) {
	// must have at least a name to search with
	if name == nil {
		return nil, nil
	}

	// query the db for it
	rows, err := s.db.Query("SELECT id, name, colors FROM cards WHERE name LIKE $1", name)
	if err != nil {
		log.Printf("error querying database: %s", err)
		return nil, errs.New("failed to search db: %s", err)
	}

	// format the query results if we have anything
	cards := []*Card{}
	for rows.Next() {
		var (
			id     *int
			name   *string
			colors *string
		)

		if err := rows.Scan(&id, &name, &colors); err != nil {
			log.Printf("ERROR: failed to scan card into struct: %s", err)
			continue
		}

		parsedID := strconv.Itoa(*id)
		card := &Card{
			ID:     parsedID,
			Name:   *name,
			Colors: colors,
		}

		cards = append(cards, card)
	}

	return cards, nil
}

//
// Shuffle functions
//

// Shuffler type defines the interface for a given Shuffle function to fulfill.
type Shuffler func(deck []*Card) ([]*Card, error)

// Shuffle will apply a Knuth shuffle to the decklist.
func Shuffle(deck []*Card) ([]*Card, error) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck, nil
}

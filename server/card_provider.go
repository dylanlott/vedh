package server

import (
	"context"
	"fmt"
	"strings"
)

// CardLookupQuery defines the normalized single-card lookup contract across games.
type CardLookupQuery struct {
	Name string
	ID   *string
}

// CardSearchQuery defines shared search inputs while allowing providers to own semantics.
type CardSearchQuery struct {
	Name          *string
	Colors        []*string
	ColorIdentity []*string
	Keywords      []*string
}

// CardImage describes a resolved image or art asset for a card.
type CardImage struct {
	URL    string
	Source string
}

// CardSourceMetadata captures how a provider gets and interprets its source data.
type CardSourceMetadata struct {
	GameType          string
	DisplayName       string
	CanonicalSource   string
	ImportPath        string
	SearchSemantics   string
	ImageSource       string
	AvailabilityNotes string
	LicenseNotes      string
}

// CardProvider defines the game-specific boundary for card data behavior.
type CardProvider interface {
	LookupCard(ctx context.Context, server *graphQLServer, query CardLookupQuery) (*Card, error)
	SearchCards(ctx context.Context, server *graphQLServer, query CardSearchQuery) ([]*Card, error)
	ResolveImage(ctx context.Context, server *graphQLServer, card *Card) (*CardImage, error)
	SourceMetadata() CardSourceMetadata
}

// CardProviderRegistry resolves provider behavior by game type rather than assuming MTG.
type CardProviderRegistry struct {
	providers map[string]CardProvider
}

func NewCardProviderRegistry() *CardProviderRegistry {
	registry := &CardProviderRegistry{providers: map[string]CardProvider{}}
	registry.Register("mtg", mtgCardProvider{})
	registry.Register("edh", mtgCardProvider{})
	return registry
}

func (r *CardProviderRegistry) Register(gameType string, provider CardProvider) {
	if r == nil || provider == nil {
		return
	}
	key := normalizeGameType(gameType)
	if key == "" {
		return
	}
	r.providers[key] = provider
}

func (r *CardProviderRegistry) ProviderForGameType(gameType string) (CardProvider, error) {
	if r == nil {
		return nil, fmt.Errorf("card provider registry is nil")
	}
	key := normalizeGameType(gameType)
	provider, ok := r.providers[key]
	if !ok {
		return nil, fmt.Errorf("no card provider registered for game type %q", gameType)
	}
	return provider, nil
}

func normalizeGameType(gameType string) string {
	return strings.ToLower(strings.TrimSpace(gameType))
}

type mtgCardProvider struct{}

func (mtgCardProvider) LookupCard(ctx context.Context, server *graphQLServer, query CardLookupQuery) (*Card, error) {
	return server.Card(ctx, query.Name, query.ID)
}

func (mtgCardProvider) SearchCards(ctx context.Context, server *graphQLServer, query CardSearchQuery) ([]*Card, error) {
	return server.Search(ctx, query.Name, query.Colors, query.ColorIdentity, query.Keywords)
}

func (mtgCardProvider) ResolveImage(ctx context.Context, server *graphQLServer, card *Card) (*CardImage, error) {
	_ = ctx
	_ = server
	if card == nil || card.ScryfallID == nil || strings.TrimSpace(*card.ScryfallID) == "" {
		return nil, nil
	}
	return &CardImage{
		URL:    fmt.Sprintf("https://api.scryfall.com/cards/%s?format=image&version=normal", strings.TrimSpace(*card.ScryfallID)),
		Source: "scryfall",
	}, nil
}

func (mtgCardProvider) SourceMetadata() CardSourceMetadata {
	return CardSourceMetadata{
		GameType:          "mtg",
		DisplayName:       "Magic: The Gathering / EDH",
		CanonicalSource:   "MTGJSON All Printings + Scryfall identifiers",
		ImportPath:        "persistence.ImportAllPrintingsJSON",
		SearchSemantics:   "Name/facename lookup backed by the cards table with MTG-style text and color filters",
		ImageSource:       "Scryfall by scryfallId",
		AvailabilityNotes: "Bulk import is already local-database friendly",
		LicenseNotes:      "Existing app behavior relies on imported MTGJSON data plus Scryfall-linked identifiers",
	}
}

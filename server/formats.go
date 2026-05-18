package server

import "strconv"

type GameFormatZone struct {
	ID            string `json:"ID"`
	Label         string `json:"Label"`
	Visibility    string `json:"Visibility"`
	Kind          string `json:"Kind"`
	SupportsCards bool   `json:"SupportsCards"`
}

type LayoutDefinition struct {
	Primary []string `json:"Primary,omitempty"`
	Shared  []string `json:"Shared,omitempty"`
	Meta    []string `json:"Meta,omitempty"`
}

type GameFormat struct {
	ID               string           `json:"ID"`
	Name             string           `json:"Name"`
	StartingLife     int              `json:"StartingLife"`
	DefaultDeckSize  int              `json:"DefaultDeckSize"`
	Zones            []*GameFormatZone `json:"Zones"`
	PhaseSequence    []string         `json:"PhaseSequence"`
	Layout           *LayoutDefinition `json:"Layout,omitempty"`
	CommanderEnabled bool             `json:"CommanderEnabled"`
}

var formatRegistry = map[string]*GameFormat{
	"EDH": {
		ID:               "EDH",
		Name:             "Commander",
		StartingLife:     40,
		DefaultDeckSize:  99,
		CommanderEnabled: true,
		PhaseSequence:    []string{"pregame", "untap", "upkeep", "draw", "main", "combat", "main2", "end"},
		Zones: []*GameFormatZone{
			{ID: "commander", Label: "Commander", Visibility: "public", Kind: "stacked", SupportsCards: true},
			{ID: "library", Label: "Library", Visibility: "count_only", Kind: "library", SupportsCards: true},
			{ID: "hand", Label: "Hand", Visibility: "private", Kind: "hand", SupportsCards: true},
			{ID: "battlefield", Label: "Battlefield", Visibility: "public", Kind: "grid", SupportsCards: true},
			{ID: "graveyard", Label: "Graveyard", Visibility: "public", Kind: "stacked", SupportsCards: true},
			{ID: "exiled", Label: "Exile", Visibility: "public", Kind: "stacked", SupportsCards: true},
			{ID: "revealed", Label: "Revealed", Visibility: "public", Kind: "stacked", SupportsCards: true},
			{ID: "controlled", Label: "Controlled", Visibility: "public", Kind: "grid", SupportsCards: true},
		},
	},
	"GENERIC_DUEL": {
		ID:               "GENERIC_DUEL",
		Name:             "Generic Duel",
		StartingLife:     20,
		DefaultDeckSize:  60,
		CommanderEnabled: false,
		PhaseSequence:    []string{"draw", "main", "battle", "end"},
		Zones: []*GameFormatZone{
			{ID: "deck", Label: "Deck", Visibility: "count_only", Kind: "library", SupportsCards: true},
			{ID: "hand", Label: "Hand", Visibility: "private", Kind: "hand", SupportsCards: true},
			{ID: "field", Label: "Field", Visibility: "public", Kind: "grid", SupportsCards: true},
			{ID: "discard", Label: "Discard", Visibility: "public", Kind: "stacked", SupportsCards: true},
			{ID: "banished", Label: "Banished", Visibility: "public", Kind: "stacked", SupportsCards: true},
		},
	},
}

func LookupFormat(id string) (*GameFormat, bool) {
	if id == "" {
		return DefaultFormat(), true
	}
	format, ok := formatRegistry[id]
	if !ok {
		return nil, false
	}
	return format, true
}

func DefaultFormat() *GameFormat {
	return formatRegistry["EDH"]
}

func findRuleValue(rules []*Rule, name string) string {
	for _, rule := range rules {
		if rule != nil && rule.Name == name {
			return rule.Value
		}
	}
	return ""
}

func ensureFormatRules(game *Game, format *GameFormat) {
	if game == nil || format == nil {
		return
	}
	game.Rules = upsertRule(game.Rules, "format", format.ID)
	game.Rules = upsertRule(game.Rules, "deck_size", strconv.Itoa(format.DefaultDeckSize))
	game.Rules = upsertRule(game.Rules, "starting_life", strconv.Itoa(format.StartingLife))
	if game.Turn != nil {
		game.Turn.Phase = normalizeTurnPhase(format, game.Turn.Phase)
	}
	for _, player := range game.Players {
		if player == nil || player.Boardstate == nil {
			continue
		}
		if player.Boardstate.Life == 0 {
			player.Boardstate.Life = format.StartingLife
		}
	}
}

func normalizeTurnPhase(format *GameFormat, phase string) string {
	if format == nil {
		format = DefaultFormat()
	}
	if len(format.PhaseSequence) == 0 {
		return phase
	}
	if phase == "" {
		return format.PhaseSequence[0]
	}
	for _, candidate := range format.PhaseSequence {
		if phase == candidate {
			return phase
		}
	}
	return format.PhaseSequence[0]
}

func upsertRule(rules []*Rule, name string, value string) []*Rule {
	for _, rule := range rules {
		if rule != nil && rule.Name == name {
			rule.Value = value
			return rules
		}
	}
	return append(rules, &Rule{Name: name, Value: value})
}

func formatFromRules(rules []*Rule) *GameFormat {
	formatID := findRuleValue(rules, "format")
	format, ok := LookupFormat(formatID)
	if !ok {
		return DefaultFormat()
	}
	return format
}

func formatDefinitions() []*GameFormat {
	return []*GameFormat{formatRegistry["EDH"], formatRegistry["GENERIC_DUEL"]}
}

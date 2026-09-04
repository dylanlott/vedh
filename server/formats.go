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
	ID               string            `json:"ID"`
	Name             string            `json:"Name"`
	StartingLife     int               `json:"StartingLife"`
	DefaultDeckSize  int               `json:"DefaultDeckSize"`
	Zones            []*GameFormatZone `json:"Zones"`
	PhaseSequence    []string          `json:"PhaseSequence"`
	Layout           *LayoutDefinition `json:"Layout,omitempty"`
	CommanderEnabled bool              `json:"CommanderEnabled"`
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
	// Deliberately no Turn.Phase normalization and no per-player Life
	// defaulting here (fix(01-05), see SUMMARY "Assigned Test Fix"): this
	// function is called by ensureGameDefaults on every load of an
	// EXISTING game (JoinGame, UpdateBoardState, AdvancePhase, GetGame,
	// UpdateGame, win_claim), not only at creation. Re-normalizing the
	// turn phase on every load silently discarded any phase name outside
	// the format's registered PhaseSequence (this tracker deliberately
	// does not enforce turn structure -- PROJECT.md: "Full Magic rules
	// enforcement...not what a tracker is for"), and re-defaulting a
	// player's Life to the format's starting life whenever it read 0
	// meant a player who legitimately reached 0 life during play had
	// their life silently reset to full the next time ANYONE called
	// UpdateBoardState -- before the auto-finish check even ran. Both
	// were real regressions from commit b1ac894, not intentional
	// behavior this function is supposed to re-apply on every read.
	// CreateGame is still where a brand-new game's Turn.Phase and initial
	// player Life get their one-time defaulting (games.go).
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

package server

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/openmtg/edh-go/pkg/deckimport"
	"github.com/zeebo/errs"
)

// PreviewDeck is the resolver for the previewDeck field. previewDeck is a
// Mutation field per the locked api-contract, so make generate binds it on
// the mutation resolver interface; this method still lives on
// *graphQLServer and its name is unchanged.
//
// publicQueries in server/authz.go is not an operation-level gate — it is
// consulted exactly once, at server/users.go:115, against a hardcoded
// literal that is not one of its members — so this resolver is public by
// virtue of not calling requireAuth, and must not be added to that map,
// where it would be dead code that reads like a security control.
func (s *graphQLServer) PreviewDeck(ctx context.Context, input InputDeckImport) (*DeckPreview, error) {
	start := time.Now()

	hasText := input.Text != nil && strings.TrimSpace(*input.Text) != ""
	hasURL := input.SourceURL != nil && strings.TrimSpace(*input.SourceURL) != ""

	switch {
	case !hasText && !hasURL:
		return s.finishPreviewDeck(ctx, input, start, deckimport.SourceUnknown, 0,
			blockedPreview(deckimport.SourceUnknown, "Paste a decklist or provide a deck link to continue."))
	case hasText && hasURL:
		return s.finishPreviewDeck(ctx, input, start, deckimport.SourceUnknown, 0,
			blockedPreview(deckimport.SourceUnknown, "Provide either a pasted decklist or a deck link, not both."))
	case hasURL:
		// A non-empty sourceURL returns a normalized "deck links are not
		// enabled" blocking error in this task; plan 01-06 replaces this
		// branch with the real provider fetch.
		return s.finishPreviewDeck(ctx, input, start, deckimport.SourceUnknown, 0,
			blockedPreview(deckimport.SourceUnknown, "Deck links are not enabled yet. Paste your decklist instead."))
	}

	// This is the single canonical parse site for the whole codebase.
	parsed := deckimport.Parse(*input.Text)
	preview, cardCount := s.buildDeckPreview(ctx, parsed)
	return s.finishPreviewDeck(ctx, input, start, parsed.Source, cardCount, preview)
}

// blockedPreview builds a DeckPreview whose only content is one blocking
// error, used for the input-shape validation cases that never reach the
// parser at all.
func blockedPreview(source deckimport.SourceType, message string) *DeckPreview {
	return &DeckPreview{
		SourceType:          string(source),
		CardCount:           0,
		Entries:             []*DeckPreviewEntry{},
		CommanderCandidates: []*Card{},
		Unresolved:          []*DeckImportIssue{},
		Warnings:            []string{},
		CanContinue:         false,
		BlockingErrors:      []string{message},
	}
}

// buildDeckPreview resolves a ParsedDeck's entries against the cards table
// and assembles the DeckPreview the resolver returns. It never fails: a
// lookup error is logged and treated as "nothing resolved" rather than
// returned to the caller, matching D-01's "never block the player."
func (s *graphQLServer) buildDeckPreview(ctx context.Context, parsed deckimport.ParsedDeck) (*DeckPreview, int) {
	if len(parsed.BlockingErrors) > 0 {
		blocking := make([]string, 0, len(parsed.BlockingErrors))
		for _, be := range parsed.BlockingErrors {
			blocking = append(blocking, be.Message)
		}
		return &DeckPreview{
			SourceType:          string(parsed.Source),
			CardCount:           0,
			Entries:             []*DeckPreviewEntry{},
			CommanderCandidates: []*Card{},
			Unresolved:          []*DeckImportIssue{},
			Warnings:            []string{},
			CanContinue:         false,
			BlockingErrors:      blocking,
		}, 0
	}

	names := make([]string, 0, len(parsed.Entries))
	seen := map[string]struct{}{}
	for _, e := range parsed.Entries {
		key := strings.ToLower(strings.TrimSpace(e.Name))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, e.Name)
	}

	lookup, lookupErr := s.lookupCardsByName(ctx, names)
	if lookupErr != nil {
		s.loggerFor(ctx).Warn("deck import card lookup failed", "err", lookupErr)
	}

	entries := make([]*DeckPreviewEntry, 0, len(parsed.Entries))
	unresolved := make([]*DeckImportIssue, 0)
	warnings := make([]string, 0, len(parsed.Warnings))
	cardCount := 0
	anyResolved := false

	for _, pe := range parsed.Entries {
		key := strings.ToLower(strings.TrimSpace(pe.Name))
		card := lookup[key]
		resolved := card != nil
		if resolved {
			cardCount += pe.Quantity
			anyResolved = true
		} else {
			// D-04's eager suggestion shape: Candidates is populated in
			// plan 01-04 and empty here.
			unresolved = append(unresolved, &DeckImportIssue{
				SourceLine: pe.SourceLine,
				RawLine:    pe.RawLine,
				Name:       pe.Name,
				Reason:     "Card not found.",
				Candidates: []*DeckSuggestion{},
			})
		}
		entries = append(entries, &DeckPreviewEntry{
			Quantity:   pe.Quantity,
			Name:       pe.Name,
			Section:    string(pe.Section),
			SourceLine: pe.SourceLine,
			Resolved:   resolved,
			Card:       card,
		})
	}

	for _, w := range parsed.Warnings {
		warnings = append(warnings, w.Message)
	}

	return &DeckPreview{
		SourceType:          string(parsed.Source),
		CardCount:           cardCount,
		Entries:             entries,
		CommanderCandidates: []*Card{},
		Unresolved:          unresolved,
		Warnings:            warnings,
		// An unmatched card must never block the player: CanContinue is
		// true whenever at least one entry resolved.
		CanContinue:    anyResolved,
		BlockingErrors: []string{},
	}, cardCount
}

// lookupCardsByName is modelled on server/cards.go:159-173's batch lookup,
// but uses QueryContext with the request context rather than Query — a
// deliberate deviation, noted here: cards.go discards the context it is
// handed, and plan 01-04 needs a sub-budget (Pattern 3) that Query cannot
// express.
func (s *graphQLServer) lookupCardsByName(ctx context.Context, names []string) (map[string]*Card, error) {
	lookup := map[string]*Card{}
	if len(names) == 0 {
		return lookup, nil
	}

	var combinedErr error
	rows, err := s.db.QueryContext(ctx,
		`SELECT name, id, colors, convertedmanacost, types, power, toughness, text, subtypes, supertypes, uuid, facename
		FROM cards
		WHERE name = ANY($1) OR facename = ANY($1);`,
		pq.Array(names),
	)
	if err != nil {
		if !isMissingRelation(err, "cards") {
			combinedErr = errs.Combine(combinedErr, err)
		}
		return lookup, combinedErr
	}
	defer rows.Close()

	for rows.Next() {
		var (
			nameVal    sql.NullString
			idVal      sql.NullString
			colors     sql.NullString
			cmc        sql.NullString
			types      sql.NullString
			power      sql.NullString
			toughness  sql.NullString
			text       sql.NullString
			subtypes   sql.NullString
			supertypes sql.NullString
			uuidVal    sql.NullString
			facename   sql.NullString
		)
		if err := rows.Scan(
			&nameVal, &idVal, &colors, &cmc, &types, &power, &toughness,
			&text, &subtypes, &supertypes, &uuidVal, &facename,
		); err != nil {
			combinedErr = errs.Combine(combinedErr, err)
			continue
		}
		card := &Card{
			Name:       nameVal.String,
			ID:         idVal.String,
			Colors:     nullStringPtr(colors),
			Cmc:        nullStringPtr(cmc),
			Types:      nullStringPtr(types),
			Power:      nullStringPtr(power),
			Toughness:  nullStringPtr(toughness),
			Text:       nullStringPtr(text),
			Subtypes:   nullStringPtr(subtypes),
			Supertypes: nullStringPtr(supertypes),
			UUID:       nullStringPtr(uuidVal),
		}
		if nameVal.Valid {
			key := strings.ToLower(nameVal.String)
			lookup[key] = preferCard(lookup[key], card)
		}
		if facename.Valid && facename.String != "" {
			key := strings.ToLower(facename.String)
			lookup[key] = preferCard(lookup[key], card)
		}
	}
	if err := rows.Err(); err != nil {
		combinedErr = errs.Combine(combinedErr, err)
	}

	return lookup, combinedErr
}

// finishPreviewDeck observes the import collectors and emits the
// deck_import_succeeded/deck_import_failed product event for one
// previewDeck call, then returns the preview unchanged. It is the single
// place both the metric and the product event are recorded, so the two can
// never drift out of sync with each other.
func (s *graphQLServer) finishPreviewDeck(ctx context.Context, input InputDeckImport, start time.Time, source deckimport.SourceType, cardCount int, preview *DeckPreview) (*DeckPreview, error) {
	elapsed := time.Since(start)

	outcome := "success"
	eventName := "deck_import_succeeded"
	if !preview.CanContinue {
		outcome = "failure"
		eventName = "deck_import_failed"
	}
	collectors.ObserveDeckImport(source, outcome, elapsed)

	sourceStr := string(source)
	outcomeStr := outcome
	durationMs := int(elapsed.Milliseconds())
	metadata := map[string]string{}
	if eventName == "deck_import_succeeded" {
		metadata["card_count"] = strconv.Itoa(cardCount)
		metadata["unresolved_count"] = strconv.Itoa(len(preview.Unresolved))
	} else {
		reason := "no_cards_resolved"
		if len(preview.BlockingErrors) > 0 {
			reason = "blocked_input"
		}
		metadata["reason"] = reason
	}

	s.recordProductEvent(ctx, ProductEvent{
		Name:       eventName,
		SessionID:  input.SessionID,
		Source:     &sourceStr,
		Outcome:    &outcomeStr,
		DurationMs: &durationMs,
		Metadata:   metadata,
	}, false)

	return preview, nil
}

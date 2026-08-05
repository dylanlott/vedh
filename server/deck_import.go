package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/openmtg/edh-go/pkg/deckimport"
	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/zeebo/errs"
)

// lowConfidenceCutoff is D-03's presentation-layer similarity cutoff: a
// suggestion score below this value is marked low-confidence in the
// DeckSuggestion the player sees. A score exactly equal to the cutoff
// counts as confident (the comparison below is strict "<", never "<="). It
// is deliberately a Go constant applied to a score already returned from a
// query, and appears nowhere in SQL, which is what makes it tunable without
// a migration or a contract change. 0.3 mirrors pg_trgm.similarity_threshold's
// own default, a reasonable starting point for "this is probably not what
// you meant." Plan 01-05's suggestion ranking is the first consumer; it is
// declared here, alongside the resolution code it belongs to.
const lowConfidenceCutoff = 0.3

// isLowConfidence reports whether score falls below lowConfidenceCutoff. A
// score exactly at the cutoff is confident.
func isLowConfidence(score float64) bool {
	return score < lowConfidenceCutoff
}

// maxSuggestionNeedles is Pattern 3 bound #3: beyond this many distinct
// unresolved needles in one preview, only the first maxSuggestionNeedles
// (in input order) receive suggestions, and one warning names both the
// shown count and the total. This is what turns a garbage paste or a
// failed-to-load snapshot into a constant-cost pathological case instead
// of a lookup per unresolved row.
const maxSuggestionNeedles = 25

// suggestionBudget is Pattern 3 bound #4: the batched suggestion query
// runs under a sub-context derived from the request context with this
// timeout. On expiry the preview degrades to zero suggestions plus one
// warning — D-01's "never block the player" applied to the suggestion
// path itself — rather than making previewDeck wait on a slow index scan.
const suggestionBudget = 750 * time.Millisecond

// suggestionsPerEntry is D-02's cap: at most this many ranked candidates
// per unresolved entry.
const suggestionsPerEntry = 3

// suggestionQuery is D-02/D-03/D-04's bounded eager suggestion lookup: one
// batched query over every deduplicated unresolved needle, using
// unnest(...) WITH ORDINALITY so the caller can map each result row back
// to its needle without a second query. Each needle is laterally joined
// against card_names ordered by the pg_trgm GiST distance operator alone
// (Pattern 2) — measured directly against the real ~35,831-row card_names
// table while writing this query: adding a second ORDER BY key inside the
// LATERAL (even just as a tiebreak) defeats the planner's ability to use
// card_names_trgm_gist for KNN ordering and forces a sequential scan plus
// a top-N sort per needle instead, which alone turned the 25-needle
// worst case from ~190ms into ~1.2s and blew the sub-budget below. The
// tiebreak that makes two equal-scoring candidates deterministic instead
// belongs only in the cheap outer sort, over the already-limited ≤75-row
// result set, where it costs nothing: q.ord first (map results back to
// their needle), then score descending, then the displayed name
// ascending. The distance operator returns the nearest candidates
// unconditionally, so D-03's "show the nearest matches even below the
// cutoff" is structurally guaranteed; isLowConfidence applies the cutoff
// in Go, never here. Both the needle array and the per-needle candidate
// limit are bound as parameters — no part of this string is composed
// with string formatting.
const suggestionQuery = `
	SELECT q.ord, q.needle, c.display, 1 - (c.name_lower <-> q.needle) AS score
	FROM unnest($1::text[]) WITH ORDINALITY AS q(needle, ord)
	CROSS JOIN LATERAL (
		SELECT cn.display, cn.name_lower
		FROM card_names cn
		ORDER BY cn.name_lower <-> q.needle
		LIMIT $2
	) c
	ORDER BY q.ord, score DESC, c.display ASC;`

// dedupeNeedles returns names deduplicated by their lower-cased, trimmed
// form, keeping only the first occurrence of each and preserving input
// order — Pattern 3 bound #2, the same shape server/cards.go's Cards()
// batch lookup already uses for its own needle de-duplication (lines
// 141-153), applied here to the suggestion path. A repeated misspelling
// becomes exactly one needle no matter how many entries name it.
func dedupeNeedles(names []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(names))
	for _, n := range names {
		key := strings.ToLower(strings.TrimSpace(n))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

// suggestFor implements D-02/D-03/D-04's bounded eager suggestion lookup
// over an already-deduplicated needle set (Pattern 3 bounds #1 and #2 are
// the caller's responsibility — this function applies bounds #3 and #4).
// It never returns an error: a query failure or an expired sub-budget
// degrades to zero suggestions plus one player-facing warning, and the
// caller's preview still succeeds — D-01's rule applied to the suggestion
// path itself. budget is a parameter (rather than the suggestionBudget
// constant read directly) purely so a test can force bound #4 to fire
// without waiting out a real 750ms query; production always calls this
// through buildDeckPreview, which passes the real constant.
func (s *graphQLServer) suggestFor(ctx context.Context, needles []string, budget time.Duration) (map[string][]*DeckSuggestion, []string) {
	start := time.Now()
	if len(needles) == 0 {
		return nil, nil
	}

	var warnings []string
	total := len(needles)
	if total > maxSuggestionNeedles {
		warnings = append(warnings, fmt.Sprintf(
			"Showing suggestions for the first %d of %d unresolved cards.",
			maxSuggestionNeedles, total))
		needles = needles[:maxSuggestionNeedles]
		collectors.IncDeckSuggestionTruncated()
	}

	qctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	rows, err := s.db.QueryContext(qctx, suggestionQuery, pq.Array(needles), suggestionsPerEntry)
	if err != nil {
		s.loggerFor(ctx).Warn("deck import suggestion lookup failed", "err", err, "needle_count", len(needles))
		outcome := "error"
		if qctx.Err() != nil {
			outcome = "timeout"
		}
		collectors.ObserveDeckSuggestion(outcome, time.Since(start))
		return nil, append(warnings, "Suggestions are temporarily unavailable.")
	}
	defer rows.Close()

	out := map[string][]*DeckSuggestion{}
	for rows.Next() {
		var ord int64
		var needle, display string
		var score float64
		if scanErr := rows.Scan(&ord, &needle, &display, &score); scanErr != nil {
			s.loggerFor(ctx).Warn("deck import suggestion row scan failed", "err", scanErr)
			continue
		}
		out[needle] = append(out[needle], &DeckSuggestion{
			Name:          display,
			Score:         score,
			LowConfidence: isLowConfidence(score),
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		s.loggerFor(ctx).Warn("deck import suggestion rows iteration failed", "err", rowsErr)
		collectors.ObserveDeckSuggestion("error", time.Since(start))
		return nil, append(warnings, "Suggestions are temporarily unavailable.")
	}

	collectors.ObserveDeckSuggestion("success", time.Since(start))
	return out, warnings
}

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

	// Rate limiting is the first thing this resolver does after reading
	// its input, and a limited call returns before ever reaching the
	// parser, the suggestion query, or a product-event write.
	clientKey := clientKeyFor(ctx, input.SessionID)
	if !s.allowRequest(ctx, ratelimit.SurfaceDeckImport, clientKey) {
		return rateLimitedPreview(), nil
	}

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
		return s.finishPreviewDeck(ctx, input, start, deckimport.SourceUnknown, 0, s.previewDeckURL(ctx, *input.SourceURL))
	}

	// This is the single canonical parse site for the whole codebase.
	parsed := deckimport.Parse(*input.Text)
	preview, cardCount := s.buildDeckPreview(ctx, parsed)
	return s.finishPreviewDeck(ctx, input, start, parsed.Source, cardCount, preview)
}

// deckProviderFetch is the fetch call previewDeckURL makes when
// providerEnabled() is true. A package-level variable, never a hardcoded
// call, so a test can substitute a spy and assert it was never invoked
// when the kill switch is off (or off-by-empty-allowlist) — proving zero
// dial attempts were made, not merely that an error was returned.
// Production always leaves this at its zero value, fetchDeckProviderURL.
var deckProviderFetch = fetchDeckProviderURL

// previewDeckURL handles a non-empty deck URL: the provider kill switch
// (D-16/T-01-28), defaulting to off, and treated as off whenever the host
// allowlist is empty even if the flag itself is set — a flag set with
// nothing allowlisted would open a fetch path that can reach nothing, a
// confusing half-state rather than a safe one. Neither the flag name, the
// allowlist contents, a host, nor any transport detail is ever named in
// the text this returns; PROJECT.md and 01-VALIDATION.md's provider
// runbook note require the disabled path to read as ordinary product
// language, not a configuration error.
//
// Once the flag is on, rawURL's hostname is looked up in
// deckProviderAdapters (server/deck_providers.go) — the D-14 checkpoint
// (plan 01-07) selected Moxfield, and that registry's only entry today is
// a deliberately incomplete moxfieldAdapter whose response contract has
// never been observed (docs/research/deck-provider-feasibility.md section
// 5). A host with no registered adapter is refused here, before any fetch
// is attempted, the same as a host absent from the allowlist — the
// allowlist and this registry are independent gates. A recognized host
// does genuinely reach the secure client and, on a successful fetch,
// genuinely reaches the adapter's normalizer; for Moxfield today that
// normalizer always fails closed, so the end-to-end outcome is still the
// same paste-fallback message, just reached by a real routing decision
// rather than a hardcoded skip.
func (s *graphQLServer) previewDeckURL(ctx context.Context, rawURL string) *DeckPreview {
	if !s.providerEnabled() {
		return blockedPreview(deckimport.SourceUnknown,
			"Deck links aren't available right now. Paste your decklist as text instead.")
	}

	adapter, ok := deckProviderAdapterFor(rawURL)
	if !ok {
		return blockedPreview(deckimport.SourceUnknown,
			"Deck links aren't available right now. Paste your decklist as text instead.")
	}

	body, err := deckProviderFetch(ctx, s.deckProviderClient(), s.deckProviderAllowedHosts, rawURL)
	if err != nil {
		s.loggerFor(ctx).Warn("deck provider fetch failed", "err", err)
		return blockedPreview(deckimport.SourceUnknown,
			"We couldn't load that deck link right now. Paste your decklist as text instead.")
	}

	// normalizeToDeckText's error never carries provider response
	// content for any adapter shipped so far (moxfieldAdapter's sentinel
	// is a static string), so logging it here is safe by construction,
	// not merely by convention.
	text, err := adapter.normalizeToDeckText(body)
	if err != nil {
		s.loggerFor(ctx).Warn("deck provider response could not be normalized", "err", err)
		return blockedPreview(deckimport.SourceUnknown,
			"We couldn't load that deck link right now. Paste your decklist as text instead.")
	}

	// Unreachable with today's only registered adapter (it never returns
	// a nil error), but this is the real success path a future adapter
	// with a verified contract will exercise: normalize, then hand off
	// to the single canonical parser, exactly like the pasted-text path.
	parsed := deckimport.ParseWithSource(text, adapter.source())
	preview, _ := s.buildDeckPreview(ctx, parsed)
	return preview
}

// deckProviderClient builds task 2's secure fetch client, scoped to this
// server's parsed host allowlist. Called only once providerEnabled() has
// already confirmed the allowlist is non-empty.
func (s *graphQLServer) deckProviderClient() *http.Client {
	return newSafeProviderClient(s.deckProviderAllowedHosts)
}

// rateLimitedPreview builds the DeckPreview a rate-limited previewDeck call
// returns: one product-language blocking error naming no limit value,
// window, or remaining count, and CanContinue false. It never reaches
// finishPreviewDeck, so a rate-limited call neither writes a product event
// nor observes the deck-import outcome counter — the point of limiting the
// surface is to shed load on it, not to spend a database write proving the
// load was shed.
func rateLimitedPreview() *DeckPreview {
	return blockedPreview(deckimport.SourceUnknown,
		"You're sending requests a bit too quickly. Please wait a moment and try again.")
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
// and assembles the DeckPreview the resolver returns, using the real
// suggestionBudget. It is a thin wrapper over
// buildDeckPreviewWithSuggestionBudget so a test can force the suggestion
// sub-budget to expire without changing this method's own behavior.
func (s *graphQLServer) buildDeckPreview(ctx context.Context, parsed deckimport.ParsedDeck) (*DeckPreview, int) {
	return s.buildDeckPreviewWithSuggestionBudget(ctx, parsed, suggestionBudget)
}

// buildDeckPreviewWithSuggestionBudget resolves a ParsedDeck's entries
// against the cards table and assembles the DeckPreview the resolver
// returns. It never fails: a lookup error is logged and treated as
// "nothing resolved" rather than returned to the caller, matching D-01's
// "never block the player."
func (s *graphQLServer) buildDeckPreviewWithSuggestionBudget(ctx context.Context, parsed deckimport.ParsedDeck, budget time.Duration) (*DeckPreview, int) {
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

	resolvedCards, printingWarnings, lookupErr := s.resolveDeckEntries(ctx, parsed.Entries)
	if lookupErr != nil {
		s.loggerFor(ctx).Warn("deck import card lookup failed", "err", lookupErr)
	}

	entries := make([]*DeckPreviewEntry, 0, len(parsed.Entries))
	unresolved := make([]*DeckImportIssue, 0)
	warnings := make([]string, 0, len(parsed.Warnings)+len(printingWarnings))
	cardCount := 0
	anyResolved := false
	var unresolvedNames []string

	for i, pe := range parsed.Entries {
		card := resolvedCards[i]
		resolved := card != nil
		if resolved {
			cardCount += pe.Quantity
			anyResolved = true
		} else {
			// D-02/D-04: Candidates is filled in below, once, from one
			// batched lookup over every unresolved entry's (deduplicated)
			// name — never per entry here.
			unresolvedNames = append(unresolvedNames, pe.Name)
			unresolved = append(unresolved, &DeckImportIssue{
				SourceLine: pe.SourceLine,
				RawLine:    pe.RawLine,
				Name:       pe.Name,
				Reason:     "Card not found.",
				Candidates: []*DeckSuggestion{},
			})
		}
		entries = append(entries, &DeckPreviewEntry{
			Quantity: pe.Quantity,
			Name:     pe.Name,
			// D-07: the parsed printing metadata is carried through onto
			// the entry regardless of how (or whether) it resolved, so a
			// later snapshot refresh or the future proxy-print feature can
			// reconcile it.
			SetCode:         stringPtrOrNil(pe.SetCode),
			CollectorNumber: stringPtrOrNil(pe.CollectorNumber),
			Category:        stringPtrOrNil(pe.Category),
			Section:         string(pe.Section),
			SourceLine:      pe.SourceLine,
			Resolved:        resolved,
			Card:            card,
		})
	}

	// D-02/D-03/D-04, Pattern 3: suggestions are computed eagerly, over the
	// failure set only (bound #1), deduplicated by lower-cased name
	// (bound #2), and returned inside this same response. suggestFor owns
	// bounds #3 (the 25-needle cap) and #4 (the sub-budget); it never
	// fails the preview.
	suggestions, suggestionWarnings := s.suggestFor(ctx, dedupeNeedles(unresolvedNames), budget)
	for _, issue := range unresolved {
		key := strings.ToLower(strings.TrimSpace(issue.Name))
		if cands, ok := suggestions[key]; ok {
			issue.Candidates = cands
		}
	}
	warnings = append(warnings, suggestionWarnings...)

	for _, w := range parsed.Warnings {
		warnings = append(warnings, w.Message)
	}
	warnings = append(warnings, printingWarnings...)

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

// cardResultColumns is the column list shared by every query in this file
// that returns a full printing row: it must stay in sync with
// scanCardResultRow's Scan call below.
const cardResultColumns = `name, id, colors, convertedmanacost, types, power, toughness, text, subtypes, supertypes, uuid, facename, setcode, number, scryfallid`

// scanCardResultRow scans one row shaped like cardResultColumns and returns
// the resulting *Card alongside its lower-cased name and face-name (each ""
// when the corresponding column was absent), so a caller can key a map by
// either without re-deriving the lower-cased form itself.
func scanCardResultRow(rows *sql.Rows) (card *Card, nameLower string, faceLower string, err error) {
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
		setcode    sql.NullString
		number     sql.NullString
		scryfallID sql.NullString
	)
	if err := rows.Scan(
		&nameVal, &idVal, &colors, &cmc, &types, &power, &toughness,
		&text, &subtypes, &supertypes, &uuidVal, &facename,
		&setcode, &number, &scryfallID,
	); err != nil {
		return nil, "", "", err
	}
	card = &Card{
		Name:            nameVal.String,
		ID:              idVal.String,
		Colors:          nullStringPtr(colors),
		Cmc:             nullStringPtr(cmc),
		Types:           nullStringPtr(types),
		Power:           nullStringPtr(power),
		Toughness:       nullStringPtr(toughness),
		Text:            nullStringPtr(text),
		Subtypes:        nullStringPtr(subtypes),
		Supertypes:      nullStringPtr(supertypes),
		UUID:            nullStringPtr(uuidVal),
		SetCode:         nullStringPtr(setcode),
		CollectorNumber: nullStringPtr(number),
		ScryfallID:      nullStringPtr(scryfallID),
	}
	if nameVal.Valid {
		nameLower = strings.ToLower(nameVal.String)
	}
	if facename.Valid && facename.String != "" {
		faceLower = strings.ToLower(facename.String)
	}
	return card, nameLower, faceLower, nil
}

// resolveDeckEntries implements D-09's two-stage printing resolution over a
// batch of parsed entries, keyed by index into entries so a caller never
// needs a second lookup to map a result back to the entry it came from.
//
// Stage one resolves every entry that supplies a set code or collector
// number to its exact printing, in one batched query using the
// ordinality-preserving unnest-over-parallel-arrays form. Stage two
// resolves everything stage one missed — including an entry whose supplied
// set code simply is not in this snapshot — by name alone, deterministically
// (see nameOnlyFallbackQuery's comment for the ordering key), and attaches a
// D-10 warning naming the requested set code whenever a set code was
// supplied but a different printing had to be used. Both stages tolerate a
// missing `cards` relation the same way server/cards.go's Cards() does, and
// accumulate per-row scan errors with the same multi-error combiner rather
// than aborting the whole preview.
//
// This is modelled on server/cards.go's Cards(), but deliberately uses
// QueryContext with the request context rather than Query: this resolver
// runs on previewDeck's unauthenticated, public request path, where a
// caller-controlled deadline should be able to cancel an in-flight lookup;
// cards.go's Cards() has no such deadline to propagate.
func (s *graphQLServer) resolveDeckEntries(ctx context.Context, entries []deckimport.ParsedEntry) (map[int]*Card, []string, error) {
	result := make(map[int]*Card, len(entries))
	var warnings []string
	var combinedErr error

	// Stage one: exact printing for every entry supplying a set code or a
	// collector number.
	var stageOneIdx []int
	var nameLowers, setLowers, numbers []string
	for i, e := range entries {
		if e.SetCode == "" && e.CollectorNumber == "" {
			continue
		}
		stageOneIdx = append(stageOneIdx, i)
		nameLowers = append(nameLowers, strings.ToLower(strings.TrimSpace(e.Name)))
		setLowers = append(setLowers, strings.ToLower(e.SetCode))
		numbers = append(numbers, e.CollectorNumber)
	}

	stageOneHit := make(map[int]bool, len(stageOneIdx))
	if len(stageOneIdx) > 0 {
		rows, err := s.db.QueryContext(ctx, exactPrintingQuery,
			pq.Array(nameLowers), pq.Array(setLowers), pq.Array(numbers))
		if err != nil {
			if !isMissingRelation(err, "cards") {
				combinedErr = errs.Combine(combinedErr, err)
			}
		} else {
			func() {
				defer rows.Close()
				for rows.Next() {
					var ord int64
					var nameVal, idVal, colors, cmc, types, power, toughness, text, subtypes, supertypes, uuidVal, facename, setcode, number, scryfallID sql.NullString
					if scanErr := rows.Scan(
						&ord, &nameVal, &idVal, &colors, &cmc, &types, &power, &toughness,
						&text, &subtypes, &supertypes, &uuidVal, &facename, &setcode, &number, &scryfallID,
					); scanErr != nil {
						combinedErr = errs.Combine(combinedErr, scanErr)
						continue
					}
					pos := int(ord) - 1
					if pos < 0 || pos >= len(stageOneIdx) {
						continue
					}
					entryIdx := stageOneIdx[pos]
					result[entryIdx] = &Card{
						Name:            nameVal.String,
						ID:              idVal.String,
						Colors:          nullStringPtr(colors),
						Cmc:             nullStringPtr(cmc),
						Types:           nullStringPtr(types),
						Power:           nullStringPtr(power),
						Toughness:       nullStringPtr(toughness),
						Text:            nullStringPtr(text),
						Subtypes:        nullStringPtr(subtypes),
						Supertypes:      nullStringPtr(supertypes),
						UUID:            nullStringPtr(uuidVal),
						SetCode:         nullStringPtr(setcode),
						CollectorNumber: nullStringPtr(number),
						ScryfallID:      nullStringPtr(scryfallID),
					}
					stageOneHit[entryIdx] = true
				}
				if rowsErr := rows.Err(); rowsErr != nil {
					combinedErr = errs.Combine(combinedErr, rowsErr)
				}
			}()
		}
	}

	// Stage two: name-only fallback, deduplicated by needle, for every
	// entry stage one did not resolve.
	var stageTwoNames []string
	seenNeedle := map[string]struct{}{}
	for i, e := range entries {
		if stageOneHit[i] {
			continue
		}
		needle := strings.ToLower(strings.TrimSpace(e.Name))
		if needle == "" {
			continue
		}
		if _, ok := seenNeedle[needle]; ok {
			continue
		}
		seenNeedle[needle] = struct{}{}
		stageTwoNames = append(stageTwoNames, needle)
	}

	byName := map[string]*Card{}
	if len(stageTwoNames) > 0 {
		rows, err := s.db.QueryContext(ctx, nameOnlyFallbackQuery, pq.Array(stageTwoNames))
		if err != nil {
			if !isMissingRelation(err, "cards") {
				combinedErr = errs.Combine(combinedErr, err)
			}
		} else {
			func() {
				defer rows.Close()
				for rows.Next() {
					card, nameLower, faceLower, scanErr := scanCardResultRow(rows)
					if scanErr != nil {
						combinedErr = errs.Combine(combinedErr, scanErr)
						continue
					}
					// The query is already ordered by the deterministic key
					// documented on nameOnlyFallbackQuery, so keeping only
					// the first row seen per key is what makes the choice
					// deterministic rather than dependent on map iteration.
					if nameLower != "" {
						if _, ok := byName[nameLower]; !ok {
							byName[nameLower] = card
						}
					}
					if faceLower != "" {
						if _, ok := byName[faceLower]; !ok {
							byName[faceLower] = card
						}
					}
				}
				if rowsErr := rows.Err(); rowsErr != nil {
					combinedErr = errs.Combine(combinedErr, rowsErr)
				}
			}()
		}
	}

	for i, e := range entries {
		if stageOneHit[i] {
			continue
		}
		needle := strings.ToLower(strings.TrimSpace(e.Name))
		card, ok := byName[needle]
		if !ok {
			continue
		}
		result[i] = card
		if e.SetCode != "" {
			// D-10: the requested printing is not in this snapshot. A
			// different printing of the same name was found instead, so
			// this entry resolves and is never reported as unresolved —
			// only warned about, in player-facing language naming the
			// requested set code.
			warnings = append(warnings, fmt.Sprintf(
				"%q: the requested printing (%s) was not found; a different printing was used instead.",
				e.Name, e.SetCode,
			))
		}
	}

	return result, warnings, combinedErr
}

// exactPrintingQuery is D-09's stage-one lookup: one batched query over
// every entry that supplied a set code or a collector number, using
// unnest(...) WITH ORDINALITY so each result row's ordinal position maps
// straight back to its position in the three parallel input arrays — no
// second lookup needed. Set code comparison lower-cases both sides so the
// upper-case and lower-case spellings of one set select the same row; an
// entry with a set code but no collector number matches any printing in
// that set (req.number = ” short-circuits the number comparison), broken
// by the same number/id ordering documented on nameOnlyFallbackQuery.
const exactPrintingQuery = `
	SELECT req.ord, c.name, c.id, c.colors, c.convertedmanacost, c.types, c.power, c.toughness,
	       c.text, c.subtypes, c.supertypes, c.uuid, c.facename, c.setcode, c.number, c.scryfallid
	FROM unnest($1::text[], $2::text[], $3::text[]) WITH ORDINALITY AS req(name_lower, set_lower, number, ord)
	JOIN LATERAL (
		SELECT *
		FROM cards c
		WHERE lower(c.name) = req.name_lower
		  AND lower(c.setcode) = req.set_lower
		  AND (req.number = '' OR c.number = req.number)
		ORDER BY c.number NULLS LAST, c.id
		LIMIT 1
	) c ON true;`

// nameOnlyFallbackQuery is D-09's stage-two lookup: name-only resolution
// for whatever stage one did not match. Today which printing a name-only
// paste resolves to is decided by row arrival order, which is arbitrary;
// this query documents and fixes a deterministic ordering key instead —
// lower-cased set code, then collector number, then row identifier (the
// `id` column, compared as text) — so the printing a player gets is a
// decision, not an accident. The caller keeps only the first row seen per
// name/face-name key, which combined with this ORDER BY is what makes the
// selection deterministic.
const nameOnlyFallbackQuery = `
	SELECT ` + cardResultColumns + `
	FROM cards
	WHERE lower(name) = ANY($1) OR lower(facename) = ANY($1)
	ORDER BY lower(setcode) ASC NULLS LAST, number ASC NULLS LAST, id ASC;`

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

// stringPtrOrNil returns nil for an empty string and a pointer to s
// otherwise. deckimport.ParsedEntry's SetCode/CollectorNumber/Category
// fields are plain strings that default to "" when absent (never a
// sql.NullString), so this is the analog of nullStringPtr for that shape.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// applyPrintingMetadata returns a shallow copy of card with SetCode,
// CollectorNumber, and Category overwritten from entry's D-07 fields
// (what the player actually typed or pasted, not necessarily the printing
// resolveDeckEntries actually found — D-10 lets those two differ) whenever
// entry names one, and SourceFormat always set from source. This is the
// same "carry the parsed value through regardless of resolution outcome"
// choice buildDeckPreview already makes for DeckPreviewEntry, applied here
// to the *Card values that end up persisted in a game's payload — D-07's
// "the normalized entry and the stored Card carry setCode, collectorNumber,
// category, and sourceFormat," and D-08's point that upsertGame's JSONB
// payload makes this free: no migration, no schema change. Returns nil
// unchanged if card is nil, so a caller never needs a separate nil check.
func applyPrintingMetadata(card *Card, entry deckimport.ParsedEntry, source deckimport.SourceType) *Card {
	if card == nil {
		return nil
	}
	c := *card
	if entry.SetCode != "" {
		c.SetCode = stringPtrOrNil(entry.SetCode)
	}
	if entry.CollectorNumber != "" {
		c.CollectorNumber = stringPtrOrNil(entry.CollectorNumber)
	}
	if entry.Category != "" {
		c.Category = stringPtrOrNil(entry.Category)
	}
	sourceStr := string(source)
	c.SourceFormat = &sourceStr
	return &c
}

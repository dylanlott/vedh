# vEDH Pokémon Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add Pokémon support to vEDH by evolving it into a multi-game board tracker with a thin Pokémon MVP, without corrupting the current Commander-first production path.

**Architecture:** Do not treat Pokémon as a new `format` inside the existing Commander model. Instead, introduce a real game-type layer that separates shared game/session concerns from game-specific board zones, deck validation, card ingestion, search semantics, and board rendering. Ship the smallest useful Pokémon slice first: a manual board tracker with Pokémon zones and card lookup, but no full rules engine.

**Tech Stack:** Go, GraphQL, SQLite/Postgres persistence, Vue 3, TypeScript, Vitest, Playwright, MTGJSON/Scryfall existing pipeline, new Pokémon card-data adapter TBD.

---

## Success Criteria

- A new game type can define its own zones, deck rules, turn flow, and card provider without pretending to be Commander.
- Existing EDH flows remain green and unchanged for current users.
- Pokémon MVP supports: create game, join game, render Pokémon-specific zones, move cards between zones, and persist/reload state.
- Pokémon support is explicitly **manual tracking**, not a rules engine.
- The card/search pipeline has a documented Pokémon source-of-truth and a tested adapter.

## Explicit Non-Goals

- No full Pokémon rules engine.
- No automated prize handling, evolution legality, attack resolution, or status-effect engine in MVP.
- No repo split unless demand later proves a separate product is the right move.

## Blast Radius

**High-risk backend areas**
- `vedh/server/schema.graphql`
- `vedh/server/models_gen.go`
- `vedh/server/formats.go`
- `vedh/server/games.go`
- `vedh/server/boardstate_log.go`
- `vedh/server/cards.go`

**High-risk frontend areas**
- `vedh/app/src/formats/registry.ts`
- `vedh/app/src/components/games/FormCreateGame.vue`
- `vedh/app/src/views/JoinGameView.vue`
- `vedh/app/src/views/BoardView.vue`
- `vedh/app/src/views/ScoreView.vue`
- `vedh/app/src/stores/games.ts`
- `vedh/app/src/graphql/queries.ts`
- `vedh/app/src/graphql/mutations.ts`
- `vedh/app/src/services/scryfall.ts`
- `vedh/app/src/services/commanderPartner.ts`

**High-risk ingestion/search areas**
- `vedh/persistence/import_all_printings_json.go`
- `vedh/persistence/import_all_printings.go`
- `vedh/persistence/migrations/20211011150143_allprintings.up.sql`

---

## Phase Order

1. Card-data source decision and adapter boundary
2. Real game-type + generic zone model
3. Pokémon manual board-tracker MVP
4. Pokémon deck/search ingestion + validation
5. QA, docs, rollout guardrails

---

### Task 1: Lock the Pokémon card-data source and adapter contract

**Files:**
- Create: `vedh/docs/plans/2026-08-08-pokemon-card-data-source-decision.md`
- Create: `vedh/server/card_provider.go`
- Create: `vedh/server/card_provider_test.go`
- Reference: `vedh/server/cards.go`
- Reference: `vedh/persistence/import_all_printings_json.go`
- Reference: `vedh/app/src/services/scryfall.ts`

**Step 1: Write the failing test**

```go
func TestCardProviderRegistryRejectsUnknownGameType(t *testing.T) {
    registry := NewCardProviderRegistry()

    _, err := registry.ProviderForGameType("pokemon")

    if err == nil {
        t.Fatal("expected missing provider error")
    }
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestCardProviderRegistryRejectsUnknownGameType
```

Expected: FAIL because the provider registry does not exist yet.

**Step 3: Write minimal implementation**
- Introduce a `CardProvider` interface that separates:
  - card lookup
  - search semantics
  - image/art resolution
  - import/source metadata
- Add a registry keyed by `gameType` instead of hard-wiring MTG assumptions into shared handlers.
- Keep MTG/EDH as the first concrete provider.
- Do **not** wire Pokémon ingestion yet; only define the seam and document the decision criteria.

**Step 4: Re-run the focused test**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestCardProviderRegistryRejectsUnknownGameType
```

Expected: PASS.

**Step 5: Write the decision doc**
- Compare candidate Pokémon data sources.
- Capture licensing/availability/image-hosting comfort.
- Define the normalized fields vEDH needs for MVP:
  - canonical id
  - name
  - supertype/category
  - subtype/species
  - hp
  - stage
  - types
  - set / collector / image url
- Record the chosen source and fallback policy.

**Step 6: Commit**

```bash
git add docs/plans/2026-08-08-pokemon-card-data-source-decision.md server/card_provider.go server/card_provider_test.go
 git commit -m "refactor: add game-specific card provider boundary"
```

---

### Task 2: Introduce a real game-type and generic zone model

**Files:**
- Modify: `vedh/server/formats.go`
- Modify: `vedh/server/schema.graphql`
- Modify: `vedh/server/games.go`
- Modify: `vedh/server/boardstate_log.go`
- Modify: `vedh/app/src/formats/registry.ts`
- Modify: `vedh/app/src/graphql/queries.ts`
- Modify: `vedh/app/src/graphql/mutations.ts`
- Modify: `vedh/app/src/stores/games.ts`
- Test: `vedh/server/games_test.go`
- Test: `vedh/server/main_test.go`
- Test: `vedh/app/src/formats/registry.test.ts` (create if missing)

**Step 1: Write the failing backend test**

```go
func TestLookupGameTypeReturnsPokemonDefinition(t *testing.T) {
    gt, ok := LookupGameType("POKEMON_TCG")
    if !ok {
        t.Fatal("expected pokemon game type")
    }
    if gt.DefaultDeckSize != 60 {
        t.Fatalf("expected 60-card default deck, got %d", gt.DefaultDeckSize)
    }
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestLookupGameTypeReturnsPokemonDefinition
```

Expected: FAIL.

**Step 3: Write the failing frontend test**

```ts
import { describe, expect, it } from 'vitest';
import { lookupFormat } from './registry';

describe('game format registry', () => {
  it('returns the pokemon game type definition', () => {
    const format = lookupFormat('POKEMON_TCG');
    expect(format.DefaultDeckSize).toBe(60);
    expect(format.Zones.map((zone) => zone.ID)).toContain('active');
  });
});
```

**Step 4: Run test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test -- src/formats/registry.test.ts
```

Expected: FAIL.

**Step 5: Write minimal implementation**
- Rename the conceptual layer from “format means Commander variant” to “game type / format definition”.
- Keep `EDH` working.
- Add `POKEMON_TCG` definition with Pokémon zones:
  - `deck`
  - `hand`
  - `active`
  - `bench`
  - `discard`
  - `lost_zone`
  - `prizes`
  - `stadium`
- Add a zone list to API responses instead of relying on fixed BoardState field names alone.
- Preserve old EDH behavior while introducing the new generic representation incrementally.
- Do not remove legacy Commander fields until replacement paths are green.

**Step 6: Re-run focused tests**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestLookupGameTypeReturnsPokemonDefinition
npm --prefix app run test -- src/formats/registry.test.ts
```

Expected: PASS.

**Step 7: Commit**

```bash
git add server/formats.go server/schema.graphql server/games.go server/boardstate_log.go app/src/formats/registry.ts app/src/graphql/queries.ts app/src/graphql/mutations.ts app/src/stores/games.ts app/src/formats/registry.test.ts
 git commit -m "refactor: add game-type and generic zone definitions"
```

---

### Task 3: Ship a Pokémon create/join/manual-board MVP

**Files:**
- Modify: `vedh/app/src/components/games/FormCreateGame.vue`
- Modify: `vedh/app/src/views/JoinGameView.vue`
- Modify: `vedh/app/src/views/BoardView.vue`
- Modify: `vedh/app/src/views/ScoreView.vue`
- Modify: `vedh/server/games.go`
- Test: `vedh/app/src/components/games/FormCreateGame.test.ts` (create if missing)
- Test: `vedh/app/e2e/pokemon-board.spec.ts` (create)

**Step 1: Write the failing form test**

```ts
it('shows pokemon as a selectable game type with no commander input', async () => {
  // render create game modal
  // switch format to POKEMON_TCG
  // assert commander picker is hidden
  // assert deck size defaults to 60
});
```

**Step 2: Run the focused test**

Run:
```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test -- src/components/games/FormCreateGame.test.ts
```

Expected: FAIL.

**Step 3: Write the failing browser test**

```ts
test('pokemon game renders pokemon zones for both players', async ({ page }) => {
  // create pokemon game
  // join second player
  // open board
  // assert Active, Bench, Prizes, Discard, Lost Zone are visible
});
```

**Step 4: Run the browser test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "pokemon game renders"
```

Expected: FAIL.

**Step 5: Write minimal implementation**
- Allow `FormCreateGame.vue` to choose Pokémon.
- Hide Commander-only controls for Pokémon.
- Default Pokémon deck size to 60.
- Update board rendering to consume zone definitions instead of hard-coded Commander columns.
- Keep MVP interactions simple:
  - drag/move cards between zones
  - show counts for deck/prizes
  - no rules automation

**Step 6: Re-run tests**

Run:
```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test -- src/components/games/FormCreateGame.test.ts
npm --prefix app run test:e2e -- --grep "pokemon game renders"
```

Expected: PASS.

**Step 7: Commit**

```bash
git add app/src/components/games/FormCreateGame.vue app/src/views/JoinGameView.vue app/src/views/BoardView.vue app/src/views/ScoreView.vue server/games.go app/src/components/games/FormCreateGame.test.ts app/e2e/pokemon-board.spec.ts
 git commit -m "feat: add pokemon manual board-tracker MVP"
```

---

### Task 4: Add Pokémon ingestion, lookup, and deck validation

**Files:**
- Create: `vedh/persistence/import_pokemon_cards.go`
- Create: `vedh/persistence/import_pokemon_cards_test.go`
- Modify: `vedh/server/cards.go`
- Modify: `vedh/server/games.go`
- Modify: `vedh/app/src/services/scryfall.ts` or replace with a generic card-art service boundary
- Test: `vedh/server/games_test.go`
- Test: `vedh/server/searchall_integration_test.go`

**Step 1: Write the failing deck-validation test**

```go
func TestCreateLibraryFromDecklistPokemonRejectsNonSixtyCardDeck(t *testing.T) {
    _, err := createLibraryFromDecklist(samplePokemonDeck(59), LookupRequiredGameType("POKEMON_TCG"), nil)
    if err == nil {
        t.Fatal("expected invalid deck size error")
    }
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestCreateLibraryFromDecklistPokemonRejectsNonSixtyCardDeck
```

Expected: FAIL.

**Step 3: Write the failing card-search test**

```go
func TestPokemonProviderSearchesByName(t *testing.T) {
    // seed one pokemon card row
    // query provider search
    // assert exact match returns pokemon card metadata
}
```

**Step 4: Run test to verify it fails**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestPokemonProviderSearchesByName
```

Expected: FAIL.

**Step 5: Write minimal implementation**
- Add Pokémon import pipeline for the chosen source.
- Normalize Pokémon rows into provider-specific storage.
- Split deck validation by game type.
- Keep EDH validator logic intact.
- Replace `scryfall.ts` assumptions with a generic card-image resolver.

**Step 6: Re-run focused tests**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run 'TestCreateLibraryFromDecklistPokemonRejectsNonSixtyCardDeck|TestPokemonProviderSearchesByName'
```

Expected: PASS.

**Step 7: Commit**

```bash
git add persistence/import_pokemon_cards.go persistence/import_pokemon_cards_test.go server/cards.go server/games.go app/src/services/scryfall.ts server/games_test.go server/searchall_integration_test.go
 git commit -m "feat: add pokemon card ingestion and validation"
```

---

### Task 5: Add regression coverage, docs, and rollout guardrails

**Files:**
- Create: `vedh/app/e2e/pokemon-create-join-smoke.spec.ts`
- Modify: `vedh/server/games_test.go`
- Modify: `vedh/server/boardstates_test.go`
- Modify: `vedh/README.md`
- Create: `vedh/docs/plans/2026-08-08-pokemon-mvp-rollout-checklist.md`

**Step 1: Write the failing regression tests**
- Add one EDH regression test proving Commander flows still work.
- Add one Pokémon smoke test proving create → join → zone render still works.

Example Go regression test:

```go
func TestEDHCreateGameStillDefaultsCommanderFields(t *testing.T) {
    // create EDH game
    // assert commander-enabled defaults remain intact
}
```

Example Playwright smoke:

```ts
test('pokemon create join smoke stays green', async ({ page, browser }) => {
  // create pokemon game
  // join from second context
  // move one card between zones
  // reload and assert persisted board state
});
```

**Step 2: Run tests to verify they fail**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server -run TestEDHCreateGameStillDefaultsCommanderFields
npm --prefix app run test:e2e -- --grep "pokemon create join smoke"
```

Expected: FAIL first if the compatibility path is incomplete.

**Step 3: Write minimal implementation**
- Fix only the compatibility gaps exposed by tests.
- Update README and rollout checklist.
- Document MVP limitations clearly so users do not expect rules automation.

**Step 4: Re-run verification**

Run:
```bash
cd /root/.openclaw/workspace/vedh
go test ./server
npm --prefix app run test
npm --prefix app run test:e2e
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/e2e server/games_test.go server/boardstates_test.go README.md docs/plans/2026-08-08-pokemon-mvp-rollout-checklist.md
 git commit -m "test: add pokemon rollout regression coverage"
```

---

## Recommended Execution Order

1. Task 1 — card-data source + provider seam
2. Task 2 — game-type + generic zone model
3. Task 3 — manual Pokémon board MVP
4. Task 4 — ingestion + validation
5. Task 5 — regression + rollout docs

## Sharp Risks

- The current GraphQL `BoardState` shape is still zone-field-specific, so trying to jump straight to Pokémon UI before generic zone support will create ugly dual-path debt.
- Card-data source and image-policy uncertainty is the biggest non-code blocker.
- `FormCreateGame.vue` and `BoardView.vue` are currently Commander-biased enough that UI work will regress EDH if not covered by regression tests.
- Search/card rendering is coupled to MTG fields like `ColorIdentity`, `ManaCost`, and `ScryfallID`; that needs a provider boundary before Pokémon is real.

## Recommended Card Breakdown

- Card 1: Decide Pokémon card source + add provider boundary
- Card 2: Generalize vEDH into a real game-type + zone-model core
- Card 3: Ship Pokémon manual board-tracker MVP
- Card 4: Add Pokémon ingestion/search/validation
- Card 5: Add regression coverage + rollout checklist

Plan complete and saved to `vedh/docs/plans/2026-08-08-vedh-pokemon-support-implementation.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?

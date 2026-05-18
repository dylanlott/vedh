# Game Format Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a game-format system so vEDH can support multiple TCG archetypes with different board layouts, zone sets, turn/phase labels, and rules defaults, while keeping Magic/Commander as one supported format rather than the hard-coded base model.

**Architecture:** Introduce a first-class format definition layer shared by the Go backend and Vue frontend. Store a format identifier plus optional per-game overrides in `Game.Rules`, derive board zones/turn phases/layout from that format definition, and migrate the board UI from hard-coded Magic zones/buttons to a format-driven renderer. Keep the first slice focused on archetype support (Magic Commander first, one non-Magic sample archetype second) rather than a full visual format editor.

**Tech Stack:** Go, gqlgen GraphQL, Postgres JSON payload storage in `games`, Vue 3, Pinia, Apollo GraphQL, Vitest, existing Dokku deploy flow.

---

## Current grounded context

The current codebase is strongly Magic-specific in both data shape and UI:

- Backend schema hard-codes zones on `BoardState` in `server/schema.graphql`
  - `Commander`, `Library`, `Graveyard`, `Exiled`, `Battlefield`, `Hand`, `Revealed`, `Controlled`
- Turn state is a plain string in `Turn.Phase`, but the allowed sequence is effectively enforced in frontend board logic
- `Game.Rules` already exists and currently stores metadata such as:
  - `format=EDH`
  - `deck_size=99`
- Game creation is Commander-centric in `app/src/components/games/FormCreateGame.vue`
- Join flow is Commander-centric in `app/src/views/JoinGameView.vue`
- Board rendering is fully hard-coded around Magic zones in `app/src/views/BoardView.vue`
- Pinia store typing is Magic-shaped in `app/src/stores/games.ts`
- Existing backend tests live in:
  - `server/games_test.go`
  - `server/boardstates_test.go`
- Existing frontend tests live in:
  - `app/__tests__/FormCreateGame.integration.spec.ts`
  - `app/__tests__/JoinGame.integration.spec.ts`

## Product framing

This feature should not start as “arbitrary custom schema designer.” That’s too big and too mushy.

Instead, ship it in this order:

1. **Format registry**
   - named supported formats/archetypes
   - each one defines zones, phase sequence, defaults, labels, and board layout hints
2. **Format-aware game creation**
   - choose format at create/join time
   - initialize board state and turn state from the chosen format
3. **Format-driven board UI**
   - render zones and turn controls from the registry instead of hard-coded Magic assumptions
4. **Optional per-game overrides later**
   - rename zones, tweak phases, hide/show sections

That keeps scope coherent and gets you from “EDH app” to “TCG table engine” without building a no-code game designer first.

## Recommended first supported formats

Ship with exactly two presets in v1:

1. **Magic: Commander**
   - proves backward compatibility
   - preserves current user expectations
2. **Generic Duel TCG**
   - sample non-Magic archetype
   - zones like `deck`, `hand`, `field`, `discard`, `banished`
   - phases like `draw`, `main`, `battle`, `end`

Do **not** add Yu-Gi-Oh/Pokémon/Lorcana-specific card rules in v1. Just support a generic layout archetype that proves the engine is no longer Magic-only.

---

### Task 1: Define the format domain model

**Files:**
- Create: `server/formats.go`
- Create: `app/src/formats/registry.ts`
- Modify: `server/models_custom.go`
- Modify: `server/models_gen.go` (generated output awareness only; do not hand-edit generated sections unless regeneration requires it)
- Test: `server/games_test.go`

**Step 1: Write the failing backend test**

Add a test in `server/games_test.go` asserting that a created game with `Rules: [{Name:"format", Value:"GENERIC_DUEL"}]` can round-trip with a known format ID and default turn phase.

Example assertion shape:

```go
func TestCreateGame_GenericDuelFormatRoundTrips(t *testing.T) {
    s := testAPI(t)
    input := InputCreateGame{
        ID: "format-generic-duel",
        Turn: &InputTurn{Player: "shakezula", Phase: "draw", Number: 1, Priority: "shakezula"},
        Players: []*InputBoardState{{User: "shakezula", Life: 20}},
    }

    game, err := s.CreateGame(authCtx("shakezula"), input)
    require.NoError(t, err)
    require.Equal(t, "GENERIC_DUEL", findRule(game.Rules, "format"))
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestCreateGame_GenericDuelFormatRoundTrips -v
```

Expected: FAIL because there is no first-class format helper/registry yet.

**Step 3: Write minimal backend format model**

Create `server/formats.go` with:

- `type FormatDefinition struct`
  - `ID string`
  - `Name string`
  - `StartingLife int`
  - `DefaultDeckSize int`
  - `Zones []ZoneDefinition`
  - `PhaseSequence []string`
  - `Layout LayoutDefinition`
  - `CommanderEnabled bool`
- `type ZoneDefinition struct`
  - `ID string`
  - `Label string`
  - `Visibility string` (`private`, `public`, `count_only`)
  - `Kind string` (`stacked`, `hand`, `library`, `grid`, etc.)
  - `SupportsCards bool`
- `var formatRegistry = map[string]FormatDefinition{ ... }`
  - include `EDH`
  - include `GENERIC_DUEL`
- helpers:
  - `func LookupFormat(id string) (FormatDefinition, bool)`
  - `func DefaultFormat() FormatDefinition`
  - `func findRuleValue(rules []*Rule, name string) string`

Create matching frontend registry file `app/src/formats/registry.ts` with the same two format IDs and the same zone/phase metadata.

**Step 4: Run the targeted test**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestCreateGame_GenericDuelFormatRoundTrips -v
```

Expected: PASS or fail later in create flow if defaults still aren’t wired.

**Step 5: Commit**

```bash
git add server/formats.go app/src/formats/registry.ts server/games_test.go
git commit -m "feat: add shared game format registry"
```

---

### Task 2: Make game creation persist format and defaults explicitly

**Files:**
- Modify: `server/games.go`
- Modify: `server/schema.graphql`
- Modify: `server/models_gen.go` (via gqlgen regeneration if schema changes)
- Modify: `server/generated.go` (via gqlgen regeneration if schema changes)
- Test: `server/games_test.go`

**Step 1: Write failing tests for create defaults**

Add tests asserting:

- `EDH` creation gets `format=EDH`, `deck_size=99`, `starting_life=40`
- `GENERIC_DUEL` creation gets `format=GENERIC_DUEL`, `deck_size=60`, `starting_life=20`
- omitted format falls back to `EDH`

**Step 2: Run the backend tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestCreateGame|TestGameFormat' -v
```

Expected: FAIL because defaults are still Commander-shaped.

**Step 3: Implement minimal schema/input support**

Update GraphQL input so `InputCreateGame` can carry a format identifier explicitly.

Recommended schema addition in `server/schema.graphql`:

```graphql
input InputCreateGame {
  ID: String!
  Turn: InputTurn!
  Handle: String
  FormatID: String
  Players: [InputBoardState!]!
}
```

Then in `server/games.go`:

- resolve `formatID := input.FormatID` or existing `Rules.format` or default `EDH`
- load format definition from registry
- normalize `Turn.Phase` to the first phase if absent/invalid
- upsert rules:
  - `format`
  - `deck_size`
  - `starting_life`
- initialize missing player life totals from format default when not explicitly provided

**Step 4: Regenerate gqlgen code if needed**

Run the repo’s gqlgen generation command if one exists; otherwise update generated files using the project’s existing GraphQL generation flow.

Document the exact command you use in the commit message or follow-up notes.

**Step 5: Run tests**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestCreateGame|TestGameFormat' -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add server/schema.graphql server/games.go server/games_test.go server/generated.go server/models_gen.go
git commit -m "feat: persist game format defaults on create"
```

---

### Task 3: Add a format descriptor query for the frontend

**Files:**
- Modify: `server/schema.graphql`
- Modify: `server/games.go` or create `server/formats_api.go`
- Modify: `app/src/graphql/queries.ts`
- Modify: `app/src/stores/games.ts`
- Test: `server/graphql_httpserver_test.go`

**Step 1: Write failing GraphQL API test**

Add a backend/API test asserting a query like `formats` returns at least `EDH` and `GENERIC_DUEL` with zones and phase sequence.

**Step 2: Run test to verify it fails**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestFormatsQuery -v
```

Expected: FAIL because no query exists.

**Step 3: Implement the query**

Add GraphQL types like:

```graphql
type GameFormat {
  ID: String!
  Name: String!
  StartingLife: Int!
  DefaultDeckSize: Int!
  CommanderEnabled: Boolean!
  PhaseSequence: [String!]!
  Zones: [GameFormatZone!]!
}

type GameFormatZone {
  ID: String!
  Label: String!
  Visibility: String!
  Kind: String!
  SupportsCards: Boolean!
}
```

Add query:

```graphql
formats: [GameFormat!]!
```

Return data from `formatRegistry`.

**Step 4: Add frontend query document**

In `app/src/graphql/queries.ts`, add `FORMATS_QUERY`.

**Step 5: Run test**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestFormatsQuery -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add server/schema.graphql server/formats_api.go server/graphql_httpserver_test.go app/src/graphql/queries.ts
git commit -m "feat: expose game format definitions over graphql"
```

---

### Task 4: Make create-game UI format-aware

**Files:**
- Modify: `app/src/components/games/FormCreateGame.vue`
- Modify: `app/src/stores/games.ts`
- Modify: `app/__tests__/FormCreateGame.integration.spec.ts`
- Create: `app/__tests__/formatRegistry.spec.ts`

**Step 1: Write failing frontend tests**

Add tests asserting:

- the create-game form shows more than one format option
- selecting `Generic Duel` updates:
  - deck size default to `60`
  - commander selector hidden/disabled
  - life default prepared as `20`
- selecting `Commander` restores commander picker and deck size `99`

**Step 2: Run tests to verify they fail**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run FormCreateGame.integration.spec.ts formatRegistry.spec.ts
```

Expected: FAIL because format registry is not wired into the form.

**Step 3: Implement minimal form behavior**

In `FormCreateGame.vue`:

- replace the hard-coded single format option with options from registry/query
- when format changes:
  - update `form.deckSize`
  - clear commander selection when `CommanderEnabled` is false
  - change labels/placeholders to format-specific copy
- include `FormatID` in the GraphQL create payload
- gate commander-only controls behind `selectedFormat.CommanderEnabled`

**Step 4: Run frontend tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run FormCreateGame.integration.spec.ts formatRegistry.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/components/games/FormCreateGame.vue app/src/stores/games.ts app/__tests__/FormCreateGame.integration.spec.ts app/__tests__/formatRegistry.spec.ts
git commit -m "feat: make game creation format aware"
```

---

### Task 5: Make join-game UI format-aware

**Files:**
- Modify: `app/src/views/JoinGameView.vue`
- Modify: `app/src/graphql/queries.ts`
- Modify: `app/__tests__/JoinGame.integration.spec.ts`

**Step 1: Write failing test**

Add a test asserting that when the loaded game format is non-Commander:

- commander chooser is hidden
- format label is shown
- join payload does not assume commander selection

**Step 2: Run test to verify it fails**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run JoinGame.integration.spec.ts
```

Expected: FAIL because the join view currently assumes Commander flow.

**Step 3: Implement minimal join behavior**

In `JoinGameView.vue`:

- derive current format from the loaded game’s rules
- hide or replace commander UI when the format doesn’t support commanders
- show format name and a short zone/turn hint
- keep decklist import behavior generic

**Step 4: Run test**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run JoinGame.integration.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/views/JoinGameView.vue app/__tests__/JoinGame.integration.spec.ts
git commit -m "feat: adapt join flow to game format"
```

---

### Task 6: Introduce a format-driven board view model

**Files:**
- Create: `app/src/formats/boardLayout.ts`
- Modify: `app/src/stores/games.ts`
- Modify: `app/src/views/BoardView.vue`
- Test: `app/__tests__/boardLayout.spec.ts`

**Step 1: Write failing unit tests**

Add tests for a helper that maps:

- game rules → format definition
- format definition + current player role → zone render sections
- phase sequence → next phase label

Test examples:

- EDH includes commander + battlefield + graveyard + exile
- Generic Duel includes deck + hand + field + discard + banished
- next phase computation wraps correctly

**Step 2: Run test to verify it fails**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run boardLayout.spec.ts
```

Expected: FAIL because helper does not exist.

**Step 3: Implement the view-model helper**

Create `app/src/formats/boardLayout.ts` with helpers such as:

- `resolveFormatFromRules(rules)`
- `getVisibleZones(format, role)`
- `getNextPhase(format, currentPhase)`
- `getZoneLabel(format, zoneId)`
- `zoneUsesHiddenCards(format, zoneId)`

Do not mutate network payloads yet. This task is only about deriving UI behavior.

**Step 4: Run tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run boardLayout.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/formats/boardLayout.ts app/__tests__/boardLayout.spec.ts
git commit -m "feat: add format-driven board layout helpers"
```

---

### Task 7: Replace hard-coded phase logic in the board UI

**Files:**
- Modify: `app/src/views/BoardView.vue`
- Modify: `app/src/graphql/mutations.ts`
- Test: `app/__tests__/boardLayout.spec.ts`

**Step 1: Write failing tests for next-phase behavior**

Add or extend tests proving:

- EDH phase order is preserved
- Generic Duel phase order is different and rendered correctly
- `Advance to …` button label comes from format definition, not static logic

**Step 2: Run tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run boardLayout.spec.ts
```

Expected: FAIL while `BoardView.vue` still hard-codes phase logic.

**Step 3: Implement minimal board changes**

In `BoardView.vue`:

- replace static phase-order computation with `getNextPhase(format, currentPhase)`
- show current format name somewhere near turn controls
- stop assuming phase names like `pregame`/`main` in display logic

**Step 4: Run tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run boardLayout.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/views/BoardView.vue app/__tests__/boardLayout.spec.ts
git commit -m "feat: drive board phase controls from format"
```

---

### Task 8: Replace hard-coded zone sections with a format-driven renderer

**Files:**
- Create: `app/src/components/board/BoardZoneSection.vue`
- Modify: `app/src/views/BoardView.vue`
- Modify: `app/src/stores/games.ts`
- Test: `app/__tests__/BoardZoneSection.spec.ts`

**Step 1: Write failing component tests**

Add tests asserting a reusable zone component can render:

- public card tiles
- hidden count-only zones
- zone labels from format metadata
- both “main player” and “opponent” variants

**Step 2: Run tests to verify they fail**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run BoardZoneSection.spec.ts
```

Expected: FAIL because component does not exist.

**Step 3: Implement minimal reusable zone renderer**

Create `BoardZoneSection.vue` and refactor `BoardView.vue` to loop through format-provided zone lists instead of manually rendering:

- Commander
n- Battlefield
- Hand
- Graveyard
- Exiled
- Revealed
- Controlled
- Library

Important: preserve current drag/drop and stack/tile mode behavior for EDH where possible, but isolate them behind zone IDs and capabilities.

**Step 4: Run tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run BoardZoneSection.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/components/board/BoardZoneSection.vue app/src/views/BoardView.vue app/__tests__/BoardZoneSection.spec.ts
git commit -m "feat: render board zones from format definitions"
```

---

### Task 9: Add backend normalization for generic zones without breaking EDH

**Files:**
- Modify: `server/boardstates.go`
- Modify: `server/games.go`
- Modify: `server/models_custom.go`
- Test: `server/boardstates_test.go`

**Step 1: Write failing backend tests**

Add tests proving:

- EDH games still round-trip the existing zones unchanged
- Generic Duel games can use a mapping layer between logical format zones and stored boardstate fields
- hidden zones remain hidden in the UI-facing data where appropriate

**Step 2: Run tests**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestUpdateBoardState|TestBoardStateFormat' -v
```

Expected: FAIL because backend still assumes Magic-shaped semantics.

**Step 3: Implement the smallest safe normalization layer**

Do **not** redesign the entire persistence schema in v1.

Instead, introduce a translation layer:

- canonical stored fields stay the same for now
- format registry maps logical zones to stored fields
- example:
  - `field` → `Battlefield`
  - `discard` → `Graveyard`
  - `banished` → `Exiled`
  - `deck` → `Library`

This is intentionally pragmatic. It gets multi-format support without a risky schema rewrite.

**Step 4: Run tests**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run 'TestUpdateBoardState|TestBoardStateFormat' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add server/boardstates.go server/games.go server/boardstates_test.go
git commit -m "feat: normalize boardstate zones by format"
```

---

### Task 10: Make game list/detail surfaces show format explicitly

**Files:**
- Modify: `app/src/stores/games.ts`
- Modify: `app/src/views/GamesView.vue`
- Modify: `app/src/views/BoardView.vue`
- Test: `app/__tests__/GamesView.spec.ts`

**Step 1: Write failing UI test**

Assert that game cards/list rows show:

- format name
- deck size or archetype summary
- not just the generic title/turn

**Step 2: Run test**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run GamesView.spec.ts
```

Expected: FAIL because format metadata is not surfaced.

**Step 3: Implement minimal UI surfacing**

Add:

- `Commander`
- `Generic Duel`

badges to games list and board header.

**Step 4: Run test**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test -- --run GamesView.spec.ts
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/src/views/GamesView.vue app/src/views/BoardView.vue app/__tests__/GamesView.spec.ts
git commit -m "feat: show game format across game surfaces"
```

---

### Task 11: Add migration/compatibility coverage for old games

**Files:**
- Modify: `server/games.go`
- Modify: `server/models_custom.go`
- Test: `server/games_test.go`
- Docs: `vedh/docs/game-formats.md`

**Step 1: Write failing compatibility tests**

Add tests for old games that:

- have no `format` rule
- have no `starting_life`
- use existing Commander-style boardstate only

Expected behavior:

- infer `EDH`
- preserve current functionality
- no migration required to load existing games

**Step 2: Run tests**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestLegacyGameDefaultsToEDH -v
```

Expected: FAIL before compatibility helpers are in place.

**Step 3: Implement compatibility helper**

Add `ensureGameFormatDefaults(game *Game)` called from existing load paths:

- `loadGameByID`
- `Games`
- any other game hydration path

That helper should:

- set `format=EDH` if absent
- set `deck_size=99` if absent on EDH
- normalize turn phase to a valid phase if empty

**Step 4: Write docs**

Create `vedh/docs/game-formats.md` describing:

- what a format is
- supported format IDs
- where defaults live
- how backward compatibility works
- how to add a new archetype

**Step 5: Run tests**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server -run TestLegacyGameDefaultsToEDH -v
```

Expected: PASS.

**Step 6: Commit**

```bash
git add server/games.go server/models_custom.go server/games_test.go vedh/docs/game-formats.md
git commit -m "docs: document game format compatibility"
```

---

### Task 12: Run full verification and do a safe frontend deploy

**Files:**
- Modify: any touched files above
- Verify: existing test suites

**Step 1: Run backend tests**

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test ./server ./... 
```

Expected: PASS.

**Step 2: Run frontend tests**

```bash
cd /root/.openclaw/workspace/vedh/app
npm test
```

Expected: PASS.

**Step 3: Run frontend build**

```bash
cd /root/.openclaw/workspace/vedh/app
npm run build
```

Expected: PASS.

**Step 4: Manual verification checklist**

Verify locally or against a deployed staging/prod environment:

- create Commander game
- create Generic Duel game
- join both games
- confirm board headers show correct format
- confirm zones differ by format
- confirm next-phase button label differs by format
- confirm Commander picker is hidden for Generic Duel
- confirm old EDH game still loads

**Step 5: Deploy frontend (when approved)**

```bash
cd /root/.openclaw/workspace/vedh
GIT_SSH_COMMAND='ssh -i /root/.ssh/id_ed25519 -o IdentitiesOnly=yes' git subtree push --prefix app dokku-app main
```

If backend GraphQL/schema changes are included and this repo’s server deploy path is also needed:

```bash
cd /root/.openclaw/workspace/vedh
GIT_SSH_COMMAND='ssh -i /root/.ssh/id_ed25519 -o IdentitiesOnly=yes' git push dokku main
```

**Step 6: Commit deploy-ready state**

```bash
git add .
git commit -m "feat: add multi-format game archetype support"
```

---

## Important design decisions baked into this plan

### 1. Use a registry, not freeform JSON first
A registry gives you controlled flexibility and a sane UX. Freeform editors can come later.

### 2. Keep `Game.Rules` as the persistence seam first
You already store game metadata there. Reusing it is lower-risk than a big schema rewrite.

### 3. Add a translation layer before redesigning `BoardState`
Current backend storage is Magic-shaped. A mapping layer gets multi-format support faster and more safely.

### 4. Commander becomes a capability, not the identity of the app
`CommanderEnabled` should gate commander-specific UI instead of Commander assumptions being spread everywhere.

### 5. Prove abstraction with one non-Magic archetype
If the system only supports Commander after the refactor, it isn’t really abstract yet.

## Explicit non-goals for v1

Do **not** include these in the first slice:

- arbitrary user-authored custom formats
- per-card rules enforcement for other games
- alternate win-condition engines by format
- format-specific card search providers
- radically different tabletop topologies like lane-based boards or hidden simultaneous action systems

## Best follow-up after this lands

Once this foundation is in, the best next slice is:

1. add **per-format counters/resources**
   - e.g. poison / energy / shields / lore / generic counters
2. add **layout presets**
   - duel / 4-player pod / vertical mobile compact
3. add **per-game light overrides**
   - rename a zone
   - hide a zone
   - tweak phase labels

That gets you a real tabletop engine without disappearing into schema philosophy.

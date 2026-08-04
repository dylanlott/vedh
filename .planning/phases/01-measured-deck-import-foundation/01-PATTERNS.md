# Phase 1: Measured Deck Import Foundation - Pattern Map

**Mapped:** 2026-08-04
**Files analyzed:** 17 new/modified
**Analogs found:** 13 / 17 (4 have no in-repo analog and are flagged explicitly)

> Sourced from `01-CONTEXT.md` (D-01..D-24) and `01-RESEARCH.md` §Recommended Project
> Structure (line 382) + §Code Examples (line 989). ACT-003 files are listed but are gated
> behind the D-14 `checkpoint:decision` — do not plan adapter code before it.

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `pkg/deckimport/scanner.go` | utility (pure) | transform | `pkg/games/games.go` | package-conventions only (weak) |
| `pkg/deckimport/sections.go` | utility (pure) | transform | `pkg/games/games.go` | package-conventions only (weak) |
| `pkg/deckimport/sourcetype.go` | utility (pure) | transform | `server/gamelog_helpers.go:8-23` (closed const set) + `server/authz.go:7-17` (closed map set) | role-match |
| `pkg/deckimport/result.go` | model | transform | `server/gamelog.go:27-33` (`Event` struct) | role-match |
| `pkg/deckimport/*_test.go` | test | — | `pkg/games/games_test.go` | exact (same dir, same CI target) |
| `pkg/deckimport/testdata/` | test fixture | file-I/O | **none** — no `testdata/` dir exists anywhere in the repo | no analog |
| `server/deck_import.go` | service + resolver | request-response / CRUD | `server/cards.go:140-230` (`Cards()` batch lookup) | exact |
| `server/deck_import_test.go` | test (DB-backed) | — | `server/games_test.go:1-40` | exact |
| `server/product_events.go` | service (writer) + resolver | event-driven (fire-and-forget) | `server/gamelog.go` + `server/gamelog_helpers.go:25-33` | exact |
| `server/product_events_test.go` | test | — | `server/graphql_metrics_test.go` (DB-free) | role-match |
| `server/metrics.go` | config (collector registry) | — | **none** — zero custom Prometheus collectors exist | no analog |
| `server/cards.go` | service | CRUD | *self* — widen existing SELECT at 162-167 | modify-in-place |
| `server/games.go` | service | transform | *self* — `createLibraryFromDecklist` 782-912 | modify-in-place |
| `server/schema.graphql` | config (contract) | — | `server/schema.graphql:34-59` (`Card`), `:15-25` (`Query`), `:160+` (`input Input*`) | exact |
| `server/deck_providers.go` (ACT-003, gated) | service | request-response (outbound) | **none** — no outbound HTTP client exists in `server/` | no analog |
| `app/src/services/productEvents.ts` | service (client) | event-driven | `app/src/services/commanderPartner.ts` (module shape) + `app/src/stores/auth.ts:13` (storage key) | role-match |
| `app/__tests__/productEvents.spec.ts` | test | — | `app/__tests__/commanderPartner.spec.ts` | exact |
| `persistence/migrations/*_product_events.{up,down}.sql` (×2 dirs) | migration | — | `persistence/migrations/20260131120000_gamelog_meta.{up,down}.sql` | exact |
| `persistence/migrations/*_card_name_search.{up,down}.sql` (×2 dirs) | migration | — | same as above | exact |

---

## The `pkg/` convention (highest-leverage finding)

`pkg/` contains **exactly one package**: `pkg/games/` (`games.go` 302 lines, `games_test.go`
248 lines). That is the entire body of evidence for the directory where RESEARCH.md wants
`pkg/deckimport/` to land. Its conventions, stated plainly:

| Convention | Evidence | Applies to `pkg/deckimport/`? |
|------------|----------|-------------------------------|
| Flat package, no subdirectories | only `games.go` + `games_test.go` | Yes — RESEARCH's 4-file flat layout matches |
| No DB, no `database/sql` import | imports are `fmt`, `log/slog`, `sync`, `time`, `github.com/google/uuid` | **Yes — load-bearing.** This is *why* `./pkg/...` runs in CI |
| Test lib is **`github.com/matryer/is`**, NOT testify | `pkg/games/games_test.go:7` | Yes. RESEARCH.md §Standard Stack recommends testify; `pkg/` precedent is `is`. `server/games_test.go:11-13` imports *both* `matryer/is` and `stretchr/testify/assert`, so either passes review — but `is` is the only lib `pkg/` has ever used |
| One top-level `TestX` with many `t.Run` subtests | `TestInMemoryGame` at `games_test.go:10` wraps 8 subtests | Optional — table-driven is a better fit for a 6-syntax corpus |
| **No `testdata/` directory, no golden files** | `find . -name testdata -type d` → zero results repo-wide; `grep -rl golden` → zero | **No precedent exists.** `pkg/deckimport/testdata/` is net-new. Nearest fixture precedent is `test/decklists/*.csv` (used by `server/` and by `persistence/migrations_test/20260131000000_seed_test_cards.up.sql`) |
| Interfaces declared before implementations, doc comment on every exported symbol | `GameService` (18), `Player` (34), `Game` (108), `PubSub` (273) | Yes |
| `slog.Default()` for logging (no injected logger) | `games.go:187, 290, 299` | The scanner should not log at all — it returns warnings as values |

**Why this matters:** `.github/workflows/test.yml` runs `make test-unit` only, which is
`$(GOTEST) -v ./pkg/... -race` (`Makefile:26-27`). `make test` → `test-api` → `./server/...`
is in **no workflow**. So `pkg/deckimport/` tests gate every PR and `server/` tests gate
nothing today.

---

## Pattern Assignments

### `pkg/deckimport/result.go` (model, transform)

**Analog:** `server/gamelog.go:27-33` — the repo's plain-struct-with-doc-comment style.

```go
// Event represents a change to boardstate
type Event struct {
	GameID  string
	Type    string
	Actor   string
	Payload map[string]interface{}
}
```

Note the exported-fields, no-json-tags, no-validation style. `ParsedEntry` / `ParsedDeck` /
`Warning` / `BlockingError` should follow it (RESEARCH.md:441-450 gives the field list).

**Sentinel error convention** (`server/gamelog.go:14-15`):

```go
// ErrEmpty is returned if an Event's payload is nil.
var ErrEmpty = fmt.Errorf("must provide an event payload")
```

`fmt.Errorf`, not `errors.New`, for package-level sentinels. Wrapping uses `%w`
(`gamelog.go:71`, `games.go:833`).

---

### `pkg/deckimport/sourcetype.go` (utility, transform — D-23/D-24 enum)

**Analog A — closed string-const set:** `server/gamelog_helpers.go:8-23`

```go
const (
	EventTypeGameCreated        = "GAME_CREATED"
	EventTypePlayerJoined       = "PLAYER_JOINED"
	EventTypeLifeChanged        = "LIFE_CHANGED"
	// ... 14 total, hardcoded, no runtime config
)
```

**Analog B — closed membership set + predicate:** `server/authz.go:7-17`

```go
var publicQueries = map[string]struct{}{
	"card":      {},
	"cards":     {},
	"search":    {},
	"searchAll": {},
}

func isPublicQuery(name string) bool {
	_, ok := publicQueries[name]
	return ok
}
```

Copy Analog A's shape for the four D-24 values, but give them a **named type**
(`type SourceType string`) rather than bare `string` — Pitfall 7 (RESEARCH.md:947) requires
that every `WithLabelValues` argument be a compile-time constant or a value of this enum
type, and a named type is what makes that reviewable. `gamelog_helpers.go` uses untyped
consts; deviate deliberately here and say why in a comment.

---

### `server/deck_import.go` (service + resolver, request-response)

**This file has the strongest analog in the phase: `server/cards.go:140-230`.** It is the
same shape — batch input, dedupe, one `pq.Array` query, scan into a keyed map.

**Resolver wiring — read this before writing any resolver.** `server/schema.resolvers.go`
is generated and every method in it `panic`s ("not implemented"). The real resolvers are
methods on `*graphQLServer`, bound by `Resolvers: s` at `server/graphql.go:265`:

```go
gqlHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
```

So `PreviewDeck` and `TrackProductEvent` are written as
`func (s *graphQLServer) PreviewDeck(ctx context.Context, input InputDeckImport) (*DeckPreview, error)`
in `server/deck_import.go` / `server/product_events.go` — **not** in `schema.resolvers.go`.
Precedent: `func (s *graphQLServer) Cards(ctx context.Context, list []string) ([]*Card, error)`
at `cards.go:140` is the `cards` query resolver.

**Dedupe pattern — copy verbatim** (`server/cards.go:141-157`), reused by RESEARCH Pattern 3
bound #2:

```go
trimmed := make([]string, 0, len(list))
seen := map[string]struct{}{}
for _, name := range list {
	n := strings.TrimSpace(name)
	if n == "" {
		continue
	}
	key := strings.ToLower(n)
	if _, ok := seen[key]; ok {
		continue
	}
	seen[key] = struct{}{}
	trimmed = append(trimmed, n)
}
if len(trimmed) == 0 {
	return []*Card{}, nil
}
```

**Batch query + `pq.Array` pattern** (`server/cards.go:159-173`) — the shape the suggestion
`LATERAL` query and the widened D-09 SELECT both follow:

```go
found := map[string]*Card{}
var combinedErr error

rows, err := s.db.Query(
	`SELECT name, id, colors, convertedmanacost, types, power, toughness, text, subtypes, supertypes, uuid, facename
	FROM cards
	WHERE name = ANY($1) OR facename = ANY($1);`,
	pq.Array(trimmed),
)
if err != nil {
	if !isMissingRelation(err, "cards") {
		combinedErr = errs.Combine(combinedErr, err)
	}
} else {
	defer rows.Close()
	for rows.Next() { /* ... */ }
}
```

Three things to carry forward and one to change:
- **Carry:** `s.db.Query` directly on `*sql.DB`, no repository layer.
- **Carry:** `errs.Combine` (`github.com/zeebo/errs`) for accumulating row errors.
- **Carry:** `isMissingRelation(err, "cards")` tolerance — the new `card_names` projection
  query should do the same (`isMissingRelation(err, "card_names")`) so a server running
  ahead of its migration degrades instead of failing. This is the existing idiom for
  exactly that situation.
- **Change:** use `QueryContext(qctx, ...)` not `Query(...)`. `cards.go` ignores the `ctx`
  it is handed; RESEARCH Pattern 3 bound #4 requires a 750 ms sub-context, which is
  impossible with `Query`. Flag this as a deliberate deviation.

**Scan style** (`server/cards.go:175-215`): every column into a `sql.NullString`, then
`nullStringPtr(x)` to get a `*string` for the GraphQL model. The four D-07 `Card` fields
follow this — declare `setcode`, `number`, `scryfallid` as `sql.NullString` locals and map
through `nullStringPtr`.

---

### `server/games.go` — `createLibraryFromDecklist` (MODIFIED, transform)

**No analog needed — this is the extraction target.** Excerpts the plan must reference:

**Preserve verbatim, commander budget (`games.go:802-815` + `840-851`):**

```go
commanderBudget := map[string]int64{}
commandersSpecified := int64(0)

for _, commander := range commanders {
	if commander == nil { continue }
	name := strings.TrimSpace(commander.Name)
	if name == "" { continue }
	commandersSpecified++
	commanderBudget[strings.ToLower(name)]++
}
// ...
key := strings.ToLower(name)
if removeCount := commanderBudget[key]; removeCount > 0 && quantity > 0 {
	if removeCount >= quantity {
		commanderBudget[key] = removeCount - quantity
		quantity = 0
	} else {
		quantity -= removeCount
		commanderBudget[key] = 0
	}
}
if quantity == 0 { continue }
```

**Preserve verbatim, size rule (`games.go:862-878`):**

```go
maxLibraryCards := int64(100) - commandersSpecified
if maxLibraryCards < 0 {
	return nil, fmt.Errorf("invalid commander count: %d", commandersSpecified)
}
var libraryCount int64
for _, entry := range entries { libraryCount += entry.qty }
if libraryCount > maxLibraryCards {
	return nil, fmt.Errorf("deck too large: library has %d cards but maximum is %d for %d commander(s)",
		libraryCount, maxLibraryCards, commandersSpecified)
}
```

**Delete and replace (`games.go:788-793`, the `csv.Reader`; and `games.go:904-908`):**

```go
r := csv.NewReader(strings.NewReader(trimmed))
r.LazyQuotes = true
r.TrimLeadingSpace = true
// ...
if found := lookup[key]; found != nil {
	cards = addX(entry.qty, cards, found)
} else {
	cards = addX(entry.qty, cards, &Card{Name: entry.name})  // ← D-06 deletes this branch
}
```

**Signature change (Pitfall 6, RESEARCH.md:940):** current signature is
`createLibraryFromDecklist(ctx, decklist string, commanders []*InputCard)`. It must become
`(ctx, parsed *deckimport.ParsedDeck, commanders []*InputCard)`. Callers: `server/games.go:552`
(CreateGame) and `:681` (JoinGame) — exactly two sites.

---

### `server/product_events.go` (service, event-driven)

**Analog:** `server/gamelog.go` (the writer) + `server/gamelog_helpers.go:25-33` (the
never-fail wrapper). This is the codebase's canonical fire-and-forget-write pattern and
D-17/D-22 are a direct extension of it.

**Never-fail wrapper** (`server/gamelog_helpers.go:25-33`):

```go
func (s *graphQLServer) logEvent(ctx context.Context, event Event) {
	if s == nil || s.db == nil {
		return
	}
	g := &pgLogger{db: s.db}
	if err := g.Add(ctx, event); err != nil {
		s.loggerFor(ctx).Warn("failed to write gamelog event", "err", err, "game_id", event.GameID, "type", event.Type)
	}
}
```

Copy exactly: no return value, `s == nil || s.db == nil` guard, `s.loggerFor(ctx).Warn`
with key-value slog pairs. `recordProductEvent` adds only the counter increment before each
`return` (RESEARCH.md:643-668).

**Writer + validation** (`server/gamelog.go:59-75`):

```go
func (g *pgLogger) Add(ctx context.Context, event Event) error {
	if event.Payload == nil { return ErrEmpty }
	if event.GameID == "" { return fmt.Errorf("missing game id") }
	if event.Type == "" { event.Type = "UNKNOWN" }
	query := `INSERT INTO gamelog (game_id, payload) VALUES($1, $2);`
	_, err := g.db.Exec(query, event.GameID, event)
	if err != nil {
		return fmt.Errorf("failed to add event to gamelog: %w", err)
	}
	return nil
}
```

Note: `ctx` is accepted and unused, and `db.Exec` is used rather than `ExecContext`. Follow
the interface shape but use `ExecContext` — the append-only-log role is identical, the
context handling is a bug this file should not inherit.

**JSONB metadata marshalling** — `Event` implements `driver.Valuer` (`gamelog.go:36-45`) so
it can be passed straight as a query arg. `product_events.metadata JSONB` can use the same
trick, or the simpler `payloadFrom` helper at `gamelog_helpers.go:35-45`.

**Interface + compile-time assertion** (`gamelog.go:11-25`):

```go
// Ensure that pgLogger fulfills EventLog
var _ EventLog = (*pgLogger)(nil)

type EventLog interface {
	Add(ctx context.Context, event Event) error
}
```

Worth copying for the product-event writer so it is fakeable in a DB-free test.

**Authoritative-event guard (D-20/D-21):** `server/authz.go` is the closest thing to an
allowlist gate. `requireAuth(ctx)` (`authz.go:21-26`) is how a resolver gets the
authenticated user the server attaches itself rather than trusting the client:

```go
func requireAuth(ctx context.Context) (*AuthUser, error) {
	if user, ok := authFromContext(ctx); ok {
		return user, nil
	}
	return nil, errors.New("authentication required")
}
```

`trackProductEvent` must *not* require auth (guests emit events) but must use
`authFromContext(ctx)` to attach `user_id` server-side, never read it from input.

---

### `server/metrics.go` (config) — NO ANALOG

There are **zero custom Prometheus collectors in this repo.** `server/graphql.go:295` is the
whole of the integration:

```go
if s.shouldExposeMetrics() {
	mux.Handle("/prometheus", s.withMetricsAuth(promhttp.Handler()))
} else if s != nil && s.cfg.MetricsEnabled {
	s.logger.Warn("prometheus metrics disabled because METRICS_TOKEN is empty")
} else {
	s.logger.Info("prometheus metrics disabled; set METRICS_ENABLED=true to expose /prometheus")
}
```

No `promauto`, no `MustRegister`, no `CounterVec` anywhere. The planner should take
RESEARCH.md Pattern 6 (line 675-719) as the source of truth and **not** look for an in-repo
model — there isn't one. The only in-repo constraint is that `promauto.New*` registers into
`prometheus.DefaultRegisterer`, which is what the already-mounted `promhttp.Handler()`
serves, so `Serve()` at `graphql.go:265-303` needs no edit.

**Test analog does exist:** `server/graphql_metrics_test.go` is DB-free (constructs
`&graphQLServer{cfg: Conf{...}}` directly, never touches `s.db`) and uses plain
`t.Fatalf` + a `for _, tc := range []struct{...}` table:

```go
func TestGraphQLServer_ShouldExposeMetrics(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsEnabled: true, MetricsToken: "secret-token"}}
	if !s.shouldExposeMetrics() {
		t.Fatalf("expected metrics to be enabled")
	}
	// ...
}
```

**Caveat:** this file still lives in `package server`, so `server/main_test.go`'s `TestMain`
runs first and hard-fails without Postgres — meaning even these DB-free tests do not
currently execute anywhere. The mechanical label-allowlist test (RESEARCH §Validation
Architecture) will inherit that problem unless it lives under `pkg/`.

---

### `server/schema.graphql` (config, contract)

**Existing `Card` type (lines 34-59)** — the four D-07 fields append here. Note the
PascalCase field convention, which is unusual for GraphQL and must be matched:

```graphql
type Card {
  FaceName: String
  Name: String!
  ID: String!
  Quantity: Int
  # ...
  TCGID: String
  ScryfallID: String
  ScreenX: Float
  ScreenY: Float
  CurrentZone: String
}
```

Add `SetCode: String`, `CollectorNumber: String`, `Category: String`, `SourceFormat: String`
— all nullable, matching the existing optional-metadata fields.

**Query field convention (lines 15-25)** — camelCase field names, PascalCase types:

```graphql
type Query {
  card(name: String!, id: String): Card
  cards(list: [String!]): [Card!]!
  search(name: String, colors: [String], ...): [Card]
}
```

`previewDeck(input: InputDeckImport!): DeckPreview!` fits this.

**Input convention (line 160+):** every input type is named `Input<Thing>` — `InputCard`,
`InputCreateGame`, `InputJoinGame`, `InputDeck`, `InputBoardState`. So `InputDeckImport` and
`InputProductEvent`, not `DeckImportInput`.

**Generation:** `gqlgen.yml` binds `autobind: github.com/openmtg/edh-go/server`, model output
`server/models_gen.go`, exec output `server/generated.go`, resolvers `layout: follow-schema`
into `server/`. Every schema edit must be followed by `make generate`
(`Makefile` → `go run github.com/99designs/gqlgen`). `generated.go` (323 KB),
`models_gen.go`, and `app/src/types/generated.ts` are never hand-edited. Expect
`make generate` to also append panicking stubs for `PreviewDeck`/`TrackProductEvent` into
`server/schema.resolvers.go`; that file is dead code (all 18 existing methods panic) and the
real implementations go on `*graphQLServer`.

---

### `app/src/services/productEvents.ts` (service, event-driven)

**Analog A — module shape:** `app/src/services/commanderPartner.ts`. Plain module, named
`export function`s, exported types at top, no Pinia, no default export:

```ts
export type CommanderPick = {
  ID: string;
  Name: string;
  Text?: string | null;
};

function normalize(text?: string | null): string {
  return (text ?? '').toLowerCase();
}

export function hasPartnerAbility(card: CommanderPick): boolean { /* ... */ }
```

**Analog B — `localStorage` key + defensive read:** `app/src/stores/auth.ts:13-25`

```ts
const STORAGE_KEY = 'edhgo/auth';

function loadPersistedProfile(): AuthProfile | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthProfile;
  } catch (error) {
    console.warn('[auth] failed to parse profile:', error);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}
```

Use `'edhgo/session-id'`. Note the `[auth]` log-prefix convention → use `[productEvents]`.

**Analog C — guarded `localStorage` access:** `app/src/services/apollo.ts:44-60` is the more
defensive pattern and the better model for a service that must never throw (D-19):

```ts
const getRawAuth = () => {
  try {
    if (typeof localStorage !== 'undefined' && localStorage) return localStorage.getItem('edhgo/auth');
  } catch (e) { /* ignore */ }
  try {
    if (typeof window !== 'undefined' && window.localStorage) return window.localStorage.getItem('edhgo/auth');
  } catch (e) { /* ignore */ }
  return null;
};
```

The comment there — "In test environments `localStorage` may not exist on the global" — is
the reason. `getSessionID()` must degrade the same way.

**Analog D — swallow-and-continue on network failure:** `app/src/services/scryfall.ts:37-52`
loops candidates and `catch { /* ignore and try next */ }`, never throwing to the caller.
Same disposition D-19 wants, minus the retry.

**Mutation document location:** gql documents live in `app/src/graphql/mutations.ts`, not
inline in services. Follow `LOGIN_MUTATION` (`mutations.ts:3-11`):

```ts
import { gql } from '@apollo/client/core';

export const LOGIN_MUTATION = gql`
  mutation Login($username: String!, $password: String!) {
    login(username: $username, password: $password) { ID Username Token }
  }
`;
```

Add `TRACK_PRODUCT_EVENT_MUTATION` there and import it into `productEvents.ts`. The client
is `import { apolloClient } from '../services/apollo'`.

---

### `app/__tests__/productEvents.spec.ts` (test)

**Analog:** `app/__tests__/commanderPartner.spec.ts` — exact match (services unit test).

```ts
import { describe, it, expect } from 'vitest';
import { hasPartnerAbility, isValidPartnerPair } from '../src/services/commanderPartner';

describe('commanderPartner helpers', () => {
  it('detects Partner keyword', () => {
    expect(hasPartnerAbility({ ID: '1', Name: 'A', Text: 'Partner' })).toBe(true);
  });
});
```

**Load-bearing constraint the planner will otherwise get wrong:** `app/package.json`
`vitest.include` is `["__tests__/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}"]`. A test
colocated at `app/src/services/productEvents.spec.ts` **will not run**. It must go in
`app/__tests__/`. Environment is `jsdom` with `globals: true`, so `localStorage` exists in
tests. Command is `npm test` → `vitest --run`.

---

### `persistence/migrations/` — four files, two directories

**Analog:** `20260131120000_gamelog_meta.{up,down}.sql` — the most recent pair and the only
one that does anything non-trivial.

**Up (idempotent-everything convention):**

```sql
ALTER TABLE IF EXISTS gamelog
    ADD COLUMN IF NOT EXISTS id BIGSERIAL,
    ADD COLUMN IF NOT EXISTS game_id TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints tc
        WHERE tc.table_name = 'gamelog' AND tc.constraint_type = 'PRIMARY KEY'
    ) THEN
        ALTER TABLE gamelog ADD PRIMARY KEY (id);
    END IF;
END$$;

UPDATE gamelog SET game_id = '' WHERE game_id IS NULL;
```

**Down (mirror, `IF EXISTS` throughout, reverse order):**

```sql
ALTER TABLE IF EXISTS gamelog
    ALTER COLUMN eventtime DROP DEFAULT;

ALTER TABLE IF EXISTS gamelog
    DROP COLUMN IF EXISTS game_id,
    DROP COLUMN IF EXISTS id;
```

Conventions to copy:
- Filename `YYYYMMDDHHMMSS_snake_name.{up,down}.sql`; timestamps are strictly increasing
  (latest is `20260131120000`, so pick something after it).
- `IF EXISTS` / `IF NOT EXISTS` on every statement, both directions.
- No transaction wrappers (golang-migrate handles it).
- Down is a true inverse, written in reverse statement order.
- `BIGSERIAL PRIMARY KEY` for append-only log tables — already the `gamelog` shape, and what
  RESEARCH Pattern 4 specifies for `product_events`.

**Parity requirement, precisely:** `persistence/migrations/` has 17 SQL files;
`persistence/migrations_test/` has 19. The test dir mirrors every production pair *plus*
`20260131000000_seed_test_cards.{up,down}.sql`, and is *missing* `20210307154621_init_db.down.sql`
(a pre-existing gap — do not "fix" it in this phase). For this phase: copy each new
`.up.sql`/`.down.sql` byte-identically into both directories → 8 files total for the two
migrations. `persistence/migrations_test/README.md` documents the `migrate ... force` reset
procedure and states the dir "is not used by production or `make migrate-local`".

**`atlas.sum`:** already stale (lists 14 files, missing two existing migrations). Leave it
alone — RESEARCH.md:515.

---

## Shared Patterns

### Never-fail-the-caller writes
**Source:** `server/gamelog_helpers.go:25-33`
**Apply to:** `server/product_events.go` (D-17, D-22), the suggestion path in
`server/deck_import.go` (D-01)

```go
if err := g.Add(ctx, event); err != nil {
	s.loggerFor(ctx).Warn("failed to write gamelog event", "err", err, "game_id", event.GameID, "type", event.Type)
}
```

Void return, `s.loggerFor(ctx).Warn`, structured key-value slog args. Every new failure path
in this phase that must not surface to the user copies this.

### Structured logging
**Source:** `server/games.go:823, 884`; `server/gamelog_helpers.go:31`
**Apply to:** all new `server/` files

```go
s.loggerFor(ctx).Warn("batch card lookup failed", "err", err)
s.loggerFor(ctx).Warn("error reading csv record", "err", err)
```

`s.loggerFor(ctx)` (request-scoped, carries the X-Request-Id set by
`s.withRequestID` at `graphql.go:277`), never `log.Printf`, never a package-level logger.
In `pkg/`, the precedent is `slog.Default()` (`pkg/games/games.go:187`) — but
`pkg/deckimport` should log nothing at all and return warnings as values.

### Error wrapping
**Source:** `server/games.go:833`, `server/gamelog.go:71`
**Apply to:** all new Go files

```go
return nil, fmt.Errorf("failed to parse quantity: %w", err)
return fmt.Errorf("failed to add event to gamelog: %w", err)
```

Lowercase message, `%w` verb. Multi-error accumulation uses `errs.Combine`
(`github.com/zeebo/errs`, `server/cards.go:170`).

### Go test style
**Source:** `pkg/games/games_test.go` (`matryer/is`), `server/games_test.go`
(`is` + `testify/assert`), `server/graphql_metrics_test.go` (stdlib `t.Fatalf`)
**Apply to:** all new `_test.go`

```go
is := is.New(t)
t.Run("should create a game", func(t *testing.T) {
	created, err := m.NewFullGame("test", []Player{})
	is.NoErr(err)
	is.True(len(m.games) > 0)
})
```

Subtest names are lowercase "should ..." phrases. DB-backed server tests get their handle
from `testAPI(t)` and their context from `authCtx(mastershake)`, and register cleanup
inline (`server/games_test.go:32-36`):

```go
t.Cleanup(func() {
	query := `DELETE FROM games WHERE id = $1;`
	_, err = api.db.Exec(query, seedGameID)
	assert.NoError(t, err)
})
```

### Public (unauthenticated) resolver registration
**Source:** `server/authz.go:7-17`
**Apply to:** `previewDeck` and `trackProductEvent`

```go
var publicQueries = map[string]struct{}{
	"card": {}, "cards": {}, "search": {}, "searchAll": {},
}
```

`previewDeck` must be added to this map (guests, pre-auth, are the whole point of the
activation funnel). `trackProductEvent` is a mutation — check whether the enforcement site
that calls `isPublicQuery` covers mutations at all; if it does not, that is a gap the plan
must close explicitly rather than assume.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `server/metrics.go` | config | — | Zero custom Prometheus collectors exist. `graphql.go:295` mounts `promhttp.Handler()` and nothing else. No `promauto`, `MustRegister`, `CounterVec`, or `HistogramVec` anywhere in non-generated code. Use RESEARCH.md Pattern 6 (line 675). |
| `server/deck_providers.go` | service | request-response (outbound) | No outbound HTTP client exists in `server/`. The only `fetch`/HTTP-client code in the repo is browser-side (`app/src/services/scryfall.ts`), which shares no security model. Use RESEARCH.md Pattern 8 (line 775) verbatim. **Gated behind the D-14 checkpoint — do not plan this file's contents yet.** |
| `pkg/deckimport/testdata/` | test fixture | file-I/O | No `testdata/` directory exists anywhere in the repo and no golden-file helper exists. Nearest precedent is `test/decklists/*.csv` (consumed by `server/` tests and by `persistence/migrations_test/20260131000000_seed_test_cards.up.sql`). `test/decklists/jarad.csv` is a required fixture regardless — it contains the four `//` double-faced names Pitfall 2 turns on. |
| `pkg/deckimport/scanner.go` grammar | utility | transform | `pkg/games/games.go` supplies package conventions only (flat layout, `matryer/is`, doc comments, no DB) — it is an in-memory pub/sub store and shares no parsing logic. The grammar itself has no in-repo model; the closest thing is the `csv.Reader` block at `server/games.go:788-860` being *replaced*. Use RESEARCH.md Pattern 1 (line 415). |

---

## Metadata

**Analog search scope:** `pkg/`, `server/`, `app/src/services/`, `app/src/stores/`,
`app/src/graphql/`, `app/__tests__/`, `persistence/migrations/`,
`persistence/migrations_test/`, `.github/workflows/`, `Makefile`, `gqlgen.yml`
**Files read in full or in targeted ranges:** 21
**Pattern extraction date:** 2026-08-04

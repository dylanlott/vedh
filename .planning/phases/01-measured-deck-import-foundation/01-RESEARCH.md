# Phase 1: Measured Deck Import Foundation - Research

**Researched:** 2026-08-04
**Domain:** Go/PostgreSQL server-side deck parsing, trigram similarity matching, Prometheus instrumentation, SSRF-hardened outbound HTTP
**Confidence:** HIGH (stack, in-repo facts, PostgreSQL extension behavior) / MEDIUM (SSRF implementation patterns, similarity cost bounds)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Deck import semantics — unresolved cards**

- **D-01:** A deck containing cards the lookup cannot match is **not blocked**. The
  preview flags every unresolved entry and `CanContinue` stays `true`, so the player can
  create or join anyway. Rationale: the ≥60% host activation target is the reason this
  milestone exists; a single typo must not stop someone at the door.

- **D-02:** Near-misses get **up to 3 ranked candidate suggestions** the player must
  explicitly accept. The parser may never auto-apply a suggestion — this upholds the
  standing constraint that fuzzy matching "may never silently rewrite a user's deck".
  Note `s.Cards()` (`server/cards.go:162`) is exact-match SQL today (`WHERE name = ANY($1)`),
  so a similarity mechanism is net-new work.

- **D-03:** When nothing clears the similarity cutoff, **show the nearest matches anyway,
  visually marked low-confidence** rather than offering nothing. Accepting one still
  requires an explicit player action, so D-02's no-silent-rewrite guarantee holds.

- **D-04:** Suggestions are computed **eagerly and returned inside the `previewDeck`
  response**, not fetched lazily per correction. One round trip; correction is instant.
  — **Reversibility:** costly — moving to lazy later means a second resolver, a changed
  `DeckPreview` shape, and a client already written against the eager contract.

- **D-05:** Unresolved entries are **excluded from the deck-size count**. Only matched
  cards count toward the `100 - commanders` cap enforced at `server/games.go:862`.

- **D-06:** Unresolved entries are **omitted from the created library**. A 100-card paste
  with 3 unmatched names produces a 97-card library. This is chosen for consistency with
  D-05 so the preview count and the board contents never disagree. It deliberately
  changes today's behavior at `server/games.go:904-908`, which silently substitutes a bare
  `&Card{Name:}` and tells nobody.

**Deck export parsing and printing metadata**

- **D-07:** Set codes, collector numbers, and category tags in real exports
  (`1 Sol Ring (C21) 263`, `1x Sol Ring (c21) 263 [Ramp]`) are **parsed and persisted**,
  not stripped. The normalized entry and the stored `Card` carry `setCode`,
  `collectorNumber`, `category`, and `sourceFormat`.
  — **Reversibility:** costly — undo touches the parser, the Go `Card` struct, the
  GraphQL `Card` type, and every game payload already written with these fields.

- **D-08:** Persisting D-07 requires **no migration and no schema change**. `upsertGame`
  (`server/games.go:746`) stores the whole game as a JSONB payload
  (`INSERT INTO games (id, payload) VALUES ($1, $2::jsonb)`), so new `Card` fields
  persist automatically. This decision is recorded explicitly because the SPEC's schema
  constraint names only `product_events` and the `is_guest`/`expires_at` user fields —
  D-07 does **not** violate it, and does not "replace the board-state model" that
  PROJECT.md marks out of scope. It extends a schemaless payload.

- **D-09:** Card lookup **disambiguates on `(name, setcode, number)`** when the paste
  supplies them, falling back to name-only when it does not. The `cards` table already
  has `setcode`, `number`, `scryfallid`, and `uuid` (one row per printing), but the
  current SELECT at `server/cards.go:163` omits them and keys results by name — so which
  printing you get today is arbitrary (last row wins). This makes an existing implicit
  choice deliberate; it does not add printing selection to a system that lacked it.

- **D-10:** A paste naming a printing **absent from the MTGJSON snapshot** resolves to
  any available printing of that name, **with a warning** that the requested printing was
  not found. The parsed `setCode` is preserved on the entry so a later MTGJSON refresh or
  the future proxy-print feature can reconcile it. It is explicitly *not* treated as
  unresolved — a stale snapshot must not silently shrink decks whenever a new set appears.

- **D-11:** `Sideboard` / `Maybeboard` sections are **dropped with a visible count**
  ("12 sideboard cards ignored"). Correct for singleton Commander, and satisfies the
  success criterion that no nonblank row disappears without becoming a total, warning, or
  error.

- **D-12:** A `Commander` section header **preselects the commander candidates** for
  player confirmation, feeding ACT-002's `CommanderCandidates`. The player can always
  override. This uses an explicit signal the player already gave rather than rediscovering
  it from card data.

**Provider feasibility gate (ACT-003)**

- **D-13:** The spike starts **neutral** — Archidekt and Moxfield are evaluated against
  the same criteria and evidence decides. OPEN-1 remains genuinely open until the spike
  reports; no candidate is favoured in planning.

- **D-14:** ACT-003 is planned as a **spike task plus decision record, then a
  `checkpoint:decision`** before any adapter code is written. The plan must assume
  neither outcome. Do not pre-write adapter tasks, and do not pre-write the paste-only
  no-go tasks — the checkpoint is where the branch is chosen.

- **D-15:** Beyond the SSRF and reliability controls already locked by constraint, the
  **only hard gate condition is a stable response shape**. Terms-of-service permission,
  rate limits, and credential requirements are recorded in the decision record for the
  human approving at D-14's checkpoint, but none of them independently fails the gate.

- **D-16:** The adapter **may use app-level credentials that vEDH holds** (an API key or
  registered account). This is distinct from user credentials, which PROJECT.md still
  forbids ("will not ask users to share credentials"). Implies a server-side secret in
  the Dokku deploy path. ACT-013's release-gate tests remain credential-free via
  fixtures, per the standing constraint that no test may depend on a live provider.
  — **Reversibility:** costly — undo means removing secret-management plumbing from
  deploy, and may invalidate whichever provider was selected at the D-14 checkpoint.

**Product events and metrics**

- **D-17:** A failed product-event write **never fails the user's action**. Game creation
  succeeds regardless; the failure is logged and increments a Prometheus rejection/drop
  counter. Protects the <2% create/join error-rate target absolutely.
  **Accepted cost and its mitigation:** funnel denominators can under-count during an
  incident, so the ≥60% activation rate is measured against imperfect data. The drop
  counter is what makes this survivable — it lets a real metric dip be distinguished from
  a measurement gap. The counter is therefore **required**, not optional.

- **D-18:** The random session ID lives in **`localStorage`** — one ID per browser,
  surviving tabs, reloads, and restarts. Counts humans rather than tabs, and makes the
  ≥15%-claim-within-7-days metric measurable across visits. A random non-identifying
  analytics ID being client-readable is acceptable. Matches where the auth token already
  lives, so no new storage pattern is introduced.
  — **Reversibility:** costly — changing storage later orphans every existing session ID
  and breaks funnel continuity across the change.

- **D-19:** Client events are delivered **fire-and-forget, one mutation per event**.
  Keeps the `board_ready` and `quick_start_viewed` timestamps sharp, which batching would
  blur — and those timestamps *are* the time-to-board metrics. The server already
  de-duplicates authoritative conversion events per constraint, so client retry safety is
  covered.

- **D-20:** The event vocabulary is **closed at exactly the PRD's 15 events** (9 client:
  `landing_primary_cta`, `quick_start_viewed`, `deck_import_started`, `game_create_started`,
  `invite_copied`, `invite_viewed`, `join_started`, `board_ready`, `account_claim_started`;
  6 server: `deck_import_succeeded`, `deck_import_failed`, `guest_session_created`,
  `game_created`, `player_joined`, `account_claimed`). A hardcoded constant, not runtime
  config. Adding an event later requires a code change and deploy — acceptable, because
  every downstream ticket through ACT-012 already knows which events it emits.

- **D-21:** Metadata keys are allowlisted **per event**, not as a global union. Each event
  declares exactly which keys it may carry (`board_ready` gets session ID, game ID, user
  role, elapsed ms; `quick_start_viewed` gets session ID only). The PRD already specifies
  the per-event keys, so the table is transcription rather than invention.

- **D-22:** An event failing the allowlist is **dropped, logged, and counted** in
  Prometheus — no GraphQL error returned. Consistent with D-19: nothing is listening for a
  result, so a counter is the only observability surface that would actually be read.

**Source-type taxonomy**

- **D-23:** **One shared source-type enum, two detectors.** Phase 1 defines the enum and a
  paste-shape detector; Phase 2's ACT-004 adds a URL-host detector producing the same enum.
  This guarantees the event metadata, the Prometheus label, and the persisted `Card` field
  mean the same thing regardless of how a deck arrived — which is what makes paste and link
  funnels comparable. Prevents Phase 2 inventing `moxfield_url` alongside `moxfield`.
  — **Reversibility:** costly — the enum values become Prometheus label values and land in
  historical `product_events` rows and persisted game payloads; renaming them later splits
  funnel data that cannot be retroactively relabelled.

- **D-24:** Enum granularity is **provider-level**: `moxfield`, `archidekt`, `plain_text`,
  `unknown`. Small and bounded, so it is safe as a Prometheus label under the
  no-high-cardinality constraint, and it answers the product question that would actually
  drive adding a second adapter — where players keep their decks.

### Claude's Discretion

- **`previewDeck` rate-limiting scope.** The SPEC already requires IP/session rate limiting
  on guest creation, deck import, and public invite lookup. Whether the new `previewDeck`
  resolver gets its own limit, shares deck import's, or stays unlimited until Phase 2
  introduces guests was raised and deliberately not discussed. Planner's call — but note
  `previewDeck` sits on the activation critical path, so a limit tuned too tight would
  throttle the funnel this phase exists to measure.
- **Similarity algorithm for D-02/D-03.** The cutoff value, the scoring function, and
  whether it is implemented as PostgreSQL `pg_trgm`, in-Go distance over a candidate set,
  or something else. Only the *product surface* was decided (up to 3, always show nearest,
  eager). Note D-03 plus D-04 together define the worst-case cost path — see Specific Ideas.
- **Paste-shape detection heuristics for D-23.** Which line features distinguish
  `moxfield` from `archidekt` from `plain_text`, and what an ambiguous paste resolves to
  (`unknown` is available in the enum).

### Deferred Ideas (OUT OF SCOPE)

- **Custom proxy prints for specific cards** — the motivation behind D-07's persistence, and
  a genuinely appealing feature, but a new user-facing capability that belongs in its own
  phase well after activation ships. Phase 1 only ensures the printing identity needed to
  build it is captured rather than discarded. Do not implement any proxy or custom-art
  behavior in this phase.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REQ-ACT-001 | Product event and activation metric foundation — `product_events` table (prod + test migrations), `trackProductEvent` with strict server-side event and field allowlists, server-authoritative conversion events, low-cardinality Prometheus counters and latency histograms, frontend event service with a persisted random session ID, documented vocabulary and example funnel query. | §Architecture Pattern 4 (`product_events` schema, PG14 dedup index without `NULLS NOT DISTINCT`), Pattern 5 (closed vocabulary + per-event key allowlist in Go), Pattern 6 (Prometheus collector registration on client_golang v1.11.0, bucket selection), Pattern 7 (frontend event service, `edhgo/` localStorage namespace), Validation Architecture (mechanical no-high-cardinality-label test, mechanical forbidden-key test) |
| REQ-ACT-002 | Canonical deck parser and preview API — extract parsing out of `createLibraryFromDecklist`, support all six syntaxes plus headers/blank lines/sideboard, preserve comma-containing names, return normalized entries + commander candidates + unresolved + warnings + blocking errors + `CanContinue`, add `previewDeck`/`DeckPreview`, feed the same normalized result into create/join, preserve lookup/commander-removal/deck-size rules. | §Architecture Pattern 1 (hand-rolled line scanner grammar, `//` disambiguation), Pattern 2 (pg_trgm GiST KNN for D-02/D-03), Pattern 3 (bounded eager-suggestion cost path), §Don't Hand-Roll, §Common Pitfalls 1–5, Validation Architecture (table-driven fixture corpus) |
| REQ-ACT-003 | Public deck provider feasibility gate and first adapter — one-day time-boxed comparison, decision record, adapter interface keyed by an allowlisted hostname, HTTPS/DNS-IP validation/redirect revalidation/3s connect/8s total/1 MiB cap, normalize through ACT-002, server-side feature flag and kill switch. | §Architecture Pattern 8 (two-layer SSRF: `ControlContext` owns IP/port, `CheckRedirect` owns scheme/host), §Common Pitfalls 8–11, §Security Domain, Validation Architecture (SSRF tests with zero live-provider dependency), §Open Questions OPEN-1 (what the spike must measure — provider choice deliberately unresolved) |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

**No `CLAUDE.md` exists** — verified absent at both `./CLAUDE.md` and `./.claude/CLAUDE.md`, and no `.claude/skills/` or `.agents/skills/` directory exists. The authoritative directives for this phase are therefore the 18 constraints in `.planning/intel/constraints.md` and the 24 locked decisions above. Treat those with the authority a CLAUDE.md would carry.

## Summary

Phase 1 is a **brownfield surgical extraction**, not new construction. Every one of the three tickets grafts onto an existing, working seam: `createLibraryFromDecklist` is called from exactly two places (`server/games.go:552` in `CreateGame`, `server/games.go:681` in `JoinGame`), which is why "one canonical parser serves host and join" is achievable without touching either mutation; `upsertGame` (`server/games.go:746`) already writes the whole game as JSONB so D-07's printing metadata persists for free; and `/prometheus` is already deployed behind `withMetricsAuth` so ACT-001 registers collectors into a secured surface rather than building one. The dominant risk in this phase is not "can we build it" — it is **getting a handful of small, load-bearing details right the first time**, because several of them (the source-type enum values, the `product_events` column set, the eager `DeckPreview` shape) are explicitly marked costly to reverse.

Four findings materially change how this phase should be planned. First, **`agnivade/levenshtein` is already in the build graph** — reachable from `server` via `gqlparser/v2/validator` — so string-distance capability costs zero new supply-chain surface, and `pg_trgm`/`fuzzystrmatch` are both `trusted = true` in PostgreSQL 14 contrib, meaning the migration can `CREATE EXTENSION` as the ordinary app user with no superuser escalation. Second, **D-03 decides the index type**: "always show nearest matches even below cutoff" cannot be served by a GIN `%` prefilter (below-threshold queries return zero rows), but is exactly what GiST's `<->` KNN ordering does natively — the PostgreSQL docs state KNN GiST "will usually beat the first formulation when only a small number of the closest matches is wanted," and three is a small number. Third, **the `cards` table has no index on `name` or `facename`** (only `cards_uuid_key` on `uuid` and the PK on `id`), so today's `WHERE name = ANY($1)` batch lookup is a sequential scan over ~100k rows on every single deck import — fixing this is a prerequisite for the latency targets, not an optimization. Fourth, **CI does not run the server test suite at all**: `.github/workflows/test.yml` runs `make test-unit` = `go test ./pkg/...`, and `./server/...` requires both a live Postgres and a ~500 MB `All Printings.json` that is not in the repo. Any test that must gate this phase has to be DB-free or CI has to change.

The parser itself should be hand-rolled. There is no maintained Go library for MTG decklist parsing, the six required syntaxes plus set codes, collector numbers, category tags, section headers, and `//` double-faced names are a small deterministic grammar, and PROJECT.md's "deck parsing must be deterministic, testable, and explainable" argues against a dependency regardless. The one genuinely subtle rule is that `//` is *both* the MTGO/Arena comment prefix and the double-faced-card name separator — and the repo's own fixture `test/decklists/jarad.csv` contains `1,Bala Ged Recovery // Bala Ged Sanctuary` unquoted, so a naive "strip everything after `//`" rule silently truncates a real card the project already ships as test data.

**Primary recommendation:** Build ACT-002's parser as a pure, DB-free line scanner in `pkg/deckimport/` (so it runs in the existing CI via `make test-unit`) and keep DB-bound resolution and suggestion in `server/deck_import.go`; back D-02/D-03 with a `pg_trgm` **GiST** index over a distinct-name projection table (~33k rows, not the ~97k-row printings table) queried once per preview through a single `LATERAL` batch capped at 25 distinct unresolved names with its own 750 ms sub-context; and implement ACT-003's SSRF controls as two composed layers — `net.Dialer.ControlContext` owning IP/port policy and `http.Client.CheckRedirect` owning scheme/host policy — with `http.Transport.Proxy` explicitly set to `nil`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Decklist grammar / tokenization | API / Backend (pure Go, no DB) | — | Constraint: one canonical parser serves host and join; the deck must not be parsed twice under divergent rules. A client-side parser would be a second implementation by definition. |
| Card name → printing resolution | Database / Storage | API / Backend | The `cards` table is the only source of truth for what a valid card name is. Backend owns the query shape and the `(name, setcode, number)` disambiguation of D-09. |
| Fuzzy suggestion ranking (D-02/D-03) | Database / Storage | — | `pg_trgm` GiST KNN does the ranking inside the index. Pulling 33k names into Go to rank them would move data to compute for no benefit. |
| `previewDeck` / `DeckPreview` contract | API / Backend | — | Locked GraphQL contract; downstream ACT-004 is written against this shape. |
| Deck-size and commander-removal rules | API / Backend | — | Already enforced at `server/games.go:862`; must be preserved, not relocated. |
| Product event **persistence** and validation | API / Backend | Database / Storage | Constraint: server-side allowlists; server attaches authenticated user IDs rather than accepting them from the client. Validation in the browser is unenforceable. |
| Product event **emission** for view/start/readiness | Browser / Client | API / Backend | Only the browser knows when a view happened or when the board became usable. D-19 keeps timestamps sharp by sending one mutation per event. |
| Random session ID generation and persistence | Browser / Client | — | D-18: `localStorage`, one ID per browser. The server cannot generate a value that survives across visits without a cookie, which is out of scope. |
| Server-authoritative conversion events | API / Backend | — | Constraint: `game_created`, `player_joined`, `guest_session_created`, `deck_import_succeeded`, `deck_import_failed`, `account_claimed` are server-authoritative; duplicate client retries must not duplicate them. |
| Technical counters and latency histograms | API / Backend | — | Constraint: Prometheus/Grafana cover technical SLIs only; PostgreSQL owns funnel analysis. |
| Outbound provider fetch + SSRF controls | API / Backend | — | A browser-side fetch would expose the allowlist to bypass and would not be subject to any of the locked network controls. |
| Provider response → normalized deck | API / Backend | — | Constraint: normalize provider output *through* ACT-002, so the provider path shares the canonical parser. |
| Deck import UI, commander review UI | *(Phase 2 — ACT-004)* | — | Explicitly out of scope. `ui.plan-gate` returned `frontend: false` for this phase. |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go standard library (`bufio`, `strings`, `unicode`) | 1.24.0 | Decklist line scanner | `[VERIFIED: go.mod:4 — "go 1.24.0"]` No maintained Go MTG-decklist parser exists (§Don't Hand-Roll); the grammar is small and PROJECT.md requires it be "deterministic, testable, and explainable". |
| PostgreSQL `pg_trgm` | bundled with PG 14 (`default_version = '1.6'`) | Trigram similarity index for D-02/D-03 candidate ranking | `[VERIFIED: postgres/postgres REL_14_STABLE contrib/pg_trgm/pg_trgm.control — "default_version = '1.6'", "trusted = true"]` Only PostgreSQL index type that answers "3 nearest names" without a sequential scan. GiST `<->` supports the below-cutoff case D-03 requires. |
| `github.com/prometheus/client_golang` | v1.11.0 (already pinned) | Counters and latency histograms | `[VERIFIED: go.mod:22 — "github.com/prometheus/client_golang v1.11.0"]` Already a direct dependency; `promhttp.Handler()` is already mounted at `server/graphql.go:295`. |
| `github.com/lib/pq` | v1.8.0 (already pinned) | `pq.Array` for batched `text[]` parameters | `[VERIFIED: go.mod:20 — "github.com/lib/pq v1.8.0"]` Already used for the batch card lookup at `server/cards.go:166` (`pq.Array(trimmed)`). |
| Go `net/http` + `net` + `net/netip` | stdlib 1.24 | SSRF-hardened outbound fetch | `[VERIFIED: go doc net.Dialer — "ControlContext func(ctx context.Context, network, address string, c syscall.RawConn) error"]` The `Control`/`ControlContext` hook is the canonical Go SSRF defense; no third-party package is required. |
| `github.com/golang-migrate/migrate/v4` | v4.14.1 (already pinned) | `product_events` and trigram-index migrations | `[VERIFIED: go.mod:14]` Established convention: 20 existing migration pairs across `persistence/migrations/` and `persistence/migrations_test/`. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/agnivade/levenshtein` | v1.2.1 (**already in build graph, indirect**) | Optional in-Go re-rank of the GiST candidate set | `[VERIFIED: go.mod:29 "github.com/agnivade/levenshtein v1.2.1 // indirect"; go mod why -m → github.com/openmtg/edh-go/server → gqlparser/v2 → validator → validator/core → agnivade/levenshtein]` API is a single function: `func ComputeDistance(a, b string) int` `[VERIFIED: $GOMODCACHE/github.com/agnivade/levenshtein@v1.2.1/levenshtein.go:20]`. Use **only** to break ties among the ≤3 candidates pg_trgm already returned — never to scan the name set. Promoting it to direct adds zero new supply-chain surface. |
| `github.com/stretchr/testify` | v1.11.1 (already pinned) | Table-driven parser assertions | `[VERIFIED: go.mod:23]` Already used across `server/*_test.go`. |
| `prometheus/client_golang/prometheus/testutil` | (bundled with v1.11.0) | Metric assertions | `[VERIFIED: $GOMODCACHE/.../prometheus/testutil/testutil.go — ToFloat64:78, CollectAndCount:122, GatherAndCount:138, CollectAndCompare:157, GatherAndCompare:169]` |
| `golang.org/x/time/rate` | v0.15.0 latest; already in `go.sum` transitively | Token-bucket rate limiting if the planner scopes a `previewDeck` limit | `[VERIFIED: go list -m -versions golang.org/x/time → ... v0.12.0 v0.13.0 v0.14.0 v0.15.0]` `[VERIFIED: grep -c "golang.org/x/time" go.sum → 4]` Not in `go.mod` today. Official Go namespace. **No rate limiting of any kind exists in this repo yet** — `grep -rn "limiter\|Limiter\|golang.org/x/time" server/ --include=*.go` returns nothing outside generated files. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `pg_trgm` GiST (`gist_trgm_ops`) | `pg_trgm` GIN (`gin_trgm_ops`) | GIN is faster for the `%` boolean prefilter at ~100k rows, but **cannot serve D-03**: below the similarity threshold `%` returns zero rows, so "always show the nearest matches anyway" would need a second, unbounded query. GiST's `<->` KNN ordering returns the 3 nearest unconditionally. `[CITED: postgresql.org/docs/current/pgtrgm.html — distance ordering "can be implemented quite efficiently by GiST indexes, but not by GIN indexes… will usually beat the first formulation when only a small number of the closest matches is wanted"]` |
| `pg_trgm` | `fuzzystrmatch` `levenshtein()` | `levenshtein()` and `levenshtein_less_equal()` are **not index-accelerated** `[CITED: postgresql.org/docs/current/fuzzystrmatch.html]` — every suggestion is a sequential scan over the card table. Also capped at 255 characters per argument. Use only as an optional post-filter on an already-small candidate set. |
| `pg_trgm` | In-Go Levenshtein over the full name set | 32,643 distinct paper card names `[VERIFIED: api.scryfall.com/cards/search?q=game:paper&unique=cards → "total_cards":32643, queried 2026-08-04]` × up to 100 query names = 3.26M `ComputeDistance` calls per preview, plus a name-cache warming and MTGJSON-refresh invalidation problem the DB does not have. Viable only behind an aggressive prefilter, which is what the trigram index already is. |
| Hand-rolled line scanner | Adopt an existing parser | No maintained Go MTG-decklist library exists (see §Don't Hand-Roll). The nearest reference implementation, `lheyberger/mtg-parser`, is Python/pyparsing and does not handle `1x` at all (`QUANTITY = Word(nums)`) `[VERIFIED: raw.githubusercontent.com/lheyberger/mtg-parser/main/src/mtg_parser/grammar.py]`. |
| `encoding/csv` (current implementation) | — | Structurally cannot satisfy REQ-A2. `csv.Reader` yields `["1","Atraxa"," Praetors' Voice"]` for `1,Atraxa, Praetors' Voice`, and the code takes `record[1]` and discards the rest `[VERIFIED: server/games.go:830 — "name := strings.TrimSpace(record[1])"]`; `1 Sol Ring` produces a single field and hits `if len(record) < 2 { return nil, ... }` `[VERIFIED: server/games.go:827-829]`. |
| Third-party SSRF package (`code.dny.dev/ssrf`, `doyensec/safeurl`) | stdlib `ControlContext` | The prefix deny-lists in those packages are worth transcribing, but adding a dependency for ~60 lines of `netip` prefix checks is not justified when the locked constraints are already fully specified and must be unit-tested regardless. |

**Installation:**

```bash
# Promote the already-present indirect dependency to direct (no new module downloaded):
go get github.com/agnivade/levenshtein@v1.2.1

# Only if the planner scopes a previewDeck rate limit:
go get golang.org/x/time@v0.15.0

# PostgreSQL extension — inside the migration, not a shell step:
#   CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

**Version verification performed this session:**

```
$ go list -m -versions github.com/agnivade/levenshtein   # → v1.1.0 v1.1.1 v1.2.0 v1.2.1 (v1.2.1 already pinned)
$ go list -m -versions golang.org/x/time                 # → ... v0.12.0 v0.13.0 v0.14.0 v0.15.0
$ go mod why -m github.com/agnivade/levenshtein          # → reachable from github.com/openmtg/edh-go/server
```

## Package Legitimacy Audit

> The `gsd-tools package-legitimacy check` seam supports `npm|pypi|crates` only and rejects `--ecosystem go` (`Error: Usage: gsd-tools package-legitimacy check --ecosystem <npm|pypi|crates> <pkg1> ...`). Verification below was performed with the ecosystem-appropriate tooling: the Go module proxy via `go list -m -versions` and the build-graph provenance check `go mod why -m`.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `github.com/agnivade/levenshtein@v1.2.1` | Go module proxy | v1.1.0 → v1.2.1 release line; latest is v1.2.1 | n/a (Go proxy publishes no download counts) | github.com/agnivade/levenshtein | OK | Approved — **already in the build graph** transitively from `gqlparser/v2/validator`, which is itself pulled by gqlgen. Promoting to direct downloads nothing new. |
| `golang.org/x/time@v0.15.0` | Go module proxy | Official Go sub-repository namespace | n/a | go.googlesource.com/time | OK | Approved *conditionally* — only add if the planner scopes a `previewDeck` rate limit. Already present in `go.sum` (4 entries) as a transitive dependency. |
| `pg_trgm` (PostgreSQL contrib) | PostgreSQL core distribution | Ships with PostgreSQL itself | n/a | git.postgresql.org/postgresql `contrib/pg_trgm` | OK | Approved — control file fetched and read this session; `trusted = true` confirms no superuser required. |
| `fuzzystrmatch` (PostgreSQL contrib) | PostgreSQL core distribution | Ships with PostgreSQL itself | n/a | git.postgresql.org/postgresql `contrib/fuzzystrmatch` | OK | Not recommended (not index-accelerated), but legitimate if the planner wants an optional post-filter. `trusted = true` confirmed. |

**Packages removed due to [SLOP] verdict:** none — no package in this research was discovered via WebSearch or training data alone. Every entry is either already in `go.mod`/`go.sum`, or is a PostgreSQL contrib module whose control file was fetched from the `postgres/postgres` repository during this session.

**Packages flagged as suspicious [SUS]:** none.

## Architecture Patterns

### System Architecture Diagram

```
                       ┌──────────────────────── BROWSER (ACT-001 scope only) ─────────────────────┐
                       │  productEvents.ts service                                                 │
                       │    localStorage 'edhgo/session-id' ──► sessionID                           │
                       │    track(name, meta) ─fire-and-forget─► trackProductEvent mutation         │
                       │    (NO UI components this phase — DeckImportPanel is ACT-004/Phase 2)      │
                       └───────────────────────────────────┬───────────────────────────────────────┘
                                                           │ GraphQL
  ┌────────────────────────────────────────────────────────▼──────────────────────────────────────────┐
  │ GO API SERVER                                                                                      │
  │                                                                                                    │
  │   previewDeck(InputDeckImport{text|sourceURL, sessionID})       createGame ──┐   joinGame ──┐      │
  │        │                                                                     │              │      │
  │        ├──► text? ──────────────────────────────┐                            ▼              ▼      │
  │        │                                        │                    ┌──────────────────────────┐  │
  │        └──► sourceURL? ──► [ACT-003 GATE]       │                    │ SAME normalized entries  │  │
  │                             feature flag OFF?   │                    │ (no second parse)        │  │
  │                             ──► normalized      │                    └────────────┬─────────────┘  │
  │                                 error +         │                                 │                │
  │                                 paste fallback  │                                 │                │
  │                             flag ON:            │                                 │                │
  │                             ┌───────────────────▼────────┐                        │                │
  │                             │ SSRF-hardened fetch        │                        │                │
  │                             │  CheckRedirect: scheme+host│                        │                │
  │                             │  ControlContext: IP+port   │                        │                │
  │                             │  3s connect / 8s total     │                        │                │
  │                             │  io.LimitReader 1 MiB+1    │                        │                │
  │                             └───────────┬────────────────┘                        │                │
  │                                         │ provider body                           │                │
  │                                         ▼                                         │                │
  │   ┌─────────────────────────────────────────────────────────────────┐             │                │
  │   │ pkg/deckimport  (PURE — no DB, runs in existing CI)              │             │                │
  │   │   scanner: line ──► {qty, name, setCode, collNum, category}      │             │                │
  │   │   section tracker: Commander / Sideboard / Maybeboard            │             │                │
  │   │   source-shape detector ──► SourceType enum (TELEMETRY ONLY)     │             │                │
  │   │   every nonblank line ──► entry | warning | blocking error       │             │                │
  │   └────────────────────────────┬────────────────────────────────────┘             │                │
  │                                │ ParsedDeck                                       │                │
  │                                ▼                                                  │                │
  │   ┌─────────────────────────────────────────────────────────────────┐             │                │
  │   │ server/deck_import.go  (DB-bound resolution)                     │             │                │
  │   │   ① exact resolve: (name,setcode,number) ─► name-only fallback   │             │                │
  │   │   ② D-10: printing missing ─► any printing + warning             │             │                │
  │   │   ③ unresolved set ─► dedupe ─► cap 25 ─► ONE LATERAL query      │             │                │
  │   │        (750ms sub-context; timeout ⇒ 0 suggestions + warning)    │             │                │
  │   │   ④ D-05/D-06: unresolved excluded from count AND library        │             │                │
  │   │   ⑤ commander removal + (100 - commanders) cap PRESERVED         │             │                │
  │   └──────────┬────────────────────────────────────┬─────────────────┘             │                │
  │              │                                    │                                │                │
  │              ▼                                    ▼                                ▼                │
  │   ┌────────────────────┐            ┌──────────────────────────┐      ┌────────────────────────┐   │
  │   │ product event      │            │ Prometheus collectors    │      │ upsertGame  (JSONB)    │   │
  │   │ writer             │            │  labels: event/outcome/  │      │  Card{setCode,         │   │
  │   │  ① name in closed  │            │          source/role/    │      │   collectorNumber,     │   │
  │   │     15-event set?  │            │          reason ONLY     │      │   category,            │   │
  │   │  ② keys in PER-    │            │  NEVER: user/session/    │      │   sourceFormat}        │   │
  │   │     EVENT allowlist│            │         game/attribution │      │  ⇒ no migration (D-08) │   │
  │   │  ③ fail ⇒ drop +   │            └──────────────┬───────────┘      └────────────────────────┘   │
  │   │     log + counter  │                           │                                                │
  │   │     (NEVER error)  │                           │                                                │
  │   └─────────┬──────────┘                           │                                                │
  └─────────────┼──────────────────────────────────────┼────────────────────────────────────────────────┘
                ▼                                      ▼
      ┌──────────────────────┐            /prometheus (withMetricsAuth,
      │ POSTGRES 14          │             METRICS_ENABLED + METRICS_TOKEN)
      │  product_events      │
      │   +(event,time) idx  │
      │   +(session,time) idx│    ┌──────────────────────────────────────────┐
      │   +partial UNIQUE    │    │ cards  (~97k printings, ~33k names)      │
      │    dedup (COALESCE   │    │  NEW: idx on lower(name), lower(facename)│
      │    — PG14 has no     │    │  NEW: card_names projection + GiST trgm  │
      │    NULLS NOT DISTINCT)│   └──────────────────────────────────────────┘
      └──────────────────────┘
```

### Recommended Project Structure

```
pkg/deckimport/                  # NEW — pure, DB-free. Runs in existing CI (`make test-unit`).
├── scanner.go                   #   line grammar: the six syntaxes + set/collector/category
├── sections.go                  #   Commander / Sideboard / Maybeboard header handling
├── sourcetype.go                #   D-23/D-24 enum + paste-shape detector (telemetry only)
├── result.go                    #   ParsedDeck, ParsedEntry, Warning, BlockingError
└── testdata/                    #   the six-syntax corpus + real-export golden fixtures

server/
├── deck_import.go               # NEW — DB-bound: resolution, D-09 disambiguation, suggestions
├── deck_import_test.go          # NEW — DB-backed integration tests
├── deck_providers.go            # NEW (ACT-003, behind the D-14 checkpoint) — adapter iface + safe client
├── product_events.go            # NEW — closed vocabulary, per-event key allowlist, writer
├── product_events_test.go       # NEW
├── metrics.go                   # NEW — all collectors in one file, one registration site
├── cards.go                     # MODIFIED — wider SELECT (setcode/number/scryfallid), name index use
├── games.go                     # MODIFIED — createLibraryFromDecklist consumes ParsedDeck
├── schema.graphql               # MODIFIED — previewDeck, DeckPreview, trackProductEvent, Card fields
└── testdata/deck_providers/     # NEW — fixture bodies; no live provider, ever

app/src/services/
└── productEvents.ts             # NEW — sessionID + fire-and-forget track(). NO components.

persistence/migrations/          # 2 files per migration, mirrored in migrations_test/
├── <ts>_product_events.up.sql
├── <ts>_product_events.down.sql
├── <ts>_card_name_search.up.sql
└── <ts>_card_name_search.down.sql
```

**Why `pkg/deckimport/` and not all of `server/`:** `.github/workflows/test.yml` runs only `make test-unit`, which is `go test ./pkg/... -race` `[VERIFIED: Makefile lines 26-27 — "test-unit:\n\t$(GOTEST) -v ./pkg/... -race"]`. `go test ./server/...` is not in any workflow, and cannot run without a live Postgres — `TestMain` calls `persistence.ForceCleanMigrations` and then `importAllPrintingsForTests`, which hard-fails on a missing `../All Printings.json` `[VERIFIED: server/main_test.go:18-38; persistence/import_all_printings_json.go:85 — 'return 0, fmt.Errorf("open json file: %w", err)']`. Confirmed empirically this session: `go test ./server/... -run TestNothingZZZ` → `test setup failed: dial tcp [::1]:5432: connect: connection refused`. Putting the deterministic grammar in `pkg/` makes REQ-A2's acceptance criteria gate every PR today with no CI change. The planner may *additionally* extend CI to run `./server/...` against a `postgres:14` service (the pattern already exists in `.github/workflows/smoke-rust.yml`), but the AllPrintings dependency makes that a larger change than this phase needs.

---

### Pattern 1: Line-oriented decklist scanner (replaces `encoding/csv`)

**What:** A hand-rolled per-line scanner with an explicit, ordered set of rules. Not a regex soup, not a CSV reader, not a parser-combinator library.

**When to use:** Every deck import path — paste and (if ACT-003 clears) provider. The canonical-parser constraint means there is exactly one of these.

**The rule order that makes all six syntaxes work.** Process each line independently, in this order:

1. **Trim** trailing `\r` (Windows pastes) and surrounding whitespace. Empty → skip silently (blank lines are the one thing allowed to vanish).
2. **Comment / section header check, anchored at line start only.** A line beginning with `//`, `#`, or `Sideboard`/`Maybeboard`/`Commander`/`Deck`/`Companion` (optionally followed by `(N)`) is a header or comment, never an entry. **The anchor is load-bearing**: `//` also separates double-faced card names, and `test/decklists/jarad.csv` ships `1,Bala Ged Recovery // Bala Ged Sanctuary` unquoted `[VERIFIED: test/decklists/jarad.csv — "1,Bala Ged Recovery // Bala Ged Sanctuary"]`. `lheyberger/mtg-parser` resolves this identically, by requiring `StringStart()` before the comment literal `[VERIFIED: raw.githubusercontent.com/lheyberger/mtg-parser/main/src/mtg_parser/grammar.py — "COMMENT_LINE = (StringStart() + (Literal('//!') | Literal('//') | Literal('#')).suppress() ..."]`.
3. **Quantity extraction.** Consume leading digits. Then consume an optional single `x`/`X` (covers `1x`). Then consume an optional single `,` (covers `1,Sol Ring`). Then consume any run of spaces/tabs (covers `1, Sol Ring` and `1 Sol Ring`). Absent leading digits → default quantity 1 and treat the whole line as a name (real exports contain bare names).
4. **Quoted-name check.** If the remainder starts with `"`, read to the matching close quote and take that verbatim as the name — this and only this is where CSV quoting semantics apply, and it covers `1,"Atraxa, Praetors' Voice"`.
5. **Trailing-annotation extraction, right to left**, off the *unquoted* remainder. Strip and record, in this order: `#tag` tokens; a trailing `` `Category` `` (Archidekt backtick form `[CITED: proxyfoundry.com — "1 Sol Ring \`Maybeboard\`"]`); a trailing `[Category]`; a trailing bare collector number (alphanumeric, follows a set group); a trailing `(SET)` group. What remains is the name.
6. **Name is everything left, trimmed** — commas inside it are just characters. This is what fixes `1 Atraxa, Praetors' Voice` and `1,Atraxa, Praetors' Voice`: after step 3 consumes `1,`, the rest is the whole name including its internal comma.
7. **Emit** an entry, or — if quantity parsing produced something impossible (negative, non-numeric where digits were expected) — a *warning* attached to that line, never a whole-deck failure. Today every error path is `return nil, err` `[VERIFIED: server/games.go:824, 828, 833, 836 — all four are "return nil, fmt.Errorf(...)"]`, which is precisely the behavior ACT-002 replaces.

**Invariant to assert in code, not just in tests:** `len(entries) + len(warnings) + len(blockingErrors) + len(droppedSectionRows) == countOfNonBlankInputLines`. This is the mechanical form of "no nonblank input row may disappear."

**Example:**

```go
// pkg/deckimport/scanner.go
// Grammar shape corroborated against lheyberger/mtg-parser's pyparsing definition:
//   QUANTITY + CARD_NAME + optional (SET) + optional COLLECTOR_NUMBER + ZeroOrMore(#TAG)
// Extended here with: `1x` quantities, `,`/`, ` separators, quoted names,
// [Category] and `Category` annotations — none of which mtg-parser handles.
type ParsedEntry struct {
	Quantity        int
	Name            string  // commas preserved verbatim
	SetCode         string  // D-07, "" when absent
	CollectorNumber string  // D-07, "" when absent
	Category        string  // D-07, "" when absent
	Section         Section // SectionMain | SectionCommander | SectionSideboard | SectionMaybeboard
	SourceLine      int     // 1-based; every warning cites this
	RawLine         string  // for the no-row-disappears audit
}
```

---

### Pattern 2: `pg_trgm` GiST KNN over a distinct-name projection

**What:** A narrow projection table holding one row per distinct searchable card name, with a GiST trigram index, queried by `<->` distance ordering.

**When to use:** Computing D-02's ≤3 ranked candidates, including D-03's below-cutoff case.

**Why a projection table and not the `cards` table directly.** The `cards` table holds one row per *printing*: ~96,589 paper printings against ~32,643 distinct card names `[VERIFIED: api.scryfall.com/cards/search?q=game:paper&unique=prints → "total_cards":96589; &unique=cards → "total_cards":32643 — both queried 2026-08-04]`. Indexing printings means a ~3× larger index and a result set where the top 3 rows are frequently three printings of the *same* card — useless as three suggestions. A distinct-name projection over `name` ∪ `facename` is ~33–35k rows, gives distinct suggestions for free, and shrinks the index proportionally.

**Why GiST and not GIN — this is decided by D-03, not by benchmarking.** With `gin_trgm_ops`, the accelerated operator is `%`, which is boolean against `pg_trgm.similarity_threshold`; a query string below the threshold matches **zero rows**, so D-03's "show the nearest matches anyway" would require a second query with no index support. With `gist_trgm_ops`, `ORDER BY name <-> $needle LIMIT 3` is an index-ordered KNN scan that always returns the 3 nearest regardless of threshold — exactly D-02 + D-03 in one operator. `[CITED: postgresql.org/docs/current/pgtrgm.html — GiST supports `<->`, `<<->`, `<<<->` for KNN; "This can be implemented quite efficiently by GiST indexes, but not by GIN indexes… It will usually beat the first formulation when only a small number of the closest matches is wanted."]`

The similarity **cutoff** then becomes a presentation concern rather than a query concern: return all 3 with their scores, and mark those below the cutoff low-confidence. That is precisely the shape D-03 asks for, and it means the cutoff can be tuned later without touching SQL.

**Example:**

```sql
-- persistence/migrations/<ts>_card_name_search.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- One row per distinct searchable name (front names and face names both).
CREATE TABLE IF NOT EXISTS card_names (
    name_lower TEXT PRIMARY KEY,
    display    TEXT NOT NULL
);

INSERT INTO card_names (name_lower, display)
SELECT DISTINCT ON (lower(n)) lower(n), n
FROM (
    SELECT name     AS n FROM cards WHERE name     IS NOT NULL AND name     <> ''
    UNION ALL
    SELECT facename AS n FROM cards WHERE facename IS NOT NULL AND facename <> ''
) s
ON CONFLICT (name_lower) DO NOTHING;

CREATE INDEX IF NOT EXISTS card_names_trgm_gist
    ON card_names USING GIST (name_lower gist_trgm_ops);

-- The exact-match path is a sequential scan today: `cards` has only
-- a PK on id and cards_uuid_key on uuid. Both lookups need real indexes.
CREATE INDEX IF NOT EXISTS cards_name_lower_idx     ON cards (lower(name));
CREATE INDEX IF NOT EXISTS cards_facename_lower_idx ON cards (lower(facename));
CREATE INDEX IF NOT EXISTS cards_name_set_num_idx   ON cards (lower(name), setcode, number);  -- D-09
```

```sql
-- ONE query for ALL unresolved names. $1 is a text[] of at most 25 distinct
-- lowercased needles. WITH ORDINALITY preserves the caller's ordering so the
-- Go side can map results back without a second lookup.
SELECT q.ord,
       q.needle,
       c.display,
       1 - (c.name_lower <-> q.needle) AS score
FROM   unnest($1::text[]) WITH ORDINALITY AS q(needle, ord)
CROSS JOIN LATERAL (
    SELECT cn.display, cn.name_lower
    FROM   card_names cn
    ORDER  BY cn.name_lower <-> q.needle
    LIMIT  3
) c;
```

**Down migration:** drop the two indexes and `card_names`, but **do not `DROP EXTENSION pg_trgm`** — a later migration or another feature may depend on it, and dropping a shared extension in a down migration is a footgun. Note that `persistence/migrations/atlas.sum` is already stale (it lists 14 files and is missing `20231018205455_add_games_id_unique.*` and `20260131120000_gamelog_meta.*`) `[VERIFIED: persistence/migrations/atlas.sum — 14 entries, no gamelog_meta]`, so it is evidently not enforced; do not attempt to regenerate it as part of this phase.

---

### Pattern 3: Bounding the eager-suggestion cost path (D-03 × D-04)

**What:** A four-part bound that makes the worst case a fixed constant instead of a function of deck size.

**When to use:** Inside `previewDeck`, wrapping the Pattern 2 query.

The cost is **proportional to the number of distinct unresolved names, not to deck size** — a correct 100-card deck runs zero KNN probes. Four bounds, in order of how much they buy:

| Bound | Mechanism | Effect |
|-------|-----------|--------|
| **1. Only unresolved names** | Suggestions are computed after exact resolution, over the failure set only | The ≥98% deck-resolution target means the expected case is 0–2 probes, not 100 |
| **2. Deduplicate** | Key the needle set by `lower(name)` before querying | `1 Sol Rng` × 4 becomes one probe. `server/cards.go:142-153` already does exactly this dedupe for the exact lookup — reuse the shape |
| **3. Cap at 25 distinct needles** | Beyond 25, suggest for the first 25 in input order and emit a warning: *"Showing suggestions for the first 25 of N unresolved cards."* | Caps the worst case at 25 KNN probes. 25 is ~12× headroom over the ≥98%-resolution target's implied ≤2 unresolved per 100-card deck, while turning the pathological all-garbage paste into a constant |
| **4. One batched query with its own sub-budget** | Single `LATERAL` statement (Pattern 2), executed under `context.WithTimeout(ctx, 750*time.Millisecond)` | One round trip instead of 25. On timeout: return **zero** suggestions plus a warning, and let the preview succeed — D-01's "never block the player" applied to the suggestion path itself |

**The bound that matters most is #1**, and it is worth stating explicitly in the plan because it inverts the intuition in CONTEXT.md's Specific Ideas: the 100-row worst case is not the normal case scaled up, it is a distinct pathological case (a garbage paste, or an MTGJSON snapshot that failed to load). Bound #3 is what stops that case from consuming a request slot.

**Instrument the bound so it is observable rather than assumed:** a histogram `vedh_deck_suggestion_duration_seconds` and a counter `vedh_deck_suggestion_truncated_total` (incremented when bound #3 fires) turn "is 25 the right cap" into a question the Grafana panels answer after beta, instead of a guess defended in review.

---

### Pattern 4: `product_events` table and index design

**What:** One append-only table, two btree indexes that serve the actual funnel queries, and one PG14-compatible dedup index.

**When to use:** ACT-001's migration. Four files across `persistence/migrations/` and `persistence/migrations_test/`, per the migration-parity constraint.

**Column types are dictated by the existing schema, and getting them wrong fails at insert time.** `users.uuid` is `VARCHAR(255)`, not a `uuid` column `[VERIFIED: schema.hcl:216-219 — 'column "uuid" { null = true  type = character_varying(255) }'; corroborated by persistence/migrations/20210307154621_init_db.up.sql — "uuid VARCHAR(255) UNIQUE"]`. `games.id` is `TEXT` `[VERIFIED: schema.hcl:175-178 — 'column "id" { null = true  type = text }']`. A `product_events.user_id uuid` column would reject every write.

**Do not add foreign keys.** ACT-005 (Phase 2) introduces an "independently retryable cleanup path for unreferenced expired guests" that deletes user rows. An FK from `product_events.user_id → users.uuid` would either block that cleanup or, with `ON DELETE CASCADE`, silently erase funnel history for exactly the guests the ≥15%-claim metric is about. It would also make an event write fail when the referenced row is gone, violating D-17's absolute "a failed product-event write never fails the user's action."

**Example:**

```sql
-- persistence/migrations/<ts>_product_events.up.sql
CREATE TABLE IF NOT EXISTS product_events (
    id          BIGSERIAL PRIMARY KEY,
    event_name  TEXT        NOT NULL,
    session_id  TEXT        NOT NULL,
    user_id     VARCHAR(255),            -- matches users.uuid; deliberately NO foreign key
    game_id     TEXT,                    -- matches games.id;    deliberately NO foreign key
    role        TEXT,                    -- 'host' | 'invitee' | ...
    source      TEXT,                    -- the D-23/D-24 enum: moxfield|archidekt|plain_text|unknown
    outcome     TEXT,                    -- 'success' | 'failure' | ...
    duration_ms INTEGER,
    metadata    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Funnel step counts grouped by day / role / source / outcome.
CREATE INDEX IF NOT EXISTS product_events_name_time_idx
    ON product_events (event_name, occurred_at DESC);

-- The index the funnel actually depends on: correlating quick_start_viewed ──► board_ready
-- within one session, and computing the time-to-board percentiles. Easy to omit; costly to omit.
CREATE INDEX IF NOT EXISTS product_events_session_time_idx
    ON product_events (session_id, occurred_at);

-- Dedup for server-authoritative conversion events.
-- PostgreSQL 14 has NO `UNIQUE NULLS NOT DISTINCT` (added in 15), so NULL game_id/user_id
-- would otherwise defeat the constraint entirely. COALESCE expressions are the PG14 form.
CREATE UNIQUE INDEX IF NOT EXISTS product_events_authoritative_once
    ON product_events (event_name,
                       COALESCE(game_id, ''),
                       COALESCE(user_id, ''),
                       session_id)
    WHERE event_name IN ('game_created',
                         'player_joined',
                         'guest_session_created',
                         'account_claimed');
```

`[VERIFIED: postgresql.org/docs/15/release-15.html §E.19.3.1.2 — "Allow unique constraints and indexes to treat NULL values as not distinct (Peter Eisentraut)… this can now be changed by creating constraints and indexes using UNIQUE NULLS NOT DISTINCT."]` Both the dev compose file and the smoke-rust CI service pin `postgres:14` `[VERIFIED: dev.docker-compose.yml:39 — "image: postgres:14"; .github/workflows/smoke-rust.yml — "image: postgres:14"]`, so the PG15 form is unavailable.

Insert with `INSERT INTO product_events (...) VALUES (...) ON CONFLICT DO NOTHING` — the untargeted form works against any unique index, including the partial one, and the two `deck_import_*` events are deliberately outside the predicate because a player may legitimately import several times.

`TIMESTAMPTZ` is a deliberate divergence from the legacy `TIMESTAMP` columns (`games.eventtime`, `users.timestamp`): day-bucketing a funnel across timezones is only correct with an absolute instant. Group with `date_trunc('day', occurred_at AT TIME ZONE 'UTC')`.

---

### Pattern 5: Closed vocabulary + per-event key allowlist in Go

**What:** One package-level table that is simultaneously the event vocabulary (D-20), the per-event metadata key allowlist (D-21), and the test fixture.

**When to use:** `server/product_events.go`. This table *is* the privacy control — the negative requirement ("raw deck text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game state cannot be stored") is satisfied structurally, because no allowlisted key is named for any of those things. That is a far stronger guarantee than a denylist, and it is mechanically testable (see §Validation Architecture).

**Example:**

```go
// server/product_events.go
type eventSpec struct {
	Authoritative bool                // server-emitted only; client attempts are rejected
	Keys          map[string]struct{} // D-21: per event, not a global union
}

// D-20: exactly 15. A hardcoded constant, not runtime config.
var eventVocabulary = map[string]eventSpec{
	// --- 9 client events ---
	"landing_primary_cta":   {Keys: keys("session_id", "utm_source", "utm_medium", "utm_campaign", "referrer_host")},
	"quick_start_viewed":    {Keys: keys("session_id")},
	"deck_import_started":   {Keys: keys("session_id", "source")},
	"game_create_started":   {Keys: keys("session_id", "role")},
	"invite_copied":         {Keys: keys("session_id", "game_id", "share_method", "source")},
	"invite_viewed":         {Keys: keys("session_id", "game_id")},
	"join_started":          {Keys: keys("session_id", "game_id")},
	"board_ready":           {Keys: keys("session_id", "game_id", "role", "elapsed_ms")},
	"account_claim_started": {Keys: keys("session_id")},
	// --- 6 server-authoritative events ---
	"deck_import_succeeded": {Authoritative: true, Keys: keys("session_id", "source", "card_count", "unresolved_count", "duration_ms")},
	"deck_import_failed":    {Authoritative: true, Keys: keys("session_id", "source", "reason", "duration_ms")},
	"guest_session_created": {Authoritative: true, Keys: keys("session_id")},
	"game_created":          {Authoritative: true, Keys: keys("session_id", "game_id", "role", "source")},
	"player_joined":         {Authoritative: true, Keys: keys("session_id", "game_id", "role")},
	"account_claimed":       {Authoritative: true, Keys: keys("session_id")},
}

const maxMetadataValueLen = 128 // bounds attribution, and stops smuggling a decklist through an allowed key
```

> **Provenance:** the *set of 15 event names* is `[VERIFIED: .planning/phases/01-measured-deck-import-foundation/01-CONTEXT.md D-20 — the 15 names are enumerated verbatim there]`. The *per-event key assignments* above are `[ASSUMED]` except for `board_ready` (session ID, game ID, user role, elapsed ms) and `quick_start_viewed` (session ID only), which D-21 states explicitly. D-21 says "The PRD already specifies the per-event keys, so the table is transcription rather than invention" — the executor must transcribe the remaining 13 from `docs/product/2026-07-23-deck-to-game-activation-prd.md`, not from this table. See §Assumptions Log A1.

**Validation order in the writer, matching D-17 and D-22:**

```go
func (s *graphQLServer) recordProductEvent(ctx context.Context, e ProductEvent) {
	// Mirrors server/gamelog_helpers.go:25-33 — logEvent() already establishes
	// "log a warning and return" as this codebase's never-fail-the-caller pattern.
	spec, ok := eventVocabulary[e.Name]
	if !ok {
		productEventsDropped.WithLabelValues("unknown_event").Inc()
		s.loggerFor(ctx).Warn("product event dropped: unknown event name", "event", e.Name)
		return // D-22: dropped, logged, counted. No GraphQL error.
	}
	for k, v := range e.Metadata {
		if _, allowed := spec.Keys[k]; !allowed {
			productEventsDropped.WithLabelValues("unknown_key").Inc()
			s.loggerFor(ctx).Warn("product event dropped: key not allowlisted", "event", e.Name, "key", k)
			return // whole event dropped, per D-22 — not the offending key alone
		}
		if len(v) > maxMetadataValueLen {
			productEventsDropped.WithLabelValues("oversized_value").Inc()
			return
		}
	}
	if err := s.insertProductEvent(ctx, e); err != nil {
		productEventsDropped.WithLabelValues("write_error").Inc()
		s.loggerFor(ctx).Warn("product event write failed", "err", err, "event", e.Name)
		return // D-17: the user's action succeeds regardless.
	}
}
```

Note the drop counter's label is a bounded `reason` (`unknown_event` | `unknown_key` | `oversized_value` | `write_error` | `client_sent_authoritative`) — never the offending event name or key, which an attacker controls and could use to explode cardinality. That subtlety is itself a pitfall; see §Common Pitfalls 7.

---

### Pattern 6: Prometheus collectors on client_golang v1.11.0

**What:** All collectors declared in one file with `promauto`, so registration has exactly one site and the label-allowlist test has exactly one thing to walk.

**When to use:** `server/metrics.go`. There are **zero custom collectors today** — `server/graphql.go:295` mounts only `promhttp.Handler()` `[VERIFIED: server/graphql.go — "mux.Handle(\"/prometheus\", s.withMetricsAuth(promhttp.Handler()))"]`, so everything in ACT-001 is net-new.

**Version constraints that matter.** The pin is v1.11.0 `[VERIFIED: go.mod:22]`, which is well before native histograms; only classic explicit buckets exist. `promauto.With(prometheus.Registerer)` returns a `Factory` `[VERIFIED: $GOMODCACHE/github.com/prometheus/client_golang@v1.11.0/prometheus/promauto/auto.go:251,258 — "type Factory struct" / "func With(r prometheus.Registerer) Factory"]`, which is how tests get an isolated registry instead of mutating the global default.

**Bucket selection.** `prometheus.DefBuckets` is `{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}` seconds `[VERIFIED: $GOMODCACHE/github.com/prometheus/client_golang@v1.11.0/prometheus/histogram.go:68]`. Two adjustments earn their keep for this phase:

- **Put a boundary exactly at each timeout.** The provider fetch has a locked 3 s connect and 8 s total budget; buckets `{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 12}` make "everything piling into the 3–5 s and 8–12 s bands" instantly readable as connect-timeouts and total-timeouts, which `DefBuckets` (no 3, no 8) cannot show.
- **`previewDeck` is faster than `DefBuckets` assumes.** A DB-bound preview should land in tens of milliseconds; `{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 4, 8}` keeps resolution where the mass actually is while still bounding the tail.

**Example:**

```go
// server/metrics.go — one file, one registration site.
var (
	// Label sets are deliberately tiny and closed. `source` takes ONLY the
	// four D-24 enum values; `outcome` and `reason` take only enumerated constants.
	deckImportTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vedh_deck_import_total",
		Help: "Deck import attempts by source type and outcome.",
	}, []string{"source", "outcome"}) // NEVER session_id / user_id / game_id

	deckImportDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vedh_deck_import_duration_seconds",
		Help:    "End-to-end previewDeck latency.",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 4, 8},
	}, []string{"source", "outcome"})

	providerFetchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "vedh_deck_provider_fetch_duration_seconds",
		Help: "Outbound provider fetch latency. Boundaries at 3s and 8s align with the locked connect/total timeouts.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 12},
	}, []string{"provider", "outcome"})

	productEventsDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vedh_product_events_dropped_total",
		Help: "Product events dropped before persistence. REQUIRED by D-17 to distinguish a real funnel dip from a measurement gap.",
	}, []string{"reason"}) // bounded enum; never the offending event name or key
)
```

`promauto.New*` registers into `prometheus.DefaultRegisterer`, which is what `promhttp.Handler()` serves — so no change to `Serve()` is needed. The collectors appear at `/prometheus` automatically, still behind `withMetricsAuth` + `METRICS_ENABLED` + `METRICS_TOKEN` `[VERIFIED: server/graphql.go:37-43 Conf struct — MetricsEnabled `envconfig:"METRICS_ENABLED"`, MetricsToken `envconfig:"METRICS_TOKEN"`]`.

**Completeness correction (2026-08-04, added during planning).** The collector list above and in the
example is **incomplete against ROADMAP Phase 1 success criterion 4**, which requires
"guest-session, import, create/join, and board-activation counters and latency histograms". This
pattern covers only the import and product-event families plus the provider fetch; it names no
guest-session, create/join, or board-activation collector. Plan `01-01` therefore declares all four
families — `vedh_guest_session_{total,duration_seconds}`, `vedh_game_create_{total,duration_seconds}`,
`vedh_game_join_{total,duration_seconds}`, `vedh_board_activation_{total,duration_seconds}` — with
label sets drawn from the same allowlist (`outcome`, `role`), so a Phase 2 or 3 emit site adds an
observation rather than inventing a name. Note that a vector with no observations exports no child
series, so declaration alone does not close criterion 4; it closes when ACT-005, ACT-006, ACT-008 and
ACT-009 emit.

---

### Pattern 7: Frontend product-event service

**What:** A single module under `app/src/services/` exporting a session ID and a fire-and-forget `track()`. No component, no store, no UI.

**When to use:** ACT-001's `app/src/services/productEvents.ts`. `app/src/services/` currently contains only `apollo.ts`, `commanderPartner.ts`, and `scryfall.ts` `[VERIFIED: ls app/src/services/]`, so this follows an established, tiny convention: plain module, named exports, no Pinia.

**Storage key.** Follow the existing namespace: `app/src/stores/auth.ts` uses `const STORAGE_KEY = 'edhgo/auth'` `[VERIFIED: app/src/stores/auth.ts:13 — "const STORAGE_KEY = 'edhgo/auth';"]`. Use `'edhgo/session-id'`.

**ID generation gotcha.** `crypto.randomUUID()` is only available in a secure context (HTTPS or `localhost`). A LAN-IP dev server over plain HTTP — a normal way to test on a tablet, which OPEN-3 makes plausible — leaves it `undefined`. Fall back to `crypto.getRandomValues` before falling back to `Math.random`.

**Example:**

```ts
// app/src/services/productEvents.ts
const STORAGE_KEY = 'edhgo/session-id'; // matches the 'edhgo/auth' namespace in stores/auth.ts

function newSessionID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();               // secure contexts only
  }
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    const b = new Uint8Array(16);
    crypto.getRandomValues(b);
    return Array.from(b, (x) => x.toString(16).padStart(2, '0')).join('');
  }
  return `${Date.now().toString(16)}${Math.random().toString(16).slice(2)}`;
}

export function getSessionID(): string {
  try {
    const existing = localStorage.getItem(STORAGE_KEY);
    if (existing) return existing;            // D-18: stable across tabs, reloads, restarts
    const fresh = newSessionID();
    localStorage.setItem(STORAGE_KEY, fresh);
    return fresh;
  } catch {
    return newSessionID();                    // private mode / storage disabled: degrade, never throw
  }
}

// D-19: one mutation per event, fire-and-forget. The caller never awaits and never sees an error.
export function track(name: string, metadata: Record<string, string | number> = {}): void {
  void apolloClient
    .mutate({ mutation: TRACK_PRODUCT_EVENT, variables: { input: { Name: name, SessionID: getSessionID(), Metadata: metadata } } })
    .catch((err) => console.debug('[productEvents] drop', name, err));
}
```

**Attribution allowlisting is a client-side *and* server-side concern.** The client should only ever read allowlisted UTM/referrer keys out of `location.search` — but the server allowlist (Pattern 5) is what actually enforces it, because the client is not a trust boundary. Both are required; only the server one is a control.

---

### Pattern 8: Two-layer SSRF-hardened HTTP client

**What:** Two independent hooks with disjoint responsibilities. `CheckRedirect` owns scheme and hostname policy; `Dialer.ControlContext` owns IP and port policy. Neither is sufficient alone, and the split is what makes each testable.

**When to use:** ACT-003, **only after the D-14 `checkpoint:decision`**. Do not write adapter code before the checkpoint. The safe-client itself, however, is provider-agnostic and can be specified now.

**Why the IP check must live in `ControlContext`, not in a pre-flight resolve.** Resolving the hostname, validating the result, and then calling `client.Get()` is TOCTOU-vulnerable: an attacker-controlled DNS server with a short TTL returns a public IP for the validation lookup and `169.254.169.254` for the connection lookup. The `Control`/`ControlContext` hook runs *after* resolution and *before* `connect(2)`, receiving the literal resolved address — there is no window. `[CITED: agwa.name/blog/post/preventing_server_side_request_forgery_in_golang — "It's insufficient to do the DNS lookup yourself and block a URL if the hostname resolves to an unsafe address; an attacker could set up a special DNS server that returns a safe address the first time it's queried, and the target address the second time when your application actually connects to the URL."]` The same design is used by `doyensec/safeurl` and `code.dny.dev/ssrf`.

**The IP predicate every naive implementation gets wrong.** `netip.Addr.IsPrivate()` covers only RFC 1918 and `fc00::/7` `[VERIFIED: $(go env GOROOT)/src/net/netip/netip.go:637-661 — "reports whether ip is in 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, or fc00::/7"]`, and `IsGlobalUnicast()` is worse than useless as a gate here because the stdlib comment says so outright: `[VERIFIED: $(go env GOROOT)/src/net/netip/netip.go:625-626 — "Match package net's IsGlobalUnicast logic. Notably private IPv4 addresses and ULA IPv6 addresses are still considered \"global unicast\"."]`. Neither catches `169.254.169.254`'s cloud-metadata neighbours in `100.64.0.0/10`, `192.0.0.0/24`, or `198.18.0.0/15`. An explicit prefix deny-list is required; the IANA-derived set used by `code.dny.dev/ssrf` is a good transcription target `[CITED: pkg.go.dev/code.dny.dev/ssrf — IPv4: 0.0.0.0/8, 10.0.0.0/8, 100.64.0.0/10, 127.0.0.0/8, 169.254.0.0/16, 172.16.0.0/12, 192.0.0.0/24, 192.0.2.0/24, 192.31.196.0/24, 192.52.193.0/24, 192.88.99.0/24, 192.168.0.0/16, 192.175.48.0/24, 198.18.0.0/15, 198.51.100.0/24, 203.0.113.0/24, 224.0.0.0/4, 240.0.0.0/4; IPv6: everything outside 2000::/3 plus 2001::/23, 2001:db8::/32, 2002::/16, 2620:4f:8000::/48, 3fff::/20]`. One thing the stdlib does get right: the `netip` predicates call `Unmap()` on `Is4In6()` addresses, so `::ffff:169.254.169.254` is caught by an IPv4 rule.

**Example:**

```go
// server/deck_providers.go
var deniedV4 = []netip.Prefix{ /* the IANA-derived set above, netip.MustParsePrefix(...) */ }
var deniedV6 = []netip.Prefix{ /* ... */ }

func safeControl(_ context.Context, network, address string, _ syscall.RawConn) error {
	if network != "tcp4" && network != "tcp6" {
		return fmt.Errorf("blocked network %q", network)      // fail closed
	}
	host, port, err := net.SplitHostPort(address)             // handles [v6]:port bracket form
	if err != nil {
		return fmt.Errorf("unparseable dial address: %w", err)
	}
	if port != "443" {
		return fmt.Errorf("blocked port %q", port)            // HTTPS only, at the socket layer too
	}
	addr, err := netip.ParseAddr(host)                        // already resolved: a literal, never a name
	if err != nil {
		return fmt.Errorf("unparseable dial IP: %w", err)
	}
	if addr.Is4In6() {
		addr = addr.Unmap()                                   // ::ffff:169.254.169.254 → 169.254.169.254
	}
	if addr.IsLoopback() || addr.IsUnspecified() || addr.IsMulticast() || addr.IsLinkLocalUnicast() {
		return fmt.Errorf("blocked address %s", addr)
	}
	for _, p := range denied(addr) {
		if p.Contains(addr) {
			return fmt.Errorf("blocked address %s (in %s)", addr, p)
		}
	}
	return nil
}

func newSafeClient(allowedHosts map[string]struct{}) *http.Client {
	dialer := &net.Dialer{
		Timeout:        3 * time.Second,   // locked: 3s connect
		ControlContext: safeControl,
	}
	return &http.Client{
		Timeout: 8 * time.Second,          // locked: 8s total — covers connect, ALL redirects, and body read
		Transport: &http.Transport{
			Proxy:               nil,      // CRITICAL: ProxyFromEnvironment would make the dialer
			                               // validate the PROXY's IP, not the target's. See Pitfall 9.
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: 3 * time.Second,
			DisableKeepAlives:   true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("too many redirects")       // bounded
			}
			if req.URL.Scheme != "https" {
				return fmt.Errorf("blocked redirect scheme %q", req.URL.Scheme)
			}
			if _, ok := allowedHosts[strings.ToLower(req.URL.Hostname())]; !ok {
				return fmt.Errorf("blocked redirect host %q", req.URL.Hostname())
			}
			return nil                                        // ControlContext independently re-checks the IP
		},
	}
}

// Body cap. Content-Length is attacker-controlled; only a counted read is a control.
const maxBody = 1 << 20 // 1 MiB
body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
if err != nil {
	return nil, err
}
if len(body) > maxBody {
	return nil, errors.New("provider response exceeded 1 MiB cap")
}
```

`http.Client.Timeout` is the right home for the 8 s total budget because it "includes connection time, any redirects, and reading the response body" `[VERIFIED: go doc net/http.Client — "Timeout specifies a time limit for requests made by this Client. The timeout includes connection time, any redirects, and reading the response body."]`. `CheckRedirect` is called "before it follows an HTTP redirect" and "if CheckRedirect returns an error, the Client's Get method returns both the previous Response (with its Body closed) and CheckRedirect's error" `[VERIFIED: go doc net/http.Client]` — i.e. returning an error is a genuine hard stop, not a soft signal.

### Anti-Patterns to Avoid

- **Auto-applying a suggestion.** D-02 forbids it, and PROJECT.md states fuzzy matching "may never silently rewrite a user's deck." The API shape should make this structurally impossible: `previewDeck` returns candidates on the *unresolved* entry; there is no field where a rewritten name could live.
- **Letting source detection influence parsing.** The D-23 enum is telemetry. If `moxfield` detection changed how lines are read, a misdetection would silently corrupt a deck and the funnel comparison it exists to enable would be measuring its own bug. Parse identically; label separately.
- **Parsing the deck twice.** Explicitly forbidden by constraint. `previewDeck` and the create/join path must consume one `ParsedDeck`. The seam is narrow — `createLibraryFromDecklist` is called from exactly two places, `server/games.go:552` and `server/games.go:681` `[VERIFIED: grep -rn createLibraryFromDecklist server/]`.
- **Returning a GraphQL error from `trackProductEvent`.** D-22 and D-19 together: nothing awaits the result, so an error is unobservable, while a counter is on a dashboard someone reads.
- **Using the *offending* value as a Prometheus label.** `productEventsDropped.WithLabelValues(e.Name)` looks helpful and is a client-controlled cardinality bomb. Log the value; label the reason.
- **`DROP EXTENSION pg_trgm` in a down migration.** Shared resource. Drop the index and the projection table only.
- **Adding an FK from `product_events` to `users` or `games`.** Breaks D-17 and collides with ACT-005's guest cleanup. See Pattern 4.
- **Trusting `Content-Length` for the 1 MiB cap.** It is a header the provider controls. Only a counted read is a control.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Top-N nearest card names | An in-Go scan over the full name set with Levenshtein | `pg_trgm` GiST `<->` KNN over a distinct-name projection (Pattern 2) | 32,643 names × up to 100 needles = 3.26M distance computations per preview, plus a cache-warming and MTGJSON-refresh-invalidation problem. The index already solves ranking, deduplication of printings, and the D-03 below-cutoff case. |
| Edit distance itself | A DIY dynamic-programming matrix | `agnivade/levenshtein.ComputeDistance` (already in the build graph) — and only as a tie-break over pg_trgm's ≤3 results | Zero new dependency. Hand-rolled Levenshtein gets rune-vs-byte handling wrong on non-ASCII card names (`Æther`, `Márton Stromgald`, `Jötun Grunt`). |
| Private/reserved IP classification | `strings.HasPrefix(ip, "10.")` or `IsPrivate()` alone | An explicit `[]netip.Prefix` deny-list checked with `Prefix.Contains` (Pattern 8) | The stdlib helpers are documented to be narrower than "not public" — `IsGlobalUnicast()` returns true for RFC1918. String prefixes miss IPv6, `Is4In6` mappings, and the entire IANA special-purpose registry. |
| DNS-rebinding protection | Resolve → validate → dial | `net.Dialer.ControlContext` | Resolve-then-dial has an exploitable TOCTOU window by construction. |
| Redirect hop validation | Following redirects manually with a loop | `http.Client.CheckRedirect` | Hand-rolled redirect loops routinely mishandle relative `Location` resolution, cross-scheme downgrades, and cookie/auth-header propagation. |
| HTTP response size cap | Reading `Content-Length` and deciding | `io.LimitReader(body, cap+1)` and comparing the read length | `Content-Length` is attacker-supplied and may be absent under chunked encoding. |
| Prometheus registration plumbing | A bespoke registry, a `sync.Once`, an init-order dance | `promauto.NewCounterVec` / `NewHistogramVec`; `promauto.With(reg)` in tests | Already available at the pinned v1.11.0; `promauto` panics loudly on duplicate registration, which is the correct failure mode for a programming error. |
| Metric assertion in tests | Scraping `/prometheus` and regexing text | `prometheus/testutil` (`ToFloat64`, `CollectAndCount`, `GatherAndCompare`) + `Registry.Gather()` | Bundled with the pinned version; `Gather()` returns structured `[]*dto.MetricFamily` so the label-allowlist rule can be asserted on names, not on scraped text. |
| Decklist grammar | A regex per format, or a parser-combinator dependency | A hand-rolled ordered-rule line scanner (Pattern 1) | No maintained Go library exists; the reference Python implementation does not even handle `1x`. The rules are few enough that a scanner is more explainable than a regex, which is what PROJECT.md requires. |
| Deck-size and commander rules | A fresh implementation inside the new service | Move `server/games.go:802-878` verbatim | Constraint: "existing card lookup, selected-commander removal, and maximum deck-size rules preserved." Rewriting invites off-by-one drift in the `100 - commanders` budget. |

**Key insight:** Almost every "build it" temptation in this phase is a place where PostgreSQL or the Go standard library already owns the hard part, and where a hand-rolled version fails on exactly the inputs an adversary or an unlucky user supplies — a non-ASCII card name, a rebinding DNS record, a chunked response with no `Content-Length`. The one thing that *should* be hand-rolled, the decklist grammar, is the one thing with no credible library and an explicit project requirement to be explainable.

## Runtime State Inventory

> ACT-002 is a refactor (extracting `createLibraryFromDecklist`) and ACT-001 adds a migration, so this inventory applies. It is deliberately narrower than a rename phase's.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | **`games.payload` JSONB rows written before this phase.** D-07 adds `setCode`, `collectorNumber`, `category`, `sourceFormat` to the Go `Card` struct. Existing payloads lack these keys; `encoding/json` will leave them as zero values on unmarshal — no migration needed, but any code that assumes the fields are populated will see `""` for historical games. Separately, **D-06 changes what gets written**: previously-created games contain bare `&Card{Name: …}` placeholders for unresolved cards (`server/games.go:904-908`); those rows are not retroactively cleaned. | Code edit only. Make the four new `Card` fields `omitempty`-tolerant and treat `""` as "unknown printing", never as an error. **Do not** write a data migration over `games.payload` — PROJECT.md forbids replacing the board-state model, and historical games are not worth rewriting. |
| Live service config | **None — verified.** `grep -rn "rate.limit\|RateLimit\|limiter\|Limiter" server/` returns nothing outside generated files, so there is no existing limiter config to reconcile. The only externally-configured surface is the Prometheus/Grafana stack under `monitoring/`, which reads metrics by name; new collector names are additive and break no existing panel. ACT-012 (Phase 4) is where Grafana panels get added. | None this phase. |
| OS-registered state | **None — verified.** Deployment is Dokku (`Makefile` targets `deploy-ui`, `deploy-server`, `docker-server`, `docker-ui`); there are no OS-level task registrations, cron entries, or named process managers in the repo. | None. |
| Secrets / env vars | **`METRICS_ENABLED` and `METRICS_TOKEN` already exist and are unchanged** `[VERIFIED: server/graphql.go:41-42]`. **D-16 introduces a new one**: if the D-14 checkpoint selects a provider requiring an app-level credential, a new `envconfig` field lands on `Conf` and a new secret must be set in the Dokku deploy path. ACT-003 also needs a kill-switch variable (e.g. `DECK_PROVIDER_ENABLED`) per the feature-flag constraint. | New env vars, set at the D-14 checkpoint — not before. The kill switch is required in **both** branches of the checkpoint: even a no-go build should ship the flag defaulting off, so Phase 5 has something to point the runbook at. |
| Build artifacts / installed packages | **`go.sum` and `go.mod` change** if `agnivade/levenshtein` is promoted to direct or `golang.org/x/time` is added. **`server/generated.go` and `app/src/types/generated.ts` are regenerated** by `make generate` (= `go run github.com/99designs/gqlgen`) after the schema edits; the constraint forbids hand-editing them. **`persistence/migrations/atlas.sum` is already stale** and should be left alone (see Pattern 2). | Run `make generate` as an explicit task after every `schema.graphql` change; commit the regenerated files. Do not touch `atlas.sum`. |

## Common Pitfalls

### Pitfall 1: The `cards` table has no name index — every import is a sequential scan

**What goes wrong:** `previewDeck` and `createGame` feel fine on the 30-row test fixtures and get slow with real data.
**Why it happens:** `schema.hcl`'s `cards` table declares exactly one `primary_key` (on `id`) and one index, `cards_uuid_key` (unique, on `uuid`) `[VERIFIED: schema.hcl — awk over the cards table yields only "primary_key {" at line 143 and 'index "cards_uuid_key" { unique = true  columns = [column.uuid] }' at 146-149]`. There is no index on `name`, `facename`, `setcode`, or `number`, and no migration in either directory creates one `[VERIFIED: grep -rn "INDEX" persistence/migrations/*.sql persistence/migrations_test/*.sql → only idx_games_id]`. So `WHERE name = ANY($1) OR facename = ANY($1)` at `server/cards.go:163-166` scans ~97k rows on every deck import.
**How to avoid:** Ship the `cards (lower(name))`, `cards (lower(facename))`, and `cards (lower(name), setcode, number)` indexes in the same migration as the trigram work. Note the lookup lowercases in Go (`key := strings.ToLower(n)` at `server/cards.go:148`) but compares case-sensitively in SQL — an expression index on `lower(...)` only helps if the query also uses `lower(...)`, so the query has to change too.
**Warning signs:** `EXPLAIN` showing `Seq Scan on cards` for the batch lookup; import latency scaling with the size of the card table rather than the size of the deck.

### Pitfall 2: `//` is both a comment prefix and a card-name separator

**What goes wrong:** `1,Bala Ged Recovery // Bala Ged Sanctuary` silently becomes `Bala Ged Recovery`, or the whole line vanishes as a comment.
**Why it happens:** MTGO/Arena exports use `//` for comments; MTG uses `//` between the faces of a double-faced card. This repo's own fixture contains three unquoted instances `[VERIFIED: test/decklists/jarad.csv — "1,Bala Ged Recovery // Bala Ged Sanctuary", "1,Boggart Trawler // Boggart Bog", "1,Bramble Familiar // Fetch Quest"]` and one quoted one (`1,"Agadeem's Awakening // Agadeem, the Undercrypt"`).
**How to avoid:** Anchor the comment rule at line start only (Pattern 1, rule 2). `lheyberger/mtg-parser` uses `StringStart()` for exactly this reason.
**Warning signs:** A deck importing one card short with no warning; a card named with a trailing space.

### Pitfall 3: `1,Atraxa, Praetors' Voice` truncates silently, and it is not an error path

**What goes wrong:** The player gets a card literally named "Atraxa" and is told nothing.
**Why it happens:** `csv.Reader` splits on every comma; the code reads `record[1]` and discards `record[2:]` `[VERIFIED: server/games.go:830]`. There is no `len(record) > 2` check.
**How to avoid:** Rule 3 of Pattern 1 consumes *one* separator after the quantity and treats the entire remainder as the name. Make this exact string a required test case — CONTEXT.md calls it out, and it is the single best regression test for D-11's no-silent-drop guarantee.
**Warning signs:** Any test corpus that only uses single-word card names will never catch this.

### Pitfall 4: `1 Sol Ring` fails outright today, and so does `1x Sol Ring`

**What goes wrong:** The most common decklist format in the world returns "invalid decklist row".
**Why it happens:** `csv.Reader` yields one field for a space-separated line, tripping `if len(record) < 2 { return nil, fmt.Errorf("invalid decklist row: expected quantity and card name") }` `[VERIFIED: server/games.go:827-829]`. For `1x Sol Ring`, `strconv.ParseInt("1x Sol Ring", 10, 64)` fails `[VERIFIED: server/games.go:831-834]`.
**How to avoid:** These are the first two rows of the fixture table (§Validation Architecture).
**Warning signs:** None at runtime — the whole deck fails, loudly. This one is at least honest, unlike Pitfall 3.

### Pitfall 5: D-05 and D-06 interact with the `100 - commanders` cap in a way that is easy to get backwards

**What goes wrong:** A 101-card paste with 2 unresolved cards is rejected as "too large" even though only 99 cards will actually be created.
**Why it happens:** Today's flow counts every parsed entry toward `libraryCount` *before* lookup, then compares against `maxLibraryCards` `[VERIFIED: server/games.go:862-878]`. Lookup happens afterwards, at line 882. D-05 requires the count to exclude unresolved entries — which means **resolution must move before the size check**, inverting the current order.
**How to avoid:** Sequence explicitly: parse → resolve → drop unresolved from the countable set (D-05) and from the library (D-06) → remove commanders → compare against `100 - commandersSpecified`. Keep the commander-budget logic at lines 802-851 verbatim; only its position relative to the count changes.
**Warning signs:** A test that only uses fully-resolvable decks will not distinguish the two orderings.

### Pitfall 6: `previewDeck` and `createGame` disagreeing about the same paste

**What goes wrong:** The preview says 99 cards, the board has 97. The player loses trust in exactly the screen this milestone is built around.
**Why it happens:** Two parse sites drift. D-06 makes this failure mode visible where it previously was not.
**How to avoid:** The constraint is explicit — "preview and final library creation consume the same normalized result." Make `createLibraryFromDecklist` take a `ParsedDeck`, not a `string`. If it still accepts a raw string anywhere, a second parse is one refactor away.
**Warning signs:** Any function signature in the new service that takes `decklist string` and is reachable from `createGame`.

### Pitfall 7: Client-controlled values reaching Prometheus labels through the *error* path

**What goes wrong:** An attacker posts 10,000 distinct bogus event names; each becomes a label value; Prometheus memory explodes.
**Why it happens:** The natural, helpful-looking implementation is `productEventsDropped.WithLabelValues(e.Name).Inc()` — the rejection counter is the one place a *rejected*, unvalidated value is still in scope. The constraint's wording bans "username, session ID, user ID, or game ID" labels and it is easy to read that as an exhaustive list rather than as examples.
**How to avoid:** Label the bounded reason (`unknown_event`, `unknown_key`, `oversized_value`, `write_error`), never the value. Put the value in the structured log, where cardinality is not a resource.
**Warning signs:** Any `WithLabelValues` whose argument is not a compile-time constant or a value of the D-24 enum type.

### Pitfall 8: `http.Transport.Proxy: http.ProxyFromEnvironment` silently disables the entire SSRF control

**What goes wrong:** Every control passes its tests and the client can still reach `169.254.169.254`.
**Why it happens:** With a proxy configured, the dialer connects to the *proxy*, so `ControlContext` validates the proxy's IP and port. The real destination travels in a `CONNECT` line the hook never sees. The widely-copied snippet in the canonical Agwa article includes `Proxy: http.ProxyFromEnvironment` `[CITED: agwa.name/blog/post/preventing_server_side_request_forgery_in_golang]`, which makes this an easy thing to inherit.
**How to avoid:** `Proxy: nil`, explicitly, with a comment saying why. Assert it in a test.
**Warning signs:** `HTTP_PROXY`/`HTTPS_PROXY` set in the Dokku environment for any reason.

### Pitfall 9: Validating the URL's hostname but not each redirect's

**What goes wrong:** `https://allowed-host.example/deck/1` returns `302 Location: https://attacker.example/` and the allowlist is bypassed.
**Why it happens:** Pre-flight validation of the user-supplied URL feels like the check, and Go follows redirects by default (up to 10). The constraint is explicit that redirects must "remain on the allowlist," but the default `CheckRedirect` is `nil`.
**How to avoid:** Set `CheckRedirect` and re-run the full scheme + hostname check against `req.URL` on every hop, bounding `len(via)`. `ControlContext` re-validates the IP per hop independently, so the two layers cover different bypasses.
**Warning signs:** A `CheckRedirect` that only checks `len(via)`.

### Pitfall 10: The 8-second total budget silently not covering the body read

**What goes wrong:** A provider trickles a slow-loris body and the request hangs far past 8 seconds.
**Why it happens:** Using `context.WithTimeout` on the request but not `http.Client.Timeout`, or vice versa without understanding which covers what.
**How to avoid:** `http.Client.Timeout` explicitly "includes connection time, any redirects, and reading the response body" `[VERIFIED: go doc net/http.Client]`. Set it to 8 s and keep `net.Dialer.Timeout` at 3 s for the connect leg.
**Warning signs:** A test that only exercises a fast handler will never notice.

### Pitfall 11: PostgreSQL 14 has no `UNIQUE NULLS NOT DISTINCT`, so the dedup index silently does nothing

**What goes wrong:** `INSERT ... ON CONFLICT DO NOTHING` never conflicts, duplicate `game_created` rows accumulate, and the funnel numerator inflates — the exact failure the constraint "duplicate client retries must not duplicate authoritative conversion events" exists to prevent.
**Why it happens:** In PG14, two rows with `game_id IS NULL` are distinct for uniqueness purposes. `NULLS NOT DISTINCT` arrived in PG15 `[VERIFIED: postgresql.org/docs/15/release-15.html §E.19.3.1.2]`, and dev + CI both pin `postgres:14`.
**How to avoid:** `COALESCE(game_id, '')` / `COALESCE(user_id, '')` expressions in the unique index (Pattern 4).
**Warning signs:** A dedup test that always supplies a non-null `game_id` will pass regardless.

### Pitfall 12: Server tests that cannot run anywhere they are needed

**What goes wrong:** The parser's acceptance criteria are covered by tests that never execute in CI and fail on a fresh developer machine.
**Why it happens:** `.github/workflows/test.yml` runs only `make test-unit` = `go test ./pkg/...`; `./server/...` is in no workflow. And `server/TestMain` unconditionally requires a live Postgres *and* `../All Printings.json`, a file absent from the repo `[VERIFIED: server/main_test.go:18-38; ls "All Printings.json" → not found]`. Reproduced this session: `go test ./server/... -run TestNothingZZZ` → `test setup failed: dial tcp [::1]:5432: connect: connection refused`.
**How to avoid:** Put the deterministic grammar in `pkg/deckimport/`. See §Validation Architecture Wave 0.
**Warning signs:** A plan whose parser tests all live in `server/`.

## Code Examples

### Batched suggestion lookup, Go side

```go
// server/deck_import.go
const (
	maxSuggestionNeedles = 25                     // Pattern 3, bound #3
	suggestionBudget     = 750 * time.Millisecond // Pattern 3, bound #4
	suggestionsPerEntry  = 3                      // D-02
)

// suggestFor returns up to 3 ranked candidates per needle. It NEVER returns an
// error to the caller: D-01 says a suggestion failure must not block the preview.
func (s *graphQLServer) suggestFor(ctx context.Context, needles []string) (map[string][]Suggestion, []string) {
	var warnings []string
	if len(needles) == 0 {
		return nil, nil
	}
	if len(needles) > maxSuggestionNeedles {
		warnings = append(warnings, fmt.Sprintf(
			"Showing suggestions for the first %d of %d unresolved cards.",
			maxSuggestionNeedles, len(needles)))
		needles = needles[:maxSuggestionNeedles]
		deckSuggestionTruncated.Inc()
	}

	qctx, cancel := context.WithTimeout(ctx, suggestionBudget)
	defer cancel()

	rows, err := s.db.QueryContext(qctx, `
		SELECT q.needle, c.display, 1 - (c.name_lower <-> q.needle) AS score
		FROM   unnest($1::text[]) WITH ORDINALITY AS q(needle, ord)
		CROSS JOIN LATERAL (
		    SELECT cn.display, cn.name_lower
		    FROM   card_names cn
		    ORDER  BY cn.name_lower <-> q.needle
		    LIMIT  $2
		) c
		ORDER BY q.ord, score DESC;`,
		pq.Array(needles), suggestionsPerEntry)
	if err != nil {
		s.loggerFor(ctx).Warn("suggestion lookup failed", "err", err, "needles", len(needles))
		return nil, append(warnings, "Suggestions are temporarily unavailable.")
	}
	defer rows.Close()
	// ... scan into map[needle][]Suggestion{Name, Score} ...
	return out, warnings
}
```

`pq.Array` for the `text[]` parameter follows the existing batch-lookup convention at `server/cards.go:166` `[VERIFIED: server/cards.go:166 — "pq.Array(trimmed),"]`.

### The never-fail event write, mirroring the existing `logEvent`

```go
// server/gamelog_helpers.go:25-33 is already the codebase's canonical
// "log and return" pattern; product events copy it and add a counter:
//
//   func (s *graphQLServer) logEvent(ctx context.Context, event Event) {
//       if s == nil || s.db == nil { return }
//       g := &pgLogger{db: s.db}
//       if err := g.Add(ctx, event); err != nil {
//           s.loggerFor(ctx).Warn("failed to write gamelog event", "err", err, ...)
//       }
//   }
```
`[VERIFIED: server/gamelog_helpers.go:25-33]`

### Example funnel query (ACT-001 deliverable, exercised by ACT-012)

```sql
-- Host activation rate: quick_start_viewed → board_ready, by day and source.
-- Uses product_events_session_time_idx for the self-join.
WITH viewed AS (
    SELECT session_id,
           min(occurred_at) AS viewed_at,
           min(source)      AS source
    FROM   product_events
    WHERE  event_name = 'quick_start_viewed'
    GROUP  BY session_id
),
ready AS (
    SELECT session_id, min(occurred_at) AS ready_at
    FROM   product_events
    WHERE  event_name = 'board_ready' AND role = 'host'
    GROUP  BY session_id
)
SELECT date_trunc('day', v.viewed_at AT TIME ZONE 'UTC') AS day,
       coalesce(v.source, 'unknown')                     AS source,
       count(*)                                          AS viewers,
       count(r.session_id)                               AS activated,
       round(100.0 * count(r.session_id) / nullif(count(*), 0), 1) AS activation_pct,
       percentile_cont(0.5)  WITHIN GROUP (ORDER BY extract(epoch FROM r.ready_at - v.viewed_at)) AS p50_seconds,
       percentile_cont(0.9)  WITHIN GROUP (ORDER BY extract(epoch FROM r.ready_at - v.viewed_at)) AS p90_seconds
FROM      viewed v
LEFT JOIN ready  r USING (session_id)
GROUP BY 1, 2
ORDER BY 1 DESC, 2;
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `set_limit(real)` / `show_limit()` to control the trigram threshold | `SET pg_trgm.similarity_threshold` (a GUC) | PostgreSQL 9.6 | `[CITED: postgresql.org/docs/current/pgtrgm.html — both listed as "Deprecated"]` Use the GUC form, or avoid the threshold entirely by using GiST `<->` ordering (Pattern 2). |
| `CREATE EXTENSION` requires superuser | Trusted extensions installable by a DB owner with `CREATE` privilege | PostgreSQL 13 | `[VERIFIED: postgres/postgres REL_14_STABLE contrib/pg_trgm/pg_trgm.control — "trusted = true"]` The migration can create the extension as the ordinary app user. Do not plan a superuser escalation step. |
| Unique indexes always treat NULLs as distinct | `UNIQUE NULLS NOT DISTINCT` | PostgreSQL **15** | `[VERIFIED: postgresql.org/docs/15/release-15.html]` **Not available here** — dev and CI both pin `postgres:14`. Use `COALESCE` expression indexes (Pitfall 11). |
| Classic Prometheus histograms with hand-picked buckets | Native (sparse) histograms | client_golang v1.14+ / Prometheus 2.40+ | **Not available here** — the pin is v1.11.0 `[VERIFIED: go.mod:22]`. Explicit `Buckets` are required, so bucket choice is a real design decision rather than something the server infers. |
| `net.Dialer.Control` | `net.Dialer.ControlContext` | Go 1.20 | `[VERIFIED: go doc net.Dialer — "If ControlContext is not nil, Control is ignored."]` Prefer `ControlContext` so cancellation and request-scoped values reach the hook. Third-party SSRF packages still use `Control`; do not copy that detail. |
| `net.IP` + `net.IPNet.Contains` | `netip.Addr` + `netip.Prefix.Contains` | Go 1.18 | Value type, no allocation, and the predicates unmap `Is4In6` automatically — which is what closes the `::ffff:169.254.169.254` bypass class. |

**Deprecated/outdated:**
- `encoding/csv` for decklists: structurally cannot satisfy REQ-A2 (Pitfalls 3 and 4).
- `persistence/migrations/atlas.sum`: stale by three migration pairs; evidently unenforced. Do not regenerate as part of this phase.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The **per-event metadata key assignments** in Pattern 5 (all 15 events except `board_ready` and `quick_start_viewed`, which D-21 specifies verbatim) | Architecture Pattern 5 | Medium. D-21 says the PRD already specifies these keys, so the correct values exist and must be transcribed from `docs/product/2026-07-23-deck-to-game-activation-prd.md`. Getting one wrong means an event is silently dropped in production (D-22) with only a counter to show for it. **The plan must include an explicit transcription task reading the PRD, not adoption of this table.** |
| A2 | A cap of **25** distinct unresolved names for eager suggestions | Architecture Pattern 3 | Low. The value is instrumented (`vedh_deck_suggestion_truncated_total`) and tunable without a contract change. Chosen as ~12× headroom over the ≥98%-resolution target's implied ≤2 unresolved per 100-card deck; not measured. |
| A3 | A **750 ms** sub-budget for the batched suggestion query | Architecture Pattern 3 | Low. Degradation is graceful by design (zero suggestions + a warning). Should be validated against a real MTGJSON-sized table before beta; no such table was available this session. |
| A4 | Similarity **cutoff value** for marking a suggestion low-confidence | Architecture Pattern 2 | Low. Explicitly the planner's discretion per CONTEXT.md. Pattern 2 deliberately makes it a presentation concern rather than a query concern, so it is cheap to change. `pg_trgm.similarity_threshold` defaults to 0.3 `[CITED: postgresql.org/docs/current/pgtrgm.html]`, which is a reasonable starting point for the low-confidence line. |
| A5 | GiST outperforms GIN **for this workload at this table size** | Architecture Pattern 2 | Low-medium. The *correctness* argument for GiST (D-03 needs below-cutoff results, which `%` cannot give) is verified from the PostgreSQL docs and is decisive on its own. The *performance* comparison at ~33k rows is not measured here; general guidance is that GIN overtakes GiST above ~100k entries, and the projection table is deliberately well under that. |
| A6 | Paste-shape detection heuristics (uppercase `(SET)` + collector number ⇒ `moxfield`; trailing `[Category]` or backtick category or a `Section (N)` header ⇒ `archidekt`) | Architecture Pattern 1, D-23/D-24 | Low. Explicitly the planner's discretion, `unknown` is an available fallback, and the anti-pattern rule ("detection never influences parsing") makes a misdetection a telemetry inaccuracy rather than a data-corruption bug. The backtick form is `[CITED: proxyfoundry.com]`, not confirmed against a live Archidekt export. |
| A7 | Prometheus histogram bucket boundaries | Architecture Pattern 6 | Low. Changing buckets later resets historical histogram data but breaks nothing structural. The reasoning (boundaries at 3 s and 8 s to match the locked timeouts) is sound independent of the specific list. |
| A8 | The `cards` table contains roughly 97k printings / 33k distinct names in this deployment | Architecture Pattern 2 | Low. Scryfall's paper counts were verified live this session (32,643 / 96,589), and `persistence/import_all_printings_json.go` imports MTGJSON `AllPrintings` cards one row per printing — but the actual deployed row count could not be measured (no local Postgres, no `All Printings.json` in the repo). Order of magnitude is what the design depends on, and that is safe. |

## Open Questions (RESOLVED)

All four are closed by the Phase 1 plan set written 2026-08-04. Each carries its resolution inline;
Q1 is resolved *as deliberately deferred* to a human checkpoint, which is a resolution of the
planning question, not of OPEN-1 itself.

1. **OPEN-1 — Which public deck provider becomes the first supported URL source.**
   - What we know: D-13 requires the spike to start neutral; D-14 puts a `checkpoint:decision` before any adapter code; D-15 makes a stable response shape the only hard gate condition; the exit note permits a documented no-go with paste-only activation, and INFO-1 says a no-go must not block the Phase 5 release gate.
   - What's unclear: everything about the outcome, deliberately.
   - **RESOLVED (planning question only) — deliberately deferred, as recommended.** Plan `01-07`
     implements exactly the recommended shape: task 1 is the timeboxed neutral spike producing
     `docs/research/deck-provider-feasibility.md`, task 2 is the `checkpoint:decision` where the
     human picks Archidekt, Moxfield, or no-go, and task 3 executes only the selected branch. No
     adapter code is pre-written and no candidate is pre-favoured anywhere in the plan set. The
     provider-agnostic secure client and the default-off kill switch are built in plan `01-06`,
     ahead of and independent of the decision, so both branches are equally cheap at the checkpoint.
     OPEN-1 itself remains genuinely open until that checkpoint resolves at execution time.
   - Recommendation: **Do not research or pre-favour a provider.** Plan ACT-003 as: (a) a timeboxed spike task producing `docs/research/deck-provider-feasibility.md`; (b) a `checkpoint:decision`; (c) nothing else pre-written. The spike should measure, for each candidate, exactly six things — request shape and whether an unauthenticated public read path exists; how a *private* deck responds (this determines the "reported as unsupported without requesting credentials" behavior); rate-limit signals in headers or status codes; response-shape stability, which D-15 makes the only hard gate; ToS/operational risk, recorded but non-blocking; and whether a fixture can be captured and maintained, since the standing constraint forbids any test depending on a live provider. The safe HTTP client of Pattern 8 and the `DECK_PROVIDER_ENABLED` kill switch are provider-agnostic and can be built and tested in **either** branch — including the no-go branch, where shipping a default-off flag gives the Phase 5 runbook something concrete to reference.

2. **`previewDeck` rate-limiting scope.**
   - What we know: the SPEC requires IP/session rate limiting on guest creation, deck import, and public invite lookup, with instrumented outcomes. **No rate limiting of any kind exists in the repo today.** `previewDeck` sits on the activation critical path.
   - What's unclear: whether `previewDeck` is "deck import" for the purposes of that constraint.
   - **RESOLVED — in scope, tuned loosely, instrumented on both paths.** Plan `01-06` task 1 limits
     both `previewDeck` and `trackProductEvent` per surface and per client key at 30/minute with a
     burst of 10, counts `vedh_rate_limit_total{surface,outcome}` on the allowed path as well as the
     limited one so the ratio is readable, and puts the bucket registry and its idle-eviction sweep
     in `pkg/ratelimit` — the only test target CI runs — with the instrumented wrapper left in
     `server/`. The `x/time/rate` module is promoted from transitive to direct.
   - Recommendation: treat it as in scope but tune it loosely — a per-session/IP token bucket generous enough that a human correcting cards several times never hits it (e.g. 30/minute), with the *outcome* instrumented (`vedh_rate_limit_total{surface,outcome}`). The instrumentation is the load-bearing part: it makes a too-tight limit visible in Grafana before it shows up as a funnel dip. `golang.org/x/time/rate` is the obvious implementation and is already in `go.sum`.

3. **Whether CI should be extended to run `./server/...`.**
   - What we know: it does not today, and cannot without a Postgres service and a ~500 MB `All Printings.json`.
   - What's unclear: whether this phase should absorb that change.
   - **RESOLVED — no; split the code instead, as recommended.** No plan touches CI. The deterministic
     grammar lives in `pkg/deckimport` (plans `01-01`, `01-02`), the vocabulary and metric
     guarantees in `pkg/telemetry` (`01-01`), and the limiter registry in `pkg/ratelimit` (`01-06`).
     Every task whose verify block runs `go test ./server` carries a `<precondition>` naming both
     Postgres and the MTGJSON snapshot, so an executor without them halts on a stated fact rather
     than on a confusing `TestMain` exit. A committed MTGJSON subset fixture remains out of scope for
     this phase, as recommended.
   - Recommendation: **No — split the code instead.** Put the deterministic grammar in `pkg/deckimport/` so REQ-A2's acceptance criteria gate every PR immediately at zero CI cost. Separately, consider a small, committed MTGJSON subset fixture (a few hundred cards, including comma-containing and double-faced names) so DB-backed tests can eventually run in CI without the full download — but scope that as its own change, not a Phase 1 dependency. Note `persistence/migrations_test/20260131000000_seed_test_cards.up.sql` (32 KB) already establishes the seeded-subset pattern.

4. **Whether `card_names` should be a table, a materialized view, or a trigger-maintained projection.**
   - What we know: the source `cards` table is refreshed by an explicit, manual operation (`make import-allprintings` / `make import-csv`), not continuously.
   - What's unclear: nothing blocking — but a stale projection means new sets are unsuggestable.
   - **RESOLVED — a plain table plus a refresh hook on the import path, as recommended.** Plan `01-04`
     task 1 creates `card_names` as a plain table populated by the migration and appends the same
     projection insert to the end of `persistence/import_all_printings_json.go`'s batch loop, logging
     and continuing on failure and treating an absent projection relation as nothing to refresh. No
     materialized view and no trigger.
   - Recommendation: a plain table populated by the migration, plus a documented refresh step appended to the MTGJSON import path (`persistence/import_all_printings_json.go` already has the natural hook at the end of its batch loop). A materialized view with `REFRESH CONCURRENTLY` is the tidier long-term answer but needs a unique index and adds a moving part this phase does not need.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All server work | ✓ | 1.24.0 / toolchain go1.24.2 `[VERIFIED: go.mod:4-6]` | — |
| Go module proxy | Dependency verification | ✓ | `go list -m -versions` succeeded for both candidate modules | — |
| PostgreSQL 14 (local) | Migrations, DB-backed tests | ✗ (not running) | image pinned `postgres:14` `[VERIFIED: dev.docker-compose.yml:39]` | `make persistence` starts it via Docker Compose |
| `psql` CLI | Ad-hoc schema/index inspection | ✗ | — | `docker compose exec postgres psql`, or Go integration tests |
| `All Printings.json` | `go test ./server/...` (`TestMain` → `importAllPrintingsForTests`) | ✗ (absent from repo) | — | **No fallback for DB-backed server tests.** This is why the parser belongs in `pkg/deckimport/`. Download per `Makefile` comment: `https://mtgjson.com/downloads/all-files` |
| `migrate` CLI (golang-migrate) | `make migrate-local` / `migrate-prod` | not checked | v4.14.1 as a Go module `[VERIFIED: go.mod:14]` | The Go module is vendored into the test harness (`persistence.ForceCleanMigrations`), so tests do not need the CLI |
| Node + npm | Frontend event service + vitest | assumed present | vitest ^1.1.9, vite ^5.4.2 `[VERIFIED: app/package.json]` | — |
| Prometheus / Grafana | Verifying collectors render | ✓ (tooling committed) | `make monitoring-up` → `monitoring/docker compose` | `prometheus/testutil` assertions cover correctness without the stack |
| Live Archidekt / Moxfield | ACT-003 spike **only** | n/a | — | **Forbidden in tests by standing constraint.** The spike is manual and timeboxed; all committed tests use `server/testdata/deck_providers/*` fixtures. |

**Missing dependencies with no fallback:**
- `All Printings.json` — blocks `go test ./server/...` entirely on any machine that lacks it, including CI. Mitigated by architecture (pure parser in `pkg/`), not by installation.

**Missing dependencies with fallback:**
- Local PostgreSQL — `make persistence` brings it up.
- `psql` — Docker exec or Go tests.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework (Go) | `go test` + `github.com/stretchr/testify v1.11.1` + `github.com/matryer/is v1.4.0` `[VERIFIED: go.mod:21,23]` |
| Framework (frontend) | `vitest ^1.1.9` with `jsdom ^22.1.0` and `@vue/test-utils ^2.4.6` `[VERIFIED: app/package.json]` |
| Config file (Go) | none — standard `go test`; `server/main_test.go` provides `TestMain` |
| Config file (frontend) | inline `"vitest"` block in `app/package.json`; `environment: jsdom`, `globals: true`, `include: ["__tests__/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}"]` `[VERIFIED: app/package.json]` — **note there is no `vitest.config.*` file** |
| Quick run command | `go test ./pkg/... -race` (DB-free, seconds) |
| Full suite command | `make test-api` (= `go test -v ./server/... -race`, needs Postgres + `All Printings.json`) then `cd app && npm test` |
| **What CI runs today** | `make test-unit` = `go test -v ./pkg/... -race` **only** `[VERIFIED: .github/workflows/test.yml + Makefile:26-27]` — plus a separate Rust smoke workflow that boots the API against a `postgres:14` service |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACT-002 | The six required syntaxes parse to identical `(qty=1, name)` pairs | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ Wave 0 |
| ACT-002 | `1,Atraxa, Praetors' Voice` and `1,"Atraxa, Praetors' Voice"` yield the *same* name (the truncation regression) | unit | `go test ./pkg/deckimport -run TestScanner_CommaNames` | ❌ Wave 0 |
| ACT-002 | `//` inside a name is a face separator; `//` at line start is a comment | unit | `go test ./pkg/deckimport -run TestScanner_DoubleFacedVsComment` | ❌ Wave 0 |
| ACT-002 | D-07 set code / collector number / category extracted and name left clean | unit | `go test ./pkg/deckimport -run TestScanner_PrintingMetadata` | ❌ Wave 0 |
| ACT-002 | D-11 sideboard/maybeboard dropped **with a count**; D-12 commander header preselects | unit | `go test ./pkg/deckimport -run TestScanner_Sections` | ❌ Wave 0 |
| ACT-002 | **No nonblank row disappears** — the accounting invariant, asserted over every golden fixture | unit (property) | `go test ./pkg/deckimport -run TestScanner_NoRowDisappears` | ❌ Wave 0 |
| ACT-002 | D-23 source detection never changes the parsed entries | unit (differential) | `go test ./pkg/deckimport -run TestSourceDetection_DoesNotAffectParse` | ❌ Wave 0 |
| ACT-002 | D-05/D-06 — unresolved excluded from count and library; `100 - commanders` cap preserved | integration | `go test ./server -run TestDeckImport_UnresolvedAccounting` | ❌ Wave 0 |
| ACT-002 | D-02/D-03 — ≤3 ranked candidates; nearest returned even below cutoff | integration | `go test ./server -run TestDeckImport_Suggestions` | ❌ Wave 0 |
| ACT-002 | Preview and create/join consume one `ParsedDeck` (no second parse) | integration | `go test ./server -run TestDeckImport_SingleParse` | ❌ Wave 0 |
| ACT-001 | Unknown event names and non-allowlisted keys are dropped, logged, counted — never errored | unit | `go test ./server -run TestProductEvents_Allowlist` | ❌ Wave 0 |
| ACT-001 | **No allowlisted key is named for a forbidden concept** (structural privacy proof) | unit (DB-free) | `go test ./server -run TestEventVocabulary_NoForbiddenKeys` | ❌ Wave 0 |
| ACT-001 | **No metric carries a high-cardinality label** (structural cardinality proof) | unit (DB-free) | `go test ./server -run TestMetrics_LabelAllowlist` | ❌ Wave 0 |
| ACT-001 | A write failure does not fail the caller (D-17) | unit | `go test ./server -run TestProductEvents_WriteFailureIsNonFatal` | ❌ Wave 0 |
| ACT-001 | Duplicate authoritative events deduplicate (the PG14 `COALESCE` index) | integration | `go test ./server -run TestProductEvents_AuthoritativeDedup` | ❌ Wave 0 |
| ACT-001 | Migration up/down passes for prod **and** test schemas | integration | `go test ./server -run TestMigrations_ProductEvents` | ❌ Wave 0 |
| ACT-001 | Session ID is stable across reloads; attribution is allowlisted | unit (vitest) | `cd app && npx vitest --run __tests__/productEvents.spec.ts` | ❌ Wave 0 |
| ACT-003 | Non-allowlisted host, non-HTTPS scheme, redirect off the allowlist, redirect loop | unit | `go test ./server -run TestSafeClient_HostAndRedirect` | ❌ Wave 0 |
| ACT-003 | Private/reserved/link-local/CGNAT/metadata IPs rejected, including `::ffff:` mapped forms | unit (table) | `go test ./server -run TestSafeControl_DeniedAddresses` | ❌ Wave 0 |
| ACT-003 | DNS rebinding fails closed | unit | `go test ./server -run TestSafeControl_Rebinding` | ❌ Wave 0 |
| ACT-003 | Response over 1 MiB fails closed; a lying `Content-Length` does not help | unit | `go test ./server -run TestSafeClient_BodyCap` | ❌ Wave 0 |
| ACT-003 | Connect and total timeouts fail closed | unit | `go test ./server -run TestSafeClient_Timeouts` | ❌ Wave 0 |
| ACT-003 | Fixture contract test detects a provider response-shape change | unit | `go test ./server -run TestProvider_FixtureContract` | ❌ Wave 0 (post-checkpoint) |
| ACT-003 | Kill switch off ⇒ normalized error + paste fallback; pasted import unaffected | unit | `go test ./server -run TestProvider_KillSwitch` | ❌ Wave 0 |

### How the three hard-to-test guarantees become mechanical

**1. The parser fixture corpus.** A single table-driven test whose rows are `{input, wantQty, wantName, wantSet, wantCollector, wantCategory}`. The non-negotiable minimum, drawn from REQ-A2 and CONTEXT.md's Specific Ideas:

| Input | Expect |
|-------|--------|
| `1 Sol Ring` | qty 1, name `Sol Ring` |
| `1x Sol Ring` | qty 1, name `Sol Ring` |
| `1,Sol Ring` | qty 1, name `Sol Ring` |
| `1, Sol Ring` | qty 1, name `Sol Ring` |
| `1,"Atraxa, Praetors' Voice"` | qty 1, name `Atraxa, Praetors' Voice` |
| `1 Atraxa, Praetors' Voice` | qty 1, name `Atraxa, Praetors' Voice` |
| **`1,Atraxa, Praetors' Voice`** | qty 1, name `Atraxa, Praetors' Voice` — **the truncation regression; today yields `Atraxa`** |
| `1 Sol Ring (C21) 263` | set `C21`, collector `263`, name `Sol Ring` |
| `1x Sol Ring (c21) 263 [Ramp]` | set `c21`, collector `263`, category `Ramp`, name `Sol Ring` |
| `` 1 Sol Ring `Maybeboard` `` | category `Maybeboard`, name `Sol Ring` |
| `1,Bala Ged Recovery // Bala Ged Sanctuary` | name `Bala Ged Recovery // Bala Ged Sanctuary` — **from the repo's own `test/decklists/jarad.csv`** |
| `// Commander` | comment/section, not an entry |
| `Sideboard (12)` | section header; the 12 following rows become a dropped-count warning (D-11) |
| `Commander` | section header; following rows feed `CommanderCandidates` (D-12) |
| `` (empty) `` / `   ` | skipped silently — the only rows allowed to vanish |
| `Sol Ring` (no quantity) | qty 1, name `Sol Ring` |
| `-1 Sol Ring` | warning, not a whole-deck failure |

Plus golden fixtures: whole-file Moxfield-shaped, Archidekt-shaped, and generic-text exports in `pkg/deckimport/testdata/`, each asserted against a checked-in expected `ParsedDeck`, and each additionally run through the accounting invariant.

**2. Proving event-allowlist rejection — and proving the privacy rule structurally.** Two distinct tests. The behavioral one submits an unknown event name, an unknown key, and an oversized value, asserting for each: no error returned to the caller, no row written, and `vedh_product_events_dropped_total{reason=...}` incremented by exactly one (via `testutil.ToFloat64`).

The structural one is the more valuable of the two, because it stays true as the code grows:

```go
// server/product_events_test.go — DB-free, runs anywhere.
func TestEventVocabulary_NoForbiddenKeys(t *testing.T) {
	forbidden := []string{"deck", "list", "url", "uri", "link", "password",
		"passwd", "secret", "token", "jwt", "auth", "clipboard", "ip", "addr", "cards", "hand"}
	for event, spec := range eventVocabulary {
		for key := range spec.Keys {
			for _, bad := range forbidden {
				if strings.Contains(strings.ToLower(key), bad) {
					t.Errorf("event %q allowlists key %q containing forbidden substring %q", event, key, bad)
				}
			}
		}
	}
}

func TestEventVocabulary_IsClosedAtFifteen(t *testing.T) {
	if len(eventVocabulary) != 15 { // D-20
		t.Fatalf("event vocabulary must contain exactly 15 events, got %d", len(eventVocabulary))
	}
}
```

This turns "raw deck text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game state cannot be stored" — a claim that would otherwise be verified only by a reviewer reading a table — into a failing build the moment someone adds `deck_url` to an event in Phase 3. Note the deliberate use of substring matching over exact names: it catches `deck_source_url` as well as `url`.

**3. Proving the no-high-cardinality-label rule mechanically.** Walk the registry rather than inspecting the source:

```go
// server/metrics_test.go — DB-free, runs anywhere.
func TestMetrics_LabelAllowlist(t *testing.T) {
	// The complete set of label names any vEDH metric may ever carry.
	allowed := map[string]bool{
		"source": true, "outcome": true, "role": true,
		"reason": true, "provider": true, "surface": true,
	}
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, fam := range families {
		if !strings.HasPrefix(fam.GetName(), "vedh_") {
			continue // ignore go_* and process_* built-ins
		}
		for _, m := range fam.GetMetric() {
			for _, lp := range m.GetLabel() {
				if !allowed[lp.GetName()] {
					t.Errorf("metric %s carries disallowed label %q (value %q)",
						fam.GetName(), lp.GetName(), lp.GetValue())
				}
			}
		}
	}
}

// Bounded-value check for the one label whose values are an enum (D-24).
func TestMetrics_SourceLabelValuesAreEnumMembers(t *testing.T) {
	valid := map[string]bool{"moxfield": true, "archidekt": true, "plain_text": true, "unknown": true}
	// ... same Gather() walk, assert lp.GetName()=="source" ⇒ valid[lp.GetValue()] ...
}
```

`Registry.Gather()` returns `[]*dto.MetricFamily` `[VERIFIED: $GOMODCACHE/github.com/prometheus/client_golang@v1.11.0/prometheus/registry.go:409]`, so both assertions are structural rather than textual. The test must run **after** the metrics have been observed at least once (a `HistogramVec`/`CounterVec` with no observations reports no child metrics), so the test file should exercise each collector with representative values first — which conveniently also proves the labels are wired.

**4. Testing SSRF controls with zero live-provider dependency.** Every control is testable against loopback and in-process fakes, which is what makes the "no test may depend on a live deck provider" constraint costless here:

- **Allowlist and redirect policy:** an `httptest.Server` (loopback) as the "provider", with its host injected into the allowlist for the positive case. The redirect cases are two `httptest.Server`s where the first `302`s to the second, which is *not* allowlisted — assert the error and assert the second server's handler was never invoked.
- **IP deny-list:** call `safeControl` **directly** as a table test over `("tcp4", "169.254.169.254:443")`, `("tcp4", "127.0.0.1:443")`, `("tcp4", "100.64.0.1:443")`, `("tcp6", "[::ffff:169.254.169.254]:443")`, `("tcp4", "192.0.0.1:443")`, `("tcp4", "198.18.0.1:443")`, `("tcp4", "8.8.8.8:80")` (wrong port), `("udp4", "8.8.8.8:443")` (wrong network), and one public control that must pass. No network is touched at all, so this is fast, hermetic, and exhaustive.
- **DNS rebinding:** a `net.Resolver` with a custom `Dial` pointing at an in-process DNS stub that returns a public address on the first query and `127.0.0.1` on the second. Assert the second connection is refused by the Control hook. If a DNS stub is judged too heavy, the direct `safeControl` table above already covers the property the hook provides; the rebinding test's added value is proving the hook is actually *wired into* the client, which can also be shown by asserting a request to a hostname resolving to loopback fails.
- **1 MiB cap:** an `httptest` handler streaming `1<<20 + 1` bytes, and a second handler advertising `Content-Length: 10` while sending 2 MiB — the second is the one that catches a `Content-Length`-trusting implementation.
- **Timeouts:** an `httptest` handler that sleeps past the budget; assert the elapsed time is bounded and the error is a timeout. Keep the test's own budget short by parameterising the client's timeouts.
- **Fixture contract:** golden provider bodies in `server/testdata/deck_providers/`, parsed by the adapter and asserted against an expected normalized deck. This is the test that "detects provider response-shape changes" — it fails when someone refreshes a fixture and the shape moved.

### Sampling Rate

- **Per task commit:** `go test ./pkg/... -race` (DB-free, seconds — this is also exactly what CI runs)
- **Per wave merge:** `go test ./pkg/... -race` + `go test ./server/... -race` (requires local Postgres + `All Printings.json`) + `cd app && npm test` + `cd app && npm run type-check`
- **Phase gate:** the full suite above green, plus `make generate` producing no uncommitted diff, before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `pkg/deckimport/` package skeleton — `scanner.go`, `sections.go`, `sourcetype.go`, `result.go`. **This is the highest-leverage Wave 0 item**: it is what puts REQ-A2's acceptance criteria inside the only test target CI actually runs.
- [ ] `pkg/deckimport/scanner_test.go` — the required-syntax corpus table (covers ACT-002 / REQ-A2)
- [ ] `pkg/deckimport/testdata/` — Moxfield-shaped, Archidekt-shaped, and generic golden exports plus expected `ParsedDeck` files
- [ ] `server/product_events_test.go` — allowlist behavior, the structural forbidden-key test, the closed-at-15 test, the non-fatal-write test
- [ ] `server/metrics_test.go` — the label-allowlist walk and the enum-value check
- [ ] `server/deck_import_test.go` — DB-backed resolution, suggestions, D-05/D-06 accounting, single-parse proof
- [ ] `server/deck_providers_test.go` + `server/testdata/deck_providers/` — SSRF table tests and fixture contract (created after the D-14 checkpoint; the safe-client tests themselves are provider-agnostic and can land before it)
- [ ] `app/__tests__/productEvents.spec.ts` — session-ID stability across simulated reloads, attribution allowlisting, fire-and-forget swallowing errors
- [ ] Migration test coverage for `product_events` up/down across **both** `persistence/migrations/` and `persistence/migrations_test/`
- [ ] Framework install: **none required** — Go stdlib testing, testify, and vitest are all already present

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no (this phase) | Guest identity and bcrypt claim rules are ACT-005 / Phase 2. `previewDeck` and `trackProductEvent` are unauthenticated by design. |
| V3 Session Management | partial | The D-18 `localStorage` session ID is an **analytics** identifier, not an authentication session. It must never be accepted as an authorization input. Enforced by the constraint "the server attaches authenticated user IDs rather than accepting them from the client" — the client-supplied `sessionID` may populate `product_events.session_id` and nothing else. |
| V4 Access Control | partial | `/prometheus` remains behind `withMetricsAuth` + `METRICS_ENABLED` + `METRICS_TOKEN`, which uses `subtle.ConstantTimeCompare` `[VERIFIED: server/graphql.go — "if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1"]`. This phase must not weaken it; the new collectors inherit it. `gameInvite` and `getGame` authorization are ACT-007 / Phase 3. |
| V5 Input Validation | **yes** | Three surfaces: the decklist grammar (bounded quantity, bounded line count, bounded total input size); the `trackProductEvent` closed vocabulary + per-event key allowlist + `maxMetadataValueLen`; and the provider URL (HTTPS-only + hostname allowlist). All three are allowlists, not denylists. |
| V6 Cryptography | partial | Only the session ID's randomness. Use `crypto.randomUUID()` / `crypto.getRandomValues` in the browser (Pattern 7); never `Math.random()` as the primary source. No new server-side cryptography in this phase. |
| V7 Error Handling & Logging | **yes** | REQ-A7: "Error messages use product language and never expose raw GraphQL, SQL, or provider responses." Provider failures must return a normalized error; the raw body goes to the structured log, never to the client. The same rule applies to unresolved-card messages — a `pq` error must not surface as a warning string. |
| V12 Files & Resources | **yes** | The 1 MiB body cap, the 25-needle suggestion cap, and a maximum accepted decklist input size (recommend 256 KiB — well above any legitimate 100-card list) all bound resource consumption on unauthenticated endpoints. |
| V13 API & Web Service | **yes** | Rate limiting on public surfaces (Open Question 2); `trackProductEvent` returning `Boolean!` and never a detailed error, per D-22. |
| V14 Configuration | **yes** | Two independent server-side kill switches are required by constraint — one for the deck-provider adapter, one for guest creation (Phase 2). Both must default to **off** so a deploy is safe before the flag is deliberately enabled. |

### Known Threat Patterns for Go + PostgreSQL + GraphQL + outbound HTTP

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SSRF to cloud metadata (`169.254.169.254`) via a crafted deck URL | Information Disclosure | `ControlContext` IP deny-list including link-local; HTTPS-only; hostname allowlist (Pattern 8) |
| DNS rebinding / TOCTOU between resolve and connect | Information Disclosure, Elevation | Validate the resolved literal inside `ControlContext`, never pre-flight (Pitfall/Pattern 8) |
| Redirect chain escaping the hostname allowlist | Information Disclosure | `CheckRedirect` re-validating scheme + host per hop, bounded hops (Pitfall 9) |
| Proxy env vars bypassing the dial-layer control | Information Disclosure | `http.Transport.Proxy: nil`, asserted in test (Pitfall 8) |
| Decompression / oversized response DoS | Denial of Service | `io.LimitReader(body, 1 MiB + 1)`; never trust `Content-Length` (Pattern 8) |
| Slow-loris provider response | Denial of Service | `http.Client.Timeout` = 8 s covering body read; `DisableKeepAlives` |
| Prometheus label-cardinality exhaustion via client-controlled event names | Denial of Service | Bounded `reason` labels only; the `TestMetrics_LabelAllowlist` structural test (Pitfall 7) |
| PII / secrets leaking into the analytics store | Information Disclosure | Per-event key allowlist + `TestEventVocabulary_NoForbiddenKeys` + `maxMetadataValueLen` (Pattern 5) |
| Forged authoritative conversion events from a client | Spoofing, Repudiation | `eventSpec.Authoritative` rejects client submission of the 6 server-owned events; the server attaches user IDs from the auth context, never from the payload |
| Duplicate conversion events inflating the funnel | Repudiation (metric integrity) | Partial unique index with `COALESCE` expressions + `ON CONFLICT DO NOTHING` (Pattern 4, Pitfall 11) |
| SQL injection via card names or event metadata | Tampering | Parameterised queries throughout; `pq.Array` for the `text[]` batch. The existing code is already parameterised — **do not** introduce `fmt.Sprintf` into the new `LATERAL` query when adding the `LIMIT`; bind it |
| Unbounded decklist input consuming memory | Denial of Service | Max input size and max line count enforced before scanning |
| Unauthenticated `previewDeck` as a card-database scraping oracle | Information Disclosure | Low severity (card data is public), but rate limiting bounds it (Open Question 2) |

## Sources

### Primary (HIGH confidence)

- **In-repo, read this session:** `go.mod`, `go.sum`, `Makefile`, `schema.hcl`, `server/cards.go`, `server/games.go` §720-940, `server/graphql.go` §37-43 & §250-340, `server/main_test.go`, `server/graphql_metrics_test.go`, `server/gamelog_helpers.go`, `server/schema.graphql`, `app/package.json`, `app/src/stores/auth.ts`, `app/src/services/scryfall.ts`, `persistence/import_all_printings_json.go`, `persistence/migrations/*.sql`, `persistence/migrations/atlas.sum`, `persistence/migrations_test/README.md`, `test/decklists/jarad.csv`, `dev.docker-compose.yml`, `.github/workflows/*.yml`, `.planning/*`
- **Go module cache, read this session:** `prometheus/client_golang@v1.11.0` (`prometheus/histogram.go`, `prometheus/registry.go`, `prometheus/promauto/auto.go`, `prometheus/testutil/testutil.go`), `agnivade/levenshtein@v1.2.1/levenshtein.go`
- **Go stdlib source, read this session:** `$(go env GOROOT)/src/net/netip/netip.go` §520-661; `go doc net.Dialer`, `go doc net/http.Client`
- **PostgreSQL source, fetched this session:** `postgres/postgres` `REL_14_STABLE` `contrib/pg_trgm/pg_trgm.control` and `contrib/fuzzystrmatch/fuzzystrmatch.control`
- **postgresql.org/docs/current/pgtrgm.html** — operators, GUCs, `gin_trgm_ops` vs `gist_trgm_ops`, KNN support
- **postgresql.org/docs/current/fuzzystrmatch.html** — `levenshtein` signatures, 255-char limit, no index acceleration
- **postgresql.org/docs/15/release-15.html** — `UNIQUE NULLS NOT DISTINCT` introduction
- **api.scryfall.com** — live card and printing counts (32,643 / 96,589, 2026-08-04)
- **raw.githubusercontent.com/lheyberger/mtg-parser** — `grammar.py`, `decklist.py` (reference decklist grammar)
- **Commands run:** `go list -m -versions`, `go mod why -m`, `go list ./...`, `go test ./pkg/...`, `go test ./server/... -run TestNothingZZZ`

### Secondary (MEDIUM confidence)

- **agwa.name/blog/post/preventing_server_side_request_forgery_in_golang** — the canonical Go SSRF `Control`-hook pattern and the TOCTOU argument; cross-checked against the stdlib `netip` source
- **pkg.go.dev/code.dny.dev/ssrf** — IANA-derived IPv4/IPv6 deny prefix lists, default ports 80/443
- **blog.doyensec.com/2022/12/13/safeurl.html** — independent confirmation that DNS-rebinding protection belongs in the `Control` hook
- **proxyfoundry.com / trinketkingdom.com** — Archidekt and Moxfield plain-text export shapes, `(SET)` and backtick-category conventions

### Tertiary (LOW confidence)

- General GIN-vs-GiST crossover guidance around ~100k entries (WebSearch aggregate, not benchmarked here). Used only to *support* a recommendation already decided on correctness grounds (D-03), never as the basis for it. See Assumptions Log A5.

## Metadata

**Confidence breakdown:**

- **Standard stack:** HIGH — every dependency is already in `go.mod`/`go.sum` or is PostgreSQL contrib whose control file was fetched from the postgres repository this session. No package name originated from training memory or an unverified search result.
- **Architecture (parser, resolution, events, metrics):** HIGH — grounded in files read this session with line citations; the GraphQL contract, migration parity, and gqlgen policy are locked constraints rather than inferences.
- **Similarity design:** HIGH on correctness (the D-03 → GiST argument follows directly from the PostgreSQL docs' own statement about GIN and distance operators), MEDIUM on performance (no MTGJSON-sized table was available to measure; see A3, A5, A8).
- **SSRF implementation:** MEDIUM-HIGH — the hook mechanics and stdlib semantics are verified against Go source and `go doc`; the deny-prefix list and the "put validation in `Control`" doctrine are corroborated across three independent sources but not exhaustively re-derived from the IANA registries.
- **Per-event metadata keys:** LOW — see Assumptions Log A1. The PRD must be read for the authoritative table; the plan needs an explicit transcription task.
- **Pitfalls:** HIGH — every pitfall cites either a specific repository line, a stdlib source comment, or a PostgreSQL release note. The three parser defects were confirmed by reading `server/games.go:827-836`, and the missing `cards` name index and the CI gap were both confirmed empirically.

**Research date:** 2026-08-04
**Valid until:** 2026-09-03 (30 days). Stable domain: PostgreSQL 14, the pinned Go dependency set, and the repository state are all fixed. Re-verify sooner only if the MTGJSON snapshot is refreshed (changes the `card_names` projection size) or if the pinned `prometheus/client_golang` version is bumped past v1.14 (native histograms would change Pattern 6's bucket advice).

---

*Phase: 1-Measured Deck Import Foundation*
*Researched: 2026-08-04*

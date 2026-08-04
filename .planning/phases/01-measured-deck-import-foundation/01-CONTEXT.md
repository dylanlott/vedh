# Phase 1: Measured Deck Import Foundation - Context

**Gathered:** 2026-08-03
**Status:** Ready for planning

<domain>
## Phase Boundary

One canonical server-side deck-import path that turns any decklist a Commander player
already has into a trustworthy preview, plus the product-event and metric substrate that
makes the whole activation funnel measurable.

Covers REQ-ACT-001 (product event and activation metric foundation), REQ-ACT-002
(canonical deck parser and preview API), and REQ-ACT-003 (public deck provider
feasibility gate and first adapter). Strictly sequential: ACT-001 → ACT-002 → ACT-003.

Partially advances REQ-A2, REQ-A6, REQ-A7. Does **not** build any activation UI — the
`DeckImportPanel` and commander-review components are ACT-004 in Phase 2. This phase
delivers the server-side contract those components will consume.

</domain>

<decisions>
## Implementation Decisions

### Deck import semantics — unresolved cards

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

### Deck export parsing and printing metadata

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

### Provider feasibility gate (ACT-003)

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

### Product events and metrics

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

### Source-type taxonomy

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

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Source specifications for this milestone
- `docs/product/2026-07-23-deck-to-game-activation-prd.md` — product intent, personas,
  user stories A1–A7, success metrics, non-goals, and the authoritative product event
  vocabulary that D-20 closes over.
- `docs/plans/2026-07-23-deck-to-game-activation-tickets.md` — ACT-001/002/003 scope,
  acceptance criteria, verification steps, and likely-files lists. Authority for technical
  contracts under SPEC > PRD precedence.

### Synthesized planning intel
- `.planning/intel/requirements.md` — full acceptance criteria and verification for
  REQ-ACT-001..003, with the load-bearing dependency graph.
- `.planning/intel/constraints.md` — all 18 constraints. The SSRF controls, privacy
  allowlists, Prometheus cardinality rule, and migration-parity requirement referenced
  throughout the decisions above live here in full.
- `.planning/intel/context.md` — PRD product framing the SPEC does not address, including
  the per-event metadata keys that D-21 transcribes.
- `.planning/INGEST-CONFLICTS.md` — INFO-1 in particular: a shipped provider adapter is
  conditional on the ACT-003 gate, and a no-go must not block the Phase 5 release gate.

### Code the parser work must respect
- `server/games.go` §782-912 — `createLibraryFromDecklist`, the function ACT-002 extracts.
  Commander-budget removal (802-851) and the `100 - commanders` size rule (862) are
  behaviors that must be preserved.
- `server/games.go` §742-759 — `upsertGame`, the JSONB persistence that makes D-08 true.
- `server/cards.go` §140-165 — `Cards()`, the exact-match name lookup that D-02 and D-09
  both need to change.
- `server/schema.graphql` — the `Card` type gaining D-07's fields; `ScryfallID` and
  `TCGID` already exist but are not populated by the current SELECT.
- `schema.hcl` — the `cards` table with `setcode`, `number`, `scryfallid`, `uuid` already
  present, and the `games` table with its `payload` JSONB column.
- `server/graphql.go` §295 — the `/prometheus` endpoint behind `withMetricsAuth`,
  `METRICS_ENABLED`, and `METRICS_TOKEN`. Only `promhttp.Handler()` is registered; there
  are **no custom collectors yet**, so every counter and histogram in ACT-001 is net-new.

### Adjacent work — read for boundaries, do not implement
- `docs/plans/2026-05-15-vedh-auth-storage-migration.md` — the JWT-to-HttpOnly-cookie
  migration. Out of scope; relevant because D-18 puts the session ID in `localStorage`
  alongside the auth token this plan will eventually move.
- `app/src/services/scryfall.ts` — client-side card image fetching. The consumer that makes
  D-07's printing metadata worth persisting.
- `app/src/stores/auth.ts` — existing `localStorage` usage, the pattern D-18 follows.
- `persistence/migrations/` and `persistence/migrations_test/` — golang-migrate,
  `YYYYMMDDHHMMSS_name.{up,down}.sql`. ACT-001's `product_events` migration must ship four
  files across both directories with passing up/down tests, per the migration-parity constraint.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`cards` table printing columns** — `setcode`, `number`, `scryfallid`, `uuid` already
  exist with one row per printing. D-09's disambiguation needs no new storage, only a wider
  SELECT and a changed WHERE.
- **`games.payload` JSONB** — makes D-07's persistence free. Any field added to the Go
  `Card` struct persists with no migration.
- **`withMetricsAuth` + `METRICS_ENABLED`/`METRICS_TOKEN`** — the metrics endpoint is already
  gated and deployed; ACT-001 registers collectors into an existing, secured surface.
- **`persistence/migrations_test/`** — the parity convention is established across ~20
  existing migration pairs; ACT-001's `product_events` migration follows a well-worn path.

### Established Patterns
- **Whole-game JSONB serialization** — board state is not relational. This is why D-07/D-08
  are cheap, and also why adding fields is the natural extension mechanism here rather than
  normalizing card data into its own table.
- **`localStorage` for client identity** — `app/src/stores/auth.ts` already stores the JWT
  there, so D-18 introduces no new storage pattern (and no new consent surface).
- **Batch card lookup by name** — `Cards()` takes a name slice and returns a slice; the
  positional/name-keyed mapping is where the arbitrary-printing behavior originates.

### Integration Points
- **`createLibraryFromDecklist` is the single seam.** Both `CreateGame` (`server/games.go:601`)
  and `JoinGame` (`server/games.go:491`) call into deck import, which is exactly why ACT-002's
  "one canonical parser serves host and join" is achievable without touching the mutations.
- **`server/schema.graphql`** — `previewDeck`, `DeckPreview`, and `trackProductEvent` are
  additive. Generated gqlgen files are updated via `make generate`, never hand-edited.
- **No frontend event service exists.** `app/src/services/` has `apollo.ts`,
  `commanderPartner.ts`, and `scryfall.ts` only. ACT-001's event service is a new file
  following that directory's conventions.

</code_context>

<specifics>
## Specific Ideas

- **The six syntaxes are non-negotiable and currently mostly broken.** `1 Sol Ring`,
  `1x Sol Ring`, `1,Sol Ring`, `1, Sol Ring`, `1,"Atraxa, Praetors' Voice"`, and
  `1 Atraxa, Praetors' Voice` must all parse. Today only the comma forms work at all:
  `csv.Reader` requires `len(record) >= 2`, so `1 Sol Ring` fails with "invalid decklist
  row" and `1x Sol Ring` fails `ParseInt`.

- **The comma bug is a silent truncation, not an error.** `1,Atraxa, Praetors' Voice`
  parses to `["1", "Atraxa", " Praetors' Voice"]`; the code takes `record[1]` and discards
  the rest. The player gets a card named "Atraxa" with no warning. Any test for D-11's
  no-silent-drop guarantee should cover this exact input.

- **Custom proxy prints are the reason D-07 persists rather than passes through.** The user
  plans a later feature swapping in custom art for specific cards. Name-only identity cannot
  express "this printing", so the metadata has to be captured at import time or it is lost.
  The feature itself is out of scope here — see Deferred Ideas.

- **Worst-case preview cost is a real risk.** D-03 (always show nearest) plus D-04 (compute
  eagerly) means a 100-card paste where nothing resolves runs similarity for all 100 rows
  inside the `previewDeck` request — on the path feeding the ≤60s median / ≤120s p90
  time-to-board target. The planner should bound this work explicitly (a cap on rows given
  suggestions, a query budget, or a fast pre-filter) rather than discovering it under load.

- **Today a bad row fails the entire deck.** There is no warning/partial model at all —
  every error path returns `nil, err`. The warnings / blocking-errors / `CanContinue`
  triad in ACT-002 is a new concept, not a refinement of an existing one.

</specifics>

<deferred>
## Deferred Ideas

- **Custom proxy prints for specific cards** — the motivation behind D-07's persistence, and
  a genuinely appealing feature, but a new user-facing capability that belongs in its own
  phase well after activation ships. Phase 1 only ensures the printing identity needed to
  build it is captured rather than discarded. Do not implement any proxy or custom-art
  behavior in this phase.

</deferred>

---

*Phase: 1-Measured Deck Import Foundation*
*Context gathered: 2026-08-03*

# Phase 2: Guest Host Activation - Research

**Researched:** 2026-08-08
**Domain:** Guest identity (Go/GraphQL/JWT/Postgres) + reusable deck-import UI (Vue 3/Pinia) + first responsive breakpoint
**Confidence:** HIGH (codebase patterns directly read this session; no new external packages)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Guest lifetime and identity durability**

- **D-2.1:** Guest rows live **indefinitely**, distinguished by the `is_guest` flag rather
  than reaped on a timer. This resolves **OPEN-2** as *neither* of its stated options —
  the question asked "24 hours or seven days"; the answer is that the backing row never
  expires at all. The 24-hour token from REQ-ACT-005 is unchanged.
  — Reversibility: one-way.
- **D-2.2:** `expires_at` **stays on the schema and stays enforced in authorization**, but
  it governs *only* whether a session credential may mint a **new** token. It never triggers
  row deletion. The independently-retryable cleanup path REQ-ACT-005 requires is still built
  and still tested, but targets only rows explicitly marked for removal — under D-2.1 it has
  no scheduled work to do. — Reversibility: reversible.
- **D-2.3:** A guest whose 24-hour token expires gets a **silent re-issue** while their row
  is alive. Because D-2.1 makes the row always alive, the token becomes an implementation
  detail rather than a user-facing deadline. Nobody is shown an "expired session" wall.
- **D-2.4:** Silent re-issue makes the re-auth value a **bearer credential**, so it must be
  a **separate secret from the analytics session ID**. Phase 1's `edhgo/session-id` stays
  exactly what it is — a low-stakes attribution identifier. A *distinct* high-entropy guest
  credential is minted, stored under its own localStorage key, and sent only to the auth
  path. — Reversibility: costly.

**Guest and display names**

- **D-2.5:** Generated guest names are **MTG-flavored adjective-noun** pairs (e.g. "Brave
  Sliver") from a **curated** word list. No pairing may be offensive or absurd. A fixed list
  bounds the namespace, so a numeric suffix is the expected overflow strategy.
- **D-2.6:** Generated names are **globally unique across all users** — the same namespace
  as real usernames, colliding with neither. Requires retry-on-conflict at generation. Names
  are **never recycled** and the namespace only ever grows (interacts with D-2.1).
- **D-2.7:** A visitor-typed display name is **free-form and NOT unique**. Two people may
  both be "Dylan"; the row ID is the identity and the name is a label. No "that name is
  taken" failure may appear on the activation path.
- **D-2.8:** D-2.6 and D-2.7 cannot both live in today's `username` column, which is
  `UNIQUE`. Therefore: add a **new nullable, non-unique `display_name` column**. `username`
  remains the unique system identifier and holds the generated MTG name for guests. **Every
  surface currently rendering `Username` must switch to `display_name ?? username`** — this
  is a wide sweep (`BoardView.vue`, `ScoreView.vue`, `GamesView.vue`, `AppNav.vue` at
  minimum) and is easy to under-do. — Reversibility: one-way.
- **D-2.9:** At claim time the permanent username is **freely chosen, prefilled** from
  whatever the user has been called — typed display name if present, generated name
  otherwise. The prefill may collide, so the claim form carries the ordinary taken-name path.
  Only the `claimGuestAccount` backend contract is in scope this phase.

**Viewport and responsive baseline**

- **D-2.10:** The activation path supports **desktop and tablet, ≥768px** — one breakpoint,
  applied to the **new activation components only**. Resolves **OPEN-3**. Zero `@media`
  queries exist in the codebase today; ACT-011 (Phase 4) inherits this shape.
- **D-2.11:** `/play` on a phone **works, with a heads-up**. The import flow functions;
  a visible notice sets expectations about the board. No hard bounce.

**Failure recovery**

- **D-2.12:** In-progress activation state — paste, corrected entries, commander choices,
  display name — lives in **`sessionStorage`**. Survives refresh and failed navigation, dies
  with the tab. **Must be explicitly cleared on successful game creation.**
- **D-2.13:** The four recoverable failure classes (provider down, preview error,
  guest-session error, create error) carry a **stable machine-readable error code from the
  server**, mapped by the client to copy *and* to the correct affordance. Model on Phase 1's
  closed event vocabulary: fixed, server-owned, allowlisted. — Reversibility: costly.

### Claude's Discretion

- **Guest-creation kill switch default.** REQ-ACT-005 mandates a feature flag and kill
  switch. Phase 1's provider switch (`DECK_PROVIDER_ENABLED`) defaults off. Decide the
  default explicitly in the plan; do not copy-paste Phase 1's default without deciding.
- **`/play`'s relationship to existing flows.** How aggressively `FormCreateGame.vue` /
  `JoinGameView.vue` / `LandingView.vue`'s CTA change is open. Both a hard replace and a
  parallel-run reading satisfy the requirement text.
- **Commander review UX for partners and backgrounds.** Whether the new commander-review
  component consumes `app/src/services/commanderPartner.ts` as-is, extends it, or supersedes
  it is open.

### Deferred Ideas (OUT OF SCOPE)

- **Claim-account UI.** `claimGuestAccount` ships as backend only this phase (REQ-A5
  "backend"); the claim form, prompt timing, and "you're about to lose this game" nudge are
  Phase 4 (ACT-010). D-2.9 records the intended prefill behavior for that phase to inherit.
- **Board responsiveness.** D-2.10 covers the activation path only; `BoardView.vue`'s own
  responsive redesign is Phase 4 (ACT-011).
- **Guest name rename.** D-2.9 lets a claimer choose a fresh username; no rename feature
  exists for anyone afterward, and none is added here.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REQ-ACT-004 | Reusable deck import and commander review UI (`DeckImportPanel.vue`, commander-review component, replaces duplicated textareas in `FormCreateGame.vue`/`JoinGameView.vue`) | Architecture Patterns (component structure), Common Pitfalls (state-preservation, a11y), Code Examples (previewDeck client wiring — first frontend consumer) |
| REQ-ACT-005 | Guest identity and account-claim backend (`is_guest`/`expires_at`/`display_name` columns, `guestSession`/`claimGuestAccount` mutations, rate limit, kill switch, cleanup job) | Migration Shape, Guest/Password-Login Auth Surface, Rate Limiting + Kill Switch, Don't Hand-Roll (bcrypt/JWT reuse) |
| REQ-ACT-006 | Public quick-start host flow (`/play`, `QuickStartView.vue`, router guard, `createGame` reuse, server-written events) | The `/play` Public-But-Stateful Route Case, Server-Written Product Events, Architectural Responsibility Map |
</phase_requirements>

## Summary

This phase adds exactly one new schema surface (three nullable/boolean columns on `users`)
and two new mutations (`guestSession`, `claimGuestAccount`) to a codebase whose GraphQL
resolvers are unusually easy to extend: `graphQLServer` itself satisfies gqlgen's
`ResolverRoot` (verified at `server/graphql.go:367`, `Resolvers: s`), so every new resolver
is a plain Go method on `*graphQLServer` in a new file (`server/guest_users.go`, matching
the ticket's own `likely_files`), not a generated-stub edit. `CreateGame` already calls
`requireAuth(ctx)` (`server/games.go:631`) and cares only about the JWT subject/username —
a guest's JWT satisfies it with zero special-casing, so `/play`'s host path is "mint a
guest, then call the existing `createGame`," exactly as REQ-ACT-006 states.

The two hardest parts of this phase are not the guest mutations themselves — auth.go,
ratelimit.go, and product_events.go already establish every pattern needed (JWT issuance,
per-surface token buckets, server-authoritative event writes with dedup) — they are (1) the
`display_name ?? username` sweep, which touches GraphQL query/mutation *strings* as well as
`.vue` templates (a query that doesn't ask for `DisplayName` will never receive it, no
matter how the template renders), and (2) the `/play` router-guard exception, which requires
a genuinely new `beforeEach` branch since every existing route is binary
public-marketing-page or `requiresAuth: true` with no "public but has its own state" case
today.

One thing this phase must build that appears nowhere in REQ-ACT-005's ticket text: D-2.3's
silent re-issue needs an actual mutation or transport to exchange the new bearer credential
for a fresh JWT. No such endpoint exists yet (`guestSession` only *creates*). See Open
Questions.

**Primary recommendation:** Add `is_guest BOOLEAN NOT NULL DEFAULT false`,
`expires_at TIMESTAMPTZ`, and `display_name VARCHAR(255)` to `users` via matched migration
pairs in both `persistence/migrations/` and `persistence/migrations_test/` (same filename,
byte-identical content, per the established pattern); add `guestSession`/`claimGuestAccount`
resolvers on `*graphQLServer` in `server/guest_users.go`; add a `ratelimit.SurfaceGuestSession`
constant (the package comment at `pkg/ratelimit/limiter.go:25-27` already anticipates this);
gate guest creation behind a new `GUEST_CREATION_ENABLED` envconfig kill switch mirroring
`DeckProviderEnabled`'s shape; and route `/play` through a new `meta: { publicStateful: true }`
(or equivalent) `beforeEach` branch, never `meta.public` or `meta.requiresAuth`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Deck paste/URL import + preview | Browser (Vue component) | API/Backend (`previewDeck` mutation, already built in Phase 1) | UI collects/displays; parsing and card resolution are server-side and already exist — this phase adds no new parsing. |
| Commander review/selection | Browser (Vue component) | — | Pure client-side selection UI over `CommanderCandidates` already returned by `previewDeck`; `commanderPartner.ts` logic already lives here. |
| Guest identity creation | API/Backend | Database/Storage (`users` table) | Must be server-authoritative: password hash, token issuance, and uniqueness retry cannot be trusted to a client. |
| Silent token re-issue | API/Backend | Browser (stores the new bearer credential) | Server validates the re-auth credential and re-signs a JWT; browser only stores/sends it. |
| `display_name ?? username` resolution | API/Backend (GraphQL field) + Browser (template fallback) | — | Server should return `DisplayName` as a real field (so every client gets it uniformly); Vue templates apply the `??` fallback for rendering. Both tiers participate — this is not purely a frontend cosmetic change. |
| In-progress activation state (paste, corrections, commander choices, display name) | Browser (`sessionStorage`) | — | D-2.12 is explicit: survives refresh/failed nav, dies with the tab. No server round-trip needed to preserve it. |
| `/play` route public-but-stateful gating | Browser (Vue Router `beforeEach`) | API/Backend (still enforces auth per-mutation independently) | The router guard is a UX convenience; the real authorization boundary is server-side per-resolver (`requireAuth`), matching the existing pattern where `publicQueries` in `authz.go` is a query-level allowlist, not a transport-level one. |
| Rate limiting / kill switch for guest creation | API/Backend | — | Mirrors `DeckProviderEnabled`/`allowRequest` exactly; a client cannot self-limit. |
| Server-authoritative event writes (`game_created`, `guest_session_created`) | API/Backend | Database (`product_events` table + dedup index) | Already the pattern for all six authoritative events (Phase 1); this phase populates two of the six for the first time. |
| First responsive breakpoint (≥768px) | Browser (scoped SFC `@media`) | — | No design-token file exists; the codebase's CSS architecture is per-component `<style scoped lang="scss">` blocks reading global CSS custom properties from `:root` in `app/src/styles/main.scss`. |

## Standard Stack

No new external packages are required this phase. Every primitive REQ-ACT-004/005/006 need
already has a working, tested implementation in the repo from Phase 1:

### Core (existing, reused — not newly installed)
| Library | Version (as pinned) | Purpose | Why reused, not replaced |
|---------|---------|---------|--------------|
| `golang-jwt/jwt/v5` | pinned in `go.mod` (already in use) | JWT issuance/parsing | `server/auth.go:110-129` (`newAuthToken`) and `:83-108` (`parseAndValidateToken`) already do exactly what a guest token needs — a guest is just a `User` row with `newAuthToken(user, 24*time.Hour)` called on it. [VERIFIED: server/auth.go:110-129] |
| `golang.org/x/crypto/bcrypt` | pinned in `go.mod` | Password hashing (guest's random non-recoverable password) | `server/users.go:104-107` (`hashPassword`) — a guest's random password must go through the identical `bcrypt.GenerateFromPassword(..., 14)` call so `claimGuestAccount` can reuse `checkPasswordHash`/bcrypt-rules unmodified. [VERIFIED: server/users.go:104-107] |
| `github.com/google/uuid` | pinned in `go.mod` | Guest row UUID | `server/users.go:23` (`uuid.New().String()`) — identical pattern for the guest's `uuid` primary identifier. [VERIFIED: server/users.go:23] |
| `pkg/ratelimit` (in-repo) | n/a (internal package) | Guest-creation rate limiting | `pkg/ratelimit/limiter.go:24-29` — the `Surface` type's own doc comment: "Guest creation and public invite lookup (later phases...) add their own constants here without changing Registry's shape at all." [VERIFIED: pkg/ratelimit/limiter.go:24-29] |
| `pkg/telemetry` (in-repo) | n/a (internal package) | `guest_session_created`/`game_created` event validation + dedup | Both events already declared `Authoritative: true` in the closed vocabulary. [VERIFIED: pkg/telemetry/vocabulary.go:68-69] — quoted verbatim: `"guest_session_created": {Authoritative: true, Keys: keys()},` and `"game_created":          {Authoritative: true, Keys: keys()},` |
| `vue-router` v4.4.3 | `app/package.json` (existing) | `/play` public route + guard | Already the router in use; `beforeEach` at `app/src/router/index.ts:79-88` needs one new branch. [VERIFIED: app/src/router/index.ts:79-88] |
| `pinia` v2.1.7 | `app/package.json` (existing) | New `guest`-aware store additions | `app/src/stores/auth.ts` and `app/src/stores/games.ts` are Pinia `defineStore` composition-API stores; extend, don't replace. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `@apollo/client` v3.11.0 (existing) | Mutation calls for `guestSession`/`claimGuestAccount`/`previewDeck` | Client wiring — no new client library needed; follow `app/src/graphql/mutations.ts`'s existing `gql` tag pattern. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| In-repo curated adjective-noun word list (D-2.5) | A third-party "random name generator" npm/Go package | User locked D-2.5 to a *curated* list ("no pairing may be offensive or absurd") — a third-party package's word list cannot be audited/curated to that bar, and pulling one in is an unjustified new dependency for ~50-100 words. Build the list in-repo (Go, next to `backronym.go`'s sibling pattern, though `backronym.go` itself is unrelated and not reusable — see Don't Hand-Roll). |
| Server-side session-credential storage in Postgres | A dedicated Redis/session store | `test/redis.go` exists but is unused by any production path found this session; adding a new datastore for one bearer-credential column is disproportionate. Store the re-auth credential's *hash* on the existing `users` row (a new nullable column) — same pattern as the password hash. |

**Installation:** none — no `npm install` / `go get` required.

## Package Legitimacy Audit

Not applicable — this phase installs no new external packages. All primitives (JWT, bcrypt,
uuid, rate limiting, event vocabulary) reuse packages already vendored and verified in Phase
1 (`go.mod`, `app/package.json`).

## Migration Shape

**Pattern discovered this session:** production and test schemas are kept in sync by
maintaining **two directories with matching filenames and byte-identical SQL content** for
every migration that affects both:

```
persistence/migrations/20260804120000_product_events.up.sql
persistence/migrations_test/20260804120000_product_events.up.sql
```

[VERIFIED: diff of the two files above, run this session — `diff` produced no output
("IDENTICAL")]. `persistence/migrations_test/` additionally carries test-only migrations
(`20260131000000_seed_test_cards.up.sql`, `README.md`) that do not exist in
`persistence/migrations/` — those are *additions*, not divergences of shared files.
[VERIFIED: `ls` diff of both directories, run this session]

Production migrations run via `persistence.NewPostgres(defaultMigrationsDir, ...)` where
`defaultMigrationsDir = "./persistence/migrations/"` [VERIFIED: persistence/sql.go:20 —
`defaultMigrationsDir = "./persistence/migrations/"`]. Tests run via
`persistence.ForceCleanMigrations("../persistence/migrations_test/", cfg.PostgresURL)` and
`persistence.NewPostgres("../persistence/migrations_test/", cfg.PostgresURL)`
[VERIFIED: server/test.go:76-79].

**Required for this phase:** create a new timestamped pair, e.g.
`persistence/migrations/{TS}_guest_users.up.sql` /
`persistence/migrations_test/{TS}_guest_users.up.sql` (same timestamp, same content), adding:

```sql
ALTER TABLE users ADD COLUMN is_guest BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN expires_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255);
```

The current `users` table has exactly four columns — `username VARCHAR(255)`,
`password VARCHAR(255)`, `uuid VARCHAR(255) UNIQUE`, `timestamp TIMESTAMP DEFAULT
CURRENT_TIMESTAMP` [VERIFIED: persistence/migrations/20210307154621_init_db.up.sql:1-6,
quoted verbatim: `CREATE TABLE IF NOT EXISTS users ( username VARCHAR(255), password
VARCHAR(255), uuid VARCHAR(255) UNIQUE, timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP );`]
— plus the later `username_unique` constraint [VERIFIED:
persistence/migrations/20211229130157_enforce_unique_usernames.up.sql:1, quoted verbatim:
`ALTER TABLE users ADD CONSTRAINT username_unique UNIQUE (username);`]. `username` has **no
column-level `UNIQUE`**, only the separately-added named constraint — a `down.sql` for the
new migration should drop the three added columns by name; it does not need to touch
`username_unique`.

Write a matching `.down.sql` dropping the three columns, in both directories, per the
established pattern (every `.up.sql` file this session has a sibling `.down.sql`).

**Also required:** a nullable column for D-2.4's guest re-auth credential (hashed, not
plaintext — see Guest/Password-Login Auth Surface below), e.g. `guest_credential_hash
VARCHAR(255)`, in the same migration pair, since it belongs to the same `users` row and the
same rollout.

## Guest/Password-Login Auth Surface

**Token issuance (reusable as-is):** `newAuthToken(user *User, ttl time.Duration)`
[VERIFIED: server/auth.go:110-129] builds a JWT with `Subject: user.ID`,
`Username: user.Username` claims and a caller-supplied TTL. `Signup`/`Login` both call it
with `time.Hour*24` [VERIFIED: server/users.go:44, :88]. A guest token is
`newAuthToken(guestUser, time.Hour*24)` — no new signing logic needed, only a new call site.

**Password-login rejection (must be added):** `Login(ctx, username, password)`
[VERIFIED: server/users.go:53-102] currently does
`SELECT "uuid", "username", "password" FROM "users" WHERE username=$1` (line 62, quoted
verbatim: `q := \`SELECT "uuid", "username", "password" FROM "users" WHERE username=$1;\``)
and checks only `checkPasswordHash`. **It has no `is_guest` check today.** REQ-ACT-005's
acceptance clause "Guest users cannot log in through the password endpoint before claiming"
must be implemented by selecting `is_guest` alongside the existing three columns and
returning an error (mirroring the existing `errs.New("failed to authenticate")` style) when
`is_guest` is true — this is a **defense-in-depth** check: because a guest's password is a
server-generated random value never revealed to any client, an attacker cannot supply the
correct password anyway, but the requirement is explicit and cheap to satisfy directly.

**Authorization / expiry enforcement (must be added, D-2.2):** no `is_guest`/`expires_at`
check exists anywhere in `authz.go` today [VERIFIED: server/authz.go, full file read this
session — contains only `isPublicQuery`, `requireAuth`, `requireMatchingUser`,
`isUserInGame`, none of which reference expiry]. Per D-2.2, `expires_at` must gate **minting
a new token only** — i.e., it belongs in whatever new mutation/logic re-issues a JWT for an
existing guest row (see Open Questions — no such mutation exists yet), not in `requireAuth`
itself (an already-issued, unexpired JWT must keep working against every existing resolver
exactly as today, since D-2.1 keeps the row alive indefinitely and D-2.3 makes token
re-issue silent).

**`AuthUser` context object is minimal today:** `type AuthUser struct { ID string;
Username string }` [VERIFIED: server/auth.go:16-19]. It carries no `IsGuest` field. Whether
any resolver needs to distinguish guest-vs-real at the `AuthUser` level (vs. querying
`users.is_guest` fresh when needed, e.g. inside `claimGuestAccount`) is a planner decision —
`claimGuestAccount` itself will need a DB lookup by the authenticated `AuthUser.ID` regardless
of what's in the JWT claims, since the claim path must verify the *current* row is still a
guest at claim time, not what the (already-issued) JWT said when it was minted.

**Resolver wiring (confirms REQ-ACT-005's `likely_files` claim):** `graphQLServer` is passed
directly as the gqlgen resolver root — `NewExecutableSchema(Config{Resolvers: s})`
[VERIFIED: server/graphql.go:367, quoted: `gqlHandler := handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),`].
The generated `server/schema.resolvers.go` stub methods on `mutationResolver`/`queryResolver`
(e.g. the `PreviewDeck` stub at line 58-60 that still `panic`s) are **dead code, never
invoked** — every real resolver (`Signup`, `Login`, `CreateGame`, `JoinGame`, `PreviewDeck`,
`TrackProductEvent`) is a plain method on `*graphQLServer` in its own topic file
(`users.go`, `games.go`, `deck_import.go`, `deck_providers.go`). **New resolvers
`GuestSession` and `ClaimGuestAccount` belong as methods on `*graphQLServer` in a new
`server/guest_users.go`**, matching the exact pattern and the ticket's own likely_files list.

## Rate Limiting + Kill Switch

**Rate limiting pattern (`pkg/ratelimit`):** a `Registry` keyed by `(Surface, clientKey)`
[VERIFIED: pkg/ratelimit/limiter.go:60-75]. Two surfaces exist today:
`SurfaceDeckImport = "deck_import"` and `SurfaceProductEvent = "product_event"`
[VERIFIED: pkg/ratelimit/limiter.go:32-37]. The package's own doc comment on `Surface`
anticipates this phase by name: "Guest creation and public invite lookup (later phases, per
`.planning/intel/constraints.md`'s 'Rate limiting on public surfaces') add their own
constants here without changing `Registry`'s shape at all" [VERIFIED:
pkg/ratelimit/limiter.go:24-29, quoted]. **Action:** add
`SurfaceGuestSession Surface = "guest_session"` to that same `const` block.

`clientKeyFor(ctx, sessionID)` [VERIFIED: server/ratelimit.go:71-81] prefers the
server-observed remote address (`"addr:" + addr`) and falls back to the caller-supplied
`sessionID` only when no address is available — the same key derivation should be reused
unmodified for `guestSession`, since the whole point (documented in the CR-02 comment at
lines 49-70) is that a caller cannot escape the limiter by rotating `sessionID`.

`allowRequest(ctx, surface, clientKey)` [VERIFIED: server/ratelimit.go:34-46] is the thin
wrapper every rate-limited resolver calls; a `nil` `s.limiter` (bare struct-literal test)
allows everything, so tests that don't care about rate limiting are unaffected — reuse this
exact call shape in the new `GuestSession` resolver: `if !s.allowRequest(ctx,
ratelimit.SurfaceGuestSession, clientKeyFor(ctx, ...)) { ... }`.

**Kill switch pattern (`DeckProviderEnabled`):**
```go
// server/graphql.go:63-69 [VERIFIED]
// DeckProviderEnabled is the provider kill switch (D-16/T-01-28).
// Defaults to false: a deploy must be safe before this is
// deliberately turned on...
DeckProviderEnabled bool `envconfig:"DECK_PROVIDER_ENABLED" default:"false"`
```
and its consuming predicate:
```go
// server/deck_providers.go:337-339 [VERIFIED]
func (s *graphQLServer) providerEnabled() bool {
	return s != nil && s.cfg.DeckProviderEnabled && len(s.deckProviderAllowedHosts) > 0
}
```
**Action:** add a new `Conf` field, e.g. `GuestCreationEnabled bool
`envconfig:"GUEST_CREATION_ENABLED" default:"false"`` (mirroring the doc-comment style and
the `false` default), and a `guestCreationEnabled()` predicate method on `*graphQLServer`
that `GuestSession` checks first, returning a typed "guest creation is currently disabled"
error (mapped to one of D-2.13's error codes) rather than proceeding. **The kill-switch
default is a Claude's Discretion item** (see User Constraints) — Phase 1's own default was
`false` ("ships dark"); the plan must explicitly decide and record whether Phase 2 ships the
same way, since unlike the deck provider (which has an optional paste fallback), a
`false` default here closes the *entire* guest-host funnel this phase exists to open.

**Existing rate-limit config fields to mirror, not reinvent:**
`DeckImportRatePerMinute`/`DeckImportRateBurst` [VERIFIED: server/graphql.go:53-54] —
add `GuestSessionRatePerMinute`/`GuestSessionRateBurst` envconfig fields feeding a
**second** `ratelimit.NewRegistry(...)` instance, or (simpler, and consistent with the
existing single-registry-multiple-surfaces design) reuse the *same* `s.limiter` Registry —
the registry's own `Allow` method already scopes bucket state by `Surface`, so one registry
serving both `SurfaceDeckImport` and the new `SurfaceGuestSession` is the pattern already in
place for `SurfaceProductEvent` sharing that same registry today [VERIFIED:
server/ratelimit.go:34-46 shows `allowRequest` taking `surface` as a parameter against a
single `s.limiter`].

## Architecture Patterns

### System Architecture Diagram

```
Browser (logged out or logged in)
   |
   | 1. GET /play  (new public-but-stateful route)
   v
QuickStartView.vue
   |
   | 2. paste text OR provider URL --> DeckImportPanel.vue
   |    (reused from ACT-004; sessionStorage snapshot on every keystroke/step,
   |    D-2.12)
   v
previewDeck mutation (already built, Phase 1)  -----------> server/deck_import.go
   |                                                          PreviewDeck()
   | 3. DeckPreview{Entries, CommanderCandidates,             (rate-limited,
   |    Unresolved, Warnings, CanContinue, BlockingErrors}     SurfaceDeckImport)
   v
CommanderReview component (new, ACT-004)
   |
   | 4. user selects commander(s), optional display name
   v
   +-- if auth.isAuthenticated --------------------> skip guest creation entirely
   |                                                  (reuse existing authUser)
   |
   +-- else --> guestSession(displayName, sessionID) mutation (NEW, ACT-005)
                    |                                    server/guest_users.go
                    | 5. only called once deck CanContinue == true
                    v                                    - generates/validates name
                AuthProfile{ID, Username, Token, IsGuest} - hashes random password
                    |                                    - issues 24h JWT
                    | stored under edhgo/auth (existing key)
                    v                                    - writes guest_session_created
   +----------------+----------------------------------- (server-authoritative, dedup'd)
   |
   v
createGame mutation (EXISTING, unmodified) --------------> server/games.go CreateGame()
   |                                                          requireAuth(ctx) succeeds
   | 6. Handle=game label, FormatID="EDH", normalized deck    for guest OR real user JWT
   v                                                          - writes game_created
router.push({ name: 'board', params: { id } })                (server-authoritative)
   |
   v
/games/:id (BoardView.vue) -- OUT OF THIS PHASE'S BOUNDARY except the
                               display_name ?? username rendering swap (D-2.8)
```

Client-emitted events along this path (criterion 5): `quick_start_viewed` (on `/play`
mount), `deck_import_started` (on first paste/URL submit), `game_create_started` (just
before the `createGame` call) — all via the existing `track()` in
`app/src/services/productEvents.ts` [VERIFIED: app/src/services/productEvents.ts:212-235].
Server-written events along this path: `guest_session_created`, `game_created` — both
already `Authoritative: true` in the vocabulary [VERIFIED: pkg/telemetry/vocabulary.go:68-69].

### Recommended Project Structure

```
app/src/
├── components/
│   └── decks/
│       ├── DeckImportPanel.vue       # NEW — text/URL input, preview, unresolved correction
│       └── CommanderReview.vue       # NEW — commander selection, consumes commanderPartner.ts
├── components/games/
│   └── FormCreateGame.vue            # MODIFIED — textarea replaced with <DeckImportPanel>
├── views/
│   ├── JoinGameView.vue              # MODIFIED — textarea replaced with <DeckImportPanel>
│   └── QuickStartView.vue            # NEW — /play, orchestrates import -> review -> create
├── stores/
│   ├── auth.ts                       # MODIFIED — AuthProfile gains IsGuest; new guest-session
│   │                                  #   credential persisted under its own storage key
│   └── games.ts                      # UNCHANGED — createGame()/joinGame() reused as-is
├── router/
│   └── index.ts                      # MODIFIED — /play gets a new public-but-stateful meta
├── graphql/
│   ├── mutations.ts                  # MODIFIED — GUEST_SESSION_MUTATION, CLAIM_GUEST_ACCOUNT_MUTATION
│   │                                  #   (backend contract only this phase); DisplayName added
│   │                                  #   to CREATE_GAME_MUTATION/JOIN_GAME_MUTATION responses
│   └── queries.ts                    # MODIFIED — DisplayName added to GAMES_QUERY, GET_GAME_QUERY,
│                                      #   GAME_UPDATED_SUBSCRIPTION Players selections
└── styles/
    └── main.scss                     # possibly gains a shared $breakpoint-tablet SCSS variable
                                       #   or --vedh-breakpoint-tablet custom property (first use)

server/
├── guest_users.go                    # NEW — GuestSession, ClaimGuestAccount resolvers,
│                                      #   name generation/collision retry, cleanup job
├── auth.go                           # possibly gains a re-auth-credential verify/reissue helper
├── users.go                          # MODIFIED — Login() gains is_guest rejection
├── authz.go                          # possibly gains an expiry-aware helper (D-2.2, new-token-only)
├── ratelimit.go                      # UNCHANGED (allowRequest reused as-is)
├── graphql.go                        # MODIFIED — Conf gains GuestCreationEnabled + rate fields
└── schema.graphql                    # MODIFIED — guestSession/claimGuestAccount mutations,
                                       #   DisplayName field on User, IsGuest if needed on client

pkg/ratelimit/limiter.go              # MODIFIED — SurfaceGuestSession constant added

persistence/migrations/{TS}_guest_users.up.sql       # NEW (is_guest, expires_at, display_name,
persistence/migrations_test/{TS}_guest_users.up.sql  #  guest_credential_hash — byte-identical pair)
```

### Pattern 1: Server-authoritative event write with dedup (reuse verbatim)
**What:** Every server-owned conversion event goes through `recordProductEvent`, which
validates against the closed vocabulary and writes with `ON CONFLICT DO NOTHING` against a
partial unique index — this is *the* mechanism that satisfies criterion 5's "a retried
client call does not double-count a conversion."
**When to use:** Every emit site for `guest_session_created` and `game_created` in the new
`GuestSession`/`CreateGame`-adjacent code paths.
**Example:**
```go
// Source: server/product_events.go:151-188 (TrackProductEvent, the client-facing sibling)
// The pattern for a SERVER-OWNED event (clientSubmitted: false) inside GuestSession:
s.recordProductEvent(ctx, ProductEvent{
    Name:      "guest_session_created",
    SessionID: input.SessionID,
    UserID:    &guestUser.ID,
    Outcome:   &outcome, // e.g. "created" or "collision_retried"
}, false) // false = server-authoritative, never subject to RejectionClientAuthoritative
```
The dedup index itself: [VERIFIED: persistence/migrations/20260804120000_product_events.up.sql:30-38,
quoted verbatim: `CREATE UNIQUE INDEX IF NOT EXISTS product_events_authoritative_once ON
product_events (event_name, COALESCE(game_id, ''), COALESCE(user_id, ''), session_id) WHERE
event_name IN ('game_created', 'player_joined', 'guest_session_created',
'account_claimed');`] — `guest_session_created` and `game_created` are both already members
of this four-event dedup set; no migration change needed for dedup itself.

### Pattern 2: Kill-switch-gated feature with required non-empty config (reuse shape)
**What:** A boolean envconfig flag that is insufficient alone — it must be ANDed with a
second, non-empty configuration value before the feature is live.
**When to use:** `providerEnabled()`'s shape is explicitly documented as the template for
"shouldExposeMetrics' shape" already, and should extend to `guestCreationEnabled()` if the
plan decides a second gate is warranted (e.g. requiring `JWT_SECRET` to be non-default, or
a curated word-list file to be non-empty).
**Example:**
```go
// Source: server/deck_providers.go:330-339
func (s *graphQLServer) providerEnabled() bool {
	return s != nil && s.cfg.DeckProviderEnabled && len(s.deckProviderAllowedHosts) > 0
}
```

### Pattern 3: sessionStorage-backed draft state (new — no precedent in codebase)
**What:** D-2.12 requires paste/corrections/commander-choices/display-name to survive
refresh and failed navigation but die with the tab.
**When to use:** `QuickStartView.vue` (and, per REQ-ACT-004's reuse mandate, potentially
`DeckImportPanel.vue` itself as a prop-driven persistence key so `FormCreateGame.vue`'s
authenticated flow can opt out of persistence while `/play` opts in).
**Example (no existing code to cite — `grep -rn "sessionStorage" app/src` returned zero
results this session [VERIFIED: search run this session, no matches]):**
```typescript
// New pattern — model on productEvents.ts's defensive storage accessor shape
// (app/src/services/productEvents.ts:59-71's getStorage()) but for sessionStorage:
const DRAFT_KEY = 'edhgo/quickstart-draft';
function saveDraft(draft: QuickStartDraft): void {
  try { sessionStorage.setItem(DRAFT_KEY, JSON.stringify(draft)); } catch { /* ignore */ }
}
function clearDraft(): void {
  try { sessionStorage.removeItem(DRAFT_KEY); } catch { /* ignore */ }
}
// clearDraft() MUST be called on successful createGame(), per D-2.12.
```

### Anti-Patterns to Avoid
- **Adding `is_guest`/`expires_at` checks inside `requireAuth`.** D-2.2 is explicit that
  expiry governs *minting a new token*, never ongoing authorization of an already-issued
  token. Putting the check in `requireAuth` would make an unexpired JWT for a since-expired
  guest fail mid-game, which is the exact "expired session wall" D-2.3 forbids.
- **Treating `DisplayName` as purely a `.vue`-file concern.** The GraphQL query/mutation
  documents in `app/src/graphql/queries.ts` and `mutations.ts` must also request the field,
  or the template-level `display_name ?? username` fallback always evaluates to `username`
  regardless of how thorough the component sweep is.
- **Building a new rate-limit Registry instance for guest sessions.** The existing
  `Registry` is already multi-surface by design (`bucketKey{surface, clientKey}`); a second
  `Registry` instance would fragment the idle-eviction sweep and the operator-visible
  `vedh_rate_limit_total` metric's mental model for no benefit.
- **Reusing `edhgo/session-id` (the analytics ID) as the guest re-auth credential.** D-2.4
  is explicit these must be different secrets under different keys — the analytics ID is a
  low-stakes, non-authenticating value read by `productEvents.ts`; conflating it with an
  authenticating bearer credential means anyone who can read Prometheus/event-table
  `session_id` values could impersonate a guest.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT signing/verification | A new token format or a second `jwt` library | `newAuthToken`/`parseAndValidateToken` in `server/auth.go` (already used by Signup/Login) | One signing key (`JWT_SECRET`), one claims shape, one verification path — a second implementation would create two ways to forge trust decisions. |
| Password hashing | A custom hash or a "cheap" hash for the throwaway guest password | `bcrypt.GenerateFromPassword(..., 14)`, same cost factor as `hashPassword` | The guest password is a real secret stored in a real `password` column that `claimGuestAccount` and `Login`'s rejection logic both interact with; a weaker hash for guests would be a real vulnerability with no offsetting benefit (the value is random and never shown to a user regardless of hash strength). |
| Rate limiting | A new middleware, a database-backed limiter, or an ad-hoc counter | `pkg/ratelimit.Registry` + a new `Surface` constant | Purpose-built, already tested, already has bounded-growth (idle eviction) and per-surface isolation — exactly what guest creation needs. |
| MTG-flavored name generation | A generic "random name" npm/Go package, or Markov-chain generation | A small in-repo curated word list + retry-on-conflict against `username` | D-2.5 requires curation for tone (no offensive/absurd pairings) — an off-the-shelf generator cannot be pre-audited to that bar, and the domain (adjective+MTG-noun) is narrow enough that a curated list is genuinely less code than integrating and constraining a library. |
| Uniqueness enforcement for generated names | Application-level "check then insert" without a DB constraint | The existing `username_unique` constraint (`server/persistence/migrations/20211229130157_enforce_unique_usernames.up.sql`) + retry-on-`23505`-violation in Go, mirroring `Signup`'s existing `strings.Contains(err.Error(), "username_unique")` check [VERIFIED: server/users.go:31] | The DB constraint is the actual source of truth against concurrent guest creation; a Go-side pre-check has a TOCTOU race under concurrent guest sign-ups. |
| Commander-partner validity logic | New multi-commander/background rules in the new review component | `app/src/services/commanderPartner.ts`'s `canAddSecondCommander`/`isValidPartnerPair`/`partnerConstraintMessage` | Already covers the Partner-pairing rules Phase 1 already decided (up to 3 suggestions, always show nearest); reimplementing risks drifting from that decision. |
| XSS escaping for generated/typed display names | A manual HTML-escape function before rendering | Vue's default `{{ }}` text interpolation (never `v-html`) | Every current `Username` render site in `app/src` uses `{{ }}` interpolation, not `v-html` [VERIFIED: grep for `v-html.*[Uu]sername\|display` across `app/src/**/*.vue` returned zero matches this session] — Vue auto-escapes interpolated text, satisfying "correctly escaped" without new code, as long as the new `display_name` renders through the same mechanism. |

**Key insight:** This phase's backend primitives (auth, rate-limit, event-write) are 100%
already-built patterns from Phase 1 needing new *call sites*, not new *mechanisms*. The
actual net-new engineering is (1) guest-name generation/collision-retry, (2) the
`display_name` sweep across both GraphQL documents and Vue templates, and (3) the `/play`
router-guard branch and its `sessionStorage` draft persistence — none of which have any
existing in-repo precedent to copy.

## Runtime State Inventory

Not applicable — this is additive schema/feature work, not a rename/refactor/migration of
existing identifiers. No existing runtime state (stored data, live service config,
OS-registered state, secrets, build artifacts) is being renamed or relocated.

## Common Pitfalls

### Pitfall 1: The `display_name` sweep is silently incomplete
**What goes wrong:** A developer updates the four `.vue` files CONTEXT.md names "at
minimum" but the underlying GraphQL query/mutation documents still only select `Username`,
so `DisplayName` is always `undefined` and every fallback silently resolves to `username`
(the generated MTG name) — the feature appears to work in a naive manual test (no crash,
name renders) while never actually showing what a guest typed.
**Why it happens:** GraphQL is a request-response contract; a field not in the selection
set is not in the response, regardless of what the server schema or resolver support.
**How to avoid:** Treat the sweep as two lists, not one: (a) every `.vue` file rendering
`.Username` in a template — found this session at `AppNav.vue:31`,
`FormCreateGame.vue:264,268` (form submission, not display — lower priority but still an
identity site), `JoinGameView.vue:270`, `ScoreView.vue:10`, `BoardView.vue` (dozens of
sites, primarily `player.Username`/`selfPlayer.Username`/`me.Username` used both for
*display* — e.g. lines 66, 255 — and for *identity comparison/lookup* — the majority of
occurrences — only the display sites need the `??` fallback; identity-comparison sites
must keep using the real `username`), `GamesView.vue:64,106,114,145-146` (line 106/114 are
identity-matching, not display — leave as `username`), and `GameAnalysisView.vue:96,125,164,167`
(not named in CONTEXT.md's "at minimum" list but is a genuine `Username`-rendering surface
found this session — flag for the plan); and (b) every GraphQL document requesting
`Players { ... Username ... }` — found this session at `app/src/graphql/queries.ts`
(`GAMES_QUERY`, `GET_GAME_QUERY`, `GAME_UPDATED_SUBSCRIPTION`) and
`app/src/graphql/mutations.ts` (`CREATE_GAME_MUTATION`, `JOIN_GAME_MUTATION`) — all six of
which must add `DisplayName` to their `Players { ... }` selection set, and the server's
`schema.graphql` `type User` must gain a `DisplayName: String` field with a resolver
returning the new column.
**Warning signs:** A component test that mocks Apollo responses with hand-written fixture
objects will pass even with an incomplete GraphQL-document sweep, because the fixture
"response" was never actually constrained by the real query string — this is a case where
passing component tests can mask an incomplete real-network sweep; a plan should verify the
query *strings* directly (e.g., a small assertion or manual grep-in-CI step), not rely
solely on component tests.

### Pitfall 2: `player.Username === auth.profile?.Username` identity comparisons break if guest usernames aren't unique-namespace-safe
**What goes wrong:** `BoardView.vue` uses `Username` (not `ID`) as the join key between
`game.Players` and `auth.profile` in dozens of places (e.g. `selfPlayer` computed at line
704-705, `isMe` at line 1246, and every zone-interaction handler). If a generated guest name
ever collided with a real user's `username` — which D-2.6 explicitly forbids ("colliding
with neither") — these comparisons would silently attribute one player's actions to another.
**Why it happens:** The comparisons predate any guest concept and assume `Username` is a
safe, unique join key (which it always has been, enforced by `username_unique`).
**How to avoid:** D-2.6's "globally unique across all users — the same namespace as real
usernames" requirement is not cosmetic; it is a correctness precondition for this exact
comparison pattern throughout `BoardView.vue`. The retry-on-conflict generation logic must
retry against the *real* `username_unique` constraint (or an equivalent `SELECT ... WHERE
username = $1` pre-check backed by that constraint), not merely against an in-memory guest
name pool. **This phase's boundary ends before `/games/:id` behavior**, but the naming
scheme this phase ships is what keeps `BoardView.vue`'s pre-existing logic correct — get the
uniqueness contract right at the source, since fixing a namespace collision after real guest
history has accumulated is exactly the "one-way" cost D-2.1/D-2.6 warn about.

### Pitfall 3: The router guard has no "public but has real behavior" case today
**What goes wrong:** `/play` gets added with `meta: { public: true }` (copy-pasting
`/login`/`/signup`'s pattern) or is left with no meta at all and accidentally falls through
`requiresAuth` handling in a way that redirects logged-out users.
**Why it happens:** The existing `beforeEach` [VERIFIED: app/src/router/index.ts:79-88] has
exactly two branches: `to.meta.public` short-circuits to `next()` before even loading the
auth store, and `to.meta.requiresAuth && !auth.isAuthenticated` redirects to `/login`.
Every route today (`app/src/router/index.ts:4-72`) is one or the other (or neither, for
`/`, `/card/:id`, `/games/404`, the catch-all) — none combine "always accessible" with
"the component itself branches its own behavior on `auth.isAuthenticated`," which is
exactly what `/play` needs (criterion 1: "never redirected... an existing authenticated
host travels the same road").
**How to avoid:** `meta.public: true` on `/play` is actually *correct* for the router-level
gate (it must never redirect, authenticated or not) — the pitfall is doing nothing further:
`QuickStartView.vue` itself must read `auth.isAuthenticated` and branch (skip guest
creation, reuse existing profile) since the router guard cannot express that distinction.
Do not invent a third `meta` flag unless the plan finds a concrete need beyond what
`meta.public` plus in-component branching already covers.

### Pitfall 4: Guest tokens as `edhgo/auth` risk being mistaken for the D-2.4 credential
**What goes wrong:** A guest's issued JWT (used for all authenticated GraphQL calls, stored
under the *existing* `edhgo/auth` key — `AuthProfile{ID, Username, Token}`) gets confused
with D-2.4's *new*, separate re-auth bearer credential. If the same value is reused for
both purposes, the "separate secret from the analytics session ID" requirement is
technically met but D-2.4's actual intent (a credential usable to silently mint a *new*
JWT after the old one expires, without re-running `guestSession` or losing row identity)
is not, since a bare JWT cannot authenticate itself once expired (`parseAndValidateToken`
[VERIFIED: server/auth.go:83-108] rejects an expired token outright — `jwt.ParseWithClaims`
returns an error for an expired `ExpiresAt` claim, and there is no separate "refresh" claim
type in `AuthClaims` [VERIFIED: server/auth.go:21-24, quoted: `type AuthClaims struct {
Username string \`json:"username"\`; jwt.RegisteredClaims }`] to distinguish a refresh token
from an access token).
**Why it happens:** The codebase has exactly one token concept (`Token` in `AuthProfile`,
used for both `Authorization: Bearer` and WS `connectionParams`), and D-2.4 requires
introducing a second, structurally different credential for the first time.
**How to avoid:** The guest's JWT (`Token`) continues to live under `edhgo/auth` and is
used exactly as today for every authenticated GraphQL/WS call. D-2.4's new value is a
*different* secret — e.g. a long random string, stored under a new key (e.g.
`edhgo/guest-credential`), sent *only* to whatever new re-issue mutation this phase (or a
follow-up) introduces, never attached as the `Authorization` header for ordinary requests.
See Open Questions for the fact that no such re-issue mutation exists in the current
schema/requirements text.

## Code Examples

### Guest token issuance (extends existing pattern)
```go
// Source: server/auth.go:110-129 (existing newAuthToken, reused verbatim for guests)
func newAuthToken(user *User, ttl time.Duration) (string, error) {
	if user == nil || user.ID == "" {
		return "", errors.New("invalid user for token")
	}
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := AuthClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
// A guest call site is simply: newAuthToken(guestUser, time.Hour*24)
```

### Signup's collision-handling pattern (model for guest name retry, D-2.6)
```go
// Source: server/users.go:24-35
stmt := `
INSERT INTO "users" (uuid, username, password)
VALUES ($1, $2, $3)
RETURNING uuid, username;
`
result, err := s.db.Query(stmt, id, username, hashed)
if err != nil {
	if strings.Contains(err.Error(), "username_unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate key value") {
		return nil, errs.New("That username is already taken. Try another one.")
	}
	return nil, errs.Wrap(err)
}
// For guest generation: on this same error class, pick a new adjective-noun pair
// (or append/increment a numeric suffix per D-2.5's overflow strategy) and retry
// the INSERT, bounded by a max-attempts constant, rather than surfacing an error
// to the visitor — D-2.7's "no 'that name is taken' failure may appear on the
// activation path" applies to generated names too (the visitor never typed one).
```

### `/play` route addition (extends existing router shape)
```typescript
// Source: app/src/router/index.ts:4-88 (existing shape) — new entry to add:
{
  path: '/play',
  name: 'quick-start',
  component: () => import('../views/QuickStartView.vue'),
  meta: { public: true }, // never requiresAuth — QuickStartView branches internally
                           // on auth.isAuthenticated, matching criterion 1's
                           // "never redirected through /login or /signup"
},
```

### Rate-limit + kill-switch call shape for the new resolver
```go
// Modeled on server/product_events.go:151-159's TrackProductEvent call shape,
// and server/deck_providers.go:337-339's providerEnabled() predicate:
func (s *graphQLServer) GuestSession(ctx context.Context, displayName *string, sessionID string) (*User, error) {
	if !s.guestCreationEnabled() {
		return nil, errGuestCreationDisabled // maps to a D-2.13 error code client-side
	}
	clientKey := clientKeyFor(ctx, sessionID)
	if !s.allowRequest(ctx, ratelimit.SurfaceGuestSession, clientKey) {
		return nil, errRateLimited // maps to a D-2.13 error code client-side
	}
	// ... generate name, hash random password, insert row, issue token ...
	s.recordProductEvent(ctx, ProductEvent{
		Name:      "guest_session_created",
		SessionID: sessionID,
		UserID:    &guestUser.ID,
	}, false)
	return guestUser, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| Deck entry as raw CSV textarea, duplicated in `FormCreateGame.vue` and `JoinGameView.vue` | Six-syntax canonical parser behind `previewDeck` (`DeckImportPanel.vue`, this phase) | Phase 1 (ACT-002) built the parser; this phase (ACT-004) is the **first** frontend consumer — `grep -rn "previewDeck\|PREVIEW_DECK" app/src` returns zero matches today [VERIFIED: search run this session] | Both textareas' placeholder text ("CSV: quantity,name per line") is already stale advice — Phase 1's parser accepts far more, but nothing in the UI has told users that yet. |
| Single `Username` identity + display field | `username` (unique, system identity) / `display_name` (nullable, free-form label) split | This phase (D-2.8) | First schema change to `users` since `username_unique` (2021-12-29) — a real, one-way migration. |
| No responsive layer (`grep -rn "@media" app/src` returns zero matches [VERIFIED: search run this session]) | First `@media (min-width: 768px)` breakpoint, scoped to new activation components | This phase (D-2.10) | Sets precedent Phase 4's ACT-011 (board responsiveness) will read as the established shape — worth a short comment at the first `@media` site explaining the 768px choice for that future reader. |

**Deprecated/outdated:**
- The literal placeholder text "Decklist (CSV: quantity,name per line)" in both
  `FormCreateGame.vue:76` and `JoinGameView.vue:31` — both call sites this phase replaces.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | A new `GUEST_CREATION_ENABLED` envconfig name (mirroring `DECK_PROVIDER_ENABLED`) is the right name/shape for the guest-creation kill switch. | Rate Limiting + Kill Switch | Low — this is a new symbol the plan is free to name; no existing contract to break. Flagged only because REQ-ACT-005's text says "feature flag/kill switch" without naming it. |
| A2 | A hashed `guest_credential_hash` column on `users` (rather than a separate table) is the right storage shape for D-2.4's re-auth credential. | Migration Shape, Pitfall 4 | Medium — if the plan instead wants credential rotation/revocation history, a separate table would be needed; a single column cannot represent "the last N credentials." Given REQ-ACT-005's scope text says nothing about rotation history, a single column matches the stated scope, but this is inference, not a locked decision. |
| A3 | The re-auth/silent-reissue mechanism (D-2.3/D-2.4) requires a **new mutation** not named anywhere in REQ-ACT-005's ticket text or `likely_files`. | Open Questions, Pitfall 4 | High if wrong in the other direction — if the planner instead assumes `guestSession` itself can be called again idempotently to "refresh," that contradicts D-2.6 (names never recycled, so calling `guestSession` again for an existing row would either generate a *second* name for the same row or require passing the existing row's identity in a way `guestSession(displayName, sessionID)`'s declared signature does not support). This is flagged as an open question, not asserted as fact, precisely because no source text confirms which shape is intended. |
| A4 | Vue's default `{{ }}` interpolation escaping is sufficient to satisfy D-2.5/REQ-ACT-005's "pass existing output escaping/validation" acceptance clause, with no additional server- or client-side sanitization needed for `display_name`. | Don't Hand-Roll (XSS row) | Medium — true for HTML/XSS in the Vue template layer (verified: no `v-html` usage found), but does not address non-HTML injection surfaces this research did not check exhaustively (e.g., whether `display_name` ever flows into a non-Vue-rendered surface such as a server log line, a GraphQL error message, or an exported CSV/report). No such surface was found this session, but the search was not exhaustive of every possible future consumer. |

## Open Questions (RESOLVED)

All three questions below were resolved during Phase 2 planning; each carries an inline
`**RESOLVED (Phase 2 planning):**` marker naming the plan and task that settles it. Nothing in
this section remains open.

1. **What is the actual transport/mutation for D-2.3's silent token re-issue?**
   - What we know: D-2.3 requires "a silent re-issue while their row is alive" whenever a
     guest's 24-hour token expires; D-2.4 requires the re-auth value to be a distinct
     bearer credential under its own localStorage key, sent only to "the auth path."
     `guestSession(displayName, sessionID)` — the only guest-creation mutation named in
     REQ-ACT-005 — takes no parameter that could identify an *existing* row to re-issue a
     token for, and calling it again would (per D-2.6) either mint a second identity or
     require a signature it doesn't have.
   - What's unclear: Whether the plan should (a) add a new mutation (e.g.
     `refreshGuestToken(credential: String!)`), (b) overload `guestSession` with an
     optional credential parameter, or (c) treat this as strictly out of scope for Phase 2
     and defer the actual re-issue wiring to Phase 4 alongside `claimGuestAccount`'s UI
     (REQ-A5's "backend only" carve-out could arguably extend here, since nothing in the
     board/session-continuity UI that would *trigger* a re-issue exists until Phase 3/4
     anyway).
   - Recommendation: The plan should explicitly decide and record this — building the
     mutation now (cheap, symmetrical with `guestSession`) versus deferring it (consistent
     with REQ-A5's precedent of shipping a backend contract ahead of its UI) are both
     defensible, but CONTEXT.md's D-2.3/D-2.4 read as in-scope decisions for *this* phase,
     not deferred ones — the Deferred Ideas section names only claim-UI, board
     responsiveness, and rename as out of scope, not re-issue.
   - **RESOLVED (Phase 2 planning):** option (a) — build the mutation now, in this phase.
     Plan `02-02` Task 2 adds `refreshGuestSession(credential: String!): User!` as a new
     Mutation field and implements `(*graphQLServer).RefreshGuestSession` in
     `server/guest_users.go`, deliberately outside `requireAuth` so it works after the JWT has
     expired. `guestSession` is NOT overloaded (that would contradict D-2.6), and nothing is
     deferred to Phase 4. The credential wire format is `<users.uuid>.<base64url 32-byte
     secret>` — minted in plan `02-01` Task 1, with only the secret half bcrypt-hashed into
     `guest_credential_hash` so the uuid half gives the refresh path an O(1) row lookup. Per
     DEC-F in `02-02`, the credential is not rotated on re-issue.

2. **Should `AuthUser` (the JWT-derived context object) carry an `IsGuest` bit, or should
   every guest-aware resolver re-query `users.is_guest` fresh?**
   - What we know: `AuthUser` today is `{ID, Username}` only [VERIFIED: server/auth.go:16-19].
     Adding `IsGuest` to the JWT claims would let `requireAuth` expose it cheaply, but
     baking guest-status into a 24-hour-lived JWT claim means the claim goes stale exactly
     at claim time (`claimGuestAccount` must flip `is_guest` to false, but an
     already-issued JWT's claim would still say `true` until the next token issuance).
   - What's unclear: Whether any resolver in this phase actually needs `IsGuest` at the
     `AuthUser` level, or whether `claimGuestAccount` (the only place it plausibly matters
     this phase) can simply re-query the row by `AuthUser.ID` directly.
   - Recommendation: Default to re-querying fresh inside `claimGuestAccount` (matches the
     "verify the *current* row is still a guest at claim time" note in Guest/Password-Login
     Auth Surface above) and avoid adding `IsGuest` to JWT claims unless a concrete
     resolver-level need surfaces during planning.
   - **RESOLVED (Phase 2 planning):** re-query fresh; `AuthUser` gains no `IsGuest` bit and the
     JWT claims are unchanged. Plan `02-02` Task 3's `ClaimGuestAccount` runs behind
     `requireAuth` and then re-reads the row by `AuthUser.ID`, because the JWT's claims were
     minted before the claim and cannot be trusted to say whether the row is still a guest; the
     UPDATE additionally carries a `WHERE is_guest` guard so a concurrent double-claim cannot
     both win. The one other guest-aware check, `Login`'s rejection in `02-02` Task 1, also
     reads `is_guest` from the row rather than from a claim.

3. **How aggressively should `FormCreateGame.vue`/`JoinGameView.vue` be refactored to share
   `DeckImportPanel.vue`, given both are also embedded in modal/page contexts with
   surrounding form chrome (game name, deck size, format select) that `/play` doesn't have?**
   - What we know: REQ-ACT-004 requires "keep the current authenticated create flow usable
     during rollout" and "host and join submit the same normalized deck representation."
     CONTEXT.md marks the exact aggressiveness as Claude's Discretion.
   - What's unclear: Whether `DeckImportPanel.vue` should be a fully generic, chrome-free
     component that all three call sites (`FormCreateGame.vue`, `JoinGameView.vue`,
     `QuickStartView.vue`) wrap differently, or whether it needs slot-based customization
     for the surrounding fields each context uniquely needs (game name/deck size only exist
     in `FormCreateGame.vue`; nothing in `JoinGameView.vue` or `QuickStartView.vue` has
     those fields).
   - Recommendation: Design `DeckImportPanel.vue` to own only text/URL input, preview,
     and unresolved-card correction (never game name/deck size/format), matching exactly
     what `previewDeck`'s `InputDeckImport` accepts (`text`, `sourceURL`, `sessionID`
     [VERIFIED: server/schema.graphql — `input InputDeckImport { text: String sourceURL:
     String sessionID: String! }`]) — this naturally bounds the component's responsibility
     and avoids the slot-customization question entirely.
   - **RESOLVED (Phase 2 planning):** the recommendation was adopted verbatim and recorded as
     DEC-J in plan `02-03`: `DeckImportPanel.vue` takes no slots and owns no surrounding chrome;
     its props are exactly what `InputDeckImport` accepts plus a persistence key, and game name,
     deck size and format stay with whichever parent owns them. `02-03` Task 1 extracts the
     panel and `02-03` Task 3 re-points all three call sites at it — `FormCreateGame.vue` and
     `JoinGameView.vue` pass no persistence key, `QuickStartView.vue` does. The
     slot-customization question is therefore closed, not deferred.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Server resolvers, migrations | ✓ | go1.25.5 (module targets `go 1.24.0`) [VERIFIED: go.mod:4] | — |
| Node.js / npm | Frontend build/test | ✓ | Node v23.5.0, npm 11.7.0 | — |
| PostgreSQL | Migrations, resolver tests | Not directly probed (`pg_isready` CLI not installed on this machine) — but `Makefile`'s `persistence` target starts it via `docker-compose -f dev.docker-compose.yml up -d postgres` [VERIFIED: Makefile] | — | Run `make persistence` before `go test ./server/...`; Docker itself confirmed available this session. |
| Docker | Local Postgres for tests | ✓ | `docker info` succeeded this session | — |
| `golang-migrate` CLI | `make migrate-local`/`make migrate-prod` (manual ops, not required for `go test`) | Not probed — `server/test.go` uses the in-process `persistence.ForceCleanMigrations`/`NewPostgres` (library, not CLI), so the CLI is not a test-time dependency | n/a | Tests do not need the CLI; only manual prod migration ops do. |

**Missing dependencies with no fallback:** none identified.

**Missing dependencies with fallback:** PostgreSQL must be started via `make persistence`
(Docker Compose) before running `go test ./server/...` locally — this is the existing
project convention, not new to this phase.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Backend framework | Go `testing` package, run via `go test ./server/... -race` (`make test-api`) [VERIFIED: Makefile] |
| Backend test DB | `persistence/migrations_test/` against a real Postgres (Docker), reset per `server/test.go:76-79` |
| Frontend framework | Vitest (`app/package.json`'s `"test": "vitest --run"`) with `@vue/test-utils`, jsdom environment |
| Frontend component tests | `app/__tests__/*.spec.ts` — existing examples: `FormCreateGame.integration.spec.ts`, `JoinGame.integration.spec.ts`, `commanderPartner.spec.ts`, `productEvents.spec.ts` [VERIFIED: `find app/__tests__` output this session] |
| E2E | Playwright (`app/e2e/create-and-join-game.spec.ts` exists today; REQ-ACT-013 (Phase 5) extends this suite for guest host/join — not this phase's job, but this phase's resolvers/routes are what Phase 5 will drive) |
| Quick run command (Go) | `go test ./server/... -run TestGuestUsers -race` (name pattern once new tests exist) |
| Quick run command (Vue) | `npx vitest run app/__tests__/QuickStartView.spec.ts` (once created) |
| Full suite command | `make test` (Go) and `npm test` (frontend, from `app/`) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REQ-ACT-004 | Paste → preview → correct → select commander → continue, state preserved across failures | component/integration | `npx vitest run app/__tests__/DeckImportPanel.spec.ts` | ❌ Wave 0/1 |
| REQ-ACT-004 | Host and join submit the same normalized deck | integration | `npx vitest run app/__tests__/FormCreateGame.integration.spec.ts app/__tests__/JoinGame.integration.spec.ts` | ✅ existing files, will need updates for the new shared component |
| REQ-ACT-005 | Guest creation requires no username/password; generated names unique & escaped | unit/integration (Go) | `go test ./server -run TestGuestUsers_Create -race` | ❌ Wave 0/1 (new `server/guest_users_test.go`) |
| REQ-ACT-005 | Guest cannot log in via password endpoint before claiming | unit (Go) | `go test ./server -run TestUsers_LoginRejectsGuest -race` | ❌ Wave 0/1 |
| REQ-ACT-005 | Cleanup never deletes a guest referenced by an active game | integration (Go) | `go test ./server -run TestGuestUsers_CleanupPreservesActiveGames -race` | ❌ Wave 0/1 |
| REQ-ACT-005 | Migration applies cleanly on both prod and test schemas | migration test | `go test ./server -run TestMigrations_GuestUsers -race` (model on `TestMigrations_CardNameSearch` at `server/deck_import_test.go:150`) | ❌ Wave 0/1 |
| REQ-ACT-006 | `/play` public while unrelated private routes remain protected | router unit test | `npx vitest run app/__tests__/router.spec.ts` (new or extend existing router coverage) | ❌ Wave 0/1 — no `router.spec.ts` found this session |
| REQ-ACT-006 | Guest and authenticated branches, error preservation | component/store test | `npx vitest run app/__tests__/QuickStartView.spec.ts` | ❌ Wave 0/1 |
| Criterion 5 | Retried client call does not double-count a conversion | integration (Go) | `go test ./server -run TestProductEvents_AuthoritativeDedup -race` (already exists — extend or add a guest-specific case; model at `server/product_events_test.go:128`) | ✅ existing test, may need a new case |

### Sampling Rate
- **Per task commit:** run the narrowest `go test -run` / `vitest run <file>` for the
  touched area.
- **Per wave merge:** `make test-api` (Go) and `npm test` (frontend, from `app/`), both full
  suites, plus `npm run type-check` (the generated GraphQL types must stay in sync after the
  schema changes).
- **Phase gate:** all of the above green, plus a manual check that `previewDeck` is
  actually reachable from `/play` in a dev-server smoke run (no Playwright coverage for
  this exists until Phase 5's REQ-ACT-013).

### Wave 0 Gaps
- [ ] `server/guest_users_test.go` — covers REQ-ACT-005's create/collision/rate-limit/expiry/
      login-rejection/claim/conflict/relationship-preservation acceptance clauses.
- [ ] A migration test for the new `{TS}_guest_users` pair, modeled on the existing
      `TestMigrations_CardNameSearch` pattern (`server/deck_import_test.go:150`).
- [ ] `app/__tests__/DeckImportPanel.spec.ts`, `app/__tests__/CommanderReview.spec.ts` —
      no existing component tests for the not-yet-built components.
- [ ] `app/__tests__/QuickStartView.spec.ts` — no existing router/store integration test
      exercising a public-but-stateful route branch (guest vs. authenticated).
- [ ] A router-level test asserting `/play` never redirects regardless of `auth.isAuthenticated`
      — no `router.spec.ts` was found in `app/__tests__` this session.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | Yes | Reuse existing JWT (`golang-jwt/jwt/v5`) + bcrypt (cost 14) exactly as `Signup`/`Login` do; no new auth primitive introduced. |
| V3 Session Management | Yes | 24-hour JWT TTL unchanged from existing tokens; new D-2.4 re-auth credential must be treated as a session-management secret (random, high-entropy, hashed at rest — mirror the password-hash pattern, never store it in plaintext even server-side). |
| V4 Access Control | Yes | `requireAuth`/`requireMatchingUser` unchanged; guest rows pass through identical authorization as real users by design (that's the point of the funnel) — the only new check is the Login-endpoint `is_guest` rejection (defense-in-depth, not the primary control, since the guest password is unguessable regardless). |
| V5 Input Validation | Yes | Display name: length-bound (mirror `username VARCHAR(255)`), no server-side HTML sanitization needed given Vue's escaping (see Don't Hand-Roll), but the server should still reject/trim control characters and enforce a max length independent of the client, since GraphQL's `String` type has no built-in length constraint. |
| V6 Cryptography | Yes | Guest's random password uses `bcrypt.GenerateFromPassword` (existing); the new re-auth credential should be generated via `crypto/rand` (Go) — never `math/rand` (note `backronym.go` uses `math/rand`, but that file is unrelated dead-utility code, not a pattern to follow for anything security-relevant) — and stored as a hash (e.g. SHA-256 or bcrypt), never plaintext, mirroring the password column's own treatment. |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Guest-row enumeration / mass guest creation (funnel abuse, resource exhaustion) | Denial of Service | `SurfaceGuestSession` rate limiting via `pkg/ratelimit`, address-anchored `clientKeyFor` (already resists sessionID rotation — see Rate Limiting section) |
| Guest impersonation via credential reuse across devices | Spoofing | D-2.4's re-auth credential must be unguessable (`crypto/rand`, sufficient entropy — mirror bcrypt cost reasoning) and never logged; the JWT itself already carries this risk today for real users identically, so no new exposure beyond what exists. |
| Cross-namespace username collision (guest name equals a real username, or two claim attempts race for the same prefill) | Tampering / Repudiation | DB-level `username_unique` constraint is the actual enforcement (not application logic); `claimGuestAccount`'s prefill (D-2.9) must go through the identical collision-handling path as `Signup` — see the "Signup's collision-handling pattern" code example above. |
| Stored/reflected XSS via `display_name` (free-form, user-typed, per D-2.7) | Tampering | Vue's default `{{ }}` interpolation (never `v-html`) for every render site — see Don't Hand-Roll. Server-side, still cap length and strip control characters as defense-in-depth (V5), since a value destined for non-Vue consumers (logs, future exports) is not protected by Vue's escaping. |
| Guest token used against the password `Login` mutation before claiming | Elevation of Privilege | Add explicit `is_guest` check in `Login` (see Guest/Password-Login Auth Surface) — defense-in-depth since the random bcrypt-hashed password is already unguessable, but REQ-ACT-005 names this as an explicit acceptance clause, so it must be tested directly, not merely relied upon implicitly. |

## Sources

### Primary (HIGH confidence — read directly this session)
- `server/auth.go`, `server/authz.go`, `server/users.go`, `server/ratelimit.go`,
  `server/product_events.go`, `server/deck_providers.go` (partial), `server/games.go`
  (JoinGame/CreateGame), `server/graphql.go` (partial), `server/schema.graphql`,
  `server/backronym.go`, `server/resolver.go`, `server/schema.resolvers.go` (partial),
  `server/formats.go` (partial)
- `pkg/ratelimit/limiter.go`, `pkg/telemetry/metrics.go` (partial), `pkg/telemetry/vocabulary.go` (partial)
- `persistence/sql.go` (partial), `persistence/migrations/*` (init_db, enforce_unique_usernames,
  product_events), `persistence/migrations_test/README.md`, directory listings of both
  migration directories
- `app/src/router/index.ts`, `app/src/stores/auth.ts`, `app/src/stores/games.ts`,
  `app/src/services/apollo.ts`, `app/src/services/productEvents.ts`,
  `app/src/services/commanderPartner.ts`, `app/src/components/games/FormCreateGame.vue`,
  `app/src/views/JoinGameView.vue`, `app/src/views/GamesView.vue` (partial),
  `app/src/views/ScoreView.vue` (partial), `app/src/views/LandingView.vue` (partial),
  `app/src/components/layout/AppNav.vue`, `app/src/graphql/queries.ts` (partial),
  `app/src/graphql/mutations.ts` (partial), `app/src/styles/main.scss`
- `docs/analytics/product-event-vocabulary.md`
- `.planning/phases/02-guest-host-activation/02-CONTEXT.md`, `.planning/REQUIREMENTS.md`,
  `.planning/STATE.md`, `.planning/intel/requirements.md` (REQ-ACT-004/005/006 sections)

### Secondary (MEDIUM confidence)
- None — no external documentation lookups were needed; every claim in this document is
  grounded in the local codebase or the project's own planning artifacts.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new packages; every reused primitive read directly this session.
- Architecture: HIGH — resolver wiring, router guard, rate-limit/kill-switch shape, and
  migration-sync mechanism all directly verified against source.
- Pitfalls: HIGH for the display_name sweep and router-guard gap (both directly observed
  gaps in current code); MEDIUM for the re-auth-credential open question, since no source
  text resolves it definitively (flagged honestly in Open Questions/Assumptions rather than
  asserted).

**Research date:** 2026-08-08
**Valid until:** 30 days (stable, internally-controlled codebase; no external API/version
drift risk since no new external dependencies are introduced)

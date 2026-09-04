# vEDH Deck-to-Game Activation Implementation Tickets

**Status:** Ready for estimation
**Date:** 2026-07-23
**PRD:** `docs/product/2026-07-23-deck-to-game-activation-prd.md`
**Code baseline:** `main` at `ef2732a`
**Delivery assumption:** One implementation agent, approximately four weeks
**Priority rule:** P0 establishes measurable guest deck-to-board activation. P1 improves post-value acquisition and operating confidence.

## Delivery Map

| Order | Ticket | Priority | Size | Depends on | Outcome |
|---|---|---:|---:|---|---|
| 1 | ACT-001 Product-event and activation-metric foundation | P0 | M | — | Funnel and technical outcomes are measurable |
| 2 | ACT-002 Canonical deck parser and preview API | P0 | L | ACT-001 | One safe import path serves create and join |
| 3 | ACT-003 Public deck-provider feasibility gate and first adapter | P0 | M | ACT-002 | At least one familiar deck URL works safely |
| 4 | ACT-004 Reusable deck-import and commander-review UI | P0 | L | ACT-002 | Host and join use the same preview experience |
| 5 | ACT-005 Guest identity and account-claim backend | P0 | L | ACT-001 | Authorized game mutations work without signup |
| 6 | ACT-006 Public quick-start host flow | P0 | L | ACT-004, ACT-005 | Guest host reaches a live board |
| 7 | ACT-007 Safe public invite preview | P0 | M | ACT-001 | Invitees understand a table before joining |
| 8 | ACT-008 Guest invite-to-board flow | P0 | L | ACT-004, ACT-005, ACT-007 | Invitee joins without signup |
| 9 | ACT-009 Board readiness, reconnect, and invite sharing | P0 | M | ACT-006, ACT-008 | Activation is trustworthy and pod growth is immediate |
| 10 | ACT-010 Post-value account claim UI | P1 | M | ACT-005, ACT-009 | Activated guests become durable users |
| 11 | ACT-011 Commander landing page and attribution | P1 | M | ACT-001, ACT-006 | Acquisition feeds the measured activation flow |
| 12 | ACT-012 Activation dashboard and quiet-beta runbook | P1 | S | ACT-001, ACT-009 | Product decisions use one funnel definition |
| 13 | ACT-013 End-to-end release gate and regression suite | P0 | L | ACT-003 through ACT-010 | The complete funnel is safe to ship |

Sizes are relative: S is up to one focused day, M is two to three days, and L is three to five days including tests and review.

## ACT-001 — Product-Event and Activation-Metric Foundation

**Priority:** P0
**Size:** M
**Dependencies:** None

### Outcome

Every activation step has a privacy-safe event definition, while the Prometheus/Grafana stack on the latest `main` branch reports technical health.

### Scope

- Add production and test migrations for a `product_events` table.
- Store event name, random session ID, optional user/game IDs, role, source, outcome, duration, approved metadata, and timestamp.
- Add `trackProductEvent` to `server/schema.graphql` and implement strict server-side event and field allowlists.
- Make `game_created`, `player_joined`, `guest_session_created`, `deck_import_succeeded`, `deck_import_failed`, and `account_claimed` server-authoritative.
- Add counters and latency histograms for guest sessions, imports, create/join results, and board activation.
- Add a frontend event service that persists a random session ID and only accepts approved attribution fields.
- Document the canonical event vocabulary and example funnel query.

### Likely Files

- `persistence/migrations/*_product_events.up.sql`
- `persistence/migrations/*_product_events.down.sql`
- `persistence/migrations_test/*_product_events.up.sql`
- `persistence/migrations_test/*_product_events.down.sql`
- `server/schema.graphql`
- `server/schema.resolvers.go`
- `server/graphql.go`
- `server/product_events.go`
- `server/product_events_test.go`
- `app/src/services/productEvents.ts`

### Acceptance Criteria

- Unknown event names and metadata keys are rejected.
- Raw deck text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game state cannot be stored through the event API.
- The server attaches authenticated user IDs rather than accepting them from the client.
- Events can be grouped by day, role, source, and outcome.
- Prometheus metrics contain no username, session ID, user ID, or game ID labels.
- Duplicate client retries do not duplicate authoritative conversion events.

### Verification

- Migration up/down tests pass for production and test schemas.
- Resolver tests cover allowed, rejected, oversized, and unauthenticated events.
- Metric tests verify counter and histogram changes without high-cardinality labels.
- Frontend unit tests verify stable session ID and attribution allowlisting.

## ACT-002 — Canonical Deck Parser and Preview API

**Priority:** P0
**Size:** L
**Dependencies:** ACT-001

### Outcome

Hosts and invitees can preview familiar pasted decklists without silent card loss or CSV-specific failures.

### Scope

- Extract parsing from `createLibraryFromDecklist` in `server/games.go` into a dedicated deck-import service.
- Support quantity/name text, `1x` text, spaced CSV, quoted CSV, section headers, blank lines, and known sideboard/maybeboard sections.
- Preserve card names containing commas.
- Return normalized entries, commander candidates, unresolved entries, warnings, blocking errors, and `CanContinue`.
- Add `previewDeck` and the `DeckPreview` GraphQL types.
- Pass a preview/import reference or normalized entries into create/join so the deck is not parsed twice with different rules.
- Preserve the current card lookup, selected-commander removal, and maximum deck-size rules.

### Likely Files

- `server/deck_import.go`
- `server/deck_import_test.go`
- `server/schema.graphql`
- `server/schema.resolvers.go`
- `server/games.go`
- `test/decklists/*`

### Acceptance Criteria

- The accepted examples in the PRD parse exactly as displayed.
- `1 Atraxa, Praetors' Voice` and `1,"Atraxa, Praetors' Voice"` resolve to the same card name.
- No nonblank input row disappears without becoming an entry, warning, or error.
- Unknown cards are shown as unresolved and are never silently substituted.
- Preview and final library creation consume the same normalized result.
- Existing authenticated create/join behavior remains compatible.

### Verification

- Table-driven Go tests cover all supported syntaxes and malformed input.
- Golden fixtures cover common Moxfield, Archidekt, and generic text exports without relying on their live services.
- Game tests prove commander removal and deck limits remain intact.
- GraphQL tests cover a successful preview and structured unresolved-card response.

## ACT-003 — Public Deck-Provider Feasibility Gate and First Adapter

**Priority:** P0
**Size:** M
**Dependencies:** ACT-002

### Outcome

The MVP supports one familiar public deck URL without making activation dependent on an unstable third party.

### Scope

- Time-box a one-day feasibility comparison of public Archidekt and Moxfield deck access.
- Record request shape, public/private behavior, rate-limit signals, response stability, and terms/operational risk.
- Select the first provider only if a stable public read path and fixture can be maintained.
- Implement an adapter interface keyed by an allowlisted hostname.
- Enforce HTTPS, DNS/IP validation, redirect revalidation, a 3-second connection timeout, 8-second total timeout, and a 1 MiB response cap.
- Normalize provider output through ACT-002.
- Protect the adapter with a server-side feature flag/kill switch.

### Likely Files

- `docs/research/deck-provider-feasibility.md`
- `server/deck_providers.go`
- `server/deck_providers_test.go`
- `server/deck_provider_<selected>.go`
- `server/testdata/deck_providers/*`
- `server/deck_import.go`

### Acceptance Criteria

- The decision record identifies the selected provider or explicitly escalates if neither clears the gate.
- Only explicitly allowlisted hostnames can be fetched.
- Redirects, DNS rebinding, private/reserved/link-local addresses, oversized responses, and timeout paths fail closed.
- Public provider failure returns a normalized error and a paste-text fallback; it never blocks pasted import.
- Private decks are reported as unsupported without requesting third-party credentials.
- Fixture contract tests detect provider response-shape changes.

### Verification

- Unit tests cover allowlist, redirect, DNS/IP, timeout, size, private-deck, invalid-response, and success cases.
- No test depends on a live provider.
- A staging smoke check imports one public fixture deck through the selected provider before release.

## ACT-004 — Reusable Deck-Import and Commander-Review UI

**Priority:** P0
**Size:** L
**Dependencies:** ACT-002

### Outcome

Create and join share one understandable deck-review experience that preserves user work when something fails.

### Scope

- Create a reusable `DeckImportPanel.vue`.
- Add text/URL input, import source detection, loading/error states, totals, warnings, and unresolved-card correction.
- Create a reusable commander-review component driven by `CommanderCandidates`.
- Preserve pasted input, corrected entries, commander choices, and display name across recoverable failures.
- Replace duplicated deck entry in `FormCreateGame.vue` and `JoinGameView.vue`.
- Keep the current authenticated create flow usable during rollout.

### Likely Files

- `app/src/components/decks/DeckImportPanel.vue`
- `app/src/components/decks/CommanderReview.vue`
- `app/src/components/games/FormCreateGame.vue`
- `app/src/views/JoinGameView.vue`
- `app/src/graphql/mutations.ts`
- `app/src/types/generated.ts`
- `app/__tests__/*`

### Acceptance Criteria

- A user can paste, preview, correct unresolved cards, select commanders, and continue without re-entering the deck.
- URL-provider failure offers text paste in the same context.
- Blocking errors are distinct from warnings.
- Keyboard and screen-reader users can reach inputs, issues, commander controls, and the continue action in a logical order.
- Host and join submit the same normalized deck representation.
- Existing authenticated creation remains available until the guest funnel passes its launch gate.

### Verification

- Component tests cover successful paste, comma-containing names, unresolved correction, provider fallback, state preservation, and commander selection.
- A generated type check passes after the schema update.
- Manual viewport checks cover desktop and the minimum supported tablet width selected for beta.

## ACT-005 — Guest Identity and Account-Claim Backend

**Priority:** P0
**Size:** L
**Dependencies:** ACT-001

### Outcome

New players can satisfy existing authorization rules without signup and later convert without losing ownership.

### Scope

- Add `is_guest` and `expires_at` to production and test user schemas.
- Add `guestSession(displayName, sessionID)` and `claimGuestAccount(username, password, sessionID)` mutations.
- Generate collision-safe readable guest names when the display name is empty.
- Create a random non-recoverable password hash for guests and issue a 24-hour guest token.
- Reject expired backing guests during authorization.
- Claim the existing user row, clear expiry, set credentials through current bcrypt rules, and issue a new full token.
- Add rate limits, feature flag/kill switch, and an independently retryable cleanup path for unreferenced expired guests.

### Likely Files

- `persistence/migrations/*_guest_users.up.sql`
- `persistence/migrations/*_guest_users.down.sql`
- `persistence/migrations_test/*_guest_users.up.sql`
- `persistence/migrations_test/*_guest_users.down.sql`
- `server/schema.graphql`
- `server/users.go`
- `server/auth.go`
- `server/authz.go`
- `server/guest_users.go`
- `server/users_test.go`

### Acceptance Criteria

- Guest creation does not require a username or password.
- Generated display names are unique and pass existing output escaping/validation.
- Guest tokens stop working after token or backing-user expiry.
- Guest users cannot log in through the password endpoint before claiming.
- Claiming preserves the same user UUID and all game/player relationships.
- Username conflicts and password validation return recoverable typed errors.
- Cleanup never deletes a guest referenced by an active game.

### Verification

- Migration tests pass.
- Resolver and authorization tests cover create, collision, rate limit, expiry, login rejection, claim, conflict, and relationship preservation.
- Security tests prove a guest cannot claim a different guest identity.

## ACT-006 — Public Quick-Start Host Flow

**Priority:** P0
**Size:** L
**Dependencies:** ACT-004, ACT-005

### Outcome

A new Commander host can move from deck to subscribed board without visiting login or signup.

### Scope

- Add public route `/play` and `QuickStartView.vue`.
- Present deck import first, followed by commander review and optional display name.
- Reuse the current authenticated user when present; otherwise create a guest session only when the deck can continue.
- Submit `Handle` and `FormatID` correctly along with the normalized deck.
- Call the existing `createGame` mutation and route to `/games/:id`.
- Emit the host activation events in the PRD.
- Add a visible recovery path for preview, guest-session, and create failures.

### Likely Files

- `app/src/router/index.ts`
- `app/src/views/QuickStartView.vue`
- `app/src/stores/auth.ts`
- `app/src/stores/games.ts`
- `app/src/graphql/mutations.ts`
- `app/src/components/decks/*`

### Acceptance Criteria

- `/play` is usable while logged out.
- A valid deck plus optional display name reaches the correct game board without an auth redirect.
- An authenticated user follows the same flow without creating a guest.
- The requested Commander format is sent to the backend.
- Recoverable failures preserve the deck and commander selections.
- `game_created` is server-authoritative; the client records start and readiness events only.

### Verification

- Router tests prove `/play` is public while unrelated private routes remain protected.
- Store/component tests cover guest and authenticated branches and error preservation.
- Playwright covers a logged-out paste-to-board happy path.

## ACT-007 — Safe Public Invite Preview

**Priority:** P0
**Size:** M
**Dependencies:** ACT-001

### Outcome

An invitee can confirm the table is real and joinable before providing a deck or identity.

### Scope

- Add `gameInvite(gameID)` with a deliberately separate minimal response type.
- Return format, status, display names, player count, capacity, and creation time only.
- Make `/join/:id` public and load preview data before any guest creation.
- Show clear states for nonexistent, finished, and full games.
- Rate-limit public invite lookup and instrument outcomes.

### Likely Files

- `server/schema.graphql`
- `server/schema.resolvers.go`
- `server/games.go`
- `server/games_test.go`
- `app/src/router/index.ts`
- `app/src/views/JoinGameView.vue`
- `app/src/graphql/queries.ts`
- `app/src/stores/games.ts`

### Acceptance Criteria

- A logged-out visitor can view a valid invite.
- The public query cannot return decklists, libraries, hands, tokens, game logs, complete board state, credentials, or hidden zones.
- Nonexistent, full, and finished states are understandable before deck input.
- The existing participant-only `getGame` authorization remains unchanged.
- Invite views record session, game, and approved attribution only.

### Verification

- Resolver tests assert the exact public response and forbidden-field absence.
- Authorization regression tests prove `getGame` remains participant-only.
- Router/component tests cover all invite states.

## ACT-008 — Guest Invite-to-Board Flow

**Priority:** P0
**Size:** L
**Dependencies:** ACT-004, ACT-005, ACT-007

### Outcome

An invited player joins a valid table with a deck and reaches the board without signup.

### Scope

- Add deck import, commander review, and optional display name below the safe invite preview.
- Reuse an existing authenticated identity or create a guest at the last responsible moment.
- Call the existing `joinGame` mutation with the canonical deck representation.
- Preserve context on race conditions such as a table becoming full.
- Route directly to the game board and emit invite/join activation events.

### Likely Files

- `app/src/views/JoinGameView.vue`
- `app/src/stores/auth.ts`
- `app/src/stores/games.ts`
- `app/src/graphql/mutations.ts`
- `app/src/components/decks/*`

### Acceptance Criteria

- A logged-out invitee joins without visiting login/signup.
- Existing authenticated users can join through the same page.
- Deck, commander, and display-name state survive a recoverable join failure.
- If the game becomes full or finishes, the page reports the new state without leaking game details.
- `player_joined` is written by the server only after the join transaction succeeds.

### Verification

- Component/store tests cover guest, authenticated, full-race, finished-race, and retry paths.
- Playwright covers a host-created invite opened by a separate guest browser context.

## ACT-009 — Board Readiness, Reconnect, and Invite Sharing

**Priority:** P0
**Size:** M
**Dependencies:** ACT-006, ACT-008

### Outcome

Reaching the board means the game is actually usable, and the host can immediately bring in the pod.

### Scope

- Define explicit board-loading, ready, degraded, reconnecting, and failed states in the games store.
- Emit `board_ready` only when the game query succeeds, the current player's board exists, and the subscription is connected or documented degraded polling is active.
- Add reconnect action and bounded polling fallback where practical.
- Add a primary invite action to the board header.
- Prefer native share when available and fall back to clipboard/manual copy.
- Record only the share method, source, and game ID.

### Likely Files

- `app/src/views/BoardView.vue`
- `app/src/stores/games.ts`
- `app/src/services/apollo.ts`
- `app/src/components/layout/*`
- `app/__tests__/*`

### Acceptance Criteria

- Route navigation alone never emits `board_ready`.
- A disconnected subscription shows a clear reconnect state rather than a silently stale board.
- The copied URL is `${window.location.origin}/join/${gameID}`.
- Share/copy success is confirmed; failure exposes a selectable manual URL.
- Invite telemetry contains no clipboard content.

### Verification

- Store tests cover ready, reconnect, degraded, and terminal failure transitions.
- Component tests cover native share, clipboard fallback, and manual fallback.
- Playwright proves both guest players are visible before the activation flow is considered complete.

## ACT-010 — Post-Value Account Claim UI

**Priority:** P1
**Size:** M
**Dependencies:** ACT-005, ACT-009

### Outcome

Activated guests can save their identity and games without leaving the board.

### Scope

- Extend the auth store with guest detection and account claim.
- Show a non-blocking “Save your games” prompt only after `board_ready`.
- Collect username/password, submit the claim mutation, and replace the guest token atomically.
- Preserve the board and active subscriptions through success and validation errors.
- Allow dismissal without affecting gameplay.

### Likely Files

- `app/src/stores/auth.ts`
- `app/src/views/BoardView.vue`
- `app/src/components/auth/ClaimGuestAccount.vue`
- `app/src/graphql/mutations.ts`
- `app/__tests__/*`

### Acceptance Criteria

- The prompt never appears before value is delivered.
- Successful claim retains the same user/game access and replaces the token.
- Username and password errors are displayed inline without route changes.
- Dismissal is remembered for the current game/session.
- Claim events contain no credentials.

### Verification

- Store/component tests cover success, username conflict, password failure, dismissal, and token replacement.
- Playwright proves a claimed guest can refresh and remain in the current game.

## ACT-011 — Commander Landing Page and Attribution

**Priority:** P1
**Size:** M
**Dependencies:** ACT-001, ACT-006

### Outcome

The public promise matches the fast Commander experience and sends users directly into the measured flow.

### Scope

- Rewrite the landing hero around “paste a Commander deck and start a table.”
- Make `/play` the primary CTA and retain login as a secondary action.
- Add a compact three-step explanation and an invite/join path.
- Capture only allowlisted UTM/referrer values into the product-event session.
- Preserve campaign attribution through host activation and invite activation.
- Avoid marketing claims the MVP cannot yet support.

### Likely Files

- `app/src/views/LandingView.vue`
- `app/src/router/index.ts`
- `app/src/services/productEvents.ts`
- `app/src/styles/main.scss`

### Acceptance Criteria

- The primary CTA opens `/play` without auth.
- Copy is Commander-specific and accurately describes paste and the selected provider.
- Attribution is normalized, length-bounded, and never copied into Prometheus labels.
- Landing CTA, quick-start, and board-ready events share the same random session ID.
- Keyboard focus and responsive layout remain usable.

### Verification

- Component tests cover CTA destination and attribution allowlisting.
- A manual content/viewport check confirms no unsupported product claims.
- Staging telemetry proves one test campaign can be followed to `board_ready`.

## ACT-012 — Activation Dashboard and Quiet-Beta Runbook

**Priority:** P1
**Size:** S
**Dependencies:** ACT-001, ACT-009

### Outcome

The first beta can be evaluated against the PRD without inventing metrics after launch.

### Scope

- Add versioned SQL queries for host activation, invite activation, time-to-board percentiles, failure reason, source, and guest claim.
- Add Grafana panels for technical request/import/create/join rates and latency using the new low-cardinality metrics.
- Document the 50-start/view minimum, bot/test filtering, cohort window, and go/no-go review.
- Add alert/runbook guidance for provider failure, create/join error rate, and subscription readiness.

### Likely Files

- `docs/analytics/deck-to-game-activation.sql`
- `docs/runbooks/deck-to-game-quiet-beta.md`
- `monitoring/grafana/provisioning/dashboards/*`

### Acceptance Criteria

- Host and invite denominators match the PRD definitions.
- SQL excludes known automated test sessions without excluding ordinary guests.
- Product-funnel queries use PostgreSQL; Grafana technical panels use Prometheus.
- No dashboard requires high-cardinality metric labels.
- The runbook names the provider kill switch and paste-only fallback.

### Verification

- Queries execute against seeded event fixtures.
- Grafana provisioning loads in the local observability stack.
- A dry run produces each PRD metric from synthetic events.

## ACT-013 — End-to-End Release Gate and Regression Suite

**Priority:** P0
**Size:** L
**Dependencies:** ACT-003 through ACT-010

### Outcome

Guest activation ships only when the complete user journey and the existing authenticated journey both work.

### Scope

- Add server unit/integration coverage required by the PRD.
- Extend the Rust smoke client for guest host and join operations.
- Add isolated-browser Playwright journeys for guest host, invitee join, two-player visibility, provider fallback, and account claim.
- Retain the existing authenticated create/join E2E as a regression test.
- Add the suite to the release workflow with deterministic fixture data.
- Document staging smoke and rollback/kill-switch steps.

### Likely Files

- `tools/smoke/src/main.rs`
- `app/e2e/create-and-join-game.spec.ts`
- `app/e2e/guest-activation.spec.ts`
- `app/e2e/helpers/testData.ts`
- `.github/workflows/*`
- `docs/runbooks/deck-to-game-release.md`

### Acceptance Criteria

- A logged-out host pastes a deck and reaches a board.
- A separate logged-out invitee opens the share link, joins, and reaches the same board.
- Both players are visible before the test records success.
- Account claim preserves game access after refresh.
- Provider failure proves text-paste fallback without clearing entered state.
- Existing signup/login and authenticated create/join coverage remains green.
- Tests do not depend on a live deck provider.

### Verification

- Go tests, frontend unit/type checks, Rust smoke, and Playwright E2E pass in CI.
- A staging run records the expected authoritative and client activation events exactly once.
- The release can switch to paste-only and disable guest creation independently.

## Suggested Four-Week Cut

### Week 1 — Measure and Import

- ACT-001
- ACT-002
- ACT-003 feasibility gate
- Begin ACT-004

### Week 2 — Host Activation

- Finish ACT-003 and ACT-004
- ACT-005
- ACT-006
- Begin ACT-013 coverage alongside implementation

### Week 3 — Invite Activation

- ACT-007
- ACT-008
- ACT-009
- Continue ACT-013

### Week 4 — Acquisition and Stabilization

- ACT-010
- ACT-011
- ACT-012
- Finish ACT-013 and run the quiet-beta release gate

## Scope Guardrails

- Do not begin NFT card tracking, generalized multi-TCG work, matchmaking, monetization, or a broader board redesign inside these tickets.
- Do not replace existing `createGame` or `joinGame`; guest identities should satisfy their current authorization model.
- Do not make a live provider a CI dependency.
- Do not add user/session/game identifiers as Prometheus labels.
- Do not move JWTs to cookies inside this release; retain the existing auth-storage migration plan as separate security work.
- If neither provider passes ACT-003's feasibility gate, stop that ticket at a documented decision and ship paste-based activation rather than weakening SSRF or reliability controls.

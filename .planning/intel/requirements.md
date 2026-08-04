# Requirements

Ingest date: 2026-08-03. Mode: new. Code baseline: `main` at `ef2732a` (verified = current HEAD).

Two requirement layers are preserved separately and deliberately NOT merged:

- **REQ-A1..A7** — product-intent requirements from the PRD (personas, user stories,
  acceptance criteria). The PRD is the authority for these.
- **REQ-ACT-001..ACT-013** — implementation requirements from the SPEC (tickets),
  which explicitly derive from the PRD. Ticket IDs, priorities, sizes, and
  `depends_on` edges are reproduced verbatim; the dependency graph is load-bearing
  for downstream phase/wave derivation and must not be renumbered or reordered.

Where the two layers differ on the same question, the SPEC won under default
precedence (SPEC > PRD) and the resolution is logged as INFO in
`.planning/INGEST-CONFLICTS.md`. Three such resolutions exist: REQ-A2 (MVP provider
requirement), REQ-A5 (account-claim phasing), and invite-preview week placement.

---

# Layer 1 — Product requirements (PRD)

## REQ-A1-quick-start-without-registering
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As a Commander host, I want to load a deck and start a table without creating an account so that I can test vEDH before committing.
- acceptance:
  - `/play` is a public route.
  - The route accepts pasted deck text and supported public deck URLs.
  - A display name is optional; a readable unique guest name is generated when omitted.
  - A valid deck preview can create a short-lived guest session and call the existing `createGame` mutation.
  - The host is routed directly to `/games/:id`.
  - `board_ready` is emitted only after the game query succeeds, the current player's board state is present, and the realtime subscription is connected or has entered a documented degraded state.
  - Existing authenticated users can use the same flow without creating a guest identity.
- scope: public quick-start route, guest sessions, host activation
- implemented_by: ACT-004, ACT-005, ACT-006, ACT-009

## REQ-A2-import-deck-familiar-formats
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As a player, I want to paste the decklist I already have so that I do not need to reformat it for vEDH.
- acceptance:
  - The canonical parser accepts, at minimum: `1 Sol Ring`; `1x Sol Ring`; `1,Sol Ring`; `1, Sol Ring`; quoted CSV such as `1,"Atraxa, Praetors' Voice"`; unquoted common text such as `1 Atraxa, Praetors' Voice`.
  - Blank lines, known section headers, and common sideboard/maybeboard headers are handled deterministically.
  - Card names containing commas remain intact.
  - The preview shows total cards, detected commanders, unresolved cards, warnings, and blocking errors.
  - The preview never silently drops a row.
  - The same parser is used for host and join flows.
  - At least one public deck-link provider ships in the MVP through an adapter interface. Text paste remains available when a provider is unavailable.
- scope: decklist parser, deck preview, provider adapters
- implemented_by: ACT-002, ACT-003, ACT-004
- precedence_note: The final acceptance clause ("At least one public deck-link provider ships in the MVP") is superseded by SPEC ACT-003, which permits shipping paste-only activation if neither candidate provider clears the feasibility gate. See INFO-1 in INGEST-CONFLICTS.md. All other clauses stand unmodified.

## REQ-A3-join-from-invite-without-registering
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As an invited player, I want to understand the table and join it without signup so that accepting an invite feels immediate.
- acceptance:
  - `/join/:id` is public.
  - A public `gameInvite` query exposes only game ID, format, status, player display names, player count, capacity, and creation time.
  - The page clearly reports nonexistent, finished, and full games before asking for a deck.
  - A valid guest can use the existing `joinGame` mutation.
  - Joining never exposes hidden zones, complete board state, tokens, credentials, or decklists through the public preview.
  - The user reaches the board without being redirected through `/login` or `/signup`.
- scope: public invite preview, guest join, invite privacy
- implemented_by: ACT-007, ACT-008

## REQ-A4-share-table-immediately
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As a host, I want a clear invite control on the board so that I can bring my pod into the game.
- acceptance:
  - The board header displays a single primary invite action.
  - The action copies `${window.location.origin}/join/${gameID}`.
  - Native share is used when available, with copy-to-clipboard as the fallback.
  - The UI confirms success and provides a manual-copy fallback on failure.
  - `invite_copied` or `invite_shared` records source and game ID without recording clipboard contents.
- scope: invite sharing, board header
- implemented_by: ACT-009

## REQ-A5-convert-guest-after-value
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As an activated guest, I want to make my identity permanent without losing the game I am playing.
- acceptance:
  - Guests see a non-blocking "Save your games" prompt after `board_ready`.
  - Claiming an account sets a unique username and password on the same user UUID.
  - Current games remain associated with the claimed identity.
  - A new full-session token replaces the guest token.
  - Username conflicts and password validation errors are recoverable without leaving the board.
  - Dismissing the prompt never blocks gameplay.
- scope: guest account claiming, post-value acquisition
- implemented_by: ACT-005, ACT-010
- precedence_note: The PRD places guest account claiming in phase v1.1. The SPEC makes the claim backend P0 (ACT-005) and includes "Account claim preserves game access after refresh" in the P0 release gate (ACT-013), while the claim UI stays P1 (ACT-010). SPEC wins. See INFO-2 in INGEST-CONFLICTS.md.

## REQ-A6-understand-funnel-performance
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As the product owner, I want an authoritative activation funnel so that acquisition work is based on real drop-off data.
- acceptance:
  - The application records the event vocabulary defined in the PRD.
  - Server-side `game_created` and `player_joined` events are authoritative.
  - Client-side view/start/readiness events include a random session ID and optional campaign attribution.
  - No deck contents, deck URLs, passwords, JWTs, IP addresses, or card-level game state are stored in product-event payloads.
  - Product events can be queried by day, role, source, and outcome.
  - Prometheus exposes technical counters and latency histograms; PostgreSQL remains the source for user-funnel analysis.
- scope: product event telemetry, activation metrics
- implemented_by: ACT-001, ACT-011, ACT-012

## REQ-A7-recover-from-expected-failures
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- description: As a new player, I want clear recovery paths when an import or realtime connection fails so that I can still reach a game.
- acceptance:
  - Provider failures offer paste-text fallback without clearing the display name or game context.
  - Unresolved cards can be corrected inline before create/join.
  - Create/join errors preserve the parsed deck.
  - A failed realtime subscription shows a reconnect action and continues read-only polling when practical.
  - Error messages use product language and do not expose raw GraphQL, SQL, or provider responses.
- scope: error recovery, degraded realtime, provider fallback
- implemented_by: ACT-003, ACT-004, ACT-008, ACT-009

---

# Layer 2 — Implementation requirements (SPEC tickets)

Dependency edges reproduced verbatim from the Delivery Map. Priority rule from source:
P0 establishes measurable guest deck-to-board activation; P1 improves post-value
acquisition and operating confidence. Sizes: S = up to one focused day, M = two to
three days, L = three to five days including tests and review.

## REQ-ACT-001-product-event-and-activation-metric-foundation
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Every activation step has a privacy-safe event definition, while the Prometheus/Grafana stack on the latest `main` branch reports technical health.
- priority: P0
- size: M
- depends_on: none
- order: 1
- scope:
  - Add production and test migrations for a `product_events` table.
  - Store event name, random session ID, optional user/game IDs, role, source, outcome, duration, approved metadata, and timestamp.
  - Add `trackProductEvent` to `server/schema.graphql` and implement strict server-side event and field allowlists.
  - Make `game_created`, `player_joined`, `guest_session_created`, `deck_import_succeeded`, `deck_import_failed`, and `account_claimed` server-authoritative.
  - Add counters and latency histograms for guest sessions, imports, create/join results, and board activation.
  - Add a frontend event service that persists a random session ID and only accepts approved attribution fields.
  - Document the canonical event vocabulary and example funnel query.
- acceptance:
  - Unknown event names and metadata keys are rejected.
  - Raw deck text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game state cannot be stored through the event API.
  - The server attaches authenticated user IDs rather than accepting them from the client.
  - Events can be grouped by day, role, source, and outcome.
  - Prometheus metrics contain no username, session ID, user ID, or game ID labels.
  - Duplicate client retries do not duplicate authoritative conversion events.
- verification:
  - Migration up/down tests pass for production and test schemas.
  - Resolver tests cover allowed, rejected, oversized, and unauthenticated events.
  - Metric tests verify counter and histogram changes without high-cardinality labels.
  - Frontend unit tests verify stable session ID and attribution allowlisting.
- likely_files: persistence/migrations/*_product_events.up.sql; persistence/migrations/*_product_events.down.sql; persistence/migrations_test/*_product_events.up.sql; persistence/migrations_test/*_product_events.down.sql; server/schema.graphql; server/schema.resolvers.go; server/graphql.go; server/product_events.go; server/product_events_test.go; app/src/services/productEvents.ts

## REQ-ACT-002-canonical-deck-parser-and-preview-api
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Hosts and invitees can preview familiar pasted decklists without silent card loss or CSV-specific failures.
- priority: P0
- size: L
- depends_on: ACT-001
- order: 2
- scope:
  - Extract parsing from `createLibraryFromDecklist` in `server/games.go` into a dedicated deck-import service.
  - Support quantity/name text, `1x` text, spaced CSV, quoted CSV, section headers, blank lines, and known sideboard/maybeboard sections.
  - Preserve card names containing commas.
  - Return normalized entries, commander candidates, unresolved entries, warnings, blocking errors, and `CanContinue`.
  - Add `previewDeck` and the `DeckPreview` GraphQL types.
  - Pass a preview/import reference or normalized entries into create/join so the deck is not parsed twice with different rules.
  - Preserve the current card lookup, selected-commander removal, and maximum deck-size rules.
- acceptance:
  - The accepted examples in the PRD parse exactly as displayed.
  - `1 Atraxa, Praetors' Voice` and `1,"Atraxa, Praetors' Voice"` resolve to the same card name.
  - No nonblank input row disappears without becoming an entry, warning, or error.
  - Unknown cards are shown as unresolved and are never silently substituted.
  - Preview and final library creation consume the same normalized result.
  - Existing authenticated create/join behavior remains compatible.
- verification:
  - Table-driven Go tests cover all supported syntaxes and malformed input.
  - Golden fixtures cover common Moxfield, Archidekt, and generic text exports without relying on their live services.
  - Game tests prove commander removal and deck limits remain intact.
  - GraphQL tests cover a successful preview and structured unresolved-card response.
- likely_files: server/deck_import.go; server/deck_import_test.go; server/schema.graphql; server/schema.resolvers.go; server/games.go; test/decklists/*

## REQ-ACT-003-public-deck-provider-feasibility-gate-and-first-adapter
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: The MVP supports one familiar public deck URL without making activation dependent on an unstable third party.
- priority: P0
- size: M
- depends_on: ACT-002
- order: 3
- scope:
  - Time-box a one-day feasibility comparison of public Archidekt and Moxfield deck access.
  - Record request shape, public/private behavior, rate-limit signals, response stability, and terms/operational risk.
  - Select the first provider only if a stable public read path and fixture can be maintained.
  - Implement an adapter interface keyed by an allowlisted hostname.
  - Enforce HTTPS, DNS/IP validation, redirect revalidation, a 3-second connection timeout, 8-second total timeout, and a 1 MiB response cap.
  - Normalize provider output through ACT-002.
  - Protect the adapter with a server-side feature flag/kill switch.
- acceptance:
  - The decision record identifies the selected provider or explicitly escalates if neither clears the gate.
  - Only explicitly allowlisted hostnames can be fetched.
  - Redirects, DNS rebinding, private/reserved/link-local addresses, oversized responses, and timeout paths fail closed.
  - Public provider failure returns a normalized error and a paste-text fallback; it never blocks pasted import.
  - Private decks are reported as unsupported without requesting third-party credentials.
  - Fixture contract tests detect provider response-shape changes.
- verification:
  - Unit tests cover allowlist, redirect, DNS/IP, timeout, size, private-deck, invalid-response, and success cases.
  - No test depends on a live provider.
  - A staging smoke check imports one public fixture deck through the selected provider before release.
- likely_files: docs/research/deck-provider-feasibility.md; server/deck_providers.go; server/deck_providers_test.go; server/deck_provider_<selected>.go; server/testdata/deck_providers/*; server/deck_import.go
- produces: docs/research/deck-provider-feasibility.md (does not exist yet — forward deliverable)
- exit_note: Per SPEC scope guardrails, if neither provider passes this feasibility gate, the ticket stops at a documented decision and paste-based activation ships instead. Weakening SSRF or reliability controls to force a provider through is explicitly disallowed.

## REQ-ACT-004-reusable-deck-import-and-commander-review-ui
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Create and join share one understandable deck-review experience that preserves user work when something fails.
- priority: P0
- size: L
- depends_on: ACT-002
- order: 4
- scope:
  - Create a reusable `DeckImportPanel.vue`.
  - Add text/URL input, import source detection, loading/error states, totals, warnings, and unresolved-card correction.
  - Create a reusable commander-review component driven by `CommanderCandidates`.
  - Preserve pasted input, corrected entries, commander choices, and display name across recoverable failures.
  - Replace duplicated deck entry in `FormCreateGame.vue` and `JoinGameView.vue`.
  - Keep the current authenticated create flow usable during rollout.
- acceptance:
  - A user can paste, preview, correct unresolved cards, select commanders, and continue without re-entering the deck.
  - URL-provider failure offers text paste in the same context.
  - Blocking errors are distinct from warnings.
  - Keyboard and screen-reader users can reach inputs, issues, commander controls, and the continue action in a logical order.
  - Host and join submit the same normalized deck representation.
  - Existing authenticated creation remains available until the guest funnel passes its launch gate.
- verification:
  - Component tests cover successful paste, comma-containing names, unresolved correction, provider fallback, state preservation, and commander selection.
  - A generated type check passes after the schema update.
  - Manual viewport checks cover desktop and the minimum supported tablet width selected for beta.
- likely_files: app/src/components/decks/DeckImportPanel.vue; app/src/components/decks/CommanderReview.vue; app/src/components/games/FormCreateGame.vue; app/src/views/JoinGameView.vue; app/src/graphql/mutations.ts; app/src/types/generated.ts; app/__tests__/*

## REQ-ACT-005-guest-identity-and-account-claim-backend
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: New players can satisfy existing authorization rules without signup and later convert without losing ownership.
- priority: P0
- size: L
- depends_on: ACT-001
- order: 5
- scope:
  - Add `is_guest` and `expires_at` to production and test user schemas.
  - Add `guestSession(displayName, sessionID)` and `claimGuestAccount(username, password, sessionID)` mutations.
  - Generate collision-safe readable guest names when the display name is empty.
  - Create a random non-recoverable password hash for guests and issue a 24-hour guest token.
  - Reject expired backing guests during authorization.
  - Claim the existing user row, clear expiry, set credentials through current bcrypt rules, and issue a new full token.
  - Add rate limits, feature flag/kill switch, and an independently retryable cleanup path for unreferenced expired guests.
- acceptance:
  - Guest creation does not require a username or password.
  - Generated display names are unique and pass existing output escaping/validation.
  - Guest tokens stop working after token or backing-user expiry.
  - Guest users cannot log in through the password endpoint before claiming.
  - Claiming preserves the same user UUID and all game/player relationships.
  - Username conflicts and password validation return recoverable typed errors.
  - Cleanup never deletes a guest referenced by an active game.
- verification:
  - Migration tests pass.
  - Resolver and authorization tests cover create, collision, rate limit, expiry, login rejection, claim, conflict, and relationship preservation.
  - Security tests prove a guest cannot claim a different guest identity.
- likely_files: persistence/migrations/*_guest_users.up.sql; persistence/migrations/*_guest_users.down.sql; persistence/migrations_test/*_guest_users.up.sql; persistence/migrations_test/*_guest_users.down.sql; server/schema.graphql; server/users.go; server/auth.go; server/authz.go; server/guest_users.go; server/users_test.go

## REQ-ACT-006-public-quick-start-host-flow
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: A new Commander host can move from deck to subscribed board without visiting login or signup.
- priority: P0
- size: L
- depends_on: ACT-004, ACT-005
- order: 6
- scope:
  - Add public route `/play` and `QuickStartView.vue`.
  - Present deck import first, followed by commander review and optional display name.
  - Reuse the current authenticated user when present; otherwise create a guest session only when the deck can continue.
  - Submit `Handle` and `FormatID` correctly along with the normalized deck.
  - Call the existing `createGame` mutation and route to `/games/:id`.
  - Emit the host activation events in the PRD.
  - Add a visible recovery path for preview, guest-session, and create failures.
- acceptance:
  - `/play` is usable while logged out.
  - A valid deck plus optional display name reaches the correct game board without an auth redirect.
  - An authenticated user follows the same flow without creating a guest.
  - The requested Commander format is sent to the backend.
  - Recoverable failures preserve the deck and commander selections.
  - `game_created` is server-authoritative; the client records start and readiness events only.
- verification:
  - Router tests prove `/play` is public while unrelated private routes remain protected.
  - Store/component tests cover guest and authenticated branches and error preservation.
  - Playwright covers a logged-out paste-to-board happy path.
- likely_files: app/src/router/index.ts; app/src/views/QuickStartView.vue; app/src/stores/auth.ts; app/src/stores/games.ts; app/src/graphql/mutations.ts; app/src/components/decks/*

## REQ-ACT-007-safe-public-invite-preview
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: An invitee can confirm the table is real and joinable before providing a deck or identity.
- priority: P0
- size: M
- depends_on: ACT-001
- order: 7
- scope:
  - Add `gameInvite(gameID)` with a deliberately separate minimal response type.
  - Return format, status, display names, player count, capacity, and creation time only.
  - Make `/join/:id` public and load preview data before any guest creation.
  - Show clear states for nonexistent, finished, and full games.
  - Rate-limit public invite lookup and instrument outcomes.
- acceptance:
  - A logged-out visitor can view a valid invite.
  - The public query cannot return decklists, libraries, hands, tokens, game logs, complete board state, credentials, or hidden zones.
  - Nonexistent, full, and finished states are understandable before deck input.
  - The existing participant-only `getGame` authorization remains unchanged.
  - Invite views record session, game, and approved attribution only.
- verification:
  - Resolver tests assert the exact public response and forbidden-field absence.
  - Authorization regression tests prove `getGame` remains participant-only.
  - Router/component tests cover all invite states.
- likely_files: server/schema.graphql; server/schema.resolvers.go; server/games.go; server/games_test.go; app/src/router/index.ts; app/src/views/JoinGameView.vue; app/src/graphql/queries.ts; app/src/stores/games.ts

## REQ-ACT-008-guest-invite-to-board-flow
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: An invited player joins a valid table with a deck and reaches the board without signup.
- priority: P0
- size: L
- depends_on: ACT-004, ACT-005, ACT-007
- order: 8
- scope:
  - Add deck import, commander review, and optional display name below the safe invite preview.
  - Reuse an existing authenticated identity or create a guest at the last responsible moment.
  - Call the existing `joinGame` mutation with the canonical deck representation.
  - Preserve context on race conditions such as a table becoming full.
  - Route directly to the game board and emit invite/join activation events.
- acceptance:
  - A logged-out invitee joins without visiting login/signup.
  - Existing authenticated users can join through the same page.
  - Deck, commander, and display-name state survive a recoverable join failure.
  - If the game becomes full or finishes, the page reports the new state without leaking game details.
  - `player_joined` is written by the server only after the join transaction succeeds.
- verification:
  - Component/store tests cover guest, authenticated, full-race, finished-race, and retry paths.
  - Playwright covers a host-created invite opened by a separate guest browser context.
- likely_files: app/src/views/JoinGameView.vue; app/src/stores/auth.ts; app/src/stores/games.ts; app/src/graphql/mutations.ts; app/src/components/decks/*

## REQ-ACT-009-board-readiness-reconnect-and-invite-sharing
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Reaching the board means the game is actually usable, and the host can immediately bring in the pod.
- priority: P0
- size: M
- depends_on: ACT-006, ACT-008
- order: 9
- scope:
  - Define explicit board-loading, ready, degraded, reconnecting, and failed states in the games store.
  - Emit `board_ready` only when the game query succeeds, the current player's board exists, and the subscription is connected or documented degraded polling is active.
  - Add reconnect action and bounded polling fallback where practical.
  - Add a primary invite action to the board header.
  - Prefer native share when available and fall back to clipboard/manual copy.
  - Record only the share method, source, and game ID.
- acceptance:
  - Route navigation alone never emits `board_ready`.
  - A disconnected subscription shows a clear reconnect state rather than a silently stale board.
  - The copied URL is `${window.location.origin}/join/${gameID}`.
  - Share/copy success is confirmed; failure exposes a selectable manual URL.
  - Invite telemetry contains no clipboard content.
- verification:
  - Store tests cover ready, reconnect, degraded, and terminal failure transitions.
  - Component tests cover native share, clipboard fallback, and manual fallback.
  - Playwright proves both guest players are visible before the activation flow is considered complete.
- likely_files: app/src/views/BoardView.vue; app/src/stores/games.ts; app/src/services/apollo.ts; app/src/components/layout/*; app/__tests__/*

## REQ-ACT-010-post-value-account-claim-ui
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Activated guests can save their identity and games without leaving the board.
- priority: P1
- size: M
- depends_on: ACT-005, ACT-009
- order: 10
- scope:
  - Extend the auth store with guest detection and account claim.
  - Show a non-blocking "Save your games" prompt only after `board_ready`.
  - Collect username/password, submit the claim mutation, and replace the guest token atomically.
  - Preserve the board and active subscriptions through success and validation errors.
  - Allow dismissal without affecting gameplay.
- acceptance:
  - The prompt never appears before value is delivered.
  - Successful claim retains the same user/game access and replaces the token.
  - Username and password errors are displayed inline without route changes.
  - Dismissal is remembered for the current game/session.
  - Claim events contain no credentials.
- verification:
  - Store/component tests cover success, username conflict, password failure, dismissal, and token replacement.
  - Playwright proves a claimed guest can refresh and remain in the current game.
- likely_files: app/src/stores/auth.ts; app/src/views/BoardView.vue; app/src/components/auth/ClaimGuestAccount.vue; app/src/graphql/mutations.ts; app/__tests__/*
- phasing_note: P1 ticket, but it is a declared dependency of the P0 release gate ACT-013 ("ACT-003 through ACT-010"), and ACT-013 acceptance requires "Account claim preserves game access after refresh". Downstream phase derivation must not treat ACT-010 as deferrable past the MVP gate.

## REQ-ACT-011-commander-landing-page-and-attribution
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: The public promise matches the fast Commander experience and sends users directly into the measured flow.
- priority: P1
- size: M
- depends_on: ACT-001, ACT-006
- order: 11
- scope:
  - Rewrite the landing hero around "paste a Commander deck and start a table."
  - Make `/play` the primary CTA and retain login as a secondary action.
  - Add a compact three-step explanation and an invite/join path.
  - Capture only allowlisted UTM/referrer values into the product-event session.
  - Preserve campaign attribution through host activation and invite activation.
  - Avoid marketing claims the MVP cannot yet support.
- acceptance:
  - The primary CTA opens `/play` without auth.
  - Copy is Commander-specific and accurately describes paste and the selected provider.
  - Attribution is normalized, length-bounded, and never copied into Prometheus labels.
  - Landing CTA, quick-start, and board-ready events share the same random session ID.
  - Keyboard focus and responsive layout remain usable.
- verification:
  - Component tests cover CTA destination and attribution allowlisting.
  - A manual content/viewport check confirms no unsupported product claims.
  - Staging telemetry proves one test campaign can be followed to `board_ready`.
- likely_files: app/src/views/LandingView.vue; app/src/router/index.ts; app/src/services/productEvents.ts; app/src/styles/main.scss

## REQ-ACT-012-activation-dashboard-and-quiet-beta-runbook
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: The first beta can be evaluated against the PRD without inventing metrics after launch.
- priority: P1
- size: S
- depends_on: ACT-001, ACT-009
- order: 12
- scope:
  - Add versioned SQL queries for host activation, invite activation, time-to-board percentiles, failure reason, source, and guest claim.
  - Add Grafana panels for technical request/import/create/join rates and latency using the new low-cardinality metrics.
  - Document the 50-start/view minimum, bot/test filtering, cohort window, and go/no-go review.
  - Add alert/runbook guidance for provider failure, create/join error rate, and subscription readiness.
- acceptance:
  - Host and invite denominators match the PRD definitions.
  - SQL excludes known automated test sessions without excluding ordinary guests.
  - Product-funnel queries use PostgreSQL; Grafana technical panels use Prometheus.
  - No dashboard requires high-cardinality metric labels.
  - The runbook names the provider kill switch and paste-only fallback.
- verification:
  - Queries execute against seeded event fixtures.
  - Grafana provisioning loads in the local observability stack.
  - A dry run produces each PRD metric from synthetic events.
- likely_files: docs/analytics/deck-to-game-activation.sql; docs/runbooks/deck-to-game-quiet-beta.md; monitoring/grafana/provisioning/dashboards/*
- produces: docs/analytics/deck-to-game-activation.sql, docs/runbooks/deck-to-game-quiet-beta.md (neither exists yet — forward deliverables)

## REQ-ACT-013-end-to-end-release-gate-and-regression-suite
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- description: Guest activation ships only when the complete user journey and the existing authenticated journey both work.
- priority: P0
- size: L
- depends_on: ACT-003 through ACT-010 (verbatim from source; expands to ACT-003, ACT-004, ACT-005, ACT-006, ACT-007, ACT-008, ACT-009, ACT-010)
- order: 13
- scope:
  - Add server unit/integration coverage required by the PRD.
  - Extend the Rust smoke client for guest host and join operations.
  - Add isolated-browser Playwright journeys for guest host, invitee join, two-player visibility, provider fallback, and account claim.
  - Retain the existing authenticated create/join E2E as a regression test.
  - Add the suite to the release workflow with deterministic fixture data.
  - Document staging smoke and rollback/kill-switch steps.
- acceptance:
  - A logged-out host pastes a deck and reaches a board.
  - A separate logged-out invitee opens the share link, joins, and reaches the same board.
  - Both players are visible before the test records success.
  - Account claim preserves game access after refresh.
  - Provider failure proves text-paste fallback without clearing entered state.
  - Existing signup/login and authenticated create/join coverage remains green.
  - Tests do not depend on a live deck provider.
- verification:
  - Go tests, frontend unit/type checks, Rust smoke, and Playwright E2E pass in CI.
  - A staging run records the expected authoritative and client activation events exactly once.
  - The release can switch to paste-only and disable guest creation independently.
- likely_files: tools/smoke/src/main.rs; app/e2e/create-and-join-game.spec.ts; app/e2e/guest-activation.spec.ts; app/e2e/helpers/testData.ts; .github/workflows/*; docs/runbooks/deck-to-game-release.md
- produces: docs/runbooks/deck-to-game-release.md (does not exist yet — forward deliverable)
- dependency_note: ACT-003 is a declared dependency, and ACT-003 may legitimately terminate in a documented no-go decision with paste-only activation. In that branch the ACT-013 acceptance clause "Provider failure proves text-paste fallback" must be reinterpreted against a paste-only build. Both source documents leave this unstated.

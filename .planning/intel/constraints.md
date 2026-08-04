# Constraints

Technical contracts extracted from the SPEC (implementation tickets), with
corroborating detail from the PRD's Technical Specifications section where the two
agree. Under default precedence (SPEC > PRD) the SPEC is authoritative for every
entry here. No contradictions were found between the two documents on any constraint
in this file.

Code baseline for all constraints: `main` at `ef2732a` (verified = current HEAD).

18 constraints: 3 api-contract, 3 schema, 6 nfr, 6 protocol.

---

## GraphQL contract additions
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-001, ACT-002, ACT-005, ACT-007); corroborated by docs/product/2026-07-23-deck-to-game-activation-prd.md
- type: api-contract
- content:
  ```graphql
  type Mutation {
    guestSession(displayName: String, sessionID: String!): User!
    previewDeck(input: InputDeckImport!): DeckPreview!
    claimGuestAccount(username: String!, password: String!, sessionID: String!): User!
    trackProductEvent(input: InputProductEvent!): Boolean!
  }

  type Query {
    gameInvite(gameID: String!): GameInvite!
  }

  input InputDeckImport {
    text: String
    sourceURL: String
    sessionID: String!
  }

  type DeckPreview {
    SourceType: String!
    CardCount: Int!
    Entries: [DeckPreviewEntry!]!
    CommanderCandidates: [Card!]!
    Unresolved: [DeckImportIssue!]!
    Warnings: [String!]!
    CanContinue: Boolean!
  }

  type GameInvite {
    ID: String!
    FormatID: String!
    Status: GameStatus!
    PlayerNames: [String!]!
    PlayerCount: Int!
    Capacity: Int!
    CreatedAt: Time!
  }
  ```
  The PRD labels these "Proposed additions". The SPEC schedules each into a specific
  ticket, which is the stronger commitment.

## Canonical deck parser grammar
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-002)
- type: api-contract
- content: One parser serves both host and join. Accepted syntaxes at minimum: `1 Sol Ring`; `1x Sol Ring`; `1,Sol Ring`; `1, Sol Ring`; quoted CSV `1,"Atraxa, Praetors' Voice"`; unquoted `1 Atraxa, Praetors' Voice`. Blank lines, known section headers, and common sideboard/maybeboard sections handled deterministically. Card names containing commas preserved intact. Returns normalized entries, commander candidates, unresolved entries, warnings, blocking errors, and `CanContinue`. No nonblank input row may disappear without becoming an entry, warning, or error. Unknown cards surface as unresolved and are never silently substituted. Preview and final library creation consume the same normalized result — the deck must not be parsed twice under divergent rules. Existing card lookup, selected-commander removal, and maximum deck-size rules preserved (PRD states the current maximum is 100 cards).

## Public invite minimal response type
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-007)
- type: api-contract
- content: `gameInvite(gameID)` uses a deliberately separate minimal response type. It returns format, status, display names, player count, capacity, and creation time only. It must not be able to return decklists, libraries, hands, tokens, game logs, complete board state, credentials, or hidden zones. The existing participant-only `getGame` authorization remains unchanged; an authorization regression test must prove this.

## product_events table
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-001)
- type: schema
- content: New `product_events` table storing event name, random session ID, optional user/game IDs, role, source, outcome, duration, approved metadata, and timestamp. Must support grouping by day, role, source, and outcome. Migrations required for both production and test schemas, with passing up/down tests.

## Guest user schema fields
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-005)
- type: schema
- content: Add `is_guest` and `expires_at` to the `users` table in both production and test user schemas. Claiming a guest updates the same row and clears guest expiry — no new row, so the user UUID and all game/player relationships survive the claim.

## Migration parity across production and test schemas
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-001, ACT-005)
- type: schema
- content: Every migration ships in four files — `persistence/migrations/*.up.sql`, `persistence/migrations/*.down.sql`, `persistence/migrations_test/*.up.sql`, `persistence/migrations_test/*.down.sql`. Up/down tests must pass for both schemas.

## Deck provider SSRF and network controls
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-003); corroborated by PRD Security & Privacy
- type: nfr
- content: HTTPS only. Hostname allowlist — only explicitly allowlisted hostnames may be fetched. DNS/IP validation with redirect revalidation and bounded redirects that must remain on the allowlist. 3-second connection timeout, 8-second total timeout, 1 MiB response cap. Private, reserved, and link-local addresses rejected even when reached through DNS rebinding. Redirects, DNS rebinding, private addresses, oversized responses, and timeout paths all fail closed. Public decks only; private decks reported as unsupported without requesting third-party credentials. Per the SPEC scope guardrails, these controls may not be weakened to force a provider through the feasibility gate.

## Guest token lifetime and authorization
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-005); corroborated by PRD Backend Services
- type: nfr
- content: Guests receive a 24-hour token backed by a random, non-recoverable password hash so they cannot authenticate through the password login endpoint before claiming. Authorization rejects expired backing guests. Tokens stop working after either token expiry or backing-user expiry. Cleanup of unreferenced expired guests is opportunistic and independently retryable, and must never delete a guest referenced by an active game. NOTE: the backing-guest expiry duration (24 hours vs seven days) is an unresolved open decision — see OPEN-2 in decisions.md.

## Product-event payload privacy allowlist
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-001); corroborated by PRD Story A6 and Security & Privacy
- type: nfr
- content: Strict server-side allowlists for both event names and metadata keys; unknown names and keys are rejected. Raw deck text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game state cannot be stored through the event API. The server attaches authenticated user IDs rather than accepting them from the client. Duplicate client retries must not duplicate authoritative conversion events. Server-authoritative events: `game_created`, `player_joined`, `guest_session_created`, `deck_import_succeeded`, `deck_import_failed`, `account_claimed`. Campaign attribution is normalized and length-bounded.

## Prometheus metric cardinality
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-001, ACT-011, ACT-012, Scope Guardrails)
- type: nfr
- content: Prometheus metrics must contain no username, session ID, user ID, or game ID labels. Attribution values must never be copied into Prometheus labels. No dashboard may require high-cardinality metric labels. Division of responsibility: PostgreSQL is the source for product-funnel analysis; Prometheus/Grafana cover technical SLIs only.

## Rate limiting on public surfaces
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-005, ACT-007); corroborated by PRD Security & Privacy
- type: nfr
- content: Guest creation, deck import, and public invite lookup all require IP/session rate limiting, with outcomes instrumented.

## Testing and release-gate coverage
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-013); corroborated by PRD Testing Requirements
- type: nfr
- content: Go tests, frontend unit/type checks, Rust smoke, and Playwright E2E must all pass in CI. Isolated-browser Playwright journeys cover guest host, invitee join, two-player visibility, provider fallback, and account claim. The existing authenticated create/join E2E is retained as a regression test. Deterministic fixture data. No test may depend on a live deck provider. A staging run must record the expected authoritative and client activation events exactly once.

## gqlgen generation policy
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md (Technical Specifications)
- type: protocol
- content: Generated gqlgen files must be updated through `make generate`, not hand-edited. The SPEC does not restate this rule but does not contradict it; it lists `server/schema.resolvers.go` and `app/src/types/generated.ts` among likely files and requires a generated type check to pass after the schema update (ACT-004).

## Provider feature flag and kill switch
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-003, ACT-005, ACT-012, ACT-013)
- type: protocol
- content: The deck-provider adapter is protected by a server-side feature flag/kill switch, as is guest creation. The release must be able to switch to paste-only and disable guest creation independently. The quiet-beta runbook must name the provider kill switch and the paste-only fallback. Public provider failure returns a normalized error with paste-text fallback and never blocks pasted import.

## Preserve existing game mutation authorization model
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (Scope Guardrails); corroborated by PRD Non-Goals and Integration Points
- type: protocol
- content: Do not replace the existing `createGame` or `joinGame` mutations; guest identities must satisfy their current authorization model. Reuse authenticated `createGame`, `joinGame`, `getGame`, and subscriptions. Do not replace the existing GraphQL game and board-state model. Existing authenticated creation remains available until the guest funnel passes its launch gate.

## board_ready emission definition
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-009); corroborated by PRD Story A1
- type: protocol
- content: `board_ready` is emitted only when the game query succeeds, the current player's board state is present, and the realtime subscription is connected or a documented degraded polling mode is active. Route navigation alone must never emit `board_ready`. The games store defines explicit loading, ready, degraded, reconnecting, and failed states. A disconnected subscription shows a clear reconnect state rather than a silently stale board.

## Account claim credential rules
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (ACT-005); corroborated by PRD Security & Privacy
- type: protocol
- content: Claiming sets credentials through the current bcrypt settings and existing username-uniqueness enforcement, on the same user row and UUID. Expiry is cleared and a new full-session token replaces the guest token atomically. Username conflicts and password validation return recoverable typed errors. A security test must prove a guest cannot claim a different guest identity.

## Scope guardrails
- source: docs/plans/2026-07-23-deck-to-game-activation-tickets.md (Scope Guardrails); corroborated by PRD Non-Goals
- type: protocol
- content: Do not begin NFT card tracking, generalized multi-TCG work, matchmaking, monetization, or a broader board redesign inside these tickets. Do not make a live provider a CI dependency. Do not add user/session/game identifiers as Prometheus labels. Do not move JWTs to HttpOnly cookies inside this release; the existing auth-storage migration plan at `docs/plans/2026-05-15-vedh-auth-storage-migration.md` remains separate follow-up security work. If neither provider passes ACT-003's feasibility gate, stop that ticket at a documented decision and ship paste-based activation rather than weakening SSRF or reliability controls. The PRD adds: no new expansion epic begins until the MVP launch gate is met.

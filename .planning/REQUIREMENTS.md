# Requirements: vEDH — Deck-to-Game Activation

**Defined:** 2026-08-03 (from doc ingest)
**Core Value:** A person with a decklist reaches a working, shareable Commander board without registering — and every step of that path is measured.
**Code baseline:** `main` at `ef2732a` (verified = HEAD at ingest)
**Sources:** `docs/product/2026-07-23-deck-to-game-activation-prd.md` (PRD),
`docs/plans/2026-07-23-deck-to-game-activation-tickets.md` (SPEC). Precedence: SPEC > PRD.

## How to read this file

Two requirement layers are preserved and **deliberately not merged**:

- **Layer 1 — product (REQ-A1..A7)**, from the PRD. What a user must be able to do.
  Each spans several implementation requirements, so each is traced *through* Layer 2
  rather than mapped to a single phase.

- **Layer 2 — implementation (REQ-ACT-001..013)**, from the SPEC tickets. Ticket IDs,
  priorities, sizes, and `depends_on` edges are verbatim. **These are the phase-mapped
  requirements: each maps to exactly one phase.** The dependency graph is load-bearing
  and must not be renumbered or reordered.

Full acceptance criteria, verification steps, and likely-files lists live in
`.planning/intel/requirements.md`. This file is the checkable scope-and-coverage view.

---

## v1 Requirements — Layer 1: Product (PRD)

### Activation

- [ ] **REQ-A1**: As a Commander host, I can load a deck and start a table without creating an account.
  `/play` is public, accepts pasted text and supported deck URLs, display name is optional (a readable unique guest name is generated when omitted), a valid preview creates a short-lived guest session and calls the existing `createGame`, and the host lands on `/games/:id`. `board_ready` emits only after the game query succeeds, the current player's board state is present, and the subscription is connected or in a documented degraded state. Authenticated users use the same flow with no guest identity.
  *Implemented by: ACT-004, ACT-005, ACT-006, ACT-009*

- [ ] **REQ-A2**: As a player, I can paste the decklist I already have without reformatting it.
  The canonical parser accepts at minimum `1 Sol Ring`, `1x Sol Ring`, `1,Sol Ring`, `1, Sol Ring`, `1,"Atraxa, Praetors' Voice"`, and `1 Atraxa, Praetors' Voice`. Blank lines, section headers, and sideboard/maybeboard sections are handled deterministically. Comma-containing names stay intact. The preview shows totals, detected commanders, unresolved cards, warnings, and blocking errors, and never silently drops a row. One parser serves host and join.
  *Implemented by: ACT-002, ACT-003, ACT-004*
  *Precedence note (INFO-1): the PRD clause "at least one public deck-link provider ships in the MVP" is **superseded**. ACT-003 may ship paste-only after a documented no-go. All other clauses stand.*

- [ ] **REQ-A3**: As an invited player, I can understand the table and join it without signup.
  `/join/:id` is public. A public `gameInvite` query exposes only game ID, format, status, player display names, player count, capacity, and creation time. Nonexistent, finished, and full games are reported before a deck is requested. A valid guest uses the existing `joinGame`. Joining never exposes hidden zones, board state, tokens, credentials, or decklists. No redirect through `/login` or `/signup`.
  *Implemented by: ACT-007, ACT-008*

- [ ] **REQ-A4**: As a host, I have a clear invite control on the board so I can bring my pod in.
  A single primary invite action in the board header copies `${window.location.origin}/join/${gameID}`, prefers native share with clipboard fallback, confirms success, offers manual copy on failure, and records `invite_copied`/`invite_shared` with source and game ID but never clipboard contents.
  *Implemented by: ACT-009*

- [ ] **REQ-A5**: As an activated guest, I can make my identity permanent without losing the game I am playing.
  A non-blocking "Save your games" prompt appears after `board_ready`. Claiming sets a unique username and password on the same user UUID, current games stay associated, a new full-session token replaces the guest token, username conflicts and password errors are recoverable without leaving the board, and dismissal never blocks gameplay.
  *Implemented by: ACT-005, ACT-010*
  *Precedence note (INFO-2): the PRD places claiming in v1.1. The SPEC makes the backend P0 (ACT-005) and includes "Account claim preserves game access after refresh" in the P0 release gate (ACT-013). **SPEC wins — claim is inside this milestone and lands before ACT-013.***

- [ ] **REQ-A6**: As the product owner, I have an authoritative activation funnel.
  The app records the PRD event vocabulary. `game_created` and `player_joined` are server-authoritative. Client view/start/readiness events carry a random session ID and optional campaign attribution. No deck contents, deck URLs, passwords, JWTs, IP addresses, or card-level game state in event payloads. Events are queryable by day, role, source, and outcome. Prometheus exposes technical counters and histograms; PostgreSQL remains the source for user-funnel analysis.
  *Implemented by: ACT-001, ACT-011, ACT-012*

- [ ] **REQ-A7**: As a new player, I get clear recovery paths when an import or realtime connection fails.
  Provider failures offer paste-text fallback without clearing display name or game context. Unresolved cards are correctable inline before create/join. Create/join errors preserve the parsed deck. A failed subscription shows a reconnect action and continues read-only polling when practical. Error messages use product language and never expose raw GraphQL, SQL, or provider responses.
  *Implemented by: ACT-003, ACT-004, ACT-008, ACT-009*

---

## v1 Requirements — Layer 2: Implementation (SPEC tickets)

Sizes: S = up to one focused day; M = two to three days; L = three to five days
including tests and review. Priority rule from source: P0 establishes measurable guest
deck-to-board activation; P1 improves post-value acquisition and operating confidence.

### Foundation

- [ ] **REQ-ACT-001**: Product event and activation metric foundation — `product_events` table (prod + test migrations), `trackProductEvent` with strict server-side event and field allowlists, server-authoritative `game_created` / `player_joined` / `guest_session_created` / `deck_import_succeeded` / `deck_import_failed` / `account_claimed`, low-cardinality Prometheus counters and latency histograms, frontend event service with a persisted random session ID, documented event vocabulary and example funnel query.
  `P0 · M · depends_on: none · order 1`

- [x] **REQ-ACT-002**: Canonical deck parser and preview API — extract parsing out of `createLibraryFromDecklist` in `server/games.go` into a deck-import service; support quantity/name, `1x`, spaced CSV, quoted CSV, headers, blank lines, sideboard/maybeboard; preserve comma-containing names; return normalized entries, commander candidates, unresolved entries, warnings, blocking errors, `CanContinue`; add `previewDeck` and `DeckPreview`; feed the same normalized result into create/join; preserve card lookup, commander removal, and deck-size rules.
  `P0 · L · depends_on: ACT-001 · order 2`

- [ ] **REQ-ACT-003**: Public deck provider feasibility gate and first adapter — one-day time-boxed comparison of public Archidekt and Moxfield access; decision record; adapter interface keyed by an allowlisted hostname; HTTPS, DNS/IP validation, redirect revalidation, 3s connect / 8s total timeout, 1 MiB cap; normalize through ACT-002; server-side feature flag and kill switch.
  `P0 · M · depends_on: ACT-002 · order 3`
  *Produces: `docs/research/deck-provider-feasibility.md` (forward deliverable).*
  *Exit note: if neither provider clears the gate, the ticket stops at a documented no-go and paste-based activation ships. Weakening SSRF or reliability controls to force a provider through is explicitly disallowed.*

### Guest host path

- [ ] **REQ-ACT-004**: Reusable deck import and commander review UI — `DeckImportPanel.vue` with text/URL input, source detection, loading/error states, totals, warnings, and unresolved-card correction; a commander-review component driven by `CommanderCandidates`; preserve pasted input, corrections, commander choices, and display name across recoverable failures; replace duplicated deck entry in `FormCreateGame.vue` and `JoinGameView.vue`; keep the authenticated create flow usable during rollout.
  `P0 · L · depends_on: ACT-002 · order 4`

- [ ] **REQ-ACT-005**: Guest identity and account-claim backend — `is_guest` and `expires_at` on prod and test user schemas; `guestSession(displayName, sessionID)` and `claimGuestAccount(username, password, sessionID)`; collision-safe readable guest names; random non-recoverable password hash and 24-hour guest token; reject expired backing guests in authorization; claim the existing row, clear expiry, set credentials through current bcrypt rules, issue a new full token; rate limits, feature flag/kill switch, and an independently retryable cleanup path for unreferenced expired guests.
  `P0 · L · depends_on: ACT-001 · order 5`

- [ ] **REQ-ACT-006**: Public quick-start host flow — public route `/play` and `QuickStartView.vue`; deck import first, then commander review and optional display name; reuse an authenticated user when present, otherwise create a guest only once the deck can continue; submit `Handle` and `FormatID` correctly with the normalized deck; call the existing `createGame` and route to `/games/:id`; emit host activation events; visible recovery for preview, guest-session, and create failures.
  `P0 · L · depends_on: ACT-004, ACT-005 · order 6`

### Guest join path

- [ ] **REQ-ACT-007**: Safe public invite preview — `gameInvite(gameID)` with a deliberately separate minimal response type returning format, status, display names, player count, capacity, and creation time only; `/join/:id` public and loading preview data before any guest creation; clear nonexistent, finished, and full states; rate-limited and instrumented.
  `P0 · M · depends_on: ACT-001 · order 7`

- [ ] **REQ-ACT-008**: Guest invite-to-board flow — deck import, commander review, and optional display name below the invite preview; reuse an authenticated identity or create a guest at the last responsible moment; call the existing `joinGame` with the canonical deck representation; preserve context on races such as a table filling up; route to the board and emit invite/join activation events.
  `P0 · L · depends_on: ACT-004, ACT-005, ACT-007 · order 8`

- [ ] **REQ-ACT-009**: Board readiness, reconnect, and invite sharing — explicit loading, ready, degraded, reconnecting, and failed states in the games store; emit `board_ready` only on real game/player/subscription state; reconnect action and bounded polling fallback; primary invite action in the board header preferring native share with clipboard/manual fallback; record only share method, source, and game ID.
  `P0 · M · depends_on: ACT-006, ACT-008 · order 9`

### Post-value acquisition

- [ ] **REQ-ACT-010**: Post-value account claim UI — guest detection and claim in the auth store; non-blocking "Save your games" prompt only after `board_ready`; collect username/password, submit the claim, replace the guest token atomically; preserve board and active subscriptions through success and validation errors; dismissible without affecting gameplay.
  `P1 · M · depends_on: ACT-005, ACT-009 · order 10`
  *Phasing note (INFO-2): P1 by label, but a declared dependency of the P0 release gate ACT-013, whose acceptance requires "Account claim preserves game access after refresh". **Not deferrable past the MVP gate.***

- [ ] **REQ-ACT-011**: Commander landing page and attribution — rewrite the hero around "paste a Commander deck and start a table"; `/play` as primary CTA with login secondary; compact three-step explanation and an invite/join path; capture only allowlisted UTM/referrer values into the product-event session; preserve attribution through host and invite activation; no marketing claims the MVP cannot support.
  `P1 · M · depends_on: ACT-001, ACT-006 · order 11`

- [ ] **REQ-ACT-012**: Activation dashboard and quiet-beta runbook — versioned SQL for host activation, invite activation, time-to-board percentiles, failure reason, source, and guest claim; Grafana panels for technical request/import/create/join rate and latency from the new low-cardinality metrics; document the 50-start/view minimum, bot/test filtering, cohort window, and go/no-go review; alert and runbook guidance for provider failure, create/join error rate, and subscription readiness.
  `P1 · S · depends_on: ACT-001, ACT-009 · order 12`
  *Produces: `docs/analytics/deck-to-game-activation.sql`, `docs/runbooks/deck-to-game-quiet-beta.md` (forward deliverables).*

### Release gate

- [ ] **REQ-ACT-013**: End-to-end release gate and regression suite — server unit/integration coverage required by the PRD; Rust smoke client extended for guest host and join; isolated-browser Playwright journeys for guest host, invitee join, two-player visibility, provider fallback, and account claim; retain the authenticated create/join E2E as a regression test; add the suite to the release workflow with deterministic fixtures; document staging smoke and rollback/kill-switch steps.
  `P0 · L · depends_on: ACT-003 through ACT-010 (= ACT-003, 004, 005, 006, 007, 008, 009, 010) · order 13`
  *Produces: `docs/runbooks/deck-to-game-release.md` (forward deliverable).*
  *Dependency note (INFO-1): ACT-003 may end in a documented no-go. In that branch the acceptance clause "Provider failure proves text-paste fallback" must be reinterpreted against a paste-only build. Both source documents leave this unstated.*

---

## Dependency graph (load-bearing — do not renumber or reorder)

```
ACT-001  (P0, M)  depends: none
ACT-002  (P0, L)  depends: ACT-001
ACT-003  (P0, M)  depends: ACT-002
ACT-004  (P0, L)  depends: ACT-002
ACT-005  (P0, L)  depends: ACT-001
ACT-006  (P0, L)  depends: ACT-004, ACT-005
ACT-007  (P0, M)  depends: ACT-001
ACT-008  (P0, L)  depends: ACT-004, ACT-005, ACT-007
ACT-009  (P0, M)  depends: ACT-006, ACT-008
ACT-010  (P1, M)  depends: ACT-005, ACT-009
ACT-011  (P1, M)  depends: ACT-001, ACT-006
ACT-012  (P1, S)  depends: ACT-001, ACT-009
ACT-013  (P0, L)  depends: ACT-003 through ACT-010
```

---

## v2 Requirements

Acknowledged, not in this roadmap. Note that the PRD's v1.1 items for account claim,
landing acquisition copy, attribution, and the activation dashboard were **pulled into
v1** (see INFO-2 and the milestone scope decision); only the items below remain deferred.

### Import hardening

- **V2-IMP-01**: A second public deck-provider adapter, only if the first remains stable
- **V2-IMP-02**: Provider health monitoring and import-quality iteration
- **V2-IMP-03**: Funnel iteration based on the first 50 host starts and invite views

### Retention and expansion

- **V2-RET-01**: Persistent decks and pods
- **V2-RET-02**: "Play again" and recent-table flows
- **V2-RET-03**: Public matchmaking and trust features
- **V2-RET-04**: Mobile companion mode
- **V2-RET-05**: Additional formats and TCGs
- **V2-RET-06**: NFT-backed custom cards, only after activation and retention are healthy

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| Full Magic rules enforcement / stack resolution | Not what a board-state tracker does |
| Public matchmaking, ranking, moderation, reputation | Activation funnel comes first |
| Native mobile apps, full mobile board redesign | Web-first |
| Voice/video chat | Outside the activation path |
| Monetization or premium entitlements | No revenue model this milestone |
| More than one required deck-link provider | One is sufficient — and even that is conditional on ACT-003 |
| Private deck URLs requiring third-party credentials | Will not ask users to share credentials |
| Moving JWTs to HttpOnly cookies | Separate follow-up: `docs/plans/2026-05-15-vedh-auth-storage-migration.md` |
| NFT wallet, ownership, minting, custom cards | Strategically downstream: `docs/plans/2026-07-16-vedh-nft-card-tracking-integration.md` |
| Generalizing the board to every TCG | Not before the Commander funnel is healthy |
| Replacing the GraphQL game / board-state model | Reuse `createGame`, `joinGame`, `getGame`, subscriptions |
| Making a live deck provider a CI dependency | Tests use fixtures only |
| User/session/game identifiers as Prometheus labels | Cardinality explosion; PostgreSQL owns funnel analysis |
| Any new expansion epic | Blocked until the MVP launch gate is met |

---

## Traceability

### Layer 2 — implementation requirements to phases (one phase each)

| Requirement | Priority | Size | Phase | Status |
|-------------|----------|------|-------|--------|
| REQ-ACT-001 | P0 | M | Phase 1 | Pending |
| REQ-ACT-002 | P0 | L | Phase 1 | Complete |
| REQ-ACT-003 | P0 | M | Phase 1 | Pending |
| REQ-ACT-004 | P0 | L | Phase 2 | Pending |
| REQ-ACT-005 | P0 | L | Phase 2 | Pending |
| REQ-ACT-006 | P0 | L | Phase 2 | Pending |
| REQ-ACT-007 | P0 | M | Phase 3 | Pending |
| REQ-ACT-008 | P0 | L | Phase 3 | Pending |
| REQ-ACT-009 | P0 | M | Phase 3 | Pending |
| REQ-ACT-010 | P1 | M | Phase 4 | Pending |
| REQ-ACT-011 | P1 | M | Phase 4 | Pending |
| REQ-ACT-012 | P1 | S | Phase 4 | Pending |
| REQ-ACT-013 | P0 | L | Phase 5 | Pending |

### Layer 1 — product requirements traced through Layer 2

A product requirement is complete when all of its implementing tickets are complete.
"Completed in" is the phase containing its last implementer.

| Requirement | Implemented by | Phases touched | Completed in | Status |
|-------------|----------------|----------------|--------------|--------|
| REQ-A1 | ACT-004, ACT-005, ACT-006, ACT-009 | 2, 3 | Phase 3 | Pending |
| REQ-A2 | ACT-002, ACT-003, ACT-004 | 1, 2 | Phase 2 | Pending |
| REQ-A3 | ACT-007, ACT-008 | 3 | Phase 3 | Pending |
| REQ-A4 | ACT-009 | 3 | Phase 3 | Pending |
| REQ-A5 | ACT-005, ACT-010 | 2, 4 | Phase 4 | Pending |
| REQ-A6 | ACT-001, ACT-011, ACT-012 | 1, 4 | Phase 4 | Pending |
| REQ-A7 | ACT-003, ACT-004, ACT-008, ACT-009 | 1, 2, 3 | Phase 3 | Pending |

**Coverage:**

- Layer 2 (phase-mapped): 13 requirements, 13 mapped to exactly one phase, 0 unmapped ✓
- Layer 1 (traced): 7 requirements, 7 fully covered by mapped implementers, 0 orphaned ✓
- Dependency graph: all 13 tickets have every `depends_on` satisfied in the same or an earlier phase ✓

---
*Requirements defined: 2026-08-03*
*Last updated: 2026-08-03 after doc ingest and roadmap creation*

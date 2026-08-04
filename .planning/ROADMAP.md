# Roadmap: vEDH — Deck-to-Game Activation

## Overview

vEDH already runs Commander games: creation, joining, realtime board state, card
interactions, scoring, smoke tests, browser E2E, and local observability all work at
baseline `main@ef2732a`. This milestone does not build the game. It replaces the *path*
to the game — today both hosts and invitees must sign up, hand-format a CSV decklist,
pick commanders on a separate screen, and cross several routes before the board is
usable. The journey below builds a measured, guest-first funnel: first the server-side
deck import backbone and the event store that will prove whether any of it works, then a
logged-out host reaching a live board, then an invited stranger reaching that same board
through a shared link, then the account claim and landing page that convert activated
guests, and finally the release gate that refuses to ship unless both the new guest
journey and the existing authenticated journey pass.

**Milestone:** v1 MVP — measurable guest deck-to-board activation. All 13 SPEC tickets
(ACT-001..ACT-013) are in scope, including the two optional P1s (ACT-011, ACT-012).

**Granularity:** standard (no `.planning/config.json` present; defaults assumed —
`granularity: standard`, `phase_id_convention: sequential`, `project_code: null`).

**Phase derivation:** phases are waves of the SPEC dependency graph, which is
reproduced verbatim in `.planning/REQUIREMENTS.md` and must not be renumbered or
reordered. Every ticket's `depends_on` is satisfied in the same phase (at an earlier
wave) or an earlier phase. The grouping matches the SPEC's four-week cut, with the
release gate split out so its go/no-go is explicit.

**Milestone success metrics:** see `.planning/PROJECT.md` → Success Metrics. Evaluated
after ≥50 host activation starts and ≥50 invite views: host activation ≥60%, host
time-to-board median ≤60s / p90 ≤120s, invite activation ≥70%, join time-to-board median
≤45s / p90 ≤90s, ≥98% deck-entry resolution, create/join error rate <2%, ≥15% guest→account
claim within 7 days.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Measured Deck Import Foundation** - Any familiar decklist reaches a trustworthy server-side preview, and every activation step is recorded privacy-safely
- [ ] **Phase 2: Guest Host Activation** - A logged-out host pastes a deck and lands on a live board without seeing login or signup
- [ ] **Phase 3: Invite, Join, and Board Readiness** - A shared link takes an invited stranger to that same board, and "on the board" means the board actually works
- [ ] **Phase 4: Post-Value Acquisition and Funnel Readout** - Activated guests can keep their identity, new arrivals land on a Commander front door, and the funnel can be read against the PRD targets
- [ ] **Phase 5: Release Gate and Regression Suite** - Guest activation ships only when the full guest journey and the existing authenticated journey both pass

## Phase Details

### Phase 1: Measured Deck Import Foundation
**Goal**: Any decklist a Commander player already has — pasted in a familiar format, or pulled from an allowlisted public deck URL if one proves viable — resolves to a trustworthy preview through a single server-side path, and every step of the activation funnel has a privacy-safe event record behind it.
**Depends on**: Nothing (first phase)
**Requirements**: REQ-ACT-001, REQ-ACT-002, REQ-ACT-003
**Also advances**: REQ-A2 (partial), REQ-A6 (partial), REQ-A7 (partial)
**Success Criteria** (what must be TRUE):
  1. A pasted decklist using any supported syntax — `1 Sol Ring`, `1x Sol Ring`, `1,Sol Ring`, `1, Sol Ring`, `1,"Atraxa, Praetors' Voice"`, `1 Atraxa, Praetors' Voice` — returns a preview with total cards, commander candidates, unresolved entries, warnings, and blocking errors, and no nonblank input row disappears without becoming one of those.
  2. Card names containing commas survive import intact, `1 Atraxa, Praetors' Voice` and `1,"Atraxa, Praetors' Voice"` resolve to the same card, unknown cards surface as unresolved rather than being silently substituted, and preview and final library creation consume the same normalized result.
  3. A recorded activation event can be grouped by day, role, source, and outcome in PostgreSQL, while unknown event names, unknown metadata keys, oversized payloads, and prohibited content (raw deck text, deck URLs, passwords, JWTs, IP addresses, hidden game state) are all rejected — and the server attaches authenticated user IDs rather than trusting the client.
  4. Prometheus reports guest-session, import, create/join, and board-activation counters and latency histograms with no username, session ID, user ID, or game ID label anywhere.
  5. A written decision record either names the first public deck provider and its adapter fetches only allowlisted HTTPS hosts — with redirects, DNS rebinding, private/reserved/link-local addresses, oversized responses, and timeouts all failing closed behind a kill switch — **or** it documents a no-go, and paste-only activation proceeds. Either outcome satisfies this phase.
**Ticket waves**: ACT-001 → ACT-002 → ACT-003 (strictly sequential; each depends on the prior)
**Open decisions to resolve here**: **OPEN-1 — which public deck provider becomes the first supported URL source.** Status: open. Owned by ACT-003's one-day time-boxed feasibility comparison of public Archidekt and Moxfield access. Do not pre-answer; `/gsd-discuss-phase 1` should pick this up.
**Branch note (INFO-1)**: a provider adapter is **not** a fixed MVP commitment. If neither candidate clears the feasibility gate, ACT-003 terminates at a documented no-go and paste-only activation ships. SSRF and reliability controls may not be weakened to force a provider through. Phase 5's release gate depends on ACT-003 either way, so a no-go must not block it — it changes what "provider fallback" means in the gate, not whether the gate can close.
**Plans**: 7 plans in 6 waves
**Contract note**: Phase 1 adds one additive field to the locked `DeckPreview` type —
`BlockingErrors: [String!]!` — because the locked api-contract's prose requires blocking errors to
be returned while its GraphQL block declares no field for them. Recorded as a dated addendum in
`.planning/intel/constraints.md`; Phase 2's ACT-004 client should expect it. `previewDeck` is a
**`Mutation`** field, per the locked contract, not a query.
**Criterion 4 note (declared here, closed later)**: plan 01-01 declares all four collector families
criterion 4 names — guest session, deck import, game create/join, board activation — with their
label sets, bucket boundaries, and cardinality proof, so no later ticket invents a metric name under
deadline. Only the import and product-event families are *observed* in Phase 1. A Prometheus vector
with no observation exports no child series, so criterion 4 is satisfied in name and shape here and
**closes when the emit sites land**: guest session in Phase 2 (ACT-005), game create in Phase 2
(ACT-006), game join in Phase 3 (ACT-008), board activation in Phase 3 (ACT-009). Do not mark
criterion 4 complete on Phase 1 alone.
Plans:
- [ ] 01-01-PLAN.md — Tracer: one pasted card becomes a measured preview end to end (migration, whole Phase 1 GraphQL contract, all criterion-4 collectors, `pkg/telemetry`, `pkg/deckimport` seed, three new `server/` files, documented vocabulary, dedup/concurrency/migration proofs)
- [ ] 01-02-PLAN.md — Full deck grammar: six syntaxes, comma and double-faced names, printing metadata, sections, source detection, golden corpus
- [ ] 01-03-PLAN.md — Browser event service: persisted session identifier, fire-and-forget emission, allowlisted campaign attribution
- [ ] 01-04-PLAN.md — Name-search migration, index-usable batch lookup, printing disambiguation and missing-printing warning
- [ ] 01-05-PLAN.md — Bounded eager ranked suggestions, and one parsed deck feeding both the preview and the created library
- [ ] 01-06-PLAN.md — Public-surface hardening: per-surface rate limits (`pkg/ratelimit` + instrumented wrapper), two-layer SSRF-safe fetch client, provider kill switch defaulting off
- [ ] 01-07-PLAN.md — Provider feasibility spike, decision checkpoint, selected branch executed, coverage decision recorded

### Phase 2: Guest Host Activation
**Goal**: A Commander host who has never made an account pastes a deck, reviews what was parsed, and lands on their own live board — and an existing authenticated host travels the same road without a guest identity being invented for them.
**Depends on**: Phase 1
**Requirements**: REQ-ACT-004, REQ-ACT-005, REQ-ACT-006
**Also advances**: REQ-A1 (partial), REQ-A2 (completes), REQ-A5 (backend), REQ-A7 (partial)
**Success Criteria** (what must be TRUE):
  1. `/play` loads while logged out; a visitor can paste (or URL-import) a deck, correct unresolved cards, choose commanders, optionally set a display name, and arrive at `/games/:id` with the requested Commander format — never redirected through `/login` or `/signup`.
  2. Omitting a display name yields a readable, unique, correctly escaped generated guest name; an already-authenticated user completes the identical flow with no guest created; and the existing authenticated create form still works, now routed through the same import component and parser.
  3. Guest tokens last 24 hours, stop working when either the token or the backing guest expires, and cannot be used against the password login endpoint before claiming; guest creation is rate-limited behind a kill switch; and cleanup of expired guests never deletes one referenced by an active game.
  4. A recoverable failure — provider down, preview error, guest-session error, create error — leaves the pasted deck, corrected entries, commander choices, and display name on screen with a visible way forward, and never shows raw GraphQL, SQL, or provider output.
  5. `game_created` and `guest_session_created` are written by the server; the client emits only `quick_start_viewed`, `deck_import_started`, and `game_create_started`, and a retried client call does not double-count a conversion.
**Ticket waves**: wave 1 — ACT-004 and ACT-005 in parallel (independent; deps satisfied in Phase 1) → wave 2 — ACT-006
**Open decisions to resolve here**: **OPEN-2 — guest expiry duration: 24 hours or seven days (the token stays 24 hours either way).** Status: open, owned by ACT-005. **OPEN-3 — does the first quiet beta target desktop only, or a minimum supported tablet breakpoint?** Status: open, first needed by ACT-004's viewport verification; revisited in Phase 4 for ACT-011. Neither is pre-answered; `/gsd-discuss-phase 2` should pick both up.
**Plans**: TBD
**UI hint**: yes

### Phase 3: Invite, Join, and Board Readiness
**Goal**: A host shares the table in one action, an invited stranger can tell the table is real before committing anything, joins it without an account, and both players can see each other on a board that is genuinely live rather than merely routed to.
**Depends on**: Phase 2
**Requirements**: REQ-ACT-007, REQ-ACT-008, REQ-ACT-009
**Also advances**: REQ-A1 (completes), REQ-A3 (completes), REQ-A4 (completes), REQ-A7 (completes)
**Success Criteria** (what must be TRUE):
  1. A logged-out visitor opening `/join/:id` sees format, status, player display names, player count, capacity, and creation time — and nonexistent, finished, and full tables are reported understandably *before* any deck or identity is requested.
  2. The public invite query is structurally incapable of returning decklists, libraries, hands, tokens, game logs, complete board state, credentials, or hidden zones, and `getGame` remains participant-only, proven by an authorization regression test.
  3. A logged-out invitee provides a deck and reaches the same board as the host with both players visible; deck, commander, and display-name state survive a recoverable join failure; a table that fills or finishes mid-join reports its new state without leaking details; and `player_joined` is written by the server only after the join transaction commits.
  4. `board_ready` emits only when the game query succeeded, the current player's board state exists, and the subscription is connected or documented degraded polling is running — route navigation alone never emits it, and a dropped subscription shows a reconnect state instead of a silently stale board.
  5. The board header carries one primary invite action that copies `${window.location.origin}/join/${gameID}`, prefers native share when available, falls back to clipboard and then to a selectable manual URL, confirms success, and records only the share method, source, and game ID — never clipboard contents.
**Ticket waves**: ACT-007 → ACT-008 → ACT-009 (sequential; ACT-008 depends on ACT-007, ACT-009 on ACT-008 and on Phase 2's ACT-006)
**Sequencing note (INFO-3)**: safe invite preview (ACT-007) belongs here with guest join, not alongside the host flow. The PRD placed it a week earlier; the SPEC ordering wins and is the one consistent with the dependency graph.
**Plans**: TBD
**UI hint**: yes

### Phase 4: Post-Value Acquisition and Funnel Readout
**Goal**: A guest who just got value can make that identity permanent without leaving the board, a stranger arriving cold lands on a page that promises exactly what the product now does, and the product owner can read the real funnel against the PRD's targets instead of inventing metrics after launch.
**Depends on**: Phase 3
**Requirements**: REQ-ACT-010, REQ-ACT-011, REQ-ACT-012
**Also advances**: REQ-A5 (completes), REQ-A6 (completes)
**Success Criteria** (what must be TRUE):
  1. After `board_ready` — and never before — a guest sees a non-blocking "Save your games" prompt; claiming sets username and password on the same user UUID, atomically swaps the guest token for a full session token, and preserves the board, the active subscription, and all game access including across a refresh.
  2. Username conflicts and password validation errors are corrected inline without a route change, dismissal is remembered for the current game/session and never blocks gameplay, and no claim event carries credentials.
  3. The landing page's primary CTA opens `/play` without auth, keeps login as a secondary action, describes paste (and the selected provider, if ACT-003 selected one) accurately, and makes no claim the MVP cannot support.
  4. The landing CTA, quick-start, and `board_ready` events from one visit share the same random session ID, with UTM/referrer values allowlisted, normalized, length-bounded, and never copied into a Prometheus label.
  5. Every PRD metric — host activation rate, host time-to-board p50/p90, invite activation rate, join time-to-board p50/p90, deck-entry resolution rate, create/join error rate, and 7-day guest claim rate — is produced from seeded event fixtures by a versioned SQL query with matching denominators and bot/test filtering that does not exclude ordinary guests, while Grafana technical panels provision cleanly into the local observability stack using only low-cardinality metrics.
**Ticket waves**: wave 1 — ACT-010, ACT-011, and ACT-012 all in parallel (every dependency is satisfied by Phases 1-3)
**Phasing note (INFO-2)**: ACT-010 is labeled P1 but is a declared dependency of the P0 release gate, whose acceptance requires "Account claim preserves game access after refresh." It lands **before** Phase 5. The PRD's deferral of claim to v1.1 is overridden.
**Open decisions to resolve here**: **OPEN-4 — which operational surface hosts product-funnel queries before a dedicated internal dashboard exists.** Status: open, owned by ACT-012; partially narrowed only (PostgreSQL for funnel, Prometheus/Grafana for technical panels), no surface named. **OPEN-3** returns here for ACT-011's responsive-layout check if it was not settled in Phase 2. `/gsd-discuss-phase 4` should pick these up.
**Plans**: TBD
**UI hint**: yes

### Phase 5: Release Gate and Regression Suite
**Goal**: Guest activation ships only when a machine can prove, on every run, that the complete logged-out journey works and that nothing about the existing authenticated journey broke on the way there.
**Depends on**: Phase 4
**Requirements**: REQ-ACT-013
**Also advances**: verifies REQ-A1 through REQ-A7 end to end
**Success Criteria** (what must be TRUE):
  1. Isolated-browser Playwright journeys pass for a logged-out host pasting a deck to a board, a separate logged-out invitee opening the share link and joining that board, both players visible before success is recorded, provider failure proving text-paste fallback without clearing entered state, and a claimed account surviving a refresh with game access intact.
  2. The existing signup/login and authenticated create/join E2E coverage stays green as a regression test, and the Rust smoke client covers guest host and guest join operations.
  3. Go tests, frontend unit and type checks, Rust smoke, and Playwright E2E all run in the release workflow on deterministic fixture data, with no test reaching a live deck provider.
  4. A staging run records each expected server-authoritative and client activation event exactly once — no duplicates, no gaps.
  5. The release can independently switch to paste-only and disable guest creation, with the staging smoke, rollback, and kill-switch steps written down in a release runbook.
**Ticket waves**: wave 1 — ACT-013 (depends on ACT-003 through ACT-010, all delivered in Phases 1-4)
**Branch note (INFO-1)**: if Phase 1 ended in a provider no-go, criterion 1's "provider failure proves text-paste fallback" is evaluated against a paste-only build — the URL path is absent rather than broken. Both source documents leave this reinterpretation unstated; settle it in `/gsd-discuss-phase 5`.
**Note**: the SPEC suggests beginning ACT-013 coverage alongside implementation from week 2 onward. Earlier phases may land test scaffolding incrementally; the gate itself closes here.
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Measured Deck Import Foundation | 0/7 | Planned | - |
| 2. Guest Host Activation | 0/TBD | Not started | - |
| 3. Invite, Join, and Board Readiness | 0/TBD | Not started | - |
| 4. Post-Value Acquisition and Funnel Readout | 0/TBD | Not started | - |
| 5. Release Gate and Regression Suite | 0/TBD | Not started | - |

## Coverage

- **Implementation requirements (Layer 2):** 13 of 13 mapped to exactly one phase. No orphans, no duplicates.
- **Product requirements (Layer 1):** 7 of 7 fully covered through their implementing tickets. See the traceability tables in `.planning/REQUIREMENTS.md`.
- **Dependency graph:** all 13 `depends_on` sets are satisfied in the same phase at an earlier wave, or in an earlier phase. Verified edge by edge.
- **Locked decisions:** none exist. The ingest set contained one PRD and one SPEC, both unlocked, and zero ADRs. Technical commitments are represented as constraints in `.planning/PROJECT.md`, not as decisions.
- **Open decisions:** OPEN-1..OPEN-4 remain `status: open` and are surfaced in their owning phase above for `/gsd-discuss-phase` to resolve.

---
*Roadmap created: 2026-08-03 from doc ingest of the 2026-07-23 deck-to-game activation PRD and SPEC*

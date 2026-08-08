# vEDH

## What This Is

vEDH ("vee-dee-aych", the virtual expressive deck handler) is a Magic: The Gathering
board-state tracker built with Go, gqlgen GraphQL, PostgreSQL, and a Vue 3 frontend.
It already runs Commander games end to end: game creation, joining, realtime board
state, card interactions, and scoring. This milestone does not build the game — it
rebuilds the *path* to it, so a Commander player can paste a decklist and be on a live
board in under a minute without creating an account.

## Core Value

A person with a decklist reaches a working, shareable Commander board without
registering — and every step of that path is measured.

## Success Metrics

The v1 launch gate. Evaluated after at least 50 host activation starts and 50 invite
views (PRD, `.planning/intel/context.md`).

| Metric | Target |
|--------|--------|
| Host activation rate | ≥60% of quick-start viewers reach `board_ready` |
| Host time to board | median ≤60s, p90 ≤120s (`quick_start_viewed` → `board_ready`) |
| Invite activation rate | ≥70% of valid invite viewers reach `player_joined` |
| Join time to board | median ≤45s, p90 ≤90s (`invite_viewed` → `board_ready`) |
| Deck resolution | ≥98% of valid pasted entries resolve to a known card or a clearly identified unresolved entry |
| Create/join reliability | request error rate <2% |
| Post-value acquisition | ≥15% of activated guests claim a permanent account within 7 days |

## Requirements

Full detail with acceptance criteria in `.planning/REQUIREMENTS.md`. Two layers are
preserved deliberately: 7 product requirements (REQ-A1..A7, from the PRD) and 13
implementation requirements (REQ-ACT-001..013, from the SPEC tickets).

### Validated

Already shipped and relied upon at baseline `main@ef2732a`:

- ✓ Authenticated Commander game creation and joining
- ✓ Realtime board state, card interactions, and scoring
- ✓ Rust smoke client and Playwright browser E2E for the authenticated journey
- ✓ Local Prometheus/Grafana observability stack

Delivered since baseline:

- ✓ **Phase 1: Measured Deck Import Foundation** — complete 2026-08-08, verified 5/5.
  REQ-ACT-001/002/003 all Complete. A canonical server-side deck parser (six syntaxes,
  comma-safe names, sections, printing metadata) feeds both `previewDeck` and library
  creation from one parse; product events land in PostgreSQL behind strict server-side
  allowlists with label-free Prometheus counters; and an Archidekt URL adapter normalizes
  through that same parser behind an SSRF-controlled fetch client and a kill switch that
  defaults off. Advances REQ-A2, REQ-A6, and REQ-A7 — none fully, so all three stay Active.

### Active

- [ ] **REQ-A1** — Host loads a deck and starts a table without creating an account
- [ ] **REQ-A2** — Player pastes the decklist they already have, in familiar formats
- [ ] **REQ-A3** — Invitee understands and joins a table from an invite without signup
- [ ] **REQ-A4** — Host shares the table from the board in one action
- [ ] **REQ-A5** — Activated guest converts to a permanent account without losing the game
- [ ] **REQ-A6** — Product owner reads an authoritative activation funnel
- [ ] **REQ-A7** — New player recovers from expected import and realtime failures

### Out of Scope

- Full Magic rules enforcement or automatic stack resolution — not what a tracker is for
- Public matchmaking, ranking, moderation, reputation — activation comes first
- Native mobile apps or a full mobile board redesign — web-first
- Voice/video chat — out of the activation path
- Monetization or premium entitlements — no revenue model in this milestone
- More than one required deck-link provider in the MVP — one is enough, and even that is conditional
- Private deck URLs requiring third-party credentials — will not ask users to share credentials
- Moving JWTs to HttpOnly cookies — tracked separately in `docs/plans/2026-05-15-vedh-auth-storage-migration.md`
- NFT wallet, ownership, minting, custom cards — downstream of activation; see `docs/plans/2026-07-16-vedh-nft-card-tracking-integration.md`
- Generalizing the board to every TCG — not before the Commander funnel is healthy
- Replacing the existing GraphQL game and board-state model — reuse, do not rewrite

## Context

**Baseline:** `main` at `ef2732a`, verified equal to current HEAD at ingest. This is an
existing, working codebase. No greenfield scaffolding is required or wanted.

**Stack (verified in-repo):** Go 1.24 with gqlgen 0.17.81 and PostgreSQL (`server/`,
`persistence/`, `schema.hcl`, `gqlgen.yml`); Vue 3.5 + Pinia + Apollo Client +
graphql-ws (`app/`); Playwright and vitest for test; Prometheus + Grafana
(`monitoring/`); Docker Compose for local dev; MTGJSON for card data; Scryfall for
client-side images; Dokku for split frontend/API deployment.

**Why this milestone exists:** every game, join, board, score, and analysis route in
`app/src/router/index.ts` requires authentication. `CreateGame` and `JoinGame` require
authenticated users in `server/games.go`. Deck input is duplicated between
`FormCreateGame.vue` and `JoinGameView.vue`, both accept only the app's CSV convention,
and the backend parser in `server/games.go` can misread unquoted card names containing
commas. No safe public invite preview exists — `getGame` is participant-only. There is
no product-funnel event store and no acquisition attribution. The existing E2E test
`app/e2e/create-and-join-game.spec.ts` captures the friction exactly: two users must
complete signup before a single game can be created and joined.

**Personas:** Commander Host/Brewer (has a deck in Moxfield, Archidekt, or plain text;
values speed and a shareable link over account features); Invited Pod Member (wants to
trust the link and see the table before deciding); Activated Guest (already got value,
now worth asking for credentials).

**No AI/LLM system is required.** Deck parsing must be deterministic, testable, and
explainable. Fuzzy card-name matching may use the existing card database but must
return explicit candidates and may never silently rewrite a user's deck.

## Constraints

Full text with sources in `.planning/intel/constraints.md` (18 constraints: 3
api-contract, 3 schema, 6 nfr, 6 protocol). Authority is the SPEC
(`docs/plans/2026-07-23-deck-to-game-activation-tickets.md`) under SPEC > PRD
precedence. Condensed:

- **API contract**: GraphQL additions are fixed — `guestSession`, `previewDeck`,
  `claimGuestAccount`, `trackProductEvent`, `gameInvite`, plus `InputDeckImport`,
  `DeckPreview`, `GameInvite` — because downstream tickets are written against these shapes.
- **API contract**: one canonical deck parser serves both host and join; preview and
  final library creation consume the same normalized result — so a deck is never parsed
  twice under divergent rules.
- **API contract**: `gameInvite` uses a deliberately separate minimal response type and
  cannot return decklists, libraries, hands, tokens, logs, board state, credentials, or
  hidden zones — a public surface must be incapable of leaking, not merely careful.
- **Schema**: new `product_events` table (event name, random session ID, optional
  user/game IDs, role, source, outcome, duration, approved metadata, timestamp),
  groupable by day, role, source, outcome.
- **Schema**: `is_guest` and `expires_at` on `users`; claiming updates the same row so
  the UUID and all game/player relationships survive.
- **Schema**: migration parity — every migration ships four files across
  `persistence/migrations/` and `persistence/migrations_test/`, with passing up/down tests.
- **Security (SSRF)**: provider fetches are HTTPS-only, hostname-allowlisted, DNS/IP
  validated with redirect revalidation, 3s connect / 8s total timeout, 1 MiB cap;
  private, reserved, and link-local addresses rejected; all failure paths fail closed.
  These controls may not be weakened to force a provider through the feasibility gate.
- **Security (guests)**: 24-hour guest token over a random non-recoverable password
  hash, so guests cannot use the password login endpoint before claiming; authorization
  rejects expired backing guests; cleanup never deletes a guest referenced by an active game.
- **Privacy**: strict server-side allowlists for event names and metadata keys; raw deck
  text, deck URLs, passwords, JWTs, clipboard values, IP addresses, and hidden game
  state can never be stored; the server attaches authenticated user IDs; duplicate
  client retries never duplicate authoritative conversion events.
- **Observability**: no username, session ID, user ID, game ID, or attribution value may
  become a Prometheus label. PostgreSQL is the source for product-funnel analysis;
  Prometheus/Grafana cover technical SLIs only.
- **Rate limiting**: guest creation, deck import, and public invite lookup all require
  IP/session rate limiting with instrumented outcomes.
- **Testing**: Go tests, frontend unit/type checks, Rust smoke, and Playwright E2E all
  pass in CI on deterministic fixtures; no test may depend on a live deck provider.
- **Protocol**: generated gqlgen files are updated via `make generate`, never hand-edited.
- **Protocol**: the deck provider adapter and guest creation each have an independent
  server-side feature flag/kill switch; the release can drop to paste-only and disable
  guest creation separately.
- **Protocol**: do not replace `createGame`, `joinGame`, `getGame`, or subscriptions —
  guest identities satisfy the existing authorization model. Authenticated creation stays
  available until the guest funnel passes its launch gate.
- **Protocol**: `board_ready` emits only when the game query succeeded, the current
  player's board exists, and the subscription is connected or documented degraded polling
  is active. Route navigation alone never emits it.
- **Protocol**: claiming sets credentials through current bcrypt settings and existing
  username uniqueness on the same row; expiry clears and a new full token replaces the
  guest token atomically; a security test must prove a guest cannot claim another guest's identity.
- **Protocol (scope)**: no NFT work, no multi-TCG generalization, no matchmaking, no
  monetization, no broader board redesign inside this milestone. No new expansion epic
  begins until the MVP launch gate is met.

## Key Decisions

**Locked decisions: none (0).**

The ingest set contained one PRD and one SPEC, both `locked: false`, and zero ADRs. No
decision in this project carries `status: locked` or `status: proposed`, because neither
can be derived from a PRD or SPEC without fabricating an ADR that does not exist. The
technical commitments that would normally appear here live in **Constraints** above and
in `.planning/intel/constraints.md`, which reflects their actual source.

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| *(none — no ADR-backed decisions exist)* | | |

### Precedence resolutions applied at ingest

Not decisions — mechanical applications of SPEC > PRD. Full detail in
`.planning/INGEST-CONFLICTS.md`.

| ID | Resolution | Effect |
|----|-----------|--------|
| INFO-1 | A shipped provider adapter is conditional on the ACT-003 feasibility gate, not guaranteed | Paste-only activation is a legitimate outcome; it must not block the release gate |
| INFO-2 | Guest account claim is inside the P0 release gate, not deferred to v1.1 | ACT-010 lands before ACT-013 |
| INFO-3 | Safe invite preview sequences with guest join, not before the host flow | Drives wave ordering in Phase 3 |

### Open decisions (unresolved — owned by a phase, not by this document)

Carried forward verbatim from the PRD's "Open Decisions". Status: **open**. These are
deliberately not answered here; each is surfaced in the phase that owns it so
`/gsd-discuss-phase` picks it up.

| ID | Question | Owning phase | Resolution path |
|----|----------|--------------|-----------------|
| ~~OPEN-1~~ **RESOLVED 2026-08-08** | Which public deck provider becomes the first supported URL source | Phase 1 (ACT-003) | **Archidekt.** The feasibility spike selected Moxfield (D-14), but its contract proved unobtainable — `api.moxfield.com/robots.txt` is a blanket `Disallow: /` and Cloudflare 403s every unauthenticated probe — so no field mapping could be written without guessing. D-14 was reversed to Archidekt, whose contract was observed and captured in-repo. Moxfield scaffold removed rather than left registered-but-failing. Operator gate before enabling: a human must read `https://archidekt.com/terms`; `DECK_PROVIDER_ENABLED` defaults false. |
| OPEN-2 | Guest expiry duration: 24 hours or seven days (token stays 24h either way) | Phase 2 (ACT-005) | Unresolved in both documents |
| OPEN-3 | Quiet beta targets desktop only, or a minimum supported tablet breakpoint | Phase 2 (ACT-004), revisited Phase 4 (ACT-011) | Unresolved; ACT-004 verification presumes an answer exists |
| OPEN-4 | Operational surface for product-funnel queries before a dedicated dashboard | Phase 4 (ACT-012) | Partially narrowed (PostgreSQL for funnel, Prometheus for technical), surface unnamed |

---
*Last updated: 2026-08-08 after Phase 1 (Measured Deck Import Foundation) completed and verified; OPEN-1 resolved to Archidekt*

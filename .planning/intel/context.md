# Context

No DOC-type documents were ingested. This file holds product framing that the PRD
carries and the SPEC does not address — personas, success metrics, non-goals,
current-state findings, risks, and phased rollout. The PRD is the authority for
everything below. It is preserved here rather than dropped, because the SPEC's
silence on these topics is not disagreement.

Code baseline: `main` at `ef2732a` (verified = current HEAD). The described work is
unbuilt against exactly this tree.

---

## Objectives and delivery envelope
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Primary objective: deck-to-game activation.
- Secondary objective: user acquisition after activation.
- Recommended delivery envelope: four weeks for one implementation agent, including stabilization.
- PRD status: "Draft for implementation". SPEC status: "Ready for estimation".

## Problem statement
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- vEDH already has working Commander game creation, joining, realtime board state, card interactions, scoring, smoke coverage, browser E2E coverage, and local observability. The path to that value is too long: both hosts and invitees must create an account, manually provide a CSV-formatted decklist, select commanders separately, and navigate several screens before the board is usable.
- The existing browser E2E test in `app/e2e/create-and-join-game.spec.ts` captures the friction: two users must complete signup before they can create and join a single game.
- The create and join forms in `app/src/components/games/FormCreateGame.vue` and `app/src/views/JoinGameView.vue` accept only the application's CSV convention, while the backend parser in `server/games.go` assumes CSV rows and can misread unquoted card names containing commas.

## Proposed solution shape
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Guest-first quick-start funnel: paste a decklist or supported public deck URL; review a parsed preview with detected commander candidates; enter an optional display name; create or join without registering; reach a subscribed interactive board; offer account creation only after value is received.
- Reuses the existing Vue 3, Pinia, Apollo, Go, gqlgen, PostgreSQL, MTGJSON, Playwright, Prometheus, and Grafana stack.

## Personas
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- **Commander Host / Brewer** — has a deck in Moxfield, Archidekt, or plain text and wants to begin testing immediately. Values speed, deck accuracy, and a shareable table link more than account features.
- **Invited Pod Member** — receives a link from a friend and wants to join with minimal context. Values trust that the link is valid, a clear player/format preview, and not having to create an account before deciding whether vEDH is useful.
- **Activated Guest** — has already reached a board and may want to preserve a display name, game access, and future history. This is the appropriate point to request permanent account credentials.

## Success metrics
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Evaluated after at least 50 host activation starts and 50 invite views.
- **Host activation rate:** at least 60% of users who open quick start reach `board_ready`.
- **Host time to board:** median at or below 60 seconds, p90 at or below 120 seconds, from `quick_start_viewed` to `board_ready`.
- **Invite activation rate:** at least 70% of valid invite viewers reach `player_joined`.
- **Join time to board:** median at or below 45 seconds, p90 at or below 90 seconds, from `invite_viewed` to `board_ready`.
- **Technical success:** at least 98% of valid pasted deck entries resolve to a known card or a clearly identified unresolved entry; create/join request error rate below 2%.
- **Post-value acquisition:** at least 15% of activated guest users claim a permanent account within seven days.

## Product event vocabulary
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Client events: `landing_primary_cta` (session ID, campaign source); `quick_start_viewed` (session ID); `deck_import_started` (session ID, source type); `game_create_started` (session ID); `invite_copied` (session ID, game ID); `invite_viewed` (session ID, game ID, campaign source); `join_started` (session ID, game ID); `board_ready` (session ID, game ID, user role, elapsed milliseconds); `account_claim_started` (session ID).
- Server events: `deck_import_succeeded` (session ID, source type, duration, card count, unresolved count); `deck_import_failed` (session ID, source type, normalized reason); `guest_session_created` (session ID); `game_created` (session ID, game ID, user role); `player_joined` (session ID, game ID); `account_claimed` (session ID).

## Current-state findings
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- All game, join, board, score, and analysis routes require authentication in `app/src/router/index.ts`.
- Authentication is username/password with a 24-hour JWT stored in browser `localStorage`.
- `CreateGame` and `JoinGame` require authenticated users in `server/games.go`.
- A safe public invite preview does not exist; `getGame` is participant-only.
- The host form contains game name, deck size, and format controls, but its submitted payload does not currently send `Handle` or `FormatID`.
- Deck input is duplicated between create and join views.
- Deck parsing is coupled to game creation and accepts only CSV-shaped input.
- There is no product-funnel event store or acquisition attribution.
- The latest `main` branch adds board/score polish and local Prometheus/Grafana tooling but does not change these activation constraints.
- The NFT card tracking plan is strategically downstream of activation and must not preempt this work.

## Non-goals
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Full Magic rules enforcement or automatic stack resolution.
- Broad public matchmaking, ranking, moderation, or reputation systems.
- Native mobile applications or a full mobile board redesign.
- Voice/video chat.
- Monetization or premium entitlements.
- More than one required deck-link provider in the MVP.
- Importing private deck URLs that require users to share third-party credentials.
- Moving JWT storage to HttpOnly cookies in this activation release.
- NFT wallet, ownership, minting, or custom-card integration.
- Generalizing the board to every TCG before the Commander activation funnel is healthy.
- Replacing the existing GraphQL game and board-state model.

## AI system requirements
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- This release does not require an AI or LLM system. Deck parsing must be deterministic, testable, and explainable. Fuzzy card-name matching may use the existing card database but must return explicit candidates and may not silently rewrite a user's deck.

## Phased rollout
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- **MVP — measurable guest activation.** Scope: product-event foundation and technical activation metrics; canonical pasted-text parser and deck preview; one allowlisted public deck-provider adapter selected by the feasibility gate, with pasted-text fallback; guest sessions; public quick-start route; safe invite preview; guest join; invite share action; board-readiness/reconnect state; updated smoke and E2E release gates. Launch gate: guest host and guest join E2E pass against staging; no regression in authenticated create/join; the selected provider's public-deck fixture suite passes and provider failure preserves the paste fallback; create/join error rate below 2% in the first quiet-beta cohort.
- **v1.1 — post-value acquisition and import hardening.** Scope: a second public deck-provider adapter only if the first remains stable; provider health monitoring and import-quality iteration; guest account claiming; Commander-focused landing-page acquisition copy; campaign attribution and activation dashboard; funnel iteration based on the first 50 host starts and invite views. Launch gate: at least one provider meets the 98% valid-entry resolution target; guest claim preserves existing games and active subscriptions; host and invite activation meet or show a credible path to the beta targets.
- **v2.0 — retention and expansion (possible scope).** Persistent decks and pods; "play again" and recent-table flows; public matchmaking and trust features; mobile companion mode; additional formats and TCGs; and only after activation and retention are healthy, NFT-backed custom cards from the existing integration plan.
- NOTE: the SPEC's ticket priorities partially override this phasing for guest account claiming. See INFO-2 in INGEST-CONFLICTS.md.

## Delivery sequence — both versions
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md and docs/plans/2026-07-23-deck-to-game-activation-tickets.md
- PRD "Recommended Delivery Sequence": week 1 — product events, canonical parser, deck preview, first provider feasibility gate. Week 2 — guest identity, quick-start host flow, safe invite preview. Week 3 — guest join, invite sharing, board readiness, reconnect handling. Week 4 — account claim, landing acquisition updates, complete E2E/CI, quiet-beta readout.
- SPEC "Suggested Four-Week Cut": week 1 — ACT-001, ACT-002, ACT-003 feasibility gate, begin ACT-004. Week 2 — finish ACT-003 and ACT-004, ACT-005, ACT-006, begin ACT-013 coverage alongside implementation. Week 3 — ACT-007, ACT-008, ACT-009. Week 4 — ACT-010, ACT-011, ACT-012, finish ACT-013 and run the quiet-beta release gate.
- The two agree except on safe invite preview (ACT-007), which the PRD places in week 2 and the SPEC in week 3. SPEC wins. See INFO-3 in INGEST-CONFLICTS.md.

## Technical risks and mitigations
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- **Provider instability** — Moxfield and Archidekt provide no contractual dependency equivalent to vEDH's own API. Mitigation: adapter isolation, timeouts, fixtures, feature flags, normalized errors, paste fallback.
- **Guest abuse and database growth** — public session creation increases bot and spam exposure. Mitigation: short token lifetime, rate limits, guest expiry, opportunistic cleanup, event monitoring, kill switch.
- **Parser ambiguity** — deck exports vary and some card names contain punctuation or commas. Mitigation: deterministic grammar, preview-before-create, no silent drops, fixtures from real exports, explicit unresolved entries.
- **Realtime readiness ambiguity** — routing to the board does not prove realtime state is usable. Mitigation: define `board_ready` from actual game/player/subscription state and provide a degraded polling mode.
- **Product-event abuse or privacy leakage** — a public analytics mutation can become a junk-data or data-leak surface. Mitigation: strict event allowlist, bounded fields, rate limits, server-authoritative conversion events, no raw content.
- **Competing roadmap work** — board polish, generalized formats, and NFT integration can consume the same implementation capacity without improving activation. Mitigation: no new expansion epic begins until the MVP launch gate is met.

## Integration points
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Existing GraphQL API: reuse authenticated `createGame`, `joinGame`, `getGame`, and subscriptions.
- PostgreSQL: guest identities, product events, games, game logs, cards.
- MTGJSON card data: card resolution and hydration.
- Scryfall: existing client-side image lookup remains unchanged.
- Deck providers: public, allowlisted, read-only imports through server adapters.
- Prometheus/Grafana: technical SLI monitoring, not user-level analytics.
- Dokku: existing split frontend/API deployment remains unchanged.

## Frontend change surface
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- Add `QuickStartView.vue` at public route `/play`. Refactor deck input into a reusable `DeckImportPanel.vue`. Refactor commander selection into a reusable component fed by `DeckPreview`. Make `/join/:id` public and render safe invite metadata before guest creation. Extend `useAuthStore` with guest session and account-claim actions. Extend `useGamesStore` with safe invite loading and explicit board-readiness state. Add a small product-events service generating and persisting a random session ID and approved attribution fields. Add an invite action to `BoardView.vue`. Preserve the authenticated create flow during rollout, routed through the same import component and parser.

## Related plans referenced but not ingested
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- `docs/plans/2026-07-16-vedh-nft-card-tracking-integration.md` — exists on disk; strategically downstream of activation, explicitly must not preempt this work.
- `docs/plans/2026-05-15-vedh-auth-storage-migration.md` — exists on disk; documents the existing localStorage bearer-token risk, which remains a follow-up outside this release.
- Neither was part of this ingest set. Neither was classified or synthesized.

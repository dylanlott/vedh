---
phase: 02-guest-host-activation
plan: 04
subsystem: activation-telemetry
tags: [graphql, vue, vitest, prometheus, product-events, guest-auth]

requires:
  - phase: 01-measured-deck-import-foundation
    provides: closed product-event vocabulary, authoritative-event rejection, dedup index, and metric collectors
  - phase: 02-guest-host-activation
    plan: 02
    provides: guest identity lifecycle and safe activation error contract
  - phase: 02-guest-host-activation
    plan: 03
    provides: reusable deck import, commander review, and draft-preserving quick-start flow
provides:
  - server-authoritative and retry-idempotent game_created attribution after persistence
  - bounded game-create and guest-session metric observations with outcome-only labels
  - exact three-event client activation funnel at mount, import, and create-attempt boundaries
  - tested authenticated-versus-guest branching and four draft-preserving failure exits
  - complete public-/play and protected-route regression coverage
affects: [02-05-guest-refresh, 02-06-human-verification, phase-03-guest-join, product-analytics]

tech-stack:
  added: []
  patterns:
    - server-owned conversion events are recorded only after persistence and skipped when optional attribution is absent
    - reusable child components emit attempt-boundary signals while the route-level orchestrator owns product-event names
    - browser session attribution travels through opaque create payloads without entering metric labels

key-files:
  created:
    - app/__tests__/router.spec.ts
  modified:
    - server/schema.graphql
    - server/generated.go
    - server/models_gen.go
    - server/games.go
    - server/games_test.go
    - app/src/components/decks/DeckImportPanel.vue
    - app/src/components/games/FormCreateGame.vue
    - app/src/graphql/mutations.ts
    - app/src/types/generated.ts
    - app/src/views/QuickStartView.vue
    - app/__tests__/QuickStartView.spec.ts
    - app/__tests__/FormCreateGame.integration.spec.ts

key-decisions:
  - "InputCreateGame.SessionID is optional attribution only: blank values skip game_created entirely, preserving compatibility and avoiding false telemetry drops."
  - "game_created is recorded after a successful game persistence with clientSubmitted=false; the existing authoritative unique index owns retry idempotency."
  - "DeckImportPanel exposes one submit-attempt signal while QuickStartView alone owns quick_start_viewed, deck_import_started, and game_create_started names."

patterns-established:
  - "Conversion ownership: clients announce attempts, while the server records authoritative success after durable persistence."
  - "Retry-safe identity: a successful guest session remains in the auth store, so a failed game create can retry without minting another guest."
  - "Route regression matrix: every protected route pattern is asserted alongside public activation and auth routes."

requirements-completed: [REQ-ACT-006]

coverage:
  - id: D1
    description: "CreateGame writes one authoritative game_created row only after persistence, deduplicates identical retries, skips missing sessions, and records bounded success/failure metrics."
    requirement: REQ-ACT-006
    verification:
      - kind: integration
        ref: "server/games_test.go#TestGames_CreateWritesGameCreated, TestGames_CreateGameCreatedDedup, TestGames_CreateWithoutSessionWritesNoEvent, TestGames_CreateFailureWritesNoEvent"
        status: pass
      - kind: integration
        ref: "server/games_test.go#TestGames_CreateAndGuestMetricsHaveBoundedLabels and server/product_events_test.go#TestProductEvents_AuthoritativeDedup"
        status: pass
      - kind: other
        ref: "go build ./... and go test ./server -race"
        status: pass
    human_judgment: false
  - id: D2
    description: "The quick-start client emits exactly quick_start_viewed, deck_import_started, and game_create_started, and forwards game label, Commander format, and browser session on create."
    requirement: REQ-ACT-006
    verification:
      - kind: automated_ui
        ref: "app/__tests__/QuickStartView.spec.ts#event ordering, exact event set, create payload, and double-submit coverage"
        status: pass
      - kind: integration
        ref: "app/__tests__/FormCreateGame.integration.spec.ts#create payload carries Handle, FormatID, and SessionID"
        status: pass
    human_judgment: false
  - id: D3
    description: "Authenticated visitors skip guest creation, guests are minted only at the final valid step, retries reuse identity, and all four recoverable failures retain safe visible state."
    requirement: REQ-ACT-006
    verification:
      - kind: automated_ui
        ref: "app/__tests__/QuickStartView.spec.ts#authenticated branch, guest timing, four failure classes, retry, draft, and Unicode coverage"
        status: pass
    human_judgment: false
  - id: D4
    description: "/play remains public for logged-out and logged-in visitors while all six pre-existing protected route patterns retain their authentication redirects."
    requirement: REQ-ACT-006
    verification:
      - kind: unit
        ref: "app/__tests__/router.spec.ts#10 route guard assertions"
        status: pass
      - kind: other
        ref: "git diff --exit-code app/src/router/index.ts"
        status: pass
    human_judgment: false

duration: 21min
completed: 2026-08-20
status: complete
---

# Phase 2 Plan 4: Measured Quick-Start Completion Summary

**Quick-start now produces a retry-safe server-owned conversion, exactly three client attempt events, and tested guest/authenticated outcomes without weakening route guards or losing recovery state.**

## Performance

- **Duration:** 21 min
- **Started:** 2026-08-20T19:17:17Z
- **Completed:** 2026-08-20T19:37:55Z
- **Tasks:** 2
- **Files modified:** 13

## Accomplishments

- Added optional browser-session attribution to `InputCreateGame` and recorded `game_created` only after persistence, with authoritative rejection and partial-index dedup behavior proven under race-enabled server tests.
- Activated the game-create collector and confirmed the existing guest-session collector on all exits, with tests enumerating bounded outcome-only metric labels.
- Completed the exact client funnel, authenticated/guest branch, retry protection, safe four-class recovery, create-payload metadata, and full route-guard regression matrix.

## Task Commits

Each TDD task has a RED test commit followed by its GREEN implementation commit:

| Task | Gate | Commit |
|------|------|--------|
| 1. Server-authoritative conversion and metric observations | RED | `d2da30a` |
| 1. Server-authoritative conversion and metric observations | GREEN | `ff5337f` |
| 2. Client events, identity branch, failures, and router proof | RED | `1070dcf` |
| 2. Client events, identity branch, failures, and router proof | GREEN | `3bca9d9` |

## Files Created/Modified

- `server/schema.graphql`, `server/generated.go`, and `server/models_gen.go` — optional `InputCreateGame.SessionID` GraphQL contract and generated Go surface.
- `server/games.go` — outcome-timed game-create metric observation and post-persistence authoritative conversion write.
- `server/games_test.go` — persistence ordering, dedup, missing-session, failure, authoritative replay, and bounded-label coverage.
- `app/src/views/QuickStartView.vue` — exact three client events plus labelled Commander create payload and explicit guest/authenticated branch.
- `app/src/components/decks/DeckImportPanel.vue` — one event emitted at each valid import-attempt boundary.
- `app/src/components/games/FormCreateGame.vue`, `app/src/graphql/mutations.ts`, and `app/src/types/generated.ts` — authenticated create attribution and matching hand-written GraphQL input contract.
- `app/__tests__/QuickStartView.spec.ts` and `app/__tests__/FormCreateGame.integration.spec.ts` — end-to-end component behavior for event ordering, identity timing, recovery, retries, duplicate prevention, Unicode, and create metadata.
- `app/__tests__/router.spec.ts` — public `/play`, public auth routes, and all six protected route-pattern assertions.

## Decisions Made

- Kept `SessionID` optional and attribution-only. Missing or whitespace input attempts no event, so older callers remain valid and ordinary creates do not inflate telemetry-drop metrics.
- Recorded `game_created` after `upsertGame` and used the authenticated user plus game and session keys already covered by the authoritative partial unique index.
- Kept event vocabulary ownership in `QuickStartView`; the shared deck panel reports only the import attempt boundary and cannot emit product event names itself.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added an import-attempt signal to the shared deck panel**
- **Found during:** Task 2 (client funnel events)
- **Issue:** `QuickStartView` could not observe the exact import-attempt boundary because `DeckImportPanel` exposed only state, resolution, failure, and continuation events; the plan's file list omitted the component that owned submission.
- **Fix:** Added one `submit` event after valid-input validation and before the preview request begins, then handled that signal in `QuickStartView` to emit `deck_import_started` once per attempt.
- **Files modified:** `app/src/components/decks/DeckImportPanel.vue`, `app/src/views/QuickStartView.vue`
- **Verification:** Focused quick-start tests assert paste/URL attempt cardinality and the complete ordered event set; the full frontend suite passes.
- **Committed in:** `3bca9d9`

---

**Total deviations:** 1 auto-fixed (1 Rule 2).
**Impact on plan:** The added component signal is the minimum reusable boundary required for correct event timing; it adds no product scope and keeps telemetry names out of the shared UI component.

## Issues Encountered

The first focused Vitest invocation supplied `--run` twice because the package script already includes it. The corrected invocation passed without code changes.

## Known Stubs

None. Nullable page/preview state, empty collection initialization, generated GraphQL fallbacks, and pre-existing TODO comments are intentional runtime or legacy patterns rather than unwired plan behavior.

## Verification

- `go build ./...` — pass.
- `go test ./server -run 'TestGames_Create|TestProductEvents_AuthoritativeDedup' -race -v` — pass, including every new create and bounded-label case.
- `go test ./server -race` — pass in 246.162 seconds, with the PostgreSQL-backed suite active rather than skipped.
- `npm --prefix app run test` — pass: 109 passed, 3 intentional skips.
- `npm --prefix app run type-check` — pass.
- Router and closed event vocabulary clean-diff gates — pass.
- Authoritative event-name grep in `QuickStartView.vue` — zero matches.
- Root `edhgo` build artifact — absent.

## User Setup Required

None — no dependencies, credentials, or external configuration were added.

## Next Phase Readiness

- The activation conversion chain now has durable server ownership and exact client attempt boundaries for Phase 2 verification and subsequent analytics.
- Plan 02-05 can build on the preserved guest identity/session lifecycle; plan 02-06 can exercise the complete quick-start path in the running browser.
- No blockers remain.

## Self-Check: PASSED

All thirteen created or modified plan files exist, and all four RED/GREEN task commits are present in git history.

---
*Phase: 02-guest-host-activation*
*Completed: 2026-08-20*

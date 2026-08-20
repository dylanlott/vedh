---
phase: 02-guest-host-activation
plan: 05
subsystem: player-identity-rendering
tags: [graphql, vue, vitest, postgres, display-name, identity]

requires:
  - phase: 02-guest-host-activation
    plan: 01
    provides: nullable users.display_name and User.DisplayName GraphQL contract
  - phase: 02-guest-host-activation
    plan: 04
    provides: complete guest/authenticated create path and persisted game payload flow
provides:
  - write-time display-name snapshots on creating and joining game players
  - one displayNameOf rendering seam with username fallback
  - DisplayName selection in all five player-bearing GraphQL documents
  - complete player-label sweep with username identity joins preserved
affects: [02-06-human-verification, phase-03-invite-preview, board-rendering, game-analysis]

tech-stack:
  added: []
  patterns:
    - cosmetic player labels are snapshotted at game writes while unique usernames remain identity keys
    - every player label resolves through one helper and Vue text interpolation

key-files:
  created:
    - app/src/services/displayName.ts
    - app/__tests__/displayName.spec.ts
  modified:
    - server/games.go
    - server/games_test.go
    - app/src/graphql/queries.ts
    - app/src/graphql/mutations.ts
    - app/src/types/generated.ts
    - app/src/components/layout/AppNav.vue
    - app/src/views/ScoreView.vue
    - app/src/views/GamesView.vue
    - app/src/views/BoardView.vue
    - app/src/views/GameAnalysisView.vue

key-decisions:
  - "CreateGame and JoinGame snapshot only the authenticated actor's users.display_name; a missing row or failed cosmetic lookup logs and proceeds with nil."
  - "displayNameOf is rendering-only: Username remains the comparison, map, turn, zone, and authorization key everywhere."

patterns-established:
  - "Player label fallback: nonblank DisplayName, then Username, then an empty string for incomplete legacy-shaped inputs."
  - "GraphQL player selection invariants are tested by printing every exported player-bearing document."

requirements-completed: [REQ-ACT-005]

coverage:
  - id: D1
    description: "CreateGame and JoinGame persist the acting user's nullable display name while legacy payloads deserialize without it and duplicate labels do not affect identity."
    requirement: REQ-ACT-005
    verification:
      - kind: integration
        ref: "server/games_test.go#TestGames_CreateStoresDisplayName, TestGames_JoinStoresDisplayName, TestGames_LegacyPayloadHasNoDisplayName, TestGames_DisplayNameIsNotAnIdentityKey"
        status: pass
      - kind: other
        ref: "go build ./... and go test ./server -race"
        status: pass
    human_judgment: false
  - id: D2
    description: "All five player-bearing GraphQL documents select DisplayName and every declared player-label surface renders through displayNameOf."
    requirement: REQ-ACT-005
    verification:
      - kind: unit
        ref: "app/__tests__/displayName.spec.ts#document and display-surface invariants"
        status: pass
      - kind: other
        ref: "npm --prefix app run type-check and displayNameOf import grep gate"
        status: pass
    human_judgment: false
  - id: D3
    description: "Duplicate display names remain safe labels, username identity resolution is unchanged, and markup characters render literally through Vue interpolation."
    requirement: REQ-ACT-005
    verification:
      - kind: automated_ui
        ref: "app/__tests__/displayName.spec.ts#identity separation and escaped interpolation"
        status: pass
      - kind: other
        ref: "repo-wide app/src v-html absence gate"
        status: pass
    human_judgment: false

duration: 14min
completed: 2026-08-20
status: complete
---

# Phase 2 Plan 5: Display Name Sweep Summary

**Player display names are now snapshotted into game payloads and rendered through one tested fallback everywhere, while every identity operation remains keyed by unique username.**

## Performance

- **Duration:** 14 min
- **Started:** 2026-08-20T19:46:44Z
- **Completed:** 2026-08-20T20:00:58Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- Added non-blocking `users.display_name` lookups to create and join, preserving nullable truth in the stored player record and retaining compatibility with payloads written before the field existed.
- Added `DisplayName` to all five `Players` selection sets and matching handwritten response types, preventing the client fallback from being silently starved.
- Routed the app nav, score, games list/title, board headings, winner labels, and analysis chart labels through `displayNameOf`, with tests pinning fallback, duplicate-name identity separation, and Vue auto-escaping.

## Task Commits

Each TDD task has a RED test commit followed by its GREEN implementation commit:

| Task | Gate | Commit |
|------|------|--------|
| 1. Persist display names on create and join | RED | `39c3f93` |
| 1. Persist display names on create and join | GREEN | `db35882` |
| 2. Shared helper, documents, and display sweep | RED | `d3002c2` |
| 2. Shared helper, documents, and display sweep | GREEN | `c77a67c` |

## Files Created/Modified

- `server/games.go` — looks up and snapshots the authenticated creator's or joiner's nullable display name without failing the game on a cosmetic lookup error.
- `server/games_test.go` — covers typed and absent names, join preservation, legacy JSON, and duplicate-label identity safety against PostgreSQL.
- `app/src/services/displayName.ts` — defines the single rendering-only player-name fallback.
- `app/src/graphql/queries.ts` and `app/src/graphql/mutations.ts` — request `DisplayName` in all five player-bearing documents.
- `app/src/types/generated.ts` — adds the matching player and create/join response shapes.
- `AppNav.vue`, `ScoreView.vue`, `GamesView.vue`, `BoardView.vue`, and `GameAnalysisView.vue` — render player labels through the helper without changing layout, breakpoints, or username-based behavior.
- `app/__tests__/displayName.spec.ts` — pins fallback cases, document completeness, declared surface usage, duplicate labels, and escaped text rendering.

## Decisions Made

- Kept nullable storage honest: a missing display name stays nil in the game payload and only the rendering layer applies the username fallback.
- Kept display-name lookup failure non-terminal because the label is cosmetic; the wrapped error is logged while game creation or joining proceeds.
- Kept chart data and all board interactions keyed by username; only their human-facing labels use `displayNameOf`.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

The first self-check shell loop used zsh's special `path` variable name and temporarily hid
`PATH` inside that one process, producing false missing-commit messages. Re-running with a
task-specific variable confirmed all files and commits exist; no repository state was changed.
The installed state handlers require named flags for metrics and decisions; their initial
positional calls were rejected without mutation, and the corrected named-flag calls succeeded.
The progress handler reported 94% but left the legacy prose bar stale, so that prose and the
last-activity label were aligned to the handler's result.

## Known Stubs

- `app/src/views/ScoreView.vue:20` — the pre-existing commander-damage panel still says "Coming soon in v2." It is unrelated to player-name rendering and does not prevent this plan's goal; the player heading above it now uses `displayNameOf`.

## Verification

- `go build ./...` — pass.
- `go test ./server -run 'TestGames_' -race -v` — pass, including all four new display-name cases with PostgreSQL active.
- `go test ./server -race` — pass in 249.455 seconds.
- `npm --prefix app run test` — pass: 124 passed, 3 intentional skips.
- `npm --prefix app run type-check` — pass.
- Five-file `displayNameOf` import gate, five GraphQL document field gate, username identity grep gates, zero-`v-html` gate, and unchanged breakpoint counts — pass.
- Root `edhgo` build artifact — absent.

## User Setup Required

None - no dependencies, credentials, migrations, or external configuration were added.

## Next Phase Readiness

- Plan 02-06 can now verify the complete running guest journey knowing typed names and generated-name fallback reach every player-label surface.
- Phase 3's invite preview can reuse `displayNameOf` and the established `DisplayName` document-field invariant without changing identity semantics.
- No blockers remain.

## Self-Check: PASSED

All twelve plan files and the summary exist, and task commits `39c3f93`, `db35882`,
`d3002c2`, and `c77a67c` are present in git history.

---
*Phase: 02-guest-host-activation*
*Completed: 2026-08-20*

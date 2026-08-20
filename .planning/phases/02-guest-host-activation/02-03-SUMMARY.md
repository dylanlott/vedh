---
phase: 02-guest-host-activation
plan: 03
subsystem: activation-ui
tags: [vue, vitest, apollo, responsive, session-storage, deck-import]

requires:
  - phase: 01-measured-deck-import-foundation
    provides: previewDeck, canonical deck parsing, eager suggestions, and commander candidates
  - phase: 02-guest-host-activation
    plan: 01
    provides: public /play tracer, guest identity branch, and existing createGame reuse
  - phase: 02-guest-host-activation
    plan: 02
    provides: closed four-code activation failure contract
provides:
  - one paste-or-URL DeckImportPanel shared by /play, authenticated create, and join
  - commander review driven by the existing partner legality helpers
  - safe client recovery copy for all four activation error codes
  - tab-scoped quick-start draft survival with explicit success clearing
  - one mobile-first 768px activation breakpoint and a non-blocking phone heads-up
affects: [02-04-quick-start-completion, 02-06-human-verification, phase-03-guest-join, phase-04-board-responsiveness]

tech-stack:
  added: []
  patterns:
    - chrome-free shared input component publishes raw state, explicit corrections, and one prepared deck representation
    - closed server error codes resolve only to frozen author copy and never to source messages
    - activation components use one ascending SCSS breakpoint while legacy descending queries remain untouched

key-files:
  created:
    - app/src/styles/breakpoints.scss
    - app/src/services/activationErrors.ts
    - app/src/services/quickStartDraft.ts
    - app/src/components/decks/DeckImportPanel.vue
    - app/src/components/decks/CommanderReview.vue
    - app/src/components/decks/MobileHeadsUpNotice.vue
    - app/__tests__/DeckImportPanel.spec.ts
    - app/__tests__/CommanderReview.spec.ts
    - app/__tests__/activationErrors.spec.ts
    - app/__tests__/quickStartDraft.spec.ts
  modified:
    - app/src/styles/main.scss
    - app/src/views/QuickStartView.vue
    - app/src/components/games/FormCreateGame.vue
    - app/src/views/JoinGameView.vue
    - app/__tests__/QuickStartView.spec.ts
    - app/__tests__/FormCreateGame.integration.spec.ts
    - app/__tests__/JoinGame.integration.spec.ts

key-decisions:
  - "DeckImportPanel preserves raw text and explicit corrections separately while publishing one prepared Decklist string to every create/join consumer."
  - "DeckImportPanel owns the single preview-error region; QuickStartView owns only terminal guest-session and create failures, preventing duplicate failure copy."
  - "The reusable panel uses no form element so it can be embedded safely inside the existing authenticated create and join forms."

patterns-established:
  - "Activation error boundary: extensions.code selects frozen title/body/affordance copy; unknown shapes return one generic entry."
  - "Draft lifecycle: save raw deck, URL, corrections, commander picks, and display name on change; clear only after createGame succeeds."
  - "Responsive activation UI: base phone stack plus exactly one min-width query reading breakpoints.$breakpoint-tablet."

requirements-completed: [REQ-ACT-004]

coverage:
  - id: D1
    description: "DeckImportPanel renders paste/URL, loading, totals, unresolved correction, safe failure, Unicode, overflow, and stale-response states."
    requirement: REQ-ACT-004
    verification:
      - kind: automated_ui
        ref: "app/__tests__/DeckImportPanel.spec.ts#16 tests and app/__tests__/activationErrors.spec.ts#10 tests"
        status: pass
      - kind: other
        ref: "verify.key-links .planning/phases/02-guest-host-activation/02-03-PLAN.md"
        status: pass
    human_judgment: false
  - id: D2
    description: "Commander selection uses commanderPartner helpers, and every quick-start draft field survives refresh or create failure but is removed on success."
    requirement: REQ-ACT-004
    verification:
      - kind: automated_ui
        ref: "app/__tests__/CommanderReview.spec.ts#7 tests and app/__tests__/QuickStartView.spec.ts#11 tests"
        status: pass
      - kind: unit
        ref: "app/__tests__/quickStartDraft.spec.ts#8 tests"
        status: pass
    human_judgment: false
  - id: D3
    description: "Authenticated create, join, and /play now consume the same shared panel and forward the same prepared deck representation."
    requirement: REQ-ACT-004
    verification:
      - kind: integration
        ref: "app/__tests__/FormCreateGame.integration.spec.ts and app/__tests__/JoinGame.integration.spec.ts#4 local component assertions"
        status: pass
      - kind: other
        ref: "zero quantity,name per line matches across app/src"
        status: pass
    human_judgment: false
  - id: D4
    description: "Activation components reflow from a phone stack to tablet layout at the shared 768px breakpoint, with the phone heads-up visible only below it."
    requirement: REQ-ACT-004
    verification:
      - kind: other
        ref: "static media-query and shared-breakpoint gates"
        status: pass
    human_judgment: true
    rationale: "Visual overflow, two-line notice wrapping, and breakpoint composition require the running-browser review assigned to plan 02-06."

duration: 22min
completed: 2026-08-20
status: complete
---

# Phase 2 Plan 3: Reusable Deck Import and Commander Review Summary

**One paste-or-URL deck panel now serves all three activation call sites, with safe four-code recovery, partner-aware commander review, and tab-scoped draft survival across refreshes and failed creates.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-08-20T18:45:01Z
- **Completed:** 2026-08-20T19:07:15Z
- **Tasks:** 3
- **Files modified:** 17

## Accomplishments

- Built the shared deck-import surface with paste/URL source selection, bounded loading and preview states, explicit unresolved-card corrections, safe activation errors, Unicode-safe display truncation, and stale-response rejection.
- Added partner-aware commander review, the phone heads-up, and a defensive session draft that preserves raw deck text, corrections, commander picks, and display name while clearing only after a successful table creation.
- Removed both stale CSV-only textareas and routed `/play`, authenticated create, and join through the same prepared deck representation without changing their surrounding payload fields.

## Task Commits

Each TDD task has a RED test commit followed by its GREEN implementation commit:

| Task | Gate | Commit |
|------|------|--------|
| 1. Responsive baseline, failure contract, and DeckImportPanel | RED | `f8147e9` |
| 1. Responsive baseline, failure contract, and DeckImportPanel | GREEN | `773d599` |
| 2. Commander review, phone notice, and draft survival | RED | `acf7891` |
| 2. Commander review, phone notice, and draft survival | GREEN | `2a52210` |
| 3. Shared authenticated create and join call sites | RED | `8b6a78c` |
| 3. Shared authenticated create and join call sites | GREEN | `8cf74a1` |
| Plan contract cleanup | REFACTOR | `9654094` |

## Files Created/Modified

- `app/src/components/decks/DeckImportPanel.vue` — shared text/URL preview, totals, correction, recovery, and prepared-deck surface.
- `app/src/components/decks/CommanderReview.vue` — bounded candidate and manual-search UI delegating all partner legality to `commanderPartner.ts`.
- `app/src/components/decks/MobileHeadsUpNotice.vue` — fixed, non-dismissible phone expectation notice.
- `app/src/services/activationErrors.ts` — frozen four-code and generic safe-copy resolver.
- `app/src/services/quickStartDraft.ts` — defensive typed tab-scoped draft lifecycle.
- `app/src/styles/breakpoints.scss` and `app/src/styles/main.scss` — shared tablet breakpoint and attention-only danger token.
- `app/src/views/QuickStartView.vue` — staged import, commander, display-name, failure-preservation, and success-clearing orchestration.
- `app/src/components/games/FormCreateGame.vue` and `app/src/views/JoinGameView.vue` — shared panel consumers retaining their existing game payload shapes.
- Seven frontend specs — 56 plan-focused passing assertions plus the preserved opt-in live integration cases.

## Decisions Made

- Kept raw input immutable and corrections explicit, while deriving a prepared deck string at the panel boundary so every consumer receives identical data without reimplementing parsing or correction behavior.
- Kept preview failures inside `DeckImportPanel` and terminal guest/create failures inside `QuickStartView`, giving every failure exactly one visible recovery block.
- Made the shared panel form-free because two of its three consumers already own submission forms; preview remains an explicit button action and Enter on the URL field remains supported.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Test Bug] Exercised stale-response handling through a disabled control**
- **Found during:** Task 1 GREEN verification
- **Issue:** The initial concurrency test attempted a second click after loading had disabled the submit button, so it could not create the two in-flight requests the behavior names.
- **Fix:** Exercised the exposed submit boundary programmatically for the second request while retaining the user-facing disabled state.
- **Files modified:** `app/__tests__/DeckImportPanel.spec.ts`
- **Verification:** The second response renders and the later-resolving first response is discarded.
- **Committed in:** `773d599`, finalized for the form-free panel in `8cf74a1`

**2. [Rule 2 - Missing Critical] Removed nested forms from the reusable component boundary**
- **Found during:** Task 3 embedding into authenticated create and join
- **Issue:** A panel-owned `<form>` would be nested inside both existing parent forms, creating invalid HTML and ambiguous submit behavior.
- **Fix:** Converted the panel submission boundary to a form-free explicit preview action, preserving URL Enter submission and updating component tests.
- **Files modified:** `app/src/components/decks/DeckImportPanel.vue`, `app/__tests__/DeckImportPanel.spec.ts`, `app/__tests__/QuickStartView.spec.ts`
- **Verification:** Component/integration tests pass, and `DeckImportPanel.vue` contains no form tag.
- **Committed in:** `8cf74a1`

**3. [Rule 2 - Missing Critical] Published a corrected prepared deck representation**
- **Found during:** Task 3 payload wiring
- **Issue:** Forwarding raw paste text alone would discard explicit correction choices and leave URL consumers without a Decklist string for the existing create/join mutations.
- **Fix:** Added the derived `decklist` field to the panel change event while retaining raw text, source URL, and corrections separately; all three parents consume the same derivation.
- **Files modified:** `app/src/components/decks/DeckImportPanel.vue`, `app/src/components/games/FormCreateGame.vue`, `app/src/views/JoinGameView.vue`
- **Verification:** Both local integration specs assert identical deck strings in the unchanged host/join payload shapes.
- **Committed in:** `8cf74a1`

**4. [Rule 1 - Contract Bug] Removed an extra correction prop from DeckImportPanel**
- **Found during:** Plan-wide interface audit
- **Issue:** The first draft-restoration implementation exposed `initialCorrections`, widening DEC-J's declared public inputs.
- **Fix:** Restored corrections through the already-declared persistence key and kept the public props to initial text, initial URL, session identifier, and persistence key.
- **Files modified:** `app/src/components/decks/DeckImportPanel.vue`, `app/src/views/QuickStartView.vue`
- **Verification:** Draft restoration tests, all panel/call-site tests, and type-check pass.
- **Committed in:** `9654094`

---

**Total deviations:** 4 auto-fixed (2 Rule 1, 2 Rule 2).
**Impact on plan:** All fixes enforce the planned reuse, persistence, and interface contracts without expanding product scope.

## Issues Encountered

None beyond the auto-fixed test and embedding issues above.

## Known Stubs

None. Empty arrays and nullable preview/error state in the changed files are intentional initialized runtime state, not unwired UI data.

## Verification

- `npm --prefix app run test` — pass: 89 passed, 3 intentional skips.
- `npm --prefix app run type-check` — pass.
- Plan-focused seven-spec run — pass: 56 passed, 2 opt-in live integration cases skipped.
- `verify.key-links` — 5/5 plan links verified.
- Shared breakpoint/media-query, no-raw-HTML, labelled icon-control, no-persistent-storage, legacy-query-untouched, and stale-copy-zero gates — pass.
- Root `edhgo` build artifact — absent.

## User Setup Required

None — no dependencies, credentials, or external configuration were added.

## Next Phase Readiness

- Plan 02-04 can add authoritative game-create telemetry and the three client funnel events on top of the now-stable quick-start orchestration.
- Plan 02-06 retains the deliberate running-browser visual review for 1024px, 768px, and phone layouts; no implementation blocker remains.

## Self-Check: PASSED

All key created artifacts exist, and all seven RED/GREEN/refactor task commits are present in git history.

---
*Phase: 02-guest-host-activation*
*Completed: 2026-08-20*

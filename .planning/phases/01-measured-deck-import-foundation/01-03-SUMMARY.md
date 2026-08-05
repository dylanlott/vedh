---
phase: 01-measured-deck-import-foundation
plan: 03
subsystem: api
tags: [vue3, apollo-client, vitest, jsdom, product-analytics, telemetry, frontend]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 01-01)
    provides: "trackProductEvent mutation, InputProductEvent/InputProductEventMeta types, server-side event/field allowlists, regenerated app/src/types/generated.ts"
provides:
  - "app/src/services/productEvents.ts: getSessionID() (D-18, persisted under edhgo/session-id), track() (D-19, fire-and-forget, one mutation per call), and captureAttribution() (T-01-17, allowlisted/trimmed/byte-bounded UTM + referrer-host)"
  - "TRACK_PRODUCT_EVENT_MUTATION gql document in app/src/graphql/mutations.ts"
  - "app/__tests__/productEvents.spec.ts: 20-case vitest suite covering session-ID stability/degradation, fire-and-forget semantics, and attribution allowlisting"
affects: [Phase 2 ACT-004 (activation UI calls track()/getSessionID() from components), Phase 2 ACT-005/ACT-006 and Phase 3 ACT-008/ACT-009 (client emit sites for board_ready/join_started/etc. reuse this module), 01-06 (rate-limit plan also claims REQ-ACT-001; does not touch this file)]

# Actuals (#2632)
actuals:
  tokens: 5000
  tasks: 2
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Defensive storage/location/referrer accessors mirroring app/src/services/apollo.ts's getRawAuth shape: try the direct global, then window-scoped, return null/empty rather than propagating — the model for any client service that must never throw"
    - "Per-event attribution table as a documented convenience, never a control — the file-header comment and the table's own comment both name pkg/telemetry.Vocabulary as the actual enforcement point"
    - "Codepoint-aware byte truncation (Array.from + TextEncoder) so a UTF-8 multi-byte sequence is never split, reusable anywhere a client needs to match a server byte limit"

key-files:
  created:
    - app/src/services/productEvents.ts
    - app/__tests__/productEvents.spec.ts
  modified:
    - app/src/graphql/mutations.ts

key-decisions:
  - "REQ-ACT-001 is not marked complete in REQUIREMENTS.md despite this plan finishing its named frontend piece ('frontend event service with a persisted random session ID') — the requirement is also claimed by plan 01-06, which has not yet executed. Following the precedent set in 01-01-SUMMARY.md's deviation record, the checkbox is left for whichever of 01-03/01-06 lands last."
  - "The four attribution keys docs/analytics/product-event-vocabulary.md records (utm_source, utm_medium, utm_campaign, referrer_host) are implemented as two separate constants — QUERY_ATTRIBUTION_KEYS (the three UTM keys, read from the URL) and a standalone REFERRER_ATTRIBUTION_KEY — because referrer_host is derived from document.referrer, never from the query string, and conflating them into one frozen array would misrepresent where each value actually comes from."
  - "captureAttribution() memoises its result in a module-level variable (cachedAttribution) rather than accepting a reset function, so 'once per page view' is enforced by the module's own lifetime rather than by caller discipline."

patterns-established:
  - "Attribution keys captured once per module lifetime (page view) and reused across every subsequent track() call in that same page view, per T-01-17"

requirements-completed: [REQ-ACT-001]

coverage:
  - id: D1
    description: "getSessionID() persists one random identifier under edhgo/session-id, is stable across repeated calls and a simulated module reload against the same storage, and degrades to a fresh unpersisted identifier (never throwing) when storage is absent, throws, or holds an empty string"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "app/__tests__/productEvents.spec.ts#getSessionID"
        status: pass
    human_judgment: false
  - id: D2
    description: "newSessionID() sources from crypto.randomUUID() first, falls through to crypto.getRandomValues() hex-encoded to 32 characters when randomUUID is absent, and only then to a non-cryptographic composition"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "app/__tests__/productEvents.spec.ts#getSessionID > falls back to the random-bytes API when randomUUID is absent from the crypto global"
        status: pass
    human_judgment: false
  - id: D3
    description: "track() sends exactly one apolloClient.mutate() per call, returns undefined synchronously (never a thenable), is never batched, and swallows a rejected mutation at console.debug level without throwing or reaching console.error; it never reads the edhgo/auth storage key"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "app/__tests__/productEvents.spec.ts#track"
        status: pass
    human_judgment: false
  - id: D4
    description: "captureAttribution() reads only the allowlisted utm_source/utm_medium/utm_campaign query keys plus a referrer reduced to its lower-cased host, trims and byte-truncates every value to at most 128 bytes on a codepoint boundary, drops empty-after-trim keys, memoises the result once per page view, and is wired into track() only for the two events (landing_primary_cta, invite_viewed) the vocabulary document marks as accepting it"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "app/__tests__/productEvents.spec.ts#captureAttribution"
        status: pass
    human_judgment: false

duration: ~8min
completed: 2026-08-04
status: complete
---

# Phase 1 Plan 3: Client Product-Event Service — Session ID, Fire-and-Forget Track, Attribution Allowlist Summary

**A plain-module `productEvents.ts` gives every browser a persisted random session ID and a fire-and-forget `track()` that attaches only allowlisted, trimmed, 128-byte-bounded campaign attribution to the two events the server vocabulary permits — no component, no store, no UI.**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-08-04T22:32:20-06:00
- **Completed:** 2026-08-04T22:39:22-06:00
- **Tasks:** 2/2
- **Files modified:** 3 (2 created, 1 modified)

## Accomplishments

- `getSessionID()` persists a per-browser random identifier under `edhgo/session-id`, proven stable across repeated calls and a simulated module reload against the same storage, and proven to degrade — never throw — when storage is absent, throws on access, or holds an empty string.
- `newSessionID()` sources from `crypto.randomUUID()` first, falls through to `crypto.getRandomValues()` hex-encoded to a 32-character string when `randomUUID` is unavailable (the LAN-IP-over-HTTP dev case), and only then to a non-cryptographic timestamp/pseudorandom composition that is documented as never relied upon for unpredictability.
- `track(name, metadata)` sends exactly one `apolloClient.mutate()` per call, is never awaited by the caller, returns `undefined` synchronously, and swallows a rejected mutation at `console.debug` level under the `[productEvents]` prefix — proven to never throw, never surface as an unhandled rejection, and never reach `console.error`.
- `captureAttribution()` reads only the three allowlisted UTM query-string keys and reduces the referrer to its lower-cased host, trims every value, truncates to at most 128 bytes on a codepoint boundary (never splitting a multi-byte UTF-8 sequence), drops empty-after-trim keys, and memoises the result once per page view so repeated `track()` calls share the same values rather than re-parsing the URL.
- `track()` merges `captureAttribution()`'s entries only for `landing_primary_cta` and `invite_viewed` — the two events `docs/analytics/product-event-vocabulary.md` marks as accepting attribution — via a small table whose own comment names `pkg/telemetry.Vocabulary` as the actual enforcing control, not this client-side convenience.
- The service never reads or writes the `edhgo/auth` storage key and never places an authorization value into event metadata, verified both by a dedicated test spying on the storage getter and by a direct `grep` for the string.
- No component, store, or route was added — this plan ships zero activation UI, matching `ui.plan-gate`'s `frontend: false` verdict for this phase.

## Task Commits

Each task was committed atomically, following the RED/GREEN TDD cycle:

1. **Task 1: Persisted session identifier and fire-and-forget event emission**
   - `11b2bbe` (test) — RED: `app/__tests__/productEvents.spec.ts` written covering both this task's and task 2's behavior (see Deviations), fails to import because the service does not exist yet.
   - `18f3af0` (feat) — GREEN: `app/src/services/productEvents.ts` (`getSessionID`, `track`) and `TRACK_PRODUCT_EVENT_MUTATION` in `app/src/graphql/mutations.ts`. All `getSessionID`/`track` cases pass; `captureAttribution` cases fail with "is not a function," expected pending task 2.
2. **Task 2: Allowlisted, normalized, length-bounded campaign attribution**
   - `9324d34` (feat) — GREEN: `captureAttribution()` added and wired into `track()`. All 20 cases in the spec pass; `npm test` and `npm run type-check` are green.

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `app/src/services/productEvents.ts` — plain module, named exports only (`getSessionID`, `track`, `captureAttribution`), no default export, no Pinia
- `app/src/graphql/mutations.ts` — added `TRACK_PRODUCT_EVENT_MUTATION`, following the existing `LOGIN_MUTATION` shape
- `app/__tests__/productEvents.spec.ts` — 20-test vitest suite (284 lines) covering session-ID stability/degradation, fire-and-forget semantics, and attribution allowlisting/truncation/referrer-reduction

## Decisions Made

- `REQ-ACT-001` is **not** flipped to complete in `REQUIREMENTS.md`, even though this plan finishes the requirement's named "frontend event service with a persisted random session ID" clause. The requirement is also claimed by plan `01-06` (not yet executed), following the exact precedent `01-01-SUMMARY.md` recorded for the same requirement. `requirements-completed: [REQ-ACT-001]` in this SUMMARY's frontmatter documents what this plan *contributed*, per the template's instruction, not a project-wide completion claim.
- The vocabulary document's "four allowlisted attribution keys" are implemented as two separate constructs — a frozen `QUERY_ATTRIBUTION_KEYS` array (the three UTM keys, read from the URL) and a standalone `REFERRER_ATTRIBUTION_KEY` constant (derived from `document.referrer`, never the URL) — because the two have genuinely different sources and conflating them into one array would misstate where `referrer_host` actually comes from.
- Attribution memoization uses a plain module-level variable rather than an exported reset function, so "captured once per page view" is a property of the module's own lifetime, not something a caller could accidentally defeat.

## Deviations from Plan

### Process adaptation (not a Rule 1-4 deviation)

**The spec file was written as a single comprehensive RED commit covering both task 1's and task 2's `<behavior>` cases, rather than task 1's cases first and task 2's as a later extension.** The plan's task 2 action explicitly says "Extend `app/__tests__/productEvents.spec.ts`" — implying an incremental two-step RED. Writing all 20 cases up front meant task 1's own verification run (`npx vitest --run __tests__/productEvents.spec.ts`) showed 9 pre-existing failures for `captureAttribution`, which does not exist until task 2. This was not treated as a task-1 failure: the 11 `getSessionID`/`track` cases were confirmed green in isolation (via `-t` filtering) before committing task 1's `feat` commit, and the full 20/20 pass was confirmed before committing task 2's `feat` commit. Net result is unchanged — one `test` commit followed by two `feat` commits, full suite green at the end — but the RED commit is shared across both tasks rather than split. Documented here rather than silently normalized, per Rule-adjacent transparency; no code behavior was affected.

**1. [Rule 1 - Bug] Test helper assumed the wrong jsdom default origin**
- **Found during:** Task 1, first RED run
- **Issue:** The `setLocationSearch` test helper built URLs against `http://localhost/`, but this project's vitest jsdom environment defaults to `http://localhost:3000/`. `window.history.pushState` throws `SecurityError: pushState cannot update history to a URL which differs in components other than in path, query, or fragment` when the origin doesn't match, failing all 20 tests with the same error regardless of which behavior they targeted.
- **Fix:** Changed the helper to build the URL from `window.location.origin` instead of a hardcoded string.
- **Files modified:** `app/__tests__/productEvents.spec.ts`
- **Verification:** Re-ran the full spec; the origin-mismatch error disappeared and each test failed/passed on its own merits.
- **Committed in:** `18f3af0` (Task 1 `feat` commit, bundled with the implementation since it was discovered while first exercising the RED spec against the real implementation)

---

**Total deviations:** 1 auto-fixed test-infrastructure bug, 1 documented process adaptation (no rule violation)
**Impact on plan:** Neither affects `productEvents.ts` production code or its behavior. No scope creep.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None — no external service configuration required. This plan touches only frontend files; no server, database, or environment variable changes.

## Next Phase Readiness

- Phase 2's activation UI (ACT-004 and later) can call `getSessionID()` and `track()` directly from Vue components without any further wiring — the module is complete, tested, and exports exactly what the artifact spec requires (`getSessionID`, `track`, `captureAttribution`).
- Plan 01-06 still owns the remainder of `REQ-ACT-001` (rate-limit instrumentation); this plan's frontend piece is done and does not block it.
- No blockers. `cd app && npm test` and `cd app && npm run type-check` are both green with this plan's changes in place.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-04*

## Self-Check: PASSED

Both created files (`app/src/services/productEvents.ts`, `app/__tests__/productEvents.spec.ts`) exist on disk, `app/src/graphql/mutations.ts` contains `TRACK_PRODUCT_EVENT_MUTATION`, and all three task commit hashes (`11b2bbe`, `18f3af0`, `9324d34`) are present in `git log --oneline --all`.

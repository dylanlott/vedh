---
phase: 02-guest-host-activation
plan: 01
subsystem: guest-auth-and-activation
tags: [graphql, postgres, jwt, bcrypt, vue, pinia, vitest, rate-limiting]

requires:
  - phase: 01-measured-deck-import-foundation
    provides: server-side deck preview, product-event recording, telemetry vocabulary, and shared rate limiting
provides:
  - durable guest user schema and public guestSession mutation
  - public /play paste-to-preview-to-board tracer using the existing createGame mutation
  - separate browser persistence for the guest re-auth credential
affects: [02-02-guest-identity-backend, 02-03-deck-import-ui, 02-04-quick-start-completion, 02-05-display-name-sweep]

tech-stack:
  added: []
  patterns:
    - public router metadata with authenticated-versus-guest branching inside the view
    - high-entropy re-auth credential stored separately from JWT and analytics identity
    - kill switch and shared per-surface rate limit checked before anonymous database writes

key-files:
  created:
    - server/guest_names.go
    - server/guest_users.go
    - server/guest_users_test.go
    - app/src/views/QuickStartView.vue
    - app/__tests__/QuickStartView.spec.ts
    - app/__tests__/authGuest.spec.ts
    - persistence/migrations/20260809120000_guest_users.up.sql
    - persistence/migrations/20260809120000_guest_users.down.sql
  modified:
    - server/graphql.go
    - server/schema.graphql
    - pkg/ratelimit/limiter.go
    - app/src/router/index.ts
    - app/src/stores/auth.ts

key-decisions:
  - "Guest creation defaults enabled; GUEST_CREATION_ENABLED is an operator kill switch, not a rollout gate."
  - "Guest rows are durable and leave expires_at NULL; the 24-hour JWT remains the expiring session boundary."
  - "Guest creation uses the existing shared rate-limit registry through SurfaceGuestSession."
  - "The guest re-auth credential is a distinct bearer secret under edhgo/guest-credential, never part of edhgo/auth or edhgo/session-id."

patterns-established:
  - "Anonymous write gate order: validate input, check kill switch, rate-limit, then write."
  - "Generated guest usernames share the real username namespace and retry database uniqueness conflicts server-side."
  - "Public activation routes branch on authentication state inside the orchestrating view."

requirements-completed: [REQ-ACT-005, REQ-ACT-006]

coverage:
  - id: D1
    description: "A logged-out visitor can use /play to preview a pasted deck, receive a guest identity, create a game through the existing mutation, and navigate to the live board."
    requirement: REQ-ACT-006
    verification:
      - kind: integration
        ref: "server/guest_users_test.go#TestGuestHost_Tracer"
        status: pass
      - kind: automated_ui
        ref: "app/__tests__/QuickStartView.spec.ts#6 tests"
        status: pass
      - kind: other
        ref: "go build ./... and npm --prefix app run type-check"
        status: pass
    human_judgment: false
  - id: D2
    description: "Guest identities are durable, uniquely named, collision-safe, rate-limited, kill-switchable, and recorded once with the expected schema."
    requirement: REQ-ACT-005
    verification:
      - kind: integration
        ref: "server/guest_users_test.go#TestGuestUsers_* and TestMigrations_GuestUsers"
        status: pass
      - kind: other
        ref: "production/test migration byte comparisons"
        status: pass
    human_judgment: false
  - id: D3
    description: "The guest re-auth credential is persisted separately from the JWT auth profile and analytics session ID and is cleared on logout."
    requirement: REQ-ACT-005
    verification:
      - kind: unit
        ref: "app/__tests__/authGuest.spec.ts#5 tests"
        status: pass
    human_judgment: false

duration: 6min
completed: 2026-08-20
status: complete
---

# Phase 2 Plan 1: Guest Host Activation Tracer Summary

**A public `/play` tracer now mints a guarded guest identity, previews a pasted deck, reuses `createGame` to reach a live board, and isolates the durable re-auth credential from both JWT and analytics storage.**

## Performance

- **Recovery duration:** 6 min
- **Recovery started:** 2026-08-20T17:54:31Z
- **Completed:** 2026-08-20T18:00:24Z
- **Tasks:** 2
- **Files modified:** 21 unique files across both task commits

Task 1 had already been implemented and committed on 2026-08-09. This recovery validated that commit in place and completed only Task 2; the duration above measures the safe-resume session rather than the earlier Task 1 implementation effort.

## Accomplishments

- Added durable guest columns, a rate-limited and kill-switchable `guestSession` mutation, curated collision-safe MTG guest names, a 24-hour JWT, and a separately hashed re-auth credential.
- Added the public `/play` quick-start view, where both logged-out and authenticated visitors preview a deck and reuse the existing `createGame` path to reach `/games/:id`.
- Persisted the guest re-auth credential under `edhgo/guest-credential`, kept it out of `edhgo/auth`, left `edhgo/session-id` untouched, and cleared it on logout.

## Task Commits

| Task | Name | Commit |
|------|------|--------|
| 1 | End-to-end logged-out paste-to-live-board tracer | `bc679ae` |
| 2 | Separate guest re-auth credential persistence | `8e3c3a4` |

## Files Created/Modified

- `persistence/migrations/20260809120000_guest_users.{up,down}.sql` and test twins — add and reverse guest identity columns with byte-identical production/test migrations.
- `server/guest_names.go` — curated adjective/noun generation using `crypto/rand` with collision overflow support.
- `server/guest_users.go` — guarded guest creation, password and re-auth-secret hashing, JWT issuance, uniqueness retry, and product-event recording.
- `server/guest_users_test.go` — real-PostgreSQL tracer, migration, collision, kill-switch, rate-limit, display-name, and dedup proofs.
- `app/src/views/QuickStartView.vue` — public paste, preview, identity branch, create, and board navigation flow.
- `app/src/stores/auth.ts` — guest session action plus defensive, separate credential storage helpers.
- `app/__tests__/QuickStartView.spec.ts` — six client tracer cases.
- `app/__tests__/authGuest.spec.ts` — five credential isolation and cleanup cases.

## Decisions Made

- `GUEST_CREATION_ENABLED` defaults to true because it controls an internal guarded write; it remains available as an operator abuse-response kill switch.
- Guest rows are durable with `expires_at = NULL`; JWTs expire after 24 hours and later re-auth uses the separate credential.
- Guest creation joins the existing rate-limit registry as `SurfaceGuestSession`, avoiding a second registry with fragmented eviction and metrics.
- Re-auth credentials use their own browser key and are never persisted in the ordinary auth profile or reused as the analytics session identifier.

## Deviations from Plan

### Recovery Workflow

**1. [Recovery - Existing Task Commit] Adopted and validated Task 1 instead of re-executing it**
- **Found at recovery start:** `bc679ae` was already `HEAD` and contained the complete Task 1 production slice, but no plan summary existed.
- **Action:** Preserved the commit, inspected its exact file set, started the repository PostgreSQL service, and reran every Task 1 automated verification before continuing.
- **Impact:** No Task 1 source was redone, reverted, or amended.

**2. [Recovery - TDD Commit Shape] Adopted the untracked Task 2 test into one atomic task commit**
- **Found at recovery start:** `app/__tests__/authGuest.spec.ts` existed untracked.
- **Action:** Ran it before implementation and observed the expected RED result (4 failures, 1 pass), then committed the recovered test and minimal implementation together as explicitly required by the safe-resume directive.
- **Impact:** Behavioral RED/GREEN evidence was preserved, while the recovery used one Task 2 commit rather than separate test and feature commits.

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Made the Task 1 test harness honor the production guest-creation default**
- **Found during:** Existing Task 1 implementation
- **Issue:** `testAPI` constructs `Conf` directly, so the envconfig `default:"true"` tag never runs and Go's false zero value would incorrectly disable every guest test.
- **Fix:** `server/test.go` explicitly sets `GuestCreationEnabled: true`; the kill-switch test overrides it to false.
- **Verification:** Full targeted guest/migration suite passes against PostgreSQL.
- **Committed in:** `bc679ae`

**2. [Rule 2 - Missing Critical] Rejected a successful-looking guest response without a re-auth credential**
- **Found during:** Task 2
- **Issue:** The GraphQL `User.GuestCredential` field is nullable even though `guestSession` must return it; persisting an auth profile without the credential would create an unrecoverable guest session.
- **Fix:** `createGuestSession` now rejects an empty credential before writing either credential or profile state.
- **Verification:** `authGuest.spec.ts` and `vue-tsc --noEmit` pass.
- **Committed in:** `8e3c3a4`

---

**Total deviations:** 2 recovery adaptations and 2 auto-fixes (1 Rule 2, 1 Rule 3).
**Impact on plan:** Recovery preserved the intended architecture and atomic task history without expanding scope.

## Issues Encountered

- PostgreSQL was initially stopped, producing a real connection-refused failure rather than a skipped pass. `make persistence` started the repository service, after which the full `-race` suite passed.

## Known Stubs

- `server/schema.resolvers.go:68` — gqlgen retains an intentional `GuestSession` panic stub on `mutationResolver`; it is dead code because the running server's `*graphQLServer.GuestSession` method implements the mutation resolver directly, matching the existing custom resolver pattern and proven by `TestGuestHost_Tracer`.

## Verification

- `go build ./...` — pass
- `go test ./server -run 'TestGuestHost_Tracer|TestGuestUsers_|TestMigrations_GuestUsers' -race -v` — pass against PostgreSQL, including explicit tracer and migration passes
- Production/test guest migration `.up.sql` and `.down.sql` byte comparisons — pass
- `npm --prefix app run test -- __tests__/QuickStartView.spec.ts` — 6/6 pass
- `npm --prefix app run test -- __tests__/authGuest.spec.ts` — 5/5 pass
- `npm --prefix app run type-check` — pass

## User Setup Required

None — no packages, credentials, or external services were added.

## Next Phase Readiness

- Plan 02-02 can consume `readGuestCredential()` to implement silent guest JWT re-issue and account-claim backend behavior.
- Plan 02-03 can expand the tracer UI into reusable deck import and commander review without changing the public-route or identity seams.
- Plan 02-06 still owns human review of the full curated adjective × noun guest-name product and the running-system visual journey.

## Self-Check: PASSED

All key created files exist, and task commits `bc679ae` and `8e3c3a4` are present in git history.

---
*Phase: 02-guest-host-activation*
*Completed: 2026-08-20*

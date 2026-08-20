---
phase: 02-guest-host-activation
plan: 02
subsystem: guest-identity-backend
tags: [graphql, postgres, jwt, bcrypt, gqlerror, account-claim, cleanup]

requires:
  - phase: 02-guest-host-activation
    plan: 01
    provides: durable guest rows, guestSession, separate re-auth credentials, product-event recording, and game access through JWT identity
provides:
  - closed four-code activation error contract with product-safe GraphQL messages
  - rate-limited silent guest-session refresh with row-expiry minting checks
  - same-row guest account claim preserving UUID and game access
  - conservative, retryable, unscheduled expired-guest cleanup safety valve
affects: [02-03-deck-import-ui, 02-04-quick-start-completion, 02-05-display-name-sweep, phase-04-account-claim-ui, secure-phase]

tech-stack:
  added: []
  patterns:
    - GraphQL activation failures expose closed server-owned codes while retaining internal causes only for logs and errors.Unwrap
    - bearer refresh checks database expiry only when minting a new token, leaving requireAuth JWT-only
    - privilege-changing guest claims re-query identity state and update the existing row with a concurrent is_guest guard

key-files:
  created:
    - server/activation_errors.go
    - server/activation_errors_test.go
    - server/guest_cleanup.go
    - server/guest_cleanup_test.go
    - docs/analytics/activation-error-codes.md
  modified:
    - server/guest_users.go
    - server/guest_users_test.go
    - server/users.go
    - server/users_test.go
    - server/deck_import.go
    - server/schema.graphql
    - server/generated.go
    - server/schema.resolvers.go

key-decisions:
  - "Activation failures expose exactly four allowlisted codes; internal provider, SQL, crypto, hash, and token causes never enter browser copy."
  - "Guest refresh retains the existing bearer credential and checks expires_at only before minting a fresh 24-hour JWT; a still-live JWT remains authorized."
  - "Guest claims update the authenticated row in place with an is_guest concurrency guard, clear all guest-only fields, and preserve the UUID used by existing games."
  - "Cleanup is a manually invoked safety valve bounded by guest status, explicit expiry, and absence from every game payload; production guests remain unscheduled and immortal."

patterns-established:
  - "Activation error boundary: product Message plus closed extensions.code, with the original error reachable through errors.Unwrap only."
  - "Guest bearer exchange: parse and validate the UUID-shaped lookup key before database or bcrypt work, then collapse all negative cases to one public response."
  - "Destructive guest maintenance: one idempotent DELETE with independent identity, expiry, and game-reference guards and no in-process scheduler."

requirements-completed: [REQ-ACT-005]

coverage:
  - id: D1
    description: "Activation failures use an exact four-code vocabulary, guest password login is refused generically, and display names are bounded and control-character safe."
    requirement: REQ-ACT-005
    verification:
      - kind: unit
        ref: "server/activation_errors_test.go#TestActivationErrors and TestActivationErrors_ClosedVocabulary"
        status: pass
      - kind: integration
        ref: "server/users_test.go#TestUsers_LoginRejectsGuest and server/guest_users_test.go#TestGuestUsers_DisplayNameValidation"
        status: pass
      - kind: other
        ref: "docs/analytics/activation-error-codes.md exact four-code comparison"
        status: pass
    human_judgment: false
  - id: D2
    description: "A valid guest bearer credential silently reissues a 24-hour JWT while expired rows cannot mint and live JWT authorization remains independent of row expiry."
    requirement: REQ-ACT-005
    verification:
      - kind: integration
        ref: "server/guest_users_test.go#TestGuestUsers_Refresh and TestGuestUsers_ExpiredRowStillAuthorized"
        status: pass
      - kind: other
        ref: "git diff --exit-code server/authz.go"
        status: pass
    human_judgment: false
  - id: D3
    description: "A guest can claim the same user row without losing game access, and explicitly expired unreferenced guests can be cleaned conservatively and idempotently."
    requirement: REQ-ACT-005
    verification:
      - kind: integration
        ref: "server/guest_users_test.go#TestGuestUsers_Claim"
        status: pass
      - kind: integration
        ref: "server/guest_cleanup_test.go#TestGuestUsers_Cleanup*"
        status: pass
      - kind: other
        ref: "grep scheduling-primitive gate for server/guest_cleanup.go"
        status: pass
    human_judgment: false

duration: 32min
completed: 2026-08-20
status: complete
---

# Phase 2 Plan 2: Guest Identity Backend Summary

**Closed activation errors, rate-limited bearer refresh, same-row account claim, and guarded cleanup now complete the durable guest identity backend without weakening JWT authorization.**

## Performance

- **Duration:** 32 min
- **Started:** 2026-08-20T18:09:11Z
- **Completed:** 2026-08-20T18:40:46Z
- **Tasks:** 3
- **Files modified:** 15

## Accomplishments

- Introduced four documented activation error codes with safe product copy, protected password login from guest rows, and enforced server-side display-name bounds.
- Added silent guest JWT refresh using one UUID lookup, one bcrypt comparison, shared rate limiting, and an exact database expiry boundary that does not alter `requireAuth`.
- Added authenticated same-row account claim with preserved UUID/game access, idempotent authoritative telemetry, and a conservative unscheduled cleanup path for operator-marked expired guests.

## Task Commits

Each TDD task has a RED test commit followed by its GREEN implementation commit:

| Task | Gate | Commit |
|------|------|--------|
| 1. Activation failure contracts, login hardening, and display-name validation | RED | `f9a3d2b` |
| 1. Activation failure contracts, login hardening, and display-name validation | GREEN | `a7b4546` |
| 2. Silent guest-session refresh | RED | `423aaca` |
| 2. Silent guest-session refresh | GREEN | `d03259a` |
| 3. Account claim and retryable cleanup | RED | `f5b90cf` |
| 3. Account claim and retryable cleanup | GREEN | `a248535` |

## Files Created/Modified

- `server/activation_errors.go` and `server/activation_errors_test.go` — closed error vocabulary, safe constructors, unwrap behavior, and closure tests.
- `docs/analytics/activation-error-codes.md` — exact browser-facing code and recovery-action contract.
- `server/guest_users.go` and `server/guest_users_test.go` — bounded display names, credential refresh, expiry boundaries, account claim, telemetry, and integration tests.
- `server/guest_cleanup.go` and `server/guest_cleanup_test.go` — guarded one-statement cleanup with preservation, idempotency, real-user, and production-shape proofs.
- `server/users.go` and `server/users_test.go` — generic password-login refusal for guest rows.
- `server/deck_import.go`, `server/deck_import_test.go`, and `server/deck_providers_test.go` — typed provider and preview failures while preserving existing blocking-preview behavior.
- `server/schema.graphql`, `server/generated.go`, and `server/schema.resolvers.go` — generated GraphQL surfaces for refresh and account claim.

## Decisions Made

- Kept activation recovery machine-readable and closed: UI behavior keys off four codes, while all underlying causes stay server-side.
- Kept row expiry out of ordinary authorization. It blocks only re-auth minting, so an already-issued valid JWT continues to authorize until its own expiry.
- Kept the guest bearer credential stable during refresh and destroyed it only when account claim makes the row permanent.
- Used a single guarded `UPDATE` for claim and a single guarded `DELETE` for cleanup, keeping concurrency and retry behavior database-authoritative.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Rejected empty claim session IDs before mutating identity state**
- **Found during:** Task 3
- **Issue:** GraphQL non-null permits an empty string, which could allow a successful account claim without its required authoritative `account_claimed` event identity.
- **Fix:** Trimmed and rejected an empty `sessionID` before password hashing or the claim update.
- **Files modified:** `server/guest_users.go`
- **Verification:** Claim integration tests pass and the successful path stores exactly one authoritative event.
- **Committed in:** `a248535`

**2. [Rule 1 - Test Bug] Compared claim recovery copy through GraphQL's public message field**
- **Found during:** Task 3 GREEN verification
- **Issue:** The initial RED test compared `gqlerror.Error.Error()`, whose formatted value includes an operation-path prefix, rather than the browser-visible `Message` field.
- **Fix:** Added a typed test helper that unwraps `*gqlerror.Error` and compares its public `Message`, preserving the exact Signup copy requirement.
- **Files modified:** `server/guest_users_test.go`
- **Verification:** `TestGuestUsers_Claim/UsernameTaken` and `/EmptyPassword` pass under `-race`.
- **Committed in:** `a248535`

**3. [Rule 1 - Test Bug] Removed expired cleanup fixtures after preservation assertions**
- **Found during:** Task 3 GREEN verification
- **Issue:** The preservation test intentionally left a past-expiry guest after deleting its protecting game fixture, so the following idempotency test correctly counted that stale row as an extra deletion.
- **Fix:** Registered per-guest test cleanup so each preservation fixture is removed after its assertions.
- **Files modified:** `server/guest_cleanup_test.go`
- **Verification:** All four cleanup tests pass together and the idempotency test reports `(1, 0)`.
- **Committed in:** `a248535`

---

**Total deviations:** 3 auto-fixed (2 Rule 1, 1 Rule 2).
**Impact on plan:** The fixes preserve the planned architecture and strengthen correctness without adding product scope.

## Issues Encountered

- A Task 2 grep pipeline returned exit 141 under an added shell `pipefail` setting because `grep -q` closed its input after finding the required pass marker. The exact plan command was rerun without the extra shell option and exited 0; the underlying test was green.

## Known Stubs

- `server/schema.resolvers.go:73` — gqlgen retains an intentional dead `RefreshGuestSession` panic stub on `mutationResolver`.
- `server/schema.resolvers.go:78` — gqlgen retains an intentional dead `ClaimGuestAccount` panic stub on `mutationResolver`.

Both fields are implemented directly by `*graphQLServer`, which is the runtime mutation resolver returned by `server/graphql.go`; the focused integration tests exercise those real methods successfully.

## Verification

- `go build ./...` — pass
- `go test ./server -run 'TestActivationErrors|TestUsers_LoginRejectsGuest|TestGuestUsers_DisplayNameValidation|TestGuestUsers_ErrorCodes' -race -v` — pass
- `go test ./server -run 'TestGuestUsers_Refresh|TestGuestUsers_ExpiredRowStillAuthorized' -race -v` — pass
- `go test ./server -run 'TestGuestUsers_Claim|TestGuestUsers_Cleanup' -race -v` — pass against PostgreSQL
- Exact Task 3 automated command — pass
- `go test ./server -race` — pass against PostgreSQL in 245.561 seconds
- `git diff --exit-code server/authz.go` — pass; authorization remained JWT-only
- Activation documentation/source comparison — exactly `create_error`, `guest_session_error`, `preview_error`, and `provider_unavailable` in both
- Cleanup scheduling grep — 0 scheduling primitives

## User Setup Required

None — no dependencies, credentials, or external configuration were added.

## Next Phase Readiness

- Plans 02-03 and 02-04 can map the closed activation codes to recovery UI and use silent guest refresh without changing authorization.
- The later account-claim UI can call `claimGuestAccount`; its prefill decision remains typed display name first, generated username otherwise.
- No blockers remain. Cleanup intentionally has no production scheduler and production guest rows continue to use `expires_at = NULL`.

## Self-Check: PASSED

All five key created artifacts exist, and all six RED/GREEN task commits are present in git history.

---
*Phase: 02-guest-host-activation*
*Completed: 2026-08-20*

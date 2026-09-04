---
phase: RI-01-shared-database-identity
plan: RI-01-02
type: execute
wave: 3
depends_on: ["RI-01-01"]
files_modified:
  - persistence/migrations/20260820091000_canonical_user_uuid.up.sql
  - persistence/migrations/20260820091000_canonical_user_uuid.down.sql
  - persistence/migrations_test/20260820091000_canonical_user_uuid.up.sql
  - persistence/migrations_test/20260820091000_canonical_user_uuid.down.sql
  - server/users.go
  - server/guest_users.go
  - server/auth.go
  - server/authz.go
  - server/identity.go
  - server/identity_test.go
  - server/users_test.go
  - server/guest_users_test.go
requirements: [D-06, D-13]
autonomous: true
scope_rationale: "The twelve files are one indivisible account-identity path: one mirrored UUID migration, the five live account/auth seams that can allocate or resolve a principal, and their three existing regression suites. Splitting those seams would permit a mixed UUID cutover that D-06 forbids."
must_haves:
  truths:
    - "Signup, login, guest creation, refresh, guest claim, token parsing, authorization, and loaders preserve one canonical UUID per D-06."
    - "Guest claim changes credentials/profile state without allocating a replacement UUID or breaking any relationship."
    - "Deletion request and recovery use a controllable clock; final expiry remains deferred to the RI-08 durable outbox orchestration."
  artifacts:
    - path: "server/identity.go"
      provides: "canonical principal plus deletion request/recovery lifecycle"
  key_links:
    - from: "server/users.go, server/guest_users.go, server/auth.go, server/authz.go"
      to: "users.canonical_user_id"
      via: "all account creation, claim, token, authorization, and loader paths"
---

<objective>Promote one native canonical UUID through every vEDH account and guest flow without claiming JCT-7. Output: mirrored UUID/deletion-lifecycle DDL and stable identity behavior across all live auth seams.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-01-01-PLAN.md @server/users.go @server/guest_users.go @server/auth.go @server/authz.go</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Lock one UUID across account, guest, token, and loader flows</name>
  <files>server/identity_test.go, server/users_test.go, server/guest_users_test.go</files>
  <action>Write `TestCanonicalIdentityLifecycleContract` for null/blank/invalid/duplicate legacy UUID upgrade, one stable canonical row, duplicate retry, and a controllable Clock at deletion request/recovery boundaries. When the native UUID/deletion lifecycle is absent, fail only this test with `EXPECTED_RED[RI-01-02-T1]: canonical identity lifecycle is not installed`. Leave signup/login/guest/token/loader seam fixtures to Task 3, where their implementation lands, so this task has one precise expected failure and no unrelated red tests.</action>
  <acceptance_criteria>Every named flow has positive and negative coverage; claim never allocates a replacement identity; invalid upgrade rows fail closed with an operator report.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-01-02-T1 --require-test TestCanonicalIdentityLifecycleContract --require-reason 'canonical identity lifecycle is not installed' -- go test ./server/... -run '^TestCanonicalIdentityLifecycleContract$' -race -v</automated></verify>
  <done>The red tests specify one stable UUID and recoverable deletion lifecycle across every vEDH identity flow.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Install the canonical UUID and deletion request/recovery schema</name>
  <files>persistence/migrations/20260820091000_canonical_user_uuid.up.sql, persistence/migrations/20260820091000_canonical_user_uuid.down.sql, persistence/migrations_test/20260820091000_canonical_user_uuid.up.sql, persistence/migrations_test/20260820091000_canonical_user_uuid.down.sql, server/identity.go, server/identity_test.go</files>
  <action>Audit and normalize existing users.uuid values, add/backfill a native UUID canonical key, validate uniqueness/non-null, and retain a measured compatibility read behind VEDH_RI_NATIVE_UUID_READ. Add 30-day deletion requested/recovered states with injected Clock and immutable lifecycle audit entries. Do not execute expiry deletion, create jank identity state, or share a symmetric token-minting secret.</action>
  <acceptance_criteria>The migration is idempotent on the current-schema fixture; request/recovery boundaries are exact; flag-off reads the validated compatibility identity without changing it.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestCanonicalIdentityLifecycleContract$' -race -v &amp;&amp; go test ./server/... -run 'TestCanonicalIdentityUpgrade|TestAccountDeletionRequestRecovery|TestRuntimeMigrationParity' -race -v &amp;&amp; diff persistence/migrations/20260820091000_canonical_user_uuid.up.sql persistence/migrations_test/20260820091000_canonical_user_uuid.up.sql</automated></verify>
  <done>The database and domain service expose one validated canonical UUID plus request/recovery state.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Wire every live account and authorization seam to the canonical UUID</name>
  <files>server/users.go, server/guest_users.go, server/auth.go, server/authz.go, server/identity.go, server/users_test.go, server/guest_users_test.go</files>
  <action>Add fixtures for signup/login, guest creation, refresh, guest-to-account claim, token mint/parse, authorization/loaders, renewal, logout/relogin, and duplicate retry, then update those seams to consume and return canonical_user_id. Assert the UUID created once is byte-for-byte stable through every subject/response/lookup and all existing game/deck relationships. Preserve current credentials and guest restrictions while making claim update the same row. Add bounded identity outcome metrics with no UUID/user labels. Keep VEDH_RI_NATIVE_UUID_READ off until RI-01-03 approves cutover.</action>
  <acceptance_criteria>All stable-UUID tests pass; a retry never creates a second principal; authorization rejects malformed, expired, deleted, or mismatched subjects.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestCanonicalUUID|TestGuestClaimCanonicalUUID|TestTokenCanonicalUUID|TestCanonicalAuthorization' -race -v</automated></verify>
  <done>Every vEDH identity seam now resolves the same canonical UUID while production reads remain gated.</done>
</task>

</tasks>
<verification>Migration parity plus canonical UUID, guest claim, token, authorization, request, and recovery suites.</verification>
<success_criteria>D-06 is a stable vEDH prerequisite; JCT-7 remains exclusively owned by RI-07-01.</success_criteria>

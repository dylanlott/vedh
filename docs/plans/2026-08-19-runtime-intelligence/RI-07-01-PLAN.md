---
phase: RI-07-jank-cutover-and-evidence-trees
plan: RI-07-01
type: execute
wave: 23
depends_on: ["RI-01-04", "RI-06-05"]
files_modified:
  - docs/contracts/jank-checkout.json
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820150000_shared_identity.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820150000_shared_identity.down.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820150000_shared_identity.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820150000_shared_identity.down.sql"
  - "${JANK_CHECKOUT_PATH}/app/migrations.go"
  - "${JANK_CHECKOUT_PATH}/app/migrations_test.go"
  - "${JANK_CHECKOUT_PATH}/app/identity.go"
  - "${JANK_CHECKOUT_PATH}/app/identity_test.go"
  - "${JANK_CHECKOUT_PATH}/app/auth.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_auth.go"
  - "${JANK_CHECKOUT_PATH}/app/models.go"
  - "${JANK_CHECKOUT_PATH}/cmd/jank-identity-reconcile/main.go"
  - "${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md"
requirements: [JCT-7, D-06, D-07]
autonomous: true
scope_rationale: "The thirteen verified-checkout files are the one-time shared-identity cutover: paired migrations, migration/identity tests, principal/auth seams, legacy reconciliation, and rollback runbook. They cannot be separated without permitting mixed integer/UUID ownership or a duplicate auth path."
must_haves:
  truths:
    - "JCT-7 acceptance is owned exclusively here: both applications use the exact vEDH canonical UUID row and jank owns no platform account, credential, sync, or link state."
    - "Every jank author/creator FK is migrated to canonical UUID or a non-reversible content-author tombstone before cutover."
    - "Guest claim preserves the same UUID and every game, post, tree, annotation, evidence link, and moderation relationship."
    - "jank user FKs are installed with only temporary REFERENCES(canonical_user_id) for jank_migrator; the privilege is revoked and jank_app has no users SELECT/DML."
  artifacts:
    - path: "docs/contracts/jank-checkout.json"
      provides: "immutable remote/base plus integration branch and mutable audited expected_head"
    - path: "${JANK_CHECKOUT_PATH}/app/identity.go"
      provides: "canonical vEDH principal adapter"
  key_links:
    - from: "jank author/owner FKs"
      to: "public.users.canonical_user_id"
      via: "deployment-only REFERENCES grant then revoke"
---

<objective>Complete JCT-7 shared-identity cutover against the explicitly verified jank checkout. Output: versioned migrations, administrative reconciliation, canonical principal verification, guest-claim preservation, and no duplicate auth path.</objective>
<execution_context>@docs/contracts/jank-checkout.json @scripts/verify-jank-checkout.sh @docs/contracts/jank-user-fk-grant.sql @docs/plans/2026-08-19-runtime-intelligence/RI-06-03-PLAN.md</execution_context>

<precondition>Before any read/edit, run `scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode pre-plan --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-01-PLAN.md`: require configured remote/branch, immutable base ancestry, HEAD=expected_head, and clean tree. During tasks use `--mode task` so only this plan's declared jank files may be dirty; refuse undeclared dirt, wrong remote, divergence, or unexpected HEAD.</precondition>

<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Create migration, reconciliation, auth, and guest-claim fixtures</name>
  <files>${JANK_CHECKOUT_PATH}/app/migrations_test.go, ${JANK_CHECKOUT_PATH}/app/identity_test.go</files>
  <action>After the pre-plan check, write `TestSharedIdentityMigrationContract` for PostgreSQL current-to-shared migration of integer users and username-authored content, duplicates/renames/conflicts/missing users, two-pass restart, direct canonical UUID FKs, and grant revocation. When migration state is absent, fail only this test with `EXPECTED_RED[RI-07-01-T1]: jank shared identity migration is not installed`. End-to-end principal/guest-claim/no-local-auth fixtures land with Task 3.</action>
  <acceptance_criteria>Every authored row maps/tombstones/blocks explicitly; guest claim changes no canonical UUID or relationship; tests run only in the recorded checkout.</acceptance_criteria>
  <verify><automated>VEDH_ROOT=$PWD &amp;&amp; JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-01-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; "$VEDH_ROOT/scripts/verify-expected-red.sh" --suite RI-07-01-T1 --require-test TestSharedIdentityMigrationContract --require-reason 'jank shared identity migration is not installed' -- go test ./... -run '^TestSharedIdentityMigrationContract$' -race -v</automated></verify>
  <done>The JCT-7 migration contract is precisely specified against the preflighted integration expected_head.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Migrate jank authors to canonical UUID with least-privilege FKs</name>
  <files>${JANK_CHECKOUT_PATH}/migrations/postgres/20260820150000_shared_identity.up.sql, ${JANK_CHECKOUT_PATH}/migrations/postgres/20260820150000_shared_identity.down.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820150000_shared_identity.up.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820150000_shared_identity.down.sql, ${JANK_CHECKOUT_PATH}/app/migrations.go, ${JANK_CHECKOUT_PATH}/app/models.go, ${JANK_CHECKOUT_PATH}/cmd/jank-identity-reconcile/main.go</files>
  <action>After preflight, add versioned migrations and a dry-run/checkpoint/resume reconciliation CLI. Use operator-reviewed mappings; username equality is candidate evidence only. Missing jank-only accounts are provisioned/recovered through vEDH-owned flow, not copied credentials. Execute the exact RI-01 contract: temporary column-scoped REFERENCES grant to jank_migrator, install/validate every canonical-user FK, revoke it, then prove jank_owner/jank_migrator/jank_app have no users SELECT/INSERT/UPDATE/DELETE/TRIGGER. Preserve immutable display snapshots and tombstone unresolved historical authors.</action>
  <acceptance_criteria>Every foreign key directly targets canonical UUID; the temporary grant is absent afterward; two-pass reconciliation is idempotent.</acceptance_criteria>
  <verify><automated>VEDH_ROOT=$PWD &amp;&amp; JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-01-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run '^TestSharedIdentityMigrationContract$' -race -v &amp;&amp; go test ./... -run 'TestSharedIdentityMigration|TestCanonicalUserFKGrantRevoked|TestIdentityReconcileReplay' -race</automated></verify>
  <done>All jank ownership relationships reference the canonical row under the agreed grant/revoke boundary.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 3: Replace jank account/token paths with the approved vEDH principal</name>
  <files>${JANK_CHECKOUT_PATH}/app/identity.go, ${JANK_CHECKOUT_PATH}/app/identity_test.go, ${JANK_CHECKOUT_PATH}/app/auth.go, ${JANK_CHECKOUT_PATH}/app/handlers_auth.go, ${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md</files>
  <action>Add end-to-end UUID stability across vEDH guest game/snapshot/participant, guest claim, jank login, post, tree, annotation, evidence link, and moderation history. Implement RI-01's approved HttpOnly session or short-lived asymmetric/introspected proof with expiry/replay/logout/rotation/origin validation. Disable local signup/login/token minting and credential reads in integrated mode. Use canonical UUID for every mutation; forum display names remain presentation snapshots. Add maintenance/read-only rollback that never reactivates duplicate account creation after cutover.</action>
  <acceptance_criteria>Canonical principal tests and all relationship-preservation tests pass; local credential paths are unreachable in integrated mode.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-01-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestCanonicalPrincipal|TestGuestClaimPreservesAllRelationships|TestIntegratedAuthNoLocalAccount' -race</automated></verify>
  <done>JCT-7 is implemented end-to-end with one account row and no duplicate credential authority.</done>
</task>
</tasks>
<integration_worktree_completion>Do not push or merge. After all tasks pass, run prepare-commit for this plan, create one atomic jank commit containing only declared jank files, run `--mode advance` with that new immediate-descendant HEAD to update expected_head and append the plan/from/to/subject/timestamp audit in docs/contracts/jank-checkout.json, then run `--mode post-plan` and require configured remote/base ancestry/new expected HEAD/clean tree.</integration_worktree_completion>
<verification>Verified checkout precondition; jank PostgreSQL migration/auth/guest-claim suites; vEDH role matrix.</verification>
<success_criteria>JCT-7 is accepted here exactly once, including cross-app guest-claim relationship preservation.</success_criteria>

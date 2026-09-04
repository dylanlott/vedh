---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-02
type: execute
wave: 29
depends_on: ["RI-08-01"]
files_modified:
  - docs/contracts/jank-checkout.json
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820160500_account_deletion_bridge.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820160500_account_deletion_bridge.down.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820160500_account_deletion_bridge.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820160500_account_deletion_bridge.down.sql"
  - "${JANK_CHECKOUT_PATH}/app/account_deletion.go"
  - "${JANK_CHECKOUT_PATH}/app/account_deletion_test.go"
  - "${JANK_CHECKOUT_PATH}/cmd/jank-deletion-worker/main.go"
  - scripts/runtime-deletion-drill.sh
  - docs/runbooks/account-deletion-orchestration.md
requirements: [D-13]
autonomous: true
must_haves:
  truths:
    - "A dedicated jank worker leases a durable command, tombstones all jank-owned relationships, stores a receipt, and acknowledges in one worker-owned PostgreSQL transaction."
    - "The jank transaction is separate from the earlier vEDH local deletion transaction; failure rolls back jank mutation/ack and retry is idempotent."
    - "jank_app cannot consume deletion commands; the consumer cannot read vEDH base tables or perform arbitrary jank DML."
    - "Every RI-08 jank read, edit, test, and drill resolves and verifies docs/contracts/jank-checkout.json first."
  artifacts:
    - path: "${JANK_CHECKOUT_PATH}/app/account_deletion.go"
      provides: "least-privilege jank deletion command application and receipt"
    - path: "scripts/runtime-deletion-drill.sh"
      provides: "target-guarded two-process failure/replay/restore drill"
  key_links:
    - from: "vEDH deletion outbox lease"
      to: "jank tombstone plus vEDH ack"
      via: "one jank-worker DB transaction using only reviewed functions"
---

<objective>Implement and verify the jank-owned consumer/receipt half of D-13 against the recorded checkout. Output: versioned migrations, least-privilege worker transaction, relationship tombstoning, automated jank deletion tests, and full two-process drill.</objective>
<execution_context>@docs/contracts/jank-checkout.json @scripts/verify-jank-checkout.sh @docs/plans/2026-08-19-runtime-intelligence/RI-08-01-PLAN.md</execution_context>
<precondition>Run verifier pre-plan mode for RI-08-02 before jank access; require configured remote/branch, immutable base ancestry, HEAD=expected_head, and clean tree. During tasks use task mode so only declared jank files may be dirty; reject wrong remote, divergence, unexpected HEAD, or undeclared dirt.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Specify jank tombstone, receipt, role, replay, and restore behavior</name>
  <files>${JANK_CHECKOUT_PATH}/app/account_deletion_test.go</files>
  <action>After task-mode verification, write `TestJankDeletionBridgeContract` for lease/apply across all jank-owned relationships, failure before/between updates, duplicate command, concurrent workers, response loss, expired lease, wrong schema/generation, unauthorized roles, UUID scrubbing, and restore-generation replay. Receipt stores command/deletion/generation/result counts but no canonical UUID or reversible mapping. When bridge storage/worker is absent, fail only this test with `EXPECTED_RED[RI-08-02-T1]: jank deletion bridge is not installed`.</action>
  <acceptance_criteria>Automated tests prove one complete jank tombstone transaction or rollback, idempotent receipt, least privilege, and no reversible identity mapping.</acceptance_criteria>
  <verify><automated>VEDH_ROOT=$PWD &amp;&amp; JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-02-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; "$VEDH_ROOT/scripts/verify-expected-red.sh" --suite RI-08-02-T1 --require-test TestJankDeletionBridgeContract --require-reason 'jank deletion bridge is not installed' -- go test ./... -run '^TestJankDeletionBridgeContract$' -race -v</automated></verify>
  <done>The verified checkout has executable jank deletion and negative-role coverage before implementation.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement jank apply/receipt functions and durable consumer</name>
  <files>${JANK_CHECKOUT_PATH}/migrations/postgres/20260820160500_account_deletion_bridge.up.sql, ${JANK_CHECKOUT_PATH}/migrations/postgres/20260820160500_account_deletion_bridge.down.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820160500_account_deletion_bridge.up.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820160500_account_deletion_bridge.down.sql, ${JANK_CHECKOUT_PATH}/app/account_deletion.go, ${JANK_CHECKOUT_PATH}/app/account_deletion_test.go, ${JANK_CHECKOUT_PATH}/cmd/jank-deletion-worker/main.go</files>
  <action>Before reading or editing any listed file, resolve and verify the recorded checkout. Add jank-owned receipt/tombstone schema and a SECURITY DEFINER apply_account_deletion function owned by jank_owner, with trusted search_path, PUBLIC execute revoked, and execute only for jank_deletion_consumer. The worker opens one PostgreSQL transaction, leases via the vEDH function, calls jank apply to generate fresh non-reversible tombstones and update every owned relationship, calls vEDH ack which scrubs pending subject UUID, then commits. Failure rolls back apply and ack together. On commit/response loss, the durable receipt makes replay a no-op. Integrated PostgreSQL is authoritative; SQLite supplies deterministic local contract fixtures without pretending to share vEDH transactions.</action>
  <acceptance_criteria>Jank apply and ack commit together in the worker transaction; the earlier vEDH transaction remains separate; retries and restore generations converge safely.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-02-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; JANK_DB_DRIVER=pgx JANK_DB_DSN="$JANK_TEST_DSN" go test ./... -run '^TestJankDeletionBridgeContract$' -race -v &amp;&amp; JANK_DB_DRIVER=pgx JANK_DB_DSN="$JANK_TEST_DSN" go test ./... -run 'TestJankAccountDeletion|TestJankDeletionReplay|TestJankDeletionRoleMatrix|TestJankDeletionRestoreGeneration' -race -v</automated></verify>
  <done>The jank worker delivers durable eventual deletion with a receipt, not a fictitious shared Go transaction.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Rehearse two-process deletion, partial failures, response loss, and restore replay</name>
  <files>scripts/runtime-deletion-drill.sh, docs/runbooks/account-deletion-orchestration.md, ${JANK_CHECKOUT_PATH}/app/account_deletion_test.go</files>
  <action>Before reading the jank test or running any jank command, resolve and verify the recorded checkout. Build a target-guarded drill for request, recovery, exact expiry, vEDH local rollback, local commit/jank pending, jank failure rollback, retry, response loss after worker commit, cleanup retry, and restored-generation re-deletion. Require machine-readable states and receipt lag, prove no authentication/private/share access after local expiry, prove jank author/owner display tombstones after receipt, and refuse empty/current/production-default targets. Update the runbook with exact commands and monitoring thresholds.</action>
  <acceptance_criteria>The drill demonstrates every partial state and eventual convergence, refuses unsafe targets, and executes the verified-checkout jank deletion tests.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-02-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestJankAccountDeletion|TestJankDeletionRestoreGeneration' -race &amp;&amp; cd - &gt;/dev/null &amp;&amp; bash -n scripts/runtime-deletion-drill.sh &amp;&amp; scripts/runtime-deletion-drill.sh --self-test-target-guards &amp;&amp; scripts/runtime-deletion-drill.sh --self-test-scenarios</automated></verify>
  <done>D-13 is proven across both processes, receipts, cleanup, and restored data using the verified jank checkout.</done>
</task>

</tasks>
<integration_worktree_completion>Create one atomic unpushed jank commit after prepare-commit, advance expected_head with its audit record, then require post-plan configured remote/base ancestry/new expected HEAD/clean tree.</integration_worktree_completion>
<verification>Every jank command begins with checkout resolution/verification; automated jank deletion, role, replay, restore, and full drill suites pass.</verification>
<success_criteria>D-13 is executable across applications with explicit eventual completion and least privilege.</success_criteria>

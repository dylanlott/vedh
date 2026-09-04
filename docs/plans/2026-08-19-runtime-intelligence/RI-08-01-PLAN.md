---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-01
type: execute
wave: 28
depends_on: ["RI-07-05"]
files_modified:
  - persistence/migrations/20260820160000_account_deletion_outbox.up.sql
  - persistence/migrations/20260820160000_account_deletion_outbox.down.sql
  - persistence/migrations_test/20260820160000_account_deletion_outbox.up.sql
  - persistence/migrations_test/20260820160000_account_deletion_outbox.down.sql
  - server/account_deletion.go
  - server/account_deletion_test.go
  - server/account_deletion_worker.go
  - server/account_deletion_worker_test.go
  - server/account_deletion_bridge.go
  - server/account_deletion_bridge_test.go
  - server/search_cache_cleanup.go
  - server/search_cache_cleanup_test.go
  - docs/runbooks/account-deletion-orchestration.md
requirements: [D-13]
autonomous: true
scope_rationale: "The thirteen files are the vEDH half of one durable deletion protocol: mirrored schema/functions, local orchestrator/worker, outbox bridge, cleanup/restore reconciliation, tests, and the cross-process contract. Splitting the local transaction from its outbox would invalidate the promised failure semantics."
must_haves:
  truths:
    - "At the 30-day boundary, one vEDH transaction rechecks recovery, deletes/revokes/tombstones all vEDH-owned state, and inserts a durable jank deletion command or rolls back."
    - "No Go transaction spans vEDH and jank processes; the job remains awaiting_jank until a separate verified receipt commits."
    - "The outbox contract is least privilege, leased/retryable/idempotent, scrubs the subject UUID after receipt, and monitors receipt age without identity labels."
    - "A non-reversible restore fingerprint and restore_generation replay re-delete restored expired data without retaining a tombstone-to-user lookup."
  artifacts:
    - path: "server/account_deletion.go"
      provides: "vEDH-local ExecuteExpiredAccountDeletion transaction"
    - path: "server/account_deletion_bridge.go"
      provides: "durable command lease/ack/fail and receipt state machine"
    - path: "docs/runbooks/account-deletion-orchestration.md"
      provides: "cross-process authority, event, retry, monitoring, partial-failure, and restore contract"
  key_links:
    - from: "vEDH local deletion commit"
      to: "account_deletion_outbox"
      via: "same PostgreSQL transaction"
    - from: "jank deletion receipt"
      to: "account deletion terminal state"
      via: "separate worker transaction and durable ack; no shared Go transaction"
---

<objective>Implement the vEDH-owned half of D-13 as an honest durable outbox protocol. Output: local deletion transaction, least-privilege bridge functions, job/receipt state, retries/monitoring, cleanup, and restore-generation replay.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-01-02-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/RI-04-03-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/RI-06-04-PLAN.md @docs/contracts/runtime-database-roles.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Specify local atomicity, outbox delivery, role negatives, and restore replay</name>
  <files>server/account_deletion_test.go, server/account_deletion_worker_test.go, server/account_deletion_bridge_test.go, server/search_cache_cleanup_test.go</files>
  <action>Write `TestDeletionOutboxContract` using a fake Clock and deterministic IDs for request, recovery, exact 30-day boundary, concurrent workers, retry, failure before/between every vEDH hook, local commit plus outbox insert, outbox failure rollback, lease expiry, response loss, duplicate ack, sanitized failure, and jank receipt delay. Specify command schema_version, command_id, deletion_id, restore_generation, subject UUID while pending, created/available/leased times, attempt count, and sanitized error code. Verify the subject is scrubbed after receipt, only a non-reversible salted restore fingerprint remains, and no tombstone-to-user mapping exists. Assert jank_app cannot read/execute bridge objects; a dedicated jank_deletion_consumer capability can only execute named lease/ack/fail functions and cannot read base vEDH or mutate general jank tables. When the local/outbox bridge is absent, fail only the named test with `EXPECTED_RED[RI-08-01-T1]: deletion outbox contract is not installed`; cleanup and restore-generation fixtures land with Task 3.</action>
  <acceptance_criteria>Tests prove local all-or-nothing state, eventual cross-process state, replay identity, least privilege, non-reversible restore matching, and explicit pending/failed/complete states.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-08-01-T1 --require-test TestDeletionOutboxContract --require-reason 'deletion outbox contract is not installed' -- go test ./server/... -run '^TestDeletionOutboxContract$' -race -v</automated></verify>
  <done>The red suite makes the two-transaction boundary and every partial-failure outcome executable.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement local deletion, durable outbox, receipts, and worker</name>
  <files>persistence/migrations/20260820160000_account_deletion_outbox.up.sql, persistence/migrations/20260820160000_account_deletion_outbox.down.sql, persistence/migrations_test/20260820160000_account_deletion_outbox.up.sql, persistence/migrations_test/20260820160000_account_deletion_outbox.down.sql, server/account_deletion.go, server/account_deletion_test.go, server/account_deletion_worker.go, server/account_deletion_worker_test.go, server/account_deletion_bridge.go, server/account_deletion_bridge_test.go</files>
  <action>Create deletion jobs/attempts/domain receipts, durable jank outbox, cleanup outbox, restore-generation ledger, and non-reversible restore fingerprint registry. Add a dedicated NOLOGIN capability role jank_deletion_consumer with EXECUTE only on SECURITY DEFINER lease/ack/fail functions; each function uses an explicit trusted search_path ending pg_temp, validates role/state/schema/generation, and has PUBLIC execute revoked. Implement ExecuteExpiredAccountDeletion(ctx, tx, deletionID, clock) to lock/recheck deadline/recovery, revoke shares/publications, delete reviews/search/profile/credentials/auth/reset/guest secrets, tombstone participants/snapshot ownership, enqueue cleanup, and insert one unique jank command in the same vEDH transaction. It never calls jank Go code or mutates jank tables. The vEDH job becomes terminal only after jank receipt and cleanup receipts. Use SKIP LOCKED leases, bounded exponential backoff, expiring leases, persistent paging after repeated failures, and no give-up terminal state for privacy work.</action>
  <acceptance_criteria>Local changes and command commit together; command uniqueness is deletion/domain/generation; jank delay is visible pending state; only a valid receipt advances the job.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestDeletionOutboxContract$' -race -v &amp;&amp; go test ./server/... -run 'TestExecuteExpiredAccountDeletion|TestDeletionLocalOutboxAtomicity|TestDeletionBridgeLeaseAck|TestDeletionBridgeRoleMatrix|TestDeletionFailureRecovery' -race -v &amp;&amp; diff persistence/migrations/20260820160000_account_deletion_outbox.up.sql persistence/migrations_test/20260820160000_account_deletion_outbox.up.sql</automated></verify>
  <done>vEDH deletion is locally atomic and durably requests, but does not falsely claim, jank completion.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Implement cleanup, monitoring, and restore-generation requeue</name>
  <files>server/search_cache_cleanup.go, server/search_cache_cleanup_test.go, server/account_deletion_worker.go, server/account_deletion_worker_test.go, docs/runbooks/account-deletion-orchestration.md</files>
  <action>Implement idempotent cache/search/index cleanup receipts and a restore-generation scanner that compares restored candidate UUIDs to the non-reversible salted fingerprint registry, reapplies local deletion, and enqueues a new generation-specific jank command. Document ownership/authority, the exact command/receipt schema, delivery/lease/retry/idempotency, both transaction boundaries, response-loss behavior, role/auth negatives, UUID scrubbing, restore replay, and monitoring. Emit bounded metrics for pending age, lease retries, receipt lag, bridge auth denial, cleanup backlog, restored-generation backlog, and terminal completion; never label UUID/deletion/command IDs. Alert on age thresholds but continue retries.</action>
  <acceptance_criteria>Restore replay creates a new generation command, cleanup is independently retryable, monitoring detects stuck delivery, and the runbook states that global completion is eventual rather than one shared transaction.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestDeletionCleanupRetry|TestDeletionMonitoring|TestRestoredDataRedeletion|TestDeletionRestoreGenerationReplay' -race -v &amp;&amp; grep -q 'No shared Go transaction' docs/runbooks/account-deletion-orchestration.md</automated></verify>
  <done>D-13 local, monitoring, cleanup, and restore semantics are explicit, tested, and honest.</done>
</task>

</tasks>
<verification>Migration parity plus local atomicity, bridge role, retry/receipt, cleanup/monitoring, and restore-generation suites.</verification>
<success_criteria>vEDH can complete its owned deletion work and durably track jank/cleanup without claiming cross-process atomicity.</success_criteria>

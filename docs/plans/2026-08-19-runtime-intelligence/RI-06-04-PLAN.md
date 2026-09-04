---
phase: RI-06-privacy-safe-publication
plan: RI-06-04
type: execute
wave: 21
depends_on: ["RI-06-02"]
files_modified:
  - server/publication_deletion.go
  - server/publication_deletion_test.go
requirements: [D-13]
autonomous: true
must_haves:
  truths:
    - "Account-expiry revocation is idempotent by deletion ID, immediately closes participant/public payloads, and retains only stable unavailable tombstones."
  artifacts:
    - path: "server/publication_deletion.go"
      provides: "vEDH-local publication/share revocation hook for RI-08"
  key_links:
    - from: "RI-08 local deletion transaction"
      to: "participant/public share revocation"
      via: "idempotent RevokePublicationsForAccount hook"
---

<objective>Implement the publication-domain portion of D-13 separately from API grants. Output: a transaction-bound idempotent hook and replay/restore tests consumed by RI-08.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-06-02-PLAN.md @server/review_deletion.go</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement idempotent publication and participant-share revocation</name>
  <files>server/publication_deletion.go, server/publication_deletion_test.go</files>
  <action>Implement RevokePublicationsForAccount(ctx, tx, canonicalUserID, deletionID) for the RI-08 vEDH-local transaction. Revoke public and participant shares, close payload/status access as policy requires, enqueue local cache/search invalidation, retain stable unavailable tombstones and canonical sources, and make replay/partial failure/restored-data re-deletion idempotent. Do not mutate jank tables or claim cross-process atomicity.</action>
  <acceptance_criteria>Expiry closes all vEDH share payloads inside the local transaction; retry is safe; jank references remain resolvable as unavailable through evidence_api.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestPublicationDeletionHook|TestPublicationDeletionReplay|TestPublicationDeletionRestore' -race -v</automated></verify>
  <done>RI-08 can revoke all vEDH publication/share state without embedding jank ownership logic.</done>
</task>

</tasks>
<verification>Publication deletion, replay, partial-failure, and restored-data tests.</verification>
<success_criteria>The D-13 publication hook is ready for durable outbox orchestration.</success_criteria>

---
phase: RI-02-immutable-game-context
plan: RI-02-03
type: execute
wave: 8
depends_on: ["RI-02-02"]
files_modified:
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/runtime_context_api.go
  - server/runtime_context_api_test.go
  - docs/runbooks/runtime-context-migration.md
requirements: [RI-CONTEXT-CUTOVER]
autonomous: false
must_haves:
  truths:
    - "Authorized users can inspect exact snapshots, lineage history, participant context, result revisions, and exclusion reasons."
    - "The production dual-write cohort cannot start without approved reconciliation and rollback evidence."
  artifacts:
    - path: "server/runtime_context_api.go"
      provides: "typed participant-authorized immutable context API"
  key_links:
    - from: "runtime context GraphQL"
      to: "immutable context tables"
      via: "canonical principal authorization"
---

<objective>Expose typed immutable context and gate production dual-write after reconciliation. Output: generated GraphQL contract, negative authorization tests, and blocking cutover decision.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-02-02-PLAN.md @server/schema.graphql</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Add authorized immutable context queries and correction commands</name>
  <files>server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/runtime_context_api.go, server/runtime_context_api_test.go</files>
  <action>Add typed snapshot, lineage history, participant/result revision, eligibility/exclusion queries and audited correction/move/split/merge mutations. Result revisions expose ending turn/round, bounded finalization source, structured declared win-condition code/vocabulary/details, source event, actor, correction link, and revision history. Authorize with canonical UUID; cover unrelated user, username collision, guest mismatch, deleted principal, moderator, and jank_app. Preserve unknown/not_recorded distinctions.</action>
  <acceptance_criteria>Generated API exposes immutable sources and audit history without raw private payloads or unauthorized access.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestRuntimeContextAPI|TestRuntimeContextAuthorizationNegative' -race -v</automated></verify>
  <done>The immutable context is consumable by RI-03/RI-04 through a typed authorized contract.</done>
</task>
<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Approve context backfill and dual-write cohort</name>
  <files>docs/runbooks/runtime-context-migration.md</files>
  <action>Review production-shaped expand/backfill/validate output, exact/partial/unknown/excluded counts, RGI-3 canonical result-field reconciliation, invalid lineage candidates, role negatives, restore, and flag rollback. Record APPROVED or REJECTED, cohort, operator, timestamp, evidence, and triggers. Rejection/absence keeps VEDH_RI_SNAPSHOT_DUAL_WRITE off and leaves this plan incomplete.</action>
  <acceptance_criteria>All ambiguity classes are understood and the decision is explicit; legacy compatibility remains through observation.</acceptance_criteria>
  <verify><automated>grep -q 'Production context cutover decision: APPROVED' docs/runbooks/runtime-context-migration.md &amp;&amp; go test ./server/... -run 'TestRuntimeContextBackfill|TestRuntimeContextAuthorizationNegative|TestFinishResultCanonicalFields|TestResultCorrectionPreservesPriorRevision' -race</automated></verify>
  <done>The context cohort is APPROVED with rollback evidence; rejection/absence halts RI-03 and is not dependency completion.</done>
</task>
</tasks>
<verification>`make generate`; `go test ./server/... -run RuntimeContext -race`.</verification>
<success_criteria>RI-03 starts only after immutable context is typed, authorized, and explicitly approved.</success_criteria>

---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-05
type: execute
wave: 32
depends_on: ["RI-08-04"]
files_modified:
  - docs/runbooks/runtime-intelligence-beta.md
  - docs/operations/runtime-intelligence-beta-report.md
requirements: [D-04]
autonomous: false
entry_precondition: "docs/operations/runtime-intelligence-beta-report.md must contain Pre-beta readiness decision: APPROVED; REJECTED or absent evidence is a hard stop, not a completed dependency."
must_haves:
  truths:
    - "Activation opens only after the RI-08-04 APPROVED record, then recruitment continues until 30 eligible testers activate."
    - "The first 29 activations leave the measurement start unset; the 30th commit sets it immutably; candidate count may exceed 30."
  artifacts:
    - path: "docs/operations/runtime-intelligence-beta-report.md"
      provides: "candidate/exclusion ledger, exactly 30 activated cohort members, and database-derived 30th activation start"
---

<objective>Open the approved activation cohort, recruit until 30 eligible activations, and verify the database-set immutable six-week start. Output: real activation/exclusion evidence and a locked start timestamp.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-08-04-PLAN.md @docs/analytics/runtime-intelligence-beta.sql</execution_context>
<tasks>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 1: Activate exactly 30 eligible testers and verify the immutable 30th-start</name>
  <files>docs/runbooks/runtime-intelligence-beta.md, docs/operations/runtime-intelligence-beta-report.md</files>
  <action>First require the exact RI-08-04 APPROVED readiness record; a rejected/absent record stops before enabling VEDH_RI_BETA_COHORT. Then recruit candidates as needed, retaining ineligible/no-game/dropout/duplicate/test exclusions separately. Do not cap candidate count at 30. Verify the first 29 accepted activations left measurement_window_started_at null and the transaction for the 30th distinct eligible activation set it to that activation's committed timestamp. Record the 30 immutable cohort IDs only in the protected database/report references, never public logs. Verify later candidate activity cannot change membership, denominator, or start.</action>
  <acceptance_criteria>Readiness is APPROVED, activated denominator is exactly 30, exclusions are separate, and the immutable start equals the 30th eligible activation commit rather than any approval/invitation/earlier activation.</acceptance_criteria>
  <verify><automated>grep -q 'Pre-beta readiness decision: APPROVED' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; grep -q 'Activated testers: 30' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; grep -q 'Measurement start source: 30th eligible activation' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; go test ./server/... -run 'TestThirtiethActivationStartsWindow|TestRuntimeBetaCohortImmutable' -race</automated></verify>
  <done>Thirty eligible testers are locked and the six-week clock has started immutably at the 30th activation.</done>
</task>

</tasks>
<verification>APPROVED readiness, exactly 30 activated testers, separate exclusions, and database-derived 30th-activation timestamp.</verification>
<success_criteria>D-04 activation and clock semantics are satisfied without a second start approval redefining time.</success_criteria>

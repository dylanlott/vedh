---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-04
type: execute
wave: 31
depends_on: ["RI-08-03"]
files_modified:
  - persistence/migrations/20260820161000_runtime_beta_measurement.up.sql
  - persistence/migrations/20260820161000_runtime_beta_measurement.down.sql
  - persistence/migrations_test/20260820161000_runtime_beta_measurement.up.sql
  - persistence/migrations_test/20260820161000_runtime_beta_measurement.down.sql
  - server/beta_measurement.go
  - server/beta_measurement_test.go
  - docs/analytics/runtime-intelligence-beta.sql
  - docs/runbooks/runtime-intelligence-beta.md
  - docs/operations/runtime-intelligence-beta-report.md
requirements: [D-03, D-04]
autonomous: false
must_haves:
  truths:
    - "Candidate/activation infrastructure exists and is tested while cohort activation remains closed."
    - "Human technical/privacy/operational readiness must be APPROVED before any candidate can activate into the beta cohort."
    - "The transaction that accepts the 30th distinct eligible activation immutably sets the six-week start; invitation, approval, first activation, and the 29th activation cannot set it."
    - "A REJECTED or absent readiness decision leaves this plan incomplete and cannot satisfy RI-08-05."
  artifacts:
    - path: "server/beta_measurement.go"
      provides: "flag-gated candidate, activation, immutable cohort lock, and loop-stage measurement"
    - path: "docs/operations/runtime-intelligence-beta-report.md"
      provides: "APPROVED-only pre-activation readiness record"
---

<objective>Build beta measurement and complete technical/human readiness before activation opens. Output: monotonic post-deletion migration, tested 30th-activation transaction, exclusion/north-star SQL, and an APPROVED-only entry gate.</objective>
<execution_context>@docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-08-03-PLAN.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Define closed-cohort, activation, 30th-start, exclusion, and north-star fixtures</name>
  <files>server/beta_measurement_test.go, docs/analytics/runtime-intelligence-beta.sql</files>
  <action>Write `TestRuntimeBetaMeasurementContract` for activation attempts while readiness is not APPROVED or the cohort flag is off; invitations/candidates beyond 30; ineligible/no-game/dropout/duplicate/test-account exclusions; eligible activation; concurrent 29th/30th/31st attempts; immutable cohort of exactly 30; measurement start exactly on the 30th commit; and rolling 14-day three-game+inspection+documented-revision numerator. Candidate exclusions remain separate. Rejection never opens activation and never counts as dependency completion. Store no notes/payloads/hidden cards/opponent IDs. When measurement storage/service is absent, fail only the named test with `EXPECTED_RED[RI-08-04-T1]: beta measurement contract is not installed`.</action>
  <acceptance_criteria>Tests prove no activation before approval, no start before 30, one immutable start at 30, exactly 30 cohort members, and separate exclusion reporting.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-08-04-T1 --require-test TestRuntimeBetaMeasurementContract --require-reason 'beta measurement contract is not installed' -- go test ./server/... -run '^TestRuntimeBetaMeasurementContract$' -race -v</automated></verify>
  <done>The beta clock and readiness ordering are executable before schema/service implementation.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement flag-gated candidate/activation and immutable 30th-start measurement</name>
  <files>persistence/migrations/20260820161000_runtime_beta_measurement.up.sql, persistence/migrations/20260820161000_runtime_beta_measurement.down.sql, persistence/migrations_test/20260820161000_runtime_beta_measurement.up.sql, persistence/migrations_test/20260820161000_runtime_beta_measurement.down.sql, server/beta_measurement.go, server/beta_measurement_test.go, docs/analytics/runtime-intelligence-beta.sql, docs/runbooks/runtime-intelligence-beta.md, docs/operations/runtime-intelligence-beta-report.md</files>
  <action>Create candidates, candidate_activation_exclusions, immutable activated cohort, cohort lock/start, readiness evidence pointer, and append-only loop-stage events. VEDH_RI_BETA_COHORT defaults off and activation requires a stored APPROVED readiness decision. Recruit candidates without a 30-candidate cap; accept only an eligible game with canonical participant and immutable snapshot. Transactionally insert the distinct activation and set measurement_window_started_at only when the locked count becomes exactly 30; later candidates remain outside the fixed denominator. The start timestamp is immutable. Add versioned SQL for activated denominator, rolling north star, funnels, exclusions, quality/inference/latency/privacy/deletion incidents with missing/unmeasured explicit.</action>
  <acceptance_criteria>Migration timestamp follows 20260820160000/160500 deletion work; activation dedupes; readiness rejection blocks; 30th activation locks one start and denominator.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestRuntimeBetaMeasurementContract$' -race -v &amp;&amp; go test ./server/... -run 'TestRuntimeBetaReadinessClosed|TestRuntimeBetaActivation|TestThirtiethActivationStartsWindow|TestRuntimeNorthStar' -race &amp;&amp; diff persistence/migrations/20260820161000_runtime_beta_measurement.up.sql persistence/migrations_test/20260820161000_runtime_beta_measurement.up.sql</automated></verify>
  <done>Measurement infrastructure is complete but cohort activation remains closed pending human approval.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 3: Approve technical, privacy, deletion, and operational readiness before activation</name>
  <files>docs/operations/runtime-intelligence-beta-report.md</files>
  <action>Review every upstream APPROVED gate, final fail-closed smoke, migration order, D-13 local/outbox/jank receipt/cleanup/restore drill, critical defects, privacy/role results, support/rollback owners, and candidate recruitment plan. Record Pre-beta readiness decision: APPROVED or REJECTED with reviewer/time/evidence. This occurs before cohort activation opens and before any measurement window exists. Rejection/absence leaves VEDH_RI_BETA_COHORT off, leaves this plan incomplete, and cannot satisfy RI-08-05.</action>
  <acceptance_criteria>Approval requires zero critical privacy/security/integrity failures, complete deletion receipts/drill, support/rollback readiness, and an unset measurement start.</acceptance_criteria>
  <verify><automated>grep -q 'Pre-beta readiness decision: APPROVED' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; grep -q 'Measurement window start: unset' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; scripts/runtime-beta-gate.sh</automated></verify>
  <done>Pre-beta readiness is APPROVED before activation; rejection/absence halts downstream and is not dependency completion.</done>
</task>

</tasks>
<verification>Beta measurement/migration tests, final fail-closed gate, and APPROVED-only pre-activation decision with unset clock.</verification>
<success_criteria>Technical and human readiness precede cohort activation, while the immutable clock remains owned by the future 30th activation.</success_criteria>

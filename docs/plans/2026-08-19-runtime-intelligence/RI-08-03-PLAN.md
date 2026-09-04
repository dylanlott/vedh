---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-03
type: execute
wave: 30
depends_on: ["RI-08-02"]
files_modified:
  - docs/contracts/jank-checkout.json
  - tools/smoke/src/main.rs
  - tools/smoke/Cargo.toml
  - app/e2e/runtime-intelligence-loop.spec.ts
  - app/e2e/runtime-intelligence-privacy.spec.ts
  - scripts/runtime-beta-gate.sh
  - .github/workflows/runtime-intelligence-beta.yml
requirements: [RI-BETA-GATE, D-01]
autonomous: true
must_haves:
  truths:
    - "Final smoke/readiness runs only after D-13 jank delivery, receipt, and restore tests exist and pass."
    - "One fail-closed gate exercises the complete D-01 loop, migration order, roles, privacy, deletion, scale, and rollback through live repository targets."
    - "Every jank command resolves and verifies docs/contracts/jank-checkout.json before access."
  artifacts:
    - path: "scripts/runtime-beta-gate.sh"
      provides: "single post-deletion pre-beta fail-closed verifier"
---

<objective>Create the final integrated smoke and readiness command after deletion is complete. Output: wired Rust smoke, Playwright flows, CI, and one fail-closed gate that cannot bypass jank verification or D-13.</objective>
<execution_context>@Makefile @tools/smoke/Cargo.toml @docs/contracts/jank-checkout.json @docs/plans/2026-08-19-runtime-intelligence/RI-08-02-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/00-VALIDATION.md</execution_context>
<precondition>Run verifier pre-plan mode for RI-08-03 before jank-backed smoke/tests; require configured remote/branch, base ancestry, HEAD=expected_head, and clean tree. This plan is read/test-only in jank, so every later jank command uses task mode with zero permitted jank dirt.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Extend wired smoke and full-loop browser fixtures after deletion delivery exists</name>
  <files>tools/smoke/src/main.rs, tools/smoke/Cargo.toml, app/e2e/runtime-intelligence-loop.spec.ts, app/e2e/runtime-intelligence-privacy.spec.ts</files>
  <action>Before any jank-backed smoke step, resolve and verify the checkout contract. Add canonical UUID/claim, eligible game/context/result, mutation retry/timeline, optional debrief, lineage n=1/Limited evidence/labels/source notice/export/revision, participant/public share, both JCT-1 imports, tree/evidence/annotations/indicators/source navigation/revision/fork/discussion, D-13 local-pending/receipt/deleted/revoked rendering, and unrelated user/moderator/guest/role/privacy negatives to the live Cargo and Playwright targets.</action>
  <acceptance_criteria>Live Rust and browser tests cover the complete loop plus D-13 cross-process states and only use the verified jank checkout.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-03-PLAN.md &amp;&amp; make test-smoke-rust &amp;&amp; npm --prefix app run test:e2e -- runtime-intelligence-loop.spec.ts runtime-intelligence-privacy.spec.ts</automated></verify>
  <done>The wired smoke/browser gates exercise the complete D-01 loop and deletion/privacy boundaries.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Compose the post-deletion fail-closed readiness gate and CI</name>
  <files>scripts/runtime-beta-gate.sh, .github/workflows/runtime-intelligence-beta.yml</files>
  <action>Resolve and verify the recorded jank checkout before invoking any jank test. Run generation; migration parity in topological timestamp order; Go/API/Vue/type/build/E2E/Rust; verified-checkout jank PostgreSQL suite; two-pass backfills; roles/FK grants/outbox consumer negatives; transaction faults; inference thresholds; projector/export parity; scale; restore/rollback; publication privacy; and the full D-13 two-process request/recovery/expiry/partial-failure/response-loss/receipt/cleanup/restore drill. Fail on missing tools/results, wrong checkout, threshold breach, pending deletion receipt, REJECTED prerequisite, or migration order reversal. Upload sanitized reports only.</action>
  <acceptance_criteria>Any critical migration/security/privacy/integrity/latency/product-loop/deletion failure returns nonzero; no fallback masks missing evidence.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-03-PLAN.md &amp;&amp; bash -n scripts/runtime-beta-gate.sh &amp;&amp; scripts/runtime-beta-gate.sh --self-test</automated></verify>
  <done>One reproducible command blocks readiness on every critical gate, including verified D-13 delivery.</done>
</task>

</tasks>
<integration_worktree_completion>No jank commit or production push/merge occurs because this plan declares no jank edits. Append a read-only verification audit entry without changing expected_head, then require post-plan remote/base ancestry/expected HEAD/clean tree.</integration_worktree_completion>
<verification>Verified checkout, wired smoke/Playwright, and post-deletion fail-closed gate self-test.</verification>
<success_criteria>Final readiness is serialized after deletion and cannot report green while jank work is missing or pending.</success_criteria>

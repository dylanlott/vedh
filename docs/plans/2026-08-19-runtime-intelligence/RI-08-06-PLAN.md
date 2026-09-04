---
phase: RI-08-pre-beta-and-six-week-beta
plan: RI-08-06
type: execute
wave: 33
depends_on: ["RI-08-05"]
files_modified:
  - docs/contracts/jank-checkout.json
  - scripts/runtime-rollback-drill.sh
  - docs/runbooks/runtime-intelligence-beta.md
  - docs/operations/runtime-intelligence-beta-report.md
requirements: [RI-BETA-OPERATIONS, D-01, D-03, D-04, D-13]
autonomous: false
must_haves:
  truths:
    - "Six consecutive weeks are measured from the immutable 30th-activation start, with no skipped weekly gate."
    - "Weekly sampling includes exact-once, inference, privacy, revocation, D-13 outbox/receipt/cleanup/restore, latency/scale, and complete-loop continuity."
    - "Exit reports exact denominator 30, north-star numerator/percentage/windows, exclusions, versions, incidents, and go/no-go."
  artifacts:
    - path: "docs/operations/runtime-intelligence-beta-report.md"
      provides: "six-week evidence and final decision"
---

<objective>Operate the six-week window from the recorded 30th activation and issue a complete north-star, quality, privacy, deletion, and rollback exit decision.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-08-05-PLAN.md @docs/analytics/runtime-intelligence-beta.sql @docs/contracts/jank-checkout.json @docs/plans/2026-08-19-runtime-intelligence/00-VALIDATION.md</execution_context>
<precondition>Before each weekly/exit jank command run verifier pre-plan/read-only mode for RI-08-06; require configured remote/branch, immutable base ancestry, HEAD=expected_head, and clean tree. No jank dirt is permitted.</precondition>
<tasks>

<task type="auto">
  <name>Task 1: Automate weekly quality, privacy, deletion, continuity, and rollback evidence</name>
  <files>scripts/runtime-rollback-drill.sh, docs/runbooks/runtime-intelligence-beta.md, docs/operations/runtime-intelligence-beta-report.md</files>
  <action>Create target-guarded reverse-dependency flag rollback and weekly command/query checklist. Before every jank-backed check, resolve and verify the checkout contract. From the locked start: W1 activation/exclusions; W2 canonical events; W3 debrief/privacy; W4 lineage/labels/Limited evidence/export/revision; W5 shares/jank imports/evidence; W6 repeat loop/north star. Each week also samples scale/latency, inference versions, migration/reconciliation, authorization/privacy, D-13 outbox pending age, jank receipt, cleanup, restore-generation re-deletion, incidents, and complete play-to-jank continuity. Pause/rollback on critical triggers.</action>
  <acceptance_criteria>Scripts refuse broad targets; every weekly section records command/query versions, timestamps, measured/not_measured, verified jank commit, incidents, and stop/rollback state.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-06-PLAN.md &amp;&amp; bash -n scripts/runtime-rollback-drill.sh &amp;&amp; scripts/runtime-rollback-drill.sh --self-test-target-guards &amp;&amp; scripts/runtime-beta-gate.sh</automated></verify>
  <done>Six-week operations have repeatable, verified-checkout sampling and fail-closed rollback procedures.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Complete six weekly gates and issue beta exit decision</name>
  <files>docs/operations/runtime-intelligence-beta-report.md</files>
  <action>Before the jank exit suite, resolve and verify the checkout. For six consecutive weeks after the immutable start, record all results and any pause/rollback/restart. At exit rerun full gate, privacy/revocation, D-13 request/recovery/local/outbox/jank receipt/partial-failure/response-loss/cleanup/restore drill, and rollback. Report denominator 30, north-star numerator/percentage/windows, candidate exclusions separately, 99.5/95/99 quality gates, inference versions/precision, query p95, receipt lag, incidents, unresolved migrations, and APPROVED or NO-GO. Include no private content/identities or post-beta teams/adapters.</action>
  <acceptance_criteria>All six weeks are nonblank and correctly dated from the 30th activation; final decision uses exact denominators and does not redefine activation/start.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-08-06-PLAN.md &amp;&amp; for w in 1 2 3 4 5 6; do grep -q "Week $w complete:" docs/operations/runtime-intelligence-beta-report.md || exit 1; done &amp;&amp; grep -q 'North-star numerator:' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; grep -q 'North-star denominator: 30' docs/operations/runtime-intelligence-beta-report.md &amp;&amp; grep -Eq 'Beta exit decision: (APPROVED|NO-GO)' docs/operations/runtime-intelligence-beta-report.md</automated></verify>
  <done>The beta concludes with complete evidence and an explicit go/no-go after six measured weeks.</done>
</task>

</tasks>
<integration_worktree_completion>No jank commit or push/merge occurs. Append a read-only exit verification audit entry with unchanged expected_head, then require post-plan remote/base ancestry/expected HEAD/clean tree.</integration_worktree_completion>
<verification>Verified jank checkout, six weekly records, full beta gate, rollback/deletion/privacy/latency/continuity exit drills.</verification>
<success_criteria>The final report truthfully evaluates D-03/D-04 and D-13 while preserving scope fences.</success_criteria>

---
phase: RI-07-jank-cutover-and-evidence-trees
plan: RI-07-05
type: execute
wave: 27
depends_on: ["RI-07-04"]
files_modified:
  - docs/contracts/jank-checkout.json
  - "${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md"
  - "${JANK_CHECKOUT_PATH}/app/app_test.go"
requirements: [RI-JANK-CUTOVER]
autonomous: false
must_haves:
  truths:
    - "Production cutover requires mapped/tombstoned authors, revision-1 parity, revoked REFERENCES, no local auth, both JCT-1 sources, bounded annotations/indicators, and passing role/privacy tests."
    - "Rollback is maintenance/read-only and never restores duplicate account creation."
  artifacts:
    - path: "${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md"
      provides: "migration, session, grants, flags, rollback, and APPROVED evidence"
---

<objective>Rehearse and approve the integrated jank identity/tree cutover only after all JCT behavior passes.</objective>
<execution_context>@docs/contracts/jank-checkout.json @docs/plans/2026-08-19-runtime-intelligence/RI-07-04-PLAN.md</execution_context>
<precondition>Run verifier pre-plan mode for RI-07-05 before jank access; require configured remote/branch, base ancestry, HEAD=expected_head, and clean tree. During tasks use task mode and permit only declared dirt.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Rehearse migrations, grants, identity, both imports, evidence, restore, and rollback</name>
  <files>${JANK_CHECKOUT_PATH}/app/app_test.go, ${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md</files>
  <action>First resolve and verify the recorded checkout. Run PostgreSQL empty/current migrations twice, SQLite development compatibility, identity/tree backfills twice, every owner/guest/moderator/role/evidence/search negative, session rotation/logout, grant revocation, both JCT-1 imports with preview parity, annotation/indicator/source-navigation suites, full tree load, backup restore, maintenance rollback, and evidence-tree read rollback. Record sanitized counts/failures. JANK_SHARED_IDENTITY rollback never enables local signup/login/token minting.</action>
  <acceptance_criteria>Authors and revision-1 rows reconcile; temporary REFERENCES is revoked; no duplicate auth, ownerless mutation, private-base read, import drift, or missing indicator remains.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-05-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; JANK_DB_DRIVER=pgx JANK_DB_DSN="$JANK_TEST_DSN" go test ./... -race</automated></verify>
  <done>The production-shaped jank cutover and rollback evidence is ready for review.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Approve canonical-user and evidence-tree cutover</name>
  <files>${JANK_CHECKOUT_PATH}/docs/runbooks/shared-identity-cutover.md</files>
  <action>First resolve and verify the recorded checkout. Review mapping conflicts, guest-claim preservation, revision counts, no-local-auth proof, grant/revoke, role negatives, both JCT-1 imports, trees/revisions/forks/discussion, bounded annotations, all compact indicators and source navigation, evidence/revocation/search/keyboard UX, restore, and rollback. Record APPROVED or REJECTED with evidence. Rejection/absence leaves this plan incomplete and flags off.</action>
  <acceptance_criteria>Approval requires zero active duplicate account path, ownerless mutation, private/base evidence access, import-version drift, or missing required indicator/navigation.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-05-PLAN.md &amp;&amp; grep -q 'Jank production cutover decision: APPROVED' "$JANK_CHECKOUT_PATH/docs/runbooks/shared-identity-cutover.md"</automated></verify>
  <done>RI-07 cutover is APPROVED; rejection/absence halts RI-08 and is not dependency completion.</done>
</task>

</tasks>
<integration_worktree_completion>Create one atomic unpushed jank commit for the runbook/test evidence after prepare-commit, advance expected_head with its audit record, then require post-plan remote/base ancestry/new expected HEAD/clean tree.</integration_worktree_completion>
<verification>Verified checkout, full jank suites, and APPROVED production cutover decision.</verification>
<success_criteria>JCT-1 through JCT-7 are production-gated with shared-row identity and privacy-safe evidence intact.</success_criteria>

---
phase: RI-07-jank-cutover-and-evidence-trees
plan: RI-07-02
type: execute
wave: 24
depends_on: ["RI-07-01"]
files_modified:
  - docs/contracts/jank-checkout.json
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820151000_evidence_trees.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/postgres/20260820151000_evidence_trees.down.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820151000_evidence_trees.up.sql"
  - "${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820151000_evidence_trees.down.sql"
  - "${JANK_CHECKOUT_PATH}/app/models.go"
  - "${JANK_CHECKOUT_PATH}/app/store/trees.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_trees.go"
  - "${JANK_CHECKOUT_PATH}/app/app_test.go"
requirements: [JCT-2, JCT-6, D-20, D-21]
autonomous: true
must_haves:
  truths:
    - "JCT-2 uses exactly supports, enables, finds, protects, recurs, replaces, competes-with, punishes, meta-answer, or bounded custom."
    - "Custom labels are searchable but never global filter categories."
    - "JCT-6 revisions/diffs/forks are immutable, attributed, standalone, previewed before save, and have independent optional threads."
  artifacts:
    - path: "${JANK_CHECKOUT_PATH}/app/store/trees.go"
      provides: "standalone immutable tree/revision/relation/fork store"
  key_links:
    - from: "tree edits/reparent/reorder"
      to: "card_tree_revisions"
      via: "preview hash and append-only revision transaction"
---

<objective>Install and implement standalone tree/revision/relation/fork foundations before any import or evidence link. Output: verified-checkout migrations, legacy revision-1 backfill, exact relationship validation, preview parity, diffs, and forks.</objective>
<execution_context>@docs/contracts/jank-checkout.json @docs/plans/2026-08-19-runtime-intelligence/RI-07-01-PLAN.md @docs/product/2026-08-17-jank-evidence-backed-card-trees-prd.md</execution_context>
<precondition>Run verifier `--mode pre-plan --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-02-PLAN.md` before reads/edits; require configured remote/branch, immutable base ancestry, HEAD=expected_head, and clean tree. During tasks use `--mode task`; only declared files may be dirty.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Install standalone trees, immutable revisions, relations, and forks</name>
  <files>${JANK_CHECKOUT_PATH}/migrations/postgres/20260820151000_evidence_trees.up.sql, ${JANK_CHECKOUT_PATH}/migrations/postgres/20260820151000_evidence_trees.down.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820151000_evidence_trees.up.sql, ${JANK_CHECKOUT_PATH}/migrations/sqlite/20260820151000_evidence_trees.down.sql, ${JANK_CHECKOUT_PATH}/app/models.go, ${JANK_CHECKOUT_PATH}/app/store/trees.go, ${JANK_CHECKOUT_PATH}/app/app_test.go</files>
  <action>First resolve and verify the recorded checkout. Add private/unlisted/public standalone trees, nullable legacy board/thread scope, immutable revisions/nodes, canonical/unresolved cards, exact nine relationships plus custom, bounded custom label, current revision, fork source/backlinks, and optional discussion pointers. Backfill every legacy tree as revision 1 without changing URL/hierarchy; preserve ambiguous cards unresolved. Require relationship for new evidence-backed trees after migration window; keep it optional for legacy revision 1.</action>
  <acceptance_criteria>Every legacy tree has revision 1; the exact vocabulary passes; unknown/free-form relation kinds fail; custom labels are bounded/searchable but absent from global filters; forks are independent.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-02-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestEvidenceTreeMigration|TestRelationshipVocabularyV1|TestCustomRelationshipBounds|TestTreeFork' -race -v</automated></verify>
  <done>JCT-2/JCT-6 persistence is immutable, exact-vocabulary, standalone, and backward compatible.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement previewed edit, diff, and attributed fork handlers</name>
  <files>${JANK_CHECKOUT_PATH}/app/store/trees.go, ${JANK_CHECKOUT_PATH}/app/handlers_trees.go, ${JANK_CHECKOUT_PATH}/app/app_test.go</files>
  <action>First resolve and verify the recorded checkout. Implement create, preview-edit, save with matching preview hash, reorder/reparent, exact relationship/custom-label validation, immutable revision, diff, and attributed fork. Diffs include added, removed, moved, relabeled, and evidence-changed nodes. Enforce canonical owner authorization on every mutation; guest/public/moderator cannot mutate by URL knowledge. A fork gets independent owner/current revision/optional thread and source tree/revision backlinks.</action>
  <acceptance_criteria>Preview/save parity, authorization negatives, all diff categories, custom bounds, fork attribution, and independent thread ownership pass.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-02-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestTreePreviewSaveParity|TestTreeAuthorization|TestTreeDiff|TestTreeFork' -race -v</automated></verify>
  <done>Owners can evolve and fork trees without overwriting history or sharing ownership/thread state.</done>
</task>

</tasks>
<integration_worktree_completion>After prepare-commit, create one atomic unpushed jank commit from declared files, advance expected_head with an audit record, then require post-plan remote/base ancestry/new expected HEAD/clean tree. Refuse undeclared dirt, divergence, or unexpected head.</integration_worktree_completion>
<verification>Every command uses the verified checkout; PostgreSQL/SQLite migration, vocabulary, preview, diff, fork, and authorization suites pass.</verification>
<success_criteria>JCT-2/JCT-6 are complete before JCT-1 imports create any tree data.</success_criteria>

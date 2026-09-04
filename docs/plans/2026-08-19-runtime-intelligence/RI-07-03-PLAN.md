---
phase: RI-07-jank-cutover-and-evidence-trees
plan: RI-07-03
type: execute
wave: 25
depends_on: ["RI-07-02", "RI-06-03"]
files_modified:
  - docs/contracts/jank-checkout.json
  - server/jank_tree_import.go
  - server/jank_tree_import_test.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - "${JANK_CHECKOUT_PATH}/app/store/trees.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_trees.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_api.go"
  - "${JANK_CHECKOUT_PATH}/app/app_test.go"
requirements: [JCT-1]
autonomous: true
scope_rationale: "The ten files are one JCT-1 end-to-end import contract: vEDH authorizes and pins both source families, while the verified jank handlers preview and save the same manifest. The gqlgen artifacts and jank redemption tests are required to prove no private-base grant or preview drift."
must_haves:
  truths:
    - "JCT-1 supports private owner-authorized lineage/snapshot manifest import without publication and authorization-filtered published-analysis slice import."
    - "Both source paths have preview/save hash parity, exact calculation/source-version pinning, subset selection, conflict display, and authorization negatives."
    - "Neither path mutates vEDH snapshots or grants jank access to private/base tables."
  artifacts:
    - path: "server/jank_tree_import.go"
      provides: "typed private-manifest and published-analysis import source contract"
    - path: "${JANK_CHECKOUT_PATH}/app/handlers_trees.go"
      provides: "previewed redemption into a private draft tree"
  key_links:
    - from: "private manifest or published-analysis slice"
      to: "jank draft tree revision 1"
      via: "session/principal-bound manifest, pinned versions, and matching preview hash"
---

<objective>Deliver both JCT-1 source paths end to end: private owner manifest without publication and authorization-filtered published-analysis import. Output: generated vEDH contracts plus verified-jank preview/redeem/save behavior.</objective>
<execution_context>@docs/contracts/jank-checkout.json @docs/contracts/evidence-api-v1.md @docs/plans/2026-08-19-runtime-intelligence/RI-07-02-PLAN.md</execution_context>
<precondition>Run verifier pre-plan mode for RI-07-03 before jank access; require configured remote/branch, base ancestry, HEAD=expected_head, and clean tree. During tasks use task mode, which permits only declared plan dirt.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create owner-private and published-analysis import manifests</name>
  <files>server/jank_tree_import.go, server/jank_tree_import_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go</files>
  <action>Add previewJankTreeImport with an exclusive source union: ownerPrivateSnapshot(lineageID, snapshotID, filter) or publishedAnalysis(publicationID, publicationVersionID, filter). Private mode reauthorizes canonical owner and returns a short-lived one-use session-bound manifest without creating publication. Published mode resolves only the authorization-filtered evidence_api slice/version. Both return included canonical/display/unresolved cards, proposed roots, conflicts, subset, pinned snapshot/source IDs, calculation/source/publication versions, redactions, and canonical preview hash. Redemption reauthorizes and rejects stale, replayed, altered, deleted, unauthorized, or recalculated source versions.</action>
  <acceptance_criteria>Both source families produce bounded manifests with exact version pins; private mode requires owner and no publication; published mode cannot bypass publication authorization.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestJankTreePrivateImportManifest|TestJankTreePublishedAnalysisImport|TestJankTreeImportPreviewParity|TestJankTreeImportAuthorization' -race -v</automated></verify>
  <done>vEDH emits two explicit, authorization-safe, source-pinned JCT-1 manifests.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Preview and save either import source into a private draft</name>
  <files>${JANK_CHECKOUT_PATH}/app/store/trees.go, ${JANK_CHECKOUT_PATH}/app/handlers_trees.go, ${JANK_CHECKOUT_PATH}/app/handlers_api.go, ${JANK_CHECKOUT_PATH}/app/app_test.go</files>
  <action>First resolve and verify the recorded checkout. Accept only the two typed vEDH manifest modes. Render included/unresolved cards, roots, subset, conflicts, source kind, pinned snapshot/publication/calculation/source versions, and redactions before save. Save requires the same principal/session, unused manifest, unchanged source/version, and matching preview hash; it creates private draft revision 1 and stores source metadata. Cover owner/unrelated/public/guest/moderator, stale/replay/tamper/recalculation, subset, and conflict cases. Never query private vEDH base tables or mutate the source snapshot.</action>
  <acceptance_criteria>Private and published imports have byte/hash preview-save parity, complete auth negatives, and exact source/version provenance on the tree.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-03-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestPrivateTreeImport|TestPublishedAnalysisTreeImport|TestTreeImportPreviewSaveParity|TestTreeImportAuthorization' -race -v</automated></verify>
  <done>JCT-1 works from either approved source without publication coercion, version drift, or private database access.</done>
</task>

</tasks>
<integration_worktree_completion>Create one atomic unpushed jank commit after prepare-commit, advance expected_head and append its audit record, then require post-plan remote/base ancestry/new expected HEAD/clean tree.</integration_worktree_completion>
<verification>vEDH manifest/generation suites plus verified-checkout import preview/save and authorization suites.</verification>
<success_criteria>JCT-1 is owned exactly here and every acceptance clause is executable for both source paths.</success_criteria>

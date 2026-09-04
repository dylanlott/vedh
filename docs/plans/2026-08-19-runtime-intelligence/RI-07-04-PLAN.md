---
phase: RI-07-jank-cutover-and-evidence-trees
plan: RI-07-04
type: execute
wave: 26
depends_on: ["RI-07-03"]
files_modified:
  - docs/contracts/jank-checkout.json
  - "${JANK_CHECKOUT_PATH}/app/store/evidence.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_evidence.go"
  - "${JANK_CHECKOUT_PATH}/app/handlers_html.go"
  - "${JANK_CHECKOUT_PATH}/app/router.go"
  - "${JANK_CHECKOUT_PATH}/app/app.go"
  - "${JANK_CHECKOUT_PATH}/app/app_test.go"
  - "${JANK_CHECKOUT_PATH}/templates/card_tree.html"
  - "${JANK_CHECKOUT_PATH}/templates/card_trees.html"
  - "${JANK_CHECKOUT_PATH}/templates/card_tree_diff.html"
  - "${JANK_CHECKOUT_PATH}/docs/contracts/evidence-api-v1.md"
requirements: [JCT-3, JCT-4, JCT-5, D-14]
autonomous: true
scope_rationale: "The ten verified-checkout files are one server-rendered evidence-tree surface: safe evidence resolution, bounded annotation/indicator storage, routes/templates, discussion links, search, and their shared integration fixture."
must_haves:
  truths:
    - "JCT-3 annotation kind is exactly observation, player-assessment, hypothesis, revision-plan, or counterexample; evidence links remain stable/revocation-safe."
    - "JCT-4 compact indicators expose eligible, observed, acted, contributor, underperformer, and source counts with denominators, versions, confidence, and Limited evidence."
    - "Selecting any indicator navigates to its authorization-filtered vEDH analysis slice or source evidence."
    - "JCT-5 discussion is optional/lazy and stable quotes never couple tree and thread lifetimes."
    - "The beta text contract enforces exact code-point bounds, plain-text escaped rendering, safe vEDH-only links, and XSS-safe public search indexing."
  artifacts:
    - path: "${JANK_CHECKOUT_PATH}/app/store/evidence.go"
      provides: "read-only evidence resolution, bounded annotation kinds, indicators, and navigation targets"
  key_links:
    - from: "tree indicator/evidence link"
      to: "evidence_api v1 authorized analysis/source URL"
      via: "live status check and stable publication/version/source reference"
---

<objective>Complete bounded evidence-linked annotations, all compact runtime indicators, authorized source navigation, optional discussion, privacy-safe search, and scale for JCT-3 through JCT-5.</objective>
<execution_context>@docs/contracts/jank-checkout.json @docs/contracts/evidence-api-v1.md @docs/plans/2026-08-19-runtime-intelligence/RI-07-03-PLAN.md</execution_context>
<precondition>Run verifier pre-plan mode for RI-07-04 before jank access; require configured remote/branch, base ancestry, HEAD=expected_head, and clean tree. During tasks use task mode and allow dirt only in declared files.</precondition>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement bounded annotation kinds, stable evidence links, and source navigation</name>
  <files>${JANK_CHECKOUT_PATH}/app/store/evidence.go, ${JANK_CHECKOUT_PATH}/app/handlers_evidence.go, ${JANK_CHECKOUT_PATH}/app/app_test.go, ${JANK_CHECKOUT_PATH}/docs/contracts/evidence-api-v1.md</files>
  <action>First pass task-mode checkout verification. Constrain annotation kind to exactly observation, player-assessment, hypothesis, revision-plan, or counterexample; reject `note` and all unknown/free-form kinds. Document conservative beta contract parameters measured in Unicode code points and enforce them at request, domain, and database boundaries: annotation body 4,000; annotation/evidence label 120; annotation tags serialization 500; tree title 160; tree description 2,000; revision message 500; custom relationship label 80; imported canonical/display/original/unresolved card label 256; imported source/display/conflict text 500; and at most 500 imported nodes per manifest. Return 413 for overlong request/import aggregates and field errors for individual text violations; never truncate silently. Attach published evidence IDs with referentially valid columns. Validate every stored navigation target as an opaque relative path or absolute HTTPS URL on the RI-00-approved vEDH origin; reject credentials, fragments carrying data, scheme-relative targets, redirects to another origin, `javascript`, `data`, `file`, embeds, iframes, images, and arbitrary remote content. Preview exact sample/version/redaction/authorization metadata and recheck evidence_api on every detail/navigation request.</action>
  <acceptance_criteria>Exact-kind validation, owner/public/revoked/redacted/deleted/forged-link matrix, preview parity, and indicator-to-authorized-source navigation pass; no raw payload/pseudonym mapping is stored.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-04-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestAnnotationKindVocabulary|TestEvidenceLink|TestEvidenceAuthorization|TestEvidenceRevoked|TestIndicatorSourceNavigation|TestAnnotationTextSecurityBounds' -race -v</automated></verify>
  <done>JCT-3 annotations and evidence links are bounded, inspectable, authorization-safe, and revocation-safe.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Render every compact indicator, revisions/forks, and lazy discussion</name>
  <files>${JANK_CHECKOUT_PATH}/app/handlers_html.go, ${JANK_CHECKOUT_PATH}/app/router.go, ${JANK_CHECKOUT_PATH}/app/app.go, ${JANK_CHECKOUT_PATH}/templates/card_tree.html, ${JANK_CHECKOUT_PATH}/templates/card_trees.html, ${JANK_CHECKOUT_PATH}/templates/card_tree_diff.html</files>
  <action>After task-mode verification, render all imported and annotation text as plain text through Go `html/template` contextual escaping, matching the existing card_tree.html interpolation path; do not cast tree text to `template.HTML` or pass it through Goldmark. Replace the current client-side card-name `innerHTML` rewrite with text-node/DOM construction so imported labels cannot become markup or attributes. If an existing forum post uses Markdown, keep the current Goldmark then `bluemonday.UGCPolicy()` path in markdown.go, but tree annotations and evidence summaries admit no raw HTML, embeds, remote media, or Markdown URL rendering. Render the six indicators with denominator/calculation/evidence/inference metadata, Limited evidence, revisions/forks, and lazy independent discussion; validated links open the exact authorized source with safe rel attributes.</action>
  <acceptance_criteria>HTTP/browser tests cover all six indicators, source navigation authorization, JCT-4 metadata, and JCT-5 optional thread/quote/backlink lifecycle without ranking language.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-04-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestTreeRuntimeIndicators|TestIndicatorSourceNavigation|TestTreeDiscussLazyThread|TestStableEvidenceQuote|TestTreeKeyboardFlow|TestTreeRenderXSS' -race -v</automated></verify>
  <done>Readers can scan every required context signal and navigate to authorized evidence while tree/thread lifecycles remain independent.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Enforce search privacy and tree/evidence scale gates</name>
  <files>${JANK_CHECKOUT_PATH}/app/app_test.go, ${JANK_CHECKOUT_PATH}/app/store/evidence.go, ${JANK_CHECKOUT_PATH}/app/handlers_html.go</files>
  <action>After task-mode verification, index only length-validated public card names, the exact five annotation kinds, controlled/custom relationship labels, escaped plain-text public annotation bodies, and safe evidence summaries. The index stores text tokens, never rendered HTML, URLs, embeds, private/revoked payloads, or pseudonym mappings. Test stored XSS payloads, template rendering, the prior client-side card-name rewrite, search-result snippets/highlighting, URL-scheme/origin bypasses, and each exact bound plus bound+1 for direct create and imported manifests. Prove malicious content remains inert in storage/render/search and oversized input is rejected without partial tree/index writes. Exercise 500 nodes and 2,000 annotations/evidence links with documented p95, cursor/bounded reads, indicator navigation, and revocation lookup.</action>
  <acceptance_criteria>Search privacy, exact-kind filtering, all indicators, source navigation, and scale fixtures complete without unbounded queries.</acceptance_criteria>
  <verify><automated>JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json) &amp;&amp; scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-07-04-PLAN.md &amp;&amp; cd "$JANK_CHECKOUT_PATH" &amp;&amp; go test ./... -run 'TestTreeSearchPrivacy|TestTreeSearchIndexXSS|TestOversizedTreeTextRejected|TestAnnotationKindVocabulary|TestCustomRelationSearch|TestTreeLoad500|TestEvidenceLinks2000' -race -v</automated></verify>
  <done>JCT-3/JCT-4/JCT-5 remain privacy-safe and measurable at beta scale.</done>
</task>

</tasks>
<integration_worktree_completion>Create one atomic unpushed jank commit after prepare-commit, advance expected_head and append the audit record, then require post-plan remote/base ancestry/new expected HEAD/clean tree.</integration_worktree_completion>
<verification>Every command uses verified checkout; evidence/annotation/indicator/navigation/discussion/search/load suites pass.</verification>
<success_criteria>Jank remains an interpretation layer over stable sanitized evidence with complete compact context.</success_criteria>

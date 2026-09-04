---
phase: RI-05-private-lineage-analytics
plan: RI-05-03
type: execute
wave: 18
depends_on: ["RI-05-02"]
files_modified:
  - app/src/components/evidence/LimitedEvidence.vue
  - app/src/components/analysis/LineageFilters.vue
  - app/src/components/analysis/SnapshotComparison.vue
  - app/src/components/analysis/CardRuntimeTable.vue
  - app/src/views/DeckAnalysisView.vue
  - app/src/router/index.ts
  - app/src/graphql/queries.ts
  - app/src/graphql/mutations.ts
  - app/src/types/generated.ts
  - app/__tests__/DeckAnalysisView.spec.ts
  - app/e2e/deck-analysis.spec.ts
  - docs/runbooks/lineage-projector-cutover.md
requirements: [D-01, D-03, D-10]
autonomous: false
scope_rationale: "The twelve files are one protected Deck Analysis route and its generated operations, source-preserving filters/export/revision actions, browser fixtures, and cutover record. Splitting would cause repeated ownership of the same route, query/mutation documents, generated types, and E2E journey."
must_haves:
  truths:
    - "The workspace displays facts from n=1, exact denominators/uncertainty, and Limited evidence when either side has fewer than three games."
    - "Inspect games and export preserve active filters/calculation version."
    - "A user can preview and create a documented successor revision."
  artifacts:
    - path: "app/src/views/DeckAnalysisView.vue"
      provides: "private lineage decision workspace"
  key_links:
    - from: "DeckAnalysisView"
      to: "analysis/export/revision APIs"
      via: "typed generated operations"
---

<objective>Ship the private lineage workspace and block its cohort switch on reconciliation, scale, and low-sample usability evidence.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-05-02-PLAN.md @app/src/views/GameAnalysisView.vue</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Build lineage analysis, comparison, export, and revision UX</name>
  <files>app/src/components/evidence/LimitedEvidence.vue, app/src/components/analysis/LineageFilters.vue, app/src/components/analysis/SnapshotComparison.vue, app/src/components/analysis/CardRuntimeTable.vue, app/src/views/DeckAnalysisView.vue, app/src/router/index.ts, app/src/graphql/queries.ts, app/src/graphql/mutations.ts, app/src/types/generated.ts, app/__tests__/DeckAnalysisView.spec.ts, app/e2e/deck-analysis.spec.ts</files>
  <action>Add protected lineage route with outcomes/pace/table equivalents, snapshot history/diff, auditable user-named snapshot labels, evidence-classed card activity, assessments, typed filters/unknown/exclusions, calculation definitions, Inspect games, reproducible export, and documented revision preview. Show pinned historical source revision separately from latest observed source revision; when they differ, render an accessible source changed since this snapshot notice that never implies history changed. Render n=1; show Limited evidence on 0–2 either side without hiding controls/data; prohibit unsupported ranking/causal language and testing-group ownership UI per D-19. E2E covers label create/rename/history/authorization, unchanged/changed/unknown source notice, n=1, 2-vs-3, filters, drill-down, export contents, correction, and revision link.</action>
  <acceptance_criteria>UI tests prove exact sample language, accessible charts/tables, preserved filter/version context, export metadata, and immutable revision flow.</acceptance_criteria>
  <verify><automated>npm --prefix app test -- DeckAnalysisView.spec.ts &amp;&amp; npm --prefix app run type-check &amp;&amp; npm --prefix app run build &amp;&amp; npm --prefix app run test:e2e -- deck-analysis.spec.ts</automated></verify>
  <done>The user can inspect, challenge, export, and act on private evidence without overstated certainty.</done>
</task>
<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Approve projector/read switch and revision usability</name>
  <files>docs/runbooks/lineage-projector-cutover.md</files>
  <action>Review shadow reconciliation, 10k rebuild/query p95, correction/deletion/access rebuild, pointer rollback, snapshot label audit/source-change notice, n=1/2-vs-3/unknown/exclusion language, drill-down/export, and revision flow. Record APPROVED or REJECTED with evidence; rejection/absence keeps VEDH_RI_ANALYTICS_READ off and leaves this plan incomplete.</action>
  <acceptance_criteria>Decision cites source parity, latency, privacy negatives, and human low-sample comprehension.</acceptance_criteria>
  <verify><automated>grep -q 'Analytics cohort decision: APPROVED' docs/runbooks/lineage-projector-cutover.md</automated></verify>
  <done>Private analytics cohort enablement is APPROVED and rollback-ready; rejection/absence halts RI-06.</done>
</task>
</tasks>
<verification>Vue unit/type/build/E2E plus projector/server phase gate.</verification>
<success_criteria>The D-01 evidence-to-revision transition is usable and D-10 remains exact.</success_criteria>

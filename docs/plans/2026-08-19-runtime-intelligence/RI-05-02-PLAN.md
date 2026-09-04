---
phase: RI-05-private-lineage-analytics
plan: RI-05-02
type: execute
wave: 17
depends_on: ["RI-05-01"]
files_modified:
  - server/deck_analysis.go
  - server/deck_analysis_test.go
  - server/deck_revisions.go
  - server/deck_revisions_test.go
  - server/analysis_export.go
  - server/analysis_export_test.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/graphql.go
  - server/metrics.go
  - pkg/telemetry/metrics.go
requirements: [RDA-4, RDA-5, RDA-6]
autonomous: true
scope_rationale: "The thirteen files are one generated Deck Analysis contract: comparison/filter/drill-down, reproducible export, and explicit successor revision all share the normalized analysis request, authorization, gqlgen artifacts, and bounded telemetry seam."
must_haves:
  truths:
    - "RDA-4 compares exact snapshots and source-pinned cohorts without better/worse claims."
    - "RDA-5 typed filters operate only on recorded facts/submitted tags and preserve unknown."
    - "RDA-6 every metric drills to accessible sources and analysis export includes calculation definitions, filters, versions, and stable source IDs."
  artifacts:
    - path: "server/analysis_export.go"
      provides: "authorization-filtered reproducible analysis export"
  key_links:
    - from: "analysis export"
      to: "calculation definition/filter/source IDs"
      via: "same normalized query contract as interactive analysis"
---

<objective>Expose the RDA-4 comparison, RDA-5 context filtering, and complete RDA-6 evidence/export contract. Output: bounded GraphQL, owner authorization, source drill-down, export, and revision link service.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-05-01-PLAN.md @server/game_analysis.go @server/deck_lineages.go</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Implement typed analysis, comparison, filters, and drill-down</name>
  <files>server/deck_analysis.go, server/deck_analysis_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/graphql.go, server/metrics.go, pkg/telemetry/metrics.go</files>
  <action>Add deckAnalysis, snapshotComparison, cardRuntimeEvidence, cursor-paged analysisSourceGames, renameSnapshotLabel, snapshotLabelHistory, and snapshotSourceStatus. Return the user-visible current label plus immutable label audit, pinned source revision, latest observed source revision/check time, and a changed-since-pinned notice without fetching/mutating historical data on read. Bound snapshot/date-session/format/result/participant-count/seat/authorized-opponent-leader/ending-turn/lesson-tag filters; expose UNKNOWN and preflight eligible count. Return calculation definitions/version, normalized active filters, exact included/excluded counts, uncertainty, and accessible source IDs. Recalculate/redact inaccessible sources. Add owner/jank/forged-source/deleted-user negatives, ensure no testing-group owner path per D-19, and measure 10k p95 at or below one second.</action>
  <acceptance_criteria>All chart/table/card values have Inspect games; filters and scale gates pass; arbitrary SQL/grouping is rejected.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestDeckAnalysis|TestSnapshotComparison|TestSnapshotLabelRevision|TestSnapshotSourceChangeNotice|TestAnalysisAuthorizationNegative|TestDeckAnalysis10k' -race -v</automated></verify>
  <done>RDA-4/RDA-5 interactive APIs are bounded, source-pinned, and owner-authorized.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Deliver reproducible analysis export with definitions and sources</name>
  <files>server/analysis_export.go, server/analysis_export_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go</files>
  <action>Add asynchronous bounded export for the current authorized analysis slice. Include schema/calculation definition and version, normalized filter parameters, inclusion/exclusion rules, numerator/denominator/uncertainty, stable lineage/snapshot/game/event/review/inference source IDs the viewer may access, generated time, and redaction manifest. Never export raw hidden payloads or inaccessible/private note text. Re-running the same pinned request over unchanged accessible sources yields the same canonical content hash.</action>
  <acceptance_criteria>Golden export fixtures reconcile to interactive results and prove definitions, filters, versions, and stable source IDs are present.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestAnalysisExport|TestAnalysisExportAuthorization|TestAnalysisExportReproducible' -race -v</automated></verify>
  <done>The missing RDA-6 export clause is implemented with complete reproducibility metadata.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 3: Link an explicit hypothesis to a successor immutable snapshot</name>
  <files>server/deck_revisions.go, server/deck_revisions_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go</files>
  <action>Implement createDeckRevision(sourceSnapshotID, normalizedDeck, hypothesis or sourceReviewNextActionID). Preview exact added/removed/quantity/leader/unresolved changes; create a successor immutable snapshot; append deck_revision_link and resolve the linked next action. Require explicit owner input and never recommend/apply cuts or infer lineage from similarity.</action>
  <acceptance_criteria>Revision creation is auditable, immutable, idempotent, and completes D-03's documented-revision fact.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestCreateDeckRevision|TestResolveReviewNextAction' -race -v</automated></verify>
  <done>Evidence and a recorded hypothesis can produce a traceable immutable successor without automatic deck decisions.</done>
</task>
</tasks>
<verification>`make generate`; deck analysis/export/revision server suites; 10k latency.</verification>
<success_criteria>RDA-4 through RDA-6 are executable, including the previously missing complete export contract.</success_criteria>

---
phase: RI-03-canonical-runtime-events
plan: RI-03-04
type: execute
wave: 12
depends_on: ["RI-03-03"]
files_modified:
  - server/game_analysis.go
  - server/game_analysis_test.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/graphql.go
  - app/src/components/evidence/EvidenceBadge.vue
  - app/src/components/evidence/TimelineFilters.vue
  - app/src/views/GameAnalysisView.vue
  - app/src/graphql/queries.ts
  - app/src/types/generated.ts
  - app/__tests__/GameAnalysisView.spec.ts
  - docs/runbooks/runtime-event-cutover.md
requirements: [RGI-5]
autonomous: false
scope_rationale: "The fourteen files are one typed timeline contract across gqlgen and the existing Game Analysis route. The generated server/client artifacts and the single view cannot be independently verified without the matching API, evidence components, test, and cutover record."
must_haves:
  truths:
    - "RGI-5 exposes an authorized typed cursor timeline with filters and source inspection."
    - "The browser does not derive analytical facts from raw event JSON."
    - "Every Derived/Inferred timeline item renders its exact calculation definition/version and provides an authorized drill-down to the definition and ordered source events."
    - "No probabilistic category default-displays until a reviewer approves its exact versioned holdout evidence."
  artifacts:
    - path: "server/game_analysis.go"
      provides: "typed authorized timeline"
  key_links:
    - from: "GameAnalysisView.vue"
      to: "runtimeTimeline"
      via: "typed cursor query and evidence metadata"
---

<objective>Ship RGI-5 game inspection and make inference enablement a genuine blocking approval. Output: typed API/UI, 10k latency evidence, and an approval/rejection record that controls default display.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-03-03-PLAN.md @app/src/views/GameAnalysisView.vue @server/schema.graphql</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Add authorized typed cursor timeline</name>
  <files>server/game_analysis.go, server/game_analysis_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/graphql.go</files>
  <action>Add bounded runtimeTimeline(gameID, first, after, filter, showSpeculative=false), returning result/context, typed events, evidence class, provenance/source IDs, confidence/inference version/causal class, exact calculationDefinitionID and calculationDefinitionVersion, and unknown/not_recorded/not_applicable. Add an authorized calculation-definition drill-down returning the immutable user-readable definition and ordered source-event links for the exact version; never substitute the latest version. Filter by player/card/class/turn. Authorize the fact, definition, and every source; cover hidden zones, unrelated user, guest mismatch, moderator, deleted principal, jank role, missing definition, mismatched version, and inaccessible source. Retain gameLogs as measured compatibility. Benchmark 10,000 events at p95 at or below 750 ms under documented beta profile.</action>
  <acceptance_criteria>Pagination is stable/bounded, negative matrix passes, and performance evidence is recorded.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestRuntimeTimeline|TestRuntimeTimelineAuthorizationNegative|TestRuntimeTimeline10k|TestTimelineCalculationDefinitionExactVersion|TestTimelineCalculationSourceAuthorization' -race -v</automated></verify>
  <done>The RGI-5 API is typed, authorized, source-linked, and meets the beta latency gate.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Replace raw browser interpretation with evidence UI</name>
  <files>app/src/components/evidence/EvidenceBadge.vue, app/src/components/evidence/TimelineFilters.vue, app/src/views/GameAnalysisView.vue, app/src/graphql/queries.ts, app/src/types/generated.ts, app/__tests__/GameAnalysisView.spec.ts</files>
  <action>Render cursor pages, result/participants/snapshots/duration/ending turn, filters, source drill-down, Observed/Derived/Inferred badges, confidence/inference version/causal-hypothesis language, and the exact calculation-definition version behind every Derived/Inferred item. The drill-down displays the typed human-readable definition and only authorized ordered sources; missing, superseded, or inaccessible definitions/sources fail closed without falling forward to a newer calculation. Preserve Show speculative and accessible loading/error/empty/table states. Remove analytical dependence on raw Payload JSON. Keep default inference rendering off unless the checkpoint's stored decision approves the exact category/version.</action>
  <acceptance_criteria>Component tests prove typed data only, speculative opt-in, accurate causality labels, and hidden-data absence.</acceptance_criteria>
  <verify><automated>npm --prefix app test -- GameAnalysisView.spec.ts -t 'calculation definition|source authorization|typed timeline' &amp;&amp; npm --prefix app run type-check &amp;&amp; npm --prefix app run build</automated></verify>
  <done>Players can inspect RGI-5 evidence without client-side fact invention or unapproved probabilistic defaults.</done>
</task>
<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 3: Decide canonical event and inference default enablement</name>
  <files>docs/runbooks/runtime-event-cutover.md</files>
  <action>Review reconciliation/fault tests, each category/version's labeled fixture and precision, source drill-down, speculative UI, causal language, privacy negatives, and 10k latency. Record separate Canonical event cohort decision and Inference default-display decision entries with APPROVED or REJECTED, reviewer/time, approved category/version list, and reasons. Canonical-event rejection/absence leaves this plan incomplete and blocks RI-04. Inference rejection is permitted only with canonical events APPROVED; it keeps default inference off while typed observed evidence may proceed.</action>
  <acceptance_criteria>Silence is rejection; only explicitly listed qualifying category/versions may default-display; 94.9% never qualifies.</acceptance_criteria>
  <verify><automated>grep -q 'Canonical event cohort decision: APPROVED' docs/runbooks/runtime-event-cutover.md &amp;&amp; grep -Eq 'Inference default-display decision: (APPROVED|REJECTED)' docs/runbooks/runtime-event-cutover.md</automated></verify>
  <done>Canonical events are APPROVED for the cohort; inference default-display follows only its exact category/version decision and stays off when rejected.</done>
</task>
</tasks>
<verification>`make generate`; server timeline suites; Vue tests/type-check/build; signed decision record.</verification>
<success_criteria>RGI-5 is shipped and inference implementation is cleanly separated from human enablement authority.</success_criteria>

---
phase: RI-05-private-lineage-analytics
plan: RI-05-01
type: execute
wave: 16
depends_on: ["RI-04-03"]
files_modified:
  - persistence/migrations/20260820130000_lineage_analytics_v1.up.sql
  - persistence/migrations/20260820130000_lineage_analytics_v1.down.sql
  - persistence/migrations_test/20260820130000_lineage_analytics_v1.up.sql
  - persistence/migrations_test/20260820130000_lineage_analytics_v1.down.sql
  - pkg/analysis/eligibility.go
  - pkg/analysis/eligibility_test.go
  - pkg/analysis/statistics.go
  - pkg/analysis/statistics_test.go
  - pkg/projection/lineage_projector.go
  - pkg/projection/lineage_projector_test.go
  - pkg/projection/testdata/lineage-golden-v1.json
  - server/deck_lineage_metadata.go
  - server/deck_lineage_metadata_test.go
requirements: [RDA-1, RDA-2, RDA-3]
autonomous: true
scope_rationale: "The thirteen files are one RDA-1 through RDA-3 analytical source contract: immutable lineage label/source metadata plus the eligibility/statistics projector and its mirrored schema/golden corpus. Labels and source-change notices must version with the projection source they describe."
must_haves:
  truths:
    - "RDA-1 analyzes immutable snapshots within stable, auditable individual-owner lineages and supports auditable user-named snapshot labels."
    - "RDA-1 distinguishes the pinned historical source revision from the latest observed source revision and shows a change notice without rewriting history."
    - "RDA-2 exposes eligible/excluded outcome and pace distributions from the first game with exact denominators."
    - "RDA-3 keeps included, observed, derived, inferred, acted, and assessed card states distinct and source-linked."
    - "Either comparison side below three games shows Limited evidence without hiding values."
  artifacts:
    - path: "pkg/projection/lineage_projector.go"
      provides: "versioned private analytical projections"
  key_links:
    - from: "lineage projector"
      to: "results/events/inferences/submitted reviews"
      via: "versioned source replay; drafts excluded"
---

<objective>Build the deterministic private analytical foundation for RDA-1 through RDA-3. Output: golden fixtures, statistics/eligibility libraries, versioned read models, and rebuildable projector.</objective>
<execution_context>@docs/product/2026-08-17-runtime-deck-analysis-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-04-03-PLAN.md</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Lock eligibility, statistics, evidence classes, and scale fixtures</name>
  <files>pkg/analysis/eligibility_test.go, pkg/analysis/statistics_test.go, pkg/projection/testdata/lineage-golden-v1.json, pkg/projection/lineage_projector_test.go, server/deck_lineage_metadata_test.go</files>
  <action>Write `TestLineageProjectionContract` covering wins/losses/draws/no-contest/incomplete, participant counts, missing turns, corrections, unknown context, hidden/private events, submitted vs draft review, 1/2/3 games per side, zero denominators, source access changes, and 10,000 games. Add RDA-1 fixtures for create/rename snapshot label revisions, blank/overlong/forged-owner labels, historical label audit, pinned source revision equal/different/unknown from latest observed source revision, and a source-current-vs-pinned-history notice that never mutates the snapshot. Keep testing-group ownership absent per D-19. Assert game-weighted rates, distributions, expected share by rules/participants, exact excluded counts, uncertainty, Limited evidence, and distinct card evidence states with accessible source IDs. When the projector contract is absent, fail only the named test with `EXPECTED_RED[RI-05-01-T1]: lineage projection contract is not installed`.</action>
  <acceptance_criteria>Every displayed number reconciles to source IDs; no sample gate hides evidence; no not-drawn or ranking/causal claim is inferred from absence/correlation.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-05-01-T1 --require-test TestLineageProjectionContract --require-reason 'lineage projection contract is not installed' -- go test ./pkg/analysis ./pkg/projection ./server/... -run '^TestLineageProjectionContract$' -race -v</automated></verify>
  <done>RDA-1/RDA-2/RDA-3 behavior is reproducible from n=1 through the 10k scale fixture.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install versioned read models and shadow projector</name>
  <files>persistence/migrations/20260820130000_lineage_analytics_v1.up.sql, persistence/migrations/20260820130000_lineage_analytics_v1.down.sql, persistence/migrations_test/20260820130000_lineage_analytics_v1.up.sql, persistence/migrations_test/20260820130000_lineage_analytics_v1.down.sql, pkg/analysis/eligibility.go, pkg/analysis/statistics.go, pkg/projection/lineage_projector.go, pkg/projection/lineage_projector_test.go, server/deck_lineage_metadata.go, server/deck_lineage_metadata_test.go</files>
  <action>Create append-only snapshot label revisions with actor/time/supersedes and current pointer, plus source-observation records that retain pinned source ID/revision/hash and the latest owner-authorized observation separately. Implement rename/audit and source-status services with bounded labels and owner authorization; a changed third-party source produces a notice and never updates snapshot contents. Create versioned lineage/card/inference summaries, projection versions/checkpoints, source watermarks, exact totals, evidence source IDs, and active pointer. Project only canonical participants/snapshots/results/events/inferences/current submitted review revisions. Build shadow, reconcile, switch atomically; corrections/deletion/access changes create new versions. Retain prior versions and switch back by pointer/flag rather than mutating sources. Do not add testing-group owner fields or roles per D-19.</action>
  <acceptance_criteria>Two rebuilds are identical, every total reconciles, pointer switches only after validation, and migration parity passes.</acceptance_criteria>
  <verify><automated>go test ./pkg/analysis ./pkg/projection ./server/... -run '^TestLineageProjectionContract$' -race -v &amp;&amp; go test ./pkg/analysis ./pkg/projection ./server/... -run 'TestLineageProjector|TestLineageShadowSwitch|TestLineageRebuildParity|TestSnapshotLabelRevision|TestSnapshotSourceChangeNotice' -race &amp;&amp; diff persistence/migrations/20260820130000_lineage_analytics_v1.up.sql persistence/migrations_test/20260820130000_lineage_analytics_v1.up.sql</automated></verify>
  <done>Private analytics are rebuildable, auditable, sample-honest, and rollback-safe.</done>
</task>
</tasks>
<verification>`go test ./pkg/analysis ./pkg/projection -race`; two rebuilds; migration parity.</verification>
<success_criteria>RDA-1 through RDA-3 have deterministic data contracts before API/UI work.</success_criteria>

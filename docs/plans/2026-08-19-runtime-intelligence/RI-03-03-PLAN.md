---
phase: RI-03-canonical-runtime-events
plan: RI-03-03
type: execute
wave: 11
depends_on: ["RI-03-02"]
files_modified:
  - persistence/migrations/20260820111000_inference_projection_registry.up.sql
  - persistence/migrations/20260820111000_inference_projection_registry.down.sql
  - persistence/migrations_test/20260820111000_inference_projection_registry.up.sql
  - persistence/migrations_test/20260820111000_inference_projection_registry.down.sql
  - pkg/inference/registry.go
  - pkg/inference/registry_test.go
  - pkg/inference/testdata/expert-holdout-v1.json
  - pkg/projection/projector.go
  - pkg/projection/projector_test.go
  - server/runtime_inferences.go
  - server/runtime_inferences_test.go
requirements: [RGI-4, D-14, D-15, D-16, D-17]
autonomous: true
scope_rationale: "The eleven files are the single authoritative registry-and-impact slice: catalog data, impact typing, both API surfaces, UI consumers, and their focused tests must agree in one change."
must_haves:
  truths:
    - "RGI-4 enrichment is additive/versioned and never alters observed records."
    - "Probabilistic default eligibility derives from stored per-category/version holdout precision at or above 95%."
    - "Explicit recorded source-target may be causal fact; other candidates remain hypotheses."
    - "Every derived or inferred fact references one immutable calculation-definition ID and exact semantic version plus its authorized source-event links."
    - "Implementation leaves inference read/default flags off pending RI-03-04 human approval."
  artifacts:
    - path: "pkg/inference/registry.go"
      provides: "version/evaluation/visibility/provenance registry"
  key_links:
    - from: "inference_evaluations"
      to: "runtime_inferences default_eligible"
      via: "computed exact category/version precision"
    - from: "runtime_inferences.calculation_definition_id/version"
      to: "calculation_definitions"
      via: "immutable exact-version foreign key returned with authorized source links"
---

<objective>Implement RGI-4 additive inference and projection evaluation without enabling default display. Output: expert holdout fixtures, immutable registry, provenance, correction recomputation, and shadow projectors.</objective>
<execution_context>@docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md Story RGI-4 @docs/plans/2026-08-19-runtime-intelligence/RI-03-02-PLAN.md</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Build expert holdout, provenance, and causality fixtures</name>
  <files>pkg/inference/registry_test.go, pkg/inference/testdata/expert-holdout-v1.json, pkg/projection/projector_test.go, server/runtime_inferences_test.go</files>
  <action>Write `TestInferenceCalculationContract` with positive/negative labels per draw, combat, mana production/expenditure, tutoring/searching, source/target, and causal-candidate category. Calculate TP/FP precision for 94.9% and at-least-95% versions; test deterministic visibility, confidence/source/version fields, explicit vs absent source-target, correction invalidation/recompute, unknown outcome, restart, and shadow parity. For every derived/inferred fixture require an immutable calculation-definition ID plus exact semantic version, rule/expression reference, input/output schema versions, source build/commit, calculation timestamp, and ordered source-event IDs; test no fact can persist or serialize with a missing/mismatched definition version. When the registry is absent, fail only the named contract test with `EXPECTED_RED[RI-03-03-T1]: inference and calculation registry is not installed`.</action>
  <acceptance_criteria>Default eligibility is calculated from immutable fixtures, never a hand-set approval boolean.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-03-03-T1 --require-test TestInferenceCalculationContract --require-reason 'inference and calculation registry is not installed' -- go test ./pkg/inference ./pkg/projection ./server/... -run '^TestInferenceCalculationContract$' -race -v</automated></verify>
  <done>The quality, provenance, additivity, and causality contracts are executable.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install registry and additive projectors with flags off</name>
  <files>persistence/migrations/20260820111000_inference_projection_registry.up.sql, persistence/migrations/20260820111000_inference_projection_registry.down.sql, persistence/migrations_test/20260820111000_inference_projection_registry.up.sql, persistence/migrations_test/20260820111000_inference_projection_registry.down.sql, pkg/inference/registry.go, pkg/projection/projector.go, server/runtime_inferences.go, server/runtime_inferences_test.go</files>
  <action>Create immutable calculation definitions, inference versions/evaluations, inferences, ordered source links, projection versions/checkpoints, and shadow read models. A calculation definition stores stable ID/key, exact semantic version, evidence class, human-readable definition, immutable rule/expression or projector reference, input/output schema and vocabulary versions, source commit/build, and status; every runtime derived/inferred record has a composite foreign key to that exact definition/version. Expose the same definition/version and authorized source-link DTO fields from server/runtime_inferences.go. Implement fixture-supported deterministic/probabilistic projectors, idempotent source/definition-version keys, correction invalidation with retained calculation-version history, and shadow reconcile/switch. Store `default_eligible` as derived evidence only. Keep VEDH_RI_INFERENCE_READ false and no default display until the RI-03-04 checkpoint records approval.</action>
  <acceptance_criteria>94.9% remains speculative, qualifying exact versions are only eligible for approval, and observed counts are unchanged.</acceptance_criteria>
  <verify><automated>go test ./pkg/inference ./pkg/projection ./server/... -run '^TestInferenceCalculationContract$' -race -v &amp;&amp; go test ./pkg/inference ./pkg/projection ./server/... -run 'TestInference|TestProjector|TestInferenceImmutability|TestCalculationDefinitionExactVersion|TestInferenceCalculationSourceLink' -race &amp;&amp; diff persistence/migrations/20260820111000_inference_projection_registry.up.sql persistence/migrations_test/20260820111000_inference_projection_registry.up.sql</automated></verify>
  <done>RGI-4 is implemented and auditable but cannot default-display without the next plan's blocking decision.</done>
</task>
</tasks>
<verification>`go test ./pkg/inference ./pkg/projection ./server/... -race`; migration parity; flags asserted false.</verification>
<success_criteria>Additive enrichment exists with quality evidence while product visibility remains unapproved and disabled.</success_criteria>

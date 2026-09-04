---
phase: RI-01-shared-database-identity
plan: RI-01-05
type: execute
wave: 4
depends_on: ["RI-01-02"]
files_modified:
  - persistence/migrations/20260820092000_card_object_identity.up.sql
  - persistence/migrations/20260820092000_card_object_identity.down.sql
  - persistence/migrations_test/20260820092000_card_object_identity.up.sql
  - persistence/migrations_test/20260820092000_card_object_identity.down.sql
  - pkg/cardidentity/identity.go
  - pkg/cardidentity/identity_test.go
  - server/card_identity.go
  - server/card_identity_test.go
  - cmd/card-identity-backfill/main.go
  - docs/runbooks/card-identity-backfill.md
requirements: [RI-CARD-IDENTITY]
autonomous: true
scope_rationale: "The ten files are one schema-to-backfill identity contract: mirrored DDL, a pure exact resolver, the server adapter, and the only resumable migration that applies the resolver. Splitting them would leave an untestable schema or an unversioned backfill classification."
must_haves:
  truths:
    - "Canonical gameplay object, printing/source identifier, game instance, and unresolved identity are distinct."
    - "Exact identifiers resolve deterministically; ambiguous or name-only inputs remain explicit and never fuzzy-match history."
    - "Backfill is resumable/idempotent, preserves original labels, and produces approval-ready quality evidence without approving itself."
  artifacts:
    - path: "pkg/cardidentity/identity.go"
      provides: "game-system-scoped exact resolution contract"
    - path: "cmd/card-identity-backfill/main.go"
      provides: "checkpointed exact-or-unresolved migration"
  key_links:
    - from: "card identity backfill"
      to: "card_objects, printing mappings, unresolved objects, and game instances"
      via: "exact resolver version and source provenance"
---

<objective>Create and backfill the canonical card/object prerequisite used by snapshots, events, reviews, analysis, and jank without embedding approval in automated work. Output: mirrored DDL, pure resolver, fixtures, and reproducible backfill evidence.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/00-RESEARCH.md @docs/plans/2026-08-19-runtime-intelligence/RI-01-02-PLAN.md @pkg/deckimport/</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Define exact, ambiguous, unresolved, and game-instance identity fixtures</name>
  <files>pkg/cardidentity/identity_test.go, server/card_identity_test.go</files>
  <action>Write `TestCardIdentityContract` covering canonical object versus printing IDs, double-faced cards, tokens, copies, custom objects, multiple physical instances, duplicate names, exact provider IDs, unknown game systems, and historical original labels. Assert ambiguous/name-only data stays unresolved and no similarity-based branch exists. When the resolver/schema contract is absent, fail only this named test with `EXPECTED_RED[RI-01-05-T1]: canonical card identity is not installed`.</action>
  <acceptance_criteria>Fixtures enumerate every current identifier shape and the expected exact, ambiguous, unresolved, or instance result.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-01-05-T1 --require-test TestCardIdentityContract --require-reason 'canonical card identity is not installed' -- go test ./pkg/cardidentity ./server/... -run '^TestCardIdentityContract$' -race -v</automated></verify>
  <done>Identity behavior is fully specified before schema/service implementation.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Add canonical object, printing, unresolved, and game-instance schema</name>
  <files>persistence/migrations/20260820092000_card_object_identity.up.sql, persistence/migrations/20260820092000_card_object_identity.down.sql, persistence/migrations_test/20260820092000_card_object_identity.up.sql, persistence/migrations_test/20260820092000_card_object_identity.down.sql, pkg/cardidentity/identity.go, server/card_identity.go</files>
  <action>Create card_objects, card_printing_mappings, unresolved_card_objects, and runtime game-object instances with game-system/provider/normalizer provenance and uniqueness constraints. Implement exact resolution returning resolved, ambiguous, or unresolved; preserve input labels and never infer from name similarity.</action>
  <acceptance_criteria>Schema constraints and services pass every fixture and migration parity check.</acceptance_criteria>
  <verify><automated>go test ./pkg/cardidentity ./server/... -run '^TestCardIdentityContract$' -race -v &amp;&amp; go test ./pkg/cardidentity ./server/... -run 'TestCardIdentity|TestRuntimeMigrationParity' -race &amp;&amp; diff persistence/migrations/20260820092000_card_object_identity.up.sql persistence/migrations_test/20260820092000_card_object_identity.up.sql</automated></verify>
  <done>Later phases can reference a stable card object, printing, physical instance, or explicit unresolved row.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Backfill card identity and produce reconciliation evidence</name>
  <files>cmd/card-identity-backfill/main.go, docs/runbooks/card-identity-backfill.md</files>
  <action>Build dry-run, bounded batch, checkpoint, resume, quarantine, and two-pass backfill with exact/ambiguous/unresolved counts and resolver version. Preserve original labels, avoid raw deck payload logs, validate constraints, and write a machine-readable reconciliation section for RI-01-06. Do not write an approval decision in this automated task.</action>
  <acceptance_criteria>Two passes create no duplicates; every source is classified; ambiguous rows remain explicit; the runbook contains evidence but no implied approval.</acceptance_criteria>
  <verify><automated>go run ./cmd/card-identity-backfill --self-test &amp;&amp; go test ./pkg/cardidentity ./server/... -run 'TestCardIdentityBackfill|TestRuntimeMigrationParity' -race</automated></verify>
  <done>The card backfill is reproducible and ready for an independent human quality decision.</done>
</task>

</tasks>
<verification>Card identity/service suites, two-pass backfill, and production/test migration parity.</verification>
<success_criteria>All later evidence can use exact canonical or explicit unresolved identity; cutover remains disabled until RI-01-06.</success_criteria>

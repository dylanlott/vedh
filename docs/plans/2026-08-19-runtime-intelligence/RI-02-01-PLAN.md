---
phase: RI-02-immutable-game-context
plan: RI-02-01
type: execute
wave: 6
depends_on: ["RI-01-03", "RI-01-04", "RI-01-06"]
files_modified:
  - persistence/migrations/20260820100000_deck_snapshots_lineages.up.sql
  - persistence/migrations/20260820100000_deck_snapshots_lineages.down.sql
  - persistence/migrations_test/20260820100000_deck_snapshots_lineages.up.sql
  - persistence/migrations_test/20260820100000_deck_snapshots_lineages.down.sql
  - persistence/migrations/20260820101000_game_participants_results.up.sql
  - persistence/migrations/20260820101000_game_participants_results.down.sql
  - persistence/migrations_test/20260820101000_game_participants_results.up.sql
  - persistence/migrations_test/20260820101000_game_participants_results.down.sql
  - pkg/deckidentity/snapshot_test.go
  - server/runtime_context_schema_test.go
requirements: [RGI-1, RGI-3]
autonomous: true
scope_rationale: "The ten files form one indivisible persistence slice: mirrored vEDH/verified-jank migrations, typed contracts, stores, and fixtures must land together so the event ledger has one schema and one tested API."
must_haves:
  truths:
    - "RGI-1: every eligible game points to exact immutable deck snapshots and stable-source lineages."
    - "RGI-3: outcomes include win/loss/draw/no_contest/abandoned/incomplete; each accepted revision has canonical ending turn/round, finalization source, and structured declared win-condition semantics."
    - "RGI-3 corrections append a superseding revision and atomically repoint current result while preserving every prior result and source event."
    - "User/tombstone, participant, snapshot, and result foreign keys are immutable and auditable."
  artifacts:
    - path: "persistence/migrations/20260820100000_deck_snapshots_lineages.up.sql"
      provides: "snapshot/lineage schema"
    - path: "persistence/migrations/20260820101000_game_participants_results.up.sql"
      provides: "participant/result revision schema"
  key_links:
    - from: "game_participants.deck_snapshot_id"
      to: "deck_snapshots.id"
      via: "required eligible-game context FK"
---

<objective>Define and enforce the immutable schema for RGI-1 and RGI-3 before live dual-write. Output: Wave-0 property fixtures and mirrored snapshot/lineage/participant/result DDL.</objective>
<execution_context>@docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-01-02-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/RI-01-03-PLAN.md</execution_context>

<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Specify snapshot, lineage, participant, and result invariants</name>
  <files>pkg/deckidentity/snapshot_test.go, server/runtime_context_schema_test.go</files>
  <action>Write `TestRuntimeContextSchemaContract` with property/table fixtures: normalized permutation preserves hash; quantity/leader/unresolved/game-system/normalizer changes it; same owner+stable vEDH/trusted external source groups; names, URLs, hashes, or list similarity never group. Cover audited move/split/merge, update/delete denial, tombstoned participant, multiple winners, and six outcomes. For RGI-3 require ending_turn_index/ending_round_index with explicit unknown state, finalization_source from a versioned bounded vocabulary, source_event_id when event-backed, declared_win_condition_code plus vocabulary_version and adapter-validated structured details rather than required free text, author/system actor, corrects_result_revision_id, and atomic current-revision advancement. When the tables/constraints are absent, fail only the named contract test with `EXPECTED_RED[RI-02-01-T1]: immutable game context schema is not installed`.</action>
  <acceptance_criteria>Tests encode every RGI-1/RGI-3 and D-09 invariant and begin red for missing schema.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-02-01-T1 --require-test TestRuntimeContextSchemaContract --require-reason 'immutable game context schema is not installed' -- go test ./pkg/deckidentity ./server/... -run '^TestRuntimeContextSchemaContract$' -race -v</automated></verify>
  <done>The schema contract is executable, including low-fidelity and correction edge cases.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install immutable context tables and constraints</name>
  <files>persistence/migrations/20260820100000_deck_snapshots_lineages.up.sql, persistence/migrations/20260820100000_deck_snapshots_lineages.down.sql, persistence/migrations_test/20260820100000_deck_snapshots_lineages.up.sql, persistence/migrations_test/20260820100000_deck_snapshots_lineages.down.sql, persistence/migrations/20260820101000_game_participants_results.up.sql, persistence/migrations/20260820101000_game_participants_results.down.sql, persistence/migrations_test/20260820101000_game_participants_results.up.sql, persistence/migrations_test/20260820101000_game_participants_results.down.sql</files>
  <action>Create append-only lineages, snapshots/cards/origins/membership/audit/revision links, participants, result revisions, participant results, and current pointers. Result revisions store ending_turn_index, ending_round_index, bounded finalization_source, optional source_event_id, structured declared_win_condition_code/vocabulary_version/details, actor, corrects_result_revision_id, and revision sequence; constraints validate the game-system vocabulary and prevent a free-text-only win condition. Hash content includes game system, format, leaders, exact/unresolved cards, quantities, roles, and normalizer version—not name/URL. Triggers reject destructive evidence mutation; D-13 tombstoning occurs only through the later audited deletion service.</action>
  <acceptance_criteria>All constraints pass fixtures; production/test migrations match; no destructive evidence down path is used for rollback.</acceptance_criteria>
  <verify><automated>go test ./pkg/deckidentity ./server/... -run '^TestRuntimeContextSchemaContract$' -race -v &amp;&amp; go test ./server/... -run 'TestRuntimeMigrationParity|TestRuntimeContextSchema|TestResultRevisionCanonicalFields' -race &amp;&amp; diff persistence/migrations/20260820101000_game_participants_results.up.sql persistence/migrations_test/20260820101000_game_participants_results.up.sql</automated></verify>
  <done>The database can represent RGI-1/RGI-3 with immutable, corrected, tombstone-safe facts.</done>
</task>
</tasks>
<verification>`go test ./pkg/deckidentity ./server/... -race`; migration parity from empty/current.</verification>
<success_criteria>Snapshots, stable lineages, participants, and complete result semantics exist before live writes.</success_criteria>

---
phase: RI-02-immutable-game-context
plan: RI-02-02
type: execute
wave: 7
depends_on: ["RI-02-01"]
files_modified:
  - pkg/deckidentity/snapshot.go
  - server/deck_snapshots.go
  - server/deck_snapshots_test.go
  - server/deck_lineages.go
  - server/deck_lineages_test.go
  - server/game_participants.go
  - server/game_participants_test.go
  - server/game_results.go
  - server/game_results_test.go
  - server/games.go
  - server/game_finish.go
  - cmd/runtime-context-backfill/main.go
  - docs/runbooks/runtime-context-migration.md
requirements: [D-04, D-05, D-09]
autonomous: true
scope_rationale: "The thirteen files form one context dual-write boundary: pure snapshot/lineage services, participant/result services, the two existing game write seams, and one resumable historical backfill. Separating the live write seams would undermine the atomicity tests this plan exists to enforce."
must_haves:
  truths:
    - "Snapshot capture precedes shuffle and commits atomically with game/participant state."
    - "Lineage assignment uses only same-owner stable source IDs; all manual changes append audit events."
    - "Every finish/correction transaction writes canonical ending turn/round, finalization source, structured declared win condition, participant outcomes, and current revision together."
    - "Backfill invents no participant, deck, result, finalization source, or win-condition fact and is resumable/idempotent."
  artifacts:
    - path: "cmd/runtime-context-backfill/main.go"
      provides: "conservative current-schema migration"
  key_links:
    - from: "server/games.go"
      to: "server/deck_snapshots.go"
      via: "pre-shuffle transaction"
---

<objective>Implement immutable context services, transactional live dual-write, and conservative historical backfill. Output: tested domain services, game integration, reconciliation, and rollback runbook.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-02-01-PLAN.md @server/games.go @server/game_finish.go</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Implement snapshots, stable-source lineages, and audit commands</name>
  <files>pkg/deckidentity/snapshot.go, server/deck_snapshots.go, server/deck_snapshots_test.go, server/deck_lineages.go, server/deck_lineages_test.go</files>
  <action>Implement CaptureDeckSnapshot, AssignSnapshotLineage, MoveSnapshot, SplitLineage, MergeLineages, and LinkDeckRevision. Content-hash dedupe may reuse immutable content but never establishes lineage. Require canonical actor/reason on manual changes and append lineage events. Add bounded capture/assignment metrics.</action>
  <acceptance_criteria>Same owner+stable source is the sole automatic grouping route; all mutation/history tests pass.</acceptance_criteria>
  <verify><automated>go test ./pkg/deckidentity ./server/... -run 'TestSnapshot|TestLineage' -race -v</automated></verify>
  <done>Snapshots and lineage history are deterministic, immutable, and D-09 compliant.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Persist participants and complete result revisions atomically</name>
  <files>server/game_participants.go, server/game_participants_test.go, server/game_results.go, server/game_results_test.go, server/games.go, server/game_finish.go</files>
  <action>Behind default-off VEDH_RI_SNAPSHOT_DUAL_WRITE, capture/reuse snapshot before shuffle and commit game JSON, participant, snapshot, and applicable result revision in one transaction. On finish, write ending_turn_index/ending_round_index, finalization_source, source event, structured declared_win_condition_code/vocabulary_version/details, per-participant outcomes, and current revision atomically. Corrections append a new revision linked to the prior revision and atomically repoint current; they never update prior result rows. Derive loss only from complete participant/rules facts. Return retryable errors on any canonical write failure and leave legacy behavior intact when disabled.</action>
  <acceptance_criteria>Fault tests prove all-or-nothing writes; retries create no duplicates; all six outcomes remain distinct.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestCreateGameSnapshotAtomic|TestJoinGameParticipantAtomic|TestFinishResultAtomic|TestFinishResultCanonicalFields|TestResultRevision|TestResultCorrectionPreservesPriorRevision' -race -v</automated></verify>
  <done>New eligible games receive immutable participant/snapshot/result context without split writes.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 3: Backfill history with explicit fidelity and exclusions</name>
  <files>cmd/runtime-context-backfill/main.go, docs/runbooks/runtime-context-migration.md</files>
  <action>Implement dry-run, bounded batch, checkpoint, quarantine, resume, two-pass reconciliation, deferred-constraint validation, and flag rollback. Classify snapshots exact/partial/unknown; resolve users only inside the exact game participant set; copy only explicit result, ending turn/round, finalization-source, and declared-win-condition facts. Use explicit unknown migration states rather than inferring structured semantics from free text. Preserve all excluded reason codes and never log raw deck/game payloads.</action>
  <acceptance_criteria>Two runs produce no duplicates and every historical game is eligible or explicitly classified/excluded.</acceptance_criteria>
  <verify><automated>go run ./cmd/runtime-context-backfill --self-test &amp;&amp; go test ./server/... -run TestRuntimeContextBackfill -race</automated></verify>
  <done>Historical context is auditable and restartable without manufactured evidence.</done>
</task>
</tasks>
<verification>`go test ./pkg/deckidentity ./server/... -race`; `go run ./cmd/runtime-context-backfill --self-test`.</verification>
<success_criteria>Live and historical context supports activation eligibility and later events/reviews.</success_criteria>

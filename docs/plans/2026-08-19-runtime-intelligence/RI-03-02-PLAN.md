---
phase: RI-03-canonical-runtime-events
plan: RI-03-02
type: execute
wave: 10
depends_on: ["RI-03-01"]
files_modified:
  - server/runtime_events.go
  - server/runtime_events_test.go
  - server/games.go
  - server/boardstates.go
  - server/game_finish.go
  - server/gamelog.go
  - server/gamelog_helpers.go
  - server/gamelog_diff.go
  - server/graphql.go
  - server/metrics.go
  - pkg/telemetry/metrics.go
  - cmd/runtime-event-backfill/main.go
  - docs/runbooks/runtime-event-cutover.md
requirements: [D-05]
autonomous: true
scope_rationale: "The thirteen files are the existing accepted-mutation surface plus its one canonical writer, telemetry seam, backfill, and runbook. They must change as one fault-injected transaction boundary so no command path can retain the prior split-write behavior."
must_haves:
  truths:
    - "Each accepted beta mutation commits live state and one canonical event set together or returns a visible retryable error."
    - "Replay after response loss returns the committed result without duplicate events."
    - "Legacy backfill is one-source-row-to-one-canonical-record, quality-classified, and idempotent."
  artifacts:
    - path: "server/runtime_events.go"
      provides: "WithGameMutation transactional boundary"
  key_links:
    - from: "games/boardstates/game_finish"
      to: "runtime_events"
      via: "one SQL transaction before subscriptions/metrics"
---

<objective>Deliver D-05 exact-once transactional mutation persistence and a conservative legacy event backfill. Output: fault-tested transaction boundary, live wiring, metrics, backfill, and rollback runbook.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-03-01-PLAN.md @server/games.go @server/gamelog.go</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Prove every transactional and retry failure boundary</name>
  <files>server/runtime_events_test.go</files>
  <action>Write `TestGameMutationAtomicityContract` with controllable fault points before event insert, between multi-event inserts, before state update, before commit, and after commit/response loss. Test duplicate/in-flight mutation UUIDs, concurrent sequence allocation, result/correction events, legacy flag-off behavior, and post-commit-only subscriptions/product metrics. When WithGameMutation is absent, fail only this named test with `EXPECTED_RED[RI-03-02-T1]: transactional game mutation boundary is not installed`.</action>
  <acceptance_criteria>Every fault proves all-or-nothing persistence and replay identity with no swallowed canonical error.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-03-02-T1 --require-test TestGameMutationAtomicityContract --require-reason 'transactional game mutation boundary is not installed' -- go test ./server/... -run '^TestGameMutationAtomicityContract$' -race -v</automated></verify>
  <done>The failure matrix demonstrates exact-once behavior before live paths are rewired.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Route accepted mutations through WithGameMutation</name>
  <files>server/runtime_events.go, server/runtime_events_test.go, server/games.go, server/boardstates.go, server/game_finish.go, server/gamelog.go, server/gamelog_helpers.go, server/gamelog_diff.go, server/graphql.go, server/metrics.go, pkg/telemetry/metrics.go</files>
  <action>Implement WithGameMutation(ctx, gameID, mutationID, fn): serialize sequence, validate, update games.payload, append participants/results/corrections/events, commit, then publish subscriptions/metrics. Route create/join/update/board/phase/priority/win/finalize behind default-off VEDH_RI_CANONICAL_EVENT_WRITE. Require mutation UUIDs for beta eligibility; legacy generated IDs are quality-marked/excluded where exact-once cannot be proven. Keep gamelog as post-commit compatibility adapter and never swallow canonical failures.</action>
  <acceptance_criteria>All live fault/retry tests pass; flag rollback returns to legacy path without deleting canonical rows; metrics have bounded labels only.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestGameMutationAtomicityContract$' -race -v &amp;&amp; go test ./server/... -run 'TestMutationFault|TestMutationRetry|TestCanonicalEventDualWrite|TestCanonicalEventMetrics' -race -v</automated></verify>
  <done>Accepted eligible mutations have atomic live state and canonical evidence or a visible retryable failure.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 3: Backfill and reconcile legacy gamelog</name>
  <files>cmd/runtime-event-backfill/main.go, docs/runbooks/runtime-event-cutover.md</files>
  <action>Build dry-run/batch/checkpoint/resume/quarantine backfill. Preserve each legacy payload privately and exactly once by legacy ID; classify diff rows as server_derived_client_state, resolve actor only inside the game participant set, preserve unknown versions, and invent no absent events. Reconcile counts/checksums, rerun twice, and document write-flag rollback.</action>
  <acceptance_criteria>Two passes create no duplicates; every row maps or has a quarantine reason; reconciliation is machine-readable.</acceptance_criteria>
  <verify><automated>go run ./cmd/runtime-event-backfill --self-test &amp;&amp; go test ./server/... -run TestRuntimeEventBackfill -race</automated></verify>
  <done>Legacy evidence is conservatively represented and ready for inference/timeline work.</done>
</task>
</tasks>
<verification>`go test ./server/... -run 'Mutation|RuntimeEventBackfill' -race`; backfill self-test.</verification>
<success_criteria>D-05 exact-once ownership is entirely in RI-03, independent from RI-01 identity prerequisites.</success_criteria>

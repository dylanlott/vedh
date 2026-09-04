---
phase: RI-03-canonical-runtime-events
plan: RI-03-01
type: execute
wave: 9
depends_on: ["RI-02-03"]
files_modified:
  - persistence/migrations/20260820110000_runtime_events.up.sql
  - persistence/migrations/20260820110000_runtime_events.down.sql
  - persistence/migrations_test/20260820110000_runtime_events.up.sql
  - persistence/migrations_test/20260820110000_runtime_events.down.sql
  - pkg/runtimeevent/envelope.go
  - pkg/runtimeevent/envelope_test.go
  - pkg/runtimeevent/testdata/magic-v1.json
  - server/runtime_events_test.go
requirements: [RGI-2, RGI-7]
autonomous: true
must_haves:
  truths:
    - "RGI-2 events are typed, ordered, append-only, idempotent, correction-capable, distinguish observed/client/derived origins, and version shuffle/randomization explicitly in Magic-v1."
    - "RGI-7 shared envelope has no Magic-only required fields; beta ships only a versioned Magic adapter."
  artifacts:
    - path: "pkg/runtimeevent/envelope.go"
      provides: "cross-game envelope plus Magic-v1 adapter"
  key_links:
    - from: "runtime_events"
      to: "game_participants/card identity"
      via: "typed actor/subject/object references"
---

<objective>Define RGI-2 and RGI-7 as executable event schema and contract fixtures. Output: golden corpus, mirrored append-only DDL, and Magic-only typed adapter.</objective>
<execution_context>@docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-02-03-PLAN.md @server/gamelog_diff.go</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Build golden event and adapter fixtures</name>
  <files>pkg/runtimeevent/envelope_test.go, pkg/runtimeevent/testdata/magic-v1.json, server/runtime_events_test.go</files>
  <action>Write `TestMagicV1EventContract` covering create/join/draw/move/tap/stack/turn/priority/resource, shuffle, randomization, win/cancel/finish/recovery snapshot, multiple copies, out-of-order client time, unknown schema version, correction, incomplete/no-contest/abandoned, explicit source-target, and 10,000 ordered events. Add named `TestMagicV1ShuffleRandomizationVocabulary` and `TestShuffleRandomizationPersistenceCorrectionIdempotency` cases proving the versioned Magic vocabulary records the observed shuffle/randomization action and typed scope/outcome metadata actually supplied by the accepted mutation, never reconstructs or overwrites unobserved order, persists once under mutation retry, and corrects by append-only reference. Assert Magic zones/life/stack/commander remain adapter fields and unsupported shared concepts return not_applicable. When the vocabulary/persistence contract is absent, fail only `TestMagicV1EventContract` with `EXPECTED_RED[RI-03-01-T1]: Magic-v1 runtime event contract is not installed`.</action>
  <acceptance_criteria>Every current event type and RGI-2 envelope field has a golden case; no second adapter is selected per D-22.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-03-01-T1 --require-test TestMagicV1EventContract --require-reason 'Magic-v1 runtime event contract is not installed' -- go test ./pkg/runtimeevent ./server/... -run '^TestMagicV1EventContract$' -race -v</automated></verify>
  <done>RGI-2/RGI-7 behavior is locked by versioned fixtures before persistence code.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install immutable events and Magic-v1 envelope</name>
  <files>persistence/migrations/20260820110000_runtime_events.up.sql, persistence/migrations/20260820110000_runtime_events.down.sql, persistence/migrations_test/20260820110000_runtime_events.up.sql, persistence/migrations_test/20260820110000_runtime_events.down.sql, pkg/runtimeevent/envelope.go, pkg/runtimeevent/envelope_test.go</files>
  <action>Create runtime_events with server sequence, client mutation UUID/ordinal, schema/adapter versions, actor/subject, origin, visibility, temporal context, typed payload, correction target, and legacy ID. Define versioned Magic-v1 shuffle and randomization event types/payload validators, preserving only the observed action/scope/outcome metadata supplied by the canonical mutation and never treating a derived ordering guess as observed truth. Enforce unique game sequence and mutation ordinal for these events, same-game correction, timeline index, and update/delete denial. Preserve unknown versions privately without interpreting them.</action>
  <acceptance_criteria>Order, retry identity, correction locality, append-only behavior, and adapter boundaries are database/test enforced.</acceptance_criteria>
  <verify><automated>go test ./pkg/runtimeevent ./server/... -run '^TestMagicV1EventContract$' -race -v &amp;&amp; go test ./pkg/runtimeevent ./server/... -run 'TestRuntimeEventEnvelope|TestRuntimeEventMigration|TestRuntimeEventImmutability|TestMagicV1ShuffleRandomizationVocabulary|TestShuffleRandomizationPersistenceCorrectionIdempotency' -race &amp;&amp; diff persistence/migrations/20260820110000_runtime_events.up.sql persistence/migrations_test/20260820110000_runtime_events.up.sql</automated></verify>
  <done>The canonical event store and Magic adapter satisfy RGI-2/RGI-7 without flattening future games.</done>
</task>
</tasks>
<verification>`go test ./pkg/runtimeevent ./server/... -run 'TestMagicV1EventContract|TestMagicV1ShuffleRandomizationVocabulary|TestShuffleRandomizationPersistenceCorrectionIdempotency|TestRuntimeEvent10k' -race`; migration parity and 10k fixture.</verification>
<success_criteria>Canonical event contracts exist before any mutation path switches writes.</success_criteria>

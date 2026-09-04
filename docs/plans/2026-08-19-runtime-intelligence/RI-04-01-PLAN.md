---
phase: RI-04-private-debrief-revisions
plan: RI-04-01
type: execute
wave: 13
depends_on: ["RI-03-04"]
files_modified:
  - persistence/migrations/20260820120000_game_reviews.up.sql
  - persistence/migrations/20260820120000_game_reviews.down.sql
  - persistence/migrations_test/20260820120000_game_reviews.up.sql
  - persistence/migrations_test/20260820120000_game_reviews.down.sql
  - server/game_reviews_test.go
  - server/review_revision_test.go
requirements: [PGR-1, PGR-2, PGR-4]
autonomous: true
must_haves:
  truths:
    - "PGR-1 debrief facts come from canonical records and opening creates no submitted review."
    - "PGR-2 allows up to three contributor assessments and one underperformer from the exact snapshot."
    - "PGR-2 assessments allow an optional explanation of at most 500 characters and preserve recorded-in-game versus selected-from-deck provenance."
    - "PGR-4 allows zero-to-five tags from the exact versioned eleven-value vocabulary, with no free-form tags."
    - "Drafts are private/mutable; submissions are append-only immutable revisions."
  artifacts:
    - path: "persistence/migrations/20260820120000_game_reviews.up.sql"
      provides: "private drafts and immutable review revisions"
  key_links:
    - from: "game_reviews"
      to: "canonical game/author/snapshot/result"
      via: "participant-authorized foreign keys"
---

<objective>Define the private debrief persistence contract for PGR-1, PGR-2, and PGR-4 after the typed timeline is complete. Output: Wave-0 fixtures and mirrored draft/revision DDL.</objective>
<execution_context>@docs/product/2026-08-17-expert-post-game-debrief-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-03-04-PLAN.md</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Specify review limits, privacy, revisions, and canonical facts</name>
  <files>server/game_reviews_test.go, server/review_revision_test.go</files>
  <action>Write `TestGameReviewSchemaContract` for finished-only foreign-key availability, canonical game/author/snapshot references without copied facts, zero-to-three contributors, one underperformer, duplicate prevention, and snapshot-only card IDs. At the persistence boundary test optional assessment explanations at 0/500/501 characters, the two provenance values, the exact eleven lesson tags under vocabulary version 1 with zero-to-five bounds, mutable draft versus immutable revision rows, draft exclusion, and the five engagement states. When storage is absent, fail only this named test with `EXPECTED_RED[RI-04-01-T1]: private review schema is not installed`. API/UI authorization and interaction fixtures land in RI-04-02/03 with their implementation rather than remaining unrelated red failures here.</action>
  <acceptance_criteria>Every PGR-1/PGR-2/PGR-4 acceptance clause is represented and tests begin red on missing storage.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-04-01-T1 --require-test TestGameReviewSchemaContract --require-reason 'private review schema is not installed' -- go test ./server/... -run '^TestGameReviewSchemaContract$' -race -v</automated></verify>
  <done>The core review contract is executable before DDL or API implementation.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install private draft and immutable revision schema</name>
  <files>persistence/migrations/20260820120000_game_reviews.up.sql, persistence/migrations/20260820120000_game_reviews.down.sql, persistence/migrations_test/20260820120000_game_reviews.up.sql, persistence/migrations_test/20260820120000_game_reviews.down.sql</files>
  <action>Create game_reviews, game_review_drafts, immutable revisions, card assessments, versioned lesson vocabulary/selections, next-action fields, timeline-note links, private search document, engagement state, dismissal/reminder state, and deletion request records. Card assessments store snapshot card, type, optional explanation with a 500-character database/domain limit, and recorded-runtime source/provenance used for the two UI labels. Seed and constrain the exact eleven PGR-4 values under vocabulary version 1; reject free-form values. Constrain engagement state to viewed/drafted/submitted/dismissed/not_seen. Enforce per-author/game identity, contributor/underperformer/tag bounds, revision immutability, draft exclusion, and author-only search. Do not copy editable factual fields.</action>
  <acceptance_criteria>Constraints enforce limits/privacy/history and migration parity passes from empty/current fixtures.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestGameReviewSchemaContract$' -race -v &amp;&amp; go test ./server/... -run 'TestReviewMigration|TestReviewCardAssessmentExplanationLimit|TestReviewLessonTagVocabularyV1|TestReviewEngagementStates|TestReviewRevision' -race &amp;&amp; diff persistence/migrations/20260820120000_game_reviews.up.sql persistence/migrations_test/20260820120000_game_reviews.up.sql</automated></verify>
  <done>The database supports private drafts and immutable submissions without duplicating canonical facts.</done>
</task>
</tasks>
<verification>`go test ./server/... -run Review -race`; migration parity.</verification>
<success_criteria>PGR-1/PGR-2/PGR-4 storage is complete and RI-04 no longer conflicts or runs parallel with RI-03 generated files.</success_criteria>

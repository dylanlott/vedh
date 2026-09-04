---
phase: RI-04-private-debrief-revisions
plan: RI-04-03
type: execute
wave: 15
depends_on: ["RI-04-02"]
files_modified:
  - app/src/components/reviews/GameDebrief.vue
  - app/src/components/reviews/ReviewHistory.vue
  - app/src/components/reviews/SnapshotCardPicker.vue
  - app/src/views/GameAnalysisView.vue
  - app/src/graphql/queries.ts
  - app/src/graphql/mutations.ts
  - app/src/types/generated.ts
  - app/__tests__/GameDebrief.spec.ts
  - app/e2e/game-debrief.spec.ts
  - server/review_deletion.go
  - server/review_deletion_test.go
  - server/product_events.go
  - server/product_events_test.go
  - docs/runbooks/game-review-deletion.md
requirements: [PGR-7, D-11, D-12]
autonomous: false
scope_rationale: "The fourteen files are the single completed-game debrief slice: one Vue form/history/picker route, its generated client operations and browser tests, the exact five-state engagement allowlist, and the domain deletion hook required before cohort approval. Splitting the UI files would create unsafe overlap on GameAnalysisView and generated operations."
must_haves:
  truths:
    - "PGR-7 debrief is one-action dismissible, never blocks navigation/rematch/invite/claim, and nags at most once in 24 hours."
    - "PGR-7 telemetry stores only viewed, drafted, submitted, dismissed, or not_seen and never review content."
    - "The compact default shows assessments/notes/next action; controlled tags remain behind Add detail."
    - "Review deletion exposes an idempotent domain hook consumed by RI-08's expiry orchestrator."
  artifacts:
    - path: "app/src/components/reviews/GameDebrief.vue"
      provides: "optional accessible debrief workflow"
  key_links:
    - from: "review deletion outbox"
      to: "RI-08 ExecuteExpiredAccountDeletion"
      via: "idempotent domain hook contract"
---

<objective>Ship the optional debrief UX, PGR-7 skip contract, and private-review deletion hook, then block cohort enablement on human privacy/usability verification.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-04-02-PLAN.md @app/src/views/GameAnalysisView.vue</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Build accessible compact debrief, history, search, and skip flow</name>
  <files>app/src/components/reviews/GameDebrief.vue, app/src/components/reviews/ReviewHistory.vue, app/src/components/reviews/SnapshotCardPicker.vue, app/src/views/GameAnalysisView.vue, app/src/graphql/queries.ts, app/src/graphql/mutations.ts, app/src/types/generated.ts, app/__tests__/GameDebrief.spec.ts, app/e2e/game-debrief.spec.ts, server/product_events.go, server/product_events_test.go</files>
  <action>Render canonical facts, up to three contributors/one underperformer, optional 500-character explanation counters/errors, and the exact recorded in this game versus selected from deck labels. Render PGR-3 notes as escaped plain text with maxlength 10000 and server error parity; never interpret markup. Keep the exact eleven version-1 lesson tags behind Add detail with zero-to-five validation and no free-form input. Render next action, autosave, search/history, and immutable edit submission. Implement one-action dismissal, nonblocking navigation/rematch/invite/claim, one reminder maximum in 24h, and persistent access from Game Analysis. Emit/store only the exact engagement states viewed/drafted/submitted/dismissed/not_seen through a server allowlist. Cover keyboard/screen-reader/mobile/save-failure states.</action>
  <acceptance_criteria>E2E proves notes-only/cards-only/detailed/dismiss/resume/edit/search/open-resolved/deep-link paths and optionality.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestReviewEngagementStates|TestReviewProductEventContentRejection' -race &amp;&amp; npm --prefix app test -- GameDebrief.spec.ts &amp;&amp; npm --prefix app run type-check &amp;&amp; npm --prefix app run test:e2e -- game-debrief.spec.ts</automated></verify>
  <done>The user can complete or skip the private debrief without friction or factual re-entry.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Implement idempotent private-review deletion hook</name>
  <files>server/review_deletion.go, server/review_deletion_test.go, docs/runbooks/game-review-deletion.md</files>
  <action>Implement DeletePrivateReviewsForAccount(ctx, tx, canonicalUserID, deletionID) for RI-08 orchestration. Delete drafts, revisions, assessments, tags, note links/search documents, and next-action private content; emit idempotent publication-revocation work; retain canonical game/snapshot/event facts. Test retry, partial failure rollback, recovery-before-expiry no-op, restored-data re-deletion, and no content in logs/metrics.</action>
  <acceptance_criteria>The hook is transactional/retryable, safe on replay, and leaves only required canonical multiplayer evidence.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestReviewDeletionHook|TestReviewDeletionReplay|TestReviewDeletionRestore' -race -v</automated></verify>
  <done>RI-08 can compose private-review deletion without reimplementing review-domain rules.</done>
</task>
<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 3: Approve review privacy and optional interaction cohort</name>
  <files>docs/runbooks/game-review-deletion.md</files>
  <action>Review notes/cards/detailed/dismiss/resume/edit/history/search/mobile/keyboard/failure flows, including 500/501 assessment explanation, 10,000/10,001 plain-text notes, inert markup, exact eleven-tag vocabulary, the two provenance labels, and all five engagement states. Inspect logs/metrics/product events for content; rehearse request/recovery/expiry hook on production-shaped data. Record APPROVED or REJECTED with evidence. Rejection/absence keeps VEDH_RI_REVIEW off and leaves this plan incomplete.</action>
  <acceptance_criteria>Human confirms optionality/accessibility and finds no private content leakage; deletion hook evidence is attached.</acceptance_criteria>
  <verify><automated>grep -q 'Review cohort decision: APPROVED' docs/runbooks/game-review-deletion.md</automated></verify>
  <done>Review cohort enablement is APPROVED; rejection/absence halts RI-05 and is not dependency completion.</done>
</task>
</tasks>
<verification>Vue unit/type/E2E plus `go test ./server/... -run ReviewDeletion -race`.</verification>
<success_criteria>Private debriefs are optional, immutable-revisioned, searchable only by owner, and deletion-ready.</success_criteria>

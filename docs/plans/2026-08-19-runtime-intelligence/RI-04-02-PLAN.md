---
phase: RI-04-private-debrief-revisions
plan: RI-04-02
type: execute
wave: 14
depends_on: ["RI-04-01"]
files_modified:
  - server/game_reviews.go
  - server/game_reviews_test.go
  - server/review_search.go
  - server/review_search_test.go
  - server/review_next_actions.go
  - server/review_next_actions_test.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/graphql.go
requirements: [PGR-3, PGR-5]
autonomous: true
scope_rationale: "The eleven files are one generated review API: draft/revision service, private note search/linking, next-action lifecycle, and the shared gqlgen surface. All three services use one author/game transaction and one generated schema, so separating them would duplicate generated-file ownership and authorization fixtures."
must_haves:
  truths:
    - "PGR-3 notes are plain text, at most 10,000 characters, private/revisioned, link to authorized turns/events, and searchable only by their owner."
    - "PGR-5 next actions are typed hypotheses, can link to resulting snapshots, and filter as open or resolved."
    - "Submit advances the immutable current revision atomically; autosave does not."
  artifacts:
    - path: "server/review_search.go"
      provides: "owner-only current-note search"
    - path: "server/review_next_actions.go"
      provides: "typed next-action resolution/filter contract"
  key_links:
    - from: "review note links"
      to: "runtimeTimeline event IDs"
      via: "same-game participant-authorized validation"
---

<objective>Implement the missing executable clauses for PGR-3 notes/search and PGR-5 next-action resolution. Output: review services, typed GraphQL, negative authorization, and generated models.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-04-01-PLAN.md @server/game_analysis.go @server/schema.graphql</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Implement draft, submit, history, and canonical card suggestions</name>
  <files>server/game_reviews.go, server/game_reviews_test.go</files>
  <action>Implement fetch/create-on-write, autosave draft, submit immutable revision/current pointer, return-to-draft, history, and deterministic snapshot card suggestions (recorded authorized activity first, then remaining snapshot cards). Choose PGR-3 plain-text storage/rendering: normalize line endings, reject invalid encoding, enforce 10,000 Unicode characters in domain and database boundaries, and return text for escaped rendering rather than interpreting markup. Validate PGR-2 500-character explanations and the exact versioned PGR-4 vocabulary. Enforce all limits and participant-author identity in transactions. Product metrics carry only viewed/drafted/submitted/dismissed/not_seen plus bounded outcome/latency, never review content.</action>
  <acceptance_criteria>Repeated submits append revisions, drafts remain excluded, facts cannot be overridden, and negative authorization passes.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestGameReviewService|TestReviewPlainTextNoteLimit|TestReviewCardAssessmentExplanationLimit|TestReviewLessonTagVocabularyV1|TestReviewRevision|TestReviewAuthorizationNegative' -race -v</automated></verify>
  <done>The core private debrief service is transactional, revisioned, and content-safe.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Add note-to-turn/event links and private-note search</name>
  <files>server/review_search.go, server/review_search_test.go, server/game_reviews.go, server/game_reviews_test.go</files>
  <action>Validate each note link against the same game and viewer-authorized timeline turn/event; store stable event IDs plus display snapshot. Index only the author's current submitted plain-text note revision, remove prior/draft content on resubmit/delete, and implement bounded owner-only search with escaped text snippets. Prove 10,001 characters fail, markup remains inert text, and other participants, moderators, jank_app, logs, Prometheus, and product_events cannot retrieve note text.</action>
  <acceptance_criteria>Valid links deep-link correctly; forged/cross-game/hidden links fail; owner search returns only current accessible private notes.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestReviewNoteLinks|TestReviewPlainTextNoteLimit|TestPrivateReviewSearch|TestPrivateReviewSearchAuthorization' -race -v</automated></verify>
  <done>PGR-3 note deep links and private-note search are fully executable and privacy-tested.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 3: Add open/resolved next-action filters and GraphQL contract</name>
  <files>server/review_next_actions.go, server/review_next_actions_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/graphql.go</files>
  <action>Implement six typed next actions, canonical card or explicit unresolved candidate, OPEN state until linked to an immutable successor snapshot, RESOLVED with link/time/actor, and owner query filters for OPEN, RESOLVED, or ALL. Add review draft/submit/history/suggestions/search/note-link/next-action queries and mutations. Bound search/filter/page inputs and preserve unknown separately.</action>
  <acceptance_criteria>Open/resolved filters reconcile exactly; linking validates owner/lineage/snapshot; generated API covers all review flows.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestReviewNextAction|TestReviewGraphQL|TestPrivateReviewSearch' -race -v</automated></verify>
  <done>PGR-5 supports a complete hypothesis-to-revision lifecycle and exact open/resolved filtering.</done>
</task>
</tasks>
<verification>`make generate`; `go test ./server/... -run 'Review|PrivateReviewSearch' -race`.</verification>
<success_criteria>PGR-3 and PGR-5 missing clauses are implemented in their primary phase.</success_criteria>

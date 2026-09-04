---
phase: RI-06-privacy-safe-publication
plan: RI-06-02
type: execute
wave: 20
depends_on: ["RI-06-01"]
files_modified:
  - server/evidence_publications.go
  - server/evidence_publications_test.go
  - server/review_participant_sharing.go
  - server/review_participant_sharing_test.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/graphql.go
requirements: [PGR-6, D-12, D-16]
autonomous: true
scope_rationale: "The nine files are one generated sharing contract: participant and public previews intentionally share authorization, immutable version pinning, and gqlgen artifacts while producing distinct recipient/public payloads."
must_haves:
  truths:
    - "PGR-6 participant sharing and public publication each require an exact-content preview of fields, recipients/scope, sources, calculation/review/inference versions, and hash."
    - "Participant sharing uses explicit game-participant scope and no additional mandatory redaction workflow."
    - "Source edits never silently change a participant share or public publication; explicit republish creates a new immutable version."
  artifacts:
    - path: "server/review_participant_sharing.go"
      provides: "exact-preview participant share lifecycle"
    - path: "server/evidence_publications.go"
      provides: "privacy-allowlisted public publication lifecycle"
  key_links:
    - from: "preview hash"
      to: "stored participant/public version"
      via: "fresh reauthorization and byte-identical canonical payload"
---

<objective>Implement PGR-6 participant/public preview and immutable sharing APIs independently from database grants and deletion. Output: generated typed APIs, authorization negatives, preview/save parity, version pinning, republish, and revoke.</objective>
<execution_context>@docs/product/2026-08-17-expert-post-game-debrief-prd.md Story PGR-6 @docs/plans/2026-08-19-runtime-intelligence/RI-06-01-PLAN.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement exact-preview participant sharing</name>
  <files>server/review_participant_sharing.go, server/review_participant_sharing_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go</files>
  <action>Add previewParticipantReviewShare, share, republish, and revoke operations. Preview returns the exact visible note/card assessments/game facts, canonical participant recipients, source review revision, source result/snapshot versions, redactions, and canonical hash. Save reauthorizes source and recipients and requires a fresh matching hash. Sharing permits only canonical game participants and adds no mandatory redaction workflow. Source edits remain pinned until explicit preview plus republish.</action>
  <acceptance_criteria>Stored participant share bytes/hash equal preview; forged/nonparticipant recipients and stale/source-changed hashes fail; private defaults remain private.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestParticipantReviewShare|TestParticipantReviewSharePreviewParity|TestParticipantReviewShareAuthorization' -race -v</automated></verify>
  <done>PGR-6 participant sharing is exact-previewed, recipient-scoped, pinned, revocable, and independent of public publication.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement public publication, republish, and revocation APIs</name>
  <files>server/evidence_publications.go, server/evidence_publications_test.go, server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/graphql.go</files>
  <action>Add preview/publish/republish/revoke operations for game excerpts, debrief excerpts, and analysis slices. Preview returns exact allowlisted fields, publication-scoped pseudonyms, sample/uncertainty, redactions, stable source IDs, review/calculation/source/inference versions, and selected probabilistic evidence. Save reauthorizes every source and requires a fresh matching hash. n=1 remains valid; edits never silently update; moderators cannot publish private reviews.</action>
  <acceptance_criteria>Preview/save parity, source/version pinning, explicit probabilistic selection, n=1, forged-source negatives, republish, and revoke all pass.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestEvidencePublicationGraphQL|TestPublicationPreviewParity|TestPublicationAuthorizationNegative|TestPublicationRepublish' -race -v</automated></verify>
  <done>Public evidence is explicit, source-pinned, privacy-transformed, immutable, and revocable.</done>
</task>

</tasks>
<verification>Generation plus participant/public preview, authorization, pinning, republish, and revoke suites.</verification>
<success_criteria>PGR-6 is fully implemented before grants/UI; participant sharing and public publication remain visibly distinct.</success_criteria>

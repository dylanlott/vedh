---
phase: RI-06-privacy-safe-publication
plan: RI-06-05
type: execute
wave: 22
depends_on: ["RI-06-03", "RI-06-04"]
files_modified:
  - app/src/components/publication/PublicationPreview.vue
  - app/src/components/publication/PublicationStatus.vue
  - app/src/views/GameAnalysisView.vue
  - app/src/views/DeckAnalysisView.vue
  - app/src/graphql/queries.ts
  - app/src/graphql/mutations.ts
  - app/src/types/generated.ts
  - app/__tests__/PublicationPreview.spec.ts
  - app/e2e/evidence-publication.spec.ts
  - docs/runbooks/evidence-publication-privacy.md
requirements: [RI-PUBLICATION-CUTOVER]
autonomous: false
scope_rationale: "The ten files are one exact-preview UI across the two existing analysis entry points, shared generated operations/types, one component/E2E fixture, and the privacy cutover record."
must_haves:
  truths:
    - "Users see exact participant/public payloads before confirmation, including fields, recipients/pseudonyms, sample, uncertainty, redactions, and pinned versions."
    - "Revoked/deleted/cache-stale reads show stable unavailable tombstones."
    - "Publication flags remain off until adversarial human privacy approval."
  artifacts:
    - path: "app/src/components/publication/PublicationPreview.vue"
      provides: "exact-content participant/public confirmation UI"
  key_links:
    - from: "preview UI"
      to: "share/publish/republish mutation"
      via: "fresh preview hash"
---

<objective>Ship exact-content sharing/publication UX and block production cutover on adversarial privacy approval.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-06-02-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/RI-06-03-PLAN.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Build participant/public preview, status, republish, and revoke UX</name>
  <files>app/src/components/publication/PublicationPreview.vue, app/src/components/publication/PublicationStatus.vue, app/src/views/GameAnalysisView.vue, app/src/views/DeckAnalysisView.vue, app/src/graphql/queries.ts, app/src/graphql/mutations.ts, app/src/types/generated.ts, app/__tests__/PublicationPreview.spec.ts, app/e2e/evidence-publication.spec.ts</files>
  <action>Render exact included/excluded fields, participant recipients or public pseudonyms, sample/uncertainty, redactions, filters/calculation/source/review/inference versions, selected probabilistic interpretation, and preview hash. Require separate confirmation for participant share, public publish, and republish. Cover n=1, source edit, stale hash, revoke, deleted source, stale cache, unsafe markup, and accessibility.</action>
  <acceptance_criteria>UI hash-matches stored content, pins all source/calculation versions, and never implies participant sharing equals public publication.</acceptance_criteria>
  <verify><automated>npm --prefix app test -- PublicationPreview.spec.ts &amp;&amp; npm --prefix app run type-check &amp;&amp; npm --prefix app run test:e2e -- evidence-publication.spec.ts</automated></verify>
  <done>Users deliberately confirm the exact recipient-scoped or public payload before visibility changes.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Approve adversarial privacy sample and publication cutover</name>
  <files>docs/runbooks/evidence-publication-privacy.md</files>
  <action>Compare generated participant/public payloads to allowlists; inspect hidden cards, identities, free text, source links, pseudonym linkability, n/uncertainty, selected inference, safe rendering, roles, caches, revoke/delete, and rollback. Record APPROVED or REJECTED with evidence. Rejection/absence keeps VEDH_RI_PUBLICATION_WRITE and VEDH_RI_PUBLICATION_READ off and leaves this plan incomplete.</action>
  <acceptance_criteria>Approval requires zero privacy-critical leaks and successful role/cache/revocation/deletion/rollback drills.</acceptance_criteria>
  <verify><automated>grep -q 'Publication cutover decision: APPROVED' docs/runbooks/evidence-publication-privacy.md</automated></verify>
  <done>Publication cutover is APPROVED; rejection/absence halts RI-07 and is not dependency completion.</done>
</task>

</tasks>
<verification>Vue unit/type/E2E, publication/evidence_api server/role suites, and APPROVED privacy decision.</verification>
<success_criteria>Privacy allowlists, immutable evidence, n=1, and exact participant sharing are production-gated.</success_criteria>

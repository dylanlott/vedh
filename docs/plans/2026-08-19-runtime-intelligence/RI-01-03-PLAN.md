---
phase: RI-01-shared-database-identity
plan: RI-01-03
type: execute
wave: 4
depends_on: ["RI-01-02"]
files_modified:
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/generated.go
  - server/models_gen.go
  - server/graphql.go
  - server/identity_api_test.go
  - docs/runbooks/runtime-identity-cutover.md
requirements: [RI-IDENTITY-CUTOVER]
autonomous: false
must_haves:
  truths:
    - "Generated GraphQL mappings return the same canonical UUID and deletion lifecycle used by domain authorization."
    - "The browser-to-vEDH-to-jank session boundary is concrete, least privilege, and explicitly approved before identity reads switch."
  artifacts:
    - path: "docs/runbooks/runtime-identity-cutover.md"
      provides: "approved session, CSRF/origin, logout, rotation, replay, and rollback contract"
  key_links:
    - from: "GraphQL principal/deletion operations"
      to: "server/identity.go"
      via: "generated typed mappings and canonical authorization"
---

<objective>Expose canonical identity through generated API contracts and make production identity/session cutover a real blocking decision. Output: generated schema/resolvers, negative tests, and an approved deployment-specific session contract.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-01-02-PLAN.md @docs/contracts/runtime-intelligence-baseline.md @server/schema.graphql</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Generate canonical-principal and deletion-lifecycle API mappings</name>
  <files>server/schema.graphql, server/schema.resolvers.go, server/generated.go, server/models_gen.go, server/graphql.go, server/identity_api_test.go</files>
  <action>Add typed canonical principal, deletion request, and deletion recovery operations backed by server/identity.go. Ensure every user/guest response returns canonical UUID, never a compatibility username identity. Cover unrelated subject, guest mismatch, expired/deleted principal, duplicate retry, and field-level credential/secret absence. Regenerate with make generate; never hand-edit generated files.</action>
  <acceptance_criteria>Generated mappings and domain authorization agree on one UUID; no response exposes credentials, guest secrets, reset material, or duplicate identity state.</acceptance_criteria>
  <verify><automated>make generate &amp;&amp; go test ./server/... -run 'TestCanonicalIdentityAPI|TestIdentityAPIAuthorizationNegative' -race -v</automated></verify>
  <done>The public vEDH contract exposes the canonical identity and deletion request/recovery lifecycle safely.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 2: Approve canonical UUID and cross-origin session cutover</name>
  <files>docs/runbooks/runtime-identity-cutover.md</files>
  <action>Using RI-00's recorded domains and proxy topology, select and document either a vEDH-owned shared-domain HttpOnly session or a vEDH-issued short-lived asymmetric/introspected proof. Review CSRF, allowed origins, replay, expiry, logout, rotation, key/issuer/audience ownership, UUID exceptions, generated API tests, and rollback. Record APPROVED or REJECTED with evidence. Shared symmetric minting secrets and duplicate account/link state are forbidden. Rejection/absence keeps VEDH_RI_NATIVE_UUID_READ off and leaves this plan incomplete.</action>
  <acceptance_criteria>The deployment-specific contract is complete, proves one vEDH identity authority, and explicitly approves or blocks the native UUID read switch.</acceptance_criteria>
  <verify><automated>grep -q 'Session boundary decision: APPROVED' docs/runbooks/runtime-identity-cutover.md &amp;&amp; grep -q 'Canonical UUID cutover decision: APPROVED' docs/runbooks/runtime-identity-cutover.md &amp;&amp; make generate &amp;&amp; go test ./server/... -run TestCanonicalIdentityAPI -race</automated></verify>
  <done>Both session and canonical UUID cutovers are APPROVED; rejection/absence halts dependent work and is not dependency completion.</done>
</task>

</tasks>
<verification>Generation, identity API and authorization suites, plus an APPROVED deployment-specific session record.</verification>
<success_criteria>The canonical UUID is typed and approved for production use without creating a second identity authority.</success_criteria>

---
phase: RI-01-shared-database-identity
plan: RI-01-04
type: execute
wave: 4
depends_on: ["RI-01-02"]
files_modified:
  - docs/contracts/jank-user-fk-grant.sql
  - docs/contracts/jank-user-fk-grant.md
  - server/runtime_role_test.go
requirements: [D-07]
autonomous: true
must_haves:
  truths:
    - "The only cross-schema user privilege is a temporary deployment-time REFERENCES(canonical_user_id) grant to jank_migrator."
    - "The grant is revoked after FK validation; jank_owner, jank_migrator, jank_app, and later jank deletion roles have no users SELECT or DML."
  artifacts:
    - path: "docs/contracts/jank-user-fk-grant.sql"
      provides: "exact temporary REFERENCES grant, FK validation, and revoke sequence"
  key_links:
    - from: "RI-07 canonical-user foreign keys"
      to: "users.canonical_user_id"
      via: "the reviewed deployment-only grant/revoke contract"
---

<objective>Codify and test the exact least-privilege foreign-key bridge independently from UUID/API cutover. Output: executable SQL, role negatives, and deployment guidance consumed by RI-07.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-01-01-PLAN.md @docs/plans/2026-08-19-runtime-intelligence/RI-01-02-PLAN.md @docs/contracts/runtime-database-roles.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Prove temporary REFERENCES is sufficient and no broader privilege exists</name>
  <files>server/runtime_role_test.go</files>
  <action>Add a migration-role fixture that grants only column-scoped REFERENCES, installs and validates a representative jank-owned FK, revokes the grant, then proves the FK remains enforced. Assert users SELECT/INSERT/UPDATE/DELETE/TRIGGER and credential-table access are denied to every jank role before, during, and after installation.</action>
  <acceptance_criteria>The test fails if FK installation needs broader access, if revoke is omitted, or if any jank role can read/mutate identity rows.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestJankUserFKGrant|TestRuntimeRoleMatrix' -race -v</automated></verify>
  <done>The database test proves the narrow grant is sufficient and temporary.</done>
</task>

<task type="auto">
  <name>Task 2: Publish the executable grant, validation, revoke, and audit contract</name>
  <files>docs/contracts/jank-user-fk-grant.sql, docs/contracts/jank-user-fk-grant.md, server/runtime_role_test.go</files>
  <action>Write the reviewed SQL as a deployment-only sequence: grant REFERENCES on canonical_user_id to jank_migrator, install every named jank foreign key, validate each constraint, revoke the grant, and query information_schema/acl state to prove removal. Document operator, transaction/nontransaction boundaries, failure cleanup, evidence capture, and that runtime roles never receive users access.</action>
  <acceptance_criteria>The contract names every allowed statement and postcondition; RI-07 can execute it without interpreting privilege intent.</acceptance_criteria>
  <verify><automated>grep -q 'GRANT REFERENCES (canonical_user_id)' docs/contracts/jank-user-fk-grant.sql &amp;&amp; grep -q 'REVOKE REFERENCES (canonical_user_id)' docs/contracts/jank-user-fk-grant.sql &amp;&amp; go test ./server/... -run TestJankUserFKGrant -race</automated></verify>
  <done>RI-07 has one tested, auditable FK installation contract with no SELECT or DML expansion.</done>
</task>

</tasks>
<verification>Role matrix and jank FK grant/revoke tests plus exact SQL contract checks.</verification>
<success_criteria>D-07 remains enforced while jank can install canonical-user foreign keys.</success_criteria>

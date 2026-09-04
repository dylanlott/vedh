---
phase: RI-01-shared-database-identity
plan: RI-01-01
type: execute
wave: 2
depends_on: ["RI-00"]
files_modified:
  - server/migration_parity_test.go
  - server/runtime_role_test.go
  - persistence/migrations/20260820090000_runtime_roles.up.sql
  - persistence/migrations/20260820090000_runtime_roles.down.sql
  - persistence/migrations_test/20260820090000_runtime_roles.up.sql
  - persistence/migrations_test/20260820090000_runtime_roles.down.sql
  - docs/contracts/runtime-database-roles.md
requirements: [D-07]
autonomous: false
must_haves:
  truths:
    - "vEDH and jank have distinct owner, migrator, and runtime roles with explicit object privileges per D-07."
    - "jank_app has no SELECT or DML on users, credentials, private evidence, or vEDH base tables."
    - "Production and test migrations install the same schema from empty and from the RI-00 current-schema fixture."
  artifacts:
    - path: "server/migration_parity_test.go"
      provides: "empty/upgrade migration parity and role capability harness"
    - path: "docs/contracts/runtime-database-roles.md"
      provides: "exact positive and negative role matrix"
  key_links:
    - from: "server/runtime_role_test.go"
      to: "20260820090000_runtime_roles.up.sql"
      via: "connect as each runtime/deployment role and assert capabilities"
---

<objective>Establish the enforceable PostgreSQL ownership boundary before canonical identity or jank foreign keys are introduced. Output: mirrored role DDL, migration fixtures, and an explicit least-privilege contract.</objective>

<execution_context>
@docs/plans/2026-08-19-runtime-intelligence/00-CONTEXT.md
@docs/plans/2026-08-19-runtime-intelligence/00-RESEARCH.md
@docs/contracts/runtime-intelligence-baseline.md
@persistence/sql.go
</execution_context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create Wave-0 migration and privilege fixtures</name>
  <read_first>persistence/sql.go; server/main_test.go; persistence/migrations/; persistence/migrations_test/</read_first>
  <files>server/migration_parity_test.go, server/runtime_role_test.go</files>
  <action>Write `TestRuntimeRoleContract` to apply production and test migration trees from empty and from the RI-00 schema fixture, compare normalized schema fingerprints, rerun backfills, and exercise flag rollback. Connect as vedh_owner, vedh_migrator, vedh_app, jank_owner, jank_migrator, and jank_app. Assert allowed schema usage and owned-object actions plus denial of PUBLIC CREATE, role escalation, arbitrary functions, credentials, guest secrets, private tables, cross-schema writes, and owner bypass. When the pre-RI schema lacks the contract, fail only this named test with `EXPECTED_RED[RI-01-01-T1]: runtime role objects and privileges are not installed`; do not convert build, database setup, or unrelated failures into the expected result.</action>
  <acceptance_criteria>Both install paths are deterministic; every role has positive and negative assertions; migration files are paired and schema-equivalent.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-01-01-T1 --require-test TestRuntimeRoleContract --require-reason 'runtime role objects and privileges are not installed' -- go test ./server/... -run '^TestRuntimeRoleContract$' -race -v</automated></verify>
  <done>The new tests fail against the pre-RI schema for named missing privileges/objects and are discoverable by the live server test command.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Install runtime roles, schemas, and restrictive defaults</name>
  <read_first>server/runtime_role_test.go; docs/plans/2026-08-19-runtime-intelligence/00-RESEARCH.md section Database roles and schemas</read_first>
  <files>persistence/migrations/20260820090000_runtime_roles.up.sql, persistence/migrations/20260820090000_runtime_roles.down.sql, persistence/migrations_test/20260820090000_runtime_roles.up.sql, persistence/migrations_test/20260820090000_runtime_roles.down.sql, docs/contracts/runtime-database-roles.md</files>
  <action>Create NOLOGIN owners, deployment-only migrators, runtime roles, and jank/evidence_api schemas. Revoke CREATE on public from PUBLIC; grant schema USAGE separately from object rights; define restrictive default privileges for future tables, sequences, views, and functions. Document the exact matrix and that user foreign-key installation receives only a temporary column-scoped REFERENCES grant in RI-01-02/RI-07—never SELECT or DML. Down migration may remove unused roles only before data exists; normal rollback is credential/flag based.</action>
  <acceptance_criteria>Role tests prove D-07 at the database boundary, future objects inherit restrictive privileges, and no runtime credential owns objects.</acceptance_criteria>
  <verify><automated>go test ./server/... -run '^TestRuntimeRoleContract$' -race -v &amp;&amp; go test ./server/... -run 'TestRuntimeMigrationParity|TestRuntimeRoleMatrix' -race -v &amp;&amp; diff persistence/migrations/20260820090000_runtime_roles.up.sql persistence/migrations_test/20260820090000_runtime_roles.up.sql</automated></verify>
  <done>The database role contract passes all positive/negative assertions and is ready for canonical UUID and jank FK work.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 3: Approve the role boundary before credential cutover</name>
  <files>docs/contracts/runtime-database-roles.md</files>
  <action>Review the production-shaped role matrix and restore evidence. Record APPROVED or REJECTED, operator, timestamp, evidence, and remediation. A rejected or absent decision leaves this plan incomplete and must not satisfy RI-01-02.</action>
  <acceptance_criteria>The decision is explicit and includes the exact temporary REFERENCES policy and proof that jank_app has no users SELECT/DML.</acceptance_criteria>
  <verify><automated>grep -q 'Role boundary decision: APPROVED' docs/contracts/runtime-database-roles.md &amp;&amp; go test ./server/... -run TestRuntimeRoleMatrix -race</automated></verify>
  <done>The role boundary is explicitly APPROVED with evidence; rejection/absence halts identity work and is not dependency completion.</done>
</task>

</tasks>

<verification>Run `go test ./server/... -run 'TestRuntimeMigrationParity|TestRuntimeRoleMatrix' -race` against empty and current-schema fixtures.</verification>
<success_criteria>D-07 is enforced before later identity/evidence migrations and the deployment-only FK privilege is narrowly specified.</success_criteria>

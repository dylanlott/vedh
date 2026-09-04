---
phase: RI-06-privacy-safe-publication
plan: RI-06-03
type: execute
wave: 21
depends_on: ["RI-06-02"]
files_modified:
  - persistence/migrations/20260820141000_evidence_api_v1.up.sql
  - persistence/migrations/20260820141000_evidence_api_v1.down.sql
  - persistence/migrations_test/20260820141000_evidence_api_v1.up.sql
  - persistence/migrations_test/20260820141000_evidence_api_v1.down.sql
  - server/evidence_publications_test.go
  - docs/contracts/evidence-api-v1.md
requirements: [D-07, D-08, D-18]
autonomous: true
must_haves:
  truths:
    - "jank_app has USAGE and SELECT only on versioned safe evidence_api views, never vEDH base/private/pseudonym-map objects or functions."
    - "Published/revoked status, exact n/uncertainty, calculation/source/inference versions, and safe source counts are stable and authorization-filtered."
  artifacts:
    - path: "docs/contracts/evidence-api-v1.md"
      provides: "versioned read-only jank evidence contract and grant matrix"
  key_links:
    - from: "evidence_api v1 views"
      to: "materialized publication versions/status"
      via: "vEDH-owned security-barrier allowlist"
---

<objective>Publish the read-only evidence_api contract and strict jank_app grants as a focused database boundary. Output: mirrored DDL, role negatives, revocation behavior, and versioned contract documentation.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-06-02-PLAN.md @docs/contracts/runtime-database-roles.md</execution_context>
<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Install evidence_api v1 safe views and grants</name>
  <files>persistence/migrations/20260820141000_evidence_api_v1.up.sql, persistence/migrations/20260820141000_evidence_api_v1.down.sql, persistence/migrations_test/20260820141000_evidence_api_v1.up.sql, persistence/migrations_test/20260820141000_evidence_api_v1.down.sql, server/evidence_publications_test.go, docs/contracts/evidence-api-v1.md</files>
  <action>Create vEDH-owned security-barrier status/publication/card/source views exposing only materialized safe fields, exact n/uncertainty/filters/calculation/source/inference metadata, stable version IDs, safe source counts, and revocation/redaction tombstones. Grant jank_app schema USAGE and SELECT on these views only. Test denial of users, credentials, private gameplay/reviews, base publication/source tables, raw payloads, pseudonym mappings, arbitrary functions, DML, and future objects. Document cache invalidation and contract versioning.</action>
  <acceptance_criteria>jank_app resolves published/revoked status but every base/private access fails; migration trees are identical.</acceptance_criteria>
  <verify><automated>go test ./server/... -run 'TestEvidenceAPIV1|TestEvidenceAPIRole|TestEvidenceAPIRevocation' -race &amp;&amp; diff persistence/migrations/20260820141000_evidence_api_v1.up.sql persistence/migrations_test/20260820141000_evidence_api_v1.up.sql</automated></verify>
  <done>jank has one stable read-only evidence contract and no private/base privilege.</done>
</task>

</tasks>
<verification>Evidence API migration parity, allowlist contract, revocation, and complete role-negative matrix.</verification>
<success_criteria>D-07/D-08/D-18 are enforced at the shared-database publication boundary.</success_criteria>

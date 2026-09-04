---
phase: RI-06-privacy-safe-publication
plan: RI-06-01
type: execute
wave: 19
depends_on: ["RI-05-03"]
files_modified:
  - persistence/migrations/20260820140000_evidence_publications.up.sql
  - persistence/migrations/20260820140000_evidence_publications.down.sql
  - persistence/migrations_test/20260820140000_evidence_publications.up.sql
  - persistence/migrations_test/20260820140000_evidence_publications.down.sql
  - pkg/publication/builder.go
  - pkg/publication/builder_test.go
  - pkg/publication/testdata/adversarial-v1.json
  - server/evidence_publications_test.go
requirements: [RGI-6, RDA-7, D-08, D-18]
autonomous: true
must_haves:
  truths:
    - "RGI-6 exposes only explicit authorization-filtered materialized evidence to jank."
    - "RDA-7 permits n=1 with exact sample/uncertainty and explicit probabilistic selection."
    - "Publication allowlists hidden/private/opponent data out and uses version-scoped pseudonyms without mutating sources."
  artifacts:
    - path: "pkg/publication/builder.go"
      provides: "pure allowlist publication transform"
  key_links:
    - from: "builder preview hash"
      to: "immutable publication version"
      via: "same canonical bytes"
---

<objective>Define the privacy boundary and immutable publication storage for RGI-6/RDA-7. Output: adversarial corpus, pure allowlist builder, scoped pseudonyms, and mirrored DDL.</objective>
<execution_context>@docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md @docs/product/2026-08-17-runtime-deck-analysis-prd.md @docs/plans/2026-08-19-runtime-intelligence/RI-05-03-PLAN.md</execution_context>
<tasks>
<task type="auto" tdd="true">
  <name>Task 1: Build adversarial allowlist, pinning, n=1, and revocation fixtures</name>
  <files>pkg/publication/testdata/adversarial-v1.json, pkg/publication/builder_test.go, server/evidence_publications_test.go</files>
  <action>Write `TestPublicationPersistenceContract` covering hands/libraries/face-down data, stable opponent UUID/name, raw event/review payloads, unselected notes, unsafe markup/URLs, linked IDs, speculative inference unselected, selected probabilistic evidence, n=1/2/multi denominators, cross-publication pseudonym comparison, source edits, revoked/deleted author, cache replay, and missing source. Assert an explicit output-key allowlist, preview/create hash parity, immutable pinning, explicit republish, tombstone, and unchanged canonical sources. When the builder/storage contract is absent, fail only this named test with `EXPECTED_RED[RI-06-01-T1]: evidence publication contract is not installed`.</action>
  <acceptance_criteria>Any unapproved key fails; pseudonyms are unlinkable across versions; RDA-7 sample/uncertainty rules are fixture-proven.</acceptance_criteria>
  <verify><automated>scripts/verify-expected-red.sh --suite RI-06-01-T1 --require-test TestPublicationPersistenceContract --require-reason 'evidence publication contract is not installed' -- go test ./pkg/publication ./server/... -run '^TestPublicationPersistenceContract$' -race -v</automated></verify>
  <done>The privacy and evidence-preservation boundary is executable before persistence.</done>
</task>
<task type="auto" tdd="true">
  <name>Task 2: Install immutable publications and pure sanitizer</name>
  <files>persistence/migrations/20260820140000_evidence_publications.up.sql, persistence/migrations/20260820140000_evidence_publications.down.sql, persistence/migrations_test/20260820140000_evidence_publications.up.sql, persistence/migrations_test/20260820140000_evidence_publications.down.sql, pkg/publication/builder.go, pkg/publication/builder_test.go</files>
  <action>Create stable publications, immutable versions, typed exact sources, version-scoped pseudonym mappings, audit events, and lifecycle/revocation state. Implement canonical allowlist builders for game excerpt, debrief excerpt, and analysis slice. Require already-authorized sources; include exact n/uncertainty/calculation/source/inference metadata; strip hidden/raw/private/unselected/stable-opponent fields; never mutate source records.</action>
  <acceptance_criteria>All adversarial fixtures pass; immutable rows reject update/delete; production/test migration parity passes.</acceptance_criteria>
  <verify><automated>go test ./pkg/publication ./server/... -run '^TestPublicationPersistenceContract$' -race -v &amp;&amp; go test ./pkg/publication ./server/... -run 'TestPublicationBuilder|TestPublicationPersistence' -race &amp;&amp; diff persistence/migrations/20260820140000_evidence_publications.up.sql persistence/migrations_test/20260820140000_evidence_publications.up.sql</automated></verify>
  <done>RGI-6/RDA-7 have a stable, source-pinned, n=1-safe materialized publication model.</done>
</task>
</tasks>
<verification>`go test ./pkg/publication ./server/... -race`; migration parity; adversarial corpus.</verification>
<success_criteria>Private evidence cannot become public except through the explicit immutable allowlist transform.</success_criteria>

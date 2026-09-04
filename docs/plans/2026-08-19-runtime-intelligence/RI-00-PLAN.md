---
phase: RI-00-release-handoff
plan: RI-00
type: execute
wave: 1
depends_on: []
files_modified:
  - docs/runbooks/runtime-intelligence-handoff.md
  - docs/contracts/runtime-intelligence-baseline.md
  - docs/operations/runtime-intelligence-inventory.example.json
  - docs/contracts/jank-checkout.json
  - scripts/runtime-inventory.sh
  - scripts/verify-jank-checkout.sh
  - scripts/verify-expected-red.sh
  - scripts/runtime-restore-rehearsal.sh
autonomous: false
requirements: [RI-ENTRY, D-02, D-19, D-22]
entry_precondition: "Active .planning Phase 5 must have an actual completed release-gate artifact; until it exists, RI-00 stays blocked and no Runtime Intelligence DDL or flag may run. This is an external milestone condition, not a plan dependency ID."
must_haves:
  truths:
    - "Active Deck-to-Game Phase 5 is closed before any runtime-intelligence migration or flag enablement."
    - "The vEDH and jank commits, domains, proxies, PostgreSQL topology, roles, schema fingerprints, row counts, event types, backups, and rollback owner are recorded from the real target environment."
    - "A production-shaped backup restores into an isolated database and the restored schema/data inventory matches the source."
    - "The inventory attests that no spreadsheet records or values enter migrations, fixtures, metrics, or analysis per D-02."
    - "RI-00 through RI-08 contain no testing-group or non-Magic implementation per D-19/D-22."
    - "A dedicated jank integration branch/worktree records immutable base_commit and remote plus mutable expected_head; later jank plans advance expected_head only through audited atomic commits."
    - "Test-first tasks use one named expected-red harness that returns success only for allowlisted test failures with the plan/task sentinel and rejects build, setup, environment, panic, timeout, race, no-test, and unrelated-test failures."
  prohibitions:
    - "No write to active `.planning` artifacts and no runtime-intelligence production DDL before approval."
    - "No spreadsheet-derived values, testing groups, or non-Magic adapter selection."
  artifacts:
    - path: "docs/runbooks/runtime-intelligence-handoff.md"
      provides: "entry, restore, rollback, ownership, and sign-off runbook"
    - path: "docs/contracts/runtime-intelligence-baseline.md"
      provides: "versioned current-schema/event/auth/deployment baseline"
    - path: "scripts/runtime-inventory.sh"
      provides: "read-only deterministic inventory collector"
    - path: "scripts/runtime-restore-rehearsal.sh"
      provides: "isolated restore and schema/data reconciliation harness"
    - path: "scripts/verify-expected-red.sh"
      provides: "named expected-failure verifier for independently executable RED tasks"
  key_links:
    - from: "scripts/runtime-inventory.sh"
      to: "docs/operations/runtime-intelligence-inventory.example.json"
      via: "stable machine-readable inventory schema"
    - from: "docs/runbooks/runtime-intelligence-handoff.md"
      to: "the actual completed active Phase 5 release-gate summary when it exists"
      via: "fail-closed external entry evidence, never a fabricated executable plan ID"
---

<objective>
Create a signed, reproducible handoff from the active v1 release to Runtime Intelligence without changing runtime behavior.

Purpose: every later migration assumes facts the repository cannot prove—production row shape, deployment domains, jank commit, role topology, backup restore behavior, and an approved rollback operator. RI-0 turns those assumptions into an entry contract while enforcing D-02/D-19/D-22.

Output: read-only inventory tooling, a production-shaped restore rehearsal, the canonical baseline contract, and a blocking go/no-go record.
</objective>

<execution_context>
@docs/plans/2026-08-19-runtime-intelligence/00-CONTEXT.md
@docs/plans/2026-08-19-runtime-intelligence/00-RESEARCH.md
@docs/plans/2026-08-19-runtime-intelligence/00-MASTER-PLAN.md
@.planning/ROADMAP.md
@.planning/STATE.md
</execution_context>

<tasks>

<task type="auto">
  <name>Task 1: Freeze the v1 release handoff and deployment inventory</name>
  <read_first>
    - .planning/ROADMAP.md — Phase 5 release gate and rollback promises
    - .planning/STATE.md — active milestone/phase; RI-0 must fail closed while it is incomplete
    - server/graphql.go — current `Conf`, flags, ports, origins, and runtime wiring
    - server/schema.graphql — current public API fingerprint
    - persistence/migrations/ — production schema history
    - persistence/migrations_test/ — test schema history that must remain in parity
    - .github/workflows/test.yml and .github/workflows/smoke-rust.yml — current automated gates
  </read_first>
  <files>scripts/runtime-inventory.sh, scripts/verify-jank-checkout.sh, scripts/verify-expected-red.sh, docs/operations/runtime-intelligence-inventory.example.json, docs/contracts/runtime-intelligence-baseline.md, docs/contracts/jank-checkout.json</files>
  <action>
Create a read-only `scripts/runtime-inventory.sh` that emits JSON and exits nonzero unless the caller supplies explicit vEDH and jank commit IDs, sanitized deployment origins/proxy topology, database DSN name (never credentials), backup identifier, and Phase 5 completion evidence. Query PostgreSQL catalogs for server version, database/schema/role names, object ownership, default privileges, migration rows, schema fingerprints, table row estimates, null/invalid/duplicate `users.uuid` counts, `games`/`gamelog` counts, distinct legacy event types, and current indexes/constraints. Record only counts and names—never payloads, credentials, review text, deck text, hidden game state, usernames, or UUID values.

Define the JSON schema in `runtime-intelligence-inventory.example.json` and document exact collection commands and current GraphQL/auth/event compatibility surfaces in `runtime-intelligence-baseline.md`. Add explicit booleans `spreadsheet_data_used=false`, `testing_groups_in_beta=false`, and `non_magic_adapter_selected=false` per D-02/D-19/D-22; the script must reject true/unknown values. Do not infer production topology from localhost defaults.

Require an explicit jank integration-worktree path and integration-branch name. If the repository is absent, the operator may supply an explicit clone destination and repository URL; never infer a conventional sibling location. Resolve and record immutable `remote` and `base_commit`, create or validate a dedicated integration branch/worktree outside the vEDH worktree, and initialize mutable `expected_head=base_commit`. `docs/contracts/jank-checkout.json` stores schema version, real path, remote, immutable base commit, integration branch, expected head, and an append-only advancement audit containing plan ID, prior head, new head, commit subject, and timestamp. No production push or merge is part of this protocol.

Implement `verify-jank-checkout.sh` modes. `--mode pre-plan --plan <PLAN>` requires the configured remote, base as an ancestor of HEAD, HEAD exactly equal to expected_head, the configured branch, and a clean worktree. `--mode task --plan <PLAN>` permits dirt only in that plan's declared `${JANK_CHECKOUT_PATH}` files while retaining the remote/base/head checks; it rejects undeclared untracked or modified paths and never demands an impossible clean tree after an edit. `--mode prepare-commit` requires the same declared-path subset before the executor creates the one atomic jank commit. `--mode advance --new-head <sha>` accepts only the immediate committed descendant produced for that plan, updates expected_head and the audit record, and `--mode post-plan` requires the new expected HEAD and a clean worktree. All modes fail closed on a wrong remote/branch, missing base ancestry, divergence, unexpected HEAD, or undeclared dirt. Every later jank-touching plan uses these modes.

Create `verify-expected-red.sh` with `--suite`, repeated `--require-test`, and `--require-reason` arguments followed by `-- <command>`. Capture combined output and require the command to fail, each allowlisted test to appear as a failing Go test, and the exact `EXPECTED_RED[<suite>]` reason. Reject compilation/package setup failures, missing tools/files/databases, connection/environment failures, panics, timeouts, races, `no tests to run`, and any failed test outside the allowlist. Exit zero only for the intended RED result; self-tests cover expected failure, accidental green, compile/setup failure, missing named test/reason, and unrelated failure.

In the same baseline and handoff artifacts, record accepted-mutation/event reconciliation, analysis query p50/p95, finished-game completeness, card-ID resolution, authorization status, and current Prometheus/product-event names. For every master-plan flag, record owner, enable prerequisite, disable trigger, and legacy fallback. Mark unavailable evidence `not_measured` with its future collection query; never fabricate zero. Record the vEDH/jank schema owners and the operators authorized for migration, rollback, privacy sampling, and deletion drills. Metric examples use bounded outcome/version/category labels only.
  </action>
  <acceptance_criteria>
    - The script is read-only, shellcheck-clean where shellcheck is available, and contains no `INSERT`, `UPDATE`, `DELETE`, `ALTER`, `DROP`, `CREATE`, or credential echo.
    - JSON output includes commits, topology, role/object ownership, schema fingerprints, migration parity, row-quality counts, backup ID, Phase 5 evidence, and the verified jank checkout artifact.
    - The three locked scope booleans are present and false; no raw application record or secret is emitted.
    - Every master-plan feature flag has an owner, enable prerequisite, disable trigger, and fallback; baseline values distinguish measured, not-measured, and unavailable.
  </acceptance_criteria>
  <verify>
    <automated>bash -n scripts/runtime-inventory.sh scripts/verify-jank-checkout.sh scripts/verify-expected-red.sh &amp;&amp; scripts/verify-jank-checkout.sh --self-test &amp;&amp; scripts/verify-expected-red.sh --self-test &amp;&amp; jq -e '.spreadsheet_data_used == false and .testing_groups_in_beta == false and .non_magic_adapter_selected == false' docs/operations/runtime-intelligence-inventory.example.json &amp;&amp; for f in VEDH_RI_NATIVE_UUID_READ VEDH_RI_SNAPSHOT_DUAL_WRITE VEDH_RI_CANONICAL_EVENT_WRITE VEDH_RI_TYPED_TIMELINE_READ VEDH_RI_INFERENCE_READ VEDH_RI_REVIEW VEDH_RI_ANALYTICS_READ VEDH_RI_PUBLICATION_WRITE VEDH_RI_PUBLICATION_READ JANK_SHARED_IDENTITY JANK_EVIDENCE_TREES VEDH_RI_BETA_COHORT; do grep -q "$f" docs/runbooks/runtime-intelligence-handoff.md || exit 1; done</automated>
  </verify>
  <done>The inventory, expected-red harness, and integration-worktree contract are safe and complete; later plans can fail closed against expected_head and advance it through audited commits.</done>
</task>

<task type="auto">
  <name>Task 2: Rehearse backup restore and baseline reconciliation</name>
  <read_first>
    - persistence/sql.go — live production migration launcher and DSN behavior
    - server/main_test.go — current destructive test reset behavior; never point it at production
    - Makefile — migration and local PostgreSQL commands
    - docs/plans/2026-08-19-runtime-intelligence/00-RESEARCH.md — migration parity and restore requirements
  </read_first>
  <files>scripts/runtime-restore-rehearsal.sh, docs/runbooks/runtime-intelligence-handoff.md</files>
  <action>
Implement `runtime-restore-rehearsal.sh` around explicit source backup and isolated target database arguments. Validate the target database name against a required `ri_rehearsal_` prefix before any restore or cleanup; refuse the production/current database. Restore the production-shaped backup, run the inventory collector against the restored target, compare migration IDs, schema fingerprints, table counts, and `games`/`gamelog` checksums that exclude JSON payload contents, then run `make generate`, `make test-unit`, and the PostgreSQL-backed API migration smoke against the target.

Document restore, verification, cleanup, failure preservation, and elapsed-time capture in the handoff runbook. Cleanup must be opt-in and target-validated; a failed rehearsal database stays available for diagnosis. The runbook records backup/replica/cache/log retention questions that RI-8 must close for D-13.
  </action>
  <acceptance_criteria>
    - The harness refuses an empty, current, `postgres`, or non-`ri_rehearsal_` target database.
    - A restore produces a machine-readable reconciliation report and preserves failures for diagnosis.
    - No destructive action targets a broad path/database or runs without an explicit validated target.
  </acceptance_criteria>
  <verify>
    <automated>bash -n scripts/runtime-restore-rehearsal.sh &amp;&amp; scripts/runtime-restore-rehearsal.sh --self-test-target-guards</automated>
  </verify>
  <done>An isolated production-shaped restore reconciles or leaves a preserved diagnostic target, with no destructive command able to address an unvalidated database.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 3: Approve the post-v1 entry gate</name>
  <read_first>
    - docs/contracts/runtime-intelligence-baseline.md — exact baseline being approved
    - docs/runbooks/runtime-intelligence-handoff.md — restore result, ownership, and open facts
    - generated runtime inventory and restore reconciliation report — production evidence
  </read_first>
  <files>docs/runbooks/runtime-intelligence-handoff.md</files>
  <action>
Present the completed inventory and restore report only after the active ROADMAP Phase 5 release-gate summary proves closure. The reviewer confirms the two repository commits, target domains/proxy, session-topology inputs required by the RI-01 canonical-session checkpoint, backup/restore result, migration/rollback/privacy/deletion owners, and D-02/D-19/D-22 attestations. Record APPROVED or REJECTED with identity/time and evidence. A REJECTED or absent decision leaves this plan incomplete; do not enable a flag or run runtime-intelligence DDL.
  </action>
  <acceptance_criteria>
    - Phase 5 closure evidence and a successful isolated restore are linked.
    - Deployment topology is concrete enough for RI-01's session decision.
    - Reviewer explicitly approves or rejects the entry; silence is not approval.
  </acceptance_criteria>
  <verify>
    <automated>test -s docs/contracts/runtime-intelligence-baseline.md &amp;&amp; test -s docs/runbooks/runtime-intelligence-handoff.md &amp;&amp; grep -q 'Entry gate decision: APPROVED' docs/runbooks/runtime-intelligence-handoff.md</automated>
  </verify>
  <done>The reviewer records APPROVED with Phase 5 closure and restore evidence; rejection/absence halts the package and is not dependency completion.</done>
</task>

</tasks>

## Artifacts this phase produces

- New files: `scripts/runtime-inventory.sh`, `scripts/runtime-restore-rehearsal.sh`, `docs/operations/runtime-intelligence-inventory.example.json`, `docs/contracts/runtime-intelligence-baseline.md`, `docs/runbooks/runtime-intelligence-handoff.md`.
- New operational categories: release handoff ID, schema fingerprint, role/object-ownership matrix, quality baseline, flag owner, restore rehearsal report, entry decision.
- No tables, APIs, application symbols, or enabled flags are produced in RI-0.

<verification>
- `bash -n scripts/runtime-inventory.sh scripts/runtime-restore-rehearsal.sh`
- `scripts/runtime-restore-rehearsal.sh --self-test-target-guards`
- On an approved isolated restore: `make generate &amp;&amp; make test-unit &amp;&amp; make test-api`
- Inventory contains no secrets or raw payloads and records explicit D-02/D-19/D-22 attestations.
</verification>

<success_criteria>
- Active Phase 5 is closed and the reviewer has approved the runtime inventory and restore evidence.
- Later plans have exact repository commits, database topology, domains, role baseline, quality baseline, and rollback owners.
- No runtime behavior, production schema, active planning artifact, testing-group scope, or adapter scope changed.
</success_criteria>

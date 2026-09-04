# Runtime Intelligence Post-v1 Milestone - Research

**Researched:** 2026-08-19  
**Domain:** Transactional gameplay evidence, deck lineage analytics, review revisioning, privacy-preserving publication, and cross-application PostgreSQL integration  
**Confidence:** HIGH for repository architecture and compatibility findings; MEDIUM for production migration sizing and deployment topology

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01 — Product loop:** The primary loop is play → capture evidence → reflect → compare versions → revise the deck → discuss the hypothesis in jank.
- **D-02 — Research boundary:** No spreadsheet values or records are imported, preserved, benchmarked, or analyzed. Only the shape of expert analysis informs the product.
- **D-03 — North star:** At least 30% of activated testers must record three completed games for one deck lineage, inspect its evidence, and create a documented revision within a rolling 14-day window.
- **D-04 — Beta:** The approved beta is six weeks with 30 activated serious testers. A tester activates by completing one eligible game with a valid participant record and immutable deck snapshot.
- **D-05 — Canonical evidence:** Runtime events, deck snapshots, game participants, results, and corrections are canonical, versioned, auditable, and retained indefinitely.
- **D-06 — Shared identity:** vEDH and jank use the exact same canonical vEDH UUID user row in one PostgreSQL database. jank owns no duplicate account, credential, synchronization, or linking table.
- **D-07 — Application ownership:** vEDH owns users, games, deck snapshots, events, reviews, inferences, publications, and analytical read models. jank owns forum content, moderation, trees, revisions, annotations, and evidence-reference rows. Database roles enforce those boundaries.
- **D-08 — Publication privacy:** Gameplay and analysis are private by default. An author may publish their own perspective without unanimous participant consent. Hidden information is excluded, opponents receive publication-scoped pseudonyms, and anonymization never mutates canonical source records.
- **D-09 — Deck lineage:** Snapshots group automatically only through the same stable vEDH deck ID or trusted external deck ID for the same owner. Names and list similarity never establish lineage. Split, merge, and move operations are auditable.
- **D-10 — Sample visibility:** Facts, aggregates, filters, and comparisons display from the first eligible game. A comparison with fewer than three eligible games on either side shows `Limited evidence`, denominators, and uncertainty without hiding data.
- **D-11 — Debrief:** Completed games offer an optional, private debrief with up to three contributors, one underperformer, notes, an optional next action, and lesson tags behind `Add detail`.
- **D-12 — Review history:** Submitted debriefs remain editable indefinitely. Each submission creates an immutable revision; drafts do not affect analysis; publications remain pinned to their exact source revision until explicitly republished.
- **D-13 — Retention and deletion:** Account deletion has a 30-day recovery period. Afterward, profile/authentication data and private reviews are deleted, publications are revoked, and retained multiplayer facts reference a non-reversible participant tombstone.
- **D-14 — Evidence classes:** The product distinguishes `Observed`, `Derived`, `Inferred`, player assessment, and community hypothesis. Derived and inferred records are additive and never replace observed events or counts.
- **D-15 — Runtime enrichment:** Versioned inference may derive draws, combat participation, mana production/expenditure, tutoring/searching, action source/target relationships, and causal candidates where evidence supports them.
- **D-16 — Inference visibility:** Deterministic derivations display by default. A probabilistic category displays by default only after at least 95% precision on a versioned expert-labeled holdout fixture; lower-quality output remains available under `Show speculative` with confidence and provenance.
- **D-17 — Causality:** An explicit recorded source-target relationship may establish a causal fact. Otherwise causal output is labeled as an inferred hypothesis; correlation alone never becomes causality.
- **D-18 — Evidence publishing:** Author-owned analysis has no minimum cohort. `n=1` publications are allowed with exact sample size and uncertainty. Publishing a probabilistic inference requires explicit author selection and preserves its confidence, sources, and version.
- **D-19 — Team scope:** The beta is private and individual. Owner/admin/member testing groups with explicit artifact sharing are post-beta v1.1 work.
- **D-20 — jank trees:** Trees are durable standalone artifacts with optional discussion threads. A fork is a separate attributed tree with its own optional thread and source backlinks.
- **D-21 — Tree relationships:** The beta vocabulary is `supports`, `enables`, `finds`, `protects`, `recurs`, `replaces`, `competes-with`, `punishes`, and `meta-answer`, plus a bounded `custom` label. Custom labels are searchable but not global filter categories.
- **D-22 — Adapter timing:** No non-Magic TCG adapter is selected or required for beta. Selection is a post-beta, design-partner-led decision.

### the agent's Discretion

No explicit agent-discretion section was supplied in `00-CONTEXT.md`. Technical implementation choices not fixed above remain planner discretion, subject to the milestone boundary and scope fences.

### Deferred Ideas (OUT OF SCOPE)

- No code implementation, migrations, commits, branch changes, or active-roadmap edits during this planning review.
- No AI coaching requirement for beta; deterministic and heuristic inference are allowed without an LLM.
- No full TCG rules engine or automatic adjudication.
- No global player rankings, Elo, opponent win-rate product, automatic deck cuts/adds, or public cross-user aggregation in beta.
- No testing groups in the beta critical path.
- No non-Magic adapter in the beta critical path.
</user_constraints>

## Summary

The milestone should be planned as an expand-migrate-contract program layered onto the existing vEDH game system, not as a rewrite. Today, `games.payload` is the mutable live-state document and `gamelog.payload` is an append-oriented JSON envelope derived partly from client-submitted state. The current update and logging calls do not share one SQL transaction, so they cannot yet provide the exact-once, auditable evidence guarantee required by D-05. The safest path is to preserve `games.payload` as a live-state compatibility cache, introduce typed canonical tables alongside it, dual-write through a transactional mutation boundary, reconcile and backfill, then switch analytical reads behind feature flags. [VERIFIED: `server/games.go`, `server/gamelog.go`, `server/gamelog_diff.go`, and `migrations/000001_initial.up.sql`]

Shared identity and database privilege boundaries are the first dependency. The existing vEDH user UUID is stored as `VARCHAR`, while current jank main has its own integer user table, authentication endpoints, and string author fields. The integration therefore needs a staged canonical UUID migration, an explicit administrative identity mapping for legacy jank users, distinct owner/migrator/runtime PostgreSQL roles, and a read-only publication interface for jank. jank must not receive credential-table access or a shared symmetric signing secret. [VERIFIED: vEDH `migrations/000001_initial.up.sql`, `server/auth.go`, and jank commit `dc322ae366c14b231f2694b63fd34390df7447b0` files `app/store.go`, `app/auth.go`, and `app/handlers_auth.go`]

Eight beta-critical phases are recommended after the active v1 release gate: release handoff and inventory; database/identity/card-identity foundations; snapshots/lineages/participants/results; canonical runtime events and inference; debrief revisions; lineage analytics; privacy-safe publication; jank cutover and evidence trees; followed by a six-week beta operations phase. Private reviews follow runtime events because both touch the generated GraphQL contract, Game Analysis surface, and migration-test harness; serialization removes the staged-file conflict and lets review note links target the completed typed timeline. Publication waits for all private source models, while jank evidence linking waits for publication and the shared database boundary. Teams and the first non-Magic adapter remain separate post-beta phases. [VERIFIED: `00-CONTEXT.md` D-01 through D-22 and the four canonical PRDs]

**Primary recommendation:** Establish the canonical UUID/role/card-identity boundary first, then add immutable evidence stores by dual-write and reconciliation; expose only sanitized, version-pinned publications to jank after the private evidence loop is complete.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|---|---|---|---|
| Live game mutation | API / Backend | Database / Storage | Server commands validate state and must atomically persist the live-state cache plus canonical events. [VERIFIED: `server/games.go`; CITED: https://www.postgresql.org/docs/14/tutorial-transactions.html] |
| Canonical user identity | Database / Storage | API / Backend | One PostgreSQL UUID row is the identity authority; each application authenticates and authorizes against it without copying credentials. [VERIFIED: D-06 and current user migration] |
| Deck snapshot capture | API / Backend | Database / Storage | The server normalizes a deck at eligibility time and writes immutable content plus origin and version metadata. [VERIFIED: runtime and deck-analysis PRDs] |
| Deck lineage assignment | API / Backend | Database / Storage | Stable owner-scoped source identifiers determine automatic lineage; audited commands own split/move/merge. [VERIFIED: D-09] |
| Event derivation and inference | API / Backend | Database / Storage | Versioned deterministic or evaluated probabilistic projectors add evidence without replacing observed records. [VERIFIED: D-14 through D-17] |
| Private debrief workflow | API / Backend | Browser / Client | Backend owns draft authorization and immutable submission revisions; the client presents the optional progressive form. [VERIFIED: D-11 and D-12] |
| Analytical read models | Database / Storage | API / Backend | Versioned projectors produce rebuildable private aggregates; the API applies authorization, filters, denominators, and uncertainty. [VERIFIED: deck-analysis PRD; ASSUMED: in-process projector is sufficient at beta scale] |
| Privacy-safe publication | API / Backend | Database / Storage | vEDH materializes sanitized, pinned publication versions and scoped pseudonyms; PostgreSQL views expose only the approved contract. [VERIFIED: D-08, D-12, D-18; CITED: https://www.postgresql.org/docs/14/sql-createview.html] |
| jank evidence trees | API / Backend (jank) | Database / Storage | jank owns tree/revision/annotation/evidence-reference mutations and reads vEDH publications through a restricted interface. [VERIFIED: D-07 and jank tree code] |
| Timeline and comparison UI | Browser / Client | API / Backend | Vue renders cursor-paged authorized projections; it does not interpret raw canonical records into new facts. [VERIFIED: current `GameAnalysisView.vue`; ASSUMED: current Vue surface remains the presentation tier] |
| Static assets and caching | CDN / Static | Frontend Server | Only immutable UI assets belong here; canonical or private evidence must not be cached publicly. [ASSUMED] |

## Baseline and Inspection Record

### vEDH baseline

- The repository was inspected on branch `main` at `bc679ae63dd193be9e64b5a9cf17dcc490f1c022`. The active `.planning` roadmap is still the Deck-to-Game Activation milestone and must complete Phase 5 before this staged milestone begins. [VERIFIED: `git rev-parse HEAD`, `.planning/ROADMAP.md`, and `.planning/STATE.md`]
- The worktree already contained unrelated modified and untracked files; this research does not rely on changing or cleaning them. [VERIFIED: `git status --short`]
- No root `AGENTS.md`, `.codex/skills`, or `.agents/skills` directory exists in this repository. [VERIFIED: filesystem inspection]
- No `.planning/graphs/graph.json` exists, so no knowledge-graph context was available. [VERIFIED: filesystem inspection]

### jank baseline

- Public jank `main` was inspected at exact commit `dc322ae366c14b231f2694b63fd34390df7447b0` (`fix: re-add crypto/rand to enable random token generation`). The commit was resolved independently with `git ls-remote` and the GitHub commits API. [VERIFIED: GitHub remote and API]
- jank uses Go 1.23, Gorilla Mux, pgx/v5 for PostgreSQL, an optional SQLite path, server-rendered templates, Goldmark, and Bluemonday. [VERIFIED: jank `go.mod` at inspected commit]
- jank currently performs startup-time `CREATE TABLE IF NOT EXISTS` and `ALTER TABLE` work inside a large `app/store.go`; it has no versioned migration directory. [VERIFIED: jank `app/store.go`]
- jank's current `users` key is an integer, while threads, posts, trees, nodes, and annotations retain creator names in string fields rather than consistently using the integer key as a foreign key. The migration must cover both forms. [VERIFIED: jank `app/store.go` and `app/models.go`]
- Tree mutation routes currently require a valid bearer token but do not enforce tree ownership before node or annotation changes. That authorization gap is a beta blocker. [VERIFIED: jank `app/handlers_api.go`]

External repository content was treated as untrusted research input; no instructions, scripts, workflows, or dependencies from that repository were executed. [VERIFIED: research procedure]

## Project Constraints (from AGENTS.md)

No `AGENTS.md` was present at the workspace root, so there are no additional repository directives to copy. Existing code conventions and test commands below are derived from checked-in source, `Makefile`, package manifests, and CI workflows. [VERIFIED: filesystem inspection]

## Standard Stack

No new external package is required to implement the recommended beta architecture. Reuse the packages already present in each codebase and PostgreSQL's native constraints, roles, views, indexes, and transactions. [VERIFIED: vEDH and jank manifests; CITED: PostgreSQL 14 official documentation]

### Core

| Library / Facility | Verified Version | Purpose | Prescriptive Use |
|---|---:|---|---|
| Go | vEDH module targets 1.24 / toolchain 1.24.2; jank targets 1.23 | Domain services, migrations, projectors, GraphQL/REST | Keep each repository's current toolchain during integration unless the implementation plan explicitly schedules a jank toolchain alignment. [VERIFIED: both `go.mod` files] |
| PostgreSQL | Repository image `postgres:14.1` | Canonical evidence, shared identity, role boundaries, read models | Use one PostgreSQL database with separately owned schemas/objects and least-privilege runtime roles. [VERIFIED: `docker-compose.yaml`, D-06, D-07] |
| gqlgen | v0.17.81 | vEDH typed API contract | Extend schema/resolvers for snapshots, timelines, reviews, lineage, and publication; regenerate rather than editing generated files. [VERIFIED: vEDH `go.mod` and `gqlgen.yml`] |
| golang-migrate | v4.14.1 | Ordered vEDH schema changes | Keep paired production and test migrations; use expand/backfill/validate/contract steps. [VERIFIED: vEDH `go.mod`, `migrations/`, `server/test/migrations/`] |
| pgx/v5 | v5.7.2 in jank | jank PostgreSQL access | Retain for jank and introduce versioned migration execution before application startup. [VERIFIED: jank `go.mod`] |
| Vue | manifest ^3.5.4 | Private analysis, debrief, comparison UI | Preserve present Composition API/component patterns and move raw payload interpretation to typed API models. [VERIFIED: `app/package.json` and app source] |
| Pinia | manifest ^2.1.7 | Client state | Store ephemeral filters/drafts only; canonical evidence remains server-side. [VERIFIED: `app/package.json`; ASSUMED: existing store conventions will be retained] |
| uPlot | 1.6.30 | Existing analysis charts | Retain for bounded time-series visualization; do not introduce another chart dependency during beta. [VERIFIED: lockfile and `GameAnalysisView.vue`] |
| Prometheus Go client | v1.11.0 | Operational and product-path telemetry | Follow single-registration and bounded-label conventions already present. [VERIFIED: vEDH `go.mod`, `server/metrics.go`, `pkg/telemetry/metrics.go`] |

### Supporting

| Library / Facility | Verified Version | Purpose | When to Use |
|---|---:|---|---|
| `github.com/google/uuid` | v1.6.0 | Canonical identifiers and tombstones | Generate entity IDs and non-reversible participant tombstones. [VERIFIED: vEDH `go.mod`] |
| Playwright | manifest ^1.60.0 | End-to-end UI verification | Product-loop, privacy, publication, and debrief scenarios. [VERIFIED: `app/package.json`] |
| Vitest | manifest ^1.1.9; installed 1.6.1 | Vue unit/component tests | Filters, evidence labels, limited-evidence state, form progression, and pseudonym rendering. [VERIFIED: `app/package.json` and local package binary] |
| Rust smoke harness | workspace toolchain nightly; local 1.99.0-nightly | Running API contract smoke tests | Preserve the current `make test-smoke-rust` gate while adding new API flows. [VERIFIED: `testing/`, `Makefile`, local environment] |
| PostgreSQL security-barrier view | PostgreSQL 14 | Sanitized publication contract | Expose materialized, authorized publication versions to jank without base-table grants. [CITED: https://www.postgresql.org/docs/14/sql-createview.html] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|---|---|---|
| Versioned projector tables | Materialized views | PostgreSQL materialized views can refresh concurrently only when populated and backed by a qualifying unique index, and refresh does not preserve ordering. Explicit projector tables better support versioning, checkpointing, shadow rebuilds, and privacy invalidation. [CITED: https://www.postgresql.org/docs/14/sql-refreshmaterializedview.html] |
| In-process beta projector | Message broker / streaming platform | A broker adds infrastructure and delivery semantics before beta volume proves it necessary. Preserve an event/checkpoint boundary so a later worker can be split out. [ASSUMED] |
| Restricted publication views | Direct jank reads of vEDH tables | Direct reads expand the privacy and schema-coupling blast radius and violate least privilege. [VERIFIED: D-07; CITED: https://www.postgresql.org/docs/14/ddl-priv.html] |
| Stable source-ID lineage | Name or list-similarity grouping | Similarity can support a user prompt but is explicitly forbidden as automatic lineage evidence. [VERIFIED: D-09] |

**Installation:** none. The beta plan should not add a runtime dependency merely to implement transactions, role grants, immutable rows, cursor pagination, or versioned projections. [VERIFIED: existing manifests and PostgreSQL capabilities]

## Package Legitimacy Audit

This research recommends no new external package installation, so the package-legitimacy gate is not applicable. Every named package is already pinned in a checked-in vEDH or inspected jank manifest. If a planner later adds a package, it must run the ecosystem-specific legitimacy and registry checks before installation. [VERIFIED: repository manifests]

## Recommended Architecture

### System Architecture Diagram

```text
Browser / Vue
  | authenticated game commands, review drafts/submissions, private queries
  v
vEDH API ------------------------------------------------------+
  | authorize canonical UUID                                  |
  | validate command + assign mutation id                      |
  v                                                            |
SQL transaction                                                |
  +--> update games.payload compatibility/live-state cache     |
  +--> insert canonical runtime_event(s)                       |
  +--> insert participant/result/snapshot corrections          |
  +--> commit --------------------------------------------------+
             |
             v
      versioned projectors / inference registry
             |
             +--> private game/lineage read models --> vEDH API --> Browser
             |
             +--> author selects review/evidence
                         |
                         v
               publication builder
               [hidden-data removal]
               [scoped pseudonyms]
               [source/version pinning]
                         |
                         v
            evidence_api security-barrier views (read only)
                         |
                         v
jank API --> jank-owned tree/revision/annotation/evidence-link tables
  |               | optional discussion thread / fork backlink
  +--------------> jank HTML/JSON UI

PostgreSQL boundary:
  vedh_owner / vedh_migrator / vedh_app  -> vEDH-owned objects
  jank_owner / jank_migrator / jank_app  -> jank schema
  jank_app -> SELECT only on evidence_api contract; no credential/base evidence access
```

The event insertion and live-state update must commit or roll back together. Publication is a separate, explicit transformation from private canonical evidence, and jank references the resulting stable version rather than reaching back into private records. [VERIFIED: D-05, D-07, D-08, D-18; CITED: https://www.postgresql.org/docs/14/tutorial-transactions.html]

### Recommended Project Structure

The following is a planning target, not a directive to reorganize unrelated code. Split by domain while preserving the existing resolver/server conventions. [VERIFIED: current repository layout; ASSUMED: filenames may be adjusted during implementation]

```text
server/
├── identity.go                 # canonical UUID/profile and deletion commands
├── deck_snapshots.go           # immutable capture + normalization
├── deck_lineages.go            # stable-origin assignment + audited changes
├── game_participants.go        # participant and result revisions
├── runtime_events.go           # transactional envelope and corrections
├── runtime_inferences.go       # version/evaluation registry and source links
├── game_reviews.go             # drafts and immutable submissions
├── evidence_publications.go    # privacy transform, versions, revocation
├── game_analysis.go            # authorized timeline/read-model API
└── schema.graphql              # typed public contract
pkg/
├── deckidentity/               # pure card/deck normalization and hashes
├── runtimeevent/               # envelope, validation, ordering, causality
├── inference/                  # deterministic rules + version fixtures
├── projection/                 # checkpoints, rebuilds, read-model versions
└── publication/                # pure sanitization/pseudonymization
migrations/                     # production ordered migrations
server/test/migrations/         # mirrored test migrations
app/src/
├── views/                      # timeline, debrief, lineage, publication
├── components/evidence/        # labels, provenance, uncertainty
├── components/reviews/         # compact form + Add detail
└── stores/                     # ephemeral filters and drafts

jank/app/
├── identity.go                 # vEDH-authenticated principal adapter
├── store/                      # split forum/tree/revision/evidence stores
├── handlers_trees.go           # owner-authorized mutation endpoints
├── handlers_evidence.go        # restricted publication reads/links
└── ...                         # existing forum/moderation surfaces
jank/migrations/                # new ordered PostgreSQL/SQLite migrations
jank/templates/                 # tree revision, fork, evidence provenance UI
```

### Reusable vEDH Patterns and Likely Files

| Existing pattern / file | Reuse | Likely change |
|---|---|---|
| `pkg/deckimport/*` pure normalization with fixtures | Model deck/card normalization and content-hash logic as a pure package with table-driven tests. [VERIFIED: code inspection] | Extend through a new sibling package; do not couple snapshot creation to GraphQL parsing. |
| `server/schema.graphql`, `server/generated.go`, `server/schema.resolvers.go` | Schema-first typed API and gqlgen generation. [VERIFIED: repository files] | Add typed timeline, lineage, review, inference, and publication operations; regenerate `generated.go` and models. |
| Resolver split such as `server/games.go`, `server/decks.go` | Keep authorization at the API boundary and domain operations in named modules. [VERIFIED: code inspection] | Refactor mutation persistence so accepted commands use a shared transaction. |
| `migrations/` plus `server/test/migrations/` | Preserve production/test schema parity. [VERIFIED: repository files] | Every runtime-intelligence migration must be mirrored and exercised by migration tests. |
| `server/metrics.go`, `pkg/telemetry/metrics.go` | Single registration and bounded labels. [VERIFIED: code inspection] | Add outcomes such as event append/reconcile/projector lag without user/game/deck IDs as labels. |
| `server/product_events.go` | Database-backed product funnel signals. [VERIFIED: code inspection] | Add activation, evidence inspection, documented revision, publication, and jank-link events with privacy-safe properties. |
| Feature flags and kill switches in current roadmap patterns | Use for dual-write, projector, new-read, publication, and jank cutover. [VERIFIED: active planning artifacts] | Each switch requires rollback semantics and observability. |
| `app/src/views/GameAnalysisView.vue` | Preserve the completed-game route and progressive enhancement. [VERIFIED: code inspection] | Replace raw JSON parsing and offset loading with typed cursor pages and evidence/provenance components. |
| `server/gamelog.go`, `server/gamelog_diff.go` | Treat as legacy adapters and a source of event naming, not the new canonical contract. [VERIFIED: code inspection] | Route accepted mutations through canonical writer, then keep old GraphQL shape during compatibility window. |
| `server/formats.go` | Retain an envelope-level game-system/format capability boundary. [VERIFIED: code inspection] | Implement Magic beta adapter only; do not select a second game. |

### Likely jank Files at Inspected Commit

| File | Current responsibility | Required planning impact |
|---|---|---|
| `app/models.go` | Forum and tree structs | Replace/augment string creators with canonical UUID principals; add tree/revision/fork/evidence models. [VERIFIED: jank source] |
| `app/store.go` | PostgreSQL/SQLite schema and most persistence | Split domain stores and extract startup DDL into ordered migrations before cutover. [VERIFIED: jank source] |
| `app/auth.go` | Username cookie and manually built HS256 token | Replace with vEDH-owned session/principal verification; jank must neither store credentials nor mint vEDH identity. [VERIFIED: jank source and D-06] |
| `app/handlers_auth.go` | Signup/login/token endpoints | Disable and remove from integrated mode after identity reconciliation; retain no duplicate account path. [VERIFIED: jank source and D-06] |
| `app/handlers_api.go` | Tree CRUD and annotations | Enforce owner/visibility checks, immutable revisions, relationship vocabulary, evidence-link validation, and fork semantics. [VERIFIED: jank source and D-20/D-21] |
| `app/handlers_html.go` | Tree/forum HTML routes | Add publication provenance, revision history, fork attribution, and privacy-safe failure behavior. [VERIFIED: jank source; ASSUMED: server-rendered UI remains] |
| `app/router.go`, `app/app.go` | Routes, middleware, store construction | Inject canonical principal and separate jank/evidence database capabilities. [VERIFIED: jank source] |
| `app/app_test.go` | HTTP/store integration tests | Add PostgreSQL role/authorization and migration/backfill scenarios; retain standalone SQLite compatibility only outside the beta contract. [VERIFIED: jank source; ASSUMED: standalone mode remains useful] |
| `templates/card_tree.html`, `templates/card_trees.html` | Existing tree UI | Render revisions, bounded relation labels, unresolved card state, linked publication/version, and optional thread/fork backlinks. [VERIFIED: jank templates and PRD] |
| `README.md`, `TODO.md` | Deployment and current limitations | Document integrated mode, required DB roles, auth boundary, migration/runbook, and standalone limitations. [VERIFIED: jank repository] |

## Canonical Data and Boundary Design

The table and column names below are planning-level contracts. Plans may adjust exact SQL names, but they should preserve ownership, immutability, version pinning, idempotency, and deletion semantics. [ASSUMED]

### 1. Database roles and schemas

Use role ownership and object grants as the first enforceable boundary:

| Role / schema | Privileges | Required invariant |
|---|---|---|
| `vedh_owner` (`NOLOGIN`) | Owns vEDH tables, sequences, functions, and views | Runtime credentials never own objects. [CITED: https://www.postgresql.org/docs/14/ddl-priv.html] |
| `vedh_migrator` | Temporarily assumes ownership/migration capability | Deployment-only; not present in the running application. [ASSUMED] |
| `vedh_app` | Explicit DML on vEDH runtime objects | No `CREATE`, superuser, `BYPASSRLS`, jank writes, or broad default grants. [CITED: https://www.postgresql.org/docs/14/role-attributes.html] |
| `jank_owner` (`NOLOGIN`) | Owns `jank` schema objects | Cannot own or mutate vEDH evidence. [VERIFIED: D-07; CITED: https://www.postgresql.org/docs/14/sql-createschema.html] |
| `jank_migrator` | Deployment-only DDL in `jank` | Cannot alter vEDH-owned base tables after the canonical-FK migration is installed. [ASSUMED] |
| `jank_app` | Explicit DML in `jank`; `USAGE` and `SELECT` on approved `evidence_api` objects | No access to credentials, private reviews, raw gameplay, hidden cards, base publication sources, or vEDH mutation functions. [VERIFIED: D-07 and D-08] |
| `evidence_api` schema | vEDH-owned, sanitized publication views/functions | Contract is versioned and read-only to jank. [CITED: https://www.postgresql.org/docs/14/sql-createview.html] |

Revoke `CREATE` on the `public` schema from `PUBLIC`, grant schema `USAGE` separately from table rights, and set restrictive default privileges for future objects. PostgreSQL schema access and object access are distinct; table grants alone do not grant schema usage. [CITED: https://www.postgresql.org/docs/14/ddl-schemas.html; CITED: https://www.postgresql.org/docs/14/sql-alterdefaultprivileges.html]

Do not move all existing vEDH objects out of `public` on the beta critical path. Changing owners and explicit grants around the existing objects, while creating new jank and evidence-interface schemas, achieves the locked ownership boundary with less application churn. A later schema relocation can be a separately tested operational migration. [ASSUMED]

Row-level security is not a substitute for API authorization when a connection pool uses one database role: `current_user` identifies the pooled application role, not the end user. If RLS is used as defense in depth, set a transaction-local principal safely, enable `FORCE ROW LEVEL SECURITY` where owners must not bypass, and test default-deny behavior. Superusers, `BYPASSRLS` roles, and normally table owners bypass row policies. [CITED: https://www.postgresql.org/docs/14/ddl-rowsecurity.html]

Any `SECURITY DEFINER` bridge function must set an explicit trusted `search_path` ending in `pg_temp`, revoke default `PUBLIC` execute access in the same transaction, validate every principal and publication/version argument, and grant execution only to the intended role. Prefer views over bridge functions when no parameterized operation is needed. [CITED: https://www.postgresql.org/docs/14/sql-createfunction.html; CITED: https://www.postgresql.org/docs/14/perm-functions.html]

### 2. Canonical users and authentication

The existing `users.uuid VARCHAR UNIQUE` must become a validated UUID primary identity before jank foreign keys or participant retention depend on it. The migration sequence is: audit null/blank/invalid/duplicate values; resolve each exception explicitly; add/backfill a native UUID column if an in-place cast is unsafe; add a validated uniqueness constraint; make it `NOT NULL`; promote it to the canonical key; then migrate all new foreign keys. [VERIFIED: current initial migration; CITED: https://www.postgresql.org/docs/14/sql-altertable.html]

Expose only a safe profile projection to jank if display data is needed. Password hashes, guest credential hashes, reset material, and authentication metadata remain in vEDH-only objects. jank's integrated routes accept an already authenticated canonical UUID principal; they do not call a local account-link table because D-06 forbids one. [VERIFIED: D-06, D-07, current vEDH and jank auth code]

The browser session transport is a required design checkpoint because current vEDH stores an HS256 JWT in browser local storage and current jank uses its own username cookie/JWT. A separate jank origin cannot safely read vEDH local storage. The two acceptable families are: a vEDH-owned HttpOnly session cookie scoped to the deployment's shared registrable domain, or vEDH-issued short-lived identity proof verified through introspection/asymmetric signatures. Do not share the vEDH HS256 minting secret with jank. Exact selection depends on the deployment domains and reverse-proxy topology, which are not in the repository. [VERIFIED: `app/src` auth storage and both servers' auth code; ASSUMED: deployment topology]

### 3. Canonical card object identity prerequisite

Current vEDH lookup paths expose inconsistent card identifiers: the GraphQL `Card` path uses the cards table `id`, search results use `uuid`, and the data also carries `scryfalloracleid`. Deck snapshots, tree nodes, and event payloads need one documented internal canonical card-object key plus a separate printing key and source provenance. [VERIFIED: `server/cards.go`, `server/search.go`, card migrations]

Create a platform-owned stable card-object identity (game-system scoped), a printing/source mapping, and an explicit unresolved record shape. Exact source matching may link legacy names; normalized or fuzzy names must never silently select a canonical card. Preserve the original submitted label and the normalizer version so later resolution does not rewrite a historical snapshot. [VERIFIED: jank PRD and current string card names; ASSUMED: internal surrogate identity is the least-coupled solution]

Current game deck hydration expands quantities into repeated card references and the state maps some cards by a shared card ID. Multiple copies therefore cannot always be distinguished as physical instances. Introduce a game-object/instance identifier where event semantics require source-target identity, distinct from canonical card-object and printing IDs. [VERIFIED: `server/games.go` deck hydration and state mapping]

### 4. Immutable deck snapshots and lineages

| Logical object | Minimum fields / constraints | Notes |
|---|---|---|
| `deck_lineages` | UUID, owner UUID/tombstone, game system, created time, status | The durable analysis identity. It is not inferred from a name or similarity. [VERIFIED: D-09] |
| `deck_snapshots` | UUID, game system/format, immutable normalized payload/hash, normalizer version, capture time, quality | Immutable canonical list at game eligibility time. [VERIFIED: D-04, D-05] |
| `deck_snapshot_cards` | snapshot, canonical card object nullable, original label, quantity, role/zone, printing nullable, resolution status | Unresolved inputs remain visible and auditable rather than being dropped. [VERIFIED: current import behavior drops unresolved cards; ASSUMED: proposed remedy] |
| `deck_snapshot_origins` | snapshot, owner, source type, stable source ID, source revision/URL, captured time | Automatic grouping uses `(owner, source type, stable source ID)` only. [VERIFIED: D-09] |
| `deck_lineage_snapshots` | lineage, snapshot, effective interval/ordinal, assignment reason, actor, audit event | Supports audited move, split, and merge without overwriting history. [VERIFIED: D-09] |
| `deck_lineage_events` | immutable event type, source/target lineages, snapshots, actor, reason, time | Split/merge/move history and rollback explanation. [VERIFIED: D-09] |
| `deck_revision_links` | from/to snapshots, author, reason/review revision, change set | Connects evidence inspection and documented deck revision for D-03. [VERIFIED: D-03 and deck-analysis PRD] |

The content hash should include the game system, format, leaders/commanders, canonical or unresolved card identities, quantities, and normalizer version; it should exclude mutable labels, deck names, and URLs. Use it for immutable-content deduplication, never as automatic lineage proof. [ASSUMED]

### 5. Participants and versioned results

Add an immutable `game_participants` row for each seat with: stable participant UUID, game ID, canonical user UUID or participant tombstone, immutable deck snapshot ID, seat/order, join status, display-name snapshot, and eligibility status/reason. The participant row, not a username string in JSON, becomes the subject and actor reference for events and reviews. [VERIFIED: D-04, D-05, D-13; current game payload uses player fields]

Represent game outcome as versioned result assertions rather than mutating one enum in JSON:

- `game_result_revisions`: result revision UUID, game ID, sequence/version, status, ending turn/round, completion source, condition/reason code, author/system actor, source event, created time, supersedes/corrects revision. [VERIFIED: runtime PRD]
- `game_participant_results`: revision plus participant and outcome `win`, `loss`, `draw`, `no_contest`, `abandoned`, or `incomplete`, with placement only when explicitly known. [VERIFIED: runtime PRD; current model only exposes `WIN`/`DRAW`]
- A game may point to the currently accepted result revision for efficient reads, but old revisions remain immutable. Corrections append and repoint; they do not update evidence in place. [VERIFIED: D-05]

### 6. Canonical runtime events and corrections

The canonical event envelope should include at least:

```text
runtime_event
  id UUID
  game_id TEXT / canonical game key
  event_sequence BIGINT              -- server assigned per game
  mutation_id UUID                   -- client idempotency scope
  mutation_event_ordinal INTEGER     -- supports multiple events per command
  event_type + schema_version
  game_system + adapter_version
  actor_participant_id nullable
  subject_participant_id nullable
  origin                             -- server_observed, client_asserted,
                                     -- server_derived_client_state, backfilled
  visibility                         -- private/owner-visible/publication-eligible
  turn/round/phase nullable
  client_occurred_at nullable
  server_recorded_at
  payload JSONB                      -- versioned, type-validated details
  corrects_event_id nullable
  legacy_gamelog_id nullable unique
```

Enforce unique `(game_id, event_sequence)` for deterministic order and unique `(game_id, mutation_id, mutation_event_ordinal)` for retry safety. A B-tree beginning `(game_id, event_sequence, id)` supports ordered, cursor-based timeline pages; PostgreSQL can satisfy ordered limited queries directly from a matching B-tree. [ASSUMED: exact constraint shape; CITED: https://www.postgresql.org/docs/14/indexes-ordering.html]

Accepted mutations must run in one SQL transaction that locks or otherwise serializes the game sequence, validates the command, updates `games.payload`, inserts all canonical events/results/corrections, and commits before publishing product metrics or asynchronous work. Current `logEvent` errors are swallowed and occur outside a shared transaction in some paths, so merely wrapping the existing logger is insufficient. [VERIFIED: `server/gamelog.go` and `server/games.go`; CITED: https://www.postgresql.org/docs/14/tutorial-transactions.html]

Require a client mutation UUID for beta-eligible commands. During a bounded compatibility window, legacy clients may use a server-generated identifier but their retries cannot be proven exactly-once and should be marked in origin/quality metadata. Keep the existing GameLogs GraphQL field as a legacy adapter while adding typed cursor-paged timeline queries; announce and measure deprecation before removal. [ASSUMED]

The existing diff logger compares client-submitted board snapshots after a server request. Classify those rows as `server_derived_client_state`, not purely `server_observed`. A correction is another append-only event referring to the corrected record; analytical projectors resolve the effective view while preserving both. [VERIFIED: `server/gamelog_diff.go`; VERIFIED: D-05 and D-14]

### 7. Inference and projection registry

Use additive records with explicit provenance:

| Object | Required content | Invariant |
|---|---|---|
| `inference_versions` | category, algorithm/ruleset version, source commit/build, status, deterministic flag | A version is immutable once used. [VERIFIED: D-15/D-16] |
| `inference_evaluations` | version, fixture version, labeled item counts, true/false positives, precision, evaluation time | Default visibility for probabilistic output derives from stored evaluation evidence, not a manual boolean. [VERIFIED: D-16] |
| `runtime_inferences` | game/lineage subject, category, output payload, confidence, causal classification, version, visibility, invalidated/superseded state | Never replaces observed or derived source evidence. [VERIFIED: D-14 through D-17] |
| `runtime_inference_sources` | inference, event/snapshot/result/review source type and ID, role/order | Every displayed inference can show source provenance. [VERIFIED: D-16/D-18] |
| `projection_checkpoints` | projector name/version, last event key, state, time | Enables restart and lag monitoring. [ASSUMED] |
| versioned read-model tables | projector version plus exact source range/watermark | Rebuild into a shadow version, compare, then switch the active pointer. [ASSUMED] |

Deterministic rules display by default. A probabilistic category can default-display only if the exact version reaches at least 95% precision on the versioned expert holdout fixture; otherwise the API returns it only when `showSpeculative=true` and includes confidence, source IDs, and version. An explicit source-target runtime event may support a causal fact; any other causal interpretation is a hypothesis. [VERIFIED: D-16 and D-17]

Run projectors in-process or as a separately invoked Go worker for beta, with idempotent upserts keyed by canonical source/version. Do not add a message broker to the beta path. Keep a clean checkpoint interface so worker extraction remains possible. [ASSUMED]

### 8. Reviews and immutable revisions

Use one logical review per `(game, author)` with mutable draft state and append-only submitted revisions:

| Object | Responsibility | Invariant |
|---|---|---|
| `game_reviews` | Stable review identity, game, author, current submitted revision, draft status | One author perspective; private by default. [VERIFIED: debrief PRD] |
| `game_review_drafts` | Mutable contributors (max 3), underperformer (max 1), notes, optional next action, tags | Drafts never affect analysis or publication. [VERIFIED: D-11/D-12] |
| `game_review_revisions` | Immutable snapshot of every submission with ordinal and source game/result version | Editing after submission creates a new row. [VERIFIED: D-12] |
| revision child rows | Contributor assessments, underperformer, tags, next action, structured details | Apply limits in domain validation and database constraints where practical. [VERIFIED: D-11] |

The API should return a compact default form and lazily reveal tags/notes under `Add detail`. Submitting writes a revision and advances the review pointer in one transaction. Publications store the exact review revision ID and remain unchanged until an explicit republish command creates a new publication version. [VERIFIED: D-11, D-12]

### 9. Publication and privacy boundary

Separate publication identity from publication versions:

- `evidence_publications`: stable author-owned ID, kind, lifecycle (`draft`/`published`/`revoked`), current version, revocation metadata. [VERIFIED: runtime and analysis PRDs]
- `evidence_publication_versions`: immutable sanitized payload, exact sample denominator, uncertainty statement, source calculation/projector version, created time, supersedes version. [VERIFIED: D-08/D-18]
- `evidence_publication_sources`: exact review revision, snapshot, lineage, event/inference IDs and roles. Probabilistic records require explicit author selection. [VERIFIED: D-12/D-18]
- `publication_pseudonyms`: publication/version-scoped mapping from participant to a generated pseudonym; mappings are inaccessible to jank. [VERIFIED: D-08]

Build the sanitized payload inside vEDH after author authorization. Remove hidden information, free-text fields not explicitly selected, stable opponent identifiers, raw event payloads, and fields not in the publication schema. Replace opponents with publication-scoped pseudonyms. Store the materialized sanitized version so jank never executes privacy logic against canonical source rows. [VERIFIED: D-08; ASSUMED: materialization provides the safest boundary]

Expose stable publication/version IDs, evidence class, visible card/snapshot facts, exact `n`, uncertainty, inference confidence/version/source summaries when selected, revocation state, and author-approved text through an `evidence_api` security-barrier view. A revoked publication leaves a tombstone sufficient for jank to explain an unavailable link; it no longer exposes its payload. [VERIFIED: D-08/D-13/D-18; CITED: https://www.postgresql.org/docs/14/sql-createview.html]

### 10. jank trees, revisions, and evidence links

Preserve current tree URLs/IDs where possible, then add:

| Object | Required shape | Compatibility rule |
|---|---|---|
| `jank.card_trees` | existing ID/title plus owner UUID, visibility, optional linked thread, current revision, fork source, source calculation/snapshot metadata | Existing scope fields remain nullable compatibility metadata; a tree can stand alone. [VERIFIED: current jank schema and D-20] |
| `jank.card_tree_revisions` | tree, immutable ordinal, author UUID/tombstone, message, source revision, created time | Every submitted edit pins a full immutable tree version or deterministic delta with rebuild guarantee. [VERIFIED: jank PRD] |
| revision nodes | stable node UUID, canonical card object nullable, original label/display snapshot, parent/position, relation kind/custom label | Ambiguous legacy names remain unresolved; no fuzzy auto-link. [VERIFIED: current string model; ASSUMED: proposed migration] |
| revision annotations | node, evidence class, bounded kind/label/tags/body, author | Community hypothesis is distinct from vEDH observed/derived/inferred evidence. [VERIFIED: D-14] |
| `jank.evidence_links` | owning tree/revision/node/annotation FK, publication ID/version, link kind, added by/time | Use explicit foreign-key columns/check constraints, not an unchecked polymorphic target string. [ASSUMED] |
| fork backlink | fork tree, source tree/revision, attribution, actor/time | A fork is independent and never overwrites its source. [VERIFIED: D-20] |

Enforce tree owner or authorized collaborator checks on every mutation before writing. Beta has individual ownership only; do not generalize this into team roles. Relationship types are the locked nine-value vocabulary plus `custom`; custom labels require length/character bounds and are indexed for search but excluded from global relation filters. [VERIFIED: D-19 through D-21]

## Migration and Backfill Order

Use additive migrations, online validation where production size warrants it, and explicit reconciliation gates. PostgreSQL can add foreign/check constraints as `NOT VALID`, enforce them for new writes, and validate existing rows later; a validated check can support a later `SET NOT NULL` without a full verification scan. A unique index can be built concurrently and attached as a constraint, but concurrent index creation cannot run inside a transaction and a failed build can leave an invalid index requiring cleanup. [CITED: https://www.postgresql.org/docs/14/sql-altertable.html; CITED: https://www.postgresql.org/docs/14/sql-createindex.html]

### Ordered migration ledger

| Order | Migration unit | Required gate before next unit |
|---:|---|---|
| 0 | **Release handoff and inventory:** complete active Phase 5; snapshot production versions, row counts, null/duplicate profiles, event types, deployments, backups, DB roles, jank database, domains, and rollback owner. | Signed inventory and restore rehearsal; no runtime-intelligence DDL before active release gate. [VERIFIED: milestone boundary; ASSUMED: operational sign-off format] |
| 1 | **Role boundary:** create owner/migrator/runtime roles; create `jank` and `evidence_api`; change or verify ownership; revoke broad `public` schema creation and default privileges; grant explicit existing-table rights. | Automated positive/negative privilege matrix passes under each runtime role. [VERIFIED: D-07; CITED: PostgreSQL privilege docs] |
| 2 | **Canonical users:** audit and normalize vEDH UUIDs; promote native UUID key; add safe profile projection; add deletion lifecycle fields/worker contract. | Zero null/invalid/duplicate canonical UUIDs; credential table invisible to `jank_app`. [VERIFIED: D-06/D-13] |
| 3 | **jank user reconciliation staging:** produce a one-time administrative mapping report from every jank integer/string identity to canonical UUID or unresolved tombstone; do not create an application linking table. | Every authored row classified; conflicts require human resolution; username equality alone is not accepted as proof. [VERIFIED: current jank schema and D-06; ASSUMED: administrative mapping artifact is allowed because it is not runtime state] |
| 4 | **Card identity:** create canonical card-object/printing mapping and unresolved representation; backfill exact identifiers; add game-object instance IDs for new runtime events. | Exact/ambiguous/unresolved counts reconciled; no fuzzy automatic decisions. [VERIFIED: current ID inconsistency and jank string cards] |
| 5 | **Snapshots and lineages:** add immutable snapshot/origin/lineage/audit tables; start snapshot dual-write for new eligible games. | Content hash repeatability fixtures pass; only stable owner-scoped origins auto-group. [VERIFIED: D-04/D-09] |
| 6 | **Participants and results:** add participant rows and result revisions; dual-write join/completion; backfill games conservatively. | Each beta-eligible completed game has valid participants, snapshot, and result revision; exclusions have reason codes. [VERIFIED: D-04/D-05] |
| 7 | **Runtime events:** create envelope/correction tables; add client mutation IDs; transactionally dual-write live state plus events; retain legacy GameLogs reads. | Fault-injection proves no live-state/event split; retry tests prove idempotency; reconciliation is zero for eligible traffic. [VERIFIED: current split-write problem; ASSUMED: client mutation contract] |
| 8 | **Gamelog backfill:** copy one legacy row to one canonical source record, classify origin/quality, link only unambiguous participants/card objects, retain raw legacy payload privately. | Source/target row counts and checksums match; no invented facts; invalid rows quarantined with reason. [VERIFIED: current gamelog design; ASSUMED: exact reconciliation method] |
| 9 | **Inference/projectors:** seed version registry, deterministic rules, evaluation fixtures, checkpoints, and shadow read models; compare to canonical records. | Rebuild is idempotent; probabilistic default visibility is computed from stored ≥95% precision evidence. [VERIFIED: D-14 through D-17] |
| 10 | **Reviews:** create draft/revision tables and APIs; add private UI; no analytical contribution from drafts. | Revision immutability, limits, ownership, and deletion behavior pass. [VERIFIED: D-11/D-12] |
| 11 | **Lineage analytics:** build versioned private projections and comparison UI over eligible facts, events, results, and submitted review revisions. | `n=1` visible; either side below 3 shows `Limited evidence`, exact denominators, and uncertainty. [VERIFIED: D-10] |
| 12 | **Publications:** create materialized sanitized versions, pseudonyms, sources, revocation, and restricted `evidence_api` contract. | Privacy red-team fixtures show no hidden/opponent/private leakage; pinning and republish semantics pass. [VERIFIED: D-08/D-12/D-18] |
| 13 | **jank content migration:** add canonical UUID creator/owner fields, revision tables, relationship vocabulary, evidence links, visibility, and forks; backfill legacy content; enforce owner auth. | Every legacy authored row mapped/tombstoned; every tree has revision 1; authorization negatives pass. [VERIFIED: jank source and D-20/D-21] |
| 14 | **jank auth cutover:** route integrated login/session to vEDH principal; disable jank signup/login/token minting; remove runtime access to legacy credentials; switch PostgreSQL DSN to `jank_app`. | No duplicate account creation path; jank cannot mint vEDH identity or query private/base evidence. [VERIFIED: D-06/D-07] |
| 15 | **Read switch and contract:** enable new timeline/analysis/publication/tree reads by cohort; monitor; retire dual-read and legacy writes only after a full compatibility window and backup. | Stable beta cohort, zero reconciliation drift, rollback drill, support/runbook approval. [ASSUMED] |

### Existing games and gamelog backfill rules

1. Copy every legacy `gamelog` row exactly once using a unique `legacy_gamelog_id`; preserve original JSON and timestamp, mark missing time explicitly, and assign `origin=backfilled` plus a more specific derivation classification when provable. Never reinterpret absence as an event. [VERIFIED: current schema; ASSUMED: backfill envelope]
2. Map actor usernames to a participant UUID only when the game participant set yields one exact match. Multiple or missing matches stay unresolved with a reason; do not use a global username guess. [VERIFIED: current event actor strings and username fallback]
3. Label diff-derived historical events as derived from client state. Preserve legacy event type and payload even if the new typed adapter cannot parse it. [VERIFIED: `server/gamelog_diff.go`]
4. Backfill a deck snapshot as `exact`, `partial`, or `unknown`. Historical pre-game deck text is not stored as a standalone immutable list. A terminal payload may contain cards across zones, but duplicated references, created/copied objects, tokens, transformations, and unresolved card IDs can make reconstruction incomplete; do not declare an exact snapshot without a fixture-proven inventory algorithm. [VERIFIED: current game/deck model; ASSUMED: listed fidelity classifications]
5. Backfill only explicit `WIN`/`DRAW` facts from current result data. Derive a loss for another participant only when the complete participant set and result semantics are unambiguous. Otherwise use `incomplete` or an unresolved migration status; never manufacture `no_contest`/`abandoned`. [VERIFIED: current result enum; ASSUMED: conservative mapping]
6. Retain all historical games, but count a game toward beta activation only when it has a valid canonical participant, immutable snapshot, and eligible completion/result under the new contract. [VERIFIED: D-04]

### Legacy jank backfill rules

1. Add nullable canonical UUID creator/owner columns and immutable display snapshots before changing reads. Backfill by an operator-reviewed export that considers integer user rows, thread/post author strings, tree/node/annotation creator strings, and moderation records. [VERIFIED: jank schema]
2. Matching usernames are candidates, not identity proof. Exact conflicts, changed names, duplicate names, and missing users block automatic assignment. Resolve them administratively or use a non-reversible content-author tombstone. Do not persist a new synchronization/linking table. [VERIFIED: D-06; ASSUMED: operator reconciliation process]
3. A jank-only user may become canonical only through a vEDH-owned account provisioning/recovery flow. Legacy jank password material must not be copied into a second credential system; require reset or a one-time vEDH-owned credential upgrade. [VERIFIED: D-06; current jank credential scheme]
4. Convert every legacy tree into revision 1 without changing its public identity or hierarchy. Preserve string card names and resolve only exact canonical mappings; ambiguous/unmatched names enter an unresolved queue. [VERIFIED: jank current model and PRD]
5. Keep current board/thread/post scope metadata for compatibility, but allow it to be null for standalone trees. Optional linked threads and fork backlinks are separate relationships. [VERIFIED: current jank model and D-20]
6. Reconcile source and target counts by table and creator classification, validate new foreign keys, switch reads to UUIDs, then disable local auth and retire credential access. Do not drop legacy identity data until the recovery window and rollback gate have passed. [ASSUMED]

## Compatibility and Rollback Strategy

### Expand → migrate → contract

```text
add nullable/new objects
  -> dual-write canonical + legacy
  -> backfill with explicit quality
  -> reconcile counts/checksums/invariants
  -> shadow projectors and compare
  -> cohort-gated new reads
  -> stop legacy writes
  -> observe through rollback window
  -> contract only after signed gate
```

Rollback means disabling new writes/reads/projectors/publication by feature flag and returning to the legacy read surface while retaining already-written canonical data. Do not write destructive down migrations for canonical evidence or rewrite immutable rows during rollback. Fix forward with additive correction/version records. [VERIFIED: D-05; ASSUMED: feature-flag naming]

### Required compatibility surfaces

- Keep `games.payload` as the live game-state cache through beta so existing game clients do not require a synchronized rewrite. Canonical tables become the evidence authority for analytics. [ASSUMED]
- Keep the existing GameLogs query temporarily and adapt canonical events to its raw shape where possible; add a new typed cursor query for the analysis UI. Measure use before removal. [VERIFIED: current GraphQL contract; ASSUMED: deprecation duration]
- Accept old clients during a bounded transition, but exclude commands without provable idempotency/snapshot eligibility from beta metrics where necessary. [ASSUMED]
- Preserve existing jank tree/thread URLs and integer public IDs if routing depends on them; add UUID ownership/revision identity alongside them. [VERIFIED: current jank routes; ASSUMED: URL preservation]
- Retain jank SQLite only for standalone development/test mode with mocked or absent vEDH evidence. The integrated beta contract is PostgreSQL because D-06 requires one shared database. [VERIFIED: D-06 and current optional SQLite support]
- Run all backfills resumably with a checkpoint, deterministic selection, bounded batches, dry-run counts, and a quarantine/error ledger. [ASSUMED]

### Account deletion and retained facts

At deletion request, mark the account recoverable for 30 days and block or define new publication behavior. At expiry, revoke all publications, delete private drafts/reviews/revisions and authentication/profile data, and replace retained participant/author references in canonical multiplayer facts with fresh non-reversible tombstone identifiers. Do not retain a lookup from tombstone to deleted UUID. Snapshot ownership references also need deletion classification: if required to retain a game fact, tombstone the owner; otherwise delete owner-private orphan data. [VERIFIED: D-13; ASSUMED: snapshot classification]

Backups, replicas, search indexes, analytics exports, caches, and logs must be included in the deletion runbook. The PRDs fix the application behavior but do not specify backup expiration or restoration re-deletion mechanics; the implementation plan needs an explicit operational policy and a test that restored data re-applies deletion tombstones/revocations. [ASSUMED]

### Revision 2 durable cross-process deletion bridge

The executable plan set resolves D-13 with a durable outbox, not a shared Go transaction. RI-08-01 commits the vEDH-owned deletion changes and one unique jank command in the same vEDH PostgreSQL transaction. RI-08-02 runs a dedicated least-privilege jank consumer against the verified checkout; that worker leases the command, invokes a jank-owned tombstone function, stores a generation receipt, invokes the vEDH ack that scrubs the pending subject UUID, and commits those jank/ack changes in a separate PostgreSQL transaction. A job is not globally complete until the jank and external-cleanup receipts exist. Lease expiry, retries, response loss, role negatives, receipt-lag monitoring, and restore-generation replay are part of the contract. [PLANNING DECISION: checker revision 2; consistent with D-07/D-13 and one shared PostgreSQL database]

Restore matching uses a non-reversible salted fingerprint registry separate from participant/content tombstones. Restored candidate UUIDs can be checked and re-deleted without preserving a tombstone-to-user lookup. The outbox subject UUID exists only while delivery is pending and is nulled by the acknowledged jank transaction. [PLANNING DECISION: checker revision 2]

### Revision 3 integration-worktree and executable-test protocol

RI-00 creates or validates a dedicated jank integration branch/worktree. Its contract keeps remote and base commit immutable while expected_head advances through one audited, atomic, unpushed jank commit per edit plan; read-only jank plans record verification without advancing. Pre-plan checks require remote/base ancestry/exact expected HEAD/cleanliness, task checks allow only declared dirt, and post-plan checks require the advanced expected HEAD and a clean tree. This replaces the contradictory requirement that an edited checkout remain equal to the original pin and clean throughout execution. Test-first RED tasks use a named harness that succeeds only on an allowlisted failing test plus exact plan/task sentinel and rejects compilation, setup, environment, panic, timeout, race, no-test, or unrelated failures. [PLANNING DECISION: checker revision 3]

### Revision 2 executable decomposition

The staged package contains 36 executable plans across 33 dependency waves: RI-00 (1), RI-01 (6), RI-02 (3), RI-03 (4), RI-04 (3), RI-05 (3), RI-06 (5), RI-07 (5), and RI-08 (6). W4 parallelizes generated identity cutover, FK-grant contract, and card identity after UUID/auth; W21 parallelizes evidence_api grants and the publication deletion hook. All other work is serialized by data, migration, generated-file, or approval dependency. RI-08 order is deletion outbox -> verified-jank receipt -> final smoke -> measurement/readiness APPROVED -> cohort activation -> six-week operations. [PLANNING DECISION: checker revision 2]

## Recommended Post-v1 Phase Split

The active Deck-to-Game Activation milestone remains authoritative. Use `RI-*` as staging labels only; the future roadmap may assign final phase numbers after the current Phase 5 release gate. [VERIFIED: `00-CONTEXT.md`]

| Phase | Goal | Depends on | Exit artifact / gate |
|---|---|---|---|
| **RI-0 — v1 release handoff and runtime inventory** | Finish the active release gate, freeze compatible baselines, audit production/jank state, topology, roles, backups, and data quality. | Active Phase 5 | Signed inventory, restore rehearsal, migration owners, exact jank commit/fork strategy, and no unresolved release blocker. [VERIFIED: milestone boundary; ASSUMED: sign-off mechanism] |
| **RI-1 — Shared database, identity, and card foundation** | Enforce role ownership; promote canonical UUID; define deletion lifecycle and shared auth design; establish canonical card/printing/game-instance identity; stage jank identity reconciliation. | RI-0 | Privilege tests, canonical UUID/card mappings, auth transport ADR, deletion runbook, unresolved identity/card queues. [VERIFIED: D-06, D-07, D-13] |
| **RI-2 — Immutable game context** | Add deck snapshots, stable-origin lineages, participants, result revisions, eligibility, and conservative game backfill. | RI-1 | Every new eligible game pins a participant and snapshot; audited lineage commands; reconciled historical quality classifications. [VERIFIED: D-04, D-05, D-09] |
| **RI-3 — Canonical runtime events and inference** | Transactional event envelope, idempotency/corrections, typed timeline, deterministic inference, evaluation/version registry, generic envelope capability. | RI-2 | Fault-injection and retry guarantees; private game evidence UI; inference provenance/visibility gates; legacy query compatibility. [VERIFIED: D-05, D-14 through D-17, D-22] |
| **RI-4 — Private post-game debrief** | Optional compact draft flow, immutable submissions, note-to-event links/search, tags, open/resolved next action, review deletion behavior. | RI-3 | Review ownership/limits/revisions pass; drafts excluded from analytics; completed-game path uses the typed timeline. [VERIFIED: D-11/D-12] |
| **RI-5 — Private lineage and deck analytics** | Versioned read models, filters, comparisons, evidence labels, denominators/uncertainty, and documented deck revision loop. | RI-3 and RI-4 | `n=1` views, Limited evidence behavior, lineage comparison, review-to-revision linkage, projector rebuild gate. [VERIFIED: D-01, D-03, D-10, D-14] |
| **RI-6 — Privacy-safe evidence publication** | Author-owned publication versions, source pinning, hidden-data removal, scoped pseudonyms, explicit probabilistic selection, republish/revoke/delete. | RI-3, RI-4, RI-5 | Privacy/adversarial suite, stable read-only evidence contract, revocation tombstones, deletion workflow. [VERIFIED: D-08, D-12, D-13, D-18] |
| **RI-7 — jank shared-identity cutover and evidence trees** | Versioned migrations, canonical principal, owner auth, standalone tree revisions, typed relations, forks, optional threads, stable evidence links. | RI-1 and RI-6 | No local account path, no unauthorized mutations/base reads, migrated revision 1 for all legacy trees, evidence-link/revocation UI. [VERIFIED: D-06, D-07, D-20/D-21] |
| **RI-8 — Six-week serious-tester beta** | Cohort enablement, support, telemetry, weekly evidence loop, incident/rollback operations, north-star evaluation. | RI-3 through RI-7 | Thirty activated serious testers enrolled/activated, six-week run completed, north-star and quality/privacy gates reported. [VERIFIED: D-03/D-04] |
| **RI-9 — Post-beta testing groups (v1.1)** | Owner/admin/member groups and explicit artifact sharing. | RI-8 | Separate post-beta plan; not a beta dependency. [VERIFIED: D-19] |
| **RI-10 — Post-beta design-partner adapter** | Select and implement the first non-Magic adapter with a real design partner. | RI-8 | Adapter decision and contract validation; not a beta dependency. [VERIFIED: D-22] |

### Dependency graph

```text
Active Deck-to-Game Activation Phase 5
                 |
                RI-0
                 |
                RI-1
                 |
                RI-2
                 |
                RI-3
                 |
                RI-4
                 |
                RI-5
                  |
                RI-6
                  |
                RI-7
                  |
                RI-8 (six-week beta)
                 / \
              RI-9 RI-10   [post-beta only]
```

RI-4 is serialized behind RI-3 because the two phases share generated GraphQL/UI/migration-test surfaces and PGR-3 note links consume RI-3's typed timeline. RI-5 consumes both machine evidence and submitted player assessment. RI-6 must see all private source shapes before its allowlist/privacy transform is frozen. RI-7 needs RI-6's stable evidence contract rather than coupling to private schemas. [VERIFIED: PRD data dependencies and checker file-ownership audit]

### Six-week beta operations plan

| Time | Operational focus | Entry / exit signal |
|---|---|---|
| Pre-beta | Restore rehearsal, migration/cutover rehearsal on production-shaped copy, privacy and role test, support runbook, cohort feature flags, baseline dashboards. | Zero critical security/privacy defects; rollback timed; 30 candidates identified. [ASSUMED] |
| Week 1 | Activate serious testers and observe first eligible game/snapshot/participant path. | Enrollment-to-activation denominator visible; failed eligibility reasons actionable. [VERIFIED: D-04; ASSUMED: weekly gate] |
| Week 2 | Evidence timeline quality and correction support; triage missing/ambiguous events. | Event reconciliation stable; no silent append failures; correction workflow exercised. [ASSUMED] |
| Week 3 | Debrief adoption and private lineage comparison. | Draft/submission/revision funnel visible; no draft leakage into analytics. [VERIFIED: D-11/D-12] |
| Week 4 | Documented deck revisions and `Limited evidence` comprehension. | Three-game lineage paths begin completing; denominators/uncertainty usability reviewed. [VERIFIED: D-03/D-10] |
| Week 5 | Opt-in publication and jank evidence-tree loop. | Publication privacy sampling and revocation drills pass; tree links remain stable. [VERIFIED: D-08/D-18/D-20] |
| Week 6 | Repeat loop, measure rolling 14-day north star, close data-quality/support issues, decide post-beta priorities. | Report numerator/denominator, inference precision versions, deletions/revocations, incidents, unresolved migrations, and go/no-go. [VERIFIED: D-03/D-04] |

Do not wait until Week 6 to compute the north star. Record each stage (`eligible game`, `evidence inspected`, `revision documented`) with canonical UUID, lineage, source version, and privacy-safe timestamps, then compute the rolling 14-day cohort continuously. Metrics and logs must not contain review notes, hidden cards, raw event payloads, or stable opponent identifiers. [VERIFIED: D-03; ASSUMED: telemetry implementation]

## PRD Story → Primary Phase Matrix

Every story is assigned one primary implementation phase. A dependency may prepare data or security foundations, but should not duplicate the story's acceptance criteria in another phase plan. [VERIFIED: four canonical PRDs]

| Story | Primary phase | Research support / acceptance focus |
|---|---|---|
| RGI-1 — Capture an immutable game context | RI-2 | Exact snapshot, stable-source lineage, participant identity. [VERIFIED: runtime PRD] |
| RGI-2 — Record normalized runtime events | RI-3 | Transactional typed append, idempotency, origin, correction. [VERIFIED: runtime PRD] |
| RGI-3 — Finish games with complete result semantics | RI-2 | Result revisions and per-participant outcome vocabulary. [VERIFIED: runtime PRD] |
| RGI-4 — Enrich observed events without rewriting them | RI-3 | Version/evaluation/source registry, additivity, quality and causality rules. [VERIFIED: runtime PRD] |
| RGI-5 — Inspect a game as evidence | RI-3 | Authorized typed cursor timeline, filters, provenance, source inspection. [VERIFIED: runtime PRD] |
| RGI-6 — Reuse gameplay evidence in jank | RI-6 | Materialized privacy-safe publication and revocation contract. [VERIFIED: runtime PRD] |
| RGI-7 — Support additional TCGs without flattening their rules | RI-3 | Generic envelope/capability contract with Magic-only beta implementation. [VERIFIED: runtime PRD and D-22] |
| PGR-1 — Start from automatically recorded facts | RI-4 | Private optional entry using canonical completed-game facts. [VERIFIED: debrief PRD] |
| PGR-2 — Identify meaningful card contributions | RI-4 | Up to three contributors and one underperformer from the exact snapshot. [VERIFIED: debrief PRD] |
| PGR-3 — Capture the lesson in the player's own words | RI-4 | Revisioned notes, timeline/event links, and owner-private search. [VERIFIED: debrief PRD] |
| PGR-4 — Classify lessons without flattening them | RI-4 | Versioned controlled tags behind `Add detail`. [VERIFIED: debrief PRD] |
| PGR-5 — Decide what to do next | RI-4 | Typed hypothesis, successor-snapshot link, open/resolved filtering. [VERIFIED: debrief PRD] |
| PGR-6 — Keep reflection private or share it deliberately | RI-6 | Exact-content preview for participant sharing and sanitized public publication. [VERIFIED: debrief PRD] |
| PGR-7 — Skip without friction | RI-4 | One-action dismissal, no blocked flow, one reminder maximum. [VERIFIED: debrief PRD] |
| RDA-1 — Analyze a deck lineage without rewriting history | RI-5 | User-facing lineage/snapshot analysis over RI-2 immutable sources. [VERIFIED: deck-analysis PRD] |
| RDA-2 — Understand outcomes and game pace | RI-5 | Outcome/turn distributions, first-game visibility, exact denominators. [VERIFIED: deck-analysis PRD] |
| RDA-3 — Inspect runtime card activity | RI-5 | Evidence-classed card activity with source drill-down. [VERIFIED: deck-analysis PRD] |
| RDA-4 — Compare deck revisions | RI-5 | Source-pinned before/after cohorts and uncertainty. [VERIFIED: deck-analysis PRD] |
| RDA-5 — Filter by runtime context | RI-5 | Typed bounded filters and explicit unknown/exclusion counts. [VERIFIED: deck-analysis PRD] |
| RDA-6 — Trace every conclusion to evidence | RI-5 | Inspect-games links plus export with definitions, filters, versions, and stable source IDs. [VERIFIED: deck-analysis PRD] |
| RDA-7 — Publish a bounded analysis to jank | RI-6 | Source-pinned sanitized payload with exact `n`, uncertainty, and selected inference metadata. [VERIFIED: deck-analysis PRD] |
| JCT-1 — Create a card tree from a real deck version | RI-7 | Owner-authorized private snapshot/lineage import and published-analysis import. [VERIFIED: jank PRD] |
| JCT-2 — Express card relationships, not just indentation | RI-7 | Locked vocabulary plus bounded searchable custom labels. [VERIFIED: jank PRD] |
| JCT-3 — Attach runtime evidence to a card or claim | RI-7 | Stable published evidence links, previews, n=1, revocation. [VERIFIED: jank PRD] |
| JCT-4 — See runtime context directly on the tree | RI-7 | Denominators, evidence/inference versions, confidence, Limited evidence. [VERIFIED: jank PRD] |
| JCT-5 — Discuss and quote evidence in a thread | RI-7 | Lazy optional thread, stable quote references, independent lifecycles. [VERIFIED: jank PRD] |
| JCT-6 — Compare and fork analytical trees | RI-7 | Immutable revisions/diffs, attributed standalone forks and backlinks. [VERIFIED: jank PRD] |
| JCT-7 — Use one vEDH identity across gameplay and discussion | RI-7 | Canonical UUID cutover, no local account, guest-claim relationship preservation. [VERIFIED: jank PRD] |

## Locked Decision → Phase Matrix

| Decision | Primary enforcement phase | Cross-cutting verification |
|---|---|---|
| D-01 product loop | RI-8 | RI-3 through RI-7 each expose one traceable loop transition. |
| D-02 no spreadsheet data | RI-0 | All plans and fixtures attest that no spreadsheet records enter the system. |
| D-03 north star | RI-8 | RI-2/RI-3/RI-5 emit the three canonical stage facts. |
| D-04 six-week beta/activation | RI-8 | RI-2 defines eligible participant + immutable snapshot. |
| D-05 canonical evidence | RI-3 | RI-2 participants/snapshots/results and all backfills use immutable/versioned records. |
| D-06 exact shared UUID/no duplicate account | RI-1 | RI-7 removes jank auth/account paths. |
| D-07 application ownership | RI-1 | Every later migration includes role-negative tests. |
| D-08 publication privacy | RI-6 | RI-7 reads only sanitized publication views. |
| D-09 stable-source lineage only | RI-2 | RI-5 never auto-groups by similarity/name. |
| D-10 first-game/limited evidence | RI-5 | RI-8 usability measurement. |
| D-11 compact optional debrief | RI-4 | RI-8 funnel measurement does not make it mandatory. |
| D-12 immutable reviews/pinned publication | RI-4 | RI-6 explicit republish contract. |
| D-13 deletion/revocation/tombstone | RI-8 | RI-1 request/recovery primitives, RI-4/RI-6/jank domain hooks, RI-8 `ExecuteExpiredAccountDeletion` orchestration and drill. |
| D-14 evidence classes additive | RI-3 | RI-4 player assessment, RI-5 display, RI-7 community hypothesis. |
| D-15 runtime enrichment | RI-3 | RI-8 quality monitoring. |
| D-16 probabilistic precision gate | RI-3 | RI-6 preserves selection/confidence/version; RI-8 reports versions. |
| D-17 causality | RI-3 | RI-5/RI-6 labels carry exact classification. |
| D-18 `n=1` publication/explicit probabilistic selection | RI-6 | RI-5 supplies denominators/uncertainty. |
| D-19 teams post-beta | RI-9 | RI-1 through RI-8 avoid team-role abstractions in product scope. |
| D-20 standalone trees/forks | RI-7 | RI-8 end-to-end use. |
| D-21 relationship vocabulary | RI-7 | Search/filter tests enforce custom-label behavior. |
| D-22 adapter post-beta | RI-10 | RI-3 creates only a generic boundary and Magic implementation. |

## Validation Architecture

Nyquist validation is enabled because `.planning/config.json` does not set `workflow.nyquist_validation` to `false`. [VERIFIED: `.planning/config.json`]

### Existing test framework

| Property | Value |
|---|---|
| Go unit | `go test ./pkg/...`; quick wrapper `make test-unit`. [VERIFIED: `Makefile`] |
| Go/API integration | Existing server test harness plus PostgreSQL and `All Printings.json`; wrapper `make test-api`. [VERIFIED: `Makefile`, `server/test`] |
| Vue unit/component | Vitest via `cd app && npm test`. [VERIFIED: `app/package.json`] |
| Type check/build | `cd app && npm run type-check` and `cd app && npm run build`. [VERIFIED: `app/package.json`] |
| Browser E2E | Playwright via `cd app && npm run test:e2e`. [VERIFIED: `app/package.json`] |
| API smoke | Running API exercised through Rust harness with `make test-smoke-rust`. [VERIFIED: `Makefile`, `testing/`] |
| jank | `go test ./...` at inspected commit. [VERIFIED: jank `Makefile`] |

### Exact verification commands

```bash
# vEDH generation and fast domain tests
make generate
make test-unit

# PostgreSQL-backed API suite (start local PostgreSQL first when absent)
make persistence
DATABASE_URL='postgres://edhgo:edhgo@127.0.0.1:5432/edhgo?sslmode=disable' \
ALL_PRINTINGS_JSON_PATH="$PWD/All Printings.json" \
make test-api

# Frontend gates
cd app && npm test
cd app && npm run type-check
cd app && npm run build

# With the API running
make test-smoke-rust

# With API and UI running
cd app && npm run test:e2e

# In the inspected/forked jank repository
go test ./...
JANK_DB_DRIVER=pgx JANK_DB_DSN='<test-postgres-dsn>' go test ./...
JANK_DB_DRIVER=sqlite JANK_DB_DSN='<temporary-sqlite-path>' go test ./...
```

`make persistence` uses Docker to provide PostgreSQL because no host `psql`, `pg_isready`, or `migrate` CLI was found in the inspected environment. Secrets in the illustrative DSNs must come from the test environment and never be logged. [VERIFIED: local environment and `Makefile`]

### Phase requirements → test map

| Phase | Behavior | Test type | Automated command / target | Exists? |
|---|---|---|---|---|
| RI-1 | UUID normalization, grants, auth principal, card identity | Migration + integration + SQL privilege matrix | `make test-api` plus new `go test ./server/... -run 'Identity|Privilege|CardIdentity'` | ❌ Wave 0 |
| RI-2 | Snapshot immutability/hash, stable-source lineage, participant/results, backfill | Unit + property + migration integration | `make test-unit`; `make test-api` | ❌ Wave 0 |
| RI-3 | Transactional append, retries, ordering, corrections, inference visibility | Unit + DB fault injection + API + UI | `make test-unit`; `make test-api`; `cd app && npm test` | ❌ Wave 0 |
| RI-4 | Draft exclusion, limits, immutable submissions/history/deletion | Unit + API + component + E2E | `make test-api`; `cd app && npm test`; `cd app && npm run test:e2e` | ❌ Wave 0 |
| RI-5 | `n=1`, denominators, Limited evidence, filters, rebuild parity | Golden projector + API + component + performance | `make test-unit`; `make test-api`; `cd app && npm test` | ❌ Wave 0 |
| RI-6 | allowlist publication, pseudonyms, version pin/republish/revoke/delete | Adversarial integration + property + E2E + role negative | `make test-api`; `cd app && npm run test:e2e` | ❌ Wave 0 |
| RI-7 | jank migration, owner auth, revisions, relations, forks, evidence revocation | PostgreSQL/SQLite integration + HTTP auth matrix | `JANK_DB_DRIVER=pgx ... go test ./...`; `go test ./...` | ❌ Wave 0 |
| RI-8 | full loop, activation/north-star telemetry, rollback and deletion drills | E2E + smoke + operational rehearsal | all commands above plus runbook scripts | ❌ Wave 0 |

### Required test suites and fixtures

1. **Migration parity:** apply every production and test migration from empty; upgrade a current-schema fixture; run backfill twice; validate constraints; exercise rollback-by-flag; then rerun forward. Check schema fingerprints between `migrations/` and `server/test/migrations/`. [VERIFIED: duplicate migration trees; ASSUMED: fingerprint harness]
2. **Role matrix:** connect as `vedh_app`, `jank_app`, and migrator roles; assert every allowed action and assert denial for credentials, private reviews/events/snapshots/publication sources, cross-schema writes, `CREATE`, owner bypass, and arbitrary functions. Include future-table default privileges. [CITED: PostgreSQL privilege docs]
3. **Transactional fault injection:** fail before event insert, between multiple inserts, before state update, before commit, and after commit/response loss. Assert all-or-nothing persistence and replay returns the same result without duplicate events. [ASSUMED]
4. **Golden event fixtures:** game create/join/draw/move/tap/stack/turn/priority/win/cancel/finish; multiple copies of a card; client-state diff; out-of-order client times; unknown event version; correction; abandoned/no-contest/incomplete. [VERIFIED: current event list and PRDs]
5. **Inference fixtures:** versioned expert holdout with positive/negative labels per category; exact precision calculation; a 94.9% version remains speculative and a ≥95% version can default only for that category/version. Include explicit and absent source-target cases. [VERIFIED: D-16/D-17]
6. **Snapshot/lineage property tests:** normalized input permutation yields the same content hash; quantity/leader/unresolved/game-system change yields a different hash; equal lists with different source IDs do not auto-group; stable same-owner source does; ownership change does not. [VERIFIED: D-09; ASSUMED: property test shapes]
7. **Backfill fixtures:** null/invalid/duplicate UUIDs; duplicate usernames; renamed user; missing event actor; unknown payload; partial deck zones; duplicate card names; ambiguous printing; legacy jank creator without user; restart at every batch checkpoint. [VERIFIED: current schemas; ASSUMED: fixtures]
8. **Publication adversarial corpus:** hidden hand/library/face-down data, stable opponent UUID/name, free-text leaks, raw source payload, linked IDs, speculative inference not selected, revoked/deleted author, malicious HTML/Markdown, cross-publication pseudonym comparison, caching. Assert output allowlist rather than a denylist. [VERIFIED: D-08/D-18 and jank sanitizer dependencies]
9. **Review fixtures:** 0–3 contributors accepted, 4 rejected; zero/one underperformer accepted, two rejected; draft invisible; repeated submits create immutable revisions; publication remains pinned; explicit republish changes version. [VERIFIED: D-11/D-12]
10. **jank authorization matrix:** anonymous/public/private reads; non-owner mutation denial for every tree/node/annotation/revision/fork endpoint; revoked evidence; missing version; custom relation bounds; global filter exclusion; optional thread; source deletion/tombstone. [VERIFIED: current authorization gap and D-20/D-21]
11. **Deletion rehearsal:** request, recover within 30 days, expire, revoke, remove private reviews/profile/auth, tombstone retained facts, search/index/cache invalidation, and restoration re-deletion. Use a controllable clock. [VERIFIED: D-13; ASSUMED: infrastructure coverage]
12. **Load/performance:** at minimum benchmark 10,000 events for one game, 10,000 eligible games in one lineage, 500 tree nodes and 2,000 evidence links, cursor-page latency, projector rebuild, and revocation lookup. These are engineering stress fixtures, not promised production SLOs. [ASSUMED]

### Sampling rate

- **Per task commit:** the narrow Go package/test, resolver test, Vitest file, or jank package touched; target under 30 seconds. [ASSUMED]
- **Per wave merge:** `make test-unit`, affected PostgreSQL integration suite, `cd app && npm test`, type check, and jank `go test ./...` when applicable. [VERIFIED: existing commands; ASSUMED: wave policy]
- **Phase gate:** generation, full Go/API/frontend build/unit suites, affected E2E and Rust smoke, migration/role matrix, reconciliation report, and phase-specific security tests all green. [ASSUMED]
- **Beta gate:** production-shaped migration/restore/rollback rehearsal plus end-to-end product loop, publication privacy, jank authorization, and deletion/revocation drills. [ASSUMED]

### Wave 0 gaps

- [ ] Shared PostgreSQL migration-upgrade/backfill fixture that starts from the current production schema. [VERIFIED: no such fixture found]
- [ ] SQL role capability matrix and future-object default privilege assertions. [VERIFIED: no such tests found]
- [ ] Deterministic clock/UUID fixtures for revisions, tombstones, pseudonyms, and deletion recovery. [ASSUMED]
- [ ] Golden canonical event/inference corpus with versioned expert labels. [VERIFIED: no such fixtures found]
- [ ] Dual-write fault-injection/idempotency harness. [VERIFIED: no such harness found]
- [ ] Privacy publication adversarial fixtures shared by vEDH and jank. [VERIFIED: no such fixtures found]
- [ ] PostgreSQL-backed jank CI job and versioned migration harness. [VERIFIED: inspected jank has no versioned migration suite]
- [ ] E2E fixtures for the complete play → inspect → debrief → compare → revise → publish → jank-tree loop. [VERIFIED: current E2E does not cover this future loop]

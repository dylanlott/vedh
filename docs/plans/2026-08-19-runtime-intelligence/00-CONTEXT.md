# Runtime Intelligence Post-v1 Milestone — Planning Context

**Prepared:** 2026-08-19  
**Status:** Ready for implementation planning  
**Source:** Approved vEDH and jank runtime-intelligence PRDs  
**Planning constraint:** Review-only staging package; do not modify the active `.planning` milestone or implement code.

## Milestone Boundary

Plan the post-v1 work required to turn vEDH into a data-first TCG gameplay and deck-analysis platform, with jank as its evidence-backed community interpretation layer. The deliverable is a dependency-ordered set of executable implementation plans through beta readiness. The plans may identify post-beta work, but must keep it outside the beta critical path.

The active Deck-to-Game Activation roadmap remains authoritative until its Phase 5 release gate completes. These staged plans must not renumber, replace, or edit its phases.

## Locked Product Decisions

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

## Scope Fences

- No code implementation, migrations, commits, branch changes, or active-roadmap edits during this planning review.
- No AI coaching requirement for beta; deterministic and heuristic inference are allowed without an LLM.
- No full TCG rules engine or automatic adjudication.
- No global player rankings, Elo, opponent win-rate product, automatic deck cuts/adds, or public cross-user aggregation in beta.
- No testing groups in the beta critical path.
- No non-Magic adapter in the beta critical path.

## Canonical References

- `docs/product/2026-08-17-runtime-game-intelligence-platform-prd.md` — canonical runtime model, identity, inference, privacy, retention, and beta contract.
- `docs/product/2026-08-17-expert-post-game-debrief-prd.md` — review workflow, immutable revisions, and next-action contract.
- `docs/product/2026-08-17-runtime-deck-analysis-prd.md` — lineage analytics, evidence classes, comparisons, and publishing contract.
- `docs/product/2026-08-17-jank-evidence-backed-card-trees-prd.md` — shared identity, tree, forum, evidence-link, revision, and fork contract.
- `.planning/ROADMAP.md` — active Deck-to-Game Activation milestone; remains unchanged and must complete before this milestone starts.
- `.planning/REQUIREMENTS.md` — active milestone requirements and existing scope fences.
- `server/schema.graphql` — current public GraphQL contract.
- `server/games.go` — current game creation, participant, and deck hydration path.
- `server/gamelog_diff.go` — current runtime event derivation starting point.
- `app/src/views/GameAnalysisView.vue` — current completed-game analysis surface.

## Planning Deliverables

The staged package must include:

1. A master milestone plan with phase boundaries, dependency graph, requirement coverage, sequencing rationale, beta gates, and post-beta deferrals.
2. Executable phase plan files with waves, dependencies, files/modules likely to change, concrete tasks, tests, acceptance criteria, rollback/compatibility notes, and artifacts produced.
3. A verification report proving every locked decision and every PRD story is covered exactly once or explicitly identified as a cross-cutting constraint.


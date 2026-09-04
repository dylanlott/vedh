# Runtime Intelligence Post-v1 Milestone — Master Implementation Plan

**Prepared:** 2026-08-19  
**Revision:** checker revision 3  
**Status:** Review-only staging package  
**Authority:** `00-CONTEXT.md`, corrected `00-RESEARCH.md`, `00-VALIDATION.md`, and the four approved 2026-08-17 PRDs  
**Execution boundary:** These plans do not modify the active `.planning` milestone. Active Deck-to-Game Phase 5 must close and the user must approve this staged package before execution.

## Milestone outcome

Deliver D-01's private-to-public loop: play → retain canonical evidence → reflect privately → compare immutable versions → create a documented revision → deliberately publish bounded evidence → discuss the hypothesis in jank, while preserving one canonical vEDH UUID, immutable evidence, privacy allowlists, honest n=1/limited-evidence language, inference quality gates, deletion/revocation, and reversible cohort flags.

The approved beta starts only after 30 serious testers have activated by completing one eligible game with a canonical participant and immutable snapshot. Recruitment may require more than 30 candidates. The six-week measurement window begins at the timestamp of the 30th activation, not invitation, enrollment, or first activation. Candidate-to-activation exclusions remain a separate denominator.

## Entry and cross-repository preconditions

RI-00 has `depends_on: []`; active Phase 5 is an external fail-closed entry precondition because no executable Phase 5 plan ID yet exists. RI-00 cannot complete until an actual Phase 5 release-gate artifact is linked, the vEDH target is inventoried/restored, and a dedicated jank integration branch/worktree is supplied or created. `docs/contracts/jank-checkout.json` records immutable remote/base_commit plus integration branch, mutable expected_head, and an append-only plan audit. Each jank plan preflights remote, base ancestry, exact expected HEAD, and cleanliness; tasks permit dirt only in declared files; edit plans finish with one atomic unpushed jank commit, advance expected_head/audit, and postflight the new clean head. Read-only plans retain expected_head and audit verification. Wrong remote, divergence, unexpected head, or undeclared dirt fails closed; no production push or merge is implied.

The package never imports spreadsheet data (D-02), never adds beta testing groups (D-19), and never selects/builds a non-Magic adapter (D-22). Those decisions are scope fences, not implementation tasks.

## Revised phase and plan graph

```text
Active Phase 5 completion evidence (external precondition, not a plan ID)
  -> RI-00 (W1 handoff/restore/entry approval)
  -> RI-01-01 (W2 roles)
  -> RI-01-02 (W3 UUID/auth)
       -> RI-01-03 (W4 generated API/session approval) --\
       -> RI-01-04 (W4 FK grant) -----------------------+-> RI-02-01 (W6)
       -> RI-01-05 (W4 card identity/backfill) -> RI-01-06 (W5 approval) --/
  -> RI-02-02 (W7) -> RI-02-03 (W8 approval)
  -> RI-03-01..04 (W9..12; event/inference/timeline approval)
  -> RI-04-01..03 (W13..15; debrief/API/UI/deletion approval)
  -> RI-05-01..03 (W16..18; analytics/API/UI approval)
  -> RI-06-01 (W19 builder) -> RI-06-02 (W20 sharing API)
       -> RI-06-03 (W21 evidence_api) --\
       -> RI-06-04 (W21 deletion hook) -+-> RI-06-05 (W22 privacy approval)
  -> RI-07-01 (W23 identity) -> RI-07-02 (W24 trees)
  -> RI-07-03 (W25 both import sources) -> RI-07-04 (W26 evidence/indicators)
  -> RI-07-05 (W27 cutover approval)
  -> RI-08-01 (W28 vEDH local deletion + durable outbox)
  -> RI-08-02 (W29 verified-jank consumer/receipt/drill)
  -> RI-08-03 (W30 final smoke/readiness command)
  -> RI-08-04 (W31 measurement infrastructure + pre-activation APPROVED gate)
  -> RI-08-05 (W32 recruit/activate; immutable start on 30th activation)
  -> RI-08-06 (W33 six-week operations)
```

RI-03 and RI-04 remain serialized because PGR-3 consumes the typed timeline and both own generated/UI surfaces. The only parallel execution wave is W4 (`RI-01-03`, `RI-01-04`, `RI-01-05`) and W21 (`RI-06-03`, `RI-06-04`); each pair/set has zero file overlap and no hidden logical dependency. D-13, smoke/readiness, pre-activation approval, cohort activation, and the six-week clock are strictly serialized.

## Plan inventory (36 executable plans, 83 tasks, 33 waves)

| Phase | Plans | Focus | Exit gate |
|---|---|---|---|
| RI-00 | `RI-00-PLAN.md` | release handoff, restore, verified jank checkout | explicit entry approval |
| RI-01 | `RI-01-01`..`RI-01-06` | roles; UUID/auth; generated cutover; FK grant; card backfill; card approval | role/UUID/card APPROVED decisions |
| RI-02 | `RI-02-01`..`03` | immutable DDL/fixtures; services/backfill; API/cutover | context cohort approval |
| RI-03 | `RI-03-01`..`04` | event schema; exact-once writes; inference; timeline/decision | separate event and inference decisions |
| RI-04 | `RI-04-01`..`03` | review DDL; notes/search/next actions; UX/deletion hook | review cohort approval |
| RI-05 | `RI-05-01`..`03` | projections; API/export/revision; workspace | analytics cohort approval |
| RI-06 | `RI-06-01`..`05` | builder; sharing API; evidence_api grants; deletion hook; privacy UI | adversarial privacy approval |
| RI-07 | `RI-07-01`..`05` | shared identity; tree schema; both JCT-1 imports; evidence/indicators/discussion; cutover | jank production approval |
| RI-08 | `RI-08-01`..`06` | vEDH outbox; jank receipt; final smoke; pre-activation approval; 30th-activation start; six-week exit | APPROVED before activation and beta exit decision |

## Cross-application ownership and grants

| Capability | Authority | Other application access |
|---|---|---|
| canonical users, credentials, deletion request/recovery/expiry | vEDH | jank authenticates the same UUID principal; no account/credential/link table |
| games, snapshots, participants, results, events, reviews, inferences, analysis, publications | vEDH | jank has no base-table access |
| published evidence contract | vEDH `evidence_api` | `jank_app` gets USAGE + SELECT only on versioned safe views |
| forum, moderation, trees, revisions, annotations, evidence-reference rows | jank | vEDH never mutates them; D-13 uses a durable command consumed by a dedicated jank worker |
| D-13 cross-process bridge | vEDH owns jobs/outbox/lease/ack functions; jank owns apply function/receipt | local vEDH deletion+command is one transaction; later jank tombstone+ack is a separate worker transaction; no shared Go transaction |
| canonical-user jank foreign keys | jank-owned tables → vEDH user key | exact temporary `GRANT REFERENCES (canonical_user_id) ... TO jank_migrator`, validate FKs, then REVOKE; no users SELECT/DML |

JCT-7 acceptance belongs only to RI-07-01. RI-01 supplies the canonical UUID/role/auth/FK-grant prerequisites and does not claim JCT-7.

## Exact-once PRD story ownership

| Story | Canonical meaning | Primary plan |
|---|---|---|
| RGI-1 | Capture an immutable game context | RI-02-01 |
| RGI-2 | Record normalized runtime events, including versioned Magic shuffle/randomization with persistence/correction/idempotency | RI-03-01 |
| RGI-3 | Finish games with complete result semantics | RI-02-01 |
| RGI-4 | Enrich observed events additively with exact calculation-definition/version and authorized source drill-down | RI-03-03 |
| RGI-5 | Inspect a game as evidence | RI-03-04 |
| RGI-6 | Reuse gameplay evidence in jank | RI-06-01 |
| RGI-7 | Support additional TCGs without flattening their rules | RI-03-01 |
| PGR-1 | Start from automatically recorded facts | RI-04-01 |
| PGR-2 | Identify meaningful card contributions | RI-04-01 |
| PGR-3 | Capture the lesson in the player's own words | RI-04-02 |
| PGR-4 | Classify lessons without flattening them | RI-04-01 |
| PGR-5 | Decide what to do next | RI-04-02 |
| PGR-6 | Keep reflection private or share it deliberately | RI-06-02 |
| PGR-7 | Skip without friction | RI-04-03 |
| RDA-1 | Analyze a deck lineage without rewriting history | RI-05-01 |
| RDA-2 | Understand outcomes and game pace | RI-05-01 |
| RDA-3 | Inspect runtime card activity | RI-05-01 |
| RDA-4 | Compare deck revisions | RI-05-02 |
| RDA-5 | Filter by runtime context | RI-05-02 |
| RDA-6 | Trace every conclusion to evidence, including export definitions/filters/source IDs | RI-05-02 |
| RDA-7 | Publish a bounded analysis to jank | RI-06-01 |
| JCT-1 | Create from private owner lineage/snapshot without publication or an authorization-filtered published-analysis slice | RI-07-03 |
| JCT-2 | Express card relationships, not just indentation | RI-07-02 |
| JCT-3 | Attach runtime evidence to a bounded annotation/card claim | RI-07-04 |
| JCT-4 | See all compact runtime indicators and navigate to authorized sources | RI-07-04 |
| JCT-5 | Discuss and quote evidence in a thread | RI-07-04 |
| JCT-6 | Compare and fork analytical trees | RI-07-02 |
| JCT-7 | Use one vEDH identity across gameplay and discussion | RI-07-01 |

## Locked-decision coverage

| Decision | Primary enforcement | Required cross-check |
|---|---|---|
| D-01 product loop | RI-05-03/RI-08-03/RI-08-06 | revision UI, complete-loop smoke, and operations exercise every transition |
| D-02 no spreadsheet data | RI-00 | inventory/fixtures attest false |
| D-03 rolling north star | RI-05-03/RI-08-04/RI-08-06 | documented revision UI, measurement, and rolling-window exit report |
| D-04 six weeks, 30 activated | RI-02-02/RI-08-04/RI-08-05/RI-08-06 | eligibility, readiness, 30th activation, and six-week operations |
| D-05 immutable canonical evidence | RI-02/RI-03 | correction/version and exact-once tests |
| D-06 shared UUID/no duplicate account | RI-01 prerequisite; RI-07 acceptance | guest claim preserves all relationships |
| D-07 ownership/least privilege | RI-01-01 and RI-07-01 | role-negative and temporary REFERENCES revoke |
| D-08 private default/safe publication | RI-06 | allowlist, scoped pseudonyms, immutable sources |
| D-09 stable-source lineage | RI-02 | no name/list similarity grouping |
| D-10 first-game/limited evidence | RI-05 | n=1 visible; either side below 3 warns |
| D-11 compact optional debrief | RI-04 | no beta activation dependency |
| D-12 immutable review/pinning | RI-04/RI-06 | edits never silently republish |
| D-13 30-day deletion | RI-08-01/02 | local deletion+outbox, verified-jank tombstone+receipt, cleanup, monitoring, and restore-generation replay |
| D-14 evidence classes/additivity | RI-03 | RI-04/05/07 preserve player/community separation |
| D-15 enrichment categories | RI-03-03 | versioned fixture-supported projectors |
| D-16 95% gate | RI-03-03/04 | implementation flags off; blocking exact-version approval |
| D-17 causality | RI-03 | explicit source-target fact; other output hypothesis |
| D-18 n=1/explicit inference publication | RI-06 | exact uncertainty/source/version preserved |
| D-19 teams post-beta | all beta plans | no owner/admin/member testing-group product scope |
| D-20 standalone trees/forks | RI-07 | optional independent threads/backlinks |
| D-21 relation vocabulary | RI-07 | custom searchable, not global category |
| D-22 non-Magic post-beta | RI-03 envelope only | Magic is the sole beta adapter |

## Inference enablement contract

RI-03-03 implements and evaluates inference while leaving read/default flags off. RI-03-04 separately implements typed display, then pauses at a blocking human checkpoint. The reviewer records APPROVED or REJECTED per category/version; absence/rejection prevents default display. A 94.9% version remains speculative, and only explicitly approved versions with stored precision ≥95% may default-display. This checkpoint is not embedded as a comment in implementation work.

## Account deletion orchestration

RI-08-01 owns the vEDH-local transaction: it rechecks the controllable 30-day deadline/recovery, deletes/revokes/tombstones every vEDH-owned domain, and inserts a unique durable jank command or rolls back. RI-08-02 owns the verified-checkout consumer: a dedicated least-privilege worker leases the command, calls a jank-owned apply function, stores the generation receipt, calls the vEDH ack that scrubs the pending subject UUID, and commits those jank/ack changes in its own PostgreSQL transaction. There is no shared Go transaction across processes. A job is globally complete only after jank and cleanup receipts. Lease expiry, retry/backoff, response loss, role negatives, receipt-age monitoring, non-reversible restore fingerprints, and restore-generation replay are all automated.

## Beta quality and operating gates

Pre-beta requires:

- migration parity from empty/current via `server/migration_parity_test.go` and `go test ./server/...`; launcher references use `persistence/sql.go`;
- exact-once transaction fault/retry tests and two-pass backfills;
- role/FK grant matrix, adversarial privacy allowlist, jank auth/owner/search negatives;
- inference 94.9%/95% evaluation, projector/export reconciliation, and human inference/privacy/cutover decisions;
- 10,000-event p95 ≤750 ms, 10,000-lineage query p95 ≤1 s, and 500-node/2,000-link jank fixtures;
- live Rust smoke in `tools/smoke/src/main.rs` wired by `tools/smoke/Cargo.toml` and `make test-smoke-rust`;
- request/recovery/expiry/replay/failure/restore deletion drill and reverse-order rollback;
- complete play→inspect→debrief→compare→revise→share/publish→jank loop.

Recruit candidates until 30 activate. Only then record `measurement_window_started_at` and begin six consecutive weeks. Weekly sampling and the final report are specified in `00-VALIDATION.md` and RI-08-06.

## Feature-flag order and rollback

Flags default off and enable in dependency order: canonical UUID → snapshot dual-write → canonical events → typed timeline → approved inference versions → review → analytics → publication write/read → jank shared identity/evidence trees → beta cohort. Rollback disables in reverse order, preserves canonical/immutable rows, and uses correction/new versions rather than destructive down migrations. Jank identity rollback enters maintenance/read-only; it never restores duplicate signup/login/token minting.

## Source coverage audit

| Source | Items | Coverage | Status |
|---|---|---|---|
| GOAL | play→evidence→reflection→revision→jank and 30-activated six-week beta | RI-02..RI-08 | COVERED |
| REQ | RGI-1..7 | exact-once table above | COVERED |
| REQ | PGR-1..7 | exact-once table above | COVERED |
| REQ | RDA-1..7 | exact-once table above | COVERED |
| REQ | JCT-1..7 | exact-once table above | COVERED |
| RESEARCH | roles/UUID/card identity, expand/backfill/cutover, event/inference, review, analytics, publication, jank, operations | RI-00..RI-08 | COVERED |
| CONTEXT | D-01..D-22 | decision table and cited tasks | COVERED |
| EXCLUDED | spreadsheet data, beta teams, non-Magic adapter, AI coach/rules engine/rankings/automatic deck edits/public cross-user aggregation | explicit prohibitions/scope fences | EXCLUDED BY DECISION |

No source item is missing and no beta plan contains executable post-beta scope.

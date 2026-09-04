# Runtime Intelligence Validation Contract

**Purpose:** map every executable plan/task to automated evidence, Wave-0 fixtures, phase gates, scale/latency checks, sampling continuity, and blocking human decisions. Commands run from the vEDH root unless a verified jank checkout command explicitly changes directory.

`JANK_CHECKOUT_PATH` is never guessed. RI-00 records immutable remote/base_commit, integration branch, mutable expected_head, and audit history. Jank plans run pre-plan verification (remote, base ancestry, HEAD=expected_head, clean), task verification (only declared dirt), then an atomic unpushed commit/expected_head advancement/post-plan clean verification; read-only plans retain expected_head. No production push/merge is implied. Commands use explicit mode and plan path:

```bash
JANK_CHECKOUT_PATH=$(jq -er '.path' docs/contracts/jank-checkout.json)
scripts/verify-jank-checkout.sh --contract docs/contracts/jank-checkout.json --mode task --plan docs/plans/2026-08-19-runtime-intelligence/RI-XX-YY-PLAN.md
```

Test-first RED verification is successful only through `scripts/verify-expected-red.sh`: it requires a nonzero precise suite, allowlisted failing test name(s), and the exact `EXPECTED_RED[plan-task]` reason, while rejecting compile/package-setup, tool/file/database/environment, panic, timeout, race, no-test, and unrelated-test failures. The immediately following implementation task runs that identical named suite directly and requires exit zero. This rule applies to RI-01-01, RI-01-02, RI-01-05, RI-02-01, RI-03-01, RI-03-02, RI-03-03, RI-04-01, RI-05-01, RI-06-01, RI-07-01, RI-08-01, RI-08-02, and RI-08-04; no plain intentionally failing `go test` is success evidence.

## Task-to-command matrix

| Plan/task | Automated command/evidence |
|---|---|
| RI-00 T1 inventory/checkout/baseline | inventory/checkout syntax+self-test, locked-scope booleans, and every flag owner/prerequisite/rollback entry |
| RI-00 T2 restore | `bash -n scripts/runtime-restore-rehearsal.sh && scripts/runtime-restore-rehearsal.sh --self-test-target-guards` |
| RI-00 T3 entry decision | require actual Phase 5 completion evidence and `Entry gate decision: APPROVED`; REJECTED does not complete the plan |
| RI-01-01 T1 migration/roles fixtures | expected-red harness requires only `TestRuntimeRoleContract` plus its exact RI-01-01-T1 sentinel; T2 runs the identical suite green |
| RI-01-01 T2 role DDL | same role/parity suite plus production/test migration diff |
| RI-01-01 T3 role checkpoint | require `Role boundary decision:` and rerun role matrix |
| RI-01-02 T1-T3 UUID/auth | canonical UUID/request/recovery fixtures, migration parity, and every live auth/authorization seam; no generated API or FK grant files |
| RI-01-03 T1 generated API | `make generate` plus canonical identity API/auth negatives |
| RI-01-03 T2 session checkpoint | require both session and canonical UUID decisions APPROVED |
| RI-01-04 T1-T2 FK grant | role-negative test plus exact temporary REFERENCES/validate/REVOKE contract |
| RI-01-05 T1-T3 card identity/backfill | exact/ambiguous/unresolved/instance suites, migration parity, two-pass backfill; no embedded approval |
| RI-01-06 T1 card checkpoint | rerun card/backfill gates and require `Card identity cutover decision: APPROVED` |
| RI-02-01 T1 context fixtures | expected-red harness for `TestRuntimeContextSchemaContract`; T2 runs that identical suite green before migration parity |
| RI-02-01 T2 context DDL | `go test ./server/... -run 'TestRuntimeMigrationParity|TestRuntimeContextSchema' -race` plus migration diff |
| RI-02-02 T1 snapshot/lineage services | `go test ./pkg/deckidentity ./server/... -run 'TestSnapshot|TestLineage' -race -v` |
| RI-02-02 T2 participant/result dual-write | `go test ./server/... -run 'TestCreateGameSnapshotAtomic|TestJoinGameParticipantAtomic|TestFinishResultAtomic|TestResultRevision' -race -v` |
| RI-02-02 T3 context backfill | `go run ./cmd/runtime-context-backfill --self-test && go test ./server/... -run TestRuntimeContextBackfill -race` |
| RI-02-03 T1 context API | `make generate && go test ./server/... -run 'TestRuntimeContextAPI|TestRuntimeContextAuthorizationNegative' -race -v` |
| RI-02-03 T2 context checkpoint | require `Production context cutover decision:` and rerun backfill/auth negatives |
| RI-03-01 T1 event fixtures | expected-red Magic-v1 contract including named shuffle/randomization vocabulary and persistence/correction/idempotency cases |
| RI-03-01 T2 event DDL/envelope | identical contract green plus shuffle/randomization, migration/immutability, and migration diff |
| RI-03-02 T1 fault fixtures | expected-red harness for `TestGameMutationAtomicityContract`; T2 runs that identical suite green before expanded fault/retry checks |
| RI-03-02 T2 transactional live wiring | mutation fault/retry/dual-write/metrics suite |
| RI-03-02 T3 event backfill | `go run ./cmd/runtime-event-backfill --self-test && go test ./server/... -run TestRuntimeEventBackfill -race` |
| RI-03-03 T1 inference fixtures | expected-red inference/calculation contract requires exact calculation-definition/version and source linkage for every derived/inferred fact |
| RI-03-03 T2 inference/projectors | identical contract green plus immutable definition registry, exact-version/source tests, migration diff, and flags off |
| RI-03-04 T1 timeline API | generation plus typed timeline, exact calculation-version drill-down, source authorization, and 10k tests |
| RI-03-04 T2 timeline UI | typed definition/version/source drill-down component tests, type-check, and build |
| RI-03-04 T3 inference decision | canonical events must be APPROVED to complete the plan; inference default-display may separately be APPROVED or REJECTED and remains off when rejected |
| RI-04-01 T1 review fixtures | expected-red harness for `TestGameReviewSchemaContract`; T2 runs that identical suite green before expanded migration/revision checks |
| RI-04-01 T2 review DDL | review migration/revision suite plus migration diff |
| RI-04-02 T1 review service | game-review/revision/auth-negative suite |
| RI-04-02 T2 note links/search | `go test ./server/... -run 'TestReviewNoteLinks|TestPrivateReviewSearch|TestPrivateReviewSearchAuthorization' -race -v` |
| RI-04-02 T3 next-action/API | `make generate && go test ./server/... -run 'TestReviewNextAction|TestReviewGraphQL|TestPrivateReviewSearch' -race -v` |
| RI-04-03 T1 review UI/skip | `npm --prefix app test -- GameDebrief.spec.ts && npm --prefix app run type-check && npm --prefix app run test:e2e -- game-debrief.spec.ts` |
| RI-04-03 T2 review deletion hook | `go test ./server/... -run 'TestReviewDeletionHook|TestReviewDeletionReplay|TestReviewDeletionRestore' -race -v` |
| RI-04-03 T3 review checkpoint | require `Review cohort decision: APPROVED`; REJECTED does not complete the plan |
| RI-05-01 T1 analytics fixtures | expected-red harness for `TestLineageProjectionContract`; T2 runs that identical suite green before shadow/rebuild checks |
| RI-05-01 T2 analytics projector | lineage shadow/rebuild parity suite plus migration diff |
| RI-05-02 T1 analysis API | `make generate && go test ./server/... -run 'TestDeckAnalysis|TestSnapshotComparison|TestAnalysisAuthorizationNegative|TestDeckAnalysis10k' -race -v` |
| RI-05-02 T2 analysis export | `make generate && go test ./server/... -run 'TestAnalysisExport|TestAnalysisExportAuthorization|TestAnalysisExportReproducible' -race -v` |
| RI-05-02 T3 documented revision | `make generate && go test ./server/... -run 'TestCreateDeckRevision|TestResolveReviewNextAction' -race -v` |
| RI-05-03 T1 analysis UI | `npm --prefix app test -- DeckAnalysisView.spec.ts && npm --prefix app run type-check && npm --prefix app run build && npm --prefix app run test:e2e -- deck-analysis.spec.ts` |
| RI-05-03 T2 analytics checkpoint | require `Analytics cohort decision: APPROVED`; REJECTED does not complete the plan |
| RI-06-01 T1 publication fixtures | expected-red harness for `TestPublicationPersistenceContract`; T2 runs that identical suite green before builder/persistence checks |
| RI-06-01 T2 publication DDL/builder | builder/persistence suite plus migration diff |
| RI-06-02 T1-T2 sharing/publication API | participant/public preview-save parity, auth, source/calculation/review/inference pinning, republish/revoke |
| RI-06-03 T1 evidence_api grants | evidence view/role/revocation suite plus 20260820141000 migration diff |
| RI-06-04 T1 publication deletion hook | local hook/replay/restore suite; no jank mutation claim |
| RI-06-05 T1 publication UI | publication preview unit/type/E2E |
| RI-06-05 T2 privacy checkpoint | require `Publication cutover decision: APPROVED` |
| RI-07-01 T1 identity fixtures | verified checkout then jank shared-identity/guest-claim/no-local-account suites |
| RI-07-01 T2 identity migration/FKs | verified checkout then shared-identity/FK-grant-revoked/replay suites |
| RI-07-01 T3 principal cutover | verified checkout then canonical-principal/guest-claim/no-local-account suites |
| RI-07-02 T1-T2 tree schema/handlers | verified checkout then exact relationship vocabulary, custom bounds, preview parity, revisions/diffs/forks/auth |
| RI-07-03 T1 dual-source manifests | vEDH private owner and published-analysis manifest/version/auth suites |
| RI-07-03 T2 dual-source save | verified checkout then private/published import preview-save parity/auth suites |
| RI-07-04 T1-T3 annotations/indicators/discussion/search/load | expected-head task mode; exact five kinds; documented text/import bounds; escaped rendering; safe vEDH-only URLs/no embeds; stored/rendered/search XSS and oversized-input tests; six indicators, navigation, JCT-5, 500/2,000 scale |
| RI-07-05 T1 rehearsal | verified checkout then full PostgreSQL suite including both imports and indicators |
| RI-07-05 T2 checkpoint | verified checkout and require `Jank production cutover decision: APPROVED` |
| RI-08-01 T1-T3 vEDH local/outbox | 20260820160000 migration parity; local atomicity; lease/ack/role/retry/monitoring/cleanup/restore-generation suites |
| RI-08-02 T1-T3 jank consumer/receipt/drill | expected-head task mode; expected-red bridge followed by identical green suite; deletion/role/replay/restore tests and drill; atomic jank commit advances audit |
| RI-08-03 T1-T2 final smoke/readiness command | verified checkout, wired Rust/Playwright, then post-deletion fail-closed gate |
| RI-08-04 T1-T2 beta measurement | closed-before-approval, 30th-start, north-star, and 20260820161000 migration parity |
| RI-08-04 T3 readiness checkpoint | require APPROVED while measurement start is unset; REJECTED does not complete |
| RI-08-05 T1 cohort activation | require prior APPROVED, exactly 30 activated, start source equals 30th eligible activation |
| RI-08-06 T1 weekly automation | verified checkout, rollback target guards, full gate, D-13 receipts/restore sampling |
| RI-08-06 T2 beta exit | verified checkout, six complete weeks, denominator 30, numerator, and APPROVED/NO-GO exit |

## Wave-0 fixtures

| Fixture | Created by | Required before |
|---|---|---|
| current-schema migration upgrade + role matrix | RI-01-01 T1 | all Runtime Intelligence DDL |
| stable UUID across signup/login/guest/refresh/claim/token/loaders | RI-01-02 T1 | any canonical FK or jank identity work |
| exact/ambiguous/unresolved card corpus | RI-01-05 T1 | snapshots/events/reviews/trees |
| snapshot/lineage/participant/result property/backfill corpus | RI-02-01 T1 | context services/dual-write |
| typed Magic event + 10k event + transaction fault corpus | RI-03-01 T1 and RI-03-02 T1 | canonical write/read switch |
| expert-labeled inference holdout with 94.9% and ≥95% | RI-03-03 T1 | any default inference decision |
| review limits/note links/search/next-action/privacy corpus | RI-04-01/02 | review cohort |
| lineage n=1/2/3/10k and exact export corpus | RI-05-01 and RI-05-02 | analytics cohort |
| publication adversarial allowlist/pseudonym/revoke corpus | RI-06-01 T1 | evidence_api/publication flags |
| jank identity/guest-claim/private-import/tree/evidence corpus | RI-07-01..03 | jank cutover |
| D-13 local/outbox/jank receipt/role/failure/replay/restore corpus | RI-08-01 T1 and RI-08-02 T1 | final readiness smoke |
| candidate/readiness-closed/activation/30th-start/north-star corpus | RI-08-04 T1 | cohort activation |

## Phase gates

| Phase | Required gate |
|---|---|
| RI-00 | inventory + verified jank checkout + isolated restore + entry approval |
| RI-01 | `go test ./server/...` migration/role/UUID/card suites; session/role/card decisions |
| RI-02 | generation + server context/backfill/auth suites; context cohort decision |
| RI-03 | runtimeevent/inference/projection/server/Vue gates; 10k p95; separate event/inference decisions |
| RI-04 | review server/Vue/E2E/deletion-hook/privacy gates; review cohort decision |
| RI-05 | analysis/projector/export/server/Vue/E2E; 10k p95; analytics decision |
| RI-06 | adversarial publication/share/evidence_api role/Vue/E2E; privacy decision |
| RI-07 | verified checkout; PostgreSQL and SQLite migrations; all jank auth/tree/evidence/search/load tests; cutover decision |
| RI-08 | vEDH outbox + verified-jank receipt/drill, then final fail-closed smoke, APPROVED before activation, immutable 30th-activation start, weekly gates, exit decision |

## Latency and scale checks

| Surface | Fixture and gate |
|---|---|
| game timeline | 10,000 authorized events; p95 ≤750 ms under documented beta profile |
| lineage analysis/export | 10,000 eligible games; interactive summary p95 ≤1 s; export asynchronous/bounded and reproducible |
| tree/evidence | 500 nodes and 2,000 annotations/evidence links; cursor/bounded reads and revocation lookup recorded |
| mutation path | fault/retry/concurrency suite; 99.5% accepted beta mutations persist one valid event set or visible retryable error |
| finished context | ≥95% beta finished games complete participant/result/format/snapshot contract |
| card identity | ≥99% card-bearing beta events resolve or preserve explicit unresolved identity |

These are measured gates, not hardcoded test-count strings. Missing results fail the pre-beta gate.

## Sampling continuity

- Per task: run the narrow command in the matrix; target under 60 seconds where possible.
- Per plan: run every task command plus generation/type/build when generated contracts changed.
- Per phase: run the phase gate and production/test migration parity from empty and current fixtures.
- Per wave merge: rerun affected Go/API/Vue/jank suites; same-wave plans have zero file overlap.
- Pre-beta: full `scripts/runtime-beta-gate.sh`, wired Rust smoke, role/privacy/restore/rollback/deletion drills.
- Week 1: candidate → activation exclusions, 30-activation lock continuity, first eligible-game completeness.
- Week 2: canonical event exact-once/reconciliation, retry/correction sample, timeline latency.
- Week 3: debrief funnel, draft/search privacy, revision continuity.
- Week 4: three-game lineage, n=1/Limited evidence comprehension, export and documented revision.
- Week 5: participant share/publication preview parity, privacy sample, revoke/cache/jank evidence link.
- Week 6: repeated complete loop, rolling 14-day north star, inference versions, D-13/rollback/restore exit drills.

Every weekly record includes command/query version, timestamp, measured/not_measured, result, incidents, and pause/rollback state. The window is anchored to the 30th activation; a pause is recorded and never silently treated as a completed week.

## Blocking human checkpoints

| Checkpoint | Approval required for |
|---|---|
| RI-00 T3 | any runtime-intelligence DDL/flag work |
| RI-01-01 T3 | production role credentials |
| RI-01-03 T2 | canonical UUID/session cutover |
| RI-01-06 T1 | card identity cutover |
| RI-02-03 T2 | snapshot/context dual-write cohort |
| RI-03-04 T3 | canonical event cohort and each inference default category/version |
| RI-04-03 T3 | private review cohort |
| RI-05-03 T2 | analytics projector/read switch |
| RI-06-05 T2 | publication write/read flags |
| RI-07-05 T2 | jank shared identity/evidence trees |
| RI-08-04 T3 | technical/human readiness before activation opens; only APPROVED completes |
| RI-08-05 T1 | operational wait until exactly 30 eligible activations; clock starts transactionally, not by approval |
| RI-08-06 T2 | beta exit go/no-go |

Each checkpoint records APPROVED or REJECTED, reviewer/operator, timestamp, evidence, allowed flag/category/version scope, and remediation. Silence or missing evidence is rejection.

## Internal planning-package checks

Run before accepting this package:

```bash
# 36 executable plan files, 83 tasks, and 33 waves; canonical RI-00 plus RI-01..RI-08 inventories in 00-MASTER-PLAN.md
find docs/plans/2026-08-19-runtime-intelligence -maxdepth 1 -name 'RI-*-PLAN.md' | sort

# Every task has name, files where applicable, action, acceptance, automated verify, and done.
rg -c '<task ' docs/plans/2026-08-19-runtime-intelligence/RI-*-PLAN.md
rg -c '<done>' docs/plans/2026-08-19-runtime-intelligence/RI-*-PLAN.md

# Story IDs appear once in executable-plan requirements and all D-01..D-22 appear in the package.
rg '^requirements:' docs/plans/2026-08-19-runtime-intelligence/RI-*-PLAN.md
for n in $(seq -w 1 22); do rg -q "D-$n" docs/plans/2026-08-19-runtime-intelligence || exit 1; done

# No obsolete/wrong paths.
! rg 'testing/src/runtime_intelligence.rs|persistence/persistence.go|server/test/migration_parity_test.go' docs/plans/2026-08-19-runtime-intelligence/RI-*-PLAN.md
! rg '^\s*- \.\./jank' docs/plans/2026-08-19-runtime-intelligence/RI-*-PLAN.md
```

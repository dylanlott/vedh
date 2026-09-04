---
phase: RI-01-shared-database-identity
plan: RI-01-06
type: execute
wave: 5
depends_on: ["RI-01-05"]
files_modified:
  - docs/runbooks/card-identity-backfill.md
requirements: [RI-CARD-IDENTITY-GATE]
autonomous: false
must_haves:
  truths:
    - "Card-identity quality is independently approved only after exact, ambiguous, unresolved, replay, and migration evidence is reviewed."
    - "Rejected or absent approval keeps dependent context/event/tree cutovers disabled and is not dependency completion."
  artifacts:
    - path: "docs/runbooks/card-identity-backfill.md"
      provides: "backfill evidence and blocking APPROVED/REJECTED decision"
  key_links:
    - from: "card identity quality decision"
      to: "RI-02 immutable snapshot work"
      via: "APPROVED-only dependency gate"
---

<objective>Independently approve or reject card identity quality after the backfill has completed. Output: a blocking evidence-backed decision, never an approval embedded in automated migration work.</objective>
<execution_context>@docs/plans/2026-08-19-runtime-intelligence/RI-01-05-PLAN.md @docs/runbooks/card-identity-backfill.md</execution_context>
<tasks>

<task type="checkpoint:human-verify" gate="blocking-human">
  <name>Task 1: Approve card identity quality and production cutover</name>
  <files>docs/runbooks/card-identity-backfill.md</files>
  <action>Review two-pass counts, exact/ambiguous/unresolved samples without raw private data, game-instance handling, migration parity, constraint validation, failure/restart behavior, and rollback. Record Card identity cutover decision: APPROVED or REJECTED with reviewer, timestamp, evidence, unresolved counts, and remediation. Rejection/absence leaves the plan incomplete and all dependent writes/reads disabled.</action>
  <acceptance_criteria>Approval is present only when every source is classified, replay is idempotent, ambiguity is explicit, and no fuzzy historical rewrite occurred.</acceptance_criteria>
  <verify><automated>go run ./cmd/card-identity-backfill --self-test &amp;&amp; go test ./pkg/cardidentity ./server/... -run TestCardIdentity -race &amp;&amp; grep -q 'Card identity cutover decision: APPROVED' docs/runbooks/card-identity-backfill.md</automated></verify>
  <done>Card identity cutover is APPROVED; rejection/absence halts RI-02 and is not dependency completion.</done>
</task>

</tasks>
<verification>Re-run card identity/backfill gates and require an APPROVED decision.</verification>
<success_criteria>RI-02 can trust card identity because implementation evidence and human authorization are separate and both passed.</success_criteria>

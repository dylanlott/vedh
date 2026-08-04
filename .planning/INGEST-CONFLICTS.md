## Conflict Detection Report

Ingest date: 2026-08-03. Mode: new. Precedence: ADR > SPEC > PRD > DOC (no per-doc overrides; both classifications have `precedence: null`).

Documents evaluated:
- docs/product/2026-07-23-deck-to-game-activation-prd.md — PRD, high confidence, locked: false
- docs/plans/2026-07-23-deck-to-game-activation-tickets.md — SPEC, medium confidence, locked: false

### BLOCKERS (0)

None. Neither document is locked, so no LOCKED-vs-LOCKED evaluation was possible. No UNKNOWN or low-confidence classifications. No synthesis cycle. No missing prerequisite.

### WARNINGS (0)

None. The competing-variants bucket requires two same-precedence sources defining the same requirement with non-identical acceptance criteria. This ingest set contains one PRD and one SPEC, so every substantive difference was resolvable by precedence and is recorded below as INFO. Nothing here requires a user decision before routing.

### INFO (8)

[INFO-1] Auto-resolved: SPEC > PRD on whether a deck provider must ship in the MVP
  Found: docs/product/2026-07-23-deck-to-game-activation-prd.md Story A2 acceptance states "At least one public deck-link provider ships in the MVP through an adapter interface." Its MVP scope and launch gate both presume a selected provider ("The selected provider's public-deck fixture suite passes").
  Found: docs/plans/2026-07-23-deck-to-game-activation-tickets.md Scope Guardrails state "If neither provider passes ACT-003's feasibility gate, stop that ticket at a documented decision and ship paste-based activation rather than weakening SSRF or reliability controls." ACT-003 acceptance permits the decision record to "explicitly escalate if neither clears the gate."
  Note: SPEC wins under default precedence. A shipped provider adapter is conditional on the ACT-003 feasibility gate, not guaranteed. Recorded on REQ-A2 as a precedence_note; all other A2 acceptance clauses stand unmodified. Downstream planning must not treat "one provider ships" as a fixed MVP commitment.

[INFO-2] Auto-resolved: SPEC > PRD on guest account-claim phasing
  Found: docs/product/2026-07-23-deck-to-game-activation-prd.md Phased Rollout places "Guest account claiming" in v1.1. The MVP scope list omits it.
  Found: docs/plans/2026-07-23-deck-to-game-activation-tickets.md makes the claim backend P0 (ACT-005, "Guest identity and account-claim backend"), the claim UI P1 (ACT-010), and lists ACT-010 among the dependencies of the P0 release gate ACT-013 ("ACT-003 through ACT-010"). ACT-013 acceptance requires "Account claim preserves game access after refresh."
  Note: SPEC wins under default precedence. Account claim is inside the MVP release gate, not deferred to v1.1. Recorded on REQ-A5 and as a phasing_note on REQ-ACT-010. This one materially affects phase derivation: a roadmapper reading only the PRD's phase table would wrongly defer claim work past the MVP.

[INFO-3] Auto-resolved: SPEC > PRD on delivery week for safe invite preview
  Found: docs/product/2026-07-23-deck-to-game-activation-prd.md Recommended Delivery Sequence places "safe invite preview" in week 2.
  Found: docs/plans/2026-07-23-deck-to-game-activation-tickets.md Suggested Four-Week Cut places ACT-007 (Safe Public Invite Preview) in week 3; week 2 is finish ACT-003/ACT-004, ACT-005, ACT-006, begin ACT-013.
  Note: SPEC wins under default precedence. The two sequences agree on every other item. The SPEC ordering is also the one consistent with the ACT dependency graph, so it should be used for wave derivation.

[INFO-4] Four cross-referenced files are absent — all are forward deliverables, not prerequisites
  Found: docs/plans/2026-07-23-deck-to-game-activation-tickets.md cross-references docs/research/deck-provider-feasibility.md, docs/analytics/deck-to-game-activation.sql, docs/runbooks/deck-to-game-quiet-beta.md, and docs/runbooks/deck-to-game-release.md. Absence of all four verified on disk at ef2732a.
  Note: Each appears in a "Likely Files" list, meaning the ticket produces it. deck-provider-feasibility.md is produced by ACT-003; deck-to-game-activation.sql and deck-to-game-quiet-beta.md by ACT-012; deck-to-game-release.md by ACT-013. These are outputs of the plan, so the missing-prerequisite BLOCKER rule does not apply and the safety gate is not tripped. Recorded as `produces:` on the corresponding requirements.

[INFO-5] Mutual cross-reference between the two ingested documents — evaluated, not a synthesis cycle
  Found: the cross-ref graph contains a two-node mutual edge. The PRD lists docs/plans/2026-07-23-deck-to-game-activation-tickets.md in cross_refs; the tickets doc names the PRD in its `**PRD:**` header field. Max traversal depth 2, far under the 50 cap.
  Note: This was NOT recorded as a cycle BLOCKER, which is a deliberate deviation from the literal cycle-detection rule. Rationale: the rule exists because recursive resolution can loop and produce garbage. Here the edge is bibliographic (a parent PRD and its derived ticket breakdown citing each other), not a derivation or inclusion edge, and precedence gives a total order over the two documents (SPEC > PRD), so merge order is deterministic and no loop is possible. Blocking would have produced zero synthesized output for a normal PRD-plus-tickets pair. Flagging explicitly so this judgment can be overridden if the strict reading is preferred.

[INFO-6] Code baseline verified against current HEAD
  Found: both documents pin "Code baseline: `main` at `ef2732a`". Repository HEAD is ef2732a.
  Note: The two documents agree, and the described work is unbuilt against exactly this tree. No drift between plan assumptions and the working tree.

[INFO-7] No ADRs and no locked documents in the ingest set
  Found: 1 PRD and 1 SPEC; both classifications carry `locked: false`; zero ADR-classified documents.
  Note: decisions.md contains no `status: locked` or `status: proposed` entries, because none can be derived from a PRD or SPEC without fabricating an ADR. Technical commitments were routed to constraints.md, which reflects their actual source. No LOCKED-vs-LOCKED evaluation was performed or possible.

[INFO-8] Four open product decisions carried forward unresolved
  Found: docs/product/2026-07-23-deck-to-game-activation-prd.md "Open Decisions" lists the first public deck provider; guest expiry duration (24 hours vs seven days, token 24 hours either way); desktop-only vs tablet breakpoint for the first quiet beta; and the operational surface for product-funnel queries before a dedicated dashboard exists.
  Note: The SPEC does not resolve any of the four. It defines a resolution mechanism for the first (ACT-003's one-day feasibility gate) and partially narrows the fourth (PostgreSQL for funnel queries, Prometheus for technical panels), but names no answers. Both documents agree these are open, so this is agreement rather than conflict and is not a competing variant. Recorded as OPEN-1..OPEN-4 in decisions.md with `status: open`, outside the locked|proposed taxonomy. OPEN-1 and OPEN-2 have downstream implementation impact and will need answers during execution.

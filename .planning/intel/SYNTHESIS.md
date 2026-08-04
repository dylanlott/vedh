# Synthesis

Entry point for downstream consumers (gsd-roadmapper). Ingest date: 2026-08-03.
Mode: new (net-new bootstrap; no prior `.planning/` state existed).
Precedence applied: ADR > SPEC > PRD > DOC, no per-doc overrides.
Code baseline: `main` at `ef2732a` — verified equal to current repository HEAD.

## Documents synthesized (2)

- **PRD** — docs/product/2026-07-23-deck-to-game-activation-prd.md (high confidence, locked: false)
  "vEDH Deck-to-Game Activation PRD". Authority for product intent: personas, user
  stories A1-A7, success metrics, non-goals, phased rollout, risks.
- **SPEC** — docs/plans/2026-07-23-deck-to-game-activation-tickets.md (medium confidence, locked: false)
  "vEDH Deck-to-Game Activation Implementation Tickets". Authority for technical
  contracts and sequencing: ACT-001..ACT-013, GraphQL, schema migrations, security.

Both are from the same 2026-07-23 workstream; the SPEC explicitly derives from the PRD.

## Decisions

- ADR-sourced decisions: **0** (no ADR documents in the ingest set)
- Locked decisions: **0** (both documents `locked: false`)
- Open decisions carried forward: **4** (OPEN-1 first deck provider; OPEN-2 guest
  expiry duration; OPEN-3 quiet-beta viewport target; OPEN-4 funnel query surface)

See `decisions.md`. Technical commitments live in `constraints.md`, not here, because
a PRD and a SPEC cannot produce ADR-status decisions without fabrication.

## Requirements (20)

Two layers, deliberately not merged. See `requirements.md`.

**Layer 1 — product (PRD), 7:**
REQ-A1-quick-start-without-registering, REQ-A2-import-deck-familiar-formats,
REQ-A3-join-from-invite-without-registering, REQ-A4-share-table-immediately,
REQ-A5-convert-guest-after-value, REQ-A6-understand-funnel-performance,
REQ-A7-recover-from-expected-failures.

**Layer 2 — implementation (SPEC), 13:**
REQ-ACT-001 through REQ-ACT-013, with ticket IDs, priorities, sizes, and `depends_on`
edges reproduced verbatim.

**Dependency graph (load-bearing — do not renumber or reorder):**

```
ACT-001  (P0, M)  depends: none
ACT-002  (P0, L)  depends: ACT-001
ACT-003  (P0, M)  depends: ACT-002
ACT-004  (P0, L)  depends: ACT-002
ACT-005  (P0, L)  depends: ACT-001
ACT-006  (P0, L)  depends: ACT-004, ACT-005
ACT-007  (P0, M)  depends: ACT-001
ACT-008  (P0, L)  depends: ACT-004, ACT-005, ACT-007
ACT-009  (P0, M)  depends: ACT-006, ACT-008
ACT-010  (P1, M)  depends: ACT-005, ACT-009
ACT-011  (P1, M)  depends: ACT-001, ACT-006
ACT-012  (P1, S)  depends: ACT-001, ACT-009
ACT-013  (P0, L)  depends: ACT-003 through ACT-010
```

Sizes from source: S = up to one focused day, M = two to three days, L = three to
five days including tests and review. Priority rule from source: P0 establishes
measurable guest deck-to-board activation; P1 improves post-value acquisition and
operating confidence.

Two notes that affect phase derivation:
- ACT-010 is P1 but is a declared dependency of the P0 release gate ACT-013, whose
  acceptance includes account claim. It is not deferrable past the MVP gate.
- ACT-003 may legitimately terminate in a documented no-go with paste-only
  activation. ACT-013 depends on it either way.

## Constraints (18)

See `constraints.md`. Breakdown by type:
- **api-contract (3)** — GraphQL contract additions; canonical deck parser grammar; public invite minimal response type.
- **schema (3)** — product_events table; guest user fields (`is_guest`, `expires_at`); migration parity across production and test schemas.
- **nfr (6)** — deck provider SSRF and network controls; guest token lifetime and authorization; product-event payload privacy allowlist; Prometheus metric cardinality; rate limiting on public surfaces; testing and release-gate coverage.
- **protocol (6)** — gqlgen generation policy; provider feature flag and kill switch; preserve existing game mutation authorization model; `board_ready` emission definition; account claim credential rules; scope guardrails.

No contradictions were found between the PRD and SPEC on any constraint.

## Context topics (13)

See `context.md`. PRD product framing the SPEC does not address, preserved rather
than dropped: objectives and delivery envelope; problem statement; proposed solution
shape; personas; success metrics; product event vocabulary; current-state findings;
non-goals; AI system requirements; phased rollout; delivery sequence (both versions);
technical risks; integration points; frontend change surface; related plans.

## Conflicts

**0 blockers, 0 competing variants, 8 auto-resolved (INFO).**

Full detail in `/Users/oberon/code/vedh/.planning/INGEST-CONFLICTS.md`.

Three genuine PRD/SPEC differences were resolved by precedence (SPEC wins):
1. **INFO-1** — whether a deck provider must ship in the MVP. SPEC makes it conditional on the ACT-003 feasibility gate.
2. **INFO-2** — guest account-claim phasing. PRD says v1.1; SPEC puts it inside the P0 release gate. Materially affects phase derivation.
3. **INFO-3** — safe invite preview lands in week 3 (SPEC), not week 2 (PRD).

Five informational findings: absent cross-refs are forward deliverables not
prerequisites (INFO-4); the PRD/SPEC mutual reference was evaluated and judged not a
synthesis cycle (INFO-5, deliberate deviation from the literal rule — see the entry);
code baseline verified against HEAD (INFO-6); no ADRs or locked docs (INFO-7); four
open decisions carried forward (INFO-8).

No PRD product framing was discarded in favor of the SPEC. The SPEC's silence on
personas, metrics, and non-goals was treated as silence, not disagreement.

## Files

- `/Users/oberon/code/vedh/.planning/intel/decisions.md`
- `/Users/oberon/code/vedh/.planning/intel/requirements.md`
- `/Users/oberon/code/vedh/.planning/intel/constraints.md`
- `/Users/oberon/code/vedh/.planning/intel/context.md`
- `/Users/oberon/code/vedh/.planning/INGEST-CONFLICTS.md`

## Status

READY — safe to route. No blockers, no user decision required before routing.

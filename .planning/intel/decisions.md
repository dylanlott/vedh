# Decisions

Source type: ADR. Ingest date: 2026-08-03. Mode: new.

## No ADR-sourced decisions in this ingest set

The ingest set contains zero ADR-classified documents (1 PRD, 1 SPEC). No entry in
this file carries `status: locked` or `status: proposed`, because neither of those
statuses can be derived from a PRD or a SPEC without fabricating an ADR that does
not exist.

Both ingested documents declare `locked: false`. No LOCKED-vs-LOCKED evaluation was
possible or performed.

Technical commitments that would normally live here (GraphQL contract, schema
migrations, SSRF controls, metric cardinality rules) were extracted to
`constraints.md` from the SPEC, which is their actual source.

---

## Open decisions carried forward (not ADR-backed)

The PRD contains an explicit "Open Decisions" section. These are recorded verbatim
because downstream planning depends on them, but they are unresolved and are NOT
decisions. They sit outside the `locked|proposed` ADR taxonomy and are marked
`status: open`.

### OPEN-1: First public deck provider
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- status: open
- decision: absent — "Which public deck provider passes the one-day feasibility spike and becomes the first supported URL source."
- scope: deck provider adapters, ACT-003
- resolution path: SPEC ACT-003 defines a one-day time-boxed feasibility comparison of public Archidekt and Moxfield deck access as the deciding mechanism. Both documents agree the question is open; neither answers it.

### OPEN-2: Guest expiry duration
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- status: open
- decision: absent — "Whether guest expiry is 24 hours or seven days; the token remains 24 hours either way."
- scope: guest identity, ACT-005
- resolution path: unresolved in both documents. SPEC ACT-005 fixes the guest *token* at 24 hours and adds an `expires_at` column for the backing guest user, but does not fix the backing-guest expiry duration. The two documents agree; neither resolves it.

### OPEN-3: Quiet-beta viewport target
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- status: open
- decision: absent — "Whether the first quiet beta targets desktop only or includes a minimum supported tablet breakpoint."
- scope: quiet beta rollout, ACT-004, ACT-011
- resolution path: unresolved. SPEC ACT-004 verification refers to "the minimum supported tablet width selected for beta", which presumes this decision is made elsewhere but does not make it.

### OPEN-4: Product-funnel query surface
- source: docs/product/2026-07-23-deck-to-game-activation-prd.md
- status: open
- decision: absent — "Which operational surface will be used for product-funnel queries before a dedicated internal dashboard exists."
- scope: analytics, ACT-012
- resolution path: SPEC ACT-012 partially narrows this ("Product-funnel queries use PostgreSQL; Grafana technical panels use Prometheus") and produces versioned SQL at `docs/analytics/deck-to-game-activation.sql`, but does not name an operational surface.

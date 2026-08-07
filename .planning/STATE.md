---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: milestone
current_phase: 01
current_phase_name: measured-deck-import-foundation
status: executing
stopped_at: Completed 01-08-PLAN.md
last_updated: "2026-08-07T22:25:53.797Z"
last_activity: 2026-08-07
last_activity_desc: Phase 01 execution started
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 10
  completed_plans: 8
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-03)

**Core value:** A person with a decklist reaches a working, shareable Commander board without registering — and every step of that path is measured.
**Current focus:** Phase 01 — measured-deck-import-foundation

## Current Position

Phase: 01 (measured-deck-import-foundation) — EXECUTING
Plan: 2 of 10
Status: Ready to execute
Last activity: 2026-08-07 — Phase 01 execution started

Progress: [████████░░] 80%

## Performance Metrics

**Velocity:**

- Total plans completed: 0
- Average duration: —
- Total execution time: 0.0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**

- Last 5 plans: —
- Trend: —

*Updated after each plan completion*
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 01 P01 | 13min | 3 tasks | 24 files |
| Phase 01 P02 | 37min | 2 tasks | 13 files |
| Phase 01 P03 | 8min | 2 tasks | 3 files |
| Phase 01 P04 | 31min | 2 tasks | 9 files |
| Phase 01 P05 | 55min | 2 tasks | 8 files |
| Phase 01 P06 | 2h | 3 tasks | 14 files |
| Phase 01 P07 | 45min | 3 tasks | 6 files |
| Phase 01 P08 | 55min | 2 tasks | 5 files |

## Accumulated Context

### Decisions

No ADR-locked decisions exist. The ingest set was one PRD and one SPEC, both
`locked: false`, zero ADRs. Technical commitments live as constraints in PROJECT.md.

Precedence resolutions applied at ingest that affect execution:

- INFO-1: a shipped deck provider is conditional on ACT-003's feasibility gate — paste-only is a legitimate outcome and must not block Phase 5
- INFO-2: guest account claim (ACT-010) is inside the MVP gate, not v1.1 — it lands in Phase 4, before Phase 5
- INFO-3: safe invite preview (ACT-007) sequences with guest join in Phase 3, not with the host flow
- [Phase ?]: previewDeck placed on Mutation (never Query) per locked api-contract; DeckPreview.BlockingErrors is a recorded additive extension
- [Phase ?]: pkg/telemetry.Vocabulary carries PRD fields with a dedicated product_events column on that column, never duplicated into metadata — a deliberate departure from 01-RESEARCH.md Pattern 5's assumed table
- [Phase ?]: Authoritative (six server-owned events) and deduplicated (four events in the dedup predicate) are kept as deliberately distinct, documented senses of the word
- [Phase ?]: Task 1's scanner.go rewrite deliberately does not depend on sections.go (Task 2's file); header words are recognized well enough to never become an Entry, with real drop/warning/preselect semantics added only in Task 2, keeping the tasks independently buildable.
- [Phase ?]: D-11's dropped-section header row is accounted for via its own summary warning (attributed to the header's SourceLine), not via DroppedSectionRows -- this is the assignment that makes the no-row-disappears invariant balance exactly.
- [Phase ?]: REQ-ACT-001 not marked complete: also claimed by plan 01-06 (not yet executed), following 01-01-SUMMARY.md precedent
- [Phase ?]: D-09/D-10 printing resolution: card_names is a plain incrementally-refreshed table (not a materialized view); GiST gist_trgm_ops chosen over GIN for card_names' trigram index because D-03 requires below-cutoff nearest matches; DeckPreviewEntry.SetCode/CollectorNumber/Category (declared since 01-01, never wired) now populated for every entry as a Rule 2 fix, since D-10's 'keeps the parsed set code on the entry' criterion is otherwise untestable at the GraphQL boundary.
- [Phase ?]: 01-05: The suggestion query's per-needle LATERAL orders by the GiST distance operator alone; a display-name tiebreak added inside it defeated the index and blew the 750ms sub-budget (measured: 190ms to 1.2s for 25 needles). The tiebreak lives only in the cheap outer sort.
- [Phase ?]: 01-05: applyPrintingMetadata overwrites a persisted library Card's SetCode/CollectorNumber/Category with what the player typed, extending D-10's reconciliation intent from the preview to the created library.
- [Phase ?]: 01-05 assigned fix: ensureFormatRules's per-load Turn.Phase re-normalization and per-player Life-zero re-defaulting were real regressions from b1ac894 (they ran on every existing-game load, not only at creation) -- removed; CreateGame's defaultLifeForAll heuristic (default life only when nobody in the call specified a nonzero life) replaces the per-player check since InputBoardState.Life is a required non-pointer Int with no wire-level 'unset' signal.
- [Phase ?]: 01-06: golang.org/x/time pinned to v0.14.0 (not the research-cited v0.15.0) to avoid an unrelated go.mod/toolchain bump against the pinned go1.24.2 and the +heroku goVersion go1.24 directive
- [Phase ?]: 01-06: previewDeck (SurfaceDeckImport) and trackProductEvent (SurfaceProductEvent) share one pkg/ratelimit.Registry built from a single Conf-derived budget (DECK_IMPORT_RATE_PER_MINUTE/BURST), since only one budget pair is specified anywhere in this phase's intel
- [Phase ?]: 01-06: DECK_PROVIDER_ENABLED defaults false and providerEnabled() also requires a non-empty DECK_PROVIDER_ALLOWED_HOSTS even when the flag is set; disabled/half-configured traffic never calls the fetch seam at all (proven via a spy)
- [Phase ?]: 01-07: Neutral comparison recommended Archidekt (D-15 hard gate observed and consistent); the user selected Moxfield anyway at the D-14 checkpoint, diverging from that recommendation -- recorded factually, not corrected
- [Phase ?]: 01-07: Task 3 executed as a Rule-2 deviation directed by the user -- hostname-keyed Moxfield adapter routing and kill-switch reuse shipped as scaffolding; normalizeToDeckText deliberately unimplemented (errMoxfieldContractUnverified) since api.moxfield.com/robots.txt disallows automated access and no authorized sample response has ever been captured
- [Phase ?]: 01-07: REQ-ACT-003 deliberately left NOT complete -- the requirement's substance (a working first-provider import) does not exist; marking it complete would overstate what shipped
- [Phase ?]: 01-08: D-14 reversed to Archidekt at UAT (gap G-01-1); Moxfield's response contract is unobtainable without authorized API access, so archidektAdapter was implemented against the observed contract and moxfieldAdapter/errMoxfieldContractUnverified/moxfieldHost were removed outright (re-addable later if authorized Moxfield access is obtained)
- [Phase ?]: 01-08: Human precondition recorded for enabling DECK_PROVIDER_ENABLED on archidekt.com -- a human must read https://archidekt.com/terms in a real browser first, since it is JS-rendered and no agent has JavaScript execution capability; building/testing the adapter does not require this

### Pending Todos

None yet.

### Blockers/Concerns

Three open decisions carried forward unresolved (`status: open`). Each is surfaced in
its owning phase for `/gsd-discuss-phase` — do not pre-answer:

- OPEN-2 (Phase 2, ACT-005): guest expiry 24 hours or seven days; token stays 24h either way
- OPEN-3 (Phase 2 ACT-004, revisited Phase 4 ACT-011): quiet beta desktop-only or a tablet breakpoint
- OPEN-4 (Phase 4, ACT-012): operational surface for product-funnel queries pre-dashboard

OPEN-1 (Phase 1, ACT-003) is resolved: the user selected Moxfield at the 01-07 D-14
checkpoint, diverging from the feasibility spike's own Archidekt recommendation. This is
not a fully closed matter, though: two implementation blockers remain before Moxfield can
be enabled for real (authorization for `api.moxfield.com`; a captured, authorized sample
response to build a real normalizer from) — see
`docs/research/deck-provider-feasibility.md` section 5 and
`.planning/phases/01-measured-deck-import-foundation/01-07-SUMMARY.md`. REQ-ACT-003
remains open in REQUIREMENTS.md pending those two blockers.

Baseline is an existing working codebase at `main@ef2732a` (verified = HEAD at ingest).
No greenfield scaffolding — this milestone changes the path to existing value.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| *(none)* | | | |

## Session Continuity

Last session: 2026-08-07T22:25:53.791Z
Stopped at: Completed 01-08-PLAN.md
Resume file: None

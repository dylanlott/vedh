---
phase: 01-measured-deck-import-foundation
plan: 07
subsystem: api
tags: [deck-import, moxfield, archidekt, ssrf, feasibility-spike, decision-record]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 06)
    provides: "server/deck_providers.go's newSafeProviderClient/fetchDeckProviderURL secure fetch client and providerEnabled()/DeckProviderEnabled kill switch, both reused unchanged; server/deck_import.go's previewDeckURL/deckProviderFetch seam this plan wires an adapter into"
  - phase: 01-measured-deck-import-foundation (plan 02)
    provides: "pkg/deckimport.Parse/ParseWithSource, the canonical parser every provider response must normalize through rather than a second grammar"
provides:
  - "docs/research/deck-provider-feasibility.md — the timeboxed neutral Archidekt/Moxfield comparison, a checked-in Archidekt fixture, and the recorded checkpoint decision (Moxfield, chosen by the user, diverging from the document's own Archidekt recommendation)"
  - "server/deck_providers.go's deckProviderAdapter/deckProviderAdapters/deckProviderAdapterFor — hostname-keyed adapter routing, and moxfieldAdapter/errMoxfieldContractUnverified — a deliberately unimplemented normalizer that fails closed rather than fabricating a Moxfield field mapping"
  - "server/deck_import.go's previewDeckURL rewired to route through the adapter registry as a second, independent gate beyond the host allowlist, before any fetch is attempted"
  - ".planning/phases/01-measured-deck-import-foundation/COVERAGE.md — every Moxfield capability recorded OPT-OUT, honestly reflecting scaffolding rather than a working import"
affects: [Phase 5 (the release runbook references DECK_PROVIDER_ENABLED as the concrete switch; this plan's Moxfield adapter cannot be turned on for real until the two open blockers in the decision record are resolved), any future phase that captures an authorized Moxfield sample response (fills in moxfieldAdapter.normalizeToDeckText)]

# Actuals (#2632)
actuals:
  tokens: 11400
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Hostname-keyed adapter registry (deckProviderAdapters/deckProviderAdapterFor) as a second, independent gate beyond the host allowlist — mirrors the separation-of-concerns shape safeControl (address policy) and newDeckProviderCheckRedirect (host/scheme policy) already established in plan 01-06."
    - "Deliberately unimplemented normalizer seam with a static sentinel error (errMoxfieldContractUnverified) rather than a fabricated field mapping, for a provider whose response contract has never been observed."

key-files:
  created:
    - server/testdata/deck_providers/archidekt_deck_2026-08-05.json
    - .planning/phases/01-measured-deck-import-foundation/COVERAGE.md
  modified:
    - docs/research/deck-provider-feasibility.md
    - server/deck_providers.go
    - server/deck_providers_test.go
    - server/deck_import.go

key-decisions:
  - "Task 1 (prior session): neutral comparison recommended Archidekt over Moxfield, tied to D-15's hard gate (response-shape stability) — Archidekt's shape was directly observed and structurally consistent across six fetches; Moxfield's could not be evaluated at all (api.moxfield.com/robots.txt blanket-disallows automated access; moxfield.com's web frontend returned a Cloudflare bot-management 403 on every probe across three User-Agents)."
  - "Task 2 (this session): the user selected Moxfield anyway, diverging from Task 1's recommendation. Recorded factually in the decision record as the user's call, not a correction to the neutral comparison, which is preserved unchanged."
  - "Task 3 (this session): executed as a documented Rule-2 deviation from the plan's anticipated Task 3 shape. The user's explicit direction at the checkpoint was 'build skeleton, contract marked unverified' rather than a fixture-tested working adapter (impossible — no Moxfield response was ever observed) or a no-go (the user chose a provider). Shipped hostname-keyed adapter routing and kill-switch reuse from plan 01-06, with the normalizer (moxfieldAdapter.normalizeToDeckText) deliberately left unimplemented, returning a static sentinel error rather than a guessed field mapping."
  - "REQ-ACT-003 is NOT marked complete (see Next Phase Readiness) — the requirement's substance (a working first-provider import) does not exist yet; marking it complete would overstate what shipped, which the checkpoint's own instructions explicitly warned against."

requirements-completed: []

coverage:
  - id: D1
    description: "A written, neutral decision record evaluates both Archidekt and Moxfield against the same six criteria, states a recommendation tied to the hard gate condition, and records the user's actual checkpoint selection (Moxfield) including its divergence from that recommendation and the two open blockers gating enablement"
    requirement: "REQ-ACT-003"
    verification:
      - kind: other
        ref: "docs/research/deck-provider-feasibility.md sections 1-5 (test -s and grep -qi archidekt/moxfield/response-shape all pass; see plan verify block)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Hostname-keyed adapter routing (deckProviderAdapters/deckProviderAdapterFor) reaches the plan 01-06 secure fetch client for a registered, allowlisted host, and refuses an allowlisted-but-unregistered host before any fetch is attempted — proving the routing decision is real, not a relabeled stub"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestProvider_DeckProviderAdapterFor"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_MoxfieldAdapterRoutesThenFailsClosed"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_UnregisteredHostNeverFetched"
        status: pass
    human_judgment: false
  - id: D3
    description: "moxfieldAdapter's normalizer deliberately refuses to guess a field mapping for any body it is handed, returning a static sentinel error that carries no provider response content, mapped onto the ordinary paste-fallback message before ever reaching a client"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestProvider_MoxfieldContractUnverified"
        status: pass
    human_judgment: false
  - id: D4
    description: "The provider scaffold does not regress plan 01-06's stability guarantees: with the kill switch off, a moxfield.com deck URL still fails closed with zero fetch attempts, and a pasted decklist is entirely unaffected regardless of the scaffold's presence"
    requirement: "REQ-ACT-003"
    verification:
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_MoxfieldScaffoldPathIsStable"
        status: pass
    human_judgment: false
  - id: D5
    description: "Every locked network control (deny-prefix count, 3s/8s timeout budgets, exact-equality host match, redirect bound) is unchanged from plan 01-06, and the full secure-client and pkg/ratelimit test suites pass unchanged"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v (all pass)"
        status: pass
      - kind: unit
        ref: "go test ./pkg/... -race (all pass)"
        status: pass
    human_judgment: false
  - id: D6
    description: "COVERAGE.md records the external-API coverage decision honestly: every Moxfield capability is OPT-OUT with a stated reason, reflecting that this ships as off-by-default scaffolding rather than a working import"
    requirement: "REQ-ACT-003"
    verification:
      - kind: other
        ref: ".planning/phases/01-measured-deck-import-foundation/COVERAGE.md"
        status: pass
    human_judgment: false

duration: ~45min (this continuation session; Task 1 ran in a prior session)
completed: 2026-08-05
status: complete
---

# Phase 1 Plan 7: Deck Provider Feasibility Decision and Moxfield Adapter Scaffold Summary

**A neutral Archidekt/Moxfield feasibility comparison recommended Archidekt, the user chose Moxfield anyway at the D-14 checkpoint, and Task 3 shipped hostname-keyed adapter routing plus a deliberately unimplemented, fail-closed normalizer — real plumbing behind the existing default-off kill switch, not a working Moxfield import.**

## Performance

- **Duration:** ~45min for this continuation (Task 2 + Task 3); Task 1 (the feasibility spike and Archidekt fixture capture) was completed in a prior session and is not re-timed here.
- **Completed:** 2026-08-05
- **Tasks:** 3/3 (Task 1 completed and committed in the prior session at `6344381`; Tasks 2 and 3 completed and committed in this session)
- **Files modified:** 6 across the whole plan (2 created in Task 1, 1 modified in Task 2, 3 created/modified in Task 3)

## Accomplishments

- `docs/research/deck-provider-feasibility.md` evaluates Archidekt and Moxfield against the same six criteria (public read path, private-deck behaviour, rate-limit signals, response-shape stability, terms of service, fixture maintainability), ties its recommendation to D-15's hard gate condition, and names Archidekt as the candidate whose stability could actually be evaluated with observed evidence.
- The user selected Moxfield at the checkpoint anyway. The decision record's section 5 records that divergence factually — attributed to the user, not to Task 1's own recommendation — alongside two explicit open blockers: (a) `api.moxfield.com/robots.txt`'s blanket `Disallow: /`, meaning authorized access is required before any real request may be made; (b) the response contract has never been observed, so no field mapping exists to build a working normalizer from.
- `server/deck_providers.go` gained hostname-keyed adapter routing (`deckProviderAdapter`/`deckProviderAdapters`/`deckProviderAdapterFor`) as a second, independent gate beyond `DECK_PROVIDER_ALLOWED_HOSTS`: a host that is allowlisted but has no registered adapter is still refused before any fetch, proven by `TestProvider_UnregisteredHostNeverFetched`.
- `moxfieldAdapter` is deliberately incomplete: `normalizeToDeckText` always returns `errMoxfieldContractUnverified`, a static sentinel carrying no provider response content, regardless of what body it is handed (`TestProvider_MoxfieldContractUnverified` proves this for nil, empty, and plausible-JSON inputs alike). This is a documented refusal to guess, not an oversight.
- `server/deck_import.go`'s `previewDeckURL` now genuinely reaches the secure client and the adapter's normalizer for a recognized Moxfield URL (`TestProvider_MoxfieldAdapterRoutesThenFailsClosed` proves the fetch spy is called exactly once), and still falls back to the same product-language paste message on the normalizer's failure — the end-user outcome is unchanged from before this task, reached now by a real routing decision instead of an unconditional stub.
- Plan 01-06's stability guarantees are intact: with the kill switch off, the Moxfield URL path still makes zero fetch attempts (`TestProvider_MoxfieldScaffoldPathIsStable/KillSwitchOffNeverFetches`), and a pasted decklist is entirely unaffected by this scaffold's presence (`TestProvider_MoxfieldScaffoldPathIsStable/PasteUnaffected`).
- `.planning/phases/01-measured-deck-import-foundation/COVERAGE.md` records every enumerated Moxfield capability (read public deck, read private deck, list user decks, read deck metadata, read commander designation, search decks) as OPT-OUT with a stated reason, rather than overstating what is actually usable today.
- Every locked network control from plan 01-06 (24-entry deny-prefix list, 3s connect/8s total timeouts, exact-equality host match, 3-hop redirect bound, 1 MiB cap) is byte-unchanged; `go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v`, the full `go test ./server -count=1` suite, and `go test ./pkg/... -race` are all green with zero failures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Timeboxed neutral feasibility comparison and decision record** (prior session) - `6344381` (docs)
2. **Task 2: Record the user's Moxfield selection at the D-14 checkpoint** (this session) - `29affa1` (docs)
3. **Task 3: Moxfield adapter scaffold behind the existing kill switch** (this session) - `2f31820` (feat)

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `docs/research/deck-provider-feasibility.md` — Task 1's neutral comparison, extended in Task 2 with section 5, the recorded checkpoint decision and its two open blockers
- `server/testdata/deck_providers/archidekt_deck_2026-08-05.json` — Task 1's checked-in real Archidekt response capture (redacted account fields only)
- `server/deck_providers.go` — Task 3: `deckProviderAdapter`, `moxfieldAdapter`, `errMoxfieldContractUnverified`, `moxfieldHost`, `deckProviderAdapters`, `deckProviderAdapterFor`
- `server/deck_import.go` — Task 3: `previewDeckURL` rewired to consult the adapter registry before fetching and to call the adapter's normalizer on a successful fetch
- `server/deck_providers_test.go` — Task 3: `TestProvider_MoxfieldContractUnverified`, `TestProvider_DeckProviderAdapterFor`, `TestProvider_MoxfieldAdapterRoutesThenFailsClosed`, `TestProvider_UnregisteredHostNeverFetched`, `TestProvider_MoxfieldScaffoldPathIsStable`
- `.planning/phases/01-measured-deck-import-foundation/COVERAGE.md` — Task 3: capability matrix, every row OPT-OUT with a stated reason

## Decisions Made

See `key-decisions` in frontmatter for full rationale on each:
- Task 1's neutral comparison recommended Archidekt over Moxfield, tied to the hard gate condition (response-shape stability); Moxfield's could not be evaluated at all within the spike's bounds.
- The user selected Moxfield anyway at the checkpoint — recorded factually as the user's call, with the neutral recommendation left unchanged as the historical record of what the comparison actually found.
- Task 3 was executed as a documented Rule-2 deviation (see below): the plan's own anticipated Task 3 shapes (a fixture-tested working adapter, or a no-go) did not fit the actual checkpoint outcome (a provider selected whose contract could not be observed), so the user was asked directly and chose "build skeleton, contract marked unverified."
- REQ-ACT-003 is **not** marked complete. See Next Phase Readiness.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Functionality, directed by explicit user instruction] Task 3 executed as a scaffold rather than either plan-anticipated shape**

- **Found during:** Task 3, before writing any code — the plan's `<action>` text anticipates exactly two outcomes ("If the checkpoint selected a provider" → a fixture-tested working adapter; "If the checkpoint recorded a no-go" → a closing section and `TestProvider_NoGoPathIsStable`). The actual checkpoint outcome (a provider selected whose response contract can never be observed within this spike's stated constraints) fits neither shape: a fixture-tested adapter is impossible without a real response to test against, and a no-go section would misrepresent what the user actually chose.
- **Resolution:** the user was asked directly and chose explicitly: "build skeleton, contract marked unverified" over inventing a fabricated field mapping or refusing to build anything. This is the shape this task actually shipped — hostname-keyed routing and kill-switch reuse are real and tested; the normalizer is a documented, fail-closed placeholder, never a guess.
- **Files modified:** `server/deck_providers.go`, `server/deck_providers_test.go`, `server/deck_import.go`, `.planning/phases/01-measured-deck-import-foundation/COVERAGE.md`, `docs/research/deck-provider-feasibility.md` (section 5)
- **Verification:** `go build ./... && go vet ./server`; `go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v` (all pass); `go test ./server -count=1` (all pass); `go test ./pkg/... -race` (all pass); `go mod verify` (clean).
- **Committed in:** `2f31820` (Task 3 commit)

---

**Total deviations:** 1 (Rule 2, directed by explicit user instruction rather than autonomously auto-fixed — recorded here per this plan's deviation-documentation requirement even though the specific resolution was user-directed, not agent-inferred).
**Impact on plan:** The plan's own two anticipated Task 3 shapes did not cover the actual checkpoint outcome. No scope creep beyond what the user explicitly authorized; no security control weakened; no fabricated data introduced anywhere.

## Known Stubs

| File | Line | Stub | Reason |
|------|------|------|--------|
| `server/deck_providers.go` | ~395 (`moxfieldAdapter.normalizeToDeckText`) | Always returns `errMoxfieldContractUnverified` regardless of input body — no Moxfield field mapping is implemented | The response contract was never observed: `api.moxfield.com/robots.txt` blanket-disallows automated access, and no authorized sample response exists. Implementing a real mapping requires resolving both open blockers in `docs/research/deck-provider-feasibility.md` section 5 first. This is a deliberate, documented fail-closed placeholder, not an oversight — it is the entire point of this task's deviation. |

This stub is also recorded in `.planning/WINDOWS.md` (see below) so it stays visible at ship time.

## Issues Encountered

None beyond the one documented deviation above.

## User Setup Required

None for this plan's code. The `DECK_PROVIDER_API_KEY`/`DECK_PROVIDER_ENABLED`/`DECK_PROVIDER_ALLOWED_HOSTS` entry in this plan's `user_setup` frontmatter is **not applicable**: no app-level credential is needed (Moxfield's public read path, if it existed for this codebase's purposes, requires no API key per the decision record's section 2.8 finding of "not established" either way), and — more importantly — the kill switch must stay at its default (`DECK_PROVIDER_ENABLED=false`) in every real deployment until both open blockers in the decision record are resolved. Turning it on for Moxfield today would enable a code path that always fails closed to the paste-fallback message; it would not enable a working import.

## Next Phase Readiness

- **REQ-ACT-003 is deliberately NOT marked complete.** The requirement's substance — a working first-provider import — does not exist. What was delivered: the neutral comparison, the recorded decision, the hostname-keyed adapter interface, the security controls (all inherited unchanged from plan 01-06), and the kill switch. What remains: an authorized Moxfield API sample response, from which `moxfieldAdapter.normalizeToDeckText` can be implemented for real. This mirrors plan 01-06-SUMMARY's own precedent of leaving REQ-ACT-003 open rather than overstating partial delivery, and follows this plan's explicit checkpoint instruction not to overstate what shipped.
- This phase's fifth success criterion ("a written decision record names the first supported provider... or documents a reasoned no-go... either outcome satisfies this phase") **is** satisfied: the decision record exists, names Moxfield, and records why it is not yet usable.
- Phase 5's release runbook has its concrete switch (`DECK_PROVIDER_ENABLED`, still defaulting to `false`) to reference, unchanged from plan 01-06, regardless of REQ-ACT-003's traceability status.
- Before Moxfield can ever be enabled for real: (a) obtain authorization for `api.moxfield.com` access (a credentialed developer program, if Moxfield operates one, was not discovered by Task 1's spike); (b) capture a real, authorized sample response and implement `moxfieldAdapter.normalizeToDeckText` against it, then add the fixture-contract test this plan's original Task 3 text anticipated; (c) revisit `COVERAGE.md`'s OPT-OUT rows against that real contract.
- No blockers to closing Phase 1 itself: `go build ./...`, `go vet ./server`, `go test ./server -count=1`, `go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v`, `go test ./pkg/... -race`, and `go mod verify` are all green.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-05*

## Self-Check: PASSED

All 7 files (docs/research/deck-provider-feasibility.md,
server/testdata/deck_providers/archidekt_deck_2026-08-05.json,
server/deck_providers.go, server/deck_import.go,
server/deck_providers_test.go,
.planning/phases/01-measured-deck-import-foundation/COVERAGE.md,
.planning/phases/01-measured-deck-import-foundation/01-07-SUMMARY.md) exist
on disk, and all four commit hashes (`6344381`, `29affa1`, `2f31820`,
`dff1d80`) are present in `git log --oneline --all`.

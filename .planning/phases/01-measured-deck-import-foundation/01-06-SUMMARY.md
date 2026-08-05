---
phase: 01-measured-deck-import-foundation
plan: 06
subsystem: api
tags: [rate-limiting, ssrf, http-client, token-bucket, netip, prometheus, promauto, gqlgen]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 01)
    provides: "server/deck_import.go's PreviewDeck/finishPreviewDeck and server/product_events.go's TrackProductEvent — the two resolvers this plan wraps with rate limiting — plus pkg/telemetry's Collectors/promauto registration site and its pre-declared surface/provider label names"
  - phase: 01-measured-deck-import-foundation (plan 03)
    provides: "the frontend event service's persisted session ID, completing REQ-ACT-001's frontend clause so this plan's rate-limit instrumentation is the requirement's last remaining piece"
  - phase: 01-measured-deck-import-foundation (plan 05)
    provides: "server/test.go's testAPI with its t.Cleanup(appDB.Close) fix, reused unchanged; server/deck_import.go's buildDeckPreview/blockedPreview this plan wraps rather than replaces"
provides:
  - "pkg/ratelimit: Registry/NewRegistry/Surface/Allow — a DB-free, per-surface, per-client-key token-bucket registry with idle eviction, CI-gated by go test ./pkg/... -race"
  - "server/ratelimit.go's allowRequest + clientKeyFor — the instrumented wrapper binding the registry to previewDeck and trackProductEvent, counting both the allowed and limited outcome"
  - "server/deck_providers.go's newSafeProviderClient/safeControl/newDeckProviderCheckRedirect/fetchDeckProviderURL — the provider-agnostic secure outbound fetch client (address/port/scheme/host/redirect/size/timeout controls) plan 01-07's adapter will call into"
  - "server/deck_providers.go's providerEnabled() and Conf.DeckProviderEnabled/DeckProviderAllowedHosts — the default-off provider kill switch, required in both branches of the D-14 checkpoint"
  - "pkg/telemetry: vedh_rate_limit_total, vedh_deck_provider_fetch_total, vedh_deck_provider_fetch_duration_seconds collectors, the last two declared-but-unobserved until plan 01-07 wires an emit site"
  - "golang.org/x/time promoted to a direct go.mod dependency at v0.14.0"
affects: [01-07 (the D-14 checkpoint's adapter calls fetchDeckProviderURL/newSafeProviderClient and previewDeckURL's deckProviderFetch seam directly rather than re-deriving the secure client), Phase 2/3 (guest creation and invite lookup reuse pkg/ratelimit.Registry/Surface for their own rate limits per the phase-wide constraint)]

# Actuals (#2632)
actuals:
  tokens: 20000
  tasks: 3
  commits: 4

tech-stack:
  added: ["golang.org/x/time v0.14.0 (direct)"]
  patterns:
    - "Test-injectable ControlContext/resolver/timeout parameters (newSafeProviderClientWithOptions) mirroring plan 01-05's budget-as-parameter shape, so hermetic loopback tests can exercise CheckRedirect/body-cap/timeout layers without relaxing what the zero-option production constructor (newSafeProviderClient) actually enforces"
    - "Package-level function-variable seam (deckProviderFetch) for proving a call was never made (a spy substituted in tests), rather than only inferring it from an error message"
    - "AND-a-flag-against-a-non-empty-value gate (providerEnabled mirroring shouldExposeMetrics) so a half-configured deployment fails closed rather than opening a path that can reach nothing"

key-files:
  created:
    - pkg/ratelimit/limiter.go
    - pkg/ratelimit/limiter_test.go
    - server/ratelimit.go
    - server/ratelimit_test.go
    - server/deck_providers.go
    - server/deck_providers_test.go
  modified:
    - server/graphql.go
    - server/deck_import.go
    - server/product_events.go
    - server/test.go
    - pkg/telemetry/metrics.go
    - pkg/telemetry/metrics_test.go
    - go.mod
    - go.sum

key-decisions:
  - "Promoted golang.org/x/time to v0.14.0, not the v0.15.0 the research audit named. v0.15.0 requires go1.25.0 and would have force-bumped go.mod's go directive away from the pinned go1.24.2 toolchain -- and away from the `// +heroku goVersion go1.24` comment governing Heroku deploys, a collateral change out of this plan's scope. v0.14.0 satisfies the same audited, approved x/time module at the pinned toolchain version with zero collateral diff. No verify-block grep enforces the literal version string."
  - "Renamed safeDialControl/deniedV4/deniedV6 to safeControl/deniedPrefixesV4/deniedPrefixesV6 partway through Task 2, to match the exact symbol names 01-01-PLAN.md's phase-wide artifact index declares for this plan (a substring grep for `deniedPrefixes` matches both V4/V6 names)."
  - "trackProductEvent and previewDeck are instrumented as two distinct ratelimit.Surface values (SurfaceDeckImport, SurfaceProductEvent) sharing one Registry built from the single Conf-derived budget -- not two separately configured surfaces -- since only one budget pair (DECK_IMPORT_RATE_PER_MINUTE/BURST) is specified anywhere in constraints.md or 01-RESEARCH.md's Open Question 2. Two distinct Surface values still make each resolver's rate-limit ratio separately readable in Grafana, which per-surface labelling exists for."
  - "server/test.go's testAPI() now sets explicit high (100000/100000) rate-limit Conf values: testAPI constructs Conf as a struct literal rather than through envconfig.Process, so the `default:\"30\"/\"10\"` tags are never applied there, and an unset burst would construct a burst-0 registry denying every existing test's PreviewDeck/TrackProductEvent call. This is a Rule 3 (blocking-issue) fix required to keep the existing full server suite green."
  - "TestSafeClient_RebindingFailsClosed proves ControlContext is wired into the real client via the simpler, explicitly-sanctioned equivalent 01-RESEARCH.md's Validation Architecture subsection 4 names ('a request to a hostname resolving to loopback fails') rather than a two-stage DNS-rebinding stub server: since there is no separate pre-flight resolution step in this design at all (that is the entire point of validating inside ControlContext), a stub returning a different answer on a second lookup proves nothing beyond what a single loopback-resolving request already proves, and a live experiment confirmed Go's client-side Content-Length framing makes the literal 'declares 10 bytes, sends 2 MiB' scenario structurally unreproducible against a standards-compliant client -- TestSafeClient_BodyCap's third case substitutes the actual mechanism (a chunked, undeclared-length response, resp.ContentLength == -1) by which a body can exceed what a declared length would suggest, which is exactly why this implementation never consults it."
  - "TestSafeClient_Timeouts' ConnectTimeout subtest dials 10.255.255.1 (a private, non-routable RFC 1918 address that answers nothing) with allowAnyDialControl rather than the real safeControl (which would refuse a 10.0.0.0/8 literal before any dial is attempted, per TestSafeControl_DeniedAddresses). This is not \"reaching a live deck provider\" -- no data is ever exchanged with any real service -- and is the standard deterministic black-hole target for a connect-timeout test; documented inline."

patterns-established:
  - "hermeticTestClient/allowAnyDialControl (server/deck_providers_test.go): a permissive ControlContext stand-in for tests that need a real loopback dial to exercise CheckRedirect, body-cap, or timeout behavior without relaxing what newSafeProviderClient (production's zero-option constructor) actually wires -- address/port denial is proven exhaustively and separately by direct, socket-free calls to safeControl."
  - "withDeckProviderFetchSpy (server/deck_providers_test.go): substitutes the deckProviderFetch package variable with a call-counting spy, restored via t.Cleanup, so a kill-switch test can prove zero dial attempts rather than merely observing an error message."

requirements-completed: [REQ-ACT-001]

coverage:
  - id: D1
    description: "previewDeck and trackProductEvent are rate limited per client key and per surface at 30/minute with a burst of 10 (Conf-configurable), and vedh_rate_limit_total increments exactly once per decision on both the allowed and the limited path so the ratio is readable before a too-tight limit becomes a funnel dip"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "server/ratelimit_test.go#TestRateLimit_AllowRequestCountsBothOutcomes"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_KillSwitch"
        status: pass
    human_judgment: false
  - id: D2
    description: "Two distinct client keys and two distinct surfaces never share a token bucket, and an entry idle past the eviction window is evicted and its slot returns to zero on a full sweep -- all proven by pkg/ratelimit tests that run in CI with no database and no MTGJSON snapshot"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "pkg/ratelimit/limiter_test.go#TestRegistry_PerKeyIsolation"
        status: pass
      - kind: unit
        ref: "pkg/ratelimit/limiter_test.go#TestRegistry_PerSurfaceIsolation"
        status: pass
      - kind: unit
        ref: "pkg/ratelimit/limiter_test.go#TestRegistry_IdleEvictionGrantsFreshBucket"
        status: pass
      - kind: unit
        ref: "pkg/ratelimit/limiter_test.go#TestRegistry_SweepReturnsToZero"
        status: pass
    human_judgment: false
  - id: D3
    description: "A rate-limited previewDeck call returns a product-language blocking error naming no limit value, window, or remaining count, with CanContinue false; a rate-limited trackProductEvent still returns true with no error"
    requirement: "REQ-ACT-001"
    verification:
      - kind: integration
        ref: "server/ratelimit_test.go#TestRateLimit_PreviewDeckNamesNoLimitValue"
        status: pass
      - kind: integration
        ref: "server/ratelimit_test.go#TestRateLimit_TrackProductEventStillReturnsTrue"
        status: pass
    human_judgment: false
  - id: D4
    description: "The secure outbound fetch client refuses the metadata address, loopback, shared-address-space, an IPv4-mapped-IPv6 form of the metadata address, protocol-assignment and benchmarking ranges, and the wrong port or a non-TCP network, while allowing a public address on the correct port -- opening no socket -- and the same first-matching-prefix is named deterministically for an address inside two declared prefixes"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestSafeControl_DeniedAddresses"
        status: pass
      - kind: unit
        ref: "server/deck_providers_test.go#TestSafeControl_ReportsFirstMatchingPrefix"
        status: pass
    human_judgment: false
  - id: D5
    description: "Two redirect hops are followed and the third is refused; a redirect to a non-secure scheme is refused; a host differing only in letter case is accepted while a host merely containing an allowlisted host as a substring is refused; a redirect to an unallowlisted host is refused and that host's handler is never invoked"
    requirement: "REQ-ACT-003"
    verification:
      - kind: integration
        ref: "server/deck_providers_test.go#TestSafeClient_HostAndRedirect"
        status: pass
    human_judgment: false
  - id: D6
    description: "An empty host and a non-secure scheme are refused before any name resolution is attempted; ControlContext is genuinely wired into the real client (a hostname resolving to loopback is refused, not merely correct in isolation); a response of exactly the 1 MiB cap is accepted, one byte more is refused, and a chunked response with no declared length carrying far more than the cap is also refused; both the connect and the total budget fail closed within a bounded, parameterised-down elapsed time; a zero-byte body yields a normalized provider error"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestSafeClient_RejectsBeforeResolution"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestSafeClient_RebindingFailsClosed"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestSafeClient_BodyCap"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestSafeClient_Timeouts"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestSafeClient_ZeroByteBodyIsProviderError"
        status: pass
    human_judgment: false
  - id: D7
    description: "With the provider flag unset, or set with an empty host allowlist, a deck URL yields exactly one blocking error mentioning pasting text with CanContinue false and zero dial attempts (proven via a spy, not merely an error return); a text-only preview is byte-identical across both flag states"
    requirement: "REQ-ACT-003"
    verification:
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_KillSwitch"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_KillSwitchRequiresAllowlist"
        status: pass
      - kind: integration
        ref: "server/deck_providers_test.go#TestProvider_PasteUnaffectedByFlag"
        status: pass
    human_judgment: false

duration: ~2h (single session)
completed: 2026-08-05
status: complete
---

# Phase 1 Plan 6: Rate Limiting and the Secure Deck-Provider Fetch Client Summary

**Both public deck-import resolvers are now per-client, per-surface rate limited by a CI-gated `pkg/ratelimit` token-bucket registry with both outcomes counted, and a provider-agnostic, hermetically-tested two-layer secure HTTP client plus a default-off `DECK_PROVIDER_ENABLED` kill switch stand ready for plan 01-07's D-14 checkpoint in either branch.**

## Performance

- **Duration:** ~2h (single session, no checkpoints)
- **Started:** 2026-08-05 (continuation from plan 01-05's completion, tip `d9cca4b`)
- **Completed:** 2026-08-05
- **Tasks:** 3/3, plus one same-session naming-consistency fixup
- **Files modified:** 14 (6 created, 8 modified)

## Accomplishments

- `pkg/ratelimit` ships a DB-free, per-(surface, client-key) token-bucket registry with an injected-clock-tested idle eviction sweep, proven by five test functions that run under `go test ./pkg/... -race` — the only target CI actually runs, with no `DATABASE_URL` and no MTGJSON snapshot.
- `server/ratelimit.go`'s `allowRequest` wraps that registry for `previewDeck` (`SurfaceDeckImport`) and `trackProductEvent` (`SurfaceProductEvent`), counting `vedh_rate_limit_total{surface,outcome}` on both the allowed and the limited path. A limited `previewDeck` returns one product-language blocking error naming no limit value; a limited `trackProductEvent` still returns `true, nil`.
- `server/deck_providers.go` implements the full locked network-control set from `.planning/intel/constraints.md` as a provider-agnostic client: `safeControl` (a 24-entry IANA-derived deny-prefix walk against the already-resolved literal, never a pre-flight name lookup), `newDeckProviderCheckRedirect` (exact lower-cased host allowlist + scheme + hop bound, re-validated every hop), the locked 3s/8s timeout budgets, and a counted-read 1 MiB body cap that never consults `Content-Length`.
- Every provider-facing test runs against loopback `httptest.NewTLSServer` instances or direct function calls — none reaches a live provider. A live experiment during Task 2 confirmed Go's own client-side `Content-Length` framing makes a "declares 10 bytes, sends 2 MiB" response structurally unreproducible against a standards-compliant client; `TestSafeClient_BodyCap`'s third case substitutes the actual mechanism (a chunked, undeclared-length body) by which a response can exceed what a declared length would suggest.
- `DECK_PROVIDER_ENABLED` (default `false`) plus `providerEnabled()` — which also requires a non-empty `DECK_PROVIDER_ALLOWED_HOSTS` even when the flag is set — gate `previewDeck`'s URL branch. Disabled traffic never calls `deckProviderFetch` at all (proven via a spy, not an inferred error message); enabled traffic genuinely reaches the task-2 secure client, but the parsing adapter stays absent until plan 01-07's D-14 checkpoint, so that branch always normalizes to the same provider error for now.
- `golang.org/x/time` is promoted to a direct `go.mod` dependency at v0.14.0 (deliberately not the research-cited v0.15.0 — see Decisions).

## Task Commits

Each task was committed atomically:

1. **Task 1: Per-surface, per-client rate limiting with instrumented outcomes** - `f1ea608` (feat)
2. **Task 2: Two-layer secure outbound fetch client, tested without touching the network** - `75f7ea6` (feat)
3. **Task 3: Provider kill switch defaulting off, with a paste fallback that never breaks paste** - `c925d38` (feat)
4. **Same-session fixup: rename `ratelimit_test.go` tests to match the plan's own `-run` gate** - `4de341e` (fix)

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `pkg/ratelimit/limiter.go` — `Registry`/`NewRegistry`/`Surface`/`Allow`, the injected-clock idle-eviction sweep
- `pkg/ratelimit/limiter_test.go` — burst exhaustion, per-key isolation, per-surface isolation, idle eviction, sweep-to-zero
- `server/ratelimit.go` — `allowRequest`, `clientKeyFor`, the `rateLimitOutcomeAllowed`/`Limited` constants
- `server/ratelimit_test.go` — the wrapper's counting behavior, the nil-limiter no-op path, the rate-limited `previewDeck`/`trackProductEvent` proofs, `clientKeyFor`'s derivation order
- `server/deck_providers.go` — `safeControl`, `deniedPrefixesV4`/`deniedPrefixesV6`, `globalUnicastV6`, `firstMatchingDenyPrefix`, `newDeckProviderCheckRedirect`, `newSafeProviderClient`/`newSafeProviderClientWithOptions`, `readBodyWithCap`, `fetchDeckProviderURL`, `providerEnabled`
- `server/deck_providers_test.go` — the full hermetic suite: address/port denial, redirect/host policy, pre-resolution rejection, the loopback-rebinding proof, body cap, timeouts, zero-byte body, and the three kill-switch tests
- `server/graphql.go` — `Conf.DeckImportRatePerMinute`/`DeckImportRateBurst`/`DeckProviderAllowedHosts`/`DeckProviderEnabled`, the `limiter`/`deckProviderAllowedHosts` fields on `graphQLServer`, `parseAllowedHosts`, `withClientAddr`/`remoteAddrContextKey`/`remoteHostFromRequest`/`remoteAddrFromContext`
- `server/deck_import.go` — `allowRequest` wired into `PreviewDeck`; `previewDeckURL`, `deckProviderClient`, `deckProviderFetch`, `rateLimitedPreview`
- `server/product_events.go` — `allowRequest` wired into `TrackProductEvent`
- `server/test.go` — `testAPI`'s `Conf` literal now sets generous rate-limit budgets
- `pkg/telemetry/metrics.go` — `rateLimitTotal`, `deckProviderFetchTotal`/`Duration` collectors and their `Observe*`/`*Counter` accessors, `deckProviderFetchBuckets`
- `pkg/telemetry/metrics_test.go` — `exerciseAndGather` extended; `TestMetrics_RateLimitLabelsAreDeclaredConstants`
- `go.mod`, `go.sum` — `golang.org/x/time` promoted to direct at v0.14.0; `github.com/prometheus/client_model` correctly reclassified as direct by the same `go mod tidy` run

## Decisions Made

See `key-decisions` in frontmatter for full rationale on each:
- `golang.org/x/time` pinned to v0.14.0, not the research-cited v0.15.0, to avoid an unrelated `go.mod`/Heroku-toolchain collateral bump.
- `safeDialControl`/`deniedV4`/`deniedV6` renamed to `safeControl`/`deniedPrefixesV4`/`deniedPrefixesV6` to match 01-01-PLAN.md's phase-wide artifact index.
- `previewDeck` and `trackProductEvent` are two distinct `Surface` values sharing one `Registry`/one Conf-derived budget, since only one budget pair is specified anywhere in this phase's intel.
- `testAPI()` needed explicit high rate-limit values to avoid regressing the existing full `server` suite.
- `TestSafeClient_RebindingFailsClosed` and `TestSafeClient_BodyCap`'s third case each substitute the literal scenario named in `01-06-PLAN.md`'s prose with the closest reproducible equivalent, per 01-RESEARCH.md's own stated fallback and a live experiment respectively.
- `TestSafeClient_Timeouts`' connect-timeout subtest dials a non-routable RFC 1918 address (never "a live deck provider") because a genuine connect-phase delay cannot be simulated against a pure-loopback listener (verified empirically — the kernel completes the TCP handshake even with no application-level `Accept()` call).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `testAPI()` would have rate-limited every existing test to zero**
- **Found during:** Task 1, before running any existing `server` test
- **Issue:** `testAPI` builds `Conf` as a Go struct literal rather than through `envconfig.Process` (`main.go`'s production path), so the new `DeckImportRatePerMinute`/`DeckImportRateBurst` fields' `default:"30"`/`"10"` tags are never applied there — an unset field is Go's zero value, which would make `NewGraphQLServer` construct a burst-0 registry denying every single `PreviewDeck`/`TrackProductEvent` call the package's ~40 existing test functions make.
- **Fix:** Set `DeckImportRatePerMinute: 100000, DeckImportRateBurst: 100000` explicitly in `testAPI`'s `Conf` literal.
- **Files modified:** `server/test.go`
- **Verification:** `go test ./server -count=1` and `go test ./server -race -count=1` both green with zero failures, run multiple times.
- **Committed in:** `f1ea608` (Task 1 commit)

---

**2. [Rule 3 - Blocking] `IsGlobalUnicast` named in a comment tripped this plan's own verify gate**
- **Found during:** Task 2, running the plan's own `<verify>` block
- **Issue:** `server/deck_providers.go`'s doc comment explaining why the standard library's global-unicast predicate is not the gate spelled out its Go identifier (`IsGlobalUnicast`) — but this task's own verify block greps the file for the *absence* of that exact string, the same protection the plan's action text applies to `ProxyFromEnvironment`. The comment failed its own gate for the one reason that is not a real defect.
- **Fix:** Rephrased the comment to describe the predicate without naming its Go identifier, matching the pattern the plan's action text already prescribes for the proxy resolver.
- **Files modified:** `server/deck_providers.go`
- **Verification:** `! grep -q 'IsGlobalUnicast' server/deck_providers.go` now exits 0.
- **Committed in:** `75f7ea6` (Task 2 commit)

---

**3. [Rule 1 - Bug] The plan's own verify gate for `-run 'TestRateLimit'` matched zero tests**
- **Found during:** the plan-level `<verification>` pass after all three tasks
- **Issue:** `server/ratelimit_test.go`'s test functions were named `TestAllowRequest_*`/`TestClientKeyFor_*`, which never matched the `-run 'TestRateLimit'` pattern this plan's own Task 1 `<verify>` block (and the plan-level `<verification>` section) runs. `go test -run` with zero matches still exits 0 ("no tests to run"), so the gate passed vacuously without exercising any of this task's actual coverage.
- **Fix:** Renamed all five test functions in `server/ratelimit_test.go` to a `TestRateLimit_` prefix.
- **Files modified:** `server/ratelimit_test.go`
- **Verification:** `go test ./server -run 'TestRateLimit' -race -v` now runs and passes all five.
- **Committed in:** `4de341e` (separate fixup commit)

---

**Total deviations:** 3 auto-fixed (2 blocking, 1 bug). No scope creep — all three are corrections required for this plan's own stated verification to mean what it claims.

## Issues Encountered

None beyond the three auto-fixed deviations above. One design tension (documented as key-decisions, not deviations, since no plan text was violated): several `<action>`-prose test scenarios (a literal "declares 10 bytes, sends 2 MiB" response; a two-stage DNS-rebinding stub) turned out to be either structurally unreproducible against Go's standard `net/http` client, or strictly weaker than a simpler equivalent 01-RESEARCH.md itself sanctions. Both were resolved in favor of the reproducible, research-sanctioned equivalent, documented inline in the test file and in this SUMMARY's key-decisions.

## User Setup Required

None for this plan's own code. Two new environment variables are deployment-affecting configuration, both defaulting to a fully safe state:
- `DECK_PROVIDER_ENABLED` (default `false`) — the provider kill switch. Must be set to `true`, alongside a non-empty `DECK_PROVIDER_ALLOWED_HOSTS`, before any deck-URL fetch path can ever be exercised.
- `DECK_PROVIDER_ALLOWED_HOSTS` (default empty, comma-separated) — the exact-match host allowlist. An empty value keeps `providerEnabled()` false regardless of the flag.
- `DECK_IMPORT_RATE_PER_MINUTE` (default `30`) / `DECK_IMPORT_RATE_BURST` (default `10`) — the shared rate budget for `previewDeck` and `trackProductEvent`. No action required to preserve current behavior; both already apply via `envconfig` defaults in `main.go`'s production path.

## Next Phase Readiness

- REQ-ACT-001 is now fully satisfied across plans 01-01 (product_events/vocabulary), 01-03 (frontend session-ID service), and this plan (rate-limit instrumentation) — marked complete in `REQUIREMENTS.md`.
- REQ-ACT-003 remains **not** marked complete: this plan delivers its NFR/security clauses (HTTPS-only, DNS/IP validation, redirect revalidation, timeouts, cap, kill switch) but not the one-day feasibility comparison, the decision record, or the adapter interface itself, all of which `01-07-PLAN.md` (frontmatter: `requirements: [REQ-ACT-003]`) still claims. Marking it complete now would misrepresent the requirement's actual traceability state.
- Plan 01-07's D-14 checkpoint has everything it needs in either branch: `newSafeProviderClient`/`fetchDeckProviderURL` for a real adapter, and `providerEnabled()`/`DECK_PROVIDER_ENABLED` already shipping default-off for a documented no-go. `previewDeckURL`'s `deckProviderFetch` seam is exactly where 01-07's adapter should be wired in, rather than re-deriving resolution logic.
- No blockers. `go build ./...`, `go vet ./server ./pkg/telemetry ./pkg/ratelimit`, `go test ./pkg/... -race`, `go test ./server -count=1`, `go test ./server -race -count=1`, and the plan's full combined `-run 'TestRateLimit|TestSafeControl|TestSafeClient|TestProvider_' -race` gate are all green. `go list -deps ./pkg/ratelimit` names no `server/` package.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-05*

## Self-Check: PASSED

All 14 files listed under `key-files` (created/modified) exist on disk, and all four
commit hashes (`f1ea608`, `75f7ea6`, `c925d38`, `4de341e`) are present in
`git log --oneline --all`.

---
status: resolved
phase: 01-measured-deck-import-foundation
source: [01-VERIFICATION.md, 01-01-SUMMARY.md, 01-02-SUMMARY.md, 01-03-SUMMARY.md, 01-04-SUMMARY.md, 01-05-SUMMARY.md, 01-06-SUMMARY.md, 01-07-SUMMARY.md]
started: 2026-08-05T22:33:38Z
updated: 2026-08-08T07:50:00Z
---

## Current Test

[testing complete]

## Tests

### 1. REQ-ACT-003 traceability sign-off
expected: Product-owner agreement that "decision record + fail-closed scaffold, no working Moxfield import" is an acceptable Phase-1 exit for REQ-ACT-003, consistent with ROADMAP criterion 5's either/or wording.
result: issue
reported: "Why can't we do Moxfield? Do Archidekt instead then if we can't do Moxfield."
severity: major
source: human_judgment
origin: 01-VERIFICATION.md human_verification[0]
note: Not a defect report — a reversal of the D-14 provider decision. The user declined to sign off on shipping REQ-ACT-003 as Pending and redirected the first provider from Moxfield to Archidekt, which the 01-07 feasibility spike originally recommended and for which a real observed response contract already exists in-repo.
resolution: The redirect was actioned, not merely accepted. `result` stays `issue` because the sign-off this test asked for genuinely never happened — the proposition was withdrawn rather than agreed. What closed the loop is gap G-01-1 below, resolved 2026-08-08 by plans 01-08/01-09/01-10; REQ-ACT-003 now ships as Complete on a working Archidekt import rather than as Pending on a fail-closed Moxfield scaffold.

### 2. No SQL composed by string formatting; vet and card/import/migration tests green (01-04 D8)
expected: No SQL string in server/deck_import.go or server/cards.go is composed with string formatting, and `go vet ./server` plus `go test ./server -run 'TestCards|TestDeckImport_|TestMigrations_' -race` pass in one invocation without a snapshot reimport.
result: pass
source: human_judgment
origin: 01-04-SUMMARY.md coverage D8 (fail-safed — verification[].kind used the non-enum value "automated")
note: Orchestrator re-verified all three refs directly — Sprintf-composed SQL grep returns zero matches, go vet is clean, and all 14 named tests pass (54.7s).

### 3. Preview card count equals created library length from one parse (01-05 D7)
expected: The preview's card count equals the created library's length for the same parsed value, and the library is built from a parsed deck value rather than re-parsing raw text under different rules.
result: pass
source: human_judgment
origin: 01-05-SUMMARY.md coverage D7 (fail-safed — verification[1].kind used the non-enum value "automated")
note: Orchestrator re-verified directly — TestDeckImport_SingleParse passes, server/games.go contains no encoding/csv and consumes deckimport.

### 4. End-to-end tracer: a pasted '1 Sol Ring' through previewDeck returns a correct DeckPreview
expected: End-to-end tracer: a pasted '1 Sol Ring' through previewDeck returns a correct DeckPreview, writes one deck_import_succeeded product_events row, and moves the vedh_deck_import_total counter by exactly one
result: pass
source: automated
coverage_id: D1
plan: 01-01

### 5. The PostgreSQL 14 COALESCE dedup index actually deduplicates (both-NULL keys, concurrent i
expected: The PostgreSQL 14 COALESCE dedup index actually deduplicates (both-NULL keys, concurrent in-flight collision, client/server replay shapes), a repeatable-event control is never deduplicated, occurred_at ties order deterministically and collapse correctly in a min()-per-session funnel query, the allowlist refuses five case types with an exact bounded-counter delta and writes no row, and the product_events migration pair applies/reverses/re-applies cleanly in scratch databases for both migration directories
result: pass
source: automated
coverage_id: D2
plan: 01-01

### 6. The 15-event vocabulary is closed and matches D-20 literally, no allowlisted metadata key 
expected: The 15-event vocabulary is closed and matches D-20 literally, no allowlisted metadata key names a forbidden concept, exactly the six PRD-Server events are marked Authoritative, every ValidateEvent rejection case returns its own bounded Rejection value, every vedh_-prefixed Prometheus label is in AllowedLabelNames, all four criterion-4 families exist with a counter and histogram, and source/role/reason label values are bounded to their declared enums
result: pass
source: automated
coverage_id: D3
plan: 01-01

### 7. All six required decklist syntaxes (1, 1x, 1X, 1,, 1, , quoted-comma, unquoted-comma) pars
expected: All six required decklist syntaxes (1, 1x, 1X, 1,, 1, , quoted-comma, unquoted-comma) parse to identical (quantity, name) pairs, and the comma-truncation regression (1,Atraxa, Praetors' Voice) yields the full name instead of truncating to \"Atraxa\"
result: pass
source: automated
coverage_id: D1
plan: 01-02

### 8. A double slash inside a line is a double-faced card-name separator preserved in the name (
expected: A double slash inside a line is a double-faced card-name separator preserved in the name (verified against the repo's own test/decklists/jarad.csv fixtures); a double slash at the start of a line is a comment producing no entry
result: pass
source: automated
coverage_id: D2
plan: 01-02

### 9. D-07 printing metadata (set code, collector number, category via bracket/backtick/hash-tag
expected: D-07 printing metadata (set code, collector number, category via bracket/backtick/hash-tag) is extracted into its own fields with the remaining name clean, and set-code letter case is preserved exactly as written
result: pass
source: automated
coverage_id: D3
plan: 01-02

### 10. The 'no nonblank row disappears' accounting invariant (Entries+Warnings+BlockingErrors+Dro
expected: The 'no nonblank row disappears' accounting invariant (Entries+Warnings+BlockingErrors+DroppedSectionRows == NonBlankLines) is asserted mechanically in code (AssertAccounting), not only in a test, and is proven over the required-syntax corpus and both real test/decklists/*.csv fixtures
result: pass
source: automated
coverage_id: D4
plan: 01-02

### 11. Parse is pure (two calls on the same text are deeply equal) and holds no mutable package-l
expected: Parse is pure (two calls on the same text are deeply equal) and holds no mutable package-level state (safe under -race from concurrent goroutines); MaxDecklistBytes/MaxDecklistLines are each enforced exactly at their boundary with no panic
result: pass
source: automated
coverage_id: D5
plan: 01-02

### 12. A Sideboard/Maybeboard header drops its rows with one summary warning stating the count; a
expected: A Sideboard/Maybeboard header drops its rows with one summary warning stating the count; a Commander header preselects candidates via SectionCommander; a card name beginning with a header word (\"1 Commander's Sphere\") stays an entry, never a header
result: pass
source: automated
coverage_id: D6
plan: 01-02

### 13. Paste-shape source detection (moxfield/archidekt/plain_text/unknown) never influences what
expected: Paste-shape source detection (moxfield/archidekt/plain_text/unknown) never influences what a line parses to: forcing every SourceType via ParseWithSource across all three golden fixtures produces deeply equal ParsedDecks apart from the Source field
result: pass
source: automated
coverage_id: D7
plan: 01-02

### 14. Three whole-file golden fixtures (Moxfield-, Archidekt-, plain-text-shaped) round-trip aga
expected: Three whole-file golden fixtures (Moxfield-, Archidekt-, plain-text-shaped) round-trip against their checked-in expected ParsedDeck JSON, each additionally passing the accounting invariant; the -update flag refuses to run when CI=true
result: pass
source: automated
coverage_id: D8
plan: 01-02

### 15. getSessionID() persists one random identifier under edhgo/session-id, is stable across rep
expected: getSessionID() persists one random identifier under edhgo/session-id, is stable across repeated calls and a simulated module reload against the same storage, and degrades to a fresh unpersisted identifier (never throwing) when storage is absent, throws, or holds an empty string
result: pass
source: automated
coverage_id: D1
plan: 01-03

### 16. newSessionID() sources from crypto.randomUUID() first, falls through to crypto.getRandomVa
expected: newSessionID() sources from crypto.randomUUID() first, falls through to crypto.getRandomValues() hex-encoded to 32 characters when randomUUID is absent, and only then to a non-cryptographic composition
result: pass
source: automated
coverage_id: D2
plan: 01-03

### 17. track() sends exactly one apolloClient.mutate() per call, returns undefined synchronously 
expected: track() sends exactly one apolloClient.mutate() per call, returns undefined synchronously (never a thenable), is never batched, and swallows a rejected mutation at console.debug level without throwing or reaching console.error; it never reads the edhgo/auth storage key
result: pass
source: automated
coverage_id: D3
plan: 01-03

### 18. captureAttribution() reads only the allowlisted utm_source/utm_medium/utm_campaign query k
expected: captureAttribution() reads only the allowlisted utm_source/utm_medium/utm_campaign query keys plus a referrer reduced to its lower-cased host, trims and byte-truncates every value to at most 128 bytes on a codepoint boundary, drops empty-after-trim keys, memoises the result once per page view, and is wired into track() only for the two events (landing_primary_cta, invite_viewed) the vocabulary document marks as accepting it
result: pass
source: automated
coverage_id: D4
plan: 01-03

### 19. The cards table carries indexes on lower-cased name, lower-cased face name, and the compos
expected: The cards table carries indexes on lower-cased name, lower-cased face name, and the composite lower-cased name/set-code/collector-number key; a card_names projection holds one row per distinct searchable name with a trigram GiST index; the migration pair applies/reverses/re-applies cleanly in scratch databases for both migration directories without ever touching the shared TestMain database
result: pass
source: automated
coverage_id: D1
plan: 01-04

### 20. Importing the MTGJSON snapshot repopulates card_names without a migration, tolerating the 
expected: Importing the MTGJSON snapshot repopulates card_names without a migration, tolerating the projection table being absent by logging and continuing
result: pass
source: automated
coverage_id: D2
plan: 01-04

### 21. A batch card lookup with a lower-cased needle matches a stored name of any letter case, an
expected: A batch card lookup with a lower-cased needle matches a stored name of any letter case, and the lookup query lower-cases both the column and the parameter so the expression indexes are usable
result: pass
source: automated
coverage_id: D3
plan: 01-04

### 22. An entry naming a set code and collector number resolves to that exact printing when it ex
expected: An entry naming a set code and collector number resolves to that exact printing when it exists, and set-code matching is case-insensitive so both spellings of the same set select the same row
result: pass
source: automated
coverage_id: D4
plan: 01-04

### 23. An entry naming a set code absent from the snapshot resolves to some available printing of
expected: An entry naming a set code absent from the snapshot resolves to some available printing of that name, carries a warning naming the requested printing, keeps the parsed set code on the entry, and is not reported as unresolved
result: pass
source: automated
coverage_id: D5
plan: 01-04

### 24. An entry naming no set code or collector number resolves by name alone, and which printing
expected: An entry naming no set code or collector number resolves by name alone, and which printing is returned is decided by a documented deterministic ordering rather than by row arrival order
result: pass
source: automated
coverage_id: D6
plan: 01-04

### 25. A trigram score exactly equal to the low-confidence cutoff is treated as confident, and a 
expected: A trigram score exactly equal to the low-confidence cutoff is treated as confident, and a score one step below it is marked low-confidence; the cutoff is applied in Go over the returned score and appears nowhere in the SQL
result: pass
source: automated
coverage_id: D7
plan: 01-04

### 26. Every unresolved card carries up to three ranked candidates in the same previewDeck respon
expected: Every unresolved card carries up to three ranked candidates in the same previewDeck response, the below-cutoff case still returns the nearest matches marked low-confidence rather than an empty list, an equal-score tie between two candidates breaks on the displayed name ascending, and a repeated misspelling collapses to one needle before any query runs
result: pass
source: automated
coverage_id: D1
plan: 01-05

### 27. Exactly 25 distinct unresolved needles produces no truncation warning and no counter movem
expected: Exactly 25 distinct unresolved needles produces no truncation warning and no counter movement; 26 produces exactly one warning naming both counts and exactly one increment of vedh_deck_suggestion_truncated_total
result: pass
source: automated
coverage_id: D2
plan: 01-05

### 28. A suggestion query whose sub-budget has already expired returns zero suggestions plus one 
expected: A suggestion query whose sub-budget has already expired returns zero suggestions plus one warning, and the surrounding preview still succeeds with the same CardCount and CanContinue as the same parsed input run with the real budget
result: pass
source: automated
coverage_id: D3
plan: 01-05

### 29. An unresolved entry counts toward neither the deck-size check nor the created library (a 1
expected: An unresolved entry counts toward neither the deck-size check nor the created library (a 100-card paste with 3 unmatched names yields a 97-card library and a 97 preview count), and a 101-card paste with 2 unmatched names is accepted because resolution now precedes the size check
result: pass
source: automated
coverage_id: D4
plan: 01-05

### 30. Exactly 100-minus-commanders resolved cards is accepted; one more is a blocking error nami
expected: Exactly 100-minus-commanders resolved cards is accepted; one more is a blocking error naming both the actual and maximum count
result: pass
source: automated
coverage_id: D5
plan: 01-05

### 31. Two rows naming the same card contribute their summed quantity to the library while remain
expected: Two rows naming the same card contribute their summed quantity to the library while remaining two separately-parsed entries, and a card named both as a selected commander and in the main deck has exactly the commander count removed from its quantity with the remainder kept
result: pass
source: automated
coverage_id: D6
plan: 01-05

### 32. Assigned test fix: go test ./server -count=1 and go test ./server -race -count=1 both pass
expected: Assigned test fix: go test ./server -count=1 and go test ./server -race -count=1 both pass with zero failures — the three pre-existing format-registry failures are resolved (one stale test expectation, two real product regressions), and a full-suite connection-exhaustion failure this plan's own added tests exposed is fixed
result: pass
source: automated
coverage_id: D8
plan: 01-05

### 33. previewDeck and trackProductEvent are rate limited per client key and per surface at 30/mi
expected: previewDeck and trackProductEvent are rate limited per client key and per surface at 30/minute with a burst of 10 (Conf-configurable), and vedh_rate_limit_total increments exactly once per decision on both the allowed and the limited path so the ratio is readable before a too-tight limit becomes a funnel dip
result: pass
source: automated
coverage_id: D1
plan: 01-06

### 34. Two distinct client keys and two distinct surfaces never share a token bucket, and an entr
expected: Two distinct client keys and two distinct surfaces never share a token bucket, and an entry idle past the eviction window is evicted and its slot returns to zero on a full sweep -- all proven by pkg/ratelimit tests that run in CI with no database and no MTGJSON snapshot
result: pass
source: automated
coverage_id: D2
plan: 01-06

### 35. A rate-limited previewDeck call returns a product-language blocking error naming no limit 
expected: A rate-limited previewDeck call returns a product-language blocking error naming no limit value, window, or remaining count, with CanContinue false; a rate-limited trackProductEvent still returns true with no error
result: pass
source: automated
coverage_id: D3
plan: 01-06

### 36. The secure outbound fetch client refuses the metadata address, loopback, shared-address-sp
expected: The secure outbound fetch client refuses the metadata address, loopback, shared-address-space, an IPv4-mapped-IPv6 form of the metadata address, protocol-assignment and benchmarking ranges, and the wrong port or a non-TCP network, while allowing a public address on the correct port -- opening no socket -- and the same first-matching-prefix is named deterministically for an address inside two declared prefixes
result: pass
source: automated
coverage_id: D4
plan: 01-06

### 37. Two redirect hops are followed and the third is refused; a redirect to a non-secure scheme
expected: Two redirect hops are followed and the third is refused; a redirect to a non-secure scheme is refused; a host differing only in letter case is accepted while a host merely containing an allowlisted host as a substring is refused; a redirect to an unallowlisted host is refused and that host's handler is never invoked
result: pass
source: automated
coverage_id: D5
plan: 01-06

### 38. An empty host and a non-secure scheme are refused before any name resolution is attempted;
expected: An empty host and a non-secure scheme are refused before any name resolution is attempted; ControlContext is genuinely wired into the real client (a hostname resolving to loopback is refused, not merely correct in isolation); a response of exactly the 1 MiB cap is accepted, one byte more is refused, and a chunked response with no declared length carrying far more than the cap is also refused; both the connect and the total budget fail closed within a bounded, parameterised-down elapsed time; a zero-byte body yields a normalized provider error
result: pass
source: automated
coverage_id: D6
plan: 01-06

### 39. With the provider flag unset, or set with an empty host allowlist, a deck URL yields exact
expected: With the provider flag unset, or set with an empty host allowlist, a deck URL yields exactly one blocking error mentioning pasting text with CanContinue false and zero dial attempts (proven via a spy, not merely an error return); a text-only preview is byte-identical across both flag states
result: pass
source: automated
coverage_id: D7
plan: 01-06

### 40. A written, neutral decision record evaluates both Archidekt and Moxfield against the same 
expected: A written, neutral decision record evaluates both Archidekt and Moxfield against the same six criteria, states a recommendation tied to the hard gate condition, and records the user's actual checkpoint selection (Moxfield) including its divergence from that recommendation and the two open blockers gating enablement
result: pass
source: automated
coverage_id: D1
plan: 01-07

### 41. Hostname-keyed adapter routing (deckProviderAdapters/deckProviderAdapterFor) reaches the p
expected: Hostname-keyed adapter routing (deckProviderAdapters/deckProviderAdapterFor) reaches the plan 01-06 secure fetch client for a registered, allowlisted host, and refuses an allowlisted-but-unregistered host before any fetch is attempted — proving the routing decision is real, not a relabeled stub
result: pass
source: automated
coverage_id: D2
plan: 01-07

### 42. moxfieldAdapter's normalizer deliberately refuses to guess a field mapping for any body it
expected: moxfieldAdapter's normalizer deliberately refuses to guess a field mapping for any body it is handed, returning a static sentinel error that carries no provider response content, mapped onto the ordinary paste-fallback message before ever reaching a client
result: pass
source: automated
coverage_id: D3
plan: 01-07

### 43. The provider scaffold does not regress plan 01-06's stability guarantees: with the kill sw
expected: The provider scaffold does not regress plan 01-06's stability guarantees: with the kill switch off, a moxfield.com deck URL still fails closed with zero fetch attempts, and a pasted decklist is entirely unaffected regardless of the scaffold's presence
result: pass
source: automated
coverage_id: D4
plan: 01-07

### 44. Every locked network control (deny-prefix count, 3s/8s timeout budgets, exact-equality hos
expected: Every locked network control (deny-prefix count, 3s/8s timeout budgets, exact-equality host match, redirect bound) is unchanged from plan 01-06, and the full secure-client and pkg/ratelimit test suites pass unchanged
result: pass
source: automated
coverage_id: D5
plan: 01-07

### 45. COVERAGE.md records the external-API coverage decision honestly: every Moxfield capability
expected: COVERAGE.md records the external-API coverage decision honestly: every Moxfield capability is OPT-OUT with a stated reason, reflecting that this ships as off-by-default scaffolding rather than a working import
result: pass
source: automated
coverage_id: D6
plan: 01-07

## Summary

total: 45
passed: 44
issues: 1
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-01-1
  truth: "REQ-ACT-003 delivers a first public deck provider adapter, keyed by an allowlisted hostname, normalizing through ACT-002 behind a server-side kill switch"
  status: resolved
  resolved: 2026-08-08
  resolved_by: [01-08-PLAN.md, 01-09-PLAN.md, 01-10-PLAN.md]
  resolution: >-
    All six `missing:` items below are delivered and independently verified (01-VERIFICATION.md,
    status passed, 5/5). archidektAdapter normalizes the committed 568 KB live capture through
    pkg/deckimport (TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste proves
    the single-canonical-parse contract, CardCount 100 from a 113-row response); fixture-contract
    tests pin the field mapping; the Moxfield scaffold and its WINDOWS.md stub window are gone;
    the suite makes zero live dials. The human precondition (read https://archidekt.com/terms) is
    NOT satisfied and remains an operator gate on *enabling* the provider — recorded in
    docs/research/deck-provider-feasibility.md section 6, COVERAGE.md "Operator Setup", and
    01-10-SUMMARY.md. DECK_PROVIDER_ENABLED still defaults false.
    Two silent-corruption defects introduced by this gap-closure work (embedded newline in a card
    name hijacking section assignment; unvalidated setCode corrupting a card name) were caught by
    code review, reproduced by the verifier, fixed in d2d8df6/53d9f60/3e5786e/e628f39, and
    re-verified — see 01-REVIEW.md iteration 2 and 01-REVIEW-FIX.md.
  reason: "User reported: Why can't we do Moxfield? Do Archidekt instead then if we can't do Moxfield."
  severity: major
  test: 1
  root_cause: "Already established by plan 01-07's feasibility spike — no diagnosis needed. Moxfield's response contract is unobtainable without authorized API access: api.moxfield.com/robots.txt is a blanket `Disallow: /` (orchestrator re-verified 2026-08-05) and Cloudflare bot-management 403s every unauthenticated probe, so the shape was never observed. The D-14 decision selected Moxfield anyway, which forced plan 01-07 Task 3 to ship an off-by-default scaffold with `moxfieldAdapter.normalizeToDeckText` returning `errMoxfieldContractUnverified` rather than a fabricated field mapping. The user has now reversed that decision to Archidekt, whose contract IS observed."
  artifacts:
    - path: "docs/research/deck-provider-feasibility.md"
      issue: "Section 5 records Moxfield as the selected provider; must be superseded with the revised Archidekt decision, dated, preserving the original for audit"
    - path: "server/deck_providers.go"
      issue: "deckProviderAdapters registers only moxfieldAdapter, whose normalizeToDeckText is a deliberate errMoxfieldContractUnverified stub"
    - path: ".planning/phases/01-measured-deck-import-foundation/COVERAGE.md"
      issue: "All 6 provider capabilities recorded OPT-OUT on the basis of the unverified Moxfield contract"
    - path: ".planning/REQUIREMENTS.md"
      issue: "REQ-ACT-003 left Pending because no working import exists; becomes completable once the Archidekt adapter lands"
    - path: ".planning/WINDOWS.md"
      issue: "Open `stub` window logged for moxfieldAdapter.normalizeToDeckText; must be resolved or re-scoped"
  missing:
    - "archidektAdapter implementing normalizeToDeckText against the real observed contract in server/testdata/deck_providers/archidekt_deck_2026-08-05.json (568 KB, MTGJSON-era live capture), registered in deckProviderAdapters under the archidekt.com hostname"
    - "Fixture-contract tests pinning the field mapping to that captured response, so an upstream shape change fails loudly rather than silently mis-parsing (the API is the internal SPA endpoint — undocumented and unversioned)"
    - "Normalization through pkg/deckimport so the Archidekt path produces the same ParsedDeck as a paste, satisfying ACT-002's single-canonical-parse contract"
    - "DECIDED by user 2026-08-07 — REMOVE the Moxfield scaffold: delete moxfieldAdapter and its errMoxfieldContractUnverified stub so deckProviderAdapters registers only providers with an observed contract, and close the corresponding open `stub` window in .planning/WINDOWS.md. Do not leave it registered-but-failing. Re-addable later if authorized Moxfield API access is ever obtained."
    - "Hermetic tests only — reuse the existing deckProviderFetch spy pattern that asserts zero dial attempts; no live Archidekt calls in the suite"
    - "Human precondition before the kill switch may be enabled: read https://archidekt.com/terms (JS-rendered, never read by any agent). robots.txt permits the /api/decks/ path but is not a licence."

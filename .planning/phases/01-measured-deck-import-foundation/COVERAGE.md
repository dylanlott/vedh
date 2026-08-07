# Phase 1 External-API Coverage: Deck Provider (Archidekt)

**Provider selected:** Archidekt — reversed from the original Moxfield selection at UAT
gap `G-01-1` (2026-08-07), per `docs/research/deck-provider-feasibility.md` section 6.
Section 5 of that document (the original Moxfield decision and its evidence) is left
intact as the historical record; it is superseded, not deleted.

**Status of this integration: a working, contract-tested import.** Plan 01-08 implemented
`archidektAdapter.normalizeToDeckText` against a real, committed 568KB response capture
(`server/testdata/deck_providers/archidekt_deck_2026-08-05.json`), and plan 01-09 pinned
that field mapping with fixture-contract tests that load the same capture from disk and run
it through the real adapter and the real canonical parser — never a hand-built shortcut.
The rows below reflect what those tests actually prove, not what the code merely attempts.

| capability | decision | reason |
|---|---|---|
| read_public_deck | INTEGRATE | |
| read_private_deck | OPT-OUT | Project policy: never request, store, or transmit a user's third-party credentials, independent of contract status (`.planning/PROJECT.md` out-of-scope list). |
| list_user_decks | OPT-OUT | Not attempted: `archidektAdapter` fetches exactly one deck by URL (`GET https://archidekt.com/api/decks/<id>/`); no endpoint for listing a user's decks was ever observed or implemented. |
| read_deck_metadata (name, format) | OPT-OUT | `archidektDeckResponse` (`server/deck_providers.go`) reads only `categories[]` and `cards[]`; no deck-level name or format field is read, mapped, or surfaced anywhere in the adapter or its tests. |
| read_commander_designation | INTEGRATE | |
| search_decks | OPT-OUT | Not attempted by this phase; out of scope for the activation-path MVP regardless of contract status. |

## What is actually shipped, with the test that proves it

- **read_public_deck (INTEGRATE).** `archidektAdapter.normalizeToDeckText`
  (`server/deck_providers.go`) buckets every `cards[]` row into main, commander, or
  maybeboard text and hands it to `pkg/deckimport.Parse` — the same canonical parser paste
  input uses (ACT-002). Hostname-keyed routing (`deckProviderAdapters`,
  `deckProviderAdapterFor`) is reachable through `previewDeckURL`
  (`server/deck_import.go`) once `DECK_PROVIDER_ENABLED=true` and `archidekt.com` is in
  `DECK_PROVIDER_ALLOWED_HOSTS`.
  - `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` —
    end-to-end proof the provider path reuses the single canonical parse: the real fixture
    resolves to exactly `CardCount == 100` (126 total quantity minus the 26-quantity
    excluded Maybeboard).
  - `TestArchidekt_QuantityAccounting` and `TestArchidekt_AssertAccountingHolds` —
    the 126-vs-100/26 contract-drift canary, asserted arithmetically against the real
    capture.
  - `TestArchidekt_MalformedInputFailsClosed` (4 subtests) — empty body, non-JSON bytes,
    no `cards[]`, and a missing card name all fail closed, never a partial deck.
  - `TestArchidekt_PrintingMetadata` / `TestArchidekt_Source` — D-07 set-code/collector
    metadata and `deckimport.SourceArchidekt` attribution both carry through.
  - `TestDeckImport_ArchidektURLPath/DisabledNeverDials` and
    `/NonAllowlistedHostNeverDialsEvenWithFlagOn` — the kill switch and the allowlist
    remain independent gates on the real path.
  - SSRF controls (HTTPS-only, DNS/IP validation, redirect revalidation, 3s connect / 8s
    total timeout, 1 MiB cap) are provider-agnostic and unchanged since plan 01-06:
    `TestSafeControl_DeniedAddresses`, `TestSafeClient_HostAndRedirect`,
    `TestSafeClient_Timeouts`, `TestSafeClient_BodyCap`.
- **read_commander_designation (INTEGRATE).** `TestArchidekt_CommanderPreselection`
  proves comma-intact commander preselection ("Cid, Timeless Artificer") survives the
  provider path, against a fixture where the same name also appears as an unrelated
  main-deck row — so the assertion genuinely filters on `Section`, not name occurrence.

## Operator setup — required before enabling for real Archidekt traffic

1. Set `DECK_PROVIDER_ALLOWED_HOSTS` to include `archidekt.com` and set
   `DECK_PROVIDER_ENABLED=true`. Both are required — `providerEnabled()`
   (`server/deck_providers.go`) ANDs the flag against a non-empty allowlist, so a
   half-configured deployment never dials.
2. **Before flipping `DECK_PROVIDER_ENABLED=true` in any real deployment, a human must
   read `https://archidekt.com/terms` in a real browser first.** The page is
   client-rendered; no agent working on this project has JavaScript execution capability,
   so no agent has ever read what Archidekt's terms actually say about automated or
   programmatic reads. `robots.txt` permitting the `/api/decks/` path is crawler
   etiquette, not a license. See
   `docs/research/deck-provider-feasibility.md` section 6, "Human precondition for
   enabling — not for building," for the full record. This precondition is recorded
   there and here; it is not resolved by either document.
3. Re-verify the 6-row matrix above against a real, current Archidekt response before
   relying on it long-term — `docs/research/deck-provider-feasibility.md` section 6
   already discloses that the response shape is an unversioned internal API, observed
   consistent across six real decks on 2026-08-05, not a published or guaranteed
   contract.

Building and testing `archidektAdapter` requires none of the above: it is exercised
entirely by hermetic tests against the committed fixture, and the kill switch defaults
off regardless of which adapter is registered.

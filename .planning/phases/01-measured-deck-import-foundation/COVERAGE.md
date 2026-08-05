# Phase 1 External-API Coverage: Deck Provider (Moxfield)

**Provider selected:** Moxfield, at the Task 2 `checkpoint:decision` in `01-07-PLAN.md`
(2026-08-05), per `docs/research/deck-provider-feasibility.md` section 5.

**Status of this integration: scaffolding, not a working import.** The response contract
was never observed — `api.moxfield.com/robots.txt` is a blanket `Disallow: /`, and no
authorized sample response has ever been captured (see the feasibility record, sections 2
and 5). Every capability below therefore starts at INTEGRATE per this document's own rule
and is deliberately subtracted to OPT-OUT, because nothing is actually integrated today:
hostname-keyed routing and the kill switch are real and hermetically tested, but the
normalizer that would turn a Moxfield response into a parsed deck
(`moxfieldAdapter.normalizeToDeckText`) always fails closed by design. This table will be
revised, capability by capability, once an authorized sample response exists and each row
can honestly move to INTEGRATE.

| capability | decision | reason |
|---|---|---|
| read_public_deck | OPT-OUT | Response contract unverified (open blocker (b)); `api.moxfield.com` access requires authorization not yet obtained (open blocker (a)). See decision record section 5. |
| read_private_deck | OPT-OUT | Project policy: never request, store, or transmit a user's third-party credentials, independent of contract status. |
| list_user_decks | OPT-OUT | No endpoint shape was ever observed for this capability; blocked by the same two open blockers as read_public_deck. |
| read_deck_metadata (name, format) | OPT-OUT | Response contract unverified; no field mapping exists to read metadata from. |
| read_commander_designation | OPT-OUT | Response contract unverified; no field mapping exists to read this from. |
| search_decks | OPT-OUT | Not attempted by this spike; blocked by the same authorization and contract blockers, and out of this phase's scope regardless. |

## What is actually shipped, for clarity

- Hostname-keyed adapter routing (`deckProviderAdapters`, `deckProviderAdapterFor` in
  `server/deck_providers.go`) — real code, hermetically tested, reachable through
  `previewDeckURL` (`server/deck_import.go`) when `DECK_PROVIDER_ENABLED=true` and
  `moxfield.com` is in `DECK_PROVIDER_ALLOWED_HOSTS`.
- The default-off kill switch from plan 01-06 (`DECK_PROVIDER_ENABLED`,
  `providerEnabled()`) — unchanged, still defaults to disabled.
- An explicit, unimplemented normalizer seam (`moxfieldAdapter.normalizeToDeckText`,
  `errMoxfieldContractUnverified`) that fails closed with a plain error rather than a
  fabricated field mapping.
- No code path in this repository ever makes a live request to any Moxfield host. Every
  test covering this scaffold uses a spied `deckProviderFetch` or a direct, in-memory call
  to the adapter's normalizer.

Re-verify before flipping `DECK_PROVIDER_ENABLED=true` for Moxfield in any real deployment:
both open blockers in `docs/research/deck-provider-feasibility.md` section 5 must be
resolved first, and this table's OPT-OUT rows must be revisited against a real, authorized
sample response before any of them can honestly move to INTEGRATE.

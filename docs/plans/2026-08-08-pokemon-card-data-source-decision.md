# Pokémon card data source decision

## Why this exists

vEDH's shared card handlers currently assume MTG-shaped data: local DB lookups, MTG-style name search, and Scryfall-backed images. Before we add Pokémon ingestion, schema fields, or UI affordances, we need a provider seam and a source decision that can tolerate different search semantics and different image-hosting rules.

## MVP normalized fields vEDH needs

For Pokémon support, the provider contract should normalize these fields regardless of upstream source shape:

- canonical id
- name
- supertype / category
- subtype / species
- hp
- stage
- types
- set
- collector number
- image url

Notes:
- `supertype / category` covers values like Pokémon / Trainer / Energy.
- `subtype / species` is intentionally loose because some sources model gameplay subtype and Pokémon species separately.
- `image url` is a provider concern, not a shared MTG assumption.

## Candidate sources

### 1) Pokémon TCG API (`pokemontcg.io`) + `PokemonTCG/pokemon-tcg-data`

**What it gives us**
- Actively used public API with predictable JSON resources.
- Optional API key, but usable without one for lower-rate access.
- Community-maintained backing dataset available as raw JSON in GitHub releases/repo.
- Hosted card image URLs already included in the API payload.
- Set metadata, collector numbers, HP, stage, subtypes, supertypes, and card types are already modeled close to what vEDH needs.

**Pros**
- Lowest integration cost for MVP.
- Search and card detail payloads are already normalized enough for app use.
- Hosted images mean we do not need to build an image mirror on day one.
- GitHub dataset gives us a bulk-import escape hatch if we later want local ingestion parity with MTGJSON.

**Cons / cautions**
- Community-run service, not an official Pokémon Company API.
- Rate limits are better with an API key, so uncached live lookups are not ideal for heavy use.
- Image hosting is convenient, but still depends on a third-party service we do not control.
- Licensing is comfortable for internal/reference MVP use, but not as clean as purely self-hosted metadata with first-party asset rights.

**Licensing / availability / image comfort**
- Availability: strong for MVP; public API and public dataset both exist.
- Licensing comfort: medium; metadata is openly distributed, but card art/text remain trademark/copyright-sensitive.
- Image-hosting comfort: medium-high for MVP because URLs are stable and already packaged with the API, but long-term durability still depends on the service.

### 2) `pkmncards.com`

**What it gives us**
- Excellent search UX and broad historical card coverage.
- Rich card pages that humans use heavily for reference.

**Pros**
- High-quality card discovery and broad catalog.
- Easy manual verification during development.

**Cons / cautions**
- Not designed as our primary machine-ingestion source.
- Scraping would add fragility and maintenance burden.
- Site copy explicitly notes card text/images are copyright Pokémon/Nintendo/Game Freak/Creatures/WotC, which makes reuse comfort lower.

**Licensing / availability / image comfort**
- Availability: fine for manual fallback/reference, weaker for production ingestion.
- Licensing comfort: low-medium for direct integration.
- Image-hosting comfort: low for app dependency; we should not plan to hotlink/scrape this as our primary asset path.

### 3) DIY bulk dataset from assorted scrapers / price sites

**What it gives us**
- Possible full local control if we assembled and cleaned it ourselves.

**Pros**
- Could eventually support a fully local import path similar to MTGJSON.
- Maximum control over schema and storage.

**Cons / cautions**
- Highest maintenance cost.
- Weakest trust and provenance story.
- Highest legal/operational ambiguity around images.
- Delays MVP without unlocking immediate gameplay value.

**Licensing / availability / image comfort**
- Availability: inconsistent.
- Licensing comfort: low.
- Image-hosting comfort: low unless we separately negotiate or self-host from a rights-safe source.

## Decision

**Choose Pokémon TCG API (`pokemontcg.io`) as the primary Pokémon source for MVP.**

Use its API shape as the canonical adapter target, and treat the `PokemonTCG/pokemon-tcg-data` repository as the preferred bulk-data companion when we later add ingestion or caching.

Why this is the best first choice:
- It matches the normalized MVP fields with minimal translation.
- It already includes image URLs, which lets the provider own image resolution cleanly.
- It gives us both an online API path and a future bulk-import path.
- It keeps Task 1 focused on the seam, not on building a custom dataset pipeline too early.

## Fallback policy

1. **Primary:** Pokémon TCG API for lookup/search/image metadata.
2. **Secondary fallback:** local cached records imported from `PokemonTCG/pokemon-tcg-data` once ingestion exists.
3. **Human/reference fallback only:** pkmncards.com for manual verification when a record looks wrong or incomplete.
4. **Failure behavior for MVP:** if the Pokémon provider is not configured or no source is available, return a provider-not-registered / provider-unavailable error rather than silently pretending MTG search semantics apply.

## Adapter implications for Task 1

The server-side contract should keep these concerns separate per game type:
- card lookup
- search semantics
- image/art resolution
- import/source metadata

MTG/EDH remains the first concrete provider, but Pokémon should plug into the same registry by `gameType` instead of inheriting shared MTG assumptions.

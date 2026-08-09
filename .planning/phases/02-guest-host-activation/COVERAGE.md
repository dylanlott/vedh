# API Coverage — Phase 2: Guest Host Activation

No external API integration: this phase builds guest identity on our own PostgreSQL `users` table
and a Vue front end over our own GraphQL server — the only third-party surface in this milestone
(the Archidekt deck-provider adapter) was decided, built, and coverage-recorded in Phase 1, and is
consumed here through the already-shipped `previewDeck` mutation without adding, widening, or
re-deciding a single provider capability.

Detector result recorded at plan time: `{"detected": false, "signals": []}` — run against the
ROADMAP Phase 2 section via `gsd-core/bin/lib/api-coverage.cjs --json`.

Phase 1's provider coverage matrix remains authoritative and unchanged:
`.planning/phases/01-measured-deck-import-foundation/COVERAGE.md`.

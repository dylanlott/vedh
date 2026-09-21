# Deck-to-game quiet beta

## Purpose

Run a measured, reversible Commander activation cohort. Product-funnel truth
comes from PostgreSQL `product_events`; Prometheus/Grafana answers technical
health questions. Do not substitute request metrics for user conversion.

## Before admitting traffic

1. Record a fixed UTC cohort start and end. Use the same values in every
   `params` CTE in `docs/analytics/deck-to-game-activation.sql`.
2. Confirm migrations, the guest host/join staging smoke, authenticated
   create/join regression, and Grafana provisioning all pass.
3. Confirm automation uses a session prefix of `test-`, `e2e-`,
   `playwright-`, `smoke-`, or `synthetic-`. Never filter generated guest
   usernames, UUID-shaped user IDs, or all guest traffic.
4. Keep `DECK_PROVIDER_ENABLED=false` until the documented human terms review
   and allowlist check are complete. Paste import is the supported fallback.
5. Confirm `GUEST_CREATION_ENABLED=true`, metrics are scraping, and product
   event drops are zero before opening the cohort.

## Minimum cohort and review window

- Do not make a product go/no-go call before **50 distinct
  `quick_start_viewed` sessions and 50 distinct valid `invite_viewed`
  sessions** survive the automation filter.
- A valid invite view is emitted only after the public invite lookup succeeds;
  missing, finished, and full-table pages are not denominator traffic.
- Freeze the acquisition cohort at the recorded end time. Allow seven more
  days before judging account claims; the claim query defaults to a matured
  14–7-day activation cohort for this reason.
- Inspect daily while collecting, but do not move the window or redefine a
  denominator in response to the result.

## Product go/no-go

Run `docs/analytics/deck-to-game-activation.sql` with the fixed window.

| Signal | Go threshold |
| --- | --- |
| Host activation | ≥60% of quick-start sessions reach host `board_ready` |
| Host time to board | p50 ≤60s and p90 ≤120s |
| Invite activation | ≥70% of valid invite viewers reach authoritative `player_joined` |
| Invite time to board | p50 ≤45s and p90 ≤90s |
| Valid entry resolution | ≥98% resolved or explicitly identified unresolved |
| Create/join technical errors | <2% of requests over the cohort |
| Guest claim | ≥15% of activated guests claim within seven days |

**Go:** both minimum denominators are met, every launch threshold passes, no
privacy or data-integrity incident is open, and event drops do not invalidate
the readout.

**Hold:** a denominator is short, the seven-day claim window is immature, or
telemetry gaps prevent a trustworthy calculation. Extend collection without
changing the cohort definition.

**No-go / rollback:** create or join errors reach 2%, board readiness is
materially degraded, a security/privacy issue appears, or the core host/invite
funnels miss target without an evidenced near-term correction.

## Technical watch and response

Use the activation panels in **vEDH App Overview**.

### Provider failure

- Watch `vedh_deck_provider_fetch_total` by `provider,outcome` and p90 fetch
  latency. Page when failures exceed 5% with at least 10 fetches in 15 minutes,
  or when the p90 approaches the eight-second total timeout.
- Set `DECK_PROVIDER_ENABLED=false`; do not relax SSRF, redirect, timeout, or
  response-size controls. Verify the same UI offers pasted text and that pasted
  preview still succeeds.

### Create/join errors

- Calculate non-success / all attempts separately for
  `vedh_game_create_total` and `vedh_game_join_total`. Page at ≥2% over 15
  minutes once at least 20 requests exist; at lower volume, review every failure
  individually and use the full cohort rate for the gate.
- Inspect bounded outcomes, correlated server logs, database health, and p90
  latency. Do not add raw error strings or identifiers as metric labels.
- If failures persist, stop new cohort admission and roll back the application
  release. Existing tables should remain available.

### Subscription readiness

- `vedh_board_activation_total{outcome="degraded"}` means the board became
  usable through bounded polling rather than realtime. Warn if three degraded
  activations occur in 15 minutes or if degraded exceeds 5% of readiness events.
- Verify websocket routing/origin configuration, then use the in-board reconnect
  action. Do not count navigation or query data alone as readiness.
- A sustained degraded state blocks go-live even if polling keeps the board
  usable.

### Measurement health

- Any sustained increase in `vedh_product_events_dropped_total` requires review.
  `write_error` can undercount the funnel; `unknown_key` or
  `client_authoritative` indicates a client/server contract mismatch.
- Compare event volume with request volume before interpreting a conversion dip.

## Kill switches and recovery

- **Provider:** `DECK_PROVIDER_ENABLED=false` immediately returns the URL path
  to paste-only behavior. Keep `DECK_PROVIDER_ALLOWED_HOSTS` narrow.
- **Guest creation:** `GUEST_CREATION_ENABLED=false` stops new guest sessions.
  Use it for abuse or guest-identity incidents; authenticated players remain the
  recovery path.
- **Application rollback:** deploy the last verified release if create/join or
  board readiness is unsafe. There is no broad client-side activation flag; do
  not improvise one during an incident.
- After any switch or rollback, rerun the staging smoke, verify `/prometheus`,
  and annotate the cohort window so affected traffic is not silently mixed into
  the decision.


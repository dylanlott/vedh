# Product Event Vocabulary

This is the documented form of `pkg/telemetry.Vocabulary` — the closed,
exactly-15-event product-event vocabulary (D-20) with each event's own
metadata key allowlist (D-21). It is transcribed from the "Product Event
Vocabulary" table in `docs/product/2026-07-23-deck-to-game-activation-prd.md`,
which is authoritative.

Fields the PRD's "Required fields" column names that already map to a
dedicated `product_events` column — session ID, game ID, role, source, and
duration (elapsed milliseconds) — are carried on that column, never
duplicated as a metadata key. Only fields with no dedicated column become
allowlisted metadata keys below. "Campaign source" is rendered as the four
allowlisted attribution keys `utm_source`, `utm_medium`, `utm_campaign`,
`referrer_host`.

## The 15 events

| Event | Client / Server | Allowlisted metadata keys | Provenance |
|---|---|---|---|
| `landing_primary_cta` | Client | `utm_source`, `utm_medium`, `utm_campaign`, `referrer_host` | PRD row (`landing_primary_cta`, "campaign source") |
| `quick_start_viewed` | Client | *(none — session ID only, on the column)* | PRD row (`quick_start_viewed`, "session ID") |
| `deck_import_started` | Client | *(none — source type carried on the `source` column)* | PRD row (`deck_import_started`, "session ID, source type") |
| `game_create_started` | Client | *(none — PRD lists only "session ID")* | PRD row (`game_create_started`, "session ID") |
| `invite_copied` | Client | `share_method` | REQ-A4 (records which share mechanism succeeded: native share vs. clipboard fallback) — no PRD row lists this key; the PRD row for `invite_copied` names only "session ID, game ID" |
| `invite_viewed` | Client | `utm_source`, `utm_medium`, `utm_campaign`, `referrer_host` | PRD row (`invite_viewed`, "session ID, game ID, campaign source") |
| `join_started` | Client | *(none — PRD lists only "session ID, game ID")* | PRD row (`join_started`, "session ID, game ID") |
| `board_ready` | Client | *(none — elapsed ms carried on the `duration_ms` column)* | PRD row (`board_ready`, "session ID, game ID, user role, elapsed milliseconds") |
| `account_claim_started` | Client | *(none — PRD lists only "session ID")* | PRD row (`account_claim_started`, "session ID") |
| `deck_import_succeeded` | Server | `card_count`, `unresolved_count` | PRD row (`deck_import_succeeded`, "session ID, source type, duration, card count, unresolved count") |
| `deck_import_failed` | Server | `reason` | PRD row (`deck_import_failed`, "session ID, source type, normalized reason") |
| `guest_session_created` | Server | *(none — PRD lists only "session ID")* | PRD row (`guest_session_created`, "session ID") |
| `game_created` | Server | *(none — user role carried on the `role` column)* | PRD row (`game_created`, "session ID, game ID, user role") |
| `player_joined` | Server | *(none — PRD lists only "session ID, game ID")* | PRD row (`player_joined`, "session ID, game ID") |
| `account_claimed` | Server | *(none — PRD lists only "session ID")* | PRD row (`account_claimed`, "session ID") |

The six **Server** rows above are exactly the six the vocabulary marks
`Authoritative: true` — a client attempting to submit one of these through
`trackProductEvent` is rejected with `RejectionClientAuthoritative`
(`client_authoritative`).

Two different things are both called "authoritative" in this phase, and
they are not the same set. `EventSpec.Authoritative` marks these **six**
server-owned names. The `product_events` migration's dedup index,
`product_events_authoritative_once`, names a **different, four-member**
set — `game_created`, `player_joined`, `guest_session_created`,
`account_claimed` — deliberately excluding `deck_import_succeeded` and
`deck_import_failed`, because a player may legitimately import several
times. Both names keep the older word for continuity with
`01-VALIDATION.md` and `01-RESEARCH.md`; read either as "the deduplicated
four," never as evidence the two senses are the same set.

## Discrepancies against `01-RESEARCH.md` Pattern 5

`01-RESEARCH.md` Pattern 5 prints an `[ASSUMED]` vocabulary table (marked
`[ASSUMED]` for 13 of its 15 rows) that this plan was explicitly told not
to adopt. Closing that assumption on the record, rather than in someone's
memory:

- **Column duplication.** Pattern 5's table lists `session_id`, `game_id`,
  `role`, `source`, and `duration_ms` (or `elapsed_ms`) as *metadata keys*
  on nearly every row — for example `board_ready: keys("session_id",
  "game_id", "role", "elapsed_ms")`. Every one of those fields already has
  a dedicated `product_events` column, so this plan carries each on its
  column instead, and the corresponding metadata key set is empty (or
  smaller) for the affected events: `quick_start_viewed`,
  `deck_import_started`, `game_create_started`, `join_started`,
  `board_ready`, `account_claim_started`, `guest_session_created`,
  `game_created`, `player_joined`, `account_claimed`, and the two
  `deck_import_*` events (which keep only their non-column fields —
  `card_count`/`unresolved_count`, and `reason`, respectively).
- **`invite_viewed` omitted its attribution keys.** Pattern 5 assumed
  `keys("session_id", "game_id")` only, dropping the PRD's "campaign
  source" requirement for this row entirely. This plan corrects that gap
  and allowlists the four attribution keys on `invite_viewed`, matching
  the PRD row.
- **`game_create_started`'s assumed `role` key has no PRD support.**
  Pattern 5 assumed `keys("session_id", "role")`; the PRD row for
  `game_create_started` lists only "session ID." No metadata key survives
  transcription for this event.
- **`invite_copied`'s `source` key has no PRD support.** Pattern 5 assumed
  `keys("session_id", "game_id", "share_method", "source")`; the PRD row
  lists only "session ID, game ID" — no `source` field at all for this
  event, and no `share_method` either. `share_method` is kept in this
  plan's vocabulary, but its provenance is REQ-A4, not the PRD table (see
  the table above); `source` is dropped, since nothing requires it.

No other discrepancy was found: every other row's transcribed key set
matches what the PRD's prose requires once the column-duplication rows are
accounted for.

## Rules enforced by `pkg/telemetry.ValidateEvent`

- **Closed vocabulary.** An event name outside the 15 above is refused with
  `RejectionUnknownEvent` (`unknown_event`).
- **Per-event allowlist (D-21).** A metadata key not in *that event's own*
  `Keys` set is refused with `RejectionUnknownKey` (`unknown_key`), even if
  the same key is allowlisted on a different event. The allowlist is never
  a global union.
- **Byte-equality comparison.** Event names and metadata keys are compared
  as raw UTF-8 bytes — no case folding, no Unicode normalization. An event
  name differing from a vocabulary entry only in letter case is not in the
  vocabulary and is refused as unknown.
- **`MaxMetadataValueBytes = 128`.** A metadata value whose byte length
  (via `len`, not rune count) exceeds this limit is refused with
  `RejectionOversizedValue` (`oversized_value`). A value exactly at the
  limit is accepted.
- **Missing-session refusal.** `product_events.session_id` is `NOT NULL`,
  and a whitespace-only string satisfies GraphQL's `String!` while
  satisfying nothing else — an event with no session cannot join any
  funnel. `strings.TrimSpace(sessionID) == ""` is refused with
  `RejectionMissingSession` (`missing_session`). This is the one place a
  whitespace trim is applied; the retained value is never otherwise
  normalized.
- **Server-authoritative refusal.** A client-submitted event whose
  `EventSpec.Authoritative` is `true` is refused with
  `RejectionClientAuthoritative` (`client_authoritative`), regardless of
  whether the tuple would otherwise be valid.

## Drop-and-count behaviour an operator will see

Every refusal and every write failure increments
`vedh_product_events_dropped_total{reason=...}` by exactly one — never zero,
never two — and writes no row. The `reason` label is always one of the
bounded `Rejection` strings above, plus `write_error` for an insert that
passed validation but failed at the database. `write_error` is a member of
this same bounded set rather than a second, unbounded label value — an
insert failure and a validation refusal are told apart by the label, not
by two different counters. A successful write increments
`vedh_product_events_written_total{outcome=...}` instead, where `outcome`
comes only from a server-owned event's own field, never forwarded from
client input.

## Metric families and where they are observed

`pkg/telemetry.NewCollectors` declares all four families ROADMAP Phase 1
criterion 4 names, plus the two product-event counters — twelve collectors
in total. Declaring a `CounterVec`/`HistogramVec` does not, by itself,
close criterion 4: an unobserved vector exports no child series at
`/prometheus` until an emit site calls one of `Collectors`' `Observe*`
methods.

| Family | Counter | Histogram | Labels | First observed |
|---|---|---|---|---|
| Deck import | `vedh_deck_import_total` | `vedh_deck_import_duration_seconds` | `source`, `outcome` | This plan (01-01) |
| Guest session | `vedh_guest_session_total` | `vedh_guest_session_duration_seconds` | `outcome` | ACT-005, Phase 2 |
| Game create | `vedh_game_create_total` | `vedh_game_create_duration_seconds` | `outcome` | ACT-006, Phase 2 |
| Game join | `vedh_game_join_total` | `vedh_game_join_duration_seconds` | `outcome` | ACT-008, Phase 3 |
| Board activation | `vedh_board_activation_total` | `vedh_board_activation_duration_seconds` | `role`, `outcome` (total); `role` (histogram) | ACT-009, Phase 3 |
| Product events | `vedh_product_events_written_total` | — | `outcome` | This plan (01-01) |
| Product events (dropped) | `vedh_product_events_dropped_total` | — | `reason` | This plan (01-01) |

Guest session, game create, game join, and board activation export **no
child series** at `/prometheus` until their Phase 2/3 emit sites land — an
operator scraping the endpoint after Phase 1 will see the deck-import and
product-event families only. That is expected, and is exactly why
`pkg/telemetry/metrics_test.go`'s registry-walk tests gather from a
registry they exercise themselves, rather than from a real scrape that
would be vacuously green for these three families.

`Role`'s two values, `host` and `invitee`, are a discretionary choice
recorded here because the PRD's vocabulary table names a "user role" field
without enumerating its values.

Every label name used above is a member of
`pkg/telemetry.AllowedLabelNames` (`source`, `outcome`, `role`, `reason`,
`provider`, `surface`). No collector declares a label for a username,
session ID, user ID, game ID, or attribution value — PostgreSQL is the
source for product-funnel analysis; Prometheus/Grafana cover technical
SLIs only.

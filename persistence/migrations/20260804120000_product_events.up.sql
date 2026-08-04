CREATE TABLE IF NOT EXISTS product_events (
    id          BIGSERIAL PRIMARY KEY,
    event_name  TEXT        NOT NULL,
    session_id  TEXT        NOT NULL,
    user_id     VARCHAR(255),
    game_id     TEXT,
    role        TEXT,
    source      TEXT,
    outcome     TEXT,
    duration_ms INTEGER,
    metadata    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Funnel step counts grouped by day / role / source / outcome.
CREATE INDEX IF NOT EXISTS product_events_name_time_idx
    ON product_events (event_name, occurred_at DESC);

-- Correlates quick_start_viewed -> board_ready within one session and drives
-- the time-to-board percentile queries.
CREATE INDEX IF NOT EXISTS product_events_session_time_idx
    ON product_events (session_id, occurred_at);

-- Dedup for the four server-authoritative conversion events that must never
-- double-count a retried client request. PostgreSQL 14 has no
-- `UNIQUE NULLS NOT DISTINCT` (added in PG15), so without the COALESCE
-- wrappers two rows with NULL game_id/user_id would be distinct and this
-- index would deduplicate nothing. deck_import_succeeded/deck_import_failed
-- are deliberately excluded: a player may legitimately import more than once.
CREATE UNIQUE INDEX IF NOT EXISTS product_events_authoritative_once
    ON product_events (event_name,
                       COALESCE(game_id, ''),
                       COALESCE(user_id, ''),
                       session_id)
    WHERE event_name IN ('game_created',
                         'player_joined',
                         'guest_session_created',
                         'account_claimed');

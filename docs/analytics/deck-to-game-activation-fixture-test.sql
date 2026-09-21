\set ON_ERROR_STOP on
BEGIN;
SET LOCAL search_path = pg_temp, public;
CREATE TEMP TABLE product_events (
    id bigint,
    event_name text NOT NULL,
    session_id text NOT NULL,
    user_id varchar(255),
    game_id text,
    role text,
    source text,
    outcome text,
    duration_ms integer,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurred_at timestamptz NOT NULL
);

INSERT INTO product_events
    (event_name, session_id, user_id, game_id, role, source, outcome, duration_ms, metadata, occurred_at)
VALUES
    ('landing_primary_cta', 'fixture-host', NULL, NULL, NULL, 'landing', NULL, NULL, '{"utm_source":"fixture"}', now() - interval '2 hours'),
    ('quick_start_viewed', 'fixture-host', 'guest-host', NULL, 'host', 'quick_start', NULL, NULL, '{}', now() - interval '119 minutes'),
    ('deck_import_succeeded', 'fixture-host', 'guest-host', NULL, NULL, 'plain_text', 'success', 100, '{"card_count":"100","unresolved_count":"0"}', now() - interval '118 minutes'),
    ('game_created', 'fixture-host', 'guest-host', 'fixture-game', 'host', 'quick_start', 'success', NULL, '{}', now() - interval '117 minutes'),
    ('board_ready', 'fixture-host', 'guest-host', 'fixture-game', 'host', 'board', 'ready', 180000, '{}', now() - interval '116 minutes'),
    ('invite_viewed', 'fixture-invite', 'guest-invite', 'fixture-game', 'invitee', 'invite', NULL, NULL, '{"referrer_host":"fixture.example"}', now() - interval '90 minutes'),
    ('player_joined', 'fixture-invite', 'guest-invite', 'fixture-game', 'invitee', 'invite', 'success', NULL, '{}', now() - interval '89 minutes'),
    ('board_ready', 'fixture-invite', 'guest-invite', 'fixture-game', 'invitee', 'board', 'degraded', 120000, '{}', now() - interval '88 minutes'),
    ('quick_start_viewed', 'e2e-excluded', 'bot', NULL, 'host', 'test', NULL, NULL, '{}', now() - interval '60 minutes'),
    ('board_ready', 'e2e-excluded', 'bot', 'bot-game', 'host', 'test', 'ready', 1, '{}', now() - interval '59 minutes'),
    ('guest_session_created', 'fixture-claim', 'guest-claim', NULL, NULL, NULL, NULL, NULL, '{}', now() - interval '10 days'),
    ('board_ready', 'fixture-claim', 'guest-claim', 'old-game', 'host', 'board', 'ready', 1000, '{}', now() - interval '9 days'),
    ('account_claimed', 'fixture-claim', 'guest-claim', NULL, NULL, NULL, NULL, NULL, '{}', now() - interval '8 days');

\ir deck-to-game-activation.sql

DO $$
BEGIN
    IF (SELECT count(*) FROM product_events WHERE session_id = 'e2e-excluded') <> 2 THEN
        RAISE EXCEPTION 'fixture setup failed';
    END IF;
    IF (SELECT count(DISTINCT session_id) FROM product_events
        WHERE session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
          AND event_name IN ('quick_start_viewed', 'invite_viewed')) <> 2 THEN
        RAISE EXCEPTION 'automation filter did not preserve the two human-shaped fixture sessions';
    END IF;
END $$;

ROLLBACK;

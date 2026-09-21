-- Returns zero only when the guest browser release journey emitted every
-- expected funnel event exactly once in its original host/invitee sessions.
-- Account-claim refresh uses a separate session so a second board_ready does
-- not contaminate the activation denominator.
WITH expected(actor, event_name) AS (
    VALUES
        ('guest-host', 'quick_start_viewed'),
        ('guest-host', 'deck_import_started'),
        ('guest-host', 'deck_import_succeeded'),
        ('guest-host', 'guest_session_created'),
        ('guest-host', 'game_create_started'),
        ('guest-host', 'game_created'),
        ('guest-host', 'board_ready'),
        ('guest-host', 'account_claim_started'),
        ('guest-host', 'account_claimed'),
        ('guest-invitee', 'invite_viewed'),
        ('guest-invitee', 'deck_import_started'),
        ('guest-invitee', 'deck_import_succeeded'),
        ('guest-invitee', 'guest_session_created'),
        ('guest-invitee', 'join_started'),
        ('guest-invitee', 'player_joined'),
        ('guest-invitee', 'board_ready')
),
expected_sessions AS (
    SELECT actor,
           event_name,
           'e2e-' || :'run_id' || '-' || actor AS session_id
    FROM expected
),
actual AS (
    SELECT session_id, event_name, count(*) AS event_count
    FROM product_events
    WHERE session_id IN (SELECT session_id FROM expected_sessions)
    GROUP BY session_id, event_name
)
SELECT count(*)
FROM expected_sessions e
LEFT JOIN actual a USING (session_id, event_name)
WHERE COALESCE(a.event_count, 0) <> 1;

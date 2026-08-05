-- Host activation funnel example: quick_start_viewed -> board_ready,
-- grouped by day, source, role, and outcome, with p50/p90 time-to-board.
--
-- Depends on product_events_session_time_idx (session_id, occurred_at),
-- which both CTEs below scan to find each session's earliest occurrence of
-- the event they group on.
--
-- This is the ACT-001 example query. ACT-012 (Phase 4) owns the versioned
-- production query set at docs/analytics/deck-to-game-activation.sql; this
-- file is the example only, not that set.
--
-- Grouping by min(occurred_at) per session — rather than counting raw rows
-- — is what makes a duplicate repeatable event (e.g. two board_ready
-- deliveries for one session under D-19's one-mutation-per-event
-- semantics) collapse to one counted activation instead of two, and is
-- also what makes a byte-identical occurred_at tie unable to change the
-- result: see server/product_events_test.go's
-- TestProductEvents_OccurredAtTieOrdering for the proof.

WITH viewed AS (
    SELECT session_id,
           min(occurred_at) AS viewed_at,
           min(source)      AS source
    FROM   product_events
    WHERE  event_name = 'quick_start_viewed'
    GROUP  BY session_id
),
ready AS (
    SELECT session_id,
           min(occurred_at) AS ready_at,
           min(role)        AS role
    FROM   product_events
    WHERE  event_name = 'board_ready'
      AND  role = 'host'
    GROUP  BY session_id
)
SELECT date_trunc('day', v.viewed_at AT TIME ZONE 'UTC')                        AS day,
       coalesce(v.source, 'unknown')                                            AS source,
       coalesce(r.role, 'host')                                                 AS role,
       CASE WHEN r.session_id IS NULL THEN 'not_activated' ELSE 'activated' END AS outcome,
       count(*)                                                                 AS sessions,
       percentile_cont(0.5) WITHIN GROUP (
           ORDER BY extract(epoch FROM r.ready_at - v.viewed_at)
       )                                                                        AS p50_seconds,
       percentile_cont(0.9) WITHIN GROUP (
           ORDER BY extract(epoch FROM r.ready_at - v.viewed_at)
       )                                                                        AS p90_seconds
FROM      viewed v
LEFT JOIN ready  r USING (session_id)
GROUP BY 1, 2, 3, 4
ORDER BY 1 DESC, 2, 3, 4;

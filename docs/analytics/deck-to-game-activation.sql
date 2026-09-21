-- vEDH deck-to-game activation readout (ACT-012)
--
-- PostgreSQL is the product-funnel source of truth. Prometheus is reserved
-- for technical request health and latency; see the quiet-beta runbook.
-- Replace each params CTE with the same fixed UTC cohort_start/cohort_end
-- before a go/no-go review. The rolling defaults are for operator previews.
--
-- Automation contract: CI, Playwright, smoke, and synthetic traffic MUST use
-- a session_id beginning test-, e2e-, playwright-, smoke-, or synthetic-.
-- Ordinary guest UUIDs and generated guest usernames are never filtered.

-- 1. Host and valid-invite conversion rates.
WITH params AS (
    SELECT now() - interval '7 days' AS cohort_start, now() AS cohort_end
),
filtered_events AS (
    SELECT e.*
    FROM product_events e, params p
    WHERE e.occurred_at >= p.cohort_start
      AND e.occurred_at < p.cohort_end
      AND e.session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
      AND COALESCE(e.source, '') <> 'test'
),
starts AS (
    SELECT 'host'::text AS funnel, session_id, min(occurred_at) AS started_at
    FROM filtered_events
    WHERE event_name = 'quick_start_viewed'
    GROUP BY session_id
    UNION ALL
    SELECT 'invite'::text AS funnel, session_id, min(occurred_at) AS started_at
    FROM filtered_events
    WHERE event_name = 'invite_viewed'
    GROUP BY session_id
),
conversions AS (
    SELECT s.funnel, s.session_id, min(e.occurred_at) AS converted_at
    FROM starts s
    JOIN filtered_events e
      ON e.session_id = s.session_id
     AND e.occurred_at >= s.started_at
     AND ((s.funnel = 'host' AND e.event_name = 'board_ready' AND e.role = 'host')
       OR (s.funnel = 'invite' AND e.event_name = 'player_joined'))
    GROUP BY s.funnel, s.session_id
)
SELECT s.funnel,
       count(*) AS starts_or_valid_views,
       count(c.session_id) AS conversions,
       round(100.0 * count(c.session_id) / NULLIF(count(*), 0), 2) AS conversion_percent
FROM starts s
LEFT JOIN conversions c USING (funnel, session_id)
GROUP BY s.funnel
ORDER BY s.funnel;

-- 2. End-to-end time to a usable board (p50/p90 in seconds).
WITH params AS (
    SELECT now() - interval '7 days' AS cohort_start, now() AS cohort_end
),
filtered_events AS (
    SELECT e.*
    FROM product_events e, params p
    WHERE e.occurred_at >= p.cohort_start
      AND e.occurred_at < p.cohort_end
      AND e.session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
      AND COALESCE(e.source, '') <> 'test'
),
starts AS (
    SELECT 'host'::text AS funnel, session_id, min(occurred_at) AS started_at
    FROM filtered_events WHERE event_name = 'quick_start_viewed' GROUP BY session_id
    UNION ALL
    SELECT 'invite'::text AS funnel, session_id, min(occurred_at) AS started_at
    FROM filtered_events WHERE event_name = 'invite_viewed' GROUP BY session_id
),
boards AS (
    SELECT s.funnel, s.session_id, s.started_at, min(e.occurred_at) AS board_at
    FROM starts s
    JOIN filtered_events e
      ON e.session_id = s.session_id
     AND e.event_name = 'board_ready'
     AND e.role = CASE WHEN s.funnel = 'host' THEN 'host' ELSE 'invitee' END
     AND e.occurred_at >= s.started_at
    GROUP BY s.funnel, s.session_id, s.started_at
),
durations AS (
    SELECT funnel, extract(epoch FROM board_at - started_at) AS seconds_to_board
    FROM boards
)
SELECT funnel,
       count(*) AS activated_sessions,
       round(percentile_cont(0.50) WITHIN GROUP (ORDER BY seconds_to_board)::numeric, 2) AS p50_seconds,
       round(percentile_cont(0.90) WITHIN GROUP (ORDER BY seconds_to_board)::numeric, 2) AS p90_seconds
FROM durations
GROUP BY funnel
ORDER BY funnel;

-- 3. Import quality and normalized failure reasons by source.
WITH params AS (
    SELECT now() - interval '7 days' AS cohort_start, now() AS cohort_end
),
filtered_events AS (
    SELECT e.*
    FROM product_events e, params p
    WHERE e.occurred_at >= p.cohort_start
      AND e.occurred_at < p.cohort_end
      AND e.session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
      AND COALESCE(e.source, '') <> 'test'
),
imports AS (
    SELECT event_name,
           COALESCE(source, 'unknown') AS source,
           COALESCE(metadata->>'reason', 'none') AS reason,
           CASE WHEN metadata->>'card_count' ~ '^\d+$' THEN (metadata->>'card_count')::bigint ELSE 0 END AS card_count,
           CASE WHEN metadata->>'unresolved_count' ~ '^\d+$' THEN (metadata->>'unresolved_count')::bigint ELSE 0 END AS unresolved_count
    FROM filtered_events
    WHERE event_name IN ('deck_import_succeeded', 'deck_import_failed')
)
SELECT source,
       event_name,
       reason,
       count(*) AS attempts,
       sum(card_count) AS card_entries,
       sum(unresolved_count) AS unresolved_entries,
       round(100.0 * (sum(card_count) - sum(unresolved_count)) / NULLIF(sum(card_count), 0), 2) AS resolved_percent
FROM imports
GROUP BY source, event_name, reason
ORDER BY source, event_name, reason;

-- 4. Acquisition source joined to the host/invite conversion outcome.
WITH params AS (
    SELECT now() - interval '7 days' AS cohort_start, now() AS cohort_end
),
filtered_events AS (
    SELECT e.*
    FROM product_events e, params p
    WHERE e.occurred_at >= p.cohort_start
      AND e.occurred_at < p.cohort_end
      AND e.session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
      AND COALESCE(e.source, '') <> 'test'
),
starts AS (
    SELECT 'host'::text AS funnel,
           q.session_id,
           q.started_at,
           COALESCE(l.metadata->>'utm_source', l.metadata->>'referrer_host', 'direct') AS acquisition_source
    FROM (
        SELECT session_id, min(occurred_at) AS started_at
        FROM filtered_events WHERE event_name = 'quick_start_viewed' GROUP BY session_id
    ) q
    LEFT JOIN LATERAL (
        SELECT metadata FROM filtered_events
        WHERE session_id = q.session_id AND event_name = 'landing_primary_cta'
        ORDER BY occurred_at LIMIT 1
    ) l ON true
    UNION ALL
    SELECT 'invite'::text,
           i.session_id,
           i.occurred_at,
           COALESCE(i.metadata->>'utm_source', i.metadata->>'referrer_host', 'direct')
    FROM (
        SELECT DISTINCT ON (session_id) session_id, occurred_at, metadata
        FROM filtered_events WHERE event_name = 'invite_viewed'
        ORDER BY session_id, occurred_at
    ) i
),
outcomes AS (
    SELECT s.*,
           EXISTS (
               SELECT 1 FROM filtered_events e
               WHERE e.session_id = s.session_id
                 AND e.occurred_at >= s.started_at
                 AND ((s.funnel = 'host' AND e.event_name = 'board_ready' AND e.role = 'host')
                   OR (s.funnel = 'invite' AND e.event_name = 'player_joined'))
           ) AS converted
    FROM starts s
)
SELECT funnel, acquisition_source, count(*) AS starts, count(*) FILTER (WHERE converted) AS conversions,
       round(100.0 * count(*) FILTER (WHERE converted) / NULLIF(count(*), 0), 2) AS conversion_percent
FROM outcomes
GROUP BY funnel, acquisition_source
ORDER BY funnel, starts DESC, acquisition_source;

-- 5. Matured seven-day guest-claim cohort. This default examines guests who
-- activated 14–7 days ago, then allows the full following seven days to claim.
WITH params AS (
    SELECT now() - interval '14 days' AS cohort_start,
           now() - interval '7 days' AS cohort_end,
           now() AS observation_end
),
eligible_events AS (
    SELECT e.*
    FROM product_events e, params p
    WHERE e.occurred_at >= p.cohort_start
      AND e.occurred_at < p.observation_end
      AND e.session_id !~* '^(test|e2e|playwright|smoke|synthetic)([-_:]|$)'
      AND COALESCE(e.source, '') <> 'test'
),
activated_guests AS (
    SELECT b.user_id, min(b.occurred_at) AS activated_at
    FROM eligible_events b, params p
    WHERE b.event_name = 'board_ready'
      AND b.user_id IS NOT NULL
      AND b.occurred_at >= p.cohort_start
      AND b.occurred_at < p.cohort_end
      AND EXISTS (
          SELECT 1 FROM eligible_events g
          WHERE g.event_name = 'guest_session_created'
            AND g.user_id = b.user_id
            AND g.occurred_at <= b.occurred_at
      )
    GROUP BY b.user_id
),
claims AS (
    SELECT a.user_id, min(c.occurred_at) AS claimed_at
    FROM activated_guests a
    JOIN eligible_events c
      ON c.user_id = a.user_id
     AND c.event_name = 'account_claimed'
     AND c.occurred_at >= a.activated_at
     AND c.occurred_at < a.activated_at + interval '7 days'
    GROUP BY a.user_id
)
SELECT count(*) AS activated_guests,
       count(c.user_id) AS claimed_within_7d,
       round(100.0 * count(c.user_id) / NULLIF(count(*), 0), 2) AS claim_percent
FROM activated_guests a
LEFT JOIN claims c USING (user_id);


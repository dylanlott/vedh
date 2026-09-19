# Dashboard Business Metrics Update

This is the next instrumentation layer after the current operational dashboards.

Right now the live dashboards are honest but mostly operational:
- `vEDH App Overview` is runtime + metrics-handler health
- `Jank App Overview` has its first real route/business-ish slice because Jank now emits route-level request metrics
- `Platform Hub Overview` is a cross-app control surface, not yet a product-value surface

The goal of this update is to add **real product and business signal** without drifting into fake KPI theater.

## vEDH App Overview

### Current ceiling

The live `vedh-api` scrape exposes:
- scrape health
- Prometheus handler request counts/in-flight
- Go runtime/process metrics

It does **not** currently expose gameplay, room, auth, join, or retention metrics. That means the dashboard cannot yet answer the questions that matter to the product:
- are people successfully creating games?
- are joins succeeding?
- where are users bouncing?
- what play patterns actually stick?

### Proposed metrics contract

Add app-level counters/histograms in `vedh-api` for the real user journey:

- `vedh_signups_total{source, result}`
- `vedh_logins_total{result}`
- `vedh_games_created_total{source}`
- `vedh_game_join_attempts_total{result}`
- `vedh_game_join_latency_seconds_bucket{result}`
- `vedh_active_games_current`
- `vedh_player_actions_total{action}`
- `vedh_workspace_context_requests_total{result}`
- `vedh_checkout_sessions_total{plan, result}`
- `vedh_paid_workspaces_current{plan}`

Notes:
- if billing/workspace-context lives outside current public vedh scope, keep those counters behind the relevant service boundary instead of pretending they belong in gameplay
- use `result` labels like `success`, `validation_error`, `auth_error`, `not_found`, `server_error`
- keep label cardinality low; never put game IDs or user IDs in labels

### Dashboard update after instrumentation

Add a second row focused on product health:
- Signups / hour
- Game creates / hour
- Join success rate
- Join p95 latency
- Active games current
- Top player action mix

Best derived business read:
- created-to-joined ratio
- join failure concentration
- active games per recent signup cohort

## Jank App Overview

### Current ceiling

Jank now has the first meaningful product-oriented metrics surface because route-level request counters exist. That lets us see `/`, `/login`, and `/search` usage, but it still stops short of the real forum questions:
- are people signing up or only browsing?
- are searches useful or dead ends?
- are threads/posts actually getting created?
- are card-tree views becoming a product wedge or just a nice demo?

### Proposed metrics contract

Add event counters around the real forum funnel:

- `jank_page_views_total{route}`
- `jank_search_queries_total{result}`
- `jank_login_attempts_total{result}`
- `jank_signup_attempts_total{result}`
- `jank_threads_created_total{board}`
- `jank_posts_created_total{board}`
- `jank_reports_created_total{category}`
- `jank_card_tree_views_total{scope}`
- `jank_card_tree_creates_total{scope}`
- `jank_visual_tree_filter_uses_total`

Notes:
- normalize `board` into a small stable label set or omit it if cardinality explodes
- `result` should distinguish `success`, `invalid_credentials`, `validation_error`, `server_error`, `empty_result`, etc.
- route-level counters already exist; these new metrics should model **intent**, not duplicate raw transport

### Dashboard update after instrumentation

Add a second product row:
- Login success rate
- Signup success rate
- Search queries / hour
- Empty-result search share
- Threads created / hour
- Posts created / hour
- Card-tree views / hour
- Card-tree create rate

Best derived business read:
- search-to-thread/create conversion
- browse-heavy vs contribution-heavy usage
- whether card trees are a novelty or a durable behavior loop

## Platform Hub Overview

### Current ceiling

The hub currently answers:
- are the core apps up?
- is the host under pressure?
- which app is noisier right now?

That is useful operator context, but not founder context.

### Proposed metrics contract

The hub should stay mostly derived rather than inventing lots of new cross-app emitters. Prefer dashboard-level composition of the app metrics above plus one optional rollup metric:

- `platform_business_value_score`

This should be a **derived score**, not a primary source of truth. Example inputs:
- weighted recent `vedh_games_created_total`
- weighted successful `vedh_game_join_attempts_total`
- weighted `jank_threads_created_total`
- weighted `jank_posts_created_total`
- weighted `jank_card_tree_creates_total`

If you do emit a real rollup metric, make it obviously synthetic and document the formula.

### Dashboard update after instrumentation

Add a founder row to the hub:
- vEDH game creates / hour
- vEDH join success rate
- Jank threads created / hour
- Jank posts created / hour
- Jank search volume + empty-result share
- blended `platform_business_value_score`

Best derived business read:
- which app is creating more genuine usage right now
- whether traffic is turning into creation/contribution instead of passive browsing
- whether product work is improving the quality of usage, not just uptime

## Implementation order

1. **vEDH instrumentation** for create/join/action/auth events
2. **Jank instrumentation** for auth/search/thread/post/card-tree events
3. refresh app dashboards with a distinct business row on each
4. update the hub last so it composes real app-level business signals instead of placeholders

## Guardrails

- No user IDs, room IDs, thread IDs, search text, or other high-cardinality identifiers in labels
- Prefer counters/histograms over free-form logs when the question is dashboard-worthy
- Keep operational rows visible; business metrics should extend the dashboards, not replace the safety rails
- If a metric would be used for decision-making, document exactly where it is emitted and what increments it

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const { createJankAppDashboard } = require('./jankAppDashboard');

test('createJankAppDashboard emphasizes engagement and business outcomes instead of mostly scrape trivia', () => {
  const dashboard = createJankAppDashboard({ datasourceUid: 'test-vm-uid' });

  assert.equal(dashboard.title, 'Jank App Overview');
  assert.equal(dashboard.schemaVersion >= 39, true);
  assert.equal(Array.isArray(dashboard.panels), true);
  assert.equal(dashboard.panels.length, 25);
  assert.deepEqual(dashboard.time, { from: 'now-24h', to: 'now' });

  const panelsByTitle = new Map(dashboard.panels.map((panel) => [panel.title, panel]));
  const expectedPanels = [
    'Target Up',
    'Searches Last 24h',
    'User Base Since Deploy',
    'New Users Last 24h',
    'Daily Active Users',
    'Weekly Active Users',
    'Monthly Active Users',
    'Threads Created Last 24h',
    'Posts Created Last 24h',
    'Search Empty Share',
    'Signup Success Rate',
    'Login Success Rate',
    'Contribution Depth',
    'Card Trees Created Last 24h',
    'Search → Thread Yield',
    'Search Outcomes Last 24h',
    'Login Outcomes Last 24h',
    'Signup Outcomes Last 24h',
    'Creation Activity (6h rolling)',
    'Funnel Activity (6h rolling)',
    'Route Demand (6h rolling)',
    'Request Rate by Route',
    'P95 Request Latency by Route',
    'Resident Memory',
    'Process CPU Usage',
  ];

  for (const title of expectedPanels) {
    assert.ok(panelsByTitle.has(title), `expected panel ${title}`);
  }

  const targetUpPanel = panelsByTitle.get('Target Up');
  assert.equal(targetUpPanel.datasource.uid, 'test-vm-uid');
  assert.match(targetUpPanel.targets[0].expr, /^up\{job="jank-app"\}$/);

  const searchesPanel = panelsByTitle.get('Searches Last 24h');
  assert.match(searchesPanel.targets[0].expr, /^sum\(increase\(jank_search_queries_total\{job="jank-app"\}\[24h\]\)\) or vector\(0\)$/);

  const userBasePanel = panelsByTitle.get('User Base Since Deploy');
  assert.match(userBasePanel.targets[0].expr, /^max\(jank_users_total\{job="jank-app"\}\)$/);

  const signupsPanel = panelsByTitle.get('New Users Last 24h');
  assert.match(signupsPanel.targets[0].expr, /^sum\(increase\(jank_signup_attempts_total\{job="jank-app",result="success"\}\[24h\]\)\) or vector\(0\)$/);

  const dailyActivePanel = panelsByTitle.get('Daily Active Users');
  assert.match(dailyActivePanel.targets[0].expr, /^max\(jank_daily_active_users_total\{job="jank-app"\}\)$/);

  const weeklyActivePanel = panelsByTitle.get('Weekly Active Users');
  assert.match(weeklyActivePanel.targets[0].expr, /^max\(jank_weekly_active_users_total\{job="jank-app"\}\)$/);

  const monthlyActivePanel = panelsByTitle.get('Monthly Active Users');
  assert.match(monthlyActivePanel.targets[0].expr, /^max\(jank_monthly_active_users_total\{job="jank-app"\}\)$/);

  const threadsPanel = panelsByTitle.get('Threads Created Last 24h');
  assert.match(threadsPanel.targets[0].expr, /^sum\(increase\(jank_threads_created_total\{job="jank-app"\}\[24h\]\)\) or vector\(0\)$/);

  const postsPanel = panelsByTitle.get('Posts Created Last 24h');
  assert.match(postsPanel.targets[0].expr, /^sum\(increase\(jank_posts_created_total\{job="jank-app"\}\[24h\]\)\) or vector\(0\)$/);

  const emptySharePanel = panelsByTitle.get('Search Empty Share');
  assert.equal(emptySharePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(emptySharePanel.targets[0].expr, /result="empty"/);
  assert.match(emptySharePanel.targets[0].expr, /increase\(jank_search_queries_total\{job="jank-app"\}\[24h\]\)/);

  const signupRatePanel = panelsByTitle.get('Signup Success Rate');
  assert.equal(signupRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(signupRatePanel.targets[0].expr, /jank_signup_attempts_total\{job="jank-app",result="success"\}/);
  assert.match(signupRatePanel.targets[0].expr, /jank_signup_attempts_total\{job="jank-app"\}\[24h\]/);

  const loginRatePanel = panelsByTitle.get('Login Success Rate');
  assert.equal(loginRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(loginRatePanel.targets[0].expr, /jank_login_attempts_total\{job="jank-app",result="success"\}/);
  assert.match(loginRatePanel.targets[0].expr, /jank_login_attempts_total\{job="jank-app"\}\[24h\]/);

  const contributionDepthPanel = panelsByTitle.get('Contribution Depth');
  assert.match(contributionDepthPanel.targets[0].expr, /increase\(jank_posts_created_total\{job="jank-app"\}\[24h\]\)/);
  assert.match(contributionDepthPanel.targets[0].expr, /increase\(jank_threads_created_total\{job="jank-app"\}\[24h\]\)/);

  const cardTreesPanel = panelsByTitle.get('Card Trees Created Last 24h');
  assert.match(cardTreesPanel.targets[0].expr, /^sum\(increase\(jank_card_trees_created_total\{job="jank-app"\}\[24h\]\)\) or vector\(0\)$/);

  const searchYieldPanel = panelsByTitle.get('Search → Thread Yield');
  assert.equal(searchYieldPanel.fieldConfig.defaults.unit, 'percent');
  assert.match(searchYieldPanel.targets[0].expr, /increase\(jank_threads_created_total\{job="jank-app"\}\[24h\]\)/);
  assert.match(searchYieldPanel.targets[0].expr, /result="match"/);

  const searchOutcomesPanel = panelsByTitle.get('Search Outcomes Last 24h');
  assert.equal(searchOutcomesPanel.type, 'bargauge');
  assert.equal(searchOutcomesPanel.targets[0].legendFormat, '{{result}}');
  assert.match(searchOutcomesPanel.targets[0].expr, /^sum by \(result\) \(increase\(jank_search_queries_total\{job="jank-app"\}\[24h\]\)\)$/);

  const loginOutcomesPanel = panelsByTitle.get('Login Outcomes Last 24h');
  assert.equal(loginOutcomesPanel.type, 'bargauge');
  assert.equal(loginOutcomesPanel.targets[0].legendFormat, '{{surface}} • {{result}}');
  assert.match(loginOutcomesPanel.targets[0].expr, /^sum by \(surface,result\) \(increase\(jank_login_attempts_total\{job="jank-app"\}\[24h\]\)\)$/);

  const signupOutcomesPanel = panelsByTitle.get('Signup Outcomes Last 24h');
  assert.equal(signupOutcomesPanel.type, 'bargauge');
  assert.equal(signupOutcomesPanel.targets[0].legendFormat, '{{surface}} • {{result}}');
  assert.match(signupOutcomesPanel.targets[0].expr, /^sum by \(surface,result\) \(increase\(jank_signup_attempts_total\{job="jank-app"\}\[24h\]\)\)$/);

  const creationActivityPanel = panelsByTitle.get('Creation Activity (6h rolling)');
  assert.equal(creationActivityPanel.type, 'timeseries');
  assert.equal(creationActivityPanel.targets.length, 3);
  assert.match(creationActivityPanel.targets[0].expr, /increase\(jank_threads_created_total\{job="jank-app"\}\[6h\]\)/);
  assert.match(creationActivityPanel.targets[1].expr, /increase\(jank_posts_created_total\{job="jank-app"\}\[6h\]\)/);
  assert.match(creationActivityPanel.targets[2].expr, /increase\(jank_card_trees_created_total\{job="jank-app"\}\[6h\]\)/);

  const funnelActivityPanel = panelsByTitle.get('Funnel Activity (6h rolling)');
  assert.equal(funnelActivityPanel.type, 'timeseries');
  assert.equal(funnelActivityPanel.targets.length, 4);
  assert.match(funnelActivityPanel.targets[0].expr, /result="match"/);
  assert.match(funnelActivityPanel.targets[1].expr, /jank_signup_attempts_total\{job="jank-app",result="success"\}/);
  assert.match(funnelActivityPanel.targets[2].expr, /jank_login_attempts_total\{job="jank-app",result="success"\}/);
  assert.match(funnelActivityPanel.targets[3].expr, /jank_threads_created_total\{job="jank-app"\}/);

  const routeDemandPanel = panelsByTitle.get('Route Demand (6h rolling)');
  assert.equal(routeDemandPanel.type, 'timeseries');
  assert.match(routeDemandPanel.targets[0].expr, /sum by \(route\) \(increase\(jank_http_requests_total\{job="jank-app",route=~"\/|\/login|\/search",route!="\/metrics"\}\[6h\]\)\)/);

  const requestRatePanel = panelsByTitle.get('Request Rate by Route');
  assert.equal(requestRatePanel.fieldConfig.defaults.unit, 'reqps');
  assert.match(requestRatePanel.targets[0].expr, /sum by \(route,status_code\) \(rate\(jank_http_requests_total\{job="jank-app",route!="\/metrics"\}\[1h\]\)\)/);

  const latencyPanel = panelsByTitle.get('P95 Request Latency by Route');
  assert.equal(latencyPanel.fieldConfig.defaults.unit, 'ms');
  assert.match(latencyPanel.targets[0].expr, /1000 \* histogram_quantile\(0\.95, sum by \(le,route\) \(rate\(jank_http_request_duration_seconds_bucket\{job="jank-app",route!="\/metrics"\}\[1h\]\)\)\)/);

  const memoryPanel = panelsByTitle.get('Resident Memory');
  assert.match(memoryPanel.targets[0].expr, /^process_resident_memory_bytes\{job="jank-app"\}$/);

  const cpuPanel = panelsByTitle.get('Process CPU Usage');
  assert.match(cpuPanel.targets[0].expr, /^100 \* rate\(process_cpu_seconds_total\{job="jank-app"\}\[\$__rate_interval\]\)$/);
});

test('generated jank dashboard JSON stays in sync with the dashboard builder', () => {
  const generated = createJankAppDashboard({ datasourceUid: 'bfcctb05vkm4ge' });
  const jsonPath = path.join(__dirname, 'jank-app-overview.json');
  const saved = JSON.parse(fs.readFileSync(jsonPath, 'utf8'));

  assert.deepEqual(saved, generated);
});

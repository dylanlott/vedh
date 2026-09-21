const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const { createVedhAppDashboard } = require('./vedhAppDashboard');

test('createVedhAppDashboard emphasizes product engagement and game health', () => {
  const dashboard = createVedhAppDashboard({ datasourceUid: 'test-vm-uid' });

  assert.equal(dashboard.title, 'vEDH App Overview');
  assert.equal(dashboard.schemaVersion >= 39, true);
  assert.equal(Array.isArray(dashboard.panels), true);
  assert.equal(dashboard.panels.length, 32);
  assert.deepEqual(dashboard.time, { from: 'now-24h', to: 'now' });

  const panelsByTitle = new Map(dashboard.panels.map((panel) => [panel.title, panel]));
  const expectedPanels = [
    'Target Up',
    'User Base Since Deploy',
    'New Users Last 24h',
    'Daily Active Users',
    'Weekly Active Users',
    'Monthly Active Users',
    'Games Created Last 24h',
    'Successful Joins Last 24h',
    'Join Success Rate',
    'Signup Success Rate',
    'Login Success Rate',
    'Join Failures Last 24h',
    'Joins per Game Created',
    'Games by Format Last 24h',
    'Signup Outcomes Last 24h',
    'Login Outcomes Last 24h',
    'Join Outcomes Last 24h',
    'Funnel Activity (6h rolling)',
    'Game Format Activity (6h rolling)',
    'Join Quality (6h rolling)',
    'Resident Memory',
    'Process CPU Usage',
    'File Descriptors',
    'Go Heap vs Stack',
    'Public Request Decisions (5m rate)',
    'Deck Import Outcomes (5m rate)',
    'Game Create Outcomes (5m rate)',
    'Game Join Outcomes (5m rate)',
    'Activation Latency p90',
    'Provider Fetch Health (5m rate)',
    'Board Readiness Mode (15m)',
    'Product Event Drops (15m)',
  ];

  for (const title of expectedPanels) {
    assert.ok(panelsByTitle.has(title), `expected panel ${title}`);
  }

  const targetUpPanel = panelsByTitle.get('Target Up');
  assert.equal(targetUpPanel.datasource.uid, 'test-vm-uid');
  assert.match(targetUpPanel.targets[0].expr, /^up\{job="vedh-api"\}$/);
  assert.equal(targetUpPanel.fieldConfig.defaults.mappings[0].options['1'].text, 'Up');

  const userBasePanel = panelsByTitle.get('User Base Since Deploy');
  assert.match(userBasePanel.targets[0].expr, /^max\(vedh_users_total\{job="vedh-api"\}\)$/);

  const signup24hPanel = panelsByTitle.get('New Users Last 24h');
  assert.match(signup24hPanel.targets[0].expr, /^sum\(increase\(vedh_signups_total\{job="vedh-api",result="success"\}\[24h\]\)\) or vector\(0\)$/);

  const dailyActivePanel = panelsByTitle.get('Daily Active Users');
  assert.match(dailyActivePanel.targets[0].expr, /^max\(vedh_daily_active_users_total\{job="vedh-api"\}\)$/);

  const weeklyActivePanel = panelsByTitle.get('Weekly Active Users');
  assert.match(weeklyActivePanel.targets[0].expr, /^max\(vedh_weekly_active_users_total\{job="vedh-api"\}\)$/);

  const monthlyActivePanel = panelsByTitle.get('Monthly Active Users');
  assert.match(monthlyActivePanel.targets[0].expr, /^max\(vedh_monthly_active_users_total\{job="vedh-api"\}\)$/);

  const games24hPanel = panelsByTitle.get('Games Created Last 24h');
  assert.match(games24hPanel.targets[0].expr, /^sum\(increase\(vedh_games_created_total\{job="vedh-api"\}\[24h\]\)\) or vector\(0\)$/);

  const joins24hPanel = panelsByTitle.get('Successful Joins Last 24h');
  assert.match(joins24hPanel.targets[0].expr, /^sum\(increase\(vedh_game_join_attempts_total\{job="vedh-api",result="success"\}\[24h\]\)\) or vector\(0\)$/);

  const joinRatePanel = panelsByTitle.get('Join Success Rate');
  assert.equal(joinRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(joinRatePanel.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(joinRatePanel.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api"\}\[24h\]/);

  const signupRatePanel = panelsByTitle.get('Signup Success Rate');
  assert.equal(signupRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(signupRatePanel.targets[0].expr, /vedh_signups_total\{job="vedh-api",result="success"\}/);
  assert.match(signupRatePanel.targets[0].expr, /vedh_signups_total\{job="vedh-api"\}\[24h\]/);

  const loginRatePanel = panelsByTitle.get('Login Success Rate');
  assert.equal(loginRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(loginRatePanel.targets[0].expr, /vedh_login_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(loginRatePanel.targets[0].expr, /vedh_login_attempts_total\{job="vedh-api"\}\[24h\]/);

  const joinFailuresPanel = panelsByTitle.get('Join Failures Last 24h');
  assert.match(joinFailuresPanel.targets[0].expr, /^sum\(increase\(vedh_game_join_attempts_total\{job="vedh-api",result!="success"\}\[24h\]\)\) or vector\(0\)$/);

  const joinsPerGamePanel = panelsByTitle.get('Joins per Game Created');
  assert.match(joinsPerGamePanel.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(joinsPerGamePanel.targets[0].expr, /vedh_games_created_total\{job="vedh-api"\}\[24h\]/);

  const formatPanel = panelsByTitle.get('Games by Format Last 24h');
  assert.equal(formatPanel.type, 'bargauge');
  assert.equal(formatPanel.targets[0].legendFormat, '{{format}}');
  assert.match(formatPanel.targets[0].expr, /^sum by \(format\) \(increase\(vedh_games_created_total\{job="vedh-api"\}\[24h\]\)\)$/);

  const signupOutcomesPanel = panelsByTitle.get('Signup Outcomes Last 24h');
  assert.equal(signupOutcomesPanel.type, 'bargauge');
  assert.equal(signupOutcomesPanel.targets[0].legendFormat, '{{result}}');
  assert.match(signupOutcomesPanel.targets[0].expr, /^sum by \(result\) \(increase\(vedh_signups_total\{job="vedh-api"\}\[24h\]\)\)$/);

  const loginOutcomesPanel = panelsByTitle.get('Login Outcomes Last 24h');
  assert.equal(loginOutcomesPanel.type, 'bargauge');
  assert.equal(loginOutcomesPanel.targets[0].legendFormat, '{{result}}');
  assert.match(loginOutcomesPanel.targets[0].expr, /^sum by \(result\) \(increase\(vedh_login_attempts_total\{job="vedh-api"\}\[24h\]\)\)$/);

  const joinOutcomesPanel = panelsByTitle.get('Join Outcomes Last 24h');
  assert.equal(joinOutcomesPanel.type, 'bargauge');
  assert.equal(joinOutcomesPanel.targets[0].legendFormat, '{{result}}');
  assert.match(joinOutcomesPanel.targets[0].expr, /^sum by \(result\) \(increase\(vedh_game_join_attempts_total\{job="vedh-api"\}\[24h\]\)\)$/);

  const funnelPanel = panelsByTitle.get('Funnel Activity (6h rolling)');
  assert.equal(funnelPanel.type, 'timeseries');
  assert.equal(funnelPanel.targets.length, 4);
  assert.match(funnelPanel.targets[0].expr, /vedh_signups_total\{job="vedh-api",result="success"\}/);
  assert.match(funnelPanel.targets[1].expr, /vedh_login_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(funnelPanel.targets[2].expr, /vedh_games_created_total\{job="vedh-api"\}/);
  assert.match(funnelPanel.targets[3].expr, /vedh_game_join_attempts_total\{job="vedh-api",result="success"\}/);

  const formatActivityPanel = panelsByTitle.get('Game Format Activity (6h rolling)');
  assert.equal(formatActivityPanel.type, 'timeseries');
  assert.match(formatActivityPanel.targets[0].expr, /sum by \(format\) \(increase\(vedh_games_created_total\{job="vedh-api"\}\[6h\]\)\)/);

  const joinQualityPanel = panelsByTitle.get('Join Quality (6h rolling)');
  assert.equal(joinQualityPanel.type, 'timeseries');
  assert.equal(joinQualityPanel.targets.length, 2);
  assert.match(joinQualityPanel.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(joinQualityPanel.targets[1].expr, /vedh_game_join_attempts_total\{job="vedh-api",result!="success"\}/);

  const memoryPanel = panelsByTitle.get('Resident Memory');
  assert.match(memoryPanel.targets[0].expr, /^process_resident_memory_bytes\{job="vedh-api"\}$/);

  const cpuPanel = panelsByTitle.get('Process CPU Usage');
  assert.match(cpuPanel.targets[0].expr, /^100 \* rate\(process_cpu_seconds_total\{job="vedh-api"\}\[\$__rate_interval\]\)$/);

  const fdPanel = panelsByTitle.get('File Descriptors');
  assert.equal(fdPanel.targets.length, 2);
  assert.match(fdPanel.targets[0].expr, /process_open_fds/);
  assert.match(fdPanel.targets[1].expr, /process_max_fds/);

  const heapPanel = panelsByTitle.get('Go Heap vs Stack');
  assert.equal(heapPanel.targets.length, 2);
  assert.match(heapPanel.targets[0].expr, /go_memstats_heap_alloc_bytes/);
  assert.match(heapPanel.targets[1].expr, /go_memstats_stack_inuse_bytes/);

  assert.match(panelsByTitle.get('Public Request Decisions (5m rate)').targets[0].expr, /vedh_rate_limit_total/);
  assert.match(panelsByTitle.get('Deck Import Outcomes (5m rate)').targets[0].expr, /vedh_deck_import_total/);
  assert.match(panelsByTitle.get('Game Create Outcomes (5m rate)').targets[0].expr, /vedh_game_create_total/);
  assert.match(panelsByTitle.get('Game Join Outcomes (5m rate)').targets[0].expr, /vedh_game_join_total/);

  const activationLatency = panelsByTitle.get('Activation Latency p90');
  assert.equal(activationLatency.targets.length, 4);
  assert.match(activationLatency.targets[0].expr, /vedh_deck_import_duration_seconds_bucket/);
  assert.match(activationLatency.targets[1].expr, /vedh_game_create_duration_seconds_bucket/);
  assert.match(activationLatency.targets[2].expr, /vedh_game_join_duration_seconds_bucket/);
  assert.match(activationLatency.targets[3].expr, /vedh_board_activation_duration_seconds_bucket/);
  assert.match(panelsByTitle.get('Provider Fetch Health (5m rate)').targets[0].expr, /vedh_deck_provider_fetch_total/);
  assert.match(panelsByTitle.get('Board Readiness Mode (15m)').targets[0].expr, /vedh_board_activation_total/);
  assert.match(panelsByTitle.get('Product Event Drops (15m)').targets[0].expr, /vedh_product_events_dropped_total/);
});

test('generated dashboard JSON stays in sync with the dashboard builder', () => {
  const generated = createVedhAppDashboard({ datasourceUid: 'bfcctb05vkm4ge' });
  const jsonPath = path.join(__dirname, 'vedh-app-overview.json');
  const saved = JSON.parse(fs.readFileSync(jsonPath, 'utf8'));

  assert.deepEqual(saved, generated);
});

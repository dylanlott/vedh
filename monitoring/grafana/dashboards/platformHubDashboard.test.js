const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const { createPlatformHubDashboard } = require('./platformHubDashboard');

test('createPlatformHubDashboard builds a cross-app hub around live vedh, jank, and node metrics', () => {
  const dashboard = createPlatformHubDashboard({ datasourceUid: 'test-vm-uid' });

  assert.equal(dashboard.title, 'Platform Hub Overview');
  assert.equal(dashboard.schemaVersion >= 39, true);
  assert.equal(Array.isArray(dashboard.panels), true);
  assert.equal(dashboard.panels.length, 18);
  assert.deepEqual(dashboard.time, { from: 'now-24h', to: 'now' });

  const panelsByTitle = new Map(dashboard.panels.map((panel) => [panel.title, panel]));
  const expectedPanels = [
    'vEDH Up',
    'Jank Up',
    'Host Load (1m)',
    'Memory Available',
    'Disk Used Root',
    'Host CPU Busy',
    'vEDH Signups',
    'vEDH Games Created',
    'vEDH Join Success Rate',
    'Jank Search Queries',
    'Jank Successful Signups',
    'Jank Content Created',
    'vEDH Request Rate',
    'Jank Request Rate',
    'Jank Search Share',
    'App Request Rate Comparison',
    'App Memory Footprint',
    'Route Mix Since Deploy',
  ];

  for (const title of expectedPanels) {
    assert.ok(panelsByTitle.has(title), `expected panel ${title}`);
  }

  const vedhUp = panelsByTitle.get('vEDH Up');
  assert.match(vedhUp.targets[0].expr, /^up\{job="vedh-api"\}$/);

  const jankUp = panelsByTitle.get('Jank Up');
  assert.match(jankUp.targets[0].expr, /^up\{job="jank-app"\}$/);

  const memAvail = panelsByTitle.get('Memory Available');
  assert.equal(memAvail.fieldConfig.defaults.unit, 'percent');
  assert.match(memAvail.targets[0].expr, /^100 \* node_memory_MemAvailable_bytes\{job="node"\} \/ node_memory_MemTotal_bytes\{job="node"\}$/);

  const diskUsed = panelsByTitle.get('Disk Used Root');
  assert.equal(diskUsed.fieldConfig.defaults.unit, 'percent');
  assert.match(diskUsed.targets[0].expr, /^100 \* \(1 - node_filesystem_avail_bytes\{job="node",mountpoint="\/",fstype!="rootfs"\} \/ node_filesystem_size_bytes\{job="node",mountpoint="\/",fstype!="rootfs"\}\)$/);

  const cpuBusy = panelsByTitle.get('Host CPU Busy');
  assert.match(cpuBusy.targets[0].expr, /^100 \* \(1 - avg\(rate\(node_cpu_seconds_total\{job="node",mode="idle"\}\[5m\]\)\)\)$/);

  const vedhSignups = panelsByTitle.get('vEDH Signups');
  assert.match(vedhSignups.targets[0].expr, /^sum\(vedh_signups_total\{job="vedh-api",result="success"\}\)$/);

  const vedhGamesCreated = panelsByTitle.get('vEDH Games Created');
  assert.match(vedhGamesCreated.targets[0].expr, /^sum\(vedh_games_created_total\{job="vedh-api"\}\)$/);

  const vedhJoinRate = panelsByTitle.get('vEDH Join Success Rate');
  assert.equal(vedhJoinRate.fieldConfig.defaults.unit, 'percent');
  assert.match(vedhJoinRate.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api",result="success"\}/);
  assert.match(vedhJoinRate.targets[0].expr, /vedh_game_join_attempts_total\{job="vedh-api"\}/);

  const jankSearchQueries = panelsByTitle.get('Jank Search Queries');
  assert.match(jankSearchQueries.targets[0].expr, /^sum\(jank_search_queries_total\{job="jank-app"\}\)$/);

  const jankSignups = panelsByTitle.get('Jank Successful Signups');
  assert.match(jankSignups.targets[0].expr, /^sum\(jank_signup_attempts_total\{job="jank-app",result="success"\}\)$/);

  const jankContentCreated = panelsByTitle.get('Jank Content Created');
  assert.match(jankContentCreated.targets[0].expr, /jank_threads_created_total/);
  assert.match(jankContentCreated.targets[0].expr, /jank_posts_created_total/);
  assert.match(jankContentCreated.targets[0].expr, /jank_card_trees_created_total/);

  const vedhRate = panelsByTitle.get('vEDH Request Rate');
  assert.match(vedhRate.targets[0].expr, /^sum\(rate\(promhttp_metric_handler_requests_total\{job="vedh-api"\}\[1h\]\)\)$/);

  const jankRate = panelsByTitle.get('Jank Request Rate');
  assert.match(jankRate.targets[0].expr, /^sum\(rate\(jank_http_requests_total\{job="jank-app",route!="\/metrics"\}\[1h\]\)\)$/);

  const jankSearchShare = panelsByTitle.get('Jank Search Share');
  assert.match(jankSearchShare.targets[0].expr, /route="\/search"/);
  assert.match(jankSearchShare.targets[0].expr, /route=~"\/\|\/login\|\/search"/);

  const requestCompare = panelsByTitle.get('App Request Rate Comparison');
  assert.equal(requestCompare.targets.length, 2);
  assert.match(requestCompare.targets[0].expr, /promhttp_metric_handler_requests_total\{job="vedh-api"\}/);
  assert.match(requestCompare.targets[1].expr, /jank_http_requests_total\{job="jank-app",route!="\/metrics"\}/);

  const memoryFootprint = panelsByTitle.get('App Memory Footprint');
  assert.equal(memoryFootprint.targets.length, 2);
  assert.match(memoryFootprint.targets[0].expr, /process_resident_memory_bytes\{job="vedh-api"\}/);
  assert.match(memoryFootprint.targets[1].expr, /process_resident_memory_bytes\{job="jank-app"\}/);

  const routeMix = panelsByTitle.get('Route Mix Since Deploy');
  assert.equal(routeMix.type, 'bargauge');
  assert.match(routeMix.targets[0].expr, /sum by \(route\) \(jank_http_requests_total\{job="jank-app",route=~"\/\|\/login\|\/search"\}\)/);
});

test('generated platform hub JSON stays in sync with the dashboard builder', () => {
  const generated = createPlatformHubDashboard({ datasourceUid: 'bfcctb05vkm4ge' });
  const jsonPath = path.join(__dirname, 'platform-hub-overview.json');
  const saved = JSON.parse(fs.readFileSync(jsonPath, 'utf8'));

  assert.deepEqual(saved, generated);
});

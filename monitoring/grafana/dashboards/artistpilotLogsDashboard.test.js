const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

let createArtistPilotLogsDashboard;
try {
  ({ createArtistPilotLogsDashboard } = require('./artistpilotLogsDashboard'));
} catch {}

test('createArtistPilotLogsDashboard keeps the dedicated ArtistPilot logs view populated with panel-level lookbacks', () => {
  assert.equal(typeof createArtistPilotLogsDashboard, 'function');

  const dashboard = createArtistPilotLogsDashboard({
    logsDatasourceUid: 'test-vlogs-uid',
  });

  assert.equal(dashboard.title, 'ArtistPilot Logs');
  assert.deepEqual(dashboard.time, { from: 'now-24h', to: 'now' });
  assert.equal(Array.isArray(dashboard.panels), true);
  assert.equal(dashboard.panels.length, 3);

  const panelsByTitle = new Map(dashboard.panels.map((panel) => [panel.title, panel]));
  const apiPanel = panelsByTitle.get('API Logs (artistpilot-api)');
  const webPanel = panelsByTitle.get('Web Logs & deploy churn (artistpilot-web)');
  const eventsPanel = panelsByTitle.get('Signup & Billing Events');

  assert.ok(apiPanel, 'expected API logs panel');
  assert.ok(webPanel, 'expected web logs panel');
  assert.ok(eventsPanel, 'expected signup and billing events panel');

  assert.equal(apiPanel.type, 'logs');
  assert.equal(apiPanel.timeFrom, '24h');
  assert.deepEqual(apiPanel.datasource, {
    type: 'victoriametrics-logs-datasource',
    uid: 'test-vlogs-uid',
  });
  assert.equal(apiPanel.targets[0].queryType, 'instant');
  assert.equal(apiPanel.targets[0].expr, 'label.com.dokku.app-name:=artistpilot-api');

  assert.equal(webPanel.type, 'logs');
  assert.equal(webPanel.timeFrom, '24h');
  assert.equal(webPanel.targets[0].expr, 'label.com.dokku.app-name:=artistpilot-web');

  assert.equal(eventsPanel.type, 'logs');
  assert.equal(eventsPanel.timeFrom, '7d');
  assert.equal(
    eventsPanel.targets[0].expr,
    'label.com.dokku.app-name:=artistpilot-api (_msg:"signup verification code generated" OR _msg:"stripe webhook processed")'
  );
});

test('generated ArtistPilot Logs dashboard JSON stays in sync with the dashboard builder', () => {
  assert.equal(typeof createArtistPilotLogsDashboard, 'function');

  const generated = createArtistPilotLogsDashboard({
    logsDatasourceUid: 'dfgpn90dvyqyod',
  });
  const jsonPath = path.join(__dirname, 'artistpilot-logs.json');

  assert.equal(fs.existsSync(jsonPath), true);
  const saved = JSON.parse(fs.readFileSync(jsonPath, 'utf8'));

  assert.deepEqual(saved, generated);
});

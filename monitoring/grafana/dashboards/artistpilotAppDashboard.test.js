const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

let createArtistPilotAppDashboard;
try {
  ({ createArtistPilotAppDashboard } = require('./artistpilotAppDashboard'));
} catch {}

test('createArtistPilotAppDashboard builds an ArtistPilot dashboard around signup and subscription engagement', () => {
  assert.equal(typeof createArtistPilotAppDashboard, 'function');

  const dashboard = createArtistPilotAppDashboard({
    datasourceUid: 'test-vm-uid',
    logsDatasourceUid: 'test-vlogs-uid',
  });

  assert.equal(dashboard.title, 'ArtistPilot App Overview');
  assert.equal(dashboard.schemaVersion >= 39, true);
  assert.equal(Array.isArray(dashboard.panels), true);
  assert.equal(dashboard.panels.length, 30);
  assert.deepEqual(dashboard.time, { from: 'now-24h', to: 'now' });

  const panelsByTitle = new Map(dashboard.panels.map((panel) => [panel.title, panel]));
  const expectedPanels = [
    'Target Up',
    'Daily Active Users',
    'Weekly Active Users',
    'Monthly Active Users',
    'Workspaces',
    'User Base Since Deploy',
    'New Users Last 24h',
    'Active Subscriptions',
    'Signup → Subscription Ratio',
    'Created Subscriptions Since Deploy',
    'Cancelled Subscriptions Since Deploy',
    'Subscription Events Since Deploy',
    'Active SSE Connections',
    'Chat Requests Since Deploy',
    'Chat Success Rate',
    'Signups by Plan & Source',
    'Active Subscriptions by Plan',
    'Subscription Lifecycle Events by Plan & Event',
    'Subscription Event Mix Since Deploy',
    'Signup Rate by Plan & Source',
    'Subscription Event Rate by Plan & Event',
    'Signups vs Active Subscriptions',
    'HTTP Status Mix Since Deploy',
    'Errors by Kind',
    'Request Rate by Status',
    'Process CPU Usage',
    'Resident Memory',
    'API/Web stderr & deploy churn',
    '4xx/5xx Request Lines',
    'Signup & Billing Events',
  ];

  for (const title of expectedPanels) {
    assert.ok(panelsByTitle.has(title), `expected panel ${title}`);
  }

  const targetUpPanel = panelsByTitle.get('Target Up');
  assert.equal(targetUpPanel.datasource.uid, 'test-vm-uid');
  assert.match(targetUpPanel.targets[0].expr, /^up\{job="artistpilot-api"\}$/);
  assert.equal(targetUpPanel.fieldConfig.defaults.mappings[0].options['1'].text, 'Up');

  const activeUsersPanel = panelsByTitle.get('Monthly Active Users');
  assert.match(activeUsersPanel.targets[0].expr, /^max\(artistpilot_monthly_active_users_total\{job="artistpilot-api"\}\)$/);

  const dailyActiveUsersPanel = panelsByTitle.get('Daily Active Users');
  assert.match(dailyActiveUsersPanel.targets[0].expr, /^max\(artistpilot_daily_active_users_total\{job="artistpilot-api"\}\)$/);

  const weeklyActiveUsersPanel = panelsByTitle.get('Weekly Active Users');
  assert.match(weeklyActiveUsersPanel.targets[0].expr, /^max\(artistpilot_weekly_active_users_total\{job="artistpilot-api"\}\)$/);

  const workspacesPanel = panelsByTitle.get('Workspaces');
  assert.match(workspacesPanel.targets[0].expr, /^max\(artistpilot_workspaces_total\{job="artistpilot-api"\}\)$/);

  const signupsPanel = panelsByTitle.get('User Base Since Deploy');
  assert.match(signupsPanel.targets[0].expr, /^max\(artistpilot_users_total\{job="artistpilot-api"\}\)$/);

  const newUsers24hPanel = panelsByTitle.get('New Users Last 24h');
  assert.match(newUsers24hPanel.targets[0].expr, /^sum\(increase\(artistpilot_signups_total\{job="artistpilot-api"\}\[24h\]\)\) or vector\(0\)$/);

  const activeSubsPanel = panelsByTitle.get('Active Subscriptions');
  assert.match(activeSubsPanel.targets[0].expr, /^sum\(artistpilot_subscriptions_active\{job="artistpilot-api"\}\) or vector\(0\)$/);

  const conversionPanel = panelsByTitle.get('Signup → Subscription Ratio');
  assert.equal(conversionPanel.fieldConfig.defaults.unit, 'percent');
  assert.match(conversionPanel.targets[0].expr, /sum\(artistpilot_subscriptions_active\{job="artistpilot-api"\}\)/);
  assert.match(conversionPanel.targets[0].expr, /sum\(artistpilot_signups_total\{job="artistpilot-api"\}\)/);

  const createdSubsPanel = panelsByTitle.get('Created Subscriptions Since Deploy');
  assert.match(createdSubsPanel.targets[0].expr, /^sum\(artistpilot_subscriptions_total\{job="artistpilot-api",event="created"\}\) or vector\(0\)$/);

  const cancelledSubsPanel = panelsByTitle.get('Cancelled Subscriptions Since Deploy');
  assert.match(cancelledSubsPanel.targets[0].expr, /^sum\(artistpilot_subscriptions_total\{job="artistpilot-api",event="cancelled"\}\) or vector\(0\)$/);

  const subEventsPanel = panelsByTitle.get('Subscription Events Since Deploy');
  assert.match(subEventsPanel.targets[0].expr, /^sum\(artistpilot_subscriptions_total\{job="artistpilot-api"\}\) or vector\(0\)$/);

  const ssePanel = panelsByTitle.get('Active SSE Connections');
  assert.match(ssePanel.targets[0].expr, /^artistpilot_sse_connections_active\{job="artistpilot-api"\}$/);

  const chatRequestsPanel = panelsByTitle.get('Chat Requests Since Deploy');
  assert.match(chatRequestsPanel.targets[0].expr, /^sum\(artistpilot_chat_requests_total\{job="artistpilot-api"\}\)$/);

  const chatSuccessRatePanel = panelsByTitle.get('Chat Success Rate');
  assert.equal(chatSuccessRatePanel.fieldConfig.defaults.unit, 'percent');
  assert.match(chatSuccessRatePanel.targets[0].expr, /artistpilot_chat_requests_total\{job="artistpilot-api",status="success"\}/);
  assert.match(chatSuccessRatePanel.targets[0].expr, /artistpilot_chat_requests_total\{job="artistpilot-api"\}/);

  const signupsByPlanPanel = panelsByTitle.get('Signups by Plan & Source');
  assert.equal(signupsByPlanPanel.type, 'bargauge');
  assert.equal(signupsByPlanPanel.targets[0].legendFormat, '{{plan}} • {{source}}');
  assert.match(signupsByPlanPanel.targets[0].expr, /^sum by \(plan,source\) \(artistpilot_signups_total\{job="artistpilot-api"\}\)$/);

  const subscriptionsByPlanPanel = panelsByTitle.get('Active Subscriptions by Plan');
  assert.equal(subscriptionsByPlanPanel.type, 'bargauge');
  assert.equal(subscriptionsByPlanPanel.targets[0].legendFormat, '{{plan}}');
  assert.match(subscriptionsByPlanPanel.targets[0].expr, /^sum by \(plan\) \(artistpilot_subscriptions_active\{job="artistpilot-api"\}\)$/);

  const lifecyclePanel = panelsByTitle.get('Subscription Lifecycle Events by Plan & Event');
  assert.equal(lifecyclePanel.type, 'bargauge');
  assert.equal(lifecyclePanel.targets[0].legendFormat, '{{plan}} • {{event}}');
  assert.match(lifecyclePanel.targets[0].expr, /^sum by \(plan,event\) \(artistpilot_subscriptions_total\{job="artistpilot-api"\}\)$/);

  const lifecycleMixPanel = panelsByTitle.get('Subscription Event Mix Since Deploy');
  assert.equal(lifecycleMixPanel.type, 'bargauge');
  assert.equal(lifecycleMixPanel.targets[0].legendFormat, '{{event}}');
  assert.match(lifecycleMixPanel.targets[0].expr, /^sum by \(event\) \(artistpilot_subscriptions_total\{job="artistpilot-api"\}\)$/);

  const signupRatePanel = panelsByTitle.get('Signup Rate by Plan & Source');
  assert.equal(signupRatePanel.fieldConfig.defaults.unit, 'reqps');
  assert.match(signupRatePanel.targets[0].expr, /sum by \(plan,source\) \(rate\(artistpilot_signups_total\{job="artistpilot-api"\}\[\$__rate_interval\]\)\)/);

  const subscriptionEventRatePanel = panelsByTitle.get('Subscription Event Rate by Plan & Event');
  assert.equal(subscriptionEventRatePanel.fieldConfig.defaults.unit, 'reqps');
  assert.match(subscriptionEventRatePanel.targets[0].expr, /sum by \(plan,event\) \(rate\(artistpilot_subscriptions_total\{job="artistpilot-api"\}\[\$__rate_interval\]\)\)/);

  const signupsVsSubscriptionsPanel = panelsByTitle.get('Signups vs Active Subscriptions');
  assert.equal(signupsVsSubscriptionsPanel.targets.length, 2);
  assert.match(signupsVsSubscriptionsPanel.targets[0].expr, /^sum\(artistpilot_signups_total\{job="artistpilot-api"\}\)$/);
  assert.match(signupsVsSubscriptionsPanel.targets[1].expr, /^sum\(artistpilot_subscriptions_active\{job="artistpilot-api"\}\)$/);

  const httpStatusMixPanel = panelsByTitle.get('HTTP Status Mix Since Deploy');
  assert.equal(httpStatusMixPanel.type, 'bargauge');
  assert.equal(httpStatusMixPanel.targets[0].legendFormat, '{{status}}');
  assert.match(httpStatusMixPanel.targets[0].expr, /^sum by \(status\) \(artistpilot_http_requests_total\{job="artistpilot-api"\}\)$/);

  const errorsPanel = panelsByTitle.get('Errors by Kind');
  assert.equal(errorsPanel.type, 'bargauge');
  assert.equal(errorsPanel.targets[0].legendFormat, '{{kind}}');
  assert.match(errorsPanel.targets[0].expr, /^sum by \(kind\) \(artistpilot_errors_total\{job="artistpilot-api"\}\)$/);

  const requestRatePanel = panelsByTitle.get('Request Rate by Status');
  assert.equal(requestRatePanel.type, 'timeseries');
  assert.equal(requestRatePanel.fieldConfig.defaults.unit, 'reqps');
  assert.match(requestRatePanel.targets[0].expr, /^sum by \(status\) \(rate\(artistpilot_http_requests_total\{job="artistpilot-api"\}\[\$__rate_interval\]\)\)$/);

  const cpuPanel = panelsByTitle.get('Process CPU Usage');
  assert.equal(cpuPanel.type, 'timeseries');
  assert.equal(cpuPanel.fieldConfig.defaults.unit, 'percent');
  assert.match(cpuPanel.targets[0].expr, /^100 \* rate\(process_cpu_seconds_total\{job="artistpilot-api"\}\[\$__rate_interval\]\)$/);

  const memoryPanel = panelsByTitle.get('Resident Memory');
  assert.equal(memoryPanel.type, 'stat');
  assert.equal(memoryPanel.fieldConfig.defaults.unit, 'bytes');
  assert.match(memoryPanel.targets[0].expr, /^process_resident_memory_bytes\{job="artistpilot-api"\}$/);

  const stderrPanel = panelsByTitle.get('API/Web stderr & deploy churn');
  assert.equal(stderrPanel.type, 'logs');
  assert.deepEqual(stderrPanel.datasource, {
    type: 'victoriametrics-logs-datasource',
    uid: 'test-vlogs-uid',
  });
  assert.equal(stderrPanel.targets[0].queryType, 'instant');
  assert.equal(
    stderrPanel.targets[0].expr,
    '(label.com.dokku.app-name:=artistpilot-api OR label.com.dokku.app-name:=artistpilot-web) stream:=stderr'
  );
  assert.equal(stderrPanel.timeFrom, '24h');
  assert.equal(stderrPanel.options.enableLogDetails, true);
  assert.equal(stderrPanel.options.sortOrder, 'Descending');
  assert.equal(stderrPanel.options.wrapLogMessage, true);

  const requestFailurePanel = panelsByTitle.get('4xx/5xx Request Lines');
  assert.equal(requestFailurePanel.type, 'logs');
  assert.deepEqual(requestFailurePanel.datasource, {
    type: 'victoriametrics-logs-datasource',
    uid: 'test-vlogs-uid',
  });
  assert.equal(
    requestFailurePanel.targets[0].expr,
    'label.com.dokku.app-name:=artistpilot-api message:~"request" message:~"status.:4..|status.:5.."'
  );
  assert.equal(requestFailurePanel.timeFrom, '24h');
  assert.equal(requestFailurePanel.options.enableLogDetails, true);
  assert.equal(requestFailurePanel.options.sortOrder, 'Descending');

  const businessEventsPanel = panelsByTitle.get('Signup & Billing Events');
  assert.equal(businessEventsPanel.type, 'logs');
  assert.deepEqual(businessEventsPanel.datasource, {
    type: 'victoriametrics-logs-datasource',
    uid: 'test-vlogs-uid',
  });
  assert.equal(
    businessEventsPanel.targets[0].expr,
    'label.com.dokku.app-name:=artistpilot-api ("signup verification code generated" OR "stripe webhook processed")'
  );
  assert.equal(businessEventsPanel.timeFrom, '7d');
  assert.equal(businessEventsPanel.options.enableLogDetails, true);
  assert.equal(businessEventsPanel.options.sortOrder, 'Descending');
});

test('generated ArtistPilot dashboard JSON stays in sync with the dashboard builder', () => {
  assert.equal(typeof createArtistPilotAppDashboard, 'function');

  const generated = createArtistPilotAppDashboard({
    datasourceUid: 'bfcctb05vkm4ge',
    logsDatasourceUid: 'dfgpn90dvyqyod',
  });
  const jsonPath = path.join(__dirname, 'artistpilot-app-overview.json');

  assert.equal(fs.existsSync(jsonPath), true);
  const saved = JSON.parse(fs.readFileSync(jsonPath, 'utf8'));

  assert.deepEqual(saved, generated);
});

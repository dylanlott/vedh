function statPanel({
  id,
  title,
  expr,
  gridPos,
  datasourceUid,
  unit = 'none',
  colorMode = 'value',
  thresholds = {
    mode: 'absolute',
    steps: [
      { color: 'red', value: null },
      { color: 'green', value: 1 },
    ],
  },
  mappings = [],
  description,
}) {
  return {
    id,
    type: 'stat',
    title,
    ...(description ? { description } : {}),
    datasource: { type: 'prometheus', uid: datasourceUid },
    gridPos,
    targets: [{ refId: 'A', expr }],
    options: {
      colorMode,
      graphMode: 'none',
      justifyMode: 'center',
      orientation: 'auto',
      reduceOptions: {
        calcs: ['lastNotNull'],
        fields: '',
        values: false,
      },
      textMode: 'auto',
    },
    fieldConfig: {
      defaults: {
        unit,
        thresholds,
        mappings,
      },
      overrides: [],
    },
  };
}

function timeSeriesPanel({
  id,
  title,
  targets,
  expr,
  gridPos,
  datasourceUid,
  unit = 'short',
  legendMode = 'list',
  min,
  thresholds,
  description,
}) {
  const defaults = { unit };
  if (min !== undefined) defaults.min = min;
  if (thresholds !== undefined) defaults.thresholds = thresholds;

  return {
    id,
    type: 'timeseries',
    title,
    ...(description ? { description } : {}),
    datasource: { type: 'prometheus', uid: datasourceUid },
    gridPos,
    targets: targets || [{ refId: 'A', expr, legendFormat: '__auto' }],
    options: {
      legend: {
        displayMode: legendMode,
        placement: 'bottom',
        showLegend: true,
      },
      tooltip: {
        mode: 'multi',
        sort: 'none',
      },
    },
    fieldConfig: {
      defaults,
      overrides: [],
    },
  };
}

function barGaugePanel({
  id,
  title,
  expr,
  gridPos,
  datasourceUid,
  unit = 'short',
  description,
  legendFormat = '{{format}}',
}) {
  return {
    id,
    type: 'bargauge',
    title,
    ...(description ? { description } : {}),
    datasource: { type: 'prometheus', uid: datasourceUid },
    gridPos,
    targets: [{ refId: 'A', expr, legendFormat }],
    options: {
      displayMode: 'basic',
      orientation: 'horizontal',
      reduceOptions: {
        calcs: ['lastNotNull'],
        fields: '',
        values: false,
      },
      showUnfilled: true,
    },
    fieldConfig: {
      defaults: { unit },
      overrides: [],
    },
  };
}

function createVedhAppDashboard({ datasourceUid }) {
  if (!datasourceUid) throw new Error('datasourceUid is required');

  return {
    annotations: { list: [] },
    editable: true,
    fiscalYearStartMonth: 0,
    graphTooltip: 0,
    id: null,
    links: [],
    liveNow: false,
    panels: [
      statPanel({
        id: 1,
        title: 'Target Up',
        description: 'Whether the vedh-api target is currently being scraped.',
        expr: 'up{job="vedh-api"}',
        gridPos: { h: 4, w: 4, x: 0, y: 0 },
        datasourceUid,
        mappings: [
          {
            type: 'value',
            options: {
              '0': { text: 'Down', color: 'red' },
              '1': { text: 'Up', color: 'green' },
            },
          },
        ],
      }),
      statPanel({
        id: 2,
        title: 'New Users Last 24h',
        description: 'Recent successful signups into vEDH.',
        expr: 'sum(increase(vedh_signups_total{job="vedh-api",result="success"}[24h])) or vector(0)',
        gridPos: { h: 4, w: 4, x: 4, y: 0 },
        datasourceUid,
      }),
      statPanel({
        id: 3,
        title: 'Daily Active Users',
        description: 'Distinct users with persisted authenticated activity in the last 24 hours.',
        expr: 'max(vedh_daily_active_users_total{job="vedh-api"})',
        gridPos: { h: 4, w: 4, x: 8, y: 0 },
        datasourceUid,
      }),
      statPanel({
        id: 4,
        title: 'Games Created Last 24h',
        description: 'Recent game creation volume across all tracked formats.',
        expr: 'sum(increase(vedh_games_created_total{job="vedh-api"}[24h])) or vector(0)',
        gridPos: { h: 4, w: 4, x: 12, y: 0 },
        datasourceUid,
      }),
      statPanel({
        id: 5,
        title: 'Successful Joins Last 24h',
        description: 'Recent successful join attempts into existing games.',
        expr: 'sum(increase(vedh_game_join_attempts_total{job="vedh-api",result="success"}[24h])) or vector(0)',
        gridPos: { h: 4, w: 4, x: 16, y: 0 },
        datasourceUid,
      }),
      statPanel({
        id: 6,
        title: 'Join Success Rate',
        description: 'How often tracked join attempts succeed over the last 24 hours.',
        expr: '100 * sum(increase(vedh_game_join_attempts_total{job="vedh-api",result="success"}[24h])) / clamp_min(sum(increase(vedh_game_join_attempts_total{job="vedh-api"}[24h])), 1)',
        gridPos: { h: 4, w: 4, x: 20, y: 0 },
        datasourceUid,
        unit: 'percent',
      }),
      statPanel({
        id: 7,
        title: 'Signup Success Rate',
        description: 'Share of recent tracked signup attempts that succeeded.',
        expr: '100 * sum(increase(vedh_signups_total{job="vedh-api",result="success"}[24h])) / clamp_min(sum(increase(vedh_signups_total{job="vedh-api"}[24h])), 1)',
        gridPos: { h: 4, w: 4, x: 0, y: 4 },
        datasourceUid,
        unit: 'percent',
      }),
      statPanel({
        id: 8,
        title: 'Login Success Rate',
        description: 'Share of recent tracked logins that succeeded.',
        expr: '100 * sum(increase(vedh_login_attempts_total{job="vedh-api",result="success"}[24h])) / clamp_min(sum(increase(vedh_login_attempts_total{job="vedh-api"}[24h])), 1)',
        gridPos: { h: 4, w: 4, x: 4, y: 4 },
        datasourceUid,
        unit: 'percent',
      }),
      statPanel({
        id: 9,
        title: 'Join Failures Last 24h',
        description: 'Recent join attempts that landed anywhere other than success.',
        expr: 'sum(increase(vedh_game_join_attempts_total{job="vedh-api",result!="success"}[24h])) or vector(0)',
        gridPos: { h: 4, w: 4, x: 8, y: 4 },
        datasourceUid,
      }),
      statPanel({
        id: 10,
        title: 'Joins per Game Created',
        description: 'A simple engagement proxy: successful joins divided by new games created over the last 24 hours.',
        expr: '(sum(increase(vedh_game_join_attempts_total{job="vedh-api",result="success"}[24h])) / clamp_min(sum(increase(vedh_games_created_total{job="vedh-api"}[24h])), 1)) or vector(0)',
        gridPos: { h: 4, w: 4, x: 12, y: 4 },
        datasourceUid,
      }),
      barGaugePanel({
        id: 11,
        title: 'Games by Format Last 24h',
        description: 'Recent games created, split by tracked format.',
        expr: 'sum by (format) (increase(vedh_games_created_total{job="vedh-api"}[24h]))',
        gridPos: { h: 4, w: 8, x: 16, y: 4 },
        datasourceUid,
        legendFormat: '{{format}}',
      }),
      barGaugePanel({
        id: 12,
        title: 'Signup Outcomes Last 24h',
        description: 'Recent signup outcomes split by result label.',
        expr: 'sum by (result) (increase(vedh_signups_total{job="vedh-api"}[24h]))',
        gridPos: { h: 8, w: 8, x: 0, y: 8 },
        datasourceUid,
        legendFormat: '{{result}}',
      }),
      barGaugePanel({
        id: 13,
        title: 'Login Outcomes Last 24h',
        description: 'Recent login outcomes split by result label.',
        expr: 'sum by (result) (increase(vedh_login_attempts_total{job="vedh-api"}[24h]))',
        gridPos: { h: 8, w: 8, x: 8, y: 8 },
        datasourceUid,
        legendFormat: '{{result}}',
      }),
      barGaugePanel({
        id: 14,
        title: 'Join Outcomes Last 24h',
        description: 'Recent join outcomes split by result label.',
        expr: 'sum by (result) (increase(vedh_game_join_attempts_total{job="vedh-api"}[24h]))',
        gridPos: { h: 8, w: 8, x: 16, y: 8 },
        datasourceUid,
        legendFormat: '{{result}}',
      }),
      timeSeriesPanel({
        id: 15,
        title: 'Funnel Activity (6h rolling)',
        description: 'Recent account-to-play flow in one founder-readable view.',
        targets: [
          { refId: 'A', expr: 'sum(increase(vedh_signups_total{job="vedh-api",result="success"}[6h]))', legendFormat: 'signups' },
          { refId: 'B', expr: 'sum(increase(vedh_login_attempts_total{job="vedh-api",result="success"}[6h]))', legendFormat: 'logins' },
          { refId: 'C', expr: 'sum(increase(vedh_games_created_total{job="vedh-api"}[6h]))', legendFormat: 'games created' },
          { refId: 'D', expr: 'sum(increase(vedh_game_join_attempts_total{job="vedh-api",result="success"}[6h]))', legendFormat: 'successful joins' },
        ],
        gridPos: { h: 8, w: 8, x: 0, y: 16 },
        datasourceUid,
      }),
      timeSeriesPanel({
        id: 16,
        title: 'Game Format Activity (6h rolling)',
        description: 'Rolling format mix for newly created games.',
        expr: 'sum by (format) (increase(vedh_games_created_total{job="vedh-api"}[6h]))',
        gridPos: { h: 8, w: 8, x: 8, y: 16 },
        datasourceUid,
      }),
      timeSeriesPanel({
        id: 17,
        title: 'Join Quality (6h rolling)',
        description: 'Recent successful joins versus failure pressure.',
        targets: [
          { refId: 'A', expr: 'sum(increase(vedh_game_join_attempts_total{job="vedh-api",result="success"}[6h]))', legendFormat: 'successful joins' },
          { refId: 'B', expr: 'sum(increase(vedh_game_join_attempts_total{job="vedh-api",result!="success"}[6h]))', legendFormat: 'failed joins' },
        ],
        gridPos: { h: 8, w: 8, x: 16, y: 16 },
        datasourceUid,
      }),
      statPanel({
        id: 18,
        title: 'Resident Memory',
        expr: 'process_resident_memory_bytes{job="vedh-api"}',
        gridPos: { h: 8, w: 6, x: 0, y: 24 },
        datasourceUid,
        unit: 'bytes',
      }),
      timeSeriesPanel({
        id: 19,
        title: 'Process CPU Usage',
        expr: '100 * rate(process_cpu_seconds_total{job="vedh-api"}[$__rate_interval])',
        gridPos: { h: 8, w: 6, x: 6, y: 24 },
        datasourceUid,
        unit: 'percent',
        min: 0,
      }),
      timeSeriesPanel({
        id: 20,
        title: 'File Descriptors',
        targets: [
          { refId: 'A', expr: 'process_open_fds{job="vedh-api"}', legendFormat: 'open' },
          { refId: 'B', expr: 'process_max_fds{job="vedh-api"}', legendFormat: 'max' },
        ],
        gridPos: { h: 8, w: 6, x: 12, y: 24 },
        datasourceUid,
      }),
      timeSeriesPanel({
        id: 21,
        title: 'Go Heap vs Stack',
        targets: [
          { refId: 'A', expr: 'go_memstats_heap_alloc_bytes{job="vedh-api"}', legendFormat: 'heap alloc' },
          { refId: 'B', expr: 'go_memstats_stack_inuse_bytes{job="vedh-api"}', legendFormat: 'stack in use' },
        ],
        gridPos: { h: 8, w: 6, x: 18, y: 24 },
        datasourceUid,
        unit: 'bytes',
      }),
      statPanel({
        id: 22,
        title: 'User Base Since Deploy',
        description: 'Current registered-user count from persisted vEDH auth state.',
        expr: 'max(vedh_users_total{job="vedh-api"})',
        gridPos: { h: 8, w: 8, x: 0, y: 32 },
        datasourceUid,
      }),
      statPanel({
        id: 23,
        title: 'Weekly Active Users',
        description: 'Distinct users with persisted authenticated activity in the last 7 days.',
        expr: 'max(vedh_weekly_active_users_total{job="vedh-api"})',
        gridPos: { h: 8, w: 8, x: 8, y: 32 },
        datasourceUid,
      }),
      statPanel({
        id: 24,
        title: 'Monthly Active Users',
        description: 'Distinct users with persisted authenticated activity in the last 30 days.',
        expr: 'max(vedh_monthly_active_users_total{job="vedh-api"})',
        gridPos: { h: 8, w: 8, x: 16, y: 32 },
        datasourceUid,
      }),
    ],
    refresh: '30s',
    schemaVersion: 41,
    tags: ['vedh', 'app', 'engagement', 'business-metrics'],
    templating: { list: [] },
    time: { from: 'now-24h', to: 'now' },
    timepicker: {},
    timezone: 'browser',
    title: 'vEDH App Overview',
    uid: 'vedh-app-overview',
    version: 2,
    weekStart: '',
  };
}

module.exports = {
  createVedhAppDashboard,
};

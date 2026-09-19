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
}) {
  return {
    id,
    type: 'bargauge',
    title,
    ...(description ? { description } : {}),
    datasource: { type: 'prometheus', uid: datasourceUid },
    gridPos,
    targets: [{ refId: 'A', expr, legendFormat: '{{route}}' }],
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

function createPlatformHubDashboard({ datasourceUid }) {
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
        title: 'vEDH Up',
        expr: 'up{job="vedh-api"}',
        gridPos: { h: 4, w: 4, x: 0, y: 0 },
        datasourceUid,
        description: 'Whether vedh-api is being scraped right now.',
        mappings: [{ type: 'value', options: { '0': { text: 'Down', color: 'red' }, '1': { text: 'Up', color: 'green' } } }],
      }),
      statPanel({
        id: 2,
        title: 'Jank Up',
        expr: 'up{job="jank-app"}',
        gridPos: { h: 4, w: 4, x: 4, y: 0 },
        datasourceUid,
        description: 'Whether jank-app is being scraped right now.',
        mappings: [{ type: 'value', options: { '0': { text: 'Down', color: 'red' }, '1': { text: 'Up', color: 'green' } } }],
      }),
      statPanel({
        id: 3,
        title: 'Host Load (1m)',
        expr: 'node_load1{job="node"}',
        gridPos: { h: 4, w: 4, x: 8, y: 0 },
        datasourceUid,
        thresholds: { mode: 'absolute', steps: [
          { color: 'green', value: null },
          { color: 'yellow', value: 2 },
          { color: 'red', value: 4 },
        ]},
      }),
      statPanel({
        id: 4,
        title: 'Memory Available',
        expr: '100 * node_memory_MemAvailable_bytes{job="node"} / node_memory_MemTotal_bytes{job="node"}',
        gridPos: { h: 4, w: 4, x: 12, y: 0 },
        datasourceUid,
        unit: 'percent',
        thresholds: { mode: 'absolute', steps: [
          { color: 'red', value: null },
          { color: 'yellow', value: 15 },
          { color: 'green', value: 40 },
        ]},
      }),
      statPanel({
        id: 5,
        title: 'Disk Used Root',
        expr: '100 * (1 - node_filesystem_avail_bytes{job="node",mountpoint="/",fstype!="rootfs"} / node_filesystem_size_bytes{job="node",mountpoint="/",fstype!="rootfs"})',
        gridPos: { h: 4, w: 4, x: 16, y: 0 },
        datasourceUid,
        unit: 'percent',
        thresholds: { mode: 'absolute', steps: [
          { color: 'green', value: null },
          { color: 'yellow', value: 70 },
          { color: 'red', value: 85 },
        ]},
      }),
      statPanel({
        id: 6,
        title: 'Host CPU Busy',
        expr: '100 * (1 - avg(rate(node_cpu_seconds_total{job="node",mode="idle"}[5m])))',
        gridPos: { h: 4, w: 4, x: 20, y: 0 },
        datasourceUid,
        unit: 'percent',
      }),
      statPanel({
        id: 7,
        title: 'vEDH Signups',
        expr: 'sum(vedh_signups_total{job="vedh-api",result="success"})',
        gridPos: { h: 4, w: 4, x: 0, y: 4 },
        datasourceUid,
        description: 'Total successful vEDH signups since deploy.',
      }),
      statPanel({
        id: 8,
        title: 'vEDH Games Created',
        expr: 'sum(vedh_games_created_total{job="vedh-api"})',
        gridPos: { h: 4, w: 4, x: 4, y: 4 },
        datasourceUid,
        description: 'Total vEDH tables created since deploy.',
      }),
      statPanel({
        id: 9,
        title: 'vEDH Join Success Rate',
        expr: '100 * sum(vedh_game_join_attempts_total{job="vedh-api",result="success"}) / clamp_min(sum(vedh_game_join_attempts_total{job="vedh-api"}), 1)',
        gridPos: { h: 4, w: 4, x: 8, y: 4 },
        datasourceUid,
        unit: 'percent',
        description: 'Share of tracked vEDH join attempts that are succeeding.',
      }),
      statPanel({
        id: 10,
        title: 'Jank Search Queries',
        expr: 'sum(jank_search_queries_total{job="jank-app"})',
        gridPos: { h: 4, w: 4, x: 12, y: 4 },
        datasourceUid,
        description: 'Total Jank search queries since deploy.',
      }),
      statPanel({
        id: 11,
        title: 'Jank Successful Signups',
        expr: 'sum(jank_signup_attempts_total{job="jank-app",result="success"})',
        gridPos: { h: 4, w: 4, x: 16, y: 4 },
        datasourceUid,
        description: 'Total successful Jank signups since deploy.',
      }),
      statPanel({
        id: 12,
        title: 'Jank Content Created',
        expr: 'sum(jank_threads_created_total{job="jank-app"}) + sum(jank_posts_created_total{job="jank-app"}) + sum(jank_card_trees_created_total{job="jank-app"})',
        gridPos: { h: 4, w: 4, x: 20, y: 4 },
        datasourceUid,
        description: 'Combined Jank thread, post, and card-tree creation count since deploy.',
      }),
      statPanel({
        id: 13,
        title: 'vEDH Request Rate',
        expr: 'sum(rate(promhttp_metric_handler_requests_total{job="vedh-api"}[1h]))',
        gridPos: { h: 4, w: 4, x: 0, y: 8 },
        datasourceUid,
        unit: 'reqps',
        description: '1h-smoothed request rate for the live vedh metrics handler.',
      }),
      statPanel({
        id: 14,
        title: 'Jank Request Rate',
        expr: 'sum(rate(jank_http_requests_total{job="jank-app",route!="/metrics"}[1h]))',
        gridPos: { h: 4, w: 4, x: 4, y: 8 },
        datasourceUid,
        unit: 'reqps',
        description: '1h-smoothed request rate for the public Jank routes.',
      }),
      statPanel({
        id: 15,
        title: 'Jank Search Share',
        expr: '100 * sum(jank_http_requests_total{job="jank-app",route="/search"}) / clamp_min(sum(jank_http_requests_total{job="jank-app",route=~"/|/login|/search"}), 1)',
        gridPos: { h: 4, w: 4, x: 8, y: 8 },
        datasourceUid,
        unit: 'percent',
        description: 'Share of tracked forum entry traffic landing on search.',
      }),
      timeSeriesPanel({
        id: 16,
        title: 'App Request Rate Comparison',
        targets: [
          { refId: 'A', expr: 'sum(rate(promhttp_metric_handler_requests_total{job="vedh-api"}[1h]))', legendFormat: 'vedh-api' },
          { refId: 'B', expr: 'sum(rate(jank_http_requests_total{job="jank-app",route!="/metrics"}[1h]))', legendFormat: 'jank-app' },
        ],
        gridPos: { h: 8, w: 12, x: 12, y: 8 },
        datasourceUid,
        unit: 'reqps',
      }),
      timeSeriesPanel({
        id: 17,
        title: 'App Memory Footprint',
        targets: [
          { refId: 'A', expr: 'process_resident_memory_bytes{job="vedh-api"}', legendFormat: 'vedh-api rss' },
          { refId: 'B', expr: 'process_resident_memory_bytes{job="jank-app"}', legendFormat: 'jank-app rss' },
        ],
        gridPos: { h: 8, w: 12, x: 0, y: 12 },
        datasourceUid,
        unit: 'bytes',
      }),
      barGaugePanel({
        id: 18,
        title: 'Route Mix Since Deploy',
        expr: 'sum by (route) (jank_http_requests_total{job="jank-app",route=~"/|/login|/search"})',
        gridPos: { h: 8, w: 12, x: 12, y: 12 },
        datasourceUid,
        description: 'Cumulative hit distribution across the currently instrumented Jank entry routes.',
      }),
    ],
    refresh: '30s',
    schemaVersion: 41,
    tags: ['platform', 'hub', 'metrics-first'],
    templating: { list: [] },
    time: { from: 'now-24h', to: 'now' },
    timepicker: {},
    timezone: 'browser',
    title: 'Platform Hub Overview',
    uid: 'platform-hub-overview',
    version: 1,
    weekStart: '',
  };
}

module.exports = {
  createPlatformHubDashboard,
};

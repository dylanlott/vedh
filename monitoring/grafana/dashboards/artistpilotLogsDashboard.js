function logsPanel({ id, title, expr, gridPos, logsDatasourceUid, timeFrom, description }) {
  return {
    id,
    type: 'logs',
    title,
    ...(description ? { description } : {}),
    ...(timeFrom ? { timeFrom } : {}),
    datasource: { type: 'victoriametrics-logs-datasource', uid: logsDatasourceUid },
    gridPos,
    targets: [
      {
        refId: 'A',
        expr,
        queryType: 'instant',
        datasource: { type: 'victoriametrics-logs-datasource', uid: logsDatasourceUid },
      },
    ],
    options: {
      dedupStrategy: 'none',
      enableLogDetails: true,
      prettifyLogMessage: false,
      showCommonLabels: false,
      showLabels: false,
      showTime: true,
      sortOrder: 'Descending',
      wrapLogMessage: true,
    },
  };
}

function createArtistPilotLogsDashboard({ logsDatasourceUid }) {
  if (!logsDatasourceUid) throw new Error('logsDatasourceUid is required');

  return {
    id: 4,
    uid: '6f2a9d46-faa0-4307-92ed-b6bf521349d6',
    title: 'ArtistPilot Logs',
    tags: ['logs', 'artistpilot'],
    timezone: 'utc',
    schemaVersion: 38,
    version: 3,
    refresh: '30s',
    time: { from: 'now-24h', to: 'now' },
    panels: [
      logsPanel({
        id: 1,
        title: 'API Logs (artistpilot-api)',
        description: 'Recent ArtistPilot API request, app, and event logs using a panel-level 24h lookback so the view stays populated during quiet hours.',
        expr: 'label.com.dokku.app-name:=artistpilot-api',
        gridPos: { h: 14, w: 24, x: 0, y: 0 },
        logsDatasourceUid,
        timeFrom: '24h',
      }),
      logsPanel({
        id: 2,
        title: 'Web Logs & deploy churn (artistpilot-web)',
        description: 'Recent ArtistPilot web/runtime logs, including deploy/startup churn, with a 24h lookback.',
        expr: 'label.com.dokku.app-name:=artistpilot-web',
        gridPos: { h: 10, w: 24, x: 0, y: 14 },
        logsDatasourceUid,
        timeFrom: '24h',
      }),
      logsPanel({
        id: 3,
        title: 'Signup & Billing Events',
        description: 'Business-flow events are lower volume, so this panel keeps a 7d lookback for verification and Stripe/webhook checks.',
        expr: 'label.com.dokku.app-name:=artistpilot-api (_msg:"signup verification code generated" OR _msg:"stripe webhook processed")',
        gridPos: { h: 10, w: 24, x: 0, y: 24 },
        logsDatasourceUid,
        timeFrom: '7d',
      }),
    ],
  };
}

module.exports = {
  createArtistPilotLogsDashboard,
};

const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

test('dashboard business metrics update doc covers vedh, jank, and platform hub instrumentation plans', () => {
  const docPath = path.join(__dirname, 'dashboard-business-metrics-update.md');
  const doc = fs.readFileSync(docPath, 'utf8');

  const requiredSnippets = [
    '# Dashboard Business Metrics Update',
    '## vEDH App Overview',
    '## Jank App Overview',
    '## Platform Hub Overview',
    '### Proposed metrics contract',
    '### Dashboard update after instrumentation',
    'vedh_workspace_context_requests_total',
    'vedh_signups_total',
    'vedh_checkout_sessions_total',
    'jank_page_views_total',
    'jank_search_queries_total',
    'jank_login_attempts_total',
    'platform_business_value_score',
  ];

  for (const snippet of requiredSnippets) {
    assert.match(doc, new RegExp(snippet.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));
  }
});

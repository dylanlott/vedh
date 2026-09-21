import { expect, test, type Page } from '@playwright/test';
import { CREATOR_DECKLIST, e2eSession, JOINER_DECKLIST } from './helpers/testData';

async function importDeckAndSelectCommander(page: Page, decklist: string) {
  const panel = page.getByTestId('deck-import-panel');
  await panel.getByLabel('Decklist').fill(decklist);
  await panel.getByTestId('deck-preview-submit').click();
  await expect(panel.getByTestId('paste-import-success')).toBeVisible();
  await panel.getByTestId('continue-to-commanders').click();
  const commander = page.getByTestId('commander-candidate').first();
  await expect(commander).toBeVisible();
  await commander.click();
}

test('logged-out host and invitee reach one board, then host claim survives refresh', async ({ browser, baseURL }, testInfo) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');
  const suffix = `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
  const hostName = `Guest Host ${suffix}`;
  const inviteeName = `Guest Invitee ${suffix}`;
  const claimedUsername = `claim-${suffix}`;

  const hostContext = await browser.newContext();
  const inviteeContext = await browser.newContext();
  const hostSession = e2eSession('guest-host');
  const inviteeSession = e2eSession('guest-invitee');
  await hostContext.addInitScript((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), hostSession);
  await inviteeContext.addInitScript((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), inviteeSession);
  try {
    const hostPage = await hostContext.newPage();
    await hostPage.goto('/play');
    await importDeckAndSelectCommander(hostPage, CREATOR_DECKLIST);
    await hostPage.getByTestId('display-name').fill(hostName);
    await hostPage.getByTestId('start-table').click();
    await expect(hostPage).toHaveURL(/\/games\/[^/]+$/);
    const gameID = hostPage.url().split('/games/')[1]?.split('?')[0];
    expect(gameID).toBeTruthy();
    await expect(hostPage.getByRole('heading', { name: hostName })).toBeVisible();

    const inviteePage = await inviteeContext.newPage();
    await inviteePage.goto(`/join/${gameID}`);
    await expect(inviteePage.getByTestId('invite-summary')).toContainText(hostName);
    await importDeckAndSelectCommander(inviteePage, JOINER_DECKLIST);
    await inviteePage.getByTestId('join-display-name').fill(inviteeName);
    await inviteePage.getByTestId('join-table').click();
    await expect(inviteePage).toHaveURL(new RegExp(`/games/${gameID}$`));

    await expect(inviteePage.getByRole('heading', { name: hostName })).toBeVisible();
    await expect(inviteePage.getByRole('heading', { name: inviteeName })).toBeVisible();
    await expect(hostPage.getByRole('heading', { name: inviteeName })).toBeVisible();

    const claim = hostPage.locator('.claim-account');
    await expect(claim.getByRole('heading', { name: 'Save your games' })).toBeVisible();
    await claim.getByLabel('Username').fill(claimedUsername);
    await claim.getByLabel('Password').fill(`Pass!${suffix}`);
    await claim.getByRole('button', { name: 'Save my games' }).click();
    await expect(hostPage.getByRole('status')).toContainText('Games saved');
    await testInfo.attach('claimed-board', {
      body: await hostPage.screenshot(),
      contentType: 'image/png',
    });

    await hostPage.evaluate((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), e2eSession('claim-refresh'));
    await hostPage.reload();
    await expect(hostPage).toHaveURL(new RegExp(`/games/${gameID}$`));
    await expect(hostPage.getByRole('heading', { name: hostName })).toBeVisible();
    await expect(hostPage.getByRole('heading', { name: inviteeName })).toBeVisible();
    await expect(hostPage.getByRole('heading', { name: 'Save your games' })).toHaveCount(0);
  } finally {
    await hostContext.close();
    await inviteeContext.close();
  }
});

test('provider failure offers paste fallback without clearing the pasted deck', async ({ page }) => {
  await page.addInitScript((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), e2eSession('provider-fallback'));
  await page.route('**/graphql', async (route) => {
    const request = route.request();
    const payload = request.method() === 'POST' ? request.postDataJSON() as {
      operationName?: string;
      variables?: { input?: { sourceURL?: string } };
    } : null;
    if (payload?.operationName === 'PreviewDeck' && payload.variables?.input?.sourceURL) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: null,
          errors: [{
            message: 'simulated provider transport detail',
            extensions: { code: 'provider_unavailable' },
          }],
        }),
      });
      return;
    }
    await route.continue();
  });

  await page.goto('/play');
  const panel = page.getByTestId('deck-import-panel');
  await panel.getByLabel('Decklist').fill(CREATOR_DECKLIST);
  await panel.getByTestId('source-url-tab').click();
  await panel.getByTestId('source-url').fill('https://archidekt.com/decks/fixture');
  await panel.getByTestId('deck-preview-submit').click();

  const error = panel.getByTestId('activation-error');
  await expect(error).toContainText('Paste instead');
  await expect(error).not.toContainText('simulated provider transport detail');
  await error.getByRole('button', { name: 'Paste instead' }).click();
  await expect(panel.getByTestId('paste-tab')).toHaveAttribute('aria-selected', 'true');
  await expect(panel.getByLabel('Decklist')).toHaveValue(CREATOR_DECKLIST);

  await panel.getByTestId('deck-preview-submit').click();
  await expect(panel.getByTestId('paste-import-success')).toBeVisible();
});

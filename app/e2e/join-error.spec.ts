import { expect, test, type Page } from '@playwright/test';

async function seedAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    window.localStorage.setItem('edhgo/auth', JSON.stringify({
      ID: 'user-e2e',
      Username: 'e2e-user',
      Token: 'token-e2e',
    }));
  });
}

test('join page keeps continue disabled for invalid invite input', async ({ page, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  await seedAuthenticatedSession(page);
  await page.goto('/join');

  await page.getByLabel('Paste invite link or game ID').fill('not a game link');

  await expect(page.getByText('Couldn’t detect a game ID from that input.')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Continue' })).toBeDisabled();
});

test('joining a missing game redirects to the game not found screen', async ({ page, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  await seedAuthenticatedSession(page);
  await page.addInitScript(() => {
    const originalFetch = window.fetch.bind(window);

    window.fetch = async (input, init) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
      const body = typeof init?.body === 'string' ? init.body : '';

      if (url.includes('/graphql') && body.includes('GetGame')) {
        return new Response(JSON.stringify({
          data: {
            getGame: {
              ID: 'non-existent-game',
              CreatedAt: '2026-07-24T00:00:00.000Z',
              Status: 'OPEN',
              Result: null,
              WinnerIDs: [],
              WinCondition: null,
              PendingWinClaim: null,
              Players: [],
              Stack: [],
              Turn: null,
              Rules: [],
            },
          },
        }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }

      if (url.includes('/graphql') && body.includes('JoinGame')) {
        return new Response(JSON.stringify({
          errors: [{ message: 'game not found' }],
        }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }

      return originalFetch(input, init);
    };
  });

  await page.goto('/join/non-existent-game');
  await page.getByRole('button', { name: 'Join game' }).click();

  await expect(page).toHaveURL(/\/games\/404$/);
  await expect(page.getByRole('heading', { name: 'Game not found' })).toBeVisible();
  await expect(page.getByText('The game you were looking for doesn’t exist or may have been closed.')).toBeVisible();
});

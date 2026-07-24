import { expect, test } from '@playwright/test';

const gameID = 'game-score-e2e';
const authProfile = {
  ID: 'user-self',
  Username: 'scorekeeper',
  Token: 'test-token',
};

const mockedGame = {
  ID: gameID,
  CreatedAt: '2026-07-24T00:00:00Z',
  Status: 'ACTIVE',
  Result: null,
  WinnerIDs: [],
  WinCondition: null,
  PendingWinClaim: null,
  Stack: [],
  Turn: { Player: 'scorekeeper', Phase: 'MAIN', Number: 4, Priority: 'scorekeeper' },
  Rules: [{ Name: 'format', Value: 'EDH' }],
  Players: [
    {
      ID: 'user-self',
      Username: 'scorekeeper',
      Boardstate: {
        Life: 34,
        Commander: [{ ID: 'cmd-1', Name: 'Atraxa, Praetors\' Voice', Tapped: false }],
        Battlefield: [{ ID: 'bf-1', Name: 'Sol Ring', Types: 'Artifact', Tapped: false }],
        Hand: [{ ID: 'hand-1', Name: 'Swords to Plowshares', Types: 'Instant', Tapped: false }],
        Graveyard: [{ ID: 'gy-1', Name: 'Cultivate', Tapped: false }],
        Exiled: [],
        Revealed: [],
        Library: [],
        Controlled: [],
      },
    },
    {
      ID: 'user-opp',
      Username: 'archenemy',
      Boardstate: {
        Life: 18,
        Commander: [{ ID: 'cmd-2', Name: 'Krenko, Mob Boss', Tapped: false }],
        Battlefield: [{ ID: 'bf-2', Name: 'Goblin Token', Types: 'Creature', Tapped: false }],
        Hand: [{ ID: 'hand-2', Name: 'Mountain', Types: 'Land', Tapped: false }, { ID: 'hand-3', Name: 'Chaos Warp', Types: 'Instant', Tapped: false }],
        Graveyard: [],
        Exiled: [],
        Revealed: [],
        Library: [],
        Controlled: [],
      },
    },
  ],
};

test('players can open the score view for a vEDH game', async ({ page, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  await page.addInitScript((profile) => {
    window.localStorage.setItem('edhgo/auth', JSON.stringify(profile));
  }, authProfile);

  await page.route('**/graphql', async route => {
    const request = route.request();
    if (request.method() !== 'POST') {
      await route.fallback();
      return;
    }

    const body = request.postDataJSON() as { operationName?: string; variables?: { gameID?: string } };

    if (body.operationName === 'GetGame' && body.variables?.gameID === gameID) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { getGame: mockedGame } }),
      });
      return;
    }

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: {} }),
    });
  });

  await page.goto(`/games/${gameID}/score`);

  await expect(page).toHaveURL(new RegExp(`/games/${gameID}/score$`));
  await expect(page.getByRole('heading', { name: 'Scoreboard' })).toBeVisible();
  await expect(page.getByText('Track commander damage and life totals across the table.')).toBeVisible();

  const selfCard = page.locator('article', { has: page.getByRole('heading', { name: 'scorekeeper' }) });
  await expect(selfCard.getByRole('heading', { name: 'scorekeeper' })).toBeVisible();
  await expect(selfCard.getByText('34 life')).toBeVisible();
  await expect(selfCard.getByText('Battlefield 1')).toBeVisible();
  await expect(selfCard.getByText('Hand 1')).toBeVisible();
  await expect(selfCard.getByText('GY 1')).toBeVisible();

  const opponentCard = page.locator('article', { has: page.getByRole('heading', { name: 'archenemy' }) });
  await expect(opponentCard.getByRole('heading', { name: 'archenemy' })).toBeVisible();
  await expect(opponentCard.getByText('18 life')).toBeVisible();
  await expect(opponentCard.getByText('Battlefield 1')).toBeVisible();
  await expect(opponentCard.getByText('Hand 2')).toBeVisible();
  await expect(opponentCard.getByText('GY 0')).toBeVisible();
  await expect(selfCard.getByRole('heading', { name: 'Commander damage' })).toBeVisible();
  await expect(opponentCard.getByRole('heading', { name: 'Commander damage' })).toBeVisible();
  await expect(selfCard.getByText('Coming soon in v2')).toBeVisible();
  await expect(opponentCard.getByText('Coming soon in v2')).toBeVisible();
});

import { expect, test, type Page } from '@playwright/test';
import { createTestUser, CREATOR_DECKLIST, JOINER_DECKLIST, type TestUser } from './helpers/testData';

async function signUp(page: Page, user: TestUser) {
  await page.goto('/signup');
  const form = page.locator('form');
  await form.getByLabel(/^Username$/).fill(user.username);
  await form.getByLabel(/^Password$/).fill(user.password);
  await form.getByLabel(/^Confirm password$/).fill(user.password);
  await form.getByRole('button', { name: 'Sign up' }).click();
  await expect(page).toHaveURL(/\/games$/);
}

test('user can create a game and another user can join it', async ({ browser, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  const userA = createTestUser('e2e-a');
  const userB = createTestUser('e2e-b');

  const creatorContext = await browser.newContext();
  const creatorPage = await creatorContext.newPage();

  await signUp(creatorPage, userA);
  await creatorPage.getByRole('button', { name: 'Create game' }).click();
  await creatorPage.getByLabel(/Decklist \(CSV: quantity,name per line\)/).fill(CREATOR_DECKLIST);
  await creatorPage.locator('form').getByRole('button', { name: 'Create game' }).click();
  await expect(creatorPage).toHaveURL(/\/games\/[^/]+$/);

  const gameID = creatorPage.url().split('/games/')[1]?.split('?')[0];
  expect(gameID).toBeTruthy();

  const joinerContext = await browser.newContext();
  const joinerPage = await joinerContext.newPage();

  await signUp(joinerPage, userB);
  await joinerPage.goto(`/join/${gameID}`);
  await expect(joinerPage.getByText(`You are about to join game ${gameID}.`)).toBeVisible();
  await joinerPage.getByLabel(/Decklist \(CSV: quantity,name per line\)/).fill(JOINER_DECKLIST);
  await joinerPage.getByRole('button', { name: 'Join game' }).click();
  await expect(joinerPage).toHaveURL(new RegExp(`/games/${gameID}$`));

  await expect(joinerPage.getByRole('heading', { name: userB.username })).toBeVisible();
  await expect(joinerPage.getByRole('heading', { name: userA.username })).toBeVisible();

  await creatorContext.close();
  await joinerContext.close();
});

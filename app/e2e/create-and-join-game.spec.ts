import { expect, test, type Page } from '@playwright/test';
import { createTestUser, CREATOR_DECKLIST, e2eSession, JOINER_DECKLIST, type TestUser } from './helpers/testData';

async function signUp(page: Page, user: TestUser) {
  await page.goto('/signup');
  const form = page.locator('form');
  await form.getByLabel(/^Username$/).fill(user.username);
  await form.getByLabel(/^Password$/).fill(user.password);
  await form.getByLabel(/^Confirm password$/).fill(user.password);
  await form.getByRole('button', { name: 'Sign up' }).click();
  await expect(page).toHaveURL(/\/games$/);
}

async function previewDeck(page: Page, decklist: string) {
  const panel = page.getByTestId('deck-import-panel');
  await panel.getByLabel('Decklist').fill(decklist);
  await panel.getByTestId('deck-preview-submit').click();
  await expect(panel.getByTestId('paste-import-success')).toBeVisible();
}

async function importDeckAndSelectCommander(page: Page, decklist: string) {
  await previewDeck(page, decklist);
  const panel = page.getByTestId('deck-import-panel');
  const continueButton = panel.getByTestId('continue-to-commanders');
  await expect(continueButton).toBeEnabled();
  await continueButton.click();
  const commander = page.getByTestId('commander-candidate').first();
  await expect(commander).toBeVisible();
  await commander.click();
}

test('user can create a game and another user can join it', async ({ browser, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  const userA = createTestUser('e2e-a');
  const userB = createTestUser('e2e-b');

  const creatorContext = await browser.newContext();
  await creatorContext.addInitScript((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), e2eSession('auth-host'));
  const creatorPage = await creatorContext.newPage();

  await signUp(creatorPage, userA);
  await creatorPage.getByRole('button', { name: 'Create game' }).click();
  await previewDeck(creatorPage, CREATOR_DECKLIST);
  await creatorPage.locator('form').getByRole('button', { name: 'Create game' }).click();
  await expect(creatorPage).toHaveURL(/\/games\/[^/]+$/);

  const gameID = creatorPage.url().split('/games/')[1]?.split('?')[0];
  expect(gameID).toBeTruthy();

  const joinerContext = await browser.newContext();
  await joinerContext.addInitScript((sessionID) => localStorage.setItem('edhgo/session-id', sessionID), e2eSession('auth-invitee'));
  const joinerPage = await joinerContext.newPage();

  await signUp(joinerPage, userB);
  await joinerPage.goto(`/join/${gameID}`);
  await expect(joinerPage.getByTestId('invite-summary')).toContainText(userA.username);
  await importDeckAndSelectCommander(joinerPage, JOINER_DECKLIST);
  await joinerPage.getByTestId('join-table').click();
  await expect(joinerPage).toHaveURL(new RegExp(`/games/${gameID}$`));

  await expect(joinerPage.getByRole('heading', { name: userB.username })).toBeVisible();
  await expect(joinerPage.getByRole('heading', { name: userA.username })).toBeVisible();

  await creatorContext.close();
  await joinerContext.close();
});

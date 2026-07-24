import { expect, test, type Page } from '@playwright/test';
import { createTestUser, type TestUser } from './helpers/testData';

async function signUp(page: Page, user: TestUser) {
  await page.goto('/signup');
  const form = page.locator('form');
  await form.getByLabel(/^Username$/).fill(user.username);
  await form.getByLabel(/^Password$/).fill(user.password);
  await form.getByLabel(/^Confirm password$/).fill(user.password);
  await form.getByRole('button', { name: 'Sign up' }).click();
  await expect(page).toHaveURL(/\/games$/);
}

test('auth redirect sends logged-out users back to their intended page after login', async ({ browser, baseURL }) => {
  test.skip(!baseURL, 'VEDH_APP_BASE_URL or Playwright baseURL is required');

  const user = createTestUser('e2e-auth');
  const registrationContext = await browser.newContext();
  const registrationPage = await registrationContext.newPage();

  await signUp(registrationPage, user);
  await registrationContext.close();

  const page = await browser.newPage();

  await page.goto('/games');
  await expect(page).toHaveURL(/\/login\?redirect=/);
  expect(new URL(page.url()).searchParams.get('redirect')).toBe('/games');

  const form = page.locator('form');
  await form.getByLabel(/^Username$/).fill(user.username);
  await form.getByLabel(/^Password$/).fill(user.password);
  await form.getByRole('button', { name: 'Log in' }).click();

  await expect(page).toHaveURL(/\/games$/);
  await expect(page.getByRole('heading', { name: 'Your games' })).toBeVisible();
});

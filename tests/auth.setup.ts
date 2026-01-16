import { test as setup, expect } from '@playwright/test';

const authFile = 'playwright/.auth/admin.json';

setup('authenticate', async ({ page }) => {
  // Default credentials for local testing only
  // In CI/CD, set GRAFANA_ADMIN_USER and GRAFANA_ADMIN_PASSWORD environment variables
  const username = process.env.GRAFANA_ADMIN_USER ?? 'admin';
  const password = process.env.GRAFANA_ADMIN_PASSWORD ?? 'admin';
  const baseURL = process.env.GRAFANA_URL ?? 'http://localhost:3000';

  await page.goto(`${baseURL}/login`);
  await page.waitForLoadState('networkidle');

  // Use multiple selectors to work across Grafana versions
  const usernameInput = page
    .getByTestId('data-testid Username input field')
    .or(page.getByPlaceholder('email or username'))
    .or(page.locator('input[name="user"]'))
    .or(page.locator('input[type="text"]').first());

  const passwordInput = page
    .getByTestId('data-testid Password input field')
    .or(page.getByPlaceholder('password'))
    .or(page.locator('input[name="password"]'))
    .or(page.locator('input[type="password"]'));

  await usernameInput.fill(username);
  await passwordInput.fill(password);

  await page.getByRole('button', { name: 'Log in' }).click();

  // Handle password change dialog if it appears (Grafana prompts on first login with default credentials)
  const skipButton = page.getByRole('button', { name: 'Skip' });
  try {
    await skipButton.waitFor({ state: 'visible', timeout: 5000 });
    await skipButton.click();
  } catch {
    // Password change dialog didn't appear, continue
  }

  // Wait for successful login - should redirect to home or dashboard
  await expect(page).not.toHaveURL(/\/login/, { timeout: 15000 });

  // Ensure we're fully logged in
  await page.waitForLoadState('networkidle');

  // Wait for cookies to be set
  await page.waitForTimeout(1000);

  await page.context().storageState({ path: authFile });
});

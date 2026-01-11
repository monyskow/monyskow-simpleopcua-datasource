import { test as setup } from '@playwright/test';

const authFile = 'playwright/.auth/admin.json';

setup('authenticate', async ({ page, request }) => {
  const username = process.env.GRAFANA_ADMIN_USER ?? 'admin';
  const password = process.env.GRAFANA_ADMIN_PASSWORD ?? 'admin';
  const baseURL = process.env.GRAFANA_URL ?? 'http://localhost:3000';

  await page.goto(`${baseURL}/login`);

  await page.getByTestId('data-testid Username input field').fill(username);
  await page.getByTestId('data-testid Password input field').fill(password);
  await page.getByRole('button', { name: 'Log in' }).click();

  await page.waitForURL(`${baseURL}/**`);

  await page.context().storageState({ path: authFile });
});

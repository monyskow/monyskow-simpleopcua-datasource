import { test, expect } from '@playwright/test';
import { OpcuaTestHelpers } from './helpers';

test.describe('OPC-UA Plugin Smoke Tests', () => {
  test('should load plugin successfully', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Navigate to plugins page
    await page.goto('/plugins');
    await page.waitForLoadState('networkidle');

    // Search for OPC-UA plugin
    const searchInput = page.getByPlaceholder(/search/i).or(page.locator('input[type="text"]')).first();

    if (await searchInput.isVisible({ timeout: 5000 })) {
      await searchInput.fill('OPC-UA');
      await page.waitForTimeout(1000);
    }

    // Plugin should be listed
    const pluginCard = page.getByText(/Simple OPC-UA|OPC-UA/i);
    const isVisible = await pluginCard.first().isVisible({ timeout: 5000 }).catch(() => false);

    expect(isVisible).toBeTruthy();
  });

  test('should have provisioned data source', async ({ page }) => {
    await page.goto('/connections/datasources');
    await page.waitForLoadState('networkidle');

    // Look for OPC-UA Test Server data source
    const dataSource = page.getByText('OPC-UA Test Server');
    await expect(dataSource).toBeVisible({ timeout: 10000 });
  });

  test('should open data source configuration', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig();

    // Verify config page loaded - look for heading or unique element
    await expect(page.getByRole('heading', { name: /OPC-UA Test Server/i })).toBeVisible();
  });

  test('should display all configuration fields', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig();

    // Check for key configuration fields
    const endpointInput = page.getByLabel(/endpoint url/i);
    await expect(endpointInput).toBeVisible();

    // Security settings should be present
    const securityElements = page.locator('select, [role="combobox"]');
    const count = await securityElements.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should load query editor in explore', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToExplore();

    // Query editor should be visible
    const addNodeButton = page.getByRole('button', { name: /add manual/i }).first();
    await expect(addNodeButton).toBeVisible({ timeout: 10000 });
  });

  test('should allow adding nodes to query', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToExplore();

    await helpers.addNode('ns=2;s=TestNode');

    // Verify node was added
    const nodeInput = page.locator('input[placeholder*="ns="]').first();
    await expect(nodeInput).toHaveValue('ns=2;s=TestNode');
  });

  test('should persist endpoint configuration', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig();

    const originalEndpoint = await helpers.getEndpoint();
    expect(originalEndpoint).toBeTruthy();

    // Reload page
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Endpoint should still be there
    const endpointAfterReload = await helpers.getEndpoint();
    expect(endpointAfterReload).toBe(originalEndpoint);
  });

  test('should work with multiple Grafana versions', async ({ page }) => {
    // This test verifies basic compatibility across versions
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check Grafana is running
    const title = await page.title();
    expect(title).toContain('Grafana');

    // Navigate to data sources
    await page.goto('/connections/datasources');
    await page.waitForLoadState('networkidle');

    // OPC-UA data source should be available
    const dataSource = page.getByText('OPC-UA Test Server');
    await expect(dataSource).toBeVisible({ timeout: 10000 });

    // Open config
    await dataSource.click();
    await page.waitForLoadState('networkidle');

    // Core configuration should work
    const endpointInput = page.getByLabel(/endpoint url/i);
    await expect(endpointInput).toBeVisible();
  });

  test('should handle navigation between pages', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);

    // Config -> Explore
    await helpers.goToDataSourceConfig();
    await expect(page.getByRole('heading', { name: /OPC-UA Test Server/i })).toBeVisible();

    await helpers.goToExplore();
    const addButton = page.getByRole('button', { name: /add manual/i }).first();
    await expect(addButton).toBeVisible({ timeout: 10000 });
  });

  test('should not crash on invalid input', async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToExplore();

    // add manual with invalid input
    await helpers.addNode('invalid-node-123');

    // Page should still be responsive
    const addButton = page.getByRole('button', { name: /add manual/i }).first();
    await expect(addButton).toBeVisible({ timeout: 5000 });
  });
});

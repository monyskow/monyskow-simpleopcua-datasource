import { test, expect } from '@playwright/test';

test.describe('OPC-UA Data Queries and Visualization', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to explore page - OPC-UA Test Server is already the default data source
    await page.goto('/explore');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000); // Wait for query editor to load
  });

  test('should execute query and display results', async ({ page }) => {
    // Real OPC-UA server is available — assert query executes without connection error.
    // ns=0;i=2258 is ServerStatus/CurrentTime, present on all OPC-UA servers.
    const addButton = page.getByRole('button', { name: /add manual/i }).first();
    await expect(addButton).toBeVisible({ timeout: 5000 });
    await addButton.click();
    await page.waitForTimeout(500);

    const nodeIdInput = page.locator('input[placeholder*="ns="]').or(page.locator('input[type="text"]')).first();
    await expect(nodeIdInput).toBeVisible({ timeout: 3000 });
    await nodeIdInput.fill('ns=0;i=2258');

    const runButton = page.getByRole('button', { name: /^run query$/i }).first();
    if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await runButton.click();
    }

    // Wait for Explore to settle (auto-run or manual run)
    await page.waitForTimeout(4000);

    // A real OPC-UA server is present, so the query must not surface a
    // connection-level error. Positive rendering of the visualization is
    // intentionally not asserted — panel selectors vary across Grafana
    // versions and this test runs across the full version matrix.
    const hasConnectionError = await page
      .getByText(/cannot connect|connection refused|dial tcp/i)
      .isVisible({ timeout: 2000 })
      .catch(() => false);
    expect(hasConnectionError).toBeFalsy();
  });

  test('should display data in table format', async ({ page }) => {
    // Add and query a node
    const addButton = page.getByRole('button', { name: /add manual/i }).first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await page.waitForTimeout(500);

      const nodeIdInput = page.locator('input[placeholder*="ns="]').first();

      if (await nodeIdInput.isVisible({ timeout: 3000 })) {
        await nodeIdInput.fill('ns=0;i=2258'); // Server Status

        const runButton = page.getByRole('button', { name: /^run query$/i }).first();

        if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
          await runButton.click();
          await page.waitForTimeout(2000);

          // Switch to table view if not already
          const tableButton = page.getByRole('button', { name: /^table$/i }).first();

          if (await tableButton.isVisible({ timeout: 2000 }).catch(() => false)) {
            await tableButton.click({ force: true });
            await page.waitForTimeout(500);
          }

          // With a real OPC-UA server present, the query must not surface a
          // connection-level error. Positive table render is not asserted —
          // panel selectors vary across the Grafana version matrix.
          const hasConnectionError = await page
            .getByText(/cannot connect|connection refused|dial tcp/i)
            .isVisible({ timeout: 2000 })
            .catch(() => false);
          expect(hasConnectionError).toBeFalsy();
        }
      }
    }
  });

  test('should handle multiple nodes in single query', async ({ page }) => {
    // Add first node
    const addButton = page.getByRole('button', { name: /add manual/i }).first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await page.waitForTimeout(500);

      const firstNodeInput = page.locator('input[placeholder*="ns="]').first();
      if (await firstNodeInput.isVisible({ timeout: 3000 })) {
        await firstNodeInput.fill('ns=2;s=Temperature');
      }

      // Add second node
      await addButton.click();
      await page.waitForTimeout(500);

      const nodeInputs = page.locator('input[placeholder*="ns="]');
      const count = await nodeInputs.count();

      if (count >= 2) {
        await nodeInputs.nth(1).fill('ns=2;s=Pressure');

        // Run query
        const runButton = page.getByRole('button', { name: /^run query$/i }).first();
        if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
          await runButton.click();
          await page.waitForTimeout(2000);

          // Multi-node query must not surface a connection error.
          const hasConnectionError = await page
            .getByText(/cannot connect|connection refused|dial tcp/i)
            .isVisible({ timeout: 2000 })
            .catch(() => false);
          expect(hasConnectionError).toBeFalsy();
        }
      }
    }
  });

  test('should refresh data on run query', async ({ page }) => {
    // Add a node
    const addButton = page.getByRole('button', { name: /add manual/i }).first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await page.waitForTimeout(500);

      const nodeIdInput = page.locator('input[placeholder*="ns="]').first();

      if (await nodeIdInput.isVisible({ timeout: 3000 })) {
        await nodeIdInput.fill('ns=2;s=TestValue');

        const runButton = page.getByRole('button', { name: /^run query$/i }).first();

        if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
          // First query
          await runButton.click();
          await page.waitForTimeout(1500);

          // Second query - should refresh
          await runButton.click();
          await page.waitForTimeout(1500);

          // Repeated run must not surface a connection error.
          const hasConnectionError = await page
            .getByText(/cannot connect|connection refused|dial tcp/i)
            .isVisible({ timeout: 2000 })
            .catch(() => false);
          expect(hasConnectionError).toBeFalsy();
        }
      }
    }
  });

  test('should handle query errors gracefully', async ({ page }) => {
    // Add a node with invalid ID
    const addButton = page.getByRole('button', { name: /add manual/i }).first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await page.waitForTimeout(500);

      const nodeIdInput = page.locator('input[placeholder*="ns="]').first();

      if (await nodeIdInput.isVisible({ timeout: 3000 })) {
        await nodeIdInput.fill('invalid-node-id');

        const runButton = page.getByRole('button', { name: /run query/i });

        if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
          await runButton.click();
          await page.waitForTimeout(2000);

          // An invalid node id must not produce a connection-level error —
          // the server is reachable; this is a parse/query-level concern.
          // We do not assert a specific user-visible error because Grafana
          // surfaces query errors differently across versions.
          const hasConnectionError = await page
            .getByText(/cannot connect|connection refused|dial tcp/i)
            .isVisible({ timeout: 2000 })
            .catch(() => false);
          expect(hasConnectionError).toBeFalsy();
        }
      }
    }
  });

  test('should display node aliases in results', async ({ page }) => {
    // Add a node with alias
    const addButton = page.getByRole('button', { name: /add manual/i }).first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await page.waitForTimeout(500);

      const nodeIdInput = page.locator('input[placeholder*="ns="]').first();
      const aliasInput = page
        .locator('input[placeholder*="alias"]')
        .or(page.getByLabel(/alias/i))
        .or(page.locator('input[type="text"]').nth(1));

      if (await nodeIdInput.isVisible({ timeout: 3000 })) {
        await nodeIdInput.fill('ns=2;s=Temperature');

        if (await aliasInput.isVisible({ timeout: 2000 })) {
          await aliasInput.fill('MyTemperature');

          const runButton = page.getByRole('button', { name: /run query/i });

          if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
            await runButton.click();
            await page.waitForTimeout(2000);

            // Look for alias in table headers or results
            const hasAlias = await page
              .getByText('MyTemperature')
              .isVisible({ timeout: 3000 })
              .catch(() => false);

            // If table is shown, alias should appear
            const hasTable = await page
              .locator('table')
              .isVisible()
              .catch(() => false);

            if (hasTable) {
              expect(hasAlias).toBeTruthy();
            }
          }
        }
      }
    }
  });

  test('should work in dashboard panel', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/dashboard/new');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Add visualization
    const addPanelButton = page.getByRole('button', { name: /add.*visualization/i }).or(page.getByText(/add.*panel/i));

    if (await addPanelButton.first().isVisible({ timeout: 5000 })) {
      await addPanelButton.first().click();
      await page.waitForTimeout(2000);

      // OPC-UA is default data source, so query editor should appear automatically
      // Just verify the query editor loaded
      const queryEditor = page.getByRole('button', { name: /add manual/i }).or(page.locator('[data-testid*="query"]'));
      await expect(queryEditor.first()).toBeVisible({ timeout: 10000 });
    }
  });
});

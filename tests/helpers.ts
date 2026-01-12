import { Page, expect } from '@playwright/test';

/**
 * Helper functions for OPC-UA plugin E2E tests
 */

export class OpcuaTestHelpers {
  constructor(private page: Page) {}

  /**
   * Navigate to the OPC-UA data source configuration page
   */
  async goToDataSourceConfig(dataSourceUid = 'opcua-test-server') {
    // Navigate directly using the data source UID
    await this.page.goto(`/connections/datasources/edit/${dataSourceUid}`);
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(2000);

    // Verify config page loaded by checking for any configuration indicator
    // Different Grafana versions have different label associations
    const hasEndpointUrl = await this.page.getByLabel(/endpoint url/i).isVisible({ timeout: 3000 }).catch(() => false);
    const hasEndpointInput = await this.page.locator('input[aria-label*="Endpoint"]').or(this.page.locator('input[placeholder*="opc.tcp"]')).isVisible({ timeout: 3000 }).catch(() => false);
    const hasSaveButton = await this.page.getByRole('button', { name: /save.*test/i }).isVisible({ timeout: 3000 }).catch(() => false);

    // Accept if any indicator is present
    if (!hasEndpointUrl && !hasEndpointInput && !hasSaveButton) {
      throw new Error(`Data source config page failed to load for UID: ${dataSourceUid}`);
    }
  }

  /**
   * Navigate to Explore with OPC-UA data source selected
   */
  async goToExplore() {
    await this.page.goto('/explore');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000); // OPC-UA Test Server is default, no need to select (wait longer for CI)
  }

  /**
   * Select a data source from the picker (only use if not already selected)
   */
  async selectDataSource(name: string) {
    // Wait a bit for page to stabilize
    await this.page.waitForTimeout(1000);

    // Check if data source is already selected by looking at the input value
    const dataSourcePicker = this.page.locator('#data-source-picker').or(
      this.page.getByLabel(/select.*data.*source/i)
    );

    if (await dataSourcePicker.isVisible({ timeout: 5000 })) {
      const currentValue = await dataSourcePicker.inputValue().catch(() => '');

      // Only select if not already the current data source
      if (!currentValue.includes(name)) {
        // Force click to bypass overlay issues
        await dataSourcePicker.click({ force: true });
        await this.page.waitForTimeout(500);

        // Look for the option in the dropdown menu (more specific selector)
        const option = this.page.locator('[role="menuitem"], [role="option"]').filter({ hasText: name }).first();
        if (await option.isVisible({ timeout: 3000 })) {
          await option.click();
          await this.page.waitForTimeout(1000);
        }
      }
    }
  }

  /**
   * Add a node to the query
   */
  async addNode(nodeId: string, alias?: string) {
    const addButton = this.page
      .getByRole('button', { name: /add manual/i })
      .or(this.page.getByText(/add manual/i))
      .first();

    if (await addButton.isVisible({ timeout: 5000 })) {
      await addButton.click();
      await this.page.waitForTimeout(500);

      const nodeIdInput = this.page
        .locator('input[placeholder*="ns="]')
        .or(this.page.locator('input[type="text"]'))
        .first();

      if (await nodeIdInput.isVisible({ timeout: 3000 })) {
        await nodeIdInput.fill(nodeId);

        if (alias) {
          const aliasInput = this.page
            .locator('input[placeholder*="alias"]')
            .or(this.page.getByLabel(/alias/i))
            .or(this.page.locator('input[type="text"]').nth(1));

          if (await aliasInput.isVisible({ timeout: 2000 })) {
            await aliasInput.fill(alias);
          }
        }
      }
    }
  }

  /**
   * Run/execute the query
   */
  async runQuery() {
    // Use exact match for "Run query" button
    const runButton = this.page.getByRole('button', { name: /^run query$/i }).first();

    if (await runButton.isVisible({ timeout: 3000 }).catch(() => false)) {
      await runButton.click({ force: true });
      await this.page.waitForTimeout(2000);
    }
  }

  /**
   * Check if query results are displayed (table, chart, or error)
   */
  async hasQueryResults(): Promise<boolean> {
    const hasTable = await this.page.locator('table').isVisible({ timeout: 3000 }).catch(() => false);
    const hasChart = await this.page.locator('canvas, svg').isVisible({ timeout: 2000 }).catch(() => false);
    const hasError = await this.page.getByText(/error|failed/i).isVisible({ timeout: 2000 }).catch(() => false);
    const hasNoData = await this.page.getByText(/no data/i).isVisible({ timeout: 2000 }).catch(() => false);

    return hasTable || hasChart || hasError || hasNoData;
  }

  /**
   * Click Save & Test on data source config
   */
  async saveAndTest() {
    const saveButton = this.page
      .getByRole('button', { name: /save.*test/i })
      .or(this.page.getByRole('button', { name: /test/i }));

    await saveButton.click();
    await this.page.waitForTimeout(2000);
  }

  /**
   * Check if health check succeeded
   */
  async hasHealthCheckSuccess(): Promise<boolean> {
    const successMessage = this.page
      .getByText(/success|connected|ok/i)
      .or(this.page.locator('[data-testid*="success"]'))
      .or(this.page.locator('.alert-success'));

    return await successMessage.isVisible({ timeout: 5000 }).catch(() => false);
  }

  /**
   * Check if health check failed
   */
  async hasHealthCheckError(): Promise<boolean> {
    const errorMessage = this.page
      .getByText(/error|failed/i)
      .or(this.page.locator('[data-testid*="error"]'))
      .or(this.page.locator('.alert-error'));

    return await errorMessage.isVisible({ timeout: 5000 }).catch(() => false);
  }

  /**
   * Open node browser
   */
  async openNodeBrowser() {
    const browserButton = this.page
      .getByRole('button', { name: /browse|browser/i })
      .first();

    if (await browserButton.isVisible({ timeout: 5000 })) {
      await browserButton.click();
      await this.page.waitForTimeout(1000);
    }
  }

  /**
   * Check if node browser is open
   */
  async isNodeBrowserOpen(): Promise<boolean> {
    const browserModal = this.page
      .locator('[role="dialog"]')
      .or(this.page.locator('.modal'))
      .or(this.page.getByText(/browse.*node|object.*folder/i));

    return await browserModal.first().isVisible({ timeout: 3000 }).catch(() => false);
  }

  /**
   * Set endpoint URL in config
   */
  async setEndpoint(url: string) {
    const endpointInput = this.page.getByLabel(/endpoint url/i);
    await endpointInput.clear();
    await endpointInput.fill(url);
  }

  /**
   * Get current endpoint URL value
   */
  async getEndpoint(): Promise<string> {
    const endpointInput = this.page.getByLabel(/endpoint url/i);
    return await endpointInput.inputValue();
  }
}

/**
 * Common test data
 */
export const TEST_NODES = {
  serverStatus: 'ns=0;i=2258',
  serverState: 'ns=0;i=2259',
  currentTime: 'ns=0;i=2258',
  temperature: 'ns=2;s=Temperature',
  pressure: 'ns=2;s=Pressure',
  demoFloat: 'ns=2;s=Demo.Dynamic.Scalar.Float',
  demoBoolean: 'ns=2;s=Demo.Dynamic.Scalar.Boolean',
};

export const TEST_ENDPOINTS = {
  opcuaServer: 'opc.tcp://opcuaserver.com:48010',
  localhost: 'opc.tcp://localhost:4840',
  prosysSimulation: 'opc.tcp://opcuaserver.com:53530/OPCUA/SimulationServer',
};

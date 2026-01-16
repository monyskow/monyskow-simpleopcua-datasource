import { Page, expect } from '@playwright/test';

/**
 * Helper functions for OPC-UA plugin E2E tests
 */

export class OpcuaTestHelpers {
  constructor(private page: Page) {}

  /**
   * Navigate to the OPC-UA data source configuration page
   * Uses the standard Grafana URL pattern that works across versions
   */
  async goToDataSourceConfig(dataSourceUid = 'opcua-test-server') {
    // Use standard /datasources/edit/ path (works in all Grafana versions)
    await this.page.goto(`/datasources/edit/${dataSourceUid}`, { waitUntil: 'networkidle' });
    await this.page.waitForTimeout(3000);

    // Verify config page loaded - use multiple selectors for cross-version compatibility
    // Grafana 10.x/11.x don't properly associate InlineField labels with inputs
    const endpointInput = this.page
      .locator('[aria-label="Endpoint URL"]')
      .or(this.page.locator('input[placeholder*="opc.tcp"]'))
      .or(this.page.getByLabel(/endpoint url/i));
    try {
      await endpointInput.first().waitFor({ state: 'visible', timeout: 5000 });
    } catch {
      // Fallback: wait additional time for slower Grafana versions
      await this.page.waitForTimeout(2000);
    }
  }

  /**
   * Navigate to Explore with OPC-UA data source selected
   */
  async goToExplore() {
    await this.page.goto('/explore');
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000);
  }

  /**
   * Select a data source from the picker (only use if not already selected)
   */
  async selectDataSource(name: string) {
    await this.page.waitForTimeout(1000);

    const dataSourcePicker = this.page.locator('#data-source-picker').or(this.page.getByLabel(/select.*data.*source/i));

    if (await dataSourcePicker.isVisible({ timeout: 5000 })) {
      const currentValue = await dataSourcePicker.inputValue().catch(() => '');

      if (!currentValue.includes(name)) {
        await dataSourcePicker.click({ force: true });
        await this.page.waitForTimeout(500);

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
    const hasTable = await this.page
      .locator('table')
      .isVisible({ timeout: 3000 })
      .catch(() => false);
    const hasChart = await this.page
      .locator('canvas, svg')
      .isVisible({ timeout: 2000 })
      .catch(() => false);
    const hasError = await this.page
      .getByText(/error|failed/i)
      .isVisible({ timeout: 2000 })
      .catch(() => false);
    const hasNoData = await this.page
      .getByText(/no data/i)
      .isVisible({ timeout: 2000 })
      .catch(() => false);

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
    const browserButton = this.page.getByRole('button', { name: /browse|browser/i }).first();

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

    return await browserModal
      .first()
      .isVisible({ timeout: 3000 })
      .catch(() => false);
  }

  /**
   * Get endpoint input locator (cross-version compatible)
   */
  private getEndpointInput() {
    // Use multiple selectors for cross-version compatibility
    // Grafana 10.x/11.x don't properly associate InlineField labels with inputs
    return this.page
      .locator('[aria-label="Endpoint URL"]')
      .or(this.page.locator('input[placeholder*="opc.tcp"]'))
      .or(this.page.getByLabel(/endpoint url/i));
  }

  /**
   * Set endpoint URL in config
   */
  async setEndpoint(url: string) {
    const endpointInput = this.getEndpointInput().first();
    await endpointInput.clear();
    await endpointInput.fill(url);
  }

  /**
   * Get current endpoint URL value
   */
  async getEndpoint(): Promise<string> {
    const endpointInput = this.getEndpointInput().first();
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

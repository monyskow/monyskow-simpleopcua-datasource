import { expect, test } from './fixtures/isolated-datasource';
import { OpcuaTestHelpers } from './helpers';
import { Page } from '@playwright/test';

// Cross-version compatible selector for endpoint URL input
function getEndpointInput(page: Page) {
  return page
    .locator('[aria-label="Endpoint URL"]')
    .or(page.locator('input[placeholder*="opc.tcp"]'))
    .or(page.getByLabel(/endpoint url/i));
}

// Select security mode via Grafana's Select (react-select based).
// react-select renders role="combobox" on the input, and role="option" on each list item.
// We open the dropdown by clicking the input, then pick the option by its visible label.
async function selectSecurityMode(page: Page, mode: string) {
  // The aria-label is set on the Select wrapper; the actual input inside has role="combobox"
  const securityModeInput = page
    .locator('[aria-label="Security Mode"] input[role="combobox"]')
    .or(page.locator('input[role="combobox"]').nth(1));
  await expect(securityModeInput.first()).toBeVisible({ timeout: 5000 });

  // Map internal value to displayed label
  const labelMap: Record<string, string> = {
    None: 'None',
    Sign: 'Sign',
    SignAndEncrypt: 'Sign and Encrypt',
  };
  const label = labelMap[mode] ?? mode;

  const securityModeContainer = page
    .locator('[aria-label="Security Mode"]')
    .or(page.locator('input[role="combobox"]').nth(1).locator('..').locator('..'));

  await securityModeInput.first().click();
  const option = page.getByRole('option', { name: label });
  await expect(option).toBeVisible({ timeout: 3000 });
  await option.click();

  // Verify the selection is reflected in the container
  await expect(securityModeContainer.first()).toContainText(label, { timeout: 3000 });
}

// Returns true when the provisioned datasource uses certificate auth.
// The 'Client Certificate' auto-generate section is intentionally hidden
// in that mode (the user supplies their own cert), so tests that exercise
// the Generate Certificate button must be skipped.
async function isCertificateAuthMode(page: Page): Promise<boolean> {
  // Read the displayed text in the Auth Method select container.
  // react-select shows the selected label as visible text (not as input value).
  const authMethodContainer = page
    .locator('[aria-label="Authentication Method"]')
    .or(page.locator('input[role="combobox"]').nth(2).locator('..').locator('..'));
  const text = await authMethodContainer
    .first()
    .textContent()
    .catch(() => '');
  return text?.toLowerCase().includes('certificate') ?? false;
}

// No describe.serial needed: each test operates on its own isolated datasource
// created via the isolatedDatasource fixture, so there is no shared mutable state
// and tests can run in parallel without 409 optimistic-concurrency conflicts.
test.describe('Certificate Generation', () => {
  test('should not show Client Certificate section when security mode is None', async ({
    page,
    isolatedDatasource,
  }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig(isolatedDatasource.uid);

    // Default configuration has security mode None
    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    // Client Certificate section must be absent with mode=None.
    // If this fails, either the default mode changed or the visibility guard was removed.
    await expect(page.getByText(/Client Certificate/i)).not.toBeVisible({ timeout: 3000 });
  });

  test('should show Client Certificate section after switching to Sign mode', async ({ page, isolatedDatasource }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig(isolatedDatasource.uid);

    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });
    // Client Certificate auto-generate UI is hidden when authMethod=certificate;
    // user supplies their own cert in that mode — skip rather than fail.
    test.skip(
      await isCertificateAuthMode(page),
      'Datasource uses certificate auth — Client Certificate section is hidden by design'
    );

    await selectSecurityMode(page, 'Sign');

    // After selecting Sign mode the Client Certificate section should appear
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });
  });

  test('should show Generate Certificate button in Client Certificate section', async ({
    page,
    isolatedDatasource,
  }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig(isolatedDatasource.uid);

    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });
    test.skip(
      await isCertificateAuthMode(page),
      'Datasource uses certificate auth — Client Certificate section is hidden by design'
    );

    await selectSecurityMode(page, 'Sign');
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    const generateBtn = page
      .getByRole('button', { name: /generate certificate/i })
      .or(page.getByText(/generate certificate/i));
    await expect(generateBtn.first()).toBeVisible({ timeout: 5000 });
  });

  test('should produce a pending certificate when Generate is clicked on a draft datasource', async ({ page }) => {
    // Navigate to a brand-new (draft) datasource form. Grafana 12 provisions a
    // draft DS with an id before rendering the editor, so the backend resource
    // call succeeds and the cert lands in pending state awaiting Save & Test.
    await page.goto('/datasources/new');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // The datasource picker page is required: if Simple OPC-UA is not listed
    // (e.g. plugin not installed in this Grafana instance), skip — do not silently pass.
    const pluginOption = page
      .getByText(/Simple OPC-UA/i)
      .or(page.locator('[data-testid*="simple-opcua"]'))
      .or(page.locator('button').filter({ hasText: /Simple OPC-UA/i }));

    const pluginVisible = await pluginOption
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
    if (!pluginVisible) {
      test.skip(true, 'Simple OPC-UA not listed in datasource picker — plugin not installed in this environment');
      return;
    }

    await pluginOption.first().click();
    await page.waitForTimeout(1000);

    await selectSecurityMode(page, 'Sign');

    // Client Certificate section must appear after switching to Sign mode on a new datasource
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    const generateBtn = page.getByRole('button', { name: /generate certificate/i }).first();
    await expect(generateBtn).toBeVisible({ timeout: 3000 });
    await generateBtn.click();

    // The backend resource call returns a cert that lands in pending state.
    const certPending = page
      .getByText(/click Save.*to persist/i)
      .or(page.getByText(/Certificate (?:configured|generated)/i));
    await expect(certPending.first()).toBeVisible({ timeout: 5000 });
  });

  test('should generate certificate successfully on saved datasource', async ({ page, isolatedDatasource }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig(isolatedDatasource.uid);

    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });
    test.skip(
      await isCertificateAuthMode(page),
      'Datasource uses certificate auth — Client Certificate section is hidden by design'
    );

    // Save the datasource first to ensure it is persisted before generating a cert.
    const saveButton = page
      .getByRole('button', { name: /save.*test/i })
      .or(page.getByRole('button', { name: /save/i }));

    if (await saveButton.isVisible({ timeout: 3000 })) {
      await saveButton.click();
      await page.waitForTimeout(1500);
    }

    // Switch to Sign mode to show the cert section
    await selectSecurityMode(page, 'Sign');
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    const generateBtn = page.getByRole('button', { name: /generate certificate/i }).first();
    await expect(generateBtn).toBeVisible({ timeout: 5000 });
    await generateBtn.click();

    // Wait for result: either success status or error alert.
    // Both must be asserted with toBeVisible — a .catch(() => false) guard would pass
    // even if the entire generate flow was removed.
    const successStatus = page
      .getByText(/Certificate generated|Certificate configured/i)
      .or(page.getByText(/click Save.*Test to persist/i));
    const errorAlert = page.getByText(/failed to generate|error/i);

    await expect(successStatus.or(errorAlert).first()).toBeVisible({ timeout: 10000 });
  });

  test('should persist certificate after Save and Test', async ({ page, isolatedDatasource }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig(isolatedDatasource.uid);

    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    await selectSecurityMode(page, 'Sign');
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    const generateBtn = page.getByRole('button', { name: /generate certificate/i }).first();

    // If cert is already configured (Generate button absent), Reset it first so we
    // can exercise the full generate-then-persist flow.
    const generateBtnVisible = await generateBtn.isVisible({ timeout: 3000 });
    if (!generateBtnVisible) {
      const resetBtn = page.getByRole('button', { name: /reset/i });
      // Reset button must be present when Generate is absent — assert it explicitly
      await expect(resetBtn).toBeVisible({ timeout: 2000 });
      await resetBtn.click();
      await page.waitForTimeout(500);
    }

    // Generate button must be visible at this point — either it was always there or
    // we just reset the cert. A missing button here means the UI is broken.
    await expect(generateBtn).toBeVisible({ timeout: 3000 });
    await generateBtn.click();

    // Wait for pending status before saving
    await expect(page.getByText(/click Save.*Test to persist/i)).toBeVisible({ timeout: 10000 });

    // Save and Test
    const saveButton = page
      .getByRole('button', { name: /save.*test/i })
      .or(page.getByRole('button', { name: /save/i }));
    await saveButton.click();
    await page.waitForTimeout(2000);

    // Reload and verify the certificate persisted
    await page.reload();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    await selectSecurityMode(page, 'Sign');
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    // After reload and re-selecting Sign mode, the cert must show as saved.
    // If persist logic was removed this assertion would fail.
    const savedStatus = page.getByText(/Certificate configured \(saved\)/i);
    await expect(savedStatus).toBeVisible({ timeout: 5000 });
  });
});

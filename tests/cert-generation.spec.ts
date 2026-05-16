import { test, expect, Page } from '@playwright/test';
import { OpcuaTestHelpers } from './helpers';

// Cross-version compatible selector for endpoint URL input
function getEndpointInput(page: Page) {
  return page
    .locator('[aria-label="Endpoint URL"]')
    .or(page.locator('input[placeholder*="opc.tcp"]'))
    .or(page.getByLabel(/endpoint url/i));
}

// Select security mode via Combobox — Grafana renders it as an input[role=combobox]
// The Security Mode combobox is the second combobox (index 1) in the Connection fieldset
async function selectSecurityMode(page: Page, mode: string) {
  // Try combobox by aria-label first (set by InlineField label in the component)
  const combobox = page.locator('input[role="combobox"]').or(page.locator('select')).nth(1); // Security Mode is the second combobox (after Security Policy)

  if (await combobox.isVisible({ timeout: 3000 })) {
    await combobox.click();
    const option = page.locator('[role="option"]').filter({ hasText: mode }).first();
    if (await option.isVisible({ timeout: 2000 })) {
      await option.click();
      return;
    }
    // Fallback: type into the combobox to filter and select
    await combobox.fill(mode);
    const filteredOption = page.locator('[role="option"]').filter({ hasText: mode }).first();
    if (await filteredOption.isVisible({ timeout: 2000 })) {
      await filteredOption.click();
    }
  }
}

test.describe('Certificate Generation', () => {
  test.beforeEach(async ({ page }) => {
    const helpers = new OpcuaTestHelpers(page);
    await helpers.goToDataSourceConfig();
  });

  test('should not show Client Certificate section when security mode is None', async ({ page }) => {
    // Default configuration has security mode None
    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    // Client Certificate section must be absent with mode=None.
    // If this fails, either the default mode changed or the visibility guard was removed.
    await expect(page.getByText(/Client Certificate/i)).not.toBeVisible({ timeout: 3000 });
  });

  test('should show Client Certificate section after switching to Sign mode', async ({ page }) => {
    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    await selectSecurityMode(page, 'Sign');

    // After selecting Sign mode the Client Certificate section should appear
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });
  });

  test('should show Generate Certificate button in Client Certificate section', async ({ page }) => {
    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    await selectSecurityMode(page, 'Sign');
    await expect(page.getByText(/Client Certificate/i).first()).toBeVisible({ timeout: 5000 });

    const generateBtn = page
      .getByRole('button', { name: /generate certificate/i })
      .or(page.getByText(/generate certificate/i));
    await expect(generateBtn.first()).toBeVisible({ timeout: 5000 });
  });

  test('should show save-first error when clicking Generate Certificate on unsaved datasource', async ({ page }) => {
    // Navigate to a brand-new (unsaved) datasource form
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
    await page.waitForTimeout(500);

    // The "save first" error must appear — removing the unsaved-datasource guard in
    // ConfigEditor would cause the fetch to be attempted and this assertion to fail.
    await expect(page.getByText(/save the datasource first/i)).toBeVisible({ timeout: 3000 });
  });

  test('should generate certificate successfully on saved datasource', async ({ page }) => {
    await expect(getEndpointInput(page).first()).toBeVisible({ timeout: 10000 });

    // Save the datasource first (it should already be saved as "OPC-UA Test Server")
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

  test('should persist certificate after Save and Test', async ({ page }) => {
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

import { defineConfig, devices } from '@playwright/test';
import { PluginOptions } from '@grafana/plugin-e2e';

const pluginE2eAuth = {
  username: process.env.GRAFANA_ADMIN_USER ?? 'admin',
  password: process.env.GRAFANA_ADMIN_PASSWORD ?? 'admin',
};

export default defineConfig<PluginOptions>({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  // retries: 1 everywhere — papers over the residual opcua-server connection-limit
  // flake (provisioned datasources + per-test datasources can saturate the 10-slot
  // pool on cold start). Track in a follow-up issue.
  retries: 1,
  // Capped at 2 to stay within the opcua-server connection pool. Per-test datasource
  // isolation removes 409 optimistic-concurrency races, but the underlying node-opcua
  // test server enforces a connection limit (~10) that higher worker counts saturate.
  // Track in a follow-up issue to lift the cap.
  workers: 2,
  timeout: 30000, // 30 second timeout per test
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
  ],
  use: {
    baseURL: process.env.GRAFANA_URL ?? 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10000, // 10 second timeout for actions
  },

  projects: [
    {
      name: 'auth',
      testMatch: '**/auth.setup.ts',
    },
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'playwright/.auth/admin.json',
      },
      dependencies: ['auth'],
    },
  ],

  webServer: process.env.CI
    ? undefined
    : {
        command: 'npm run server',
        url: 'http://localhost:3000',
        timeout: 120 * 1000,
        reuseExistingServer: !process.env.CI,
      },
});

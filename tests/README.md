# OPC-UA Plugin E2E Tests

This directory contains end-to-end tests for the Simple OPC-UA Grafana data source plugin using Playwright.

## Test Structure

- **auth.setup.ts** - Authentication setup that runs before all tests
- **helpers.ts** - Reusable test helper functions and test data
- **smoke.spec.ts** - Basic smoke tests to verify plugin loads and core functionality works
- **datasource-config.spec.ts** - Tests for data source configuration UI
- **query-editor.spec.ts** - Tests for query editor and node browser
- **data-queries.spec.ts** - Tests for executing queries and displaying data

## Running Tests

### Prerequisites

1. Build the plugin:
```bash
npm run build
mage buildAll
```

2. Start Grafana with the plugin:
```bash
npm run server
```

### Run All Tests

```bash
npm run e2e
```

### Run Specific Test File

```bash
npx playwright test tests/smoke.spec.ts
```

### Run Tests in UI Mode (Interactive)

```bash
npx playwright test --ui
```

### Run Tests in Debug Mode

```bash
npx playwright test --debug
```

### Run Tests in Headed Mode (See Browser)

```bash
npx playwright test --headed
```

## Test Reports

After running tests, view the HTML report:

```bash
npx playwright show-report
```

## CI/CD Integration

The tests automatically run in CI against multiple Grafana versions:

- Minimum supported version (10.4.0)
- Latest LTS version
- Latest stable version

The version matrix is automatically determined by the `grafana/plugin-actions/e2e-version` GitHub Action based on the `grafanaDependency` in plugin.json.

## Test Against Specific Grafana Version

```bash
# Start Grafana with specific version
GRAFANA_VERSION=11.0.0 npm run server

# Run tests in another terminal
npm run e2e
```

## Writing New Tests

Use the helper class for common operations:

```typescript
import { test, expect } from '@playwright/test';
import { OpcuaTestHelpers } from './helpers';

test('my test', async ({ page }) => {
  const helpers = new OpcuaTestHelpers(page);

  await helpers.goToExplore();
  await helpers.addNode('ns=2;s=Temperature', 'Temp1');
  await helpers.runQuery();

  expect(await helpers.hasQueryResults()).toBeTruthy();
});
```

## Test Data

The tests use a provisioned OPC-UA Test Server data source configured in `provisioning/datasources/datasources.yaml`.

Common test nodes are available in `helpers.ts`:
- `TEST_NODES.serverStatus` - ns=0;i=2258
- `TEST_NODES.temperature` - ns=2;s=Temperature
- `TEST_NODES.pressure` - ns=2;s=Pressure

## Debugging Failed Tests

1. Check screenshots in `test-results/` directory
2. View trace files with: `npx playwright show-trace test-results/.../trace.zip`
3. Check Grafana logs: `docker logs monyskow-simpleopcua-datasource`

## Known Limitations

- Tests use public OPC-UA test servers which may be unavailable
- Some tests verify UI behavior rather than data accuracy
- Network timeouts may cause flaky tests - retries are configured in CI

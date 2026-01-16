# E2E Test Documentation

This directory contains End-to-End (E2E) tests for the Simple OPC-UA Grafana plugin using Playwright.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Test Suites](#test-suites)
- [Running Tests](#running-tests)
- [Test Configuration](#test-configuration)
- [Troubleshooting](#troubleshooting)

## Overview

The E2E test suite validates the complete user experience of the OPC-UA plugin across multiple Grafana versions (10.4.x, 11.x, 12.x). Tests run in a real browser (Chromium) against actual Grafana instances running in Docker containers.

**Test Coverage:**

- Authentication flows
- Plugin installation and visibility
- Data source configuration
- Query editor functionality
- Data querying and visualization
- Cross-version compatibility

**Grafana Versions Tested:**

- 10.4.19 (Grafana 10.x LTS)
- 11.1.13 (Grafana 11.1.x)
- 11.4.8 (Grafana 11.4.x)
- 12.1.5 (Grafana 12.1.x)
- 12.3.1 (Grafana 12.3.x - latest)

## Test Structure

```
tests/
├── README.md                      # This file
├── auth.setup.ts                  # Authentication setup (runs before all tests)
├── helpers.ts                     # Shared test utilities
├── data-queries.spec.ts           # Query execution and data visualization tests
├── datasource-config.spec.ts      # Data source configuration tests
├── plugin-metadata.spec.ts        # Plugin installation and metadata tests
├── query-editor.spec.ts           # Query editor UI tests
└── smoke.spec.ts                  # Basic smoke tests
```

## Test Suites

### 1. Authentication Setup (`auth.setup.ts`)

**Purpose:** Establishes authenticated session before running any tests.

**What it does:**

- Navigates to Grafana login page
- Fills in username/password (admin/admin)
- Handles password change dialog (clicks "Skip" if it appears)
- Waits for successful redirect to home page
- Saves authentication state to `playwright/.auth/admin.json`

**Key Features:**

- Uses multiple selectors for cross-version compatibility
- Handles Grafana's "Update your password" prompt
- Validates successful login by checking URL redirect
- Stores session cookies for reuse in other tests

---

### 2. Data Queries Tests (`data-queries.spec.ts`)

**Suite:** `OPC-UA Data Queries and Visualization` (8 tests)

| Test                                           | Description                                                                  |
| ---------------------------------------------- | ---------------------------------------------------------------------------- |
| `should execute query and display results`     | Adds node `ns=2;s=Demo.Dynamic.Scalar.Float`, runs query, verifies execution |
| `should display data in table format`          | Queries server status node, switches to table view                           |
| `should handle multiple nodes in single query` | Adds Temperature + Pressure nodes, runs combined query                       |
| `should refresh data on run query`             | Runs same query twice to verify refresh                                      |
| `should handle query errors gracefully`        | Tests invalid node ID, ensures no crash                                      |
| `should display node aliases in results`       | Sets alias "MyTemperature", checks table headers                             |
| `should work in dashboard panel`               | Creates dashboard, adds panel, verifies query editor                         |

**Coverage:** Query execution, visualization, error handling, aliases, dashboards

---

### 3. Data Source Configuration Tests (`datasource-config.spec.ts`)

**Suite:** `OPC-UA Data Source Configuration` (8 tests)

| Test                                                                | Description                                                |
| ------------------------------------------------------------------- | ---------------------------------------------------------- |
| `should load the configuration page`                                | Verifies config page loads with correct heading            |
| `should display endpoint URL field`                                 | Checks endpoint input with default value                   |
| `should display security policy selector`                           | Verifies security policy dropdown (None/Basic)             |
| `should display security mode selector`                             | Verifies security mode dropdown (None/Sign/SignAndEncrypt) |
| `should display authentication method selector`                     | Verifies auth dropdown (anonymous/username)                |
| `should test connection successfully`                               | Clicks "Save & Test", verifies workflow                    |
| `should allow changing endpoint URL`                                | Changes to `opc.tcp://localhost:4840`                      |
| `should show username/password fields when auth method is username` | Switches auth, checks fields appear                        |
| `should persist configuration after save`                           | Saves, reloads, verifies persistence                       |

**Coverage:** Config UI, security settings, authentication, persistence

---

### 4. Plugin Metadata Tests (`plugin-metadata.spec.ts`)

**Suite:** `OPC-UA Plugin Metadata and Compatibility` (11 tests)

| Test                                             | Description                                            |
| ------------------------------------------------ | ------------------------------------------------------ |
| `should display correct plugin information`      | Verifies plugin name visible on detail page            |
| `should show plugin as installed`                | **Uses 10+ selectors** for cross-version compatibility |
| `should display plugin description and metadata` | Checks for OPC-UA/industrial keywords                  |
| `should be categorized as data source`           | Filters plugins by type, finds OPC-UA                  |
| `should be discoverable via search`              | Searches: opcua, industrial, plc                       |
| `should allow creating new data source instance` | Finds "Create instance" button                         |
| `should work with current Grafana version`       | Opens config, proves compatibility                     |
| `should display backend plugin indicator`        | Checks for "Save & Test" button                        |
| `should support alerting (has backend)`          | Checks alerting page for OPC-UA                        |
| `should have required plugin files`              | Monitors console for load errors                       |

**Coverage:** Installation status, search, backend features, cross-version UI

---

### 5. Query Editor Tests (`query-editor.spec.ts`)

**Suite:** `OPC-UA Query Editor` (9 tests)

| Test                                                     | Description                       |
| -------------------------------------------------------- | --------------------------------- |
| `should display query editor with node list`             | Verifies query editor loads       |
| `should have add manual button`                          | Checks "Add manual" button exists |
| `should allow adding a node manually`                    | Adds node, verifies input value   |
| `should allow setting node alias`                        | Sets alias on node                |
| `should allow removing a node`                           | Removes node from list            |
| `should display node browser button`                     | Checks "Browse" button            |
| `should open node browser when browse button is clicked` | Opens browser dialog              |
| `should display root nodes in browser`                   | Verifies Objects folder loads     |
| `should persist query configuration`                     | Reloads page, checks persistence  |

**Coverage:** Query editor UI, node management, browser, persistence

---

### 6. Smoke Tests (`smoke.spec.ts`)

**Suite:** `OPC-UA Plugin Smoke Tests` (10 tests)

| Test                                         | Description                        |
| -------------------------------------------- | ---------------------------------- |
| `should load plugin successfully`            | Searches plugins page              |
| `should have provisioned data source`        | Checks "OPC-UA Test Server" exists |
| `should open data source configuration`      | Opens config page                  |
| `should display all configuration fields`    | Counts security fields             |
| `should load query editor in explore`        | Opens Explore, checks query editor |
| `should allow adding nodes to query`         | Adds test node                     |
| `should persist endpoint configuration`      | Reloads, checks endpoint           |
| `should work with multiple Grafana versions` | Basic version check                |
| `should handle navigation between pages`     | Config → Explore                   |
| `should not crash on invalid input`          | Tests error resilience             |

**Coverage:** Basic functionality, provisioning, navigation, error resilience

---

### 7. Test Helpers (`helpers.ts`)

**Class:** `OpcuaTestHelpers`

**Navigation Methods:**

- `goToDataSourceConfig(uid?)` - Navigate to data source config
- `goToExplore()` - Navigate to Explore page
- `selectDataSource(name)` - Select from data source picker

**Query Methods:**

- `addNode(nodeId, alias?)` - Add node to query
- `runQuery()` - Execute query
- `hasQueryResults()` - Check for results (table/chart/error)

**Config Methods:**

- `saveAndTest()` - Click "Save & Test"
- `hasHealthCheckSuccess()` - Check success message
- `hasHealthCheckError()` - Check error message
- `setEndpoint(url)` - Set endpoint URL
- `getEndpoint()` - Get current endpoint

**Browser Methods:**

- `openNodeBrowser()` - Open node browser
- `isNodeBrowserOpen()` - Check if open

**Test Data:**

```typescript
TEST_NODES = {
  serverStatus: 'ns=0;i=2258',
  temperature: 'ns=2;s=Temperature',
  pressure: 'ns=2;s=Pressure',
  // ...
};

TEST_ENDPOINTS = {
  opcuaServer: 'opc.tcp://opcuaserver.com:48010',
  localhost: 'opc.tcp://localhost:4840',
};
```

## Running Tests

### Single Grafana Version

```bash
# Start Grafana + plugin
npm run server

# Run all tests (in another terminal)
npm run e2e

# Run specific test file
npx playwright test tests/smoke.spec.ts

# Run specific test by name
npx playwright test -g "should load plugin successfully"

# Run with UI
npx playwright test --ui

# Run in headed mode (see browser)
npx playwright test --headed

# Run with debug inspector
npx playwright test --debug
```

### All Grafana Versions (Automated)

```bash
npm run e2e:all
```

**What it does:**

1. Builds plugin (frontend + backend)
2. For each Grafana version (10.4.19, 11.1.13, 11.4.8, 12.1.5, 12.3.1):
   - Starts Grafana container
   - Runs full test suite (46 tests)
   - Saves results
   - Stops container
3. Displays summary

**Output:** `e2e-results-YYYYMMDD-HHMMSS.txt`

**Expected:** ~10 minutes total (2 min per version)

### All Grafana Versions (Manual Testing)

```bash
# Start all 5 versions on different ports
npm run server:all
```

**Access URLs (login: admin / admin):**

- Grafana 10.4.19: http://localhost:3000
- Grafana 11.1.13: http://localhost:3001
- Grafana 11.4.8: http://localhost:3002
- Grafana 12.1.5: http://localhost:3003
- Grafana 12.3.1: http://localhost:3004

**Useful pages to test:**

- Plugin: `/plugins/monyskow-simpleopcua-datasource`
- Data sources: `/connections/datasources`
- Config: `/connections/datasources/edit/opcua-test-server`
- Explore: `/explore`

```bash
# Stop all when done
npm run server:all:stop
```

### View Test Reports

```bash
# Open HTML report
npx playwright show-report

# View specific trace
npx playwright show-trace test-results/.../trace.zip
```

## Test Configuration

**File:** `playwright.config.ts`

| Setting        | Value                       | Purpose            |
| -------------- | --------------------------- | ------------------ |
| `timeout`      | 30000ms                     | Max time per test  |
| `retries`      | 1 (CI), 0 (local)           | Retry flaky tests  |
| `workers`      | 2 (CI), unlimited (local)   | Parallel execution |
| `baseURL`      | http://localhost:3000       | Grafana URL        |
| `storageState` | playwright/.auth/admin.json | Auth cookies       |

**Projects:**

1. **auth** - Setup project (runs `auth.setup.ts`)
2. **chromium** - Main tests (depends on auth)

**CI Mode (`CI=true`):**

- Enables 1 retry
- Uses 2 workers
- Disables interactive report

## Troubleshooting

### Tests Fail with "Not authenticated"

**Symptom:** Tests see login page instead of content

**Fix:**

```bash
rm -f playwright/.auth/admin.json
npm run e2e
```

### Tests Timeout Waiting for Elements

**Symptom:** `TimeoutError: Locator not found`

**Debug:**

```bash
# Check plugin loaded
docker logs <container> | grep opcua

# Check Grafana errors
docker logs <container> | grep ERROR

# Run with debug
npx playwright test --debug
```

**Fix:** Increase wait time or check selectors

### Selector Not Found in Older Grafana

**Symptom:** Passes in v12, fails in v10/v11

**Fix:** Add fallback selectors:

```typescript
const button = page
  .getByRole('button', { name: 'New Label' })
  .or(page.getByRole('button', { name: 'Old Label' }))
  .or(page.getByText(/Label/i));
```

### Docker Port Conflicts

**Symptom:** `Port 3000 already in use`

**Fix:**

```bash
docker ps | grep grafana | awk '{print $1}' | xargs docker rm -f
# Or
npm run server:all:stop
```

### Plugin Not Loading

**Symptom:** "Plugin not found"

**Fix:**

```bash
# Rebuild
npm run build
mage -v buildAll

# Check dist exists
ls -la dist/

# Verify mount
docker inspect <container> | grep -A5 Mounts
```

## Test Best Practices

1. **Use helpers** - Import `OpcuaTestHelpers` to reduce duplication
2. **Version compatibility** - Use `.or()` for fallback selectors
3. **Meaningful waits** - Prefer `waitForLoadState()` over `waitForTimeout()`
4. **Soft assertions** - Use `.catch(() => false)` for optional elements
5. **Debug artifacts** - Screenshots/videos saved in `test-results/`
6. **Clear test names** - Describe what's being validated

## Performance

| Metric          | Time                     |
| --------------- | ------------------------ |
| Single version  | ~2 minutes (46 tests)    |
| All versions    | ~10 minutes (5 versions) |
| Auth setup      | ~4 seconds               |
| Individual test | ~3-10 seconds            |

**Optimization:**

- Use `CI=true` in scripts
- Run tests in parallel
- Cache Docker images
- Skip optional waits in CI

## CI/CD Integration

**File:** `.github/workflows/ci.yml`

**Triggers:**

- Pull requests
- Push to main
- Version tags

**Matrix:**

- Tests run against all 5 Grafana versions in parallel
- Artifacts (screenshots, videos) saved on failure

## Contributing

When adding tests:

1. Use appropriate test file (smoke, config, query, etc.)
2. Reuse helpers from `helpers.ts`
3. Test all versions: `npm run e2e:all`
4. Add version fallbacks for UI changes
5. Update this README
6. Handle errors gracefully

## Resources

- [Playwright Docs](https://playwright.dev)
- [Grafana Plugin E2E](https://github.com/grafana/plugin-tools/tree/main/packages/plugin-e2e)
- [Plugin README](../README.md)

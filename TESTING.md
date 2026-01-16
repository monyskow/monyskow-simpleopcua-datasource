# Testing Strategy for Simple OPC-UA Plugin

This document outlines the comprehensive testing strategy for ensuring plugin compatibility across all supported Grafana versions.

## Automated Multi-Version Testing

### How It Works

The plugin is automatically tested against multiple Grafana versions on every push and pull request:

1. **Version Detection** - GitHub Actions automatically determines which Grafana versions to test based on `grafanaDependency: ">=10.4.0"` in [src/plugin.json](src/plugin.json)

2. **Matrix Testing** - Tests run in parallel across:

   - **Minimum Version**: Grafana 10.4.0
   - **LTS Version**: Current long-term support release
   - **Latest Version**: Current stable release

3. **Automatic Execution** - Tests run without manual intervention on:
   - Every push to main/master
   - Every pull request
   - Manual workflow dispatch

### CI/CD Workflow

Located in [.github/workflows/ci.yml](.github/workflows/ci.yml):

```yaml
resolve-versions:
  # Resolves Grafana versions to test
  uses: grafana/plugin-actions/e2e-version

playwright-tests:
  # Runs tests in parallel against each version
  strategy:
    matrix:
      GRAFANA_IMAGE: ${{fromJson(needs.resolve-versions.outputs.matrix)}}
```

## Test Coverage

### 48 End-to-End Tests Across 5 Suites

#### 1. Smoke Tests (11 tests)

- Plugin loads successfully
- Provisioned data source exists
- Configuration page opens
- Query editor loads
- Basic navigation works
- Multi-version compatibility

#### 2. Data Source Configuration (9 tests)

- Configuration UI renders
- Endpoint URL field works
- Security policy/mode selectors
- Authentication methods (anonymous, username/password, certificate)
- Connection testing (Save & Test)
- Configuration persistence

#### 3. Query Editor (10 tests)

- Query editor displays
- Add/remove nodes
- Set node aliases
- Node browser opens
- Browse OPC-UA tree
- Query persistence

#### 4. Data Queries & Visualization (8 tests)

- Execute queries
- Display results (table/chart)
- Multiple nodes in query
- Refresh data
- Error handling
- Alias display
- Dashboard integration

#### 5. Plugin Metadata (10 tests)

- Plugin information correct
- Installed status
- Description and metadata
- Searchable and discoverable
- Backend indicator
- Alerting support
- No console errors

### Additional Testing

- **Unit Tests**: Frontend (Jest) and Backend (Go)
- **Type Checking**: TypeScript validation
- **Linting**: ESLint code quality
- **Build Validation**: Multi-platform builds

## Running Tests Locally

### Quick Start

```bash
# 1. Install dependencies (first time only)
npm install
npx playwright install chromium --with-deps

# 2. Build plugin
npm run build
mage buildAll

# 3. Start Grafana (Terminal 1)
npm run server

# 4. Run tests (Terminal 2)
npm run e2e
```

### Test Specific Grafana Version

```bash
# Test minimum supported version
GRAFANA_VERSION=10.4.0 npm run server
npm run e2e

# Test latest version
GRAFANA_VERSION=12.0.0 npm run server
npm run e2e

# Test LTS version
GRAFANA_VERSION=11.4.0 npm run server
npm run e2e
```

### Useful Commands

```bash
# Run only smoke tests
npx playwright test smoke.spec.ts

# Run in headed mode (see browser)
npx playwright test --headed

# Debug mode
npx playwright test --debug

# UI mode (interactive)
npx playwright test --ui

# View report
npx playwright show-report
```

## Test Results in CI

### Where to Find Results

1. **GitHub Actions** - View test runs in the "Actions" tab
2. **PR Checks** - Test status shown on pull requests
3. **Test Reports** - Published to GitHub Pages (if configured)

### What Gets Reported

- ✅ Pass/Fail status for each Grafana version
- 📊 Test execution time
- 📸 Screenshots on failure
- 🎬 Video recordings on failure
- 📋 Detailed trace files for debugging

## Ensuring Compatibility

### Before Release

1. **All CI tests must pass** - No merging with failing tests
2. **Test locally** - Verify against common versions
3. **Review test report** - Check for flaky tests
4. **Update compatibility** - If dropping version support, update plugin.json

### Adding New Features

When adding features:

1. Add E2E tests in appropriate suite
2. Use test helpers from `tests/helpers.ts`
3. Test against multiple versions locally
4. Ensure CI passes before merge

### Example: Adding New Test

```typescript
// tests/my-feature.spec.ts
import { test, expect } from '@playwright/test';
import { OpcuaTestHelpers } from './helpers';

test('my new feature works', async ({ page }) => {
  const helpers = new OpcuaTestHelpers(page);
  await helpers.goToExplore();

  // Your test code here

  expect(true).toBeTruthy();
});
```

## Troubleshooting

### Tests Fail in CI But Pass Locally

- Check Grafana version (CI may test older versions)
- Review CI logs for specific errors
- Test locally with same Grafana version as CI

### Tests Are Flaky

- Increase timeouts in test
- Add explicit waits for elements
- Check for race conditions
- Review network requests

### New Grafana Version Released

- CI automatically picks up new versions
- Monitor first test run with new version
- Update tests if API changes detected

## Maintenance

### Regular Tasks

- ✅ Monitor CI test results
- ✅ Fix flaky tests promptly
- ✅ Update test data as needed
- ✅ Add tests for new features
- ✅ Review and refactor test code

### When to Update Tests

- New feature added → Add tests
- Bug fixed → Add regression test
- Grafana API changes → Update selectors
- New authentication method → Add config test
- New visualization type → Add query test

## Benefits

### Automated Version Testing Provides

✅ **Confidence** - Know plugin works across versions
✅ **Early Detection** - Catch breaking changes immediately
✅ **Documentation** - Tests serve as usage examples
✅ **Regression Prevention** - Prevent old bugs from returning
✅ **Faster Development** - Catch issues before manual testing
✅ **Release Quality** - Ship with confidence

## Summary

The Simple OPC-UA plugin has comprehensive automated testing that ensures:

- ✅ Plugin works on all supported Grafana versions (>=10.4.0)
- ✅ Core functionality verified on every commit
- ✅ Breaking changes detected immediately
- ✅ Quality maintained across releases
- ✅ Documentation for developers and users

**Result**: High confidence in multi-version compatibility without manual testing effort.

---

For detailed test documentation, see:

- [tests/README.md](tests/README.md) - Comprehensive test docs
- [tests/QUICKSTART.md](tests/QUICKSTART.md) - Quick start guide
- [.github/workflows/ci.yml](.github/workflows/ci.yml) - CI configuration

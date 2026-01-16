# E2E Tests Quick Start

## First Time Setup

1. **Install dependencies** (if not already done):

```bash
npm install
```

2. **Install Playwright browsers**:

```bash
npx playwright install chromium --with-deps
```

## Running Tests Locally

### Option 1: Automatic (Recommended for first time)

This will build the plugin, start Grafana, and run tests:

```bash
# Build plugin
npm run build
mage buildAll

# Start Grafana (in one terminal)
npm run server

# Wait for Grafana to start (30-60 seconds), then in another terminal:
npm run e2e
```

### Option 2: Quick Test Run (If already running)

If Grafana is already running from previous session:

```bash
npm run e2e
```

## Test Against Specific Grafana Version

```bash
# Terminal 1: Start specific Grafana version
GRAFANA_VERSION=10.4.0 npm run server

# Terminal 2: Run tests
npm run e2e

# Test other versions
GRAFANA_VERSION=11.0.0 npm run server  # then run e2e
GRAFANA_VERSION=11.4.0 npm run server  # then run e2e
GRAFANA_VERSION=12.0.0 npm run server  # then run e2e
```

## Useful Commands

```bash
# Run only smoke tests (fastest)
npx playwright test smoke.spec.ts

# Run specific test file
npx playwright test datasource-config.spec.ts

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run tests in debug mode (step through)
npx playwright test --debug

# Run tests in UI mode (interactive)
npx playwright test --ui

# Run single test by name
npx playwright test -g "should load plugin successfully"

# View last test report
npx playwright show-report
```

## Troubleshooting

### Tests failing with connection errors

**Problem**: Tests can't connect to Grafana

**Solution**:

1. Make sure Grafana is running: `docker ps`
2. Check Grafana is accessible: `curl http://localhost:3000`
3. Wait longer for Grafana to start (check logs: `docker logs monyskow-simpleopcua-datasource`)

### Plugin not loading

**Problem**: Plugin files not found

**Solution**:

```bash
# Rebuild plugin
npm run build
mage buildAll

# Restart Grafana
docker compose down
npm run server
```

### Auth errors

**Problem**: Tests can't authenticate

**Solution**:

```bash
# Remove auth cache
rm -rf playwright/.auth/

# Rerun tests (will re-authenticate)
npm run e2e
```

### Grafana version mismatch

**Problem**: Tests pass locally but fail in CI

**Solution**:
Test against multiple versions locally:

```bash
for version in 10.4.0 11.0.0 11.4.0 12.0.0; do
  echo "Testing Grafana $version"
  GRAFANA_VERSION=$version npm run server &
  sleep 60  # Wait for startup
  npm run e2e
  docker compose down
done
```

## CI/CD Testing

Tests run automatically in CI against multiple Grafana versions:

- Check `.github/workflows/ci.yml`
- View test results in GitHub Actions
- Matrix strategy tests: min version (10.4.0), LTS, and latest

## What Gets Tested

✅ Plugin loads successfully
✅ Configuration UI works
✅ Query editor works
✅ Node browser works
✅ Data queries execute
✅ Dashboard integration
✅ Multiple Grafana versions
✅ Error handling
✅ Navigation and persistence

## Next Steps

- Add more test cases in `tests/*.spec.ts`
- Update test helpers in `tests/helpers.ts`
- Check coverage: tests should cover all user workflows
- Test real OPC-UA servers in addition to test servers

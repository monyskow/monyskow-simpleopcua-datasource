# Testing Guide

Quick setup instructions for Grafana plugin reviewers and end users to evaluate the Simple OPC-UA plugin.

## Table of Contents

- [What Do I Run?](#what-do-i-run)
- [Quick Start](#quick-start-⭐)
- [Test Environment Overview](#test-environment-overview)
- [Manual Testing Guide](#manual-testing-guide)
- [Docker Compose Options](#docker-compose-options)
- [Automated Testing](#automated-testing)
- [For Grafana Plugin Reviewers](#for-grafana-plugin-reviewers)

---

## What Do I Run?

Pick the row that matches your goal. All commands assume the plugin is already built (`npm run build && mage buildAll`).

| Scenario                                       | Command                                                                                                                                                                 | Approximate time                      |
| ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| Browse the plugin manually (single combo)      | `docker compose -f docker-compose.full.yaml up -d` (14 datasources) or `GRAFANA_VERSION=x AUTH_CONFIG=y docker compose -f docker-compose.e2e.yaml up -d` (single combo) | instant start, browse freely          |
| Single E2E run (current Grafana, default auth) | `npm run e2e`                                                                                                                                                           | ~2 min                                |
| All Grafana versions, sequential, default auth | `npm run e2e:all`                                                                                                                                                       | ~14 min (7 versions)                  |
| Full version × auth matrix locally             | `npm run e2e:matrix`                                                                                                                                                    | ~105–140 min (35 combos)              |
| All versions in parallel for manual testing    | `npm run server:all`                                                                                                                                                    | instant start, ports 3000–3006        |
| Trigger CI matrix                              | `gh workflow run e2e-matrix.yml`                                                                                                                                        | ~10 min wall-clock (35 parallel jobs) |

---

## Quick Start ⭐

Get testing in under 5 minutes.

### Prerequisites

- Docker
- Node.js 22+
- Go 1.22+

### Build and Start

```bash
# 1. Build the plugin
npm install
npm run build
mage buildAll

# 2. Start complete test environment
docker compose -f docker-compose.full.yaml up

# 3. Open Grafana
# http://localhost:3000
```

### Login Credentials

- **Grafana**: admin / admin
- **OPC-UA servers**: user1 / password1

### What's Pre-configured

- ✅ 14 data sources covering all security/auth combinations
- ✅ 5 OPC-UA test servers (containerized simulators)
- ✅ Test dashboard with sample queries
- ✅ Pre-configured connections ready to test

---

## Test Environment Overview

### OPC-UA Test Servers

Five containerized OPC-UA servers are included for comprehensive testing:

| Server                  | Port  | Security | Authentication | Purpose                           |
| ----------------------- | ----- | -------- | -------------- | --------------------------------- |
| `opcua-nosecurity`      | 50000 | None     | Anonymous only | Test unsecured connections        |
| `opcua-secure-anon`     | 50001 | Enabled  | Anonymous only | Test security with anonymous auth |
| `opcua-secure-userpass` | 50002 | Enabled  | User/Pass only | Test username/password auth       |
| `opcua-all-auth`        | 50003 | Enabled  | All methods    | Test all auth combinations        |
| `opcua-node`            | 4840  | Enabled  | All methods    | Test Aes256_Sha256_RsaPss policy  |

**Test credentials**: user1 / password1

### Pre-configured Data Sources

14 data sources test all valid security policy × security mode × authentication combinations:

| #   | Name                          | Security Policy       | Security Mode  | Auth Method | Server                |
| --- | ----------------------------- | --------------------- | -------------- | ----------- | --------------------- |
| 01  | None + Anonymous              | None                  | None           | Anonymous   | opcua-nosecurity      |
| 02  | None + User/Pass              | None                  | None           | User/Pass   | opcua-all-auth        |
| 03  | Basic256Sha256/Sign + Anon    | Basic256Sha256        | Sign           | Anonymous   | opcua-secure-anon     |
| 04  | Basic256Sha256/Sign + User    | Basic256Sha256        | Sign           | User/Pass   | opcua-secure-userpass |
| 05  | Basic256Sha256/Encrypt + Anon | Basic256Sha256        | SignAndEncrypt | Anonymous   | opcua-secure-anon     |
| 06  | Basic256Sha256/Encrypt + User | Basic256Sha256        | SignAndEncrypt | User/Pass   | opcua-secure-userpass |
| 07  | Aes128Sha256/Sign + Anon      | Aes128_Sha256_RsaOaep | Sign           | Anonymous   | opcua-secure-anon     |
| 08  | Aes128Sha256/Sign + User      | Aes128_Sha256_RsaOaep | Sign           | User/Pass   | opcua-secure-userpass |
| 09  | Aes128Sha256/Encrypt + Anon   | Aes128_Sha256_RsaOaep | SignAndEncrypt | Anonymous   | opcua-secure-anon     |
| 10  | Aes128Sha256/Encrypt + User   | Aes128_Sha256_RsaOaep | SignAndEncrypt | User/Pass   | opcua-secure-userpass |
| 11  | Aes256Sha256/Sign + Anon      | Aes256_Sha256_RsaPss  | Sign           | Anonymous   | opcua-node            |
| 12  | Aes256Sha256/Sign + User      | Aes256_Sha256_RsaPss  | Sign           | User/Pass   | opcua-node            |
| 13  | Aes256Sha256/Encrypt + Anon   | Aes256_Sha256_RsaPss  | SignAndEncrypt | Anonymous   | opcua-node            |
| 14  | Aes256Sha256/Encrypt + User   | Aes256_Sha256_RsaPss  | SignAndEncrypt | User/Pass   | opcua-node            |

**Access in Grafana**: http://localhost:3000/connections/datasources

### E2E real OPC-UA server (auth/security matrix)

The `docker-compose.e2e.yaml` file starts a single `node-opcua` server (built from `test-servers/node-opcua-server/`) on port **4840**. This is the server used by all automated Playwright tests.

**Test credentials**: user1 / password1, admin / admin123

**Supported security policies**: None, Basic256Sha256, Aes256_Sha256_RsaPss

The `AUTH_CONFIG` environment variable selects which provisioned data source configuration Grafana loads. It maps to a file in `provisioning/datasources/`:

| `AUTH_CONFIG` value  | Provisioning file                         |
| -------------------- | ----------------------------------------- |
| `anon-none`          | `datasources-e2e-anon-none.yaml`          |
| `userpass-none`      | `datasources-e2e-userpass-none.yaml`      |
| `userpass-b256-sign` | `datasources-e2e-userpass-b256-sign.yaml` |
| `cert-b256-sign`     | `datasources-e2e-cert-b256-sign.yaml`     |
| `cert-aes256-sign`   | `datasources-e2e-cert-aes256-sign.yaml`   |

Default (when `AUTH_CONFIG` is unset): `anon-none`.

#### Manual testing of a specific (version × auth) combination

`docker-compose.e2e.yaml` also accepts `GRAFANA_VERSION` (default `12.2.0`). Combine both env vars to spin up any matrix cell for browser-based exploration:

```bash
docker compose -f docker-compose.e2e.yaml down 2>/dev/null
GRAFANA_VERSION=12.4.3 AUTH_CONFIG=cert-b256-sign \
  docker compose -f docker-compose.e2e.yaml up --wait --build

# Open http://localhost:3000  (Grafana: admin / admin)
# The provisioned "OPC-UA Test Server" data source uses the selected auth config.

docker compose -f docker-compose.e2e.yaml down
```

For an automated sweep of **all** combinations locally, use `npm run e2e:matrix` (see Automated Testing below).

---

## Manual Testing Guide

### Step-by-Step Testing Workflow

#### 1. Test Data Source Configuration

```
Navigate: http://localhost:3000/connections/datasources
→ Click any pre-configured data source
→ Click "Save & Test" button
→ Verify: Green success message appears
```

**What to verify**:

- Configuration UI loads correctly
- Security policy dropdown shows: None, Basic256Sha256, Aes128_Sha256_RsaOaep, Aes256_Sha256_RsaPss
- Security mode dropdown shows: None, Sign, SignAndEncrypt
- Auth method dropdown shows: Anonymous, Username/Password, Certificate
- "Save & Test" validates connection successfully

#### 2. Test Node Browser

```
Navigate: http://localhost:3000/explore
→ Select any data source from dropdown
→ Click "Browse Nodes" button
→ Expand "Objects" folder
→ Browse the address space
```

**What to verify**:

- Node browser dialog opens
- Root folders visible (Objects, Types, Views)
- Folders expand to show child nodes
- Variable nodes show data type icons
- Click "+" icon adds node to query

#### 3. Test Query Execution

```
In Explore page (with data source selected):
→ Click "Add manual" button
→ Enter node ID: ns=2;s=Demo.Dynamic.Scalar.Float
→ Click "Run query" button
→ Verify: Data appears in graph or table
```

**What to verify**:

- Manual node entry works
- Query executes without errors
- Data visualizes correctly
- Multiple nodes can be added to single query
- Node aliases can be set

#### 4. Test Security Features

**No Security** (datasources #1-2):

- Use: "01. None + Anonymous"
- Connects without certificates or encryption

**Basic256Sha256** (datasources #3-6):

- Use: "03. Basic256Sha256/Sign + Anon"
- Plugin auto-generates client certificate
- Connection succeeds with message signing

**Aes128_Sha256_RsaOaep** (datasources #7-10):

- Use: "07. Aes128Sha256/Sign + Anon"
- Tests modern AES encryption

**Aes256_Sha256_RsaPss** (datasources #11-14):

- Use: "11. Aes256Sha256/Sign + Anon"
- Tests strongest encryption policy

#### 5. Test Authentication Methods

**Anonymous** (odd-numbered datasources):

- Use: "01. None + Anonymous"
- No credentials required

**Username/Password** (even-numbered datasources):

- Use: "02. None + User/Pass"
- Credentials: user1 / password1
- Stored securely in Grafana

#### 6. Test Dashboard Integration

```
Navigate: http://localhost:3000/dashboards
→ Click "OPC-UA Test Dashboard"
→ Verify: Panels display data from test servers
→ Edit panel to modify queries
```

**What to verify**:

- Pre-configured panels load
- Data refreshes on time interval
- Query editor works in panel edit mode
- Multiple data sources can be used in single dashboard

---

## Docker Compose Options

Choose the right environment for your testing needs:

| Configuration                | Datasources       | Dashboards | OPC-UA Servers             | Port | Use Case                                                                               |
| ---------------------------- | ----------------- | ---------- | -------------------------- | ---- | -------------------------------------------------------------------------------------- |
| `docker-compose.yaml`        | None              | None       | None                       | 3000 | Manual testing, clean slate                                                            |
| `docker-compose.e2e.yaml`    | 1 (test)          | None       | 1 (node-opcua real server) | 3000 | E2E tests (`AUTH_CONFIG` env var selects one of 5 auth configs; used by `npm run e2e`) |
| `docker-compose.full.yaml`   | 14 (docker-based) | Yes        | 5 containers               | 3000 | **Complete integration testing** ⭐                                                    |
| `docker-compose.prosys.yaml` | 14 (ProSys)       | Yes        | ProSys on host             | 3001 | ProSys simulator testing                                                               |

Before first run with `docker-compose.prosys.yaml`: `npm run server:prosys-certs` to generate the client cert + render the datasources YAML.

**Recommendation**: Use `docker-compose.full.yaml` for comprehensive plugin testing.

---

## Automated Testing

For comprehensive E2E testing documentation, see [tests/README.md](../tests/README.md).

**Test coverage**: 51 tests across plugin metadata, configuration, query editor, data queries, and cert generation.

### Single Grafana version

```bash
# Build the plugin
npm run build && mage buildAll

# Start the test environment (Grafana + node-opcua server)
docker compose -f docker-compose.e2e.yaml up -d --wait

# Run all E2E tests
npm run e2e

# Run specific test suite
npx playwright test smoke.spec.ts    # Quick validation
npx playwright test --ui             # Interactive mode
```

### All Grafana versions, sequential

```bash
npm run e2e:all
```

Backed by `scripts/e2e-all-versions.sh`. Reads `.grafana-versions` (currently 7 versions), starts each Grafana container in turn, runs the full test suite, then stops the container before moving to the next version.

**Expected runtime:** ~14 minutes total (~2 min per version)

**Output file:** `e2e-results-YYYYMMDD-HHMMSS.txt` at repo root — contains pass/fail summary for every version.

#### .grafana-versions and version management

File at repo root. 7 lines, one Grafana version per line. All entries are manually curated — edit the file directly to pin / bump versions. The convention is: lines 1–6 are LTS / known-good anchors, line 7 is the latest stable. Update line 7 manually when a new Grafana release surfaces on Docker Hub that you want to start testing against.

### Full version × auth matrix locally

```bash
npm run e2e:matrix
```

Backed by `scripts/e2e-all-auth-versions.sh`. Nested loop over all versions in `.grafana-versions` (7) × all 5 auth configs = **35 combinations**. Uses `docker-compose.e2e.yaml` (provisioned data source + real OPC-UA server) — mirrors what CI runs.

**Expected runtime:** ~105–140 minutes sequential. Use sparingly — for routine dev iteration prefer single-axis (`npm run e2e:all`) or single-combo manual testing. The CI matrix (`e2e-matrix.yml`) provides the same coverage in ~10 min wall-clock via parallel jobs.

**Output file:** `e2e-matrix-results-YYYYMMDD-HHMMSS.txt` at repo root — one line per combination, pass/fail summary at the end.

### CI: E2E matrix (e2e-matrix.yml)

Defined in `.github/workflows/e2e-matrix.yml`. Runs **7 Grafana versions × 5 auth configs = 35 jobs** in parallel.

**Auth configs:**

| `AUTH_CONFIG`        | Description                                                      |
| -------------------- | ---------------------------------------------------------------- |
| `anon-none`          | Anonymous authentication, no security policy                     |
| `userpass-none`      | Username/password authentication, no security policy             |
| `userpass-b256-sign` | Username/password with Basic256Sha256 + Sign mode                |
| `cert-b256-sign`     | Certificate authentication with Basic256Sha256 + Sign mode       |
| `cert-aes256-sign`   | Certificate authentication with Aes256_Sha256_RsaPss + Sign mode |

**Triggers:** push of a `v*` tag and `workflow_dispatch`.

**On failure:** Playwright HTML reports are uploaded as artifacts with 7-day retention.

**Summary:** The `summary` job collects per-job results and posts a pass/fail results table to `GITHUB_STEP_SUMMARY`.

Note: the CI matrix and `.grafana-versions` now use the same 7 Grafana versions; the local `e2e:all` script uses the versions in `.grafana-versions`.

---

## For Grafana Plugin Reviewers

### Validation Checklist

Use this checklist to validate plugin functionality during review:

#### Installation & Loading

- [ ] Plugin appears in plugin catalog
- [ ] Plugin marked as "Installed"
- [ ] Plugin metadata displays correctly (name, description, author)
- [ ] Backend plugin indicator visible

#### Configuration

- [ ] Data source configuration page loads
- [ ] All security policies available: None, Basic256Sha256, Aes128_Sha256_RsaOaep, Aes256_Sha256_RsaPss
- [ ] All security modes available: None, Sign, SignAndEncrypt
- [ ] All auth methods available: Anonymous, Username/Password, Certificate
- [ ] "Save & Test" validates connection
- [ ] Configuration persists after save

#### Query Editor

- [ ] Query editor loads in Explore
- [ ] "Add manual" button adds node input
- [ ] "Browse Nodes" button opens node browser
- [ ] Node browser displays address space
- [ ] Nodes can be added from browser
- [ ] Node aliases can be set
- [ ] Query configuration persists

#### Data Queries

- [ ] Queries execute successfully
- [ ] Data displays in graph/table
- [ ] Multiple nodes in single query work
- [ ] Template variables supported
- [ ] Errors handled gracefully (invalid node IDs)

#### Security & Authentication

- [ ] No security mode works (anonymous)
- [ ] Sign mode works (signed messages)
- [ ] SignAndEncrypt mode works (encrypted)
- [ ] Anonymous authentication works
- [ ] Username/password authentication works
- [ ] Client certificates auto-generated for secure connections

#### Dashboard Integration

- [ ] Queries work in dashboard panels
- [ ] Multiple data sources in single dashboard
- [ ] Data refresh works
- [ ] Panel edit mode functional

#### Health & Error Handling

- [ ] Health check endpoint responds
- [ ] Invalid endpoint shows error
- [ ] Invalid credentials show error
- [ ] Invalid node IDs handled gracefully
- [ ] Connection timeout handled

### Screenshots Reference

Example screenshots available in `src/img/`:

- `screenshot-config.png` - Configuration UI
- `screenshot-query-editor.png` - Query editor
- `screenshot-browser.png` - Node browser

### Automated Validation

Run E2E tests for automated validation:

```bash
npm run build && mage buildAll
docker compose -f docker-compose.full.yaml up -d
npm run e2e
```

Expected: All 51 tests pass across smoke, configuration, query editor, data query, and cert generation suites.

---

## Need Help?

- [README.md](../README.md) - Full documentation
- [tests/README.md](../tests/README.md) - E2E testing details
- [GitHub Issues](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues) - Report issues

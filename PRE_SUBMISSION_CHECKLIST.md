# Pre-Submission Checklist for Grafana Labs

This checklist covers all requirements before submitting the Simple OPC-UA plugin to Grafana Labs for verification and publication in the Plugin Catalog.

## Table of Contents

- [Code Quality](#code-quality)
- [Documentation](#documentation)
- [Assets and Media](#assets-and-media)
- [Metadata](#metadata)
- [Build and Distribution](#build-and-distribution)
- [Testing](#testing)
- [Security](#security)
- [Legal and Licensing](#legal-and-licensing)
- [CI/CD](#cicd)
- [Final Pre-Submission Steps](#final-pre-submission-steps)

---

## Code Quality

### Linting and Code Standards

- [x] **ESLint passes without errors** - `npm run lint`

  - Status: 0 errors, 4 warnings (deprecation warnings only)
  - ⚠️ **Action Required**: Fix deprecation warnings before submission:
    - Replace deprecated `Select` component with `Combobox` in [ConfigEditor.tsx:127,141,173](src/components/ConfigEditor/ConfigEditor.tsx#L127)
    - Update deprecated `DataQuery` type in [types.ts:57](src/types.ts#L57)

- [x] **TypeScript compilation** - `npm run typecheck`

  - Status: ✅ No type errors

- [ ] **Code formatting** - Verify Prettier formatting
  ```bash
  npx prettier --check .
  ```

### Backend Code Quality

- [x] **Go code builds successfully** - `mage buildAll`

  - Status: ✅ Built for 6 platforms (darwin/linux/windows, amd64/arm64/arm)

- [ ] **Go linting** - Run `golangci-lint`

  ```bash
  golangci-lint run ./...
  ```

- [ ] **Go formatting** - Check with `gofmt`
  ```bash
  gofmt -l pkg/
  ```

---

## Documentation

### README.md

- [x] **Clear project description** - Explains what the plugin does
- [x] **Installation instructions** - From catalog and manual
- [x] **Configuration guide** - Connection settings and auth methods
- [x] **Usage examples** - Query building and template variables
- [x] **Development setup** - Prerequisites and build commands
- [x] **Testing instructions** - Unit and E2E tests
- [x] **License information** - Apache 2.0 referenced

### CHANGELOG.md

- [x] **Follows Keep a Changelog format** - Uses standard format
- [x] **Version 1.0.0 documented** - Initial release documented
- [ ] **Future releases prepared** - Create placeholders for upcoming versions

### Additional Documentation

- [x] **E2E test documentation** - Comprehensive [tests/README.md](tests/README.md)
- [x] **CLAUDE.md** - Development guidance for AI assistants
- [ ] **CONTRIBUTING.md** - Guidelines for contributors (recommended)
  - How to report bugs
  - How to request features
  - Development workflow
  - Pull request process

---

## Assets and Media

### Screenshots

- [x] **Minimum 3-6 screenshots** - 4 screenshots present in `src/img/`

  - [x] screenshot-1.png (292KB) - Dashboard with OPC-UA Data
  - [x] screenshot-2.png (252KB) - Query Editor
  - [x] screenshot-3.png (595KB) - Node Browser
  - [x] screenshot-4.png (207KB) - Data Source Configuration

- [ ] **Screenshot quality check**
  - Verify resolution: 1920x1080 or 1280x720 recommended
  - Check clarity and readability
  - Verify no sensitive information (credentials, private IPs)
  - Ensure consistent Grafana theme across screenshots

### Logo

- [x] **Logo file present** - `src/img/logo.svg` (597 bytes)
- [ ] **Logo quality check**
  - SVG format with transparent background
  - Works at both small (40x40) and large (200x200) sizes
  - Looks good in light and dark themes
  - No trademarked elements without permission

---

## Metadata

### plugin.json

- [x] **Required fields present**

  - [x] `type: "datasource"`
  - [x] `name: "Simple OPC-UA"`
  - [x] `id: "monyskow-simpleopcua-datasource"`
  - [x] `info.description`
  - [x] `info.author` (name, email, url)
  - [x] `info.keywords` (9 keywords)
  - [x] `info.logos` (small, large)
  - [x] `info.links` (GitHub, Documentation, License)
  - [x] `info.screenshots` (4 screenshots)
  - [x] `dependencies.grafanaDependency: ">=10.4.0"`

- [x] **Backend configuration**

  - [x] `backend: true`
  - [x] `executable: "gpx_simpleopcua"`
  - [x] `alerting: true`

- [x] **Category** - `category: "iot"` (appropriate)

- [ ] **Version and update fields** - Uses placeholders `%VERSION%` and `%TODAY%`
  - Verify build process replaces these correctly

### Naming Conventions

- [x] **Plugin ID format** - `{author}-{pluginname}-{plugintype}` ✅ `monyskow-simpleopcua-datasource`
- [x] **Author namespace** - `monyskow` (GitHub username)

---

## Build and Distribution

### Build Artifacts

- [x] **dist/ folder structure**

  ```
  dist/
  ├── CHANGELOG.md
  ├── LICENSE
  ├── README.md
  ├── plugin.json
  ├── module.js
  ├── module.js.map
  ├── gpx_simpleopcua_darwin_amd64
  ├── gpx_simpleopcua_darwin_arm64
  ├── gpx_simpleopcua_linux_amd64
  ├── gpx_simpleopcua_linux_arm
  ├── gpx_simpleopcua_linux_arm64
  ├── gpx_simpleopcua_windows_amd64.exe
  ├── go_plugin_build_manifest
  └── img/
      ├── logo.svg
      └── screenshot-*.png
  ```

- [x] **Backend binaries executable** - All 6 platform binaries present
- [x] **Frontend bundle** - module.js present with source map

### Plugin Signing

- [ ] **Sign the plugin** - Required for public plugins

  ```bash
  npm run sign
  ```

  - Need `GRAFANA_ACCESS_POLICY_TOKEN` environment variable
  - Creates `MANIFEST.txt` and signature files
  - See: https://grafana.com/docs/grafana/latest/developers/plugins/sign-a-plugin/

- [ ] **Verify signature**
  ```bash
  npx @grafana/sign-plugin@latest --verify dist/
  ```

### Plugin Validation

- [ ] **Run Grafana plugin validator**
  ```bash
  docker run --pull=always \
    -v $PWD/dist:/plugin \
    grafana/plugin-validator-cli -config /plugin
  ```
  - Checks plugin.json metadata
  - Validates file structure
  - Checks for common issues

---

## Testing

### Unit Tests

- [x] **Frontend unit tests pass** - `npm run test:ci`

  - Status: ✅ Passes with no tests (uses `--passWithNoTests`)

- [x] **Backend unit tests pass** - `mage test`

  - Status: ✅ Go tests pass

- [ ] **Code coverage** - Check coverage reports
  ```bash
  npm run test:coverage
  mage coverage
  ```
  - Aim for >60% coverage minimum
  - > 80% coverage recommended

### End-to-End Tests

- [x] **E2E tests implemented** - 46 tests across 5 test suites
- [x] **Tests pass on all supported Grafana versions** - `npm run e2e:all`
  - [x] Grafana 10.4.19 ✅
  - [x] Grafana 11.1.13 ✅
  - [x] Grafana 11.4.8 ✅
  - [x] Grafana 12.1.5 ✅
  - [x] Grafana 12.3.1 ✅

### Manual Testing Checklist

- [ ] **Install plugin in Grafana** - Verify installation process
- [ ] **Configure data source** - Test all auth methods
  - [ ] Anonymous authentication
  - [ ] Username/Password authentication
  - [ ] Certificate authentication
- [ ] **Create queries** - Test query editor functionality
  - [ ] Add nodes manually
  - [ ] Use node browser
  - [ ] Set aliases
  - [ ] Use template variables
- [ ] **Visualize data** - Create dashboards and panels
  - [ ] Table visualization
  - [ ] Graph visualization
  - [ ] Multiple nodes in single query
- [ ] **Test health check** - Save & Test functionality
- [ ] **Test error handling** - Invalid node IDs, connection errors
- [ ] **Test alerting** - If alerting is supported

---

## Security

### Code Security

- [ ] **No hardcoded credentials** - Search codebase

  ```bash
  grep -r "password\|secret\|token\|key" --include="*.ts" --include="*.tsx" --include="*.go" | grep -v "test"
  ```

- [x] **Secure credential storage** - Uses Grafana's `secureJsonData`

  - [x] Passwords stored in `secureJsonData`
  - [x] Certificates stored in `secureJsonData`

- [ ] **Input validation** - Check for:

  - [ ] XSS vulnerabilities in frontend
  - [ ] SQL injection (if using databases)
  - [ ] Command injection (if executing shell commands)
  - [ ] Path traversal vulnerabilities

- [ ] **Dependency vulnerabilities** - Run security audits
  ```bash
  npm audit
  go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth
  ```

### HTTPS and Encryption

- [x] **Supports encrypted OPC-UA connections** - Security policies implemented
  - [x] None
  - [x] Basic256Sha256
  - [x] Security modes: None, Sign, SignAndEncrypt

---

## Legal and Licensing

### License

- [x] **LICENSE file present** - Apache License 2.0
- [x] **Copyright year correct** - Copyright 2024 monyskow
- [x] **License referenced in package.json** - `"license": "Apache-2.0"`

### Third-Party Dependencies

- [ ] **Review all dependencies** - Check licenses of npm and Go modules

  ```bash
  npx license-checker --summary
  go-licenses report ./... --template licenses.tpl
  ```

- [ ] **Ensure compatible licenses** - Apache 2.0 compatible:

  - ✅ MIT, BSD, ISC, Apache 2.0
  - ❌ GPL (without explicit exception)

- [ ] **Attribution for included assets** - If using third-party code/assets
  - Add NOTICE file if required
  - Credit original authors

---

## CI/CD

### GitHub Actions

- [x] **CI workflow present** - `.github/workflows/ci.yml`
- [x] **CI runs on PR and push** - Configured for main/master branches

- [ ] **CI workflow validation**
  - [x] Frontend build ✅
  - [x] Lint ✅
  - [x] TypeScript check ✅
  - [x] Unit tests ✅
  - [x] Backend build ✅
  - [x] Backend tests ✅
  - [x] E2E tests (matrix) ✅
  - [x] Plugin metadata validation ✅
  - [ ] Plugin signing (needs token)

### Release Workflow

- [ ] **Create release workflow** - Automate releases

  - Trigger on version tags (e.g., `v1.0.0`)
  - Build plugin
  - Sign plugin
  - Create GitHub release
  - Upload signed ZIP

- [ ] **Version tag format** - Use semantic versioning
  - Format: `v1.0.0`, `v1.1.0`, `v2.0.0`
  - Tag should trigger release workflow

---

## Final Pre-Submission Steps

### 1. Fix All Warnings

- [ ] Fix ESLint deprecation warnings
  - [ ] Replace `Select` with `Combobox` in ConfigEditor.tsx
  - [ ] Update deprecated `DataQuery` type in types.ts

### 2. Code Cleanup

- [ ] Remove debug console.log statements
- [ ] Remove commented-out code
- [ ] Remove unused imports
- [ ] Remove TODOs or create GitHub issues for them

### 3. Update Documentation

- [ ] Review README.md for accuracy
- [ ] Update CHANGELOG.md with complete feature list
- [ ] Add CONTRIBUTING.md if accepting contributions

### 4. Final Build

- [ ] Clean build directories

  ```bash
  rm -rf dist/ node_modules/
  npm ci
  npm run build
  mage buildAll
  ```

- [ ] Verify dist/ contents
- [ ] Test plugin locally in Grafana
  ```bash
  npm run server
  # Access http://localhost:3000, test plugin
  ```

### 5. Sign Plugin

- [ ] Obtain Grafana Access Policy Token

  - Sign up at https://grafana.com/
  - Create access policy token with plugin signing permissions

- [ ] Sign the plugin

  ```bash
  export GRAFANA_ACCESS_POLICY_TOKEN="your-token"
  npm run sign
  ```

- [ ] Verify signature
  ```bash
  npx @grafana/sign-plugin@latest --verify dist/
  ```

### 6. Run Validator

- [ ] Run Grafana plugin validator

  ```bash
  docker run --pull=always \
    -v $PWD/dist:/plugin \
    grafana/plugin-validator-cli -config /plugin
  ```

- [ ] Fix any issues reported by validator

### 7. Create GitHub Release

- [ ] Create git tag

  ```bash
  git tag -a v1.0.0 -m "Release v1.0.0"
  git push origin v1.0.0
  ```

- [ ] Create GitHub release from tag
  - Include CHANGELOG.md content in release notes
  - Upload signed ZIP file: `monyskow-simpleopcua-datasource-1.0.0.zip`

### 8. Submit to Grafana Labs

- [ ] Go to https://grafana.com/plugins/publish
- [ ] Fill out submission form

  - Plugin ID: `monyskow-simpleopcua-datasource`
  - GitHub repository URL
  - Release tag: `v1.0.0`
  - Plugin ZIP URL (from GitHub release)

- [ ] Wait for review
  - Grafana Labs reviews submissions within 5-10 business days
  - Monitor email for review feedback
  - Be prepared to make requested changes

---

## Post-Submission

### After Approval

- [ ] **Update README badges** - Add Grafana plugin catalog badge
- [ ] **Announce release** - Share on social media, forums, blog
- [ ] **Monitor GitHub issues** - Respond to user feedback
- [ ] **Plan next version** - Based on user feedback

---

## Resources

- [Grafana Plugin Tools](https://grafana.com/developers/plugin-tools/)
- [Plugin Publishing Guide](https://grafana.com/docs/grafana/latest/developers/plugins/publish-a-plugin/)
- [Plugin Metadata](https://grafana.com/docs/grafana/latest/developers/plugins/metadata/)
- [Plugin Signing](https://grafana.com/docs/grafana/latest/developers/plugins/sign-a-plugin/)
- [Plugin Validator](https://github.com/grafana/plugin-validator)

---

## Summary

**Completed**: 35/60 items (58%)

**Critical Items to Complete Before Submission**:

1. ✅ Fix ESLint deprecation warnings (Select → Combobox, DataQuery type)
2. ✅ Run code formatting check (`npx prettier --check .`)
3. ✅ Run Go linting (`golangci-lint run ./...`)
4. ✅ Verify screenshot quality and resolution
5. ✅ Verify logo works at small and large sizes
6. ✅ Check code coverage (aim for >60%)
7. ✅ Run security audits (`npm audit`)
8. ✅ Review dependency licenses
9. ✅ Sign the plugin (requires token)
10. ✅ Run Grafana plugin validator

**Estimated Time to Complete**: 4-6 hours

**Priority Order**:

1. Fix code warnings (30 min)
2. Sign plugin (1 hour, including token setup)
3. Run validator and fix issues (1 hour)
4. Security audit and dependency review (1 hour)
5. Documentation review and updates (1 hour)
6. Final testing and verification (1 hour)

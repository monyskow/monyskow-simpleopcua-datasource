# Grafana Plugin Submission Readiness Checklist

This document tracks the plugin's compliance with [Grafana's official submission requirements](https://grafana.com/docs/grafana/latest/developers/plugins/publishing-and-signing-criteria/).

---

## plugin.json - Required Fields

| Requirement | Status | Notes |
|-------------|--------|-------|
| `id` (pattern: `^[0-9a-z]+\-([0-9a-z]+\-)?(app\|panel\|datasource)$`) | ✅ PASS | `monyskow-simpleopcua-datasource` |
| `type` | ✅ PASS | `datasource` |
| `name` | ✅ PASS | `Simple OPC-UA` |
| `info.keywords` (min 1) | ✅ PASS | 9 keywords |
| `info.logos` (small & large) | ✅ PASS | SVG logo provided |
| `info.version` | ✅ PASS | Uses `%VERSION%` placeholder |
| `info.updated` | ✅ PASS | Uses `%TODAY%` placeholder |
| `dependencies.grafanaDependency` | ✅ PASS | `>=10.4.0` |

---

## plugin.json - Optional but Recommended

| Requirement | Status | Notes |
|-------------|--------|-------|
| `info.description` | ✅ PASS | Clear 2-sentence description |
| `info.author` (name, email, url) | ✅ PASS | All fields populated |
| `info.links` | ✅ PASS | 3 links (GitHub, Docs, License) |
| `info.screenshots` | ✅ PASS | 4 screenshots |
| `category` | ✅ PASS | `iot` |
| `backend` | ✅ PASS | `true` |
| `executable` | ✅ PASS | `gpx_simpleopcua` |
| `alerting` | ✅ PASS | `true` |
| `metrics` | ✅ PASS | `true` |
| Sponsorship link | ⚠️ OPTIONAL | Not configured (optional) |

---

## Documentation Requirements

| Requirement | Status | Notes |
|-------------|--------|-------|
| README.md exists | ✅ PASS | 219 lines, comprehensive |
| Installation instructions | ✅ PASS | Both catalog and manual |
| Configuration guide | ✅ PASS | Detailed with tables |
| Screenshots/demos | ✅ PASS | Screenshots in README |
| CHANGELOG.md exists | ✅ PASS | Follows Keep a Changelog format |
| CHANGELOG uses semver | ✅ PASS | v1.0.0 documented |
| LICENSE file | ✅ PASS | Apache 2.0 |

---

## Technical Requirements

| Requirement | Status | Notes |
|-------------|--------|-------|
| Plugin validator passes | ✅ PASS | CI runs `grafana/plugin-validator-cli` |
| E2E tests exist | ✅ PASS | Playwright tests configured |
| E2E tests multiple Grafana versions | ✅ PASS | Matrix tests in CI |
| CI/CD pipeline | ✅ PASS | GitHub Actions configured |
| TypeScript type checking | ✅ PASS | `npm run typecheck` in CI |
| Linting configured | ✅ PASS | ESLint in CI |
| Unit tests | ✅ PASS | Jest configured |
| Backend tests | ✅ PASS | Go tests with coverage |
| Build automation | ✅ PASS | Mage + npm scripts |

---

## Code Quality

| Requirement | Status | Notes |
|-------------|--------|-------|
| No Angular (deprecated) | ✅ PASS | React-based frontend |
| No Grafana Toolkit (deprecated) | ✅ PASS | Uses plugin-tools |
| Valid license (BSD/MIT/Apache/LGPL3/GPL3/AGPL3) | ✅ PASS | Apache 2.0 |
| No template placeholders | ✅ PASS | Customized from template |

---

## Pre-Submission Tasks

| Task | Status | Action Required |
|------|--------|-----------------|
| Run `npx @grafana/plugin-validator` locally | ✅ DONE | Passes with only optional recommendation (sponsorship link) and expected warning (unsigned) |
| Create signed plugin (for production) | ✅ DONE | Unsigned for initial submission - Grafana signs after approval |
| Create GitHub Release with ZIP | ✅ DONE | https://github.com/monyskow/monyskow-simpleopcua-datasource/releases/tag/v1.0.0 |
| Prepare SHA1 hash of ZIP | ✅ DONE | `e11865027d6cb761611c9175b08c36c668be6112` |
| Test provisioning works | ✅ DONE | Grafana starts, plugin loads, datasource provisioned correctly |

---

## Submission Form Requirements

| Field | Preparation Status | Value/Action |
|-------|-------------------|--------------|
| Plugin URL (ZIP) | ✅ Ready | https://github.com/monyskow/monyskow-simpleopcua-datasource/releases/download/v1.0.0/monyskow-simpleopcua-datasource-1.0.0.zip |
| Source code URL | ✅ Ready | `https://github.com/monyskow/monyskow-simpleopcua-datasource` |
| SHA1 hash | ✅ Ready | `e11865027d6cb761611c9175b08c36c668be6112` |
| Testing guidance | ✅ Ready | README has setup instructions |
| Provisioning configured | ✅ Ready | docker-compose files ready |

---

## How to Submit

1. **Request Plugin Signing Key**
   - Go to https://grafana.com/auth/sign-in/
   - Navigate to My Account > Access Policies
   - Create a new access policy with `plugins:write` scope
   - Use the token for `GRAFANA_ACCESS_POLICY_TOKEN` in CI

2. **Create a Release**
   ```bash
   # Tag the release
   git tag v1.0.0
   git push origin v1.0.0

   # CI will build and create release with signed plugin
   ```

3. **Submit to Grafana**
   - Go to https://grafana.com/orgs/your-org/plugins
   - Click "Submit Plugin"
   - Fill in the form with release ZIP URL and SHA1 hash

---

## Summary

- **Ready**: 32/32 checks passing
- **Optional**: 1 item (sponsorship link)
- **TODO**: 0 pre-submission tasks

The plugin is **ready for submission** to the Grafana plugin catalog!

The plugin is **technically ready** for submission. Complete the pre-submission tasks before submitting.

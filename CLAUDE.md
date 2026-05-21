# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Framework contract (commands, agents, model-selection rules, output style, hook contract) lives in `.claude/CLAUDE.md` and is loaded automatically. This file covers only what's specific to **this plugin**.

## What this is

A Grafana datasource plugin that connects to OPC-UA servers (industrial automation protocol — IEC 62541; common in MES/SCADA). The backend speaks OPC-UA over TCP and exposes browse/read/subscribe to the frontend, which renders configuration UI and query editors. When in doubt about OPC-UA semantics (NodeIds, namespaces, security policies, message-security modes), prefer the `gopcua/opcua` source over guessing.

## Tech stack

- **Frontend:** TypeScript 5.5, React 18, `@grafana/ui` 12.2.x, Webpack 5. Node 22.
- **Backend:** Go 1.25, `github.com/gopcua/opcua`, `github.com/grafana/grafana-plugin-sdk-go`.
- **Build orchestration:** Mage (backend) + npm scripts (frontend). E2E via Playwright.

## Build / test / lint

Non-standard commands worth remembering:

- `mage test` — Go tests (race detector on)
- `mage buildAll` — backend binaries for all platforms, copied into `dist/`
- `npm run build` — webpack frontend bundle
- `npm run typecheck` — `tsc --noEmit`
- `npm run lint:fix` — ESLint + Prettier combined
- `npm run test:ci` — Jest frontend tests
- `npm run server` — start `docker-compose.e2e.yaml` (Grafana + one test OPC-UA server)
- `npm run server:all` — 4 Grafana versions in parallel on ports 3000–3003
- `npm run e2e` — Playwright against default Grafana
- `npm run e2e:all` — sequential across the supported Grafana versions
- `npm run e2e:matrix` — full auth × Grafana version matrix (35 jobs)

`docker-compose.full.yaml` brings up 5 OPC-UA test servers for manual exploration; `docker-compose.prosys.yaml` targets the external ProSys simulator.

## Layout

- `src/` — frontend (datasource class, config + query editors, plugin.json)
- `pkg/plugin/` — backend; `datasource.go` handlers, `resources.go` HTTP endpoints, `opcua/` client/auth/browse/certs
- `tests/` — Playwright e2e (auth, config, query editor, cert generation, matrix)
- `scripts/` — multi-version bash helpers
- `.github/workflows/` — CI, weekly Grafana-version bump, release signing

## Grafana compatibility

- `plugin.json` declares `grafanaDependency: ">=10.4.0"`. The config editor uses `Select` (available since G7) and the UID-based resource API (available since G9), so 10.4 is the effective floor.
- CI matrix covers: **10.4.0, 10.4.19, 11.0.0, 11.6.14, 12.0.0, 12.4.3, 13.0.1**. Don't widen it without checking UID resource API support (G9+, already satisfied).
- A weekly workflow (`bump-grafana-latest.yml`) auto-PRs new stable versions into the matrix.

## Grafana 13 quirks (don't re-discover these)

- **UID-based resource URLs only.** G13 returns 404 on `/api/datasources/{numericId}/resources/...`. Use `/api/datasources/uid/{uid}/resources/...` (supported on G9+, so no version branching needed).
- **"What's new" splash modal** blocks every authenticated page on first load in G13.0.1+. `GF_FEATURE_TOGGLES_SPLASHSCREEN=false` does NOT disable it (`AllowSelfServe:false`). `tests/auth.setup.ts` dismisses it by clicking close with a soft 5s timeout — fine if the modal isn't there.
- **Dashboard panel selectors changed.** Use role-scoped locators, not text-matched OR-chains.

## E2E rules

- **Each test creates its own datasource** with a randomized UID (`opcua-iso-{workerIndex}-{random}`) via `tests/fixtures/isolated-datasource.ts`, and deletes it on teardown. This avoids Grafana's HTTP 409 on concurrent PATCHes. Don't share datasources across tests.
- **Playwright `workers: 2`, `retries: 1`** is deliberate. The node-opcua test server caps around 10 concurrent connections; raising workers or retries papers over real flake. If a test is flaky, fix the test — don't bump these.
- Auth setup has multiple fallback selectors (testId → placeholder → name → position) because the login form layout shifts between Grafana versions. Keep the fallback chain when editing.

## Plugin signing & release

- Releases are signed with `@grafana/sign-plugin --signatureType community`. Requires `GRAFANA_ACCESS_POLICY_TOKEN` in CI secrets.
- Locally the plugin is unsigned; Grafana must allow unsigned plugins for dev (`GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS` in the dev docker-compose files).

## Pre-PR checklist

`lint:fix` → `typecheck` → `test:ci` → `mage test` → `mage buildAll` → `npm run build` → at minimum `npm run e2e` (matrix runs in CI).

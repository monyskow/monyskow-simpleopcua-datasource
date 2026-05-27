# Plan: fix(e2e): "What's new" modal blocks tests on Grafana 13.0.1 (#44)

## Kontekst

`npm run e2e:matrix` — all 5 auth combos on Grafana 13.0.1 FAIL. Root cause is the new `SplashScreenModal` (heading: "Grafana Assistant is now available to OSS users") that overlays the page after login and blocks every Playwright interaction. The plugin works on 13.0.1; only automated tests are blocked.

Issue: https://github.com/monyskow/monyskow-simpleopcua-datasource/issues/44

## Podejście

Disable the splash screen via Grafana feature toggle env var in `docker-compose.e2e.yaml`. Researched in Grafana source:

- The modal is rendered in `AppChrome.tsx` gated by `useBooleanFlagValue('splashScreen', false)`.
- The feature toggle `splashScreen` was registered with `Expression: "true"` in `pkg/services/featuremgmt/registry.go` at `v13.0.1` (default ON for that release).
- It has since been flipped back to `Expression: "false"` on `main`, but 13.0.1 ships with it ON.
- Grafana standard env var maps `[feature_toggles] splashScreen = false` to `GF_FEATURE_TOGGLES_SPLASHSCREEN=false`.

Single env var line in the compose file — cleanest possible fix, no Playwright code change needed. The dismissal would otherwise need per-user storage write (`useUserStorage` → server-side), which is harder than overriding the flag.

**Fallback (only if env var does not work):** add a tolerant dismissal in `tests/auth.setup.ts` — click the close `IconButton` with `aria-label="Close"` inside the dialog with `aria-label="What's new in Grafana"`, wrapped in try/catch with short timeout (mirrors existing "Skip password change" pattern lines 33-40).

## Zmiany w plikach

- `docker-compose.e2e.yaml` — in the `grafana.environment` block (lines 61-66), add:
  ```yaml
  GF_FEATURE_TOGGLES_SPLASHSCREEN: false
  ```
  Place after `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS` (line 66). The value is a YAML boolean — Grafana parses env strings, both `false` and `"false"` work; keep unquoted to match the existing `GF_DATAPROXY_LOGGING: 1` style.

That is the entire change. No code, no provisioning, no Playwright edits.

## Kolejność

1. Add the env line to `docker-compose.e2e.yaml`.
2. Rebuild and run a single combo against 13.0.1 to verify:
   ```sh
   GRAFANA_VERSION=13.0.1 AUTH_CONFIG=anon-none \
     docker compose -f docker-compose.e2e.yaml up --wait --build
   npm run e2e
   ```
3. If green, run full matrix: `npm run e2e:matrix`. Expect all 13.0.1 rows to pass.
4. If still failing (modal still present): inspect `test-results/.../error-context.md`. If the dialog is gone but a different blocker appears → separate issue. If the dialog is still there → toggle name typo / wrong env-var mapping; fall back to Playwright dismissal in `tests/auth.setup.ts`.

## Test plan

- [ ] AC (b): `GF_FEATURE_TOGGLES_SPLASHSCREEN: false` present in `docker-compose.e2e.yaml`.
- [ ] AC: `npm run e2e:matrix` 13.0.1 × 5 auth combos — all pass (52/52 per combo expected, matching existing baselines).
- [ ] Regression: run matrix against 12.2.0 and 11.4.x as well — the env var is a no-op on versions where the flag does not exist, but confirm no warning spam in `grafana` container logs (`docker logs monyskow-simpleopcua-datasource | grep -i splash`).
- [ ] No code change in `tests/`, so existing test suite behavior on currently-green versions is unchanged.

## Ryzyka / open questions

- **Unknown env var mapping for hyphenated/camelCase keys.** Grafana env var rule lowercases and ignores camelCase — `splashScreen` → `GF_FEATURE_TOGGLES_SPLASHSCREEN`. Verified pattern in `defaults.ini` `[feature_toggles]` section; same convention used by other Grafana plugins. Low risk.
- **Toggle could be stripped/renamed in a future Grafana version.** If it disappears, the env var silently does nothing — no harm. Modal would also be gone in newer Grafana (already flipped to false on main).
- **If env var does not disable the modal** (unlikely but possible): fall back to dismissal block in `tests/auth.setup.ts` after the Skip-password block — target `page.getByRole('dialog', { name: "What's new in Grafana" }).getByRole('button', { name: 'Close' })` with short timeout in try/catch.
- **No ADR needed.** Single config tweak in test infra, no new dependency, no module boundary change.

# Plan: fix(plugin): ConfigEditor crashes with React error #130 on Grafana 10.4-11.4 (Combobox undefined) (#43)

## Context

`ConfigEditor.tsx` imports `Combobox` from `@grafana/ui`. `Combobox` is undefined in Grafana 10.4.x and 11.0-11.4 (stabilized only in 11.5+/12.x), so the config page throws React error #130 there. `plugin.json` claims `grafanaDependency: ">=10.4.0"` but the matrix run shows pass only on 12.1+. The declaration is dishonest.

Issue: https://github.com/monyskow/monyskow-simpleopcua-datasource/issues/43

## Approach

**Direction X — Honest declaration.** Bump `grafanaDependency` to `">=12.1.0"` and trim the test matrix to the versions empirically known to pass. This is a ~5-minute change with no code logic touched.

Rationale (solo dev pragmatism):

- `@grafana/ui` is already pinned to `^12.2.0` in build deps — the plugin is implicitly 12+.
- Direction Y (`Combobox` → `Select` rewrite + manual re-verify on three old versions) costs 30-60 min and only buys back EOL'd LTS lines (10.4 EOL'd 2025-11, 11.x EOL'd 2026).
- CLAUDE.md: "Solo dev = nobody depends on old API." No known users on 10.4/11.x. If anyone surfaces post-release, we revisit.
- Side benefit: saves ~3 versions × 5 auth configs = 15 matrix jobs in CI.

**ADR needed?** No. Bumping a min-supported version is a deps-shrink, not a new dependency, module-boundary, or API contract change. It is a breaking change for downstream users, but per solo-dev framework: nobody depends, just note in CHANGELOG when shipping.

## Files changed

- `src/plugin.json` — `grafanaDependency: ">=10.4.0"` → `">=12.1.0"`
- `.grafana-versions` — delete lines 1-3 (10.4.19, 11.1.13, 11.4.8)
- `.github/workflows/e2e-matrix.yml` — drop the three old versions from matrix (lines 21-23) and from the summary `GRAFANA_VERSIONS` string (line 117). Also align: matrix currently lists `12.4.3` (line 26) which is missing from `.grafana-versions` — pre-existing drift; out of scope for this fix, leave a note.
- `tests/README.md` — update line 18 ("10.4.x, 11.x, 12.x" → "12.x, 13.x"), lines 31-33 (drop old versions), line 257, lines 277-279, and adjust "~12 minutes total" rough estimate on line 266.
- `TESTING.md` — grep confirms no version-count references that need editing (only a `#11-14` datasource-id reference, unrelated).

## Order

1. `src/plugin.json` — honest declaration (smallest commit, the core fix).
2. `.grafana-versions` — align supported list.
3. `.github/workflows/e2e-matrix.yml` — drop matrix rows and summary versions.
4. `tests/README.md` — docs consistent with code.
5. Run `npm run e2e:matrix` (or trigger workflow) — confirm all remaining rows pass.

Each step can be its own commit, or squashed at PR time.

## Test plan

- [ ] AC: `src/plugin.json` declares `>=12.1.0` (grep verify).
- [ ] AC: `.grafana-versions` contains only 12.1.5, 12.3.1, 12.4.3, 13.0.1 (or current 12+/13 set).
- [ ] AC: `.github/workflows/e2e-matrix.yml` matrix and summary versions match `.grafana-versions`.
- [ ] AC: `tests/README.md` "Grafana Versions Tested" lists only supported versions.
- [ ] Regression / empirical: `npm run e2e:matrix` (or `workflow_dispatch` of `e2e-matrix.yml`) — every row (Grafana × auth_config) passes. No row for 10.4/11.x should appear.
- [ ] Manual smoke: open ConfigEditor on Grafana 12.1.5 — no React error #130 in console, security policy/mode/auth dropdowns render and work.
- [ ] `npm run build` succeeds (no type errors from unchanged `Combobox` import).

## Risks / open questions

- **Users on 10.4/11.x in production.** Likelihood low (plugin is young, no telemetry of installs). Mitigation: CHANGELOG entry on `/cx-ship` calling out the min-version bump as a breaking change, recommending `Combobox`-aware Grafana (11.5+/12.x). Grafana's plugin manager will refuse install on incompatible versions once `grafanaDependency` is honest — that's the desired behavior.
- **Pre-existing drift.** `.grafana-versions` (6 lines) and matrix (7 entries — adds `12.4.3`) disagree today. This plan removes 3 lines from both, leaving the drift intact. Out of scope; flag as a follow-up if it matters.
- **`Combobox` not removed from code.** Still imported and used. That is intentional under Direction X — the type signature is fine on 12+. If we later drop `>=12.1` to something `Combobox`-less again, that's a new ticket.

## Concrete diffs

### `src/plugin.json` (line 63)

```diff
   "dependencies": {
-    "grafanaDependency": ">=10.4.0",
+    "grafanaDependency": ">=12.1.0",
     "plugins": []
   }
```

### `.grafana-versions` (lines 1-3)

```diff
-10.4.19
-11.1.13
-11.4.8
 12.1.5
 12.3.1
 13.0.1
```

(Note: `12.4.3` from the GH matrix is not present here today — pre-existing drift, leave as-is.)

### `.github/workflows/e2e-matrix.yml`

Lines 20-27:

```diff
         grafana_version:
-          - '10.4.19'
-          - '11.1.13'
-          - '11.4.8'
           - '12.1.5'
           - '12.3.1'
           - '12.4.3'
           - '13.0.1'
```

Line 117:

```diff
-          GRAFANA_VERSIONS="10.4.19 11.1.13 11.4.8 12.1.5 12.3.1 12.4.3 13.0.1"
+          GRAFANA_VERSIONS="12.1.5 12.3.1 12.4.3 13.0.1"
```

### `tests/README.md`

Line 18:

```diff
-The E2E test suite validates the complete user experience of the OPC-UA plugin across multiple Grafana versions (10.4.x, 11.x, 12.x). Tests run in a real browser (Chromium) against actual Grafana instances running in Docker containers.
+The E2E test suite validates the complete user experience of the OPC-UA plugin across multiple Grafana versions (12.x, 13.x). Tests run in a real browser (Chromium) against actual Grafana instances running in Docker containers.
```

Lines 29-36:

```diff
 **Grafana Versions Tested:**

-- 10.4.19 (Grafana 10.x LTS)
-- 11.1.13 (Grafana 11.1.x)
-- 11.4.8 (Grafana 11.4.x)
 - 12.1.5 (Grafana 12.1.x)
 - 12.3.1 (Grafana 12.3.x)
 - 13.0.1 (Grafana 13.x - latest-stable)
```

Line 257:

```diff
-2. For each Grafana version (10.4.19, 11.1.13, 11.4.8, 12.1.5, 12.3.1, 13.0.1):
+2. For each Grafana version (12.1.5, 12.3.1, 13.0.1):
```

Line 266 (estimate):

```diff
-**Expected:** ~12 minutes total (2 min per version)
+**Expected:** ~6 minutes total (2 min per version)
```

Lines 271-282:

```diff
-# Start all 6 versions on different ports
+# Start all supported versions on different ports
 npm run server:all
```

**Access URLs (login: admin / admin):**

-- Grafana 10.4.19: http://localhost:3000
-- Grafana 11.1.13: http://localhost:3001
-- Grafana 11.4.8: http://localhost:3002
-- Grafana 12.1.5: http://localhost:3003
-- Grafana 12.3.1: http://localhost:3004
-- Grafana 13.0.1: http://localhost:3005
+- Grafana 12.1.5: http://localhost:3000
+- Grafana 12.3.1: http://localhost:3001
+- Grafana 13.0.1: http://localhost:3002

```

(Port-renumbering only if `npm run server:all` script actually re-maps. If it hardcodes ports per version, leave the URLs intact and just remove the 10/11 lines. Developer to verify against `package.json` / docker-compose before editing port numbers.)
```

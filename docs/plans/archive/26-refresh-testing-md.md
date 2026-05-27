# Plan: Refresh TESTING.md for recent E2E/CI features (#26)

## Context

Issue [#26](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues/26): `TESTING.md` and `tests/README.md` drifted after four features landed without doc updates: `.grafana-versions` + weekly bump workflow, `scripts/*-all-versions.sh`, real `node-opcua` container in `docker-compose.e2e.yaml`, and `e2e-matrix.yml` 7×5 matrix. Docs-only task; surface = exactly two markdown files. Authored by `tech-lead` (Opus); execution by `developer` (Sonnet).

## Approach

Edit `TESTING.md` top-down (Test Environment Overview → Docker Compose Options → Automated Testing), then patch six concrete drift items in `tests/README.md`. No new files, no code changes, no framework changes.

## File changes

### `TESTING.md` (in-place edits only)

- **§ Test Environment Overview** — add new subsection `### E2E real OPC-UA server (auth/security matrix)` after the existing "Pre-configured Data Sources" table. Content: node-opcua container on port 4840, creds `user1/password1` and `admin/admin123` (from `docker-compose.e2e.yaml` header), supported security policies; `AUTH_CONFIG` env var → `provisioning/datasources/datasources-e2e-<config>.yaml` mapping; 5-row table mapping each `AUTH_CONFIG` value to its provisioned data source file.
- **§ Docker Compose Options (table, line ~204)** — change row `docker-compose.e2e.yaml`: "OPC-UA Servers" `None` → `1 (node-opcua real server)`; "Use Case" → mention `AUTH_CONFIG` env-var-driven provisioning selecting one of 5 auth configs.
- **§ Automated Testing (lines 212–229)** — replace stale `npm run server` / "Terminal 1 / Terminal 2" snippet with `docker compose -f docker-compose.e2e.yaml up -d --wait` then `npm run e2e`. Verify the "46+ tests" claim by counting `test(` occurrences in `tests/*.spec.ts` (current ground truth ≈ 52: smoke 10, plugin-metadata 10, datasource-config 10, query-editor 9, data-queries 7, cert-generation 6); update number accordingly.
- **§ Automated Testing** — append new subsection `### Run against all Grafana versions locally`: document `npm run e2e:all` (= `scripts/e2e-all-versions.sh`), reads `.grafana-versions` (currently 6 versions), runtime ~12 min, output `e2e-results-YYYYMMDD-HHMMSS.txt`.
- **§ Automated Testing** — append new subsection `### .grafana-versions and version management`: file shape (lines 1–5 manually curated anchors, line 6 latest-stable slot); weekly `bump-grafana-latest.yml` cron `0 9 * * 1` (Mon 09:00 UTC) opens a PR; pin by editing the file directly.
- **§ Automated Testing** — append new subsection `### CI: E2E matrix (e2e-matrix.yml)`: matrix 7 Grafana versions × 5 auth configs = 35 jobs; list five auth configs with one-line descriptions (`anon-none`, `userpass-none`, `userpass-b256-sign`, `cert-b256-sign`, `cert-aes256-sign`); triggers `push: tags v*` and `workflow_dispatch`; Playwright reports uploaded on failure with 7-day retention; summary job renders a results table to `GITHUB_STEP_SUMMARY`.

### `tests/README.md` (6 audit items only — no other edits)

1. **Overview → Grafana Versions Tested** (lines 27–33): 5 → 6 entries; add `13.0.1` (latest-stable slot). Keep list sourced from `.grafana-versions`.
2. **Running Tests → All Grafana Versions (Automated)** (lines ~252–262): version list 5 → 6; expected runtime "~10 minutes" → "~12 minutes (6 versions)".
3. **Running Tests → All Grafana Versions (Manual Testing)** (lines ~271–277): URL list 5 → 6 (add `localhost:3005` for `13.0.1`).
4. **Performance table** (lines ~410–415): "All versions ~10 minutes (5 versions)" → "~12 minutes (6 versions)".
5. **CI/CD Integration** (lines ~424–437): file reference `ci.yml` → `e2e-matrix.yml`; "5 Grafana versions" → "7 Grafana versions × 5 auth configs" (matrix workflow runs the full 7, not the local 6); update triggers to match (`workflow_dispatch`, `push: tags v*`).
6. **Test Structure tree** (lines ~37–47): add `cert-generation.spec.ts` entry.

## Order

1. `TESTING.md` § Docker Compose Options table fix (smallest, isolated, low conflict risk).
2. `TESTING.md` § Test Environment Overview — add real OPC-UA subsection.
3. `TESTING.md` § Automated Testing — replace stale `npm run server` block + verify test count.
4. `TESTING.md` § Automated Testing — append three new subsections (`Run against all Grafana versions locally`, `.grafana-versions and version management`, `CI: E2E matrix`). Update Table of Contents (lines 5–12) once at the end.
5. `tests/README.md` — apply 6 audit items in a single pass.
6. Final read-through for cross-references and broken anchors.

## Test plan

No automated test suite for docs. `developer` (Sonnet) verification:

- [ ] AC group 1 (Automated Testing refresh) — snippet matches `package.json` (`e2e`, `e2e:all`) and current compose flow.
- [ ] AC group 2 (`.grafana-versions` subsection) — content matches actual file (6 lines) and `bump-grafana-latest.yml` (cron, line-6 slot semantics).
- [ ] AC group 3 (E2E matrix CI subsection) — dimensions, auth configs, triggers, artifact retention match `e2e-matrix.yml`.
- [ ] AC group 4 (Docker Compose table) — `docker-compose.e2e.yaml` row reflects `opcua-server` service + `AUTH_CONFIG` provisioning.
- [ ] AC group 5 (real OPC-UA subsection) — credentials and `AUTH_CONFIG` mapping match `docker-compose.e2e.yaml` and the five `datasources-e2e-*.yaml` files in `provisioning/datasources/`.
- [ ] AC group 6 (`tests/README.md` audit) — all 6 items addressed, nothing beyond scope touched.
- [ ] `npx markdownlint TESTING.md tests/README.md` if available locally, else `prettier --check '*.md' 'tests/*.md'`.
- [ ] Manual read-through: TOC anchors resolve; no dead links to removed sections.

## Risks / open questions

- **7 vs 6 versions discrepancy.** `e2e-matrix.yml` matrix explicitly lists 7 (adds `12.4.3`), `.grafana-versions` lists 6. Issue spec says "7 Grafana × 5 auth" for CI. Plan keeps both: local = 6 (from `.grafana-versions`), CI matrix = 7 (hardcoded in workflow). `developer` should not try to "reconcile" by editing the workflow or the file — that is out of scope.
- **Test count claim ("46+").** Recount by `grep -c "test(" tests/*.spec.ts`. Current ≈ 52. Use the recounted number; do not invent a range.
- **Runtime claim ("~12 min").** Sourced from issue; not independently measured here. If `developer` knows a better number from recent CI runs, use that; otherwise keep ~12 min and flag in PR description.
- **TOC update.** TESTING.md TOC (lines 5–12) is flat (one entry per H2). New content is H3 under existing H2 sections — no TOC change required. Confirm before adding entries.
- **No ADR needed.** This is content drift on existing docs, no new dependency, no module boundary change, no breaking API. No `/cx-architecture` gate.

---

**Plan path:** `docs/plans/26-refresh-testing-md.md`

**Summary:** Docs-only refresh of `TESTING.md` (one new subsection in Test Environment Overview, three new subsections in Automated Testing, one table-row fix, one stale-command replacement) plus six targeted edits in `tests/README.md`. Key risks: the 7-vs-6 versions split between CI workflow and local script (do not reconcile), and the unverified "46+ tests" / "~12 min" claims (recount / accept and flag).

**ADR needed?** No.

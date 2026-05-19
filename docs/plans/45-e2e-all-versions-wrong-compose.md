# Plan: e2e-all-versions.sh uses wrong compose file (#45)

## Goal

Make `scripts/e2e-all-versions.sh` (invoked by `npm run e2e:all`) drive `docker-compose.e2e.yaml` instead of the implicit default `docker-compose.yaml`, so the local "all versions" sweep actually exercises a provisioned datasource + real OPC-UA server — matching what `e2e:matrix` and CI already do.

## Acceptance criteria (from issue)

1. `scripts/e2e-all-versions.sh` invokes `docker compose -f docker-compose.e2e.yaml ...` for both `up` and `down`.
2. Default `AUTH_CONFIG` (env unset → `anon-none`) preserved — matches CI matrix default-auth row.
3. `TESTING.md` decision table: remove the inline "uses `docker-compose.yaml` (no OPC-UA server)" callout from the `npm run e2e:all` row.
4. `npm run e2e:all` against 12.1.5 and 12.3.1 → both PASS.
5. Failures on 10.4–11.4 expected and out-of-scope (#43).

## Files to change

- `scripts/e2e-all-versions.sh`
  - **Line 53** — `docker compose down --remove-orphans` → add `-f docker-compose.e2e.yaml`.
  - **Line 60** — `docker compose up -d --build` (the line that exports `GRAFANA_VERSION`/`GRAFANA_IMAGE`/`ANONYMOUS_AUTH_ENABLED`) → add `-f docker-compose.e2e.yaml`.
  - **Line 99** — final `docker compose down --remove-orphans` → add `-f docker-compose.e2e.yaml`.
  - Do **not** add `AUTH_CONFIG=...` to the env prefix — leaving it unset relies on the `${AUTH_CONFIG:-anon-none}` default in `docker-compose.e2e.yaml` line 60. Preserves AC2.
- `TESTING.md`
  - **Line 25** — strip the trailing ` — **Note: uses ... data source only**` segment from the `npm run e2e:all` row. Approximate-time cell becomes just `~12 min (6 versions)`.

No changes to `package.json` (`e2e:all` already just shells out to the script), no changes to `docker-compose.e2e.yaml` (it already accepts every env var the script exports), no changes elsewhere.

## Order of changes (smallest reasonable steps)

1. Patch the three `docker compose` invocations in `scripts/e2e-all-versions.sh` (one commit).
2. Strip the TESTING.md caveat (line 25) in the same commit — single logical fix.
3. Smoke-check by `developer` (Sonnet): run `npm run e2e:all` after editing `.grafana-versions` down to **a single working version** (e.g. `12.3.1`) so the loop finishes in ~2 min. Restore `.grafana-versions` before commit.
4. Hand off to `tester` (Sonnet) for the real two-version verification on `12.1.5` and `12.3.1` (AC4).

## How we verify

| AC  | Concrete check                                                                                                                                        |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `grep -n 'docker compose' scripts/e2e-all-versions.sh` → every match has `-f docker-compose.e2e.yaml`.                                                |
| 2   | Script does not export `AUTH_CONFIG`. `docker compose -f docker-compose.e2e.yaml config` during a run shows `datasources-e2e-anon-none.yaml` mounted. |
| 3   | `grep -n 'no OPC-UA server' TESTING.md` → no matches.                                                                                                 |
| 4   | `tester` re-runs `npm run e2e:all` with `.grafana-versions` reduced to `12.1.5` + `12.3.1` (or full file if Docker bandwidth available) → both PASS.  |
| 5   | If 10.4–11.4 fail, confirm failure surface matches #43/#44 (Combobox crash, splash modal); do not block this PR.                                      |

**Developer smoke check first.** `tester` will need each Grafana image pulled (≈800 MB × 6) and ~12 min wall-clock for the full sweep — make sure the single-version smoke passes before invoking that.

## Risks

- **(a) Env var compatibility.** Verified by reading `docker-compose.e2e.yaml`: it consumes exactly `GRAFANA_VERSION`, `GRAFANA_IMAGE`, `ANONYMOUS_AUTH_ENABLED`, plus `AUTH_CONFIG` (defaulted to `anon-none`). All three the script exports map cleanly. No additional env vars needed.
- **(b) Other callers of default `docker-compose.yaml`.** `npm run server` and `npm run e2e:matrix` already use `-f`. The bare `docker-compose.yaml` stays as the "manual / clean slate" option documented in TESTING.md line 266 — unchanged, unaffected.
- **(c) CI impact.** `grep -rn e2e:all .github/workflows` → no matches. CI uses `e2e-matrix.yml` directly, which already calls the matrix script. Zero CI surface change.
- **(d) Newly-exposed failures.** This fix is exactly the change that bug #43 (Combobox on older Grafana) reportedly slipped past. After the fix, 10.4–11.4 will likely fail in the local sweep. AC5 explicitly carves this out — surface in PR description, do not chase.
- **(e) `--wait` vs curl-loop.** `docker-compose.e2e.yaml` declares healthchecks; the script uses a curl loop + `sleep 10`. Switching to `docker compose ... up --wait` would be cleaner (sibling matrix script does this on line 51) but is out of scope here. Note for future refactor only.

## Out of scope

- Adding auth dimension to `e2e:all` (#39).
- Fixing underlying compat on older Grafana (#43, #44).
- Replacing the curl health loop with `--wait` (cosmetic, separate PR if ever).

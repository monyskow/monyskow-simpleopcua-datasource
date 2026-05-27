# Plan: fix(frontend): cert "saved" status lost after reload on provisioned datasource (#29)

## Kontekst

Regression caught by E2E `tests/cert-generation.spec.ts:206` ("should persist certificate after Save and Test") — 51/52 chromium tests pass; this one fails. After Save & Test + reload, the UI shows `No certificate configured` even though the cert was saved.

Root cause (likely): `ConfigEditor.tsx:63` derives "saved?" from `secureJsonFields.clientCert && secureJsonFields.clientKey`. For provisioned datasources (the E2E setup uses `provisioning/datasources/datasources-e2e-anon-none.yaml`), Grafana re-applies the YAML on every restart and the YAML does not declare `clientCert`/`clientKey`, so `secureJsonFields` for those keys returns absent — even after a successful API write.

Issue: [#29](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues/29). Failing test: `tests/cert-generation.spec.ts:206`.

## Podejście

Move "is client cert configured?" out of `secureJsonFields` (which provisioning controls) into `jsonData.clientCertConfigured: boolean` (which provisioning does not overwrite once written by API patch). The backend sets the flag during the `/generate-certificate` resource call by patching the datasource via Grafana's admin API; the frontend reads it on reload.

**Verify hypothesis first** — before coding, confirm via DevTools Network on `/cx-build` that:

1. `PATCH /api/datasources/uid/:uid` after Save & Test includes `secureJsonData.clientCert` / `clientKey`.
2. On reload, `GET /api/datasources/uid/:uid` returns `secureJsonFields.clientCert === false` (or missing) despite the prior PATCH. If true, hypothesis (a)/(b) confirmed; proceed with the flag approach. If (c) — payload missing the cert fields — pivot: include them in Save & Test payload (no backend change needed).

## Zmiany w plikach

- `src/types.ts` — add `clientCertConfigured?: boolean` to `OpcuaDataSourceOptions`.
- `src/components/ConfigEditor/ConfigEditor.tsx:63` — change `hasClientCertificateSaved` to OR with `jsonData.clientCertConfigured`. On `onResetClientCertificate` (line 139), also set `jsonData.clientCertConfigured = false`. On successful generate (line 121), do **not** set the flag — only the backend flips it to `true` after the secrets actually land in Grafana storage.
- `pkg/plugin/resources.go:94` (`handleGenerateCertificate`) — after generating the PEMs, call Grafana's admin API (`PUT /api/datasources/uid/:uid`) using `backend.GrafanaConfig` + plugin SA token to patch `jsonData.clientCertConfigured = true` **and** `secureJsonData.clientCert` / `clientKey`. This makes the resource handler the single source of truth for cert persistence and removes the dependency on user clicking Save & Test for the secret to land.
  - Alternative if admin-API approach is messy: keep current "return PEMs to frontend, frontend stores them in secureJsonData" flow, but additionally have frontend write `jsonData.clientCertConfigured = true` alongside the secrets, and have the backend (in `CheckHealth` or first query) backfill the flag on legacy datasources where `len(s.ClientCert) > 0 && !s.ClientCertConfigured`. Decide during step 2 below.
- `pkg/plugin/models/settings.go:20` — add `ClientCertConfigured bool \`json:"clientCertConfigured"\`` (read from jsonData).

## Kolejność

1. **Verify** — run `npx playwright test tests/cert-generation.spec.ts --headed --debug` (or one-shot with PWDEBUG) and inspect Network: confirm PATCH payload + reloaded GET response shape. ~15 min, gates the whole plan.
2. **Pick path** — if hypothesis confirmed: backend-driven flag (preferred, no race with user clicking Save). If admin-API call from inside CallResource turns out to need extra config plumbing, fall back to frontend-driven flag with backend backfill.
3. **Backend** — `models/settings.go` field, `resources.go` flag-write (or backfill in `datasource.go` if going the frontend-driven route). Add unit test in `resources_test.go` for the patch shape.
4. **Frontend** — `types.ts` field, `ConfigEditor.tsx` OR-check + reset clears flag. No other UI changes.
5. **Verify** — re-run failing E2E. Run the full chromium suite to confirm no regression in the other 51.

## Test plan

- [ ] AC: failing `tests/cert-generation.spec.ts:206` now passes (regression guard).
- [ ] AC: cert section after reload on provisioned DS shows `Certificate configured (saved)`.
- [ ] Manual: non-provisioned datasource (created via UI) — generate, save, reload — still shows saved. Add a Playwright variant only if hypothesis-verification step shows the non-provisioned flow also touches the same code path differently.
- [ ] Go unit: `resources_test.go` — `handleGenerateCertificate` writes flag (mock admin API or extract the patch builder into a pure function and test that).
- [ ] Backfill (if frontend-driven path chosen): test that an existing DS with `secureJsonFields.clientCert === true` and missing `clientCertConfigured` is treated as saved on next read. Solo dev = no real users yet, so backfill can be one read-time normalization in `ParseSettings` rather than a migration.

## Ryzyka / open questions

- **Admin API from CallResource** — calling `PUT /api/datasources` from within a plugin resource handler requires a service-account token; Grafana 10+ exposes one via `backend.GrafanaConfig.PluginAppClientSecret` or similar. If not trivially available, fall back to frontend-driven flag (step 2 alternative).
- **Stale flag** — flag says `true` but secret was manually deleted from DB. Out of scope; surfaces as a connection error on test, not a UI lie users will act on.
- **Non-provisioned flow** — provisioning only overwrites fields it declares, so non-provisioned DSes should already work with the current `secureJsonFields` check. Verify explicitly in step 1 — the OR-fallback to `jsonData.clientCertConfigured` is harmless in either case.
- **Backwards compat** — solo dev, nobody depends on the old shape. If we pick the frontend-driven path, do one read-time normalization in `ParseSettings` (`if !ClientCertConfigured && len(ClientCert) > 0 { ClientCertConfigured = true }`) and don't ship a migration.
- **ADR needed?** No. Adding one boolean to `jsonData` is not ADR-class. If the verification step reveals we need a new dep (e.g. a Grafana admin client lib) or want to redesign cert storage entirely (e.g. move PEMs to disk), STOP and run `/cx-architecture`.

---

## Summary

- **Path:** `docs/plans/29-cert-saved-status-lost-on-reload.md`
- **Plan:** Introduce `jsonData.clientCertConfigured: boolean` as the persistence-of-cert signal (provisioning doesn't clobber `jsonData` fields it didn't declare, unlike `secureJsonFields`). Backend preferably writes the flag during cert generation; frontend reads it alongside the existing `secureJsonFields` check. Verify hypothesis via DevTools Network before committing to the path.
- **Key risks:** (1) admin-API-from-CallResource may be plumbing-heavy → frontend-driven flag is the fallback. (2) hypothesis (c) — cert never reaches PATCH payload at all — flips the fix entirely; the verification step in `Kolejność:1` exists to catch this before any code is written.
- **ADR needed?** No — single boolean addition. Re-evaluate if scope expands to new dep or cert storage redesign.

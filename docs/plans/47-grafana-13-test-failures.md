# Plan: Fix remaining Grafana 13.0.1 E2E test failures (#47)

## Goal

Make `should work in dashboard panel` and `should persist certificate after Save and Test` pass on Grafana 13.0.1 while keeping G12.x green.

## Diagnosis

### Test A: `tests/data-queries.spec.ts:237-255` — `should work in dashboard panel`

- G13 dashboard editor renders a side panel with `button "Panel"` (img alt: `"Add new panel button"`) and a paragraph `"Drag or click to add a panel"`.
- Current selector `getByRole('button', { name: /add.*visualization/i }).or(getByText(/add.*panel/i))` matches the **paragraph**, not the button. `.click()` on it is a no-op, query editor never loads, line 253 times out.

### Test B: `tests/cert-generation.spec.ts:187-236` — `should persist certificate after Save and Test`

- Resource call `GET /api/datasources/{id}/resources/generate-certificate` returns **HTTP 404 `{"message":"Not found"}`** on G13.0.1 only. Frontend catches the error in `src/components/ConfigEditor/ConfigEditor.tsx:113-148` and shows `Certificate Generation Error: Failed to generate certificate`. Status stays `No certificate configured`. Test never reaches Save.
- **Root cause:** Grafana 13 removed the legacy numeric-ID resource path. Modern path is `/api/datasources/uid/<uid>/resources/...`. The plugin frontend at `ConfigEditor.tsx:122` builds the URL with `options.id` (numeric) — works on G12.x, returns 404 on G13.
- Sibling test at `tests/cert-generation.spec.ts:148` tolerates both success and error so it stays green, masking the regression.

## Fix strategy

### Test A (low risk, frontend test only)

- File: `tests/data-queries.spec.ts` around lines 239-244.
- Extend the OR chain with a G13-matching role-based locator so the **button** is clicked, not the paragraph:
  - Add `getByRole('button', { name: /add new panel/i })` (matches the img alt on G13).
  - Keep existing G12 locators first; G13 fallback last.
- Optionally tighten the text fallback to a button-scoped locator to avoid matching the paragraph again on future versions.

### Test B (frontend-only fix, low risk)

- File: `src/components/ConfigEditor/ConfigEditor.tsx:122`.
- Replace numeric-ID URL with UID-based URL:
  - Before: `url: \`/api/datasources/${options.id}/resources/generate-certificate\``
  - After: `url: \`/api/datasources/uid/${options.uid}/resources/generate-certificate\``
- Grep for other call sites using the same numeric-ID resource path pattern — fix all of them.
- The UID-based path has been supported in Grafana for years, so this fix works on G12.x too (no version branch needed).
- Update unit test `src/components/ConfigEditor/ConfigEditor.test.tsx:284` to match the new URL.
- Keep persist test assertions strict — do not relax to accept the error branch (sibling test #148 currently masks this regression; consider tightening it as a follow-up).

## Order

1. Fix Test A first (selector change, isolated to one file, easy verify).
2. Then Test B: swap URL to UID path -> update unit test -> rerun matrix.

Rationale: both are now localized one-file changes. Test A first because it's a test-only diff with no production-code impact.

## Verification

- Single test, headed, G13.0.1 only (fastest loop):
  - `GRAFANA_VERSION=13.0.1 npx playwright test tests/data-queries.spec.ts -g "should work in dashboard panel" --headed`
  - `GRAFANA_VERSION=13.0.1 npx playwright test tests/cert-generation.spec.ts -g "should persist certificate" --headed`
  - (Adjust env var name to whatever `scripts/e2e-all-versions.sh` exports; check the script if unsure.)
- Single combo via matrix runner: `scripts/e2e-all-versions.sh 13.0.1` (or the documented single-combo invocation).
- Full sweep before PR: `npm run e2e:matrix` — must pass on all currently-supported versions (>=12.1.0 per #43/#46).

## Risks / open questions

- Test A selector: `/add new panel/i` could match more than one element on some G12.x layouts. Mitigate by keeping G12 selectors first in the `.or` chain and using `.first()` after click resolution, or scope to the side panel container.
- Test B: UID-path swap should be backwards compatible (G9+ supports UID resource paths). If any older Grafana version still on the matrix rejects the UID path, fall back to numeric-ID via a runtime guard — but the supported matrix is >=12.1.0 per #43/#46, so this is theoretical.
- Sibling test `cert-generation.spec.ts:148` currently masks the regression. After fix, consider tightening it too — out of scope for this issue, log as follow-up.
- Reverting #44 not in scope; this builds on it.

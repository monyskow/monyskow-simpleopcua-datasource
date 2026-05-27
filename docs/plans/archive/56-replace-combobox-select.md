# Plan: replace Combobox with Select in ConfigEditor + lower floor to 10.4.0 (#56)

## Kontekst

Wykonanie ADR `docs/decisions/0001-grafana-support-floor.md`: zamiana `Combobox` (G12.1+) na `Select` (G7+) i obniżenie `grafanaDependency` do `>=10.4.0`. Matryca CI rośnie z 4×5=20 do 7×5=35 jobów. ADR rozstrzyga wszystkie sub-decyzje — to plan czysto wykonawczy.

Issue: #56. ADR: `docs/decisions/0001-grafana-support-floor.md`.

## Podejście

Trzy logiczne warstwy zmian, każda zostawia repo bootable:

1. **Kod komponentu + jego testy** — `Combobox` → `Select` w `ConfigEditor.tsx` + dopasowanie mocka i `getAllByRole` w `ConfigEditor.test.tsx`. Po tym kroku plugin nadal działa na G12+ — wymagania floor=12.1 jeszcze nie poluzowano, więc cofnięcie tylko tego commita to clean revert.
2. **Floor + matryca** — `plugin.json`, `.grafana-versions`, `e2e-matrix.yml`, docs. Tu schodzimy do 10.4.0.
3. **Doc-touchup** — `CLAUDE.md` (sekcja Grafana compatibility), `TESTING.md` (liczby wersji, opis matrycy).

E2E spec selektory w `tests/cert-generation.spec.ts` zostawiamy do pre-flight sprawdzenia — jeśli `input[role="combobox"]` nadal trafia w `Select`, nie ruszamy; jeśli nie, dorabiamy w **kroku 1b** osobnym commitem przed pełnym e2e:matrix.

## Pre-flight (zanim ruszysz pierwszą edycję)

Pre-flight to ~10 min czytania i jeden lokalny run — wszystko mechaniczne, bez decyzji ADR-class.

1. **API `Select` — confirmed via `node_modules/@grafana/ui/dist/types/components/Select/types.d.ts`:**
   - `options?: Array<SelectableValue<T>>` (`SelectableValue<T>` z `@grafana/data` — kompatybilny z naszym `{ label, value }[]`, dodatkowe pola opcjonalne).
   - `onChange: (value: SelectableValue<T>, actionMeta) => void` — `value.value` ma ten sam kształt co `Combobox`'owe `v.value` (string union narrow do `SecurityPolicy`/`SecurityMode`/`AuthMethod`).
   - `value?: T | SelectableValue<T> | null` — można przekazywać raw string jak w `Combobox`.
   - `placeholder?: string`, `isClearable?: boolean`, `width?: number` — wszystkie obecne.
   - `allowCustomValue?: boolean` — **nie używamy**. Trzy dropdowny to fixed enums (OPC-UA security policies to zamknięty zbiór wg IEC 62541, security mode = `None`/`Sign`/`SignAndEncrypt`, auth = nasze 3 wartości). Free-text byłby błędem.
2. **Sprawdź czy `input[role="combobox"]` w `tests/cert-generation.spec.ts:41` nadal działa z `Select`.** `Select` z `@grafana/ui` bazuje na `react-select` i renderuje `role="combobox"` na inputie — selektor _powinien_ trafiać. Weryfikacja: po kroku 1 odpal `npm run server` + `CI=1 npx playwright test tests/cert-generation.spec.ts` lokalnie. Jeśli pęknie, decyzję podejmij w **kroku 1b** (selektor lub option-id wzorzec).
3. **Sprawdź `.grafana-versions` vs TESTING.md drift.** Plik dziś ma 3 wpisy (`12.1.5`/`12.3.1`/`13.0.1`), TESTING.md mówi "6 versions" i "30 combos" — drift istniał już przed tym issue. Po naszej zmianie będzie 7 wersji × 5 auth = 35; spójność TESTING.md z `.grafana-versions` to część tego planu (krok 3).
4. **CI matrix w `e2e-matrix.yml` jest hardcoded** (lines 21–24 i 114). `.grafana-versions` _nie_ zasila CI — zasila tylko skrypty lokalne. Trzeba aktualizować oba miejsca w workflow YAML.

## Zmiany w plikach

### Krok 1 — komponent + jego testy (jeden commit, repo bootable, floor nadal 12.1)

- `src/components/ConfigEditor/ConfigEditor.tsx`
  - **Line 5** — import: `Combobox` → `Select`.
  - **Lines 189–195** — Security Policy `<Combobox>` → `<Select>`. Props: `width={30}`, `options={SECURITY_POLICY_OPTIONS}`, `value={jsonData.securityPolicy || 'None'}`, `onChange={(v) => onJsonDataChange('securityPolicy', v.value!)}`, `aria-label="Security Policy"` (dorzucamy — pomaga testom i a11y; `Combobox` go nie miał).
  - **Lines 198–204** — Security Mode j.w. (`aria-label="Security Mode"`).
  - **Lines 274–280** — Auth Method j.w. (`aria-label="Authentication Method"` — nie kolidować z istniejącym labelem "Method").
  - Uwaga: `v.value!` (non-null assert) bo `SelectableValue.value` jest opcjonalny. Alternatywa bez assert: `if (v.value)` guard. Wybór: zostawiamy assert (`Select` z fixed `options` zawsze zwraca `value`, brak `isClearable`).
- `src/components/ConfigEditor/ConfigEditor.test.tsx`
  - **Lines 21–49** — przepisanie komentarza i mocka: `MockCombobox` → `MockSelect`, eksport `Select` zamiast `Combobox`. `Select` w jsdom ma ten sam problem co `Combobox` (`react-select` używa pomiarów DOM-owych których jsdom nie ma) — mock jest nadal potrzebny. Mock renderuje plain `<select>` jak teraz, signature `onChange({ value })` pasuje 1:1 do tego co produkcyjny kod oczekuje.
  - **Lines 387, 392, 407** — zmień komentarze "Combobox" → "Select". Logika testu (`getAllByRole('combobox')`) bez zmian — plain `<select>` ma role=combobox tylko gdy ma atrybut, więc używamy `getAllByRole('combobox')` jak dziś (w naszym mocku to nadal `<select>`). **Weryfikacja:** uruchom `npm run test:ci` po edycji; jeśli `getAllByRole` zwróci 0, zmień na `getAllByTestId` lub dodaj `role="combobox"` w mocku.

### Krok 1b (opcjonalny, tylko jeśli pre-flight #2 pokaże regresję) — e2e selektor

- `tests/cert-generation.spec.ts` — jeśli `Select` renderuje option list inaczej niż `Combobox`'owe `combobox-option-<value>` (lines 23–34), przepisz na `page.getByRole('option', { name: ... })`. Zostaw `input[role="combobox"]` przy line 41 jeśli nadal działa.

### Krok 2 — floor + matryca (jeden commit)

- `src/plugin.json`
  - **Line 63** — `"grafanaDependency": ">=12.1.0"` → `">=10.4.0"`.
- `.grafana-versions`
  - Zastąp 3 obecne linie 7 wpisami wg ADR: `10.4.0`, `10.4.19`, `11.0.0`, `11.6.14`, `12.0.0`, `12.4.3`, `13.0.1`. Kolejność: rosnąco, ostatnia linia = latest-of-latest-major (wymaganie `bump-grafana-latest.sh` po fix #52).
- `.github/workflows/e2e-matrix.yml`
  - **Lines 21–24** — rozszerz `grafana_version` na 7 pozycji (kolejność jak `.grafana-versions`).
  - **Line 114** — `GRAFANA_VERSIONS="..."` w summary stepu — to samo 7 wartości oddzielone spacjami.

### Krok 3 — docs (jeden commit; opcjonalnie squash z krokiem 2 przed PR)

- `CLAUDE.md`
  - **Line 45** — zaktualizuj: `">=10.4.0"`. Usuń zdanie o `Combobox` jako blokerze floor.
  - **Line 46** — zaktualizuj listę wersji matrycy do 7 z ADR. Zmień zdanie "Don't widen it without checking `Combobox` + UID resource API support" na samo "UID resource API" (G9+ — wciąż aktualne dla floor 10.4).
- `TESTING.md`
  - **Line 25** — "(6 versions)" → "(7 versions)". Wall-clock w górę proporcjonalnie (~14 min).
  - **Line 26** — "30 combos" → "35 combos". Wall-clock w górę (~105–140 min).
  - **Line 304** — "(currently 6 versions)" → "(currently 7 versions)".
  - **Line 312** — "6 lines" → "7 lines".
  - **Line 314, 319** — "Lines 1–5" → "Lines 1–6" (latest-of-latest-major to ostatnia, czyli linia 7).
  - **Line 327** — "all versions in `.grafana-versions` (6) × all 5 auth configs = **30 combinations**" → "(7) × 5 = **35 combinations**".
  - **Line 335** — już mówi "7 Grafana versions × 5 auth configs = 35 jobs" — sprawdź czy nadal aktualne; **line 353** zdanie "adds `12.4.3` relative to `.grafana-versions`" trzeba usunąć (po naszej zmianie obie listy są identyczne).

## Kolejność (każdy krok = bootable commit)

1. **Pre-flight** — szybkie czytanie types, ewentualny lokalny `npm run e2e` na G13 z odpalonym `npm run server`, żeby zobaczyć czy `Select` selektor pasuje przed migracją (sanity, opcjonalne).
2. **Commit 1** — `feat(config-editor): replace Combobox with Select`. Po tym kroku `lint:fix` → `typecheck` → `test:ci` muszą być zielone na floor=12.1; e2e na G12.1/G13.0 bez regresji.
3. **(Warunkowo) Commit 1b** — `test(e2e): adjust selector for Select-based dropdowns` — tylko jeśli pre-flight #2 pokazał regresję.
4. **Commit 2** — `feat(compat): lower grafanaDependency floor to 10.4.0 and expand CI matrix`. Po tym kroku pełny `npm run e2e:matrix` powinien być zielony na 35 jobach.
5. **Commit 3** — `docs: update floor + matrix references in CLAUDE.md and TESTING.md` (lub squash do commita 2 przed PR).

**Git bisect safety:** po commicie 1 plugin działa na G12+ (floor niezmieniony). Po commicie 2 plugin działa na G10.4+ (Select istnieje). Bisect który wyląduje "w środku" jest spójny — nie ma broken intermediate state.

## Test plan

Kolejność uruchamiania po każdym commicie (krok mówi co dodać):

| Krok | Polecenia                                                                                                                                                                                                                                                     | AC pokryte                                                        |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| 1    | `npm run lint:fix && npm run typecheck && npm run test:ci`                                                                                                                                                                                                    | AC: brak Combobox w `src/`, Jest zielony, typecheck zielony       |
| 1    | `mage test && mage buildAll && npm run build`                                                                                                                                                                                                                 | AC: build artefaktów nie pęka (sanity — backend nie tknięty)      |
| 1    | `npm run server` (G13.0.1) → `npm run e2e` smoke                                                                                                                                                                                                              | AC: ConfigEditor renderuje na G13                                 |
| 2    | (po update `.grafana-versions`) `npm run server` z `GRAFANA_VERSION=10.4.0` ręcznie → otwórz `/connections/datasources/new` i klik przez ConfigEditor: 3 dropdowny renderują, `Save & Test` przechodzi (bez OPC-UA serwera może rzucić connection error — ok) | AC: ConfigEditor renderuje na G10.4.0 (smoke przed pełną matrycą) |
| 2    | `npm run e2e:all` lokalnie (35 min na 7 wersji × 1 auth=anon-none)                                                                                                                                                                                            | AC: każda wersja z `.grafana-versions` zielona na default auth    |
| 2    | `npm run e2e:matrix` lokalnie (90–140 min na 35 combos) — lub poczekać na CI jeśli lokalnie za drogo                                                                                                                                                          | AC: 35-job matryca zielona                                        |
| 3    | `grep -rn 'Combobox' src/` → 0 hitów (poza ew. doc komentarzami które też usuwamy)                                                                                                                                                                            | AC: brak Combobox w `src/`                                        |

**Edge cases do pokrycia w `ConfigEditor.test.tsx`** (już istnieją, sprawdź że nadal przechodzą po zmianie mocka):

- `getAllByRole('combobox')` zwraca 3 elementy w deterministycznej kolejności (Security Policy, Security Mode, Auth Method).
- `onChange` z `Select` przekazuje pełny `SelectableValue` — nasze handlery używają `v.value`, więc mock musi to symulować (już symuluje, sygnatura `onChange({ value })` jest zachowana).

## Ryzyka / open questions

- **`Select` props mismatch — typ `value`.** `Combobox` akceptuje raw `T`, `Select<T>` akceptuje `T | SelectableValue<T> | null`. Raw string nadal działa, ale TypeScript może wymagać explicit type param `<Select<SecurityPolicy>>`. **Mitigacja:** typecheck po kroku 1; jeśli sypie, dodaj generic param.
- **`v.value` może być `undefined`.** `SelectableValue.value` jest opcjonalny w typach `@grafana/ui` (bo `Select` ogólnie może być clearable). Nasze użycie ma fixed options bez `isClearable`, więc runtime nigdy nie zwróci `undefined`. **Mitigacja:** `v.value!` lub guard `if (v.value)`. Decyzja w planie: assert, bo fixed options + brak clearable = bezpieczne.
- **E2E `tests/cert-generation.spec.ts` selektor option-id.** `Combobox` ma deterministyczne `id="combobox-option-<value>"`, `Select`/`react-select` używa innego wzorca (`react-select-<n>-option-<idx>`). **Mitigacja:** krok 1b — jeśli pre-flight smoke pęka, przepisz na `getByRole('option', { name: ... })`.
- **Nieznane G10/G11 quirki.** Ryzyko znane z issue #56 i ADR. Out of scope tego planu — log do CLAUDE.md przyjdzie w follow-up issue (#57 wspomniany w issue body). Jeśli e2e:matrix odkryje cross-cutting bug który blokuje 3+ wersji → **STOP, follow-up jako osobne issue** (nie rozszerzaj tego PR).
- **Wall-clock e2e:matrix 90–140 min.** Akceptowalne per ADR (sub-decyzja 3 + 4). Nie podnosić `workers` ponad 2 — `node-opcua` cap.

## Reverse / abort plan

Jeśli e2e:matrix pokaże że G10.4 ma ≥1 dodatkowy blocker poza Combobox (np. brakujący `@grafana/ui` komponent w innym miejscu, theme tokens, runtime API):

- **Opcja A (preferowana, low cost):** zatrzymaj się przed mergem PR. Zostaw commit 1 jako osobny PR (`Select` działa wszędzie, więc zamiana sama z siebie nie szkodzi — _może_ poprawić UX nawet bez obniżania floor). Commit 2+3 odrzuć; nowy issue na konkretny G10 blocker. Po jego naprawie wróć do floor=10.4. ADR _nie_ wymaga zmian — tylko follow-up zostaje otwarty dłużej.
- **Opcja B (full abort):** revert wszystkich 3 commitów. Mało prawdopodobne — `Select` jest niezależnie poprawny.
- **Opcja C (ADR revision):** jeśli G10/G11 quirków jest dużo i koszt mitigacji jest wysoki — wróć do `architect` przez `/cx-architecture`. ADR-class decyzja.

Reverse cost dla każdej z opcji = niski (każdy commit jest atomowy i bootable).

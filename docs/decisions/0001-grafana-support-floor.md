# 0001. Grafana support floor: lower to 10.4.0, expand matrix to first+last patch per major

- Status: accepted
- Date: 2026-05-20

## Kontekst

Plugin `monyskow-simpleopcua-datasource` deklaruje dziś `grafanaDependency: ">=12.1.0"`. Powodem był pojedynczy komponent: `Combobox` z `@grafana/ui`, dostępny od 12.1 — używany trzykrotnie w `src/components/ConfigEditor/ConfigEditor.tsx` (Security Policy, Security Mode, Auth Method). Bez `Combobox` config page rzuca przy mount.

CI matrix po fix #52 obejmuje 4 wersje (`12.1.5`, `12.3.1`, `12.4.3`, `13.0.1`) × 5 trybów auth = 20 jobów. `.grafana-versions` jest źródłem prawdy; tylko ostatnia linia (latest-of-latest-major) floatuje przez weekly bump workflow, reszta jest pinned.

Realna baza użytkowników OPC-UA pluginu to środowiska MES/SCADA — tam Grafana 10.4 LTS i 11.x są nadal w produkcji, bo update Grafany w hali fabrycznej to projekt, nie click. Issue #53 prosi o obniżenie floor do 10.4.0 i rozszerzenie matrycy do first+last patch per major:

```
10.4.0
10.4.19
11.0.0
11.6.14
12.0.0
12.4.3
13.0.1
```

7 wersji × 5 auth = **35 jobów** w `e2e-matrix.yml`.

W kontekście są cztery niezależne sub-decyzje:

1. Jak rozwiązać blok `Combobox` < 12.1.
2. Co trafia do `.grafana-versions` przy security re-releases (baseline patch vs `-security-NN`).
3. Czy matryca jest jednolita (`must-green`) czy split (`supported` + `compat-tracking`).
4. Reguła bumpu — co floatuje, co pinned.

## Decyzja

**Obniżamy `grafanaDependency` do `>=10.4.0`. Matryca: 7 wersji wg tabeli z issue #53. `Combobox` → `Select` (zamiana w całym `ConfigEditor`). Security re-releases: bazowy patch w `.grafana-versions` (nie `-security-NN`). Matryca jest jednolita (wszystko must-green); split na `compat-tracking` rezerwujemy jako reakcję, nie projekt z góry. Bump workflow bez zmian: floatuje tylko latest-of-latest-major, reszta pinned.**

Uzasadnienie w jednym zdaniu: `Select` jest stabilną częścią `@grafana/ui` od G7+, więc zamiana jest boring tech (jeden komponent, jeden styl, zero version sniffing), reszta decyzji to minimalizacja powierzchni utrzymania dla solo dev.

### Reverse cost (najpierw)

**Medium.** Po komponentach:

- **`Combobox` → `Select`**: low. Trzy miejsca, czysta zamiana props (`value`/`onChange`/`options`). Jeśli za rok podniesiemy floor z powrotem na 12.1, mechaniczna re-zamiana lub po prostu zostawiamy `Select` (działa nadal). Nikt nie zależy od `Combobox`-specyficznego UX.
- **Matryca 35 jobów**: low–medium. GitHub Actions na publicznym repo = darmowe; wall-clock rośnie ~1.75x. Jeśli za rok uznamy „za wolno na każdy PR", wycinamy intermediate patches, wracamy do first per major = 5 jobów. Mechaniczna zmiana `.grafana-versions`.
- **`.grafana-versions` policy (baseline patch)**: low. Jedna linia per wersja, znana zasada „baseline = `MAJOR.MINOR.0` lub ostatni stabilny patch, bez `-security-NN` suffixu".
- **Akumulacja compat-quirków**: medium. To główny koszt cofnięcia. Każda znaleziona quirka G10/G11 zostanie zakopana w testach/setupie jako „don't re-discover" — przy podnoszeniu floor będziemy musieli wyłuskać które obejścia są nadal potrzebne. To jest realny dług, ale przewidywalny i lokalny (Playwright setup + auth fallbacki).

## Alternatywy

### Sub-decyzja 1: Combobox replacement strategy

- **(a) Zastąp `Combobox` `Select`-em we wszystkich miejscach** — **wybrane**. Plusy: jeden komponent, jeden styl, brak version sniffingu, `Select` istnieje od dawna w `@grafana/ui`, znany każdemu Grafana dev. Minusy: tracimy `Combobox`-specyficzny look (auto-size, inline filter na list-style); dla 3 dropdownów z fixed-options to nieistotne.
- **(b) Version-branched component** — odrzucone. Plusy: zachowuje `Combobox` na G12+. Minusy: musimy wprowadzić wrapper, version detect (`config.bootData.grafanaVersion` z `@grafana/runtime`), gałąź renderingu, dual snapshot w testach. Dla solo dev to gwarantowany loose end przy każdej zmianie Grafana API.
- **(c) Własny `Combobox`-like polyfill** — odrzucone bez dyskusji. Najwięcej kodu, najgorszy wybór, gwałci „boring tech wins".

### Sub-decyzja 2: Security-release policy w `.grafana-versions`

- **(a) Baseline patch (`13.0.1`)** — **wybrane**. Plusy: plik się nie zmienia przy każdym security re-release Grafany; mniej noise w git logu; bump workflow został właśnie naprawiony w #52 specjalnie pod tę regułę. Minusy: nie testujemy realnie security-patched obrazu w CI.
- **(b) Najświeższa re-release (`13.0.1-security-01`)** — odrzucone. Plusy: testujemy realny obraz, który użytkownicy mają w produkcji. Minusy: re-release pojawia się co kilka tygodni, plik churn, niepotrzebny noise w PR-ach. Security patche w Grafanie nie dotykają plugin SDK / `@grafana/ui` API, więc value testowania `-security-NN` dla pluginu = bliski zero. Można dodać oddzielny weekly `security-smoke` job poza matrycą, jeśli kiedyś będzie potrzeba (osobne ADR).

### Sub-decyzja 3: Matryca jednolita vs split

- **(a) Wszystko must-green** — **wybrane**. Plusy: prosta sygnalizacja, prosta reguła „red = blok release'u". Minusy: jeśli stara wersja zacznie się sypać, blokujemy się sami.
- **(b) Split `supported` (must-green) + `compat-tracking` (continue-on-error)** — odrzucone na ten moment, **zarezerwowane jako reakcja**. Wprowadzamy split tylko gdy konkretna wersja w `.grafana-versions` udowodni, że jest niestabilna mimo prawidłowego pluginu (np. flake po stronie Grafany). Dziś complexity bez konkretnej potrzeby. Reverse cost wprowadzenia split-a później = low (jeden `continue-on-error: true` na job).

### Sub-decyzja 4: Bump workflow shape

- **(a) Floatuje tylko latest-of-latest-major, reszta pinned** — **wybrane, status quo**. Plusy: stare floory są stabilne (one job = jeden image = przewidywalny test); tylko najnowszy major dostaje cotygodniowy bump. Minusy: nie wykrywamy regresji w intermediate majorach, jeśli Grafana je wyda — ale Grafana dla starych majorów wydaje praktycznie tylko security patches, a tych świadomie nie śledzimy (sub-decyzja 2).
- **(b) Float per major** — odrzucone. Plusy: zawsze aktualne. Minusy: 7 floatujących linii = 7x weekly PR, każdy do code review. Solo dev = nie.

### Główne odrzucone alternatywy dla całej ADR

- **Zostaw floor na `>=12.1.0`** — odrzucone. Wartość biznesowa floor=10.4 jest realna (MES/SCADA target audience), koszt zamiany `Combobox`→`Select` jest niski, koszt CI niezauważalny dla solo OSS.
- **Obniż jeszcze niżej, do 9.x** — odrzucone. G9 ma osobne breaking changes w plugin SDK i resource API (UID-based routing wprowadzony częściowo dopiero w 9.0, wcześniej numericId), brak nowoczesnego `@grafana/plugin-sdk-go` testowania, znikoma realna populacja użytkowników. ROI ujemny.

## Konsekwencje

### Co zyskujemy

- Pokrycie realnej populacji użytkowników OPC-UA (G10.4 LTS, G11 nadal w fabrykach).
- Sygnał regresji na pierwszym i ostatnim patchu każdego majora — jeśli plugin pęknie na `10.4.0` ale nie na `10.4.19`, wiemy że to konkretny patch Grafany.
- Mniej dependencji na `Combobox`-specyficzne API = łatwiejsze utrzymanie `ConfigEditor` w przyszłości.

### Co tracimy / co staje się trudniejsze

- E2E wall-clock ~1.75x (20 → 35 jobów; sequential per worker `npm run e2e:all` rośnie proporcjonalnie).
- Trzeba przejrzeć każdy `@grafana/ui` import w `ConfigEditor` i `QueryEditor` pod kątem „czy ten komponent istnieje w G10.4?". `@grafana/ui` ^12.2 jako dev dependency zostaje — ale komponenty wybierane do renderingu muszą mieć runtime-fallback do G10.4 API.
- Nowe „G10 quirks" / „G11 quirks" sekcje w `CLAUDE.md` po pierwszych e2e runach. To jest spodziewany koszt, ale konkretne quirki są nieznane do czasu uruchomienia matrycy.
- Potencjalna `jsdom` flakiness przy mockowaniu `Select` w testach (już istnieje precedens dla `Combobox` w `ConfigEditor.test.tsx`).

### Co musi się stać teraz (follow-up)

1. Zamiana `Combobox` → `Select` w `src/components/ConfigEditor/ConfigEditor.tsx` (3 miejsca) + update jest istniejącego mocka w `ConfigEditor.test.tsx`.
2. Aktualizacja `.grafana-versions` na 7 wpisów z tabeli issue #53.
3. `plugin.json` → `"grafanaDependency": ">=10.4.0"`.
4. Update `CLAUDE.md` (sekcja „Grafana compatibility") — nowy floor, nowa matryca, usunięcie zdania o `Combobox` jako blokerze floor.
5. Pierwsze pełne `npm run e2e:matrix` przeciw nowej matrycy + zebranie quirków G10/G11 do osobnej sekcji w `CLAUDE.md` (analog do istniejących G13 quirks).
6. Weryfikacja, że `scripts/bump-grafana-latest.sh` po fix #52 nadal trafia w prawidłową linię (latest-of-latest-major = ostatnia w pliku) niezależnie od liczby wpisów.

### Reverse cost (szczegółowo)

Medium overall — głównie z powodu akumulacji compat-quirków przez rok. Mechanicznie odwracalne:

- `Combobox` zamiana: 1 commit.
- Matryca: 1 commit (edycja `.grafana-versions`).
- Floor w `plugin.json`: 1 commit.
- Compat-quirki: każdy do oddzielnej oceny przy podnoszeniu floor — to jest realny koszt, niemechaniczny.

## Linki

- Issue: #53 „Lower grafanaDependency floor to 10.4.0"
- Powiązane PR-y: #52 (bump-grafana-latest fix), #49 (UID-based resource URL), #48 (G13 splash modal)
- Kod: `src/components/ConfigEditor/ConfigEditor.tsx`, `src/plugin.json`, `.grafana-versions`, `.github/workflows/e2e-matrix.yml`
- Następne ADR (kandydaci): „Compat-tracking matrix split policy" (jeśli sub-decyzja 3 zostanie zrewidowana), „Security-release smoke job" (jeśli sub-decyzja 2 zostanie zrewidowana)

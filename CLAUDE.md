# CLAUDE.md

Plik kontekstu dla Claude Code. Czytany na początku każdej sesji. Trzymaj go zwięzłym — to jest stale ładowane do kontekstu modelu.

---

## Kim jest użytkownik

Solo developer pracujący w domenie **MES / ERP / manufacturing** (Infor M3, Critical Manufacturing, Operator Systems MES) plus side projecty: analytics, custom tooling.

**Stack:**

- Go (backend, CLI)
- Python z polars / duckdb (analytics, ETL)
- TypeScript / Node (frontend, narzędzia)
- C# (.NET, integracje z M3)
- SQL — głównie MSSQL i DuckDB

**Preferencje:**

- Boring tech wins. Stdlib przed frameworkiem.
- Prostota przed elastycznością. Nie buduj abstrakcji „na zapas".
- Jeden plik > sześć plików, jeśli to wystarczy.
- Mniej zależności = mniej długu.

**Język:**

- Rozmowa z Claude: polski lub angielski (dopasuj się do tego, co pisze użytkownik).
- Wszystko, co trafia do repo lub GitHuba — domyślnie **angielski**: kod, komentarze, nazwy zmiennych, commit messages, treść issues, PR-ów, komentarzy, ADR-y, notatki w `docs/`, release notes, changelog, README.
- Wyjątek tylko gdy użytkownik wprost poprosi o inny język dla konkretnego artefaktu („ten ADR napisz po polsku") — wtedy ten jeden artefakt po polsku, ale default się nie zmienia.

---

## Filozofia frameworka

Trzy filary:

1. **Wymuszenie mechaniczne** — hooki, branch protection, CI. Tam, gdzie maszyna może egzekwować, człowiek nie powinien się tym zajmować.
2. **Siedem komend** pokrywających cały flow: `/cx-idea` → `/cx-architecture` → `/cx-backlog` → `/cx-issue` → `/cx-build` → `/cx-merge` → `/cx-ship`.
3. **Sześcioro agentów + dwa wewnętrzne (parser + merger)** = osiem perspektyw, każda z dopasowanym modelem.

---

## Komendy

Wszystkie komendy mają prefix `cx-` (od „claude-x", roboczo nazwa frameworka). Pełne instrukcje: `.claude/commands/`.

| Komenda            | Cel                                                | Etap flow |
| ------------------ | -------------------------------------------------- | --------- |
| `/cx-idea`         | Szkic pomysłu → notatka w `docs/ideas/`            | Inception |
| `/cx-architecture` | Decyzja architektoniczna → ADR w `docs/decisions/` | Design    |
| `/cx-backlog`      | Przegląd / planowanie issues w GitHubie            | Planning  |
| `/cx-issue`        | Tworzenie / refinement pojedynczego issue          | Planning  |
| `/cx-build`        | Implementacja: branch → kod → PR                   | Execution |
| `/cx-merge`        | Squash merge PR + sprzątanie branchy               | Execution |
| `/cx-ship`         | Release: tag, release notes, changelog             | Delivery  |

---

## Agenci (sześć perspektyw)

Pełne definicje: `.claude/agents/`.

| Agent       | Model  | Rola                                                             |
| ----------- | ------ | ---------------------------------------------------------------- |
| `developer` | Sonnet | Pisanie kodu — głównie wykonawca w `/cx-build`                   |
| `explorer`  | Haiku  | Szybkie szukanie po repo: gdzie jest X, kto woła Y               |
| `architect` | Opus   | Decyzje długoterminowe: ADR, wybory technologii, granice modułów |
| `pm`        | Sonnet | Refinement issues, priorytetyzacja, dopytywanie o intencję       |
| `gh-parser` | Haiku  | Wewnętrzny parser: agregacja `gh` outputu, format, walidacja     |
| `merger`    | Haiku  | Wewnętrzny: squash merge PR + cleanup branchy (lokalnie/remote)  |
| `reviewer`  | Sonnet | Code review, zapach kodu, drobne ulepszenia                      |
| `tester`    | Sonnet | Generowanie testów, edge cases, weryfikacja założeń              |

`gh-parser` i `merger` są **wewnętrzne** — nie wołasz ich bezpośrednio, używają ich komendy. `gh-parser` tylko czyta, `merger` wykonuje merge + cleanup po pre-checkach komendy.

---

## Reguła doboru modelu (kluczowa)

**Haiku to DEFAULT.** Eskalacja do Sonneta/Opusa wymaga uzasadnienia, nie odwrotnie.

- **Haiku** — mechanika: parsing `gh`, agregacja, format, status checks, walidacje, transformacje danych.
- **Sonnet** — osąd: kodowanie, code review, planowanie, dopytywanie o intencję.
- **Opus** — decyzje długoterminowe: architektura (ADR), krytyczne release'y, kontrowersyjne wybory bibliotek.

Jeśli zadanie polega na **przetwarzaniu danych z deterministycznym wynikiem** → Haiku.
Jeśli wymaga **oceny jakości, intencji lub kompromisu** → Sonnet.
Jeśli ma **długoterminowe konsekwencje** (architektura, release na produkcję) → Opus.

---

## Default agent (zwykła rozmowa)

Gdy użytkownik pisze do mnie bez wywoływania komendy — odpowiadam z perspektywy **głównej sesji**, nie sub-agenta. Zachowuję się jak doświadczony pair-programmer:

- Czytam kontekst (CLAUDE.md, ostatni `git log`, zmienione pliki) zanim odpowiem na pytanie o repo.
- Nie zgaduję domeny — pytam, jeśli nie wiem.
- Krótkie odpowiedzi, bez ścian tekstu.

Sub-agentów wołam **tylko** gdy:

- Komenda explicitly tego wymaga (np. `/cx-build` deleguje do `developer`),
- Zadanie wymaga przeszukania repo szerzej niż 3 grep/find (→ `explorer`),
- Decyzja jest ADR-class (→ `architect`).

---

## Anti-patterns (czego nie robić)

1. **Nie commituj automatycznie.** Commit i push to ostatni krok człowieka. Wyjątek: `/cx-build` może utworzyć branch + PR, ale nie merge.
2. **Nie nadpisuj `main`.** Hook `block-main-push.sh` wymusza, ale agent też ma to wiedzieć.
3. **Nie dodawaj dependencies bez ADR.** Każda nowa biblioteka spoza stdlib → `/cx-architecture` najpierw.
4. **Nie pisz testów „na pokaz".** Testy mają łapać regresje w realnych ścieżkach, nie pokrywać linie.
5. **Nie generuj komentarzy oczywistych.** Komentarz odpowiada na „dlaczego", nie „co". Jeśli nazwa zmiennej już to mówi — usuń komentarz.
6. **Nie wprowadzaj backwards-compat shimów** dla kodu, który nie jest jeszcze nigdzie deployed. Solo dev = nikt nie zależy od starego API.
7. **Nie eskaluj do Opusa „na wszelki wypadek".** Haiku załatwia 70% zadań mechanicznych. Sonnet 25%. Opus 5%.
8. **Nie twórz plików dokumentacyjnych proaktywnie.** README, CHANGELOG — jeśli framework lub komenda ich nie wymaga, nie piszemy.

---

## Kontrakt z hookami

Hooki w `.claude/hooks/` egzekwują:

- `block-main-push.sh` — push do `main` blokowany lokalnie (drugi pas: branch protection na GH).
- `validate-branch-name.sh` — branche muszą mieć prefix: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/`, `spike/`.
- `validate-commit-msg.sh` — Conventional Commits.
- `auto-format.sh` — uruchamia formatery dla zmienionych plików (gofmt, ruff, prettier, dotnet format).
- `lint-feedback.sh` — lint po edycji (golangci-lint, ruff, eslint).
- `session-context.sh` — wstrzykuje skrót `git log` + zmienione pliki na początku sesji.
- `dangerous-ops.sh` — łapie `rm -rf`, `git reset --hard`, `force push` → wymaga potwierdzenia.

Jeśli hook blokuje — **nie omijaj go z `--no-verify`**. Hook ma rację. Jeśli się myli, popraw hook.

---

## Wersja frameworka

Aktualna wersja w `VERSION` (root template'u). Każdy projekt utworzony przez `init-project.sh` zapisuje swoją wersję w `.solo-dev-version`. `update-project.sh` używa tego do wykrywania driftu.

---

## Co jest w tym repo

To repo (`solo-dev-baseline`) pełni **podwójną rolę**:

- **Template repo** — używane przez GitHub „Use this template" do tworzenia nowych projektów.
- **Source of truth** — `scripts/init-project.sh` i `scripts/update-project.sh` kopiują framework do projektów.

Dla projektów docelowych: `scripts/` nie jest kopiowane — to wewnętrzne narzędzia template'u.

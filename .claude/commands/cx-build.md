---
description: Implementacja issue — branch → plan → kod → testy → review → PR
argument-hint: <numer-issue> [--plan]
allowed-tools: Bash(gh:*), Bash(git:*), Bash(jq:*), Read, Write, Edit, Grep, Glob
---

# /cx-build — implementacja end-to-end

Bierze numer issue, prowadzi przez cały cykl: branch → plan → implementacja → testy → review → PR. **NIE merge'uje** — to ostatni krok człowieka.

## Argumenty

- `$ARGUMENTS` — numer issue (np. `42`). Wymagany. Jeśli puste, zapytaj.
- `--plan` (opcjonalna flaga) — wymusza pełny plan przez `tech-lead`, nawet jeśli issue nie ma label `epic`/`complex`. Przykład: `/cx-build 42 --plan`.

## Krok po kroku

### 1. Pobranie kontekstu issue

- Deleguj do `gh-parser` (Haiku): _„`gh issue view $ARGUMENTS --json number,title,body,labels` jako markdown."_
- Pokaż użytkownikowi tytuł + acceptance criteria.
- Zapytaj: _„Acceptance criteria są jasne? Brak otwartych pytań? [y/n]"_. Jeśli `n` → przerwij i sugeruj `/cx-issue $ARGUMENTS` najpierw.

### 2. Branch

- Wygeneruj nazwę: prefix z labela (`feat/`, `fix/`, `chore/`, `refactor/`, `test/`, `spike/`) + slug z tytułu + `-#NNN`.
  - Przykład: issue #42 z label `feat`, tytuł "Handle empty CSV" → `feat/handle-empty-csv-42`.
- Sprawdź że jesteś na `main` i czysto: `git status --porcelain` puste, `git rev-parse --abbrev-ref HEAD` == `main`.
- `git checkout -b <branch>`. Hook `validate-branch-name.sh` to sprawdzi.

### 3a. Plan

Dwie gałęzie, zależne od trigger:

**Gałąź A — full plan (`tech-lead`):** jeśli labels z kroku 1 zawierają `epic` lub `complex`, ALBO user wywołał `/cx-build $ARGUMENTS --plan`:

- Deleguj do `tech-lead` (Opus):
  - Wklej AC + tytuł + skrót CLAUDE.md.
  - Polecenie: _„Stwórz plan w `docs/plans/<numer>-<slug>.md` według szablonu. Max 1 strona. Wskaż pliki konkretnie. Identyfikuj ryzyka."_
- Pokaż treść planu. Zapytaj: _„Plan OK? [y/n/edit]"_.
  - `edit` → user edytuje plik ręcznie, potem `y`.
  - `n` → przerwij workflow.
- Jeśli `tech-lead` zwrócił „STOP: potrzebny `/cx-architecture`" → przerwij, zasugeruj ADR.
- Plan zostaje w repo i zostanie zacommitowany razem z kodem w kroku 6.

**Gałąź B — lightweight (default):** dla pozostałych zadań (małe fixe, drobne feature'y bez label) krok jest **obowiązkiem `developera`** w kroku 3 — wypisuje plan inline i czeka na akceptację (patrz Reguła #1 w `developer.md`). Tu w komendzie nic explicite nie robimy.

### 3. Implementacja

- Deleguj do `developer` (Sonnet):
  - Wklej acceptance criteria.
  - Wklej skrót CLAUDE.md (stack, preferencje).
  - Polecenie: _„Jeśli `docs/plans/<numer>-*.md` istnieje — czytaj go najpierw i trzymaj się go. Jeśli nie istnieje — wypisz inline plan (pliki + kolejność + jak zweryfikujesz) i czekaj na `y` przed pierwszą edycją (Reguła #1). Następnie zaimplementuj zgodnie z AC. Najmniejsza możliwa zmiana. Po edycji uruchom istniejące testy."_
- Po pracy `developera`: pokaż listę zmienionych plików + jego komentarz.

### 4. Testy

- Deleguj do `tester` (Sonnet):
  - Polecenie: _„Sprawdź czy nowy kod ma test pokrywający wszystkie acceptance criteria. Dopisz brakujące testy (regression test obowiązkowy dla `fix/_`). Uruchom suite."\*
- Jeśli testy failują → wracaj do `developera` z listą błędów. Max 2 iteracje, potem przerwij i pokaż użytkownikowi co poszło nie tak.

### 5. Self-review

- Deleguj do `reviewer` (Sonnet):
  - Polecenie: _„Review diffu (`git diff main...HEAD`). Format: blockers / discussion / nits / verdict."_
- Pokaż użytkownikowi raport.
- Jeśli **blockers** istnieją → wróć do `developera`. Inaczej → krok 6.

### 6. Commit + PR

- Zaproponuj commit message (Conventional Commits): `<type>(<scope>): <opis> (#<issue>)`.
- Pokaż użytkownikowi → poczekaj na `y/n`.
- Jeśli `y`: `git add` + `git commit`. Hook `validate-commit-msg.sh` waliduje.
- Push: `git push -u origin <branch>` (hook `block-main-push.sh` jest w innym branchu, więc OK).
- `gh pr create --title "<conv-commit>" --body "<...>" --assignee @me`. Body zawiera: `Closes #<issue>` + krótkie podsumowanie + checklist AC.
- Pokaż URL PR-a.

## Czego NIE robisz

- Nie merge'ujesz PR-a.
- Nie zamykasz issue ręcznie (`Closes #N` w PR body to zrobi po merge).
- Nie pushujesz do `main`.
- Nie usuwasz brancha — to po merge.

## Reguły

1. **Stop na blokerze.** Jeśli `developer`, `tester` lub `reviewer` zgłaszają twardy problem → przerwij i pytaj użytkownika.
2. **Każda iteracja to commit.** Lepiej 3 commity niż 1 wielki + amend.
3. **Hook ma rację.** Jeśli `validate-commit-msg` blokuje — popraw message, nie omijaj.

## Format outputu (po sukcesie)

```
✓ PR utworzone: https://github.com/<user>/<repo>/pull/87

Branch: feat/handle-empty-csv-42
Commits: 1 (feat(parser): handle empty CSV (#42))
Files: parser.go, parser_test.go
Tests: 14/14 pass
Review: APPROVE

Następny krok: review w GitHubie i merge ręczny.
```

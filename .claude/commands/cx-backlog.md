---
description: Przegląd backlogu — agreguj issues, zaproponuj kolejność
argument-hint: [filter] (np. "open", "label:bug", "milestone:v1")
allowed-tools: Bash(gh:*), Bash(jq:*), Read
---

# /cx-backlog — przegląd i planowanie

Pokazuję obecny stan backlogu i pomagam zdecydować *co dalej*. Mechanika (pobieranie z `gh`) → `gh-parser` (Haiku). Osąd (priorytety) → `pm` (Sonnet).

## Argument

`$ARGUMENTS` — opcjonalny filtr w składni `gh issue list` (np. `state:open label:bug`, `milestone:v1`). Jeśli puste → wszystkie open.

## Krok po kroku

1. **Deleguj do `gh-parser`** (Haiku):
   - Polecenie: *„Pobierz `gh issue list --state open --json number,title,labels,updatedAt,assignees --limit 100` z dodatkowym filtrem `$ARGUMENTS` jeśli niepusty. Zwróć tabelę markdown posortowaną po `updatedAt` desc, kolumny: `#`, `tytuł` (skrócony do 60 zn.), `labels`, `dni od update`."*
   - Odbierz tabelę.

2. **Pokaż użytkownikowi tabelę** wraz z agregatami:
   - Total open: N
   - Po labelach (top 5): `bug: 3`, `feat: 7`, `chore: 2`...
   - Stale (>30 dni bez update): liczba

3. **Zapytaj użytkownika**: *„Chcesz przejrzeć szczegóły konkretnych issues, zaproponować kolejność, czy zamknąć stale?"* — opcje:
   - `details <numbers>` — pełny widok przez `gh issue view`.
   - `plan` — `pm` proponuje kolejność top-5 z uzasadnieniem.
   - `stale` — lista do potencjalnego zamknięcia (deleguj `pm` do oceny).
   - `none` — zakończ, użytkownik tylko chciał zobaczyć stan.

4. **Jeśli `plan`**: deleguj do `pm` (Sonnet) z listą issues + skrótem CLAUDE.md (kontekst stacku/domeny). PM zwraca top-5 w kolejności + jedno-zdaniowe uzasadnienie każdego.

## Czego NIE robisz

- Nie tworzysz, nie zamykasz, nie modyfikujesz issues. Tylko pokazujesz.
- Nie eskalujesz do Opusa. Backlog review = mechanika + lekki osąd.

## Format outputu

```
## Backlog ({filter or "all open"})

| #    | Tytuł                          | Labels       | Stale (d) |
|------|--------------------------------|--------------|-----------|
| 42   | Refactor parser for empty CSV  | feat,parser  | 2         |
| ...  |                                |              |           |

Total: 17 open
By label: feat:7  bug:3  chore:5  spike:2
Stale (>30d): 4

Co dalej? [details / plan / stale / none]
```

---
description: Decyzja architektoniczna — utwórz ADR w docs/decisions/
argument-hint: <tytuł-decyzji>
allowed-tools: Read, Write, Edit, Bash(ls:*), Bash(grep:*), Bash(date:*), Glob
---

# /cx-architecture — Architecture Decision Record

Wymuszasz strukturalne myślenie nad decyzją, której odwrócenie kosztuje. Output: ADR w `docs/decisions/NNNN-tytuł.md`.

## Argument

`$ARGUMENTS` — tytuł decyzji (np. "use polars instead of pandas"). Jeśli puste, zapytaj.

## Krok po kroku

1. **Deleguj do `architect`** (model: opus). Przekaż mu:
   - Tytuł z `$ARGUMENTS`.
   - Kontekst: zawartość `CLAUDE.md` (sekcja stack + preferencje).
   - Listę istniejących ADR: `ls docs/decisions/*.md` (jeśli są).
   - Polecenie: *„Utwórz ADR zgodnie z szablonem w swojej definicji. Numer = max(istniejące) + 1, format `NNNN-slug.md`."*

2. **Architekt może dopytać** użytkownika o brakujący kontekst (1-3 pytania, nie więcej). Przekazuj odpowiedzi.

3. **Po zapisaniu ADR przez architekta**:
   - Sprawdź że plik powstał (`ls docs/decisions/`).
   - Pokaż użytkownikowi ścieżkę i streszczenie: *decyzja + kluczowy trade-off + reverse cost*.

4. **Status początkowy**: zwykle `proposed`. Jeśli użytkownik potwierdzi *„zatwierdzam"* — `architect` zmienia na `accepted`.

## Czego NIE robisz

- Nie piszesz ADR sam (to robi `architect` z modelem opus).
- Nie commitujesz pliku.
- Nie linkujesz do nieistniejących issues.

## Reguły

1. **Tylko realne decyzje.** Jeśli użytkownik pyta *„czy ADR dla logowania?"* a logowanie to `log/slog`-stdlib bez alternatyw — odmów: *„To nie wymaga ADR. ADR jest tam, gdzie są realnie różne opcje."*
2. **Minimum 2 alternatywy.** Architekt MUSI wymienić odrzucone opcje, nie tylko wybór.
3. **Reverse cost first.** Pierwsze pytanie do architekta: *jak trudno wycofać tę decyzję za rok?*

## Format outputu (po wykonaniu)

```
✓ ADR utworzone: docs/decisions/0007-use-polars.md

Decyzja: użyj polars zamiast pandas dla nowych pipeline'ów ETL.
Kluczowy trade-off: szybkość + memory efficiency vs. mniejsza społeczność / mniej tutoriali.
Reverse cost: średni — interface DataFrame podobny, migracja per-pipeline.

Następne kroki sugerowane przez architekta:
- [ ] follow-up issue: dopisać polars do docs/stack.md
- [ ] spike: porównanie performance na realnym datasecie
```

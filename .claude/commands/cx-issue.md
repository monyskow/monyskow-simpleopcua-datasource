---
description: Tworzenie / refinement issue — z luźnego pomysłu w trackowane zadanie
argument-hint: [tytuł lub numer-issue]
allowed-tools: Bash(gh:*), Bash(jq:*), Read, Write
---

# /cx-issue — issue refinement

Dwa tryby:
- **Create** (`$ARGUMENTS` to tytuł) — tworzymy nowe issue od zera lub z `docs/ideas/`.
- **Refine** (`$ARGUMENTS` to numer, np. `42`) — bierzemy istniejące i ulepszamy.

## Detekcja trybu

- Jeśli `$ARGUMENTS` matchuje `^[0-9]+$` → **Refine**.
- Jeśli pusty → zapytaj: *„Numer istniejącego issue, czy tytuł nowego?"*
- Inaczej → **Create**.

## Tryb CREATE

1. **Zapytaj**: *„Czy ten pomysł istnieje już jako notatka w `docs/ideas/`?"*. Jeśli tak — wczytaj zawartość notatki i przekaż jako kontekst.
2. **Deleguj do `pm`** (Sonnet):
   - Tytuł: `$ARGUMENTS`.
   - Kontekst: zawartość notatki (jeśli była) + skrót CLAUDE.md.
   - Polecenie: *„Zaproponuj title + body w formacie Acceptance Criteria zgodnie z twoją definicją. Jedno-dwa pytania uzupełniające."*
3. **Po dialogu z PM**: pokaż użytkownikowi finalną wersję body. Zapytaj: *„Tworzymy issue? [y/n]"*
4. **Jeśli y**: utwórz przez `gh issue create --title "..." --body-file <tmpfile>`. Zaproponuj label (PM sugeruje) — pytaj o zatwierdzenie.
5. **Po utworzeniu**: pokaż URL i numer.

## Tryb REFINE

1. **Deleguj do `gh-parser`** (Haiku):
   - Polecenie: *„`gh issue view $NUMBER --json number,title,body,labels,state,comments` — zwróć JSON sparsowany do markdown."*
2. **Pokaż użytkownikowi obecny stan** issue.
3. **Deleguj do `pm`** (Sonnet) z istniejącym body:
   - Polecenie: *„Wskaż braki: czy są acceptance criteria? czy out-of-scope jest jawny? czy są open questions, które blokują start? Zaproponuj poprawiony body."*
4. **Po propozycji PM**: pokaż diff body (stary vs nowy). Zapytaj: *„Aktualizujemy issue? [y/n]"*
5. **Jeśli y**: `gh issue edit $NUMBER --body-file <tmpfile>`.

## Czego NIE robisz

- Nie zamykasz issues (to świadoma decyzja użytkownika).
- Nie dodajesz milestone bez pytania.
- Nie kopiujesz body z `docs/ideas/` 1:1 — PM ma to przerobić.
- Nie wołasz `developer` — `/cx-issue` kończy się utworzeniem trackowanego zadania, kod to `/cx-build`.

## Format outputu

```
✓ Issue #43 utworzone: https://github.com/<user>/<repo>/issues/43

Title: Handle empty CSV in importer
Labels: feat, parser

Acceptance criteria:
- [ ] importer.Parse(empty) returns ([]Row{}, nil) instead of panic
- [ ] regression test in parser_test.go
- [ ] log warning (info-level)

Następny krok: gdy gotowy do pracy → /cx-build 43
```

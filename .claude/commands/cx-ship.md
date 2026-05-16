---
description: Release — tag, release notes, changelog
argument-hint: <wersja> (np. v1.2.0) lub auto
allowed-tools: Bash(gh:*), Bash(git:*), Bash(jq:*), Read, Write, Edit
---

# /cx-ship — release

Tworzy nowy release: tag, release notes z mergowanych PR-ów, opcjonalnie wpis w `CHANGELOG.md`. Mechanika → `gh-parser` (Haiku). Decyzja o wersji → użytkownik (z sugestią). Krytyczny release → `architect` (Opus) jako sanity check.

## Argument

- `$ARGUMENTS` = wersja (np. `v1.2.0`) lub `auto` (auto-increment patch).
- Jeśli puste — zapytaj: *„Major / minor / patch?"*

## Krok po kroku

### 1. Pre-checks

- Jesteśmy na `main` i czysto? (`git status --porcelain` puste, branch == `main`).
- `git fetch && git status` — czy local == remote?
- Ostatni tag: `git describe --tags --abbrev=0` → np. `v1.1.3`.
- Jeśli `auto`: bump patch → `v1.1.4`.

### 2. Zebranie zmian

- Deleguj do `gh-parser` (Haiku):
  - Polecenie: *„Wylistuj PR-y mergowane od taga `<last-tag>` do HEAD. `gh pr list --state merged --search 'merged:>$LAST_TAG_DATE' --json number,title,labels,author --limit 100`. Pogrupuj po type (z conventional commit prefix w title): feat / fix / chore / docs / refactor / test."*
- Odbierz strukturalną listę.

### 3. Generacja release notes

- Format markdown:
  ```
  ## v1.2.0 — YYYY-MM-DD
  
  ### Features
  - feat(parser): handle empty CSV (#42)
  
  ### Fixes
  - fix(api): retry on 503 (#51)
  
  ### Other
  - chore(deps): bump duckdb (#48)
  ```
- Pokaż użytkownikowi.

### 4. Sanity check (opcjonalny — tylko dla major / krytyczny)

- Jeśli `$VERSION` ma `.0.0` lub user wpisał *„krytyczny"* → deleguj do `architect` (Opus):
  - Polecenie: *„Przejrzyj listę zmian przed major release. Czy są breaking changes nieoznaczone? Czy migracje zostały udokumentowane? Czy ADR-y `proposed` powinny być `accepted` przed release?"*
- Jeśli architect zgłasza concerns → przerwij, pokaż użytkownikowi.

### 5. CHANGELOG.md

- Jeśli `CHANGELOG.md` istnieje → prepend nową sekcję na górze (po nagłówku).
- Jeśli nie istnieje — zapytaj: *„Utworzyć CHANGELOG.md? [y/n]"*. Jeśli `y` — utwórz z tą jedną sekcją.

### 6. Tag + release

- Pokaż użytkownikowi PLAN: *„Utworzę tag `<v>`, push do origin, GitHub release z release notes. [y/n]"*
- Jeśli `y`:
  - `git add CHANGELOG.md` (jeśli zmieniony) + `git commit -m "chore(release): <v>"`.
  - `git tag -a <v> -m "Release <v>"`.
  - `git push origin main && git push origin <v>`.
  - `gh release create <v> --title "<v>" --notes-file <release-notes-file>`.
  - Pokaż URL release.

## Czego NIE robisz

- Nie tworzysz release jeśli pre-checks fail.
- Nie pomijasz architect dla major bez powodu.
- Nie pushujesz force.
- Nie modyfikujesz starych release notes / tagów.

## Reguły

1. **Semver luźny solo dev:**
   - Major (`X.0.0`) — breaking change w API publicznym.
   - Minor (`x.Y.0`) — nowa feature, backward-compat.
   - Patch (`x.y.Z`) — fix / chore / docs.
2. **Jeden release = jedna spójna paczka.** Jeśli na `main` jest niedokończona feature → przerwij.
3. **Release-drafter (`.github/workflows/release-drafter.yml`) działa równolegle.** Sprawdź czy nie ma już szkicu — jeśli tak, użyj go jako bazy zamiast generować od zera.

## Format outputu (po sukcesie)

```
✓ Release v1.2.0 opublikowany: https://github.com/<user>/<repo>/releases/tag/v1.2.0

Tag: v1.2.0 (commit: abc1234)
Zmian od v1.1.3: 8 PR-ów (3 feat, 2 fix, 3 chore)
CHANGELOG.md: zaktualizowany

Następny krok: deploy (poza scope frameworka).
```

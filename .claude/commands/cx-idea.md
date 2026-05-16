---
description: Szkic pomysłu — zamień luźną myśl w notatkę w docs/ideas/
argument-hint: <jedno-zdanie-tytułu>
allowed-tools: Read, Write, Bash(date:*), Bash(ls:*)
---

# /cx-idea — szkic pomysłu

Zamieniam luźny pomysł użytkownika w **trwałą notatkę** w `docs/ideas/`. To NIE jest jeszcze issue ani ADR — to surowa myśl, która ma przetrwać kontekst sesji.

## Argument

`$ARGUMENTS` — jedno zdanie tytułu pomysłu. Jeśli puste, zapytaj użytkownika *o czym pomysł?* zanim cokolwiek zapiszesz.

## Krok po kroku

1. **Wygeneruj nazwę pliku**: `docs/ideas/YYYY-MM-DD-slug.md` (slug z `$ARGUMENTS`, lower-kebab-case, max 50 znaków).
2. **Sprawdź czy plik nie istnieje** (`ls docs/ideas/`). Jeśli kolizja — dodaj sufiks `-2`, `-3`.
3. **Zadaj 1-2 pytania** uzupełniające (nie więcej):
   - *Co konkretnie chcesz, żeby się działo?*
   - *Dlaczego teraz, nie kiedyś?* (opcjonalnie)
4. **Zapisz plik** z szablonem poniżej.
5. **Pokaż użytkownikowi ścieżkę** + sugestię: *„Gdy będziesz gotów, użyj `/cx-issue` żeby zamienić to w trackowane zadanie, lub `/cx-architecture` jeśli to decyzja długoterminowa."*

## Czego NIE robisz

- Nie tworzysz issue na GitHubie (to robota `/cx-issue`).
- Nie piszesz ADR (to `/cx-architecture`).
- Nie deklarujesz scope ani estimacji.
- Nie wołasz sub-agentów — to czysta operacja zapisu pliku.

## Szablon

```markdown
# {{tytuł}}

- Date: {{YYYY-MM-DD}}
- Status: idea

## Co

Krótki opis pomysłu w 2-3 zdaniach.

## Dlaczego

Co problem rozwiązuje, jaki ma kontekst (1-2 zdania).

## Otwarte pytania

- pytanie 1
- pytanie 2

## Możliwe następne kroki

- [ ] eksploracja: ...
- [ ] spike: ...
- [ ] issue: ...
```

## Format outputu

Po zapisie pokaż:
- Ścieżkę pliku (klikalną).
- Co zapisałeś (skrót, 2-3 linie).
- Sugestię następnej komendy.

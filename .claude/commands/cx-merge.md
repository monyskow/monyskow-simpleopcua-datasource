---
description: Merge PR (squash) + sprzątanie branchy
argument-hint: <numer-PR> (opcjonalnie — auto-detect z aktualnego brancha)
allowed-tools: Bash(gh:*), Bash(git:*), Bash(jq:*), Read
---

# /cx-merge — squash merge + cleanup

Pre-checks → squash merge → delete remote branch → cleanup lokalny. Cała mechanika delegowana do `merger` (Haiku). Komenda **nie pyta** użytkownika o nic ponad finalne potwierdzenie planu — wszystko inne wynika z PR-a.

## Argument

- `$ARGUMENTS` = numer PR (np. `87`).
- Jeśli puste — auto-detect z aktualnego brancha: `gh pr view --json number -q .number`. Jeśli auto-detect zwróci pusto → poproś użytkownika o numer.

## Krok po kroku

### 1. Pobranie stanu PR

- Deleguj do `gh-parser` (Haiku):
  - _„`gh pr view $PR --json number,title,state,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,headRefName,baseRefName,isDraft`. Zwróć jako tabelę markdown."_
- Odbierz dane.

### 2. Pre-checks (twarde)

Wszystkie muszą być spełnione. Jeśli którykolwiek fail → **stop**, pokaż użytkownikowi co i czemu, zasugeruj akcję.

| Check            | Warunek                                                                  | Jak zaadresować przy fail                                                                                         |
| ---------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| State            | `state == OPEN`                                                          | Już zmergowany / zamknięty — nic do roboty.                                                                       |
| Draft            | `isDraft == false`                                                       | „Wyjmij z draft w GitHubie."                                                                                      |
| Base             | `baseRefName == main`                                                    | Inny base = inny flow, nie ten command.                                                                           |
| Review           | `reviewDecision in (APPROVED, null)`                                     | Jeśli `CHANGES_REQUESTED` — popraw blockery najpierw. `null` (brak review) jest OK dla solo dev.                  |
| CI               | wszystkie `statusCheckRollup[].conclusion` ∈ (SUCCESS, NEUTRAL, SKIPPED) | Lista failed checks + sugestia: „rerun w GH UI / zobacz logi".                                                    |
| Mergeable        | `mergeable == MERGEABLE`                                                 | Jeśli `CONFLICTING` — rebase brancha ręcznie.                                                                     |
| MergeStateStatus | ∈ (CLEAN, HAS_HOOKS, UNSTABLE)                                           | `BEHIND` → branch nie up-to-date z base, ale GitHub i tak pozwoli na squash; ostrzeż. `BLOCKED` / `DIRTY` → stop. |

### 3. Plan + potwierdzenie

Pokaż użytkownikowi:

```
Plan:
- Merge PR #87 "feat(parser): handle empty CSV" (squash)
- Skasuj remote branch: feat/handle-empty-csv-42
- Switch na main, pull --ff-only
- Sprzątnij stale lokalne tracking refs (fetch --prune)
- Usuń lokalne branche z upstream=gone (-d, nie -D)

Wykonać? [y/n]
```

Czekaj na `y`. `n` → koniec, nic się nie dzieje.

### 4. Egzekucja

- Deleguj do `merger` (Haiku):
  - _„Wykonaj merge + cleanup dla PR #$PR. Branch lokalny brancha PR-a: `$HEAD_REF`. Trzymaj się swojego kontraktu — squash, --delete-branch, --ff-only, -d nie -D."_
- Odbierz raport od `merger`.

### 5. Output

Pokaż raport `merger` + URL zmergowanego PR-a.

## Czego NIE robisz

- Nie używasz `--admin` ani `--no-verify` żeby ominąć failed checks.
- Nie robisz force push, nie resetujesz hard.
- Nie usuwasz brancha który ma niezmergowane commity.
- Nie pytasz o strategy merge — squash jest defaultem frameworka.
- Nie zamykasz issue ręcznie — `Closes #N` w PR body zrobi to po merge.

## Reguły

1. **Pre-checks są twarde.** Solo dev = jedyna linia obrony. Failed CI = nie merguj.
2. **Squash zawsze.** Jeden PR = jeden commit na main. Iteracje zostają w historii brancha (który i tak znika).
3. **`merger` nie kombinuje.** Jeśli zwróci błąd → pokaż użytkownikowi komendę do ręcznego naprawienia. Nie próbuj drugi raz z innymi flagami.
4. **Hook ma rację.** Jeśli `block-main-push.sh` strzela podczas pull/fetch — popraw hook, nie omijaj.

## Format outputu (po sukcesie)

```
✓ PR #87 zmergowany: https://github.com/<user>/<repo>/pull/87

Strategia: squash
Commit na main: abc1234
Remote branch: feat/handle-empty-csv-42 (usunięty)
Lokalnie usunięte branche: feat/handle-empty-csv-42
Stale refs wyczyszczone: 2

Następny krok: /cx-ship gdy nazbiera się paczka, albo następne /cx-build.
```

# CONTRIBUTING

Kontrakt dla **człowieka** i **agenta** pracującego w tym repo. Krótki, egzekwowalny, bez ścian tekstu.

Jeśli jesteś Claude Code — czytasz to razem z `CLAUDE.md`. To nie sugestie, to reguły.

---

## Branching

**Tylko branch z prefiksem.** Hook `validate-branch-name.sh` + workflow `branch-name.yml` to wymuszają.

| Prefix      | Kiedy                                                   |
| ----------- | ------------------------------------------------------- |
| `feat/`     | Nowa funkcja lub ulepszenie                             |
| `fix/`      | Naprawa buga                                            |
| `chore/`    | Maintenance, deps bump, konfiguracja                    |
| `docs/`     | Wyłącznie zmiany w dokumentacji                         |
| `refactor/` | Zmiana struktury bez zmiany zachowania                  |
| `test/`     | Dodanie / poprawa testów (bez fixu produkcyjnego)       |
| `spike/`    | Timeboxed eksploracja — kod do potencjalnego wyrzucenia |

**Format:** `<prefix>/<lower-kebab-slug>[-<issue-number>]`

Przykłady:

- `feat/handle-empty-csv-42`
- `fix/retry-on-503-51`
- `chore/bump-duckdb`
- `spike/polars-vs-pandas`

Nieakceptowane: `dev`, `my-branch`, `Marcin/foo`, `FEAT/Foo`.

---

## Commits

**Conventional Commits.** Hook `validate-commit-msg.sh` + workflow `pr-title.yml` to wymuszają.

```
<type>(<scope>)?!?: <description>

[optional body]

[optional footer]
```

Typy: `feat | fix | chore | docs | refactor | test | perf | style | build | ci | revert`.

Reguły:

1. **Imperatyw, nie czas przeszły.** `add X`, nie `added X`.
2. **Subject ≤ 72 znaki.** Detal w body.
3. **`!` po typie/scope = breaking change.** `feat(api)!: rename endpoints`.
4. **Issue w footerze lub subject.** `Closes #42` (footer) lub `feat(parser): handle empty CSV (#42)` (subject).

Złe:

- `update code` — brak typu.
- `feat: stuff` — brak treści.
- `fix: fixed the bug` — co za bug?
- `WIP` — nie do mergowanej historii. Jeśli musisz, squashuj przed PR.

---

## Pull requests

**Każda zmiana w `main` przechodzi przez PR.** Push do `main` jest blokowany:

- Lokalnie: hook `block-main-push.sh`.
- Zdalnie: branch protection (skonfiguruj ręcznie po init, patrz README).

**PR title** = pierwszy commit subject (Conv. Commits). Workflow `pr-title.yml` to waliduje.

**PR body**: szablon w `.github/pull_request_template.md`. Wymagane sekcje:

- `Summary` — 1-3 zdania _co i dlaczego_.
- `Closes #N` — link do issue. Jeśli brak issue, wytłumacz dlaczego.
- `Acceptance criteria` — odhaczone = zrobione.
- `Tests` — co weryfikowane.
- `ADR` — wskaż lub zaznacz „nie dotyczy".

**Merge:**

- Zawsze **squash merge** (chyba że świadomie inaczej).
- Tytuł squash commitu = tytuł PR (zachowuje Conv. Commits).
- Delete branch after merge.

---

## Issues

Trzy szablony (`.github/ISSUE_TEMPLATE/`):

- **feature.yml** — nowa funkcja. Wymaga: Why, Acceptance criteria.
- **bug.yml** — bug. Wymaga: What/Expected/Repro/AC z regression test.
- **spike.yml** — timeboxed eksploracja. Wymaga: Question, Timebox, Deliverable.

Blank issues są wyłączone. Luźne pomysły → `/cx-idea` → `docs/ideas/`.

Issue gotowe do pracy ma:

- jasne acceptance criteria,
- pustą listę open questions (lub zamknięte odpowiedziami w komentarzach),
- jeden label typu (feat / fix / chore / spike).

---

## Documentation

**ADR (Architecture Decision Records)** w `docs/decisions/NNNN-tytuł.md`. Tworzy je `/cx-architecture` (agent `architect`, model Opus). Każdy ADR ma: kontekst, decyzję, ≥2 alternatywy, konsekwencje, reverse cost.

**Idea notes** w `docs/ideas/YYYY-MM-DD-slug.md`. Tworzy `/cx-idea`. To brudnopis — może zniknąć, gdy stanie się issue lub ADR-em.

**README, CHANGELOG, code comments** — pisz tylko gdy potrzebne. Nie generuj proaktywnie. Komentarz w kodzie = _dlaczego_, nie _co_.

---

## Protected branches

`main` jest chroniona. Skonfiguruj **po pierwszym push** w GitHub → Settings → Rules:

- ✅ Require a pull request before merging
- ✅ Require status checks: `branch-name`, `pr-title`, relevantne joby CI
- ✅ Require linear history (no merge commits)
- ✅ Block force pushes
- ✅ Restrict deletions
- ❌ Bypass list: puste (nawet admin nie pomija)

Drugi pas (lokalnie): `block-main-push.sh`.

---

## Dependencies

**Nowa dependency = ADR.** Bez wyjątków, nawet dla „małych" libów.

Każda dep ma koszt:

- maintenance burden (upgrade'y, breaking changes),
- powierzchnia ataku (CVE-y),
- nauka (jeden więcej rzecz do umienia w głowie).

`/cx-architecture` z opisem: _dlaczego stdlib nie wystarczy, jakie alternatywy, reverse cost_.

---

## Stop the line

**Jeśli hook blokuje — nie obchodź go.** Nigdy `--no-verify`. Hook ma rację. Jeśli się myli — popraw hook (osobny PR, label `chore`).

**Jeśli test failuje na main** — to incydent. Stop, znajdź root cause (skill `systematic-debugging`), naprawiaj.

**Jeśli zaproponowano destructive action** (`rm -rf`, `git reset --hard`, force push) — hook `dangerous-ops.sh` blokuje. Jeśli operacja jest naprawdę potrzebna — wykonaj ręcznie w terminalu po świadomej decyzji człowieka. Agent nie omija.

---

## Solo dev specifics

Tutaj nie ma zespołu. Więc:

- **„Approve" oznacza sam zatwierdziłem swój PR.** To OK. Ale ZAWSZE przeczytaj diff zanim klikniesz merge.
- **Nie ma stakeholdera do balansowania.** Priorytety to _co odblokowuje co_ + _co najbardziej boli_.
- **Estymacje nie są zobowiązaniem.** Są planem. Spike pozwala je urealnić.
- **Konsystencja w czasie > konsystencja w zespole.** „Przyszły ja za 6 miesięcy" to twój reviewer.

---

## Reference

- `CLAUDE.md` — kontekst projektu i reguły dla Claude Code.
- `.claude/agents/` — definicje sześciu agentów.
- `.claude/commands/` — sześć komend (`/cx-*`).
- `.claude/hooks/` — siedem hooków egzekwujących reguły.
- `.claude/skills/` — opcjonalne meta-skille (brainstorming, systematic-debugging, TDD, verification).
- `.github/` — workflows i issue/PR templates.
- `docs/decisions/` — ADR-y.
- `docs/ideas/` — notatki / brudnopis.
- `.solo-dev-version` — wersja frameworka tego projektu.

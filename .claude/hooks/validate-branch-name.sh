#!/usr/bin/env bash
# validate-branch-name.sh — PreToolUse hook for Bash
# Validates `git checkout -b <name>` and `git switch -c <name>` against allowed prefixes.
# Allowed: feat/ fix/ chore/ docs/ refactor/ test/ spike/
set -euo pipefail

INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")

# Match: `git checkout -b <name>` or `git switch -c <name>` (also -B for checkout)
BRANCH_NAME=""
if [[ "$CMD" =~ git[[:space:]]+checkout[[:space:]]+-[bB][[:space:]]+([^[:space:];|&]+) ]]; then
  BRANCH_NAME="${BASH_REMATCH[1]}"
elif [[ "$CMD" =~ git[[:space:]]+switch[[:space:]]+-c[[:space:]]+([^[:space:];|&]+) ]]; then
  BRANCH_NAME="${BASH_REMATCH[1]}"
elif [[ "$CMD" =~ git[[:space:]]+branch[[:space:]]+([^[:space:]-][^[:space:];|&]*) ]]; then
  BRANCH_NAME="${BASH_REMATCH[1]}"
else
  exit 0
fi

# Allowed prefixes
if [[ "$BRANCH_NAME" =~ ^(feat|fix|chore|docs|refactor|test|spike)/[a-z0-9][a-z0-9._-]*$ ]]; then
  exit 0
fi

# Special-cased branches that are always OK
case "$BRANCH_NAME" in
  main|master|develop) exit 0 ;;
esac

cat >&2 <<EOF
BLOCKED: branch name "$BRANCH_NAME" doesn't match required pattern.

Required prefix (one of): feat/ fix/ chore/ docs/ refactor/ test/ spike/
Format: <prefix>/<lower-kebab-slug>[-<issue-number>]

Examples:
  feat/handle-empty-csv-42
  fix/retry-on-503-51
  chore/bump-duckdb
  spike/polars-vs-pandas

See CONTRIBUTING.md → Branching.
EOF
exit 2

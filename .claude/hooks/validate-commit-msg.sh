#!/usr/bin/env bash
# validate-commit-msg.sh — PreToolUse hook for Bash
# Validates `git commit -m "..."` follows Conventional Commits.
# Pattern: <type>(<scope>)?: <description>  [optional body / footer]
set -euo pipefail

INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")

# Only act on `git commit -m "..."` (skip commits without -m, e.g. opens editor)
if [[ ! "$CMD" =~ git[[:space:]]+commit ]]; then
  exit 0
fi
if [[ ! "$CMD" =~ -m[[:space:]]+ ]]; then
  exit 0
fi

# Extract the message between -m " ... " or -m ' ... '
MSG=""
if [[ "$CMD" =~ -m[[:space:]]+\"([^\"]*)\" ]]; then
  MSG="${BASH_REMATCH[1]}"
elif [[ "$CMD" =~ -m[[:space:]]+\'([^\']*)\' ]]; then
  MSG="${BASH_REMATCH[1]}"
elif [[ "$CMD" =~ -m[[:space:]]+\$\(cat[[:space:]]*\<\<\'?EOF\'? ]]; then
  # heredoc form — extract content between EOF markers
  MSG=$(echo "$CMD" | awk '/<<.?EOF.?$/{flag=1;next}/^[[:space:]]*EOF[[:space:]]*$/{flag=0}flag' | head -1)
fi

if [[ -z "$MSG" ]]; then
  # Couldn't parse — be permissive, let git itself reject if needed
  exit 0
fi

# Conventional Commits regex (subject line):
# type(scope?)!?: description
# types: feat fix chore docs refactor test perf style build ci revert
SUBJECT="${MSG%%$'\n'*}"
PATTERN='^(feat|fix|chore|docs|refactor|test|perf|style|build|ci|revert)(\([a-z0-9._-]+\))?!?:[[:space:]].+'

if [[ "$SUBJECT" =~ $PATTERN ]]; then
  exit 0
fi

cat >&2 <<EOF
BLOCKED: commit message doesn't match Conventional Commits.

Subject was: "$SUBJECT"

Required format: <type>(<scope>)?: <description>
Types: feat | fix | chore | docs | refactor | test | perf | style | build | ci | revert

Examples:
  feat(parser): handle empty CSV (#42)
  fix: retry on 503
  chore(deps): bump duckdb to 0.10
  refactor(api)!: rename endpoints (BREAKING)

See CONTRIBUTING.md → Commits.
EOF
exit 2

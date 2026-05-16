#!/usr/bin/env bash
# session-context.sh — SessionStart hook
# Injects a brief git context into the session: branch, last 5 commits, dirty files.
# Output goes to stdout — Claude Code adds it to the session context.
set -euo pipefail

# Only run if we're in a git repo
if ! git rev-parse --git-dir >/dev/null 2>&1; then
  exit 0
fi

BRANCH=$({ git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "unborn"; } | head -1)
LAST_COMMITS=$(git log --oneline -5 2>/dev/null || echo "(no commits yet)")
[[ -z "$LAST_COMMITS" ]] && LAST_COMMITS="(no commits yet)"
DIRTY=$(git status --porcelain 2>/dev/null || echo "")

cat <<EOF
[session-context] git snapshot:

Branch: $BRANCH

Last 5 commits:
$LAST_COMMITS
EOF

if [[ -n "$DIRTY" ]]; then
  echo ""
  echo "Uncommitted changes:"
  echo "$DIRTY" | head -20
  LINES=$(echo "$DIRTY" | wc -l | tr -d ' ')
  if [[ "$LINES" -gt 20 ]]; then
    echo "... and $((LINES - 20)) more"
  fi
fi

# Framework version pointer (if present)
if [[ -f .solo-dev-version ]]; then
  VERSION=$(jq -r '.version // "unknown"' .solo-dev-version 2>/dev/null || echo "unknown")
  echo ""
  echo "Solo Dev framework version: $VERSION"
fi

exit 0

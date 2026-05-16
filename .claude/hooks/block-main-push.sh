#!/usr/bin/env bash
# block-main-push.sh — PreToolUse hook for Bash
# Blocks `git push` targeting the main/master branch.
# Branch protection on GitHub is the second line of defense; this catches it locally.
set -euo pipefail

INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")

# Only inspect git push commands
if [[ ! "$CMD" =~ ^[[:space:]]*git[[:space:]]+push ]]; then
  exit 0
fi

# Detect explicit push target: `git push <remote> main` or `git push origin HEAD:main`
# Also catches: `git push --force` to main
TARGET_BRANCH=$(echo "$CMD" | awk '{
  for (i=1; i<=NF; i++) {
    if ($i == "push") {
      # Look for branch ref in following args
      remote=""; branch="";
      for (j=i+1; j<=NF; j++) {
        if ($j ~ /^-/) continue
        if (remote == "") { remote=$j; continue }
        if (branch == "") { branch=$j; break }
      }
      print branch
      exit
    }
  }
}')

# If no explicit branch given, push goes to current branch — check that
if [[ -z "$TARGET_BRANCH" ]]; then
  CURRENT=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
  TARGET_BRANCH="$CURRENT"
fi

# Strip `HEAD:` prefix if present
TARGET_BRANCH="${TARGET_BRANCH#HEAD:}"

case "$TARGET_BRANCH" in
  main|master)
    echo "BLOCKED: direct push to '$TARGET_BRANCH' is not allowed." >&2
    echo "Use a feature branch and open a PR. See CONTRIBUTING.md." >&2
    exit 2
    ;;
esac

exit 0

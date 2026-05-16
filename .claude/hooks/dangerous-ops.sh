#!/usr/bin/env bash
# dangerous-ops.sh — PreToolUse hook for Bash
# Blocks destructive operations that have no easy reverse.
# Patterns: rm -rf /, git reset --hard, git push --force, git clean -fdx, etc.
# Exit 2 → blocked with feedback to the model.
set -euo pipefail

INPUT=$(cat)
CMD=$(jq -r '.tool_input.command // empty' <<<"$INPUT")

if [[ -z "$CMD" ]]; then
  exit 0
fi

# Helper to fail with a uniform message
deny() {
  cat >&2 <<EOF
BLOCKED: dangerous operation detected.

Command: $CMD
Reason: $1

If you genuinely need to do this, ask the user explicitly with full context.
Do NOT bypass with --no-verify, env hacks, or splitting the command.
EOF
  exit 2
}

# 1. rm -rf on / or $HOME or current dir
if [[ "$CMD" =~ rm[[:space:]]+(-[a-zA-Z]*r[a-zA-Z]*f|-rf|-fr)[a-zA-Z]*[[:space:]]+(/|/\*|~|\$HOME|\.|\./?\*?)[[:space:]]*($|[\;\|\&]) ]]; then
  deny "rm -rf targeting root, home, or current directory"
fi
# Generic rm -rf with no path is suspicious
if [[ "$CMD" =~ rm[[:space:]]+-rf?[[:space:]]+\* ]]; then
  deny "rm -rf with glob — too broad"
fi

# 2. git reset --hard / git checkout -- . / git restore .
if [[ "$CMD" =~ git[[:space:]]+reset[[:space:]]+--hard ]]; then
  deny "git reset --hard discards uncommitted changes irreversibly"
fi
if [[ "$CMD" =~ git[[:space:]]+(checkout|restore)[[:space:]]+(--[[:space:]]+)?\. ]]; then
  deny "git checkout/restore . discards uncommitted changes"
fi

# 3. git clean -fd / -fdx
if [[ "$CMD" =~ git[[:space:]]+clean[[:space:]]+-[a-z]*f ]]; then
  deny "git clean -f deletes untracked files"
fi

# 4. git push --force / -f to any branch (covers main via block-main-push.sh too)
if [[ "$CMD" =~ git[[:space:]]+push[[:space:]]+.*(--force|-f([[:space:]]|$)) ]]; then
  # --force-with-lease is OK
  if [[ ! "$CMD" =~ --force-with-lease ]]; then
    deny "git push --force can overwrite upstream history; use --force-with-lease if needed"
  fi
fi

# 5. git branch -D (force-delete) on protected branches
if [[ "$CMD" =~ git[[:space:]]+branch[[:space:]]+-D[[:space:]]+(main|master|develop) ]]; then
  deny "force-deleting protected branch"
fi

# 6. SQL DROP DATABASE / TRUNCATE on production-looking DBs
if [[ "$CMD" =~ (DROP[[:space:]]+DATABASE|TRUNCATE[[:space:]]+TABLE) ]]; then
  deny "destructive SQL — confirm with user explicitly"
fi

# 7. gh repo delete
if [[ "$CMD" =~ gh[[:space:]]+repo[[:space:]]+delete ]]; then
  deny "gh repo delete is irreversible — confirm with user"
fi

# 8. dd / mkfs / shred on devices
if [[ "$CMD" =~ (^|[[:space:]])(dd|mkfs|shred)([[:space:]]|$) ]]; then
  deny "low-level disk command"
fi

exit 0

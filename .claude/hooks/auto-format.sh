#!/usr/bin/env bash
# auto-format.sh — PostToolUse hook for Edit/Write
# Runs the appropriate formatter for the just-edited file.
# Silent on success; warns to stdout on failure but does NOT block.
set -euo pipefail

INPUT=$(cat)
FILE=$(jq -r '.tool_input.file_path // empty' <<<"$INPUT")

if [[ -z "$FILE" || ! -f "$FILE" ]]; then
  exit 0
fi

# Pick formatter by extension. Skip silently if formatter not installed.
case "$FILE" in
  *.go)
    if command -v gofmt >/dev/null 2>&1; then
      gofmt -w "$FILE" 2>/dev/null || true
    fi
    if command -v goimports >/dev/null 2>&1; then
      goimports -w "$FILE" 2>/dev/null || true
    fi
    ;;
  *.py)
    if command -v ruff >/dev/null 2>&1; then
      ruff format "$FILE" 2>/dev/null || true
    elif command -v black >/dev/null 2>&1; then
      black --quiet "$FILE" 2>/dev/null || true
    fi
    ;;
  *.ts|*.tsx|*.js|*.jsx|*.json|*.md|*.yml|*.yaml|*.css|*.scss|*.html)
    if command -v prettier >/dev/null 2>&1; then
      prettier --write --log-level=silent "$FILE" 2>/dev/null || true
    fi
    ;;
  *.cs)
    # dotnet format requires a project context — only run if csproj nearby
    DIR=$(dirname "$FILE")
    if command -v dotnet >/dev/null 2>&1 && find "$DIR" -maxdepth 3 -name "*.csproj" 2>/dev/null | head -1 | grep -q .; then
      dotnet format --include "$FILE" >/dev/null 2>&1 || true
    fi
    ;;
  *.sh|*.bash)
    if command -v shfmt >/dev/null 2>&1; then
      shfmt -w "$FILE" 2>/dev/null || true
    fi
    ;;
esac

exit 0

#!/usr/bin/env bash
# lint-feedback.sh — PostToolUse hook for Edit/Write
# Runs linter on the edited file and surfaces warnings to the model (stdout).
# Does NOT block — feedback only.
set -euo pipefail

INPUT=$(cat)
FILE=$(jq -r '.tool_input.file_path // empty' <<<"$INPUT")

if [[ -z "$FILE" || ! -f "$FILE" ]]; then
  exit 0
fi

OUT=""

case "$FILE" in
  *.go)
    if command -v golangci-lint >/dev/null 2>&1; then
      OUT=$(golangci-lint run --fast --out-format=line-number "$FILE" 2>&1 || true)
    elif command -v go >/dev/null 2>&1; then
      OUT=$(go vet "$FILE" 2>&1 || true)
    fi
    ;;
  *.py)
    if command -v ruff >/dev/null 2>&1; then
      OUT=$(ruff check --output-format=concise "$FILE" 2>&1 || true)
    fi
    ;;
  *.ts|*.tsx|*.js|*.jsx)
    if command -v eslint >/dev/null 2>&1; then
      OUT=$(eslint --no-color "$FILE" 2>&1 || true)
    fi
    ;;
  *.cs)
    # No lightweight linter — rely on dotnet build/format separately
    ;;
  *.sh|*.bash)
    if command -v shellcheck >/dev/null 2>&1; then
      OUT=$(shellcheck -f gcc "$FILE" 2>&1 || true)
    fi
    ;;
esac

# Filter out empty / OK lines. If anything left → surface as model feedback.
TRIMMED=$(echo "$OUT" | grep -v '^[[:space:]]*$' || true)
if [[ -n "$TRIMMED" ]]; then
  cat <<EOF
[lint-feedback] $FILE has lint warnings:

$TRIMMED

(non-blocking — fix only if relevant to current change)
EOF
fi

exit 0

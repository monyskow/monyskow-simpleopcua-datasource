#!/bin/bash
# Interactive dev launcher: pick (or pass as args) a Grafana version and auth
# config, then bring up docker-compose.e2e.yaml for manual plugin development.
# Useful when you want a specific combo without memorizing env var names.

set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Versions from the single source of truth. while-read loop instead of
# `mapfile` for bash 3.2 (macOS default) compatibility.
VERSIONS=()
while IFS= read -r line || [ -n "$line" ]; do
  [ -n "$line" ] && VERSIONS+=("$line")
done <"$REPO_ROOT/.grafana-versions"

# Auth configs match the 5 provisioning files at
# provisioning/datasources/datasources-e2e-*.yaml
AUTH_CONFIGS=(anon-none userpass-none userpass-b256-sign cert-b256-sign cert-aes256-sign)

# --- Version selection ---

if [ -n "$1" ]; then
  # Validate the provided version against the known list
  VALID_VERSION=0
  for v in "${VERSIONS[@]}"; do
    [ "$v" = "$1" ] && VALID_VERSION=1 && break
  done

  if [ $VALID_VERSION -eq 0 ]; then
    echo "Error: unknown Grafana version '$1'"
    echo "Valid versions:"
    printf '  %s\n' "${VERSIONS[@]}"
    exit 1
  fi

  VERSION="$1"
else
  # Interactive picker — requires a terminal; not meant for CI use
  echo "Select Grafana version:"
  select v in "${VERSIONS[@]}"; do
    [ -n "$v" ] && VERSION="$v" && break
    echo "Invalid selection, try again."
  done
fi

# EOF/Ctrl-D on `select` exits the loop without assignment. Without this guard
# the compose call would silently fall back to the yaml's default version.
[ -z "$VERSION" ] && { echo "No version selected." >&2; exit 1; }

# --- Auth config selection ---

if [ -n "$2" ]; then
  # Validate the provided auth config against the known list
  VALID_AUTH=0
  for a in "${AUTH_CONFIGS[@]}"; do
    [ "$a" = "$2" ] && VALID_AUTH=1 && break
  done

  if [ $VALID_AUTH -eq 0 ]; then
    echo "Error: unknown auth config '$2'"
    echo "Valid auth configs:"
    printf '  %s\n' "${AUTH_CONFIGS[@]}"
    exit 1
  fi

  AUTH="$2"
else
  echo "Select auth config:"
  select a in "${AUTH_CONFIGS[@]}"; do
    [ -n "$a" ] && AUTH="$a" && break
    echo "Invalid selection, try again."
  done
fi

[ -z "$AUTH" ] && { echo "No auth config selected." >&2; exit 1; }

# --- Bring up the stack ---

echo "Starting Grafana $VERSION with auth config $AUTH..."

GRAFANA_VERSION=$VERSION AUTH_CONFIG=$AUTH \
  docker compose -f "$REPO_ROOT/docker-compose.e2e.yaml" up -d --wait --build

echo "Grafana ready at http://localhost:3000"

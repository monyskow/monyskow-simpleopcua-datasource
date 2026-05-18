#!/bin/bash
# Run E2E tests against the full (Grafana version × auth config) matrix locally.
# Mirrors what `.github/workflows/e2e-matrix.yml` runs in CI, but sequentially
# on the dev machine. Use sparingly — ~3-4 min per combo × 30 combos = ~90-120 min.

set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"

# Versions from the single source of truth. while-read loop instead of `mapfile`
# for bash 3.2 (macOS default) compatibility.
VERSIONS=()
while IFS= read -r line || [ -n "$line" ]; do
  [ -n "$line" ] && VERSIONS+=("$line")
done <"$REPO_ROOT/.grafana-versions"

# Auth configs match the 5 provisioning files at provisioning/datasources/datasources-e2e-*.yaml
AUTH_CONFIGS=(anon-none userpass-none userpass-b256-sign cert-b256-sign cert-aes256-sign)

RESULTS_FILE="$REPO_ROOT/e2e-matrix-results-$(date +%Y%m%d-%H%M%S).txt"
RESULTS=()
FAILED=0

TOTAL=$((${#VERSIONS[@]} * ${#AUTH_CONFIGS[@]}))
echo "=========================================="
echo "E2E matrix: ${#VERSIONS[@]} versions × ${#AUTH_CONFIGS[@]} auth configs = $TOTAL combos"
echo "Results: $RESULTS_FILE"
echo "=========================================="

# Build the plugin once before the loop
echo "Building plugin..."
npm run build
mage -v buildAll
echo ""

CURRENT=0
for VERSION in "${VERSIONS[@]}"; do
  for AUTH in "${AUTH_CONFIGS[@]}"; do
    CURRENT=$((CURRENT + 1))
    echo ""
    echo "=========================================="
    echo "[$CURRENT/$TOTAL] $VERSION / $AUTH"
    echo "=========================================="

    # Clean state from previous combo
    docker compose -f docker-compose.e2e.yaml down --remove-orphans >/dev/null 2>&1 || true
    mkdir -p playwright/.auth
    rm -f playwright/.auth/*.json 2>/dev/null || true

    GRAFANA_VERSION=$VERSION AUTH_CONFIG=$AUTH \
      docker compose -f docker-compose.e2e.yaml up -d --wait --build

    set +e
    CI=1 npm run e2e
    EXIT=$?
    set -e

    if [ $EXIT -eq 0 ]; then
      RESULTS+=("$VERSION / $AUTH: PASS")
      echo "[$CURRENT/$TOTAL] $VERSION / $AUTH: PASS"
    else
      RESULTS+=("$VERSION / $AUTH: FAIL")
      echo "[$CURRENT/$TOTAL] $VERSION / $AUTH: FAIL"
      FAILED=1
    fi

    docker compose -f docker-compose.e2e.yaml down --remove-orphans >/dev/null 2>&1 || true
  done
done

# Summary
{
  echo "E2E Matrix Results - $(date)"
  echo "========================================"
  printf '%s\n' "${RESULTS[@]}"
  echo "========================================"
  if [ $FAILED -eq 0 ]; then
    echo "Overall: ALL PASSED ($TOTAL/$TOTAL)"
  else
    PASSED=$((TOTAL - $(printf '%s\n' "${RESULTS[@]}" | grep -c FAIL)))
    echo "Overall: SOME FAILED ($PASSED/$TOTAL passed)"
  fi
} | tee "$RESULTS_FILE"

exit $FAILED

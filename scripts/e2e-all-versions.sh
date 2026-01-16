#!/bin/bash
# Run E2E tests against all Grafana versions (same as CI)

set -e

VERSIONS=(
  "10.4.19"
  "11.1.13"
  "11.4.8"
  "12.1.5"
  "12.3.1"
)

RESULTS=()
FAILED=0

# Results file with timestamp
RESULTS_FILE="e2e-results-$(date +%Y%m%d-%H%M%S).txt"

echo "=========================================="
echo "Running E2E tests against ${#VERSIONS[@]} Grafana versions"
echo "Results will be saved to: $RESULTS_FILE"
echo "=========================================="

# Initialize results file
echo "E2E Test Results - $(date)" > "$RESULTS_FILE"
echo "========================================" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Build the plugin once before testing
echo "Building plugin..."
echo "Building frontend..."
npm run build
echo "Building backend..."
mage -v buildAll
echo "Plugin build complete"
echo ""

for VERSION in "${VERSIONS[@]}"; do
  echo ""
  echo "=========================================="
  echo "Testing Grafana $VERSION"
  echo "=========================================="

  # Write to results file
  echo "" >> "$RESULTS_FILE"
  echo "========================================" >> "$RESULTS_FILE"
  echo "Testing Grafana $VERSION" >> "$RESULTS_FILE"
  echo "========================================" >> "$RESULTS_FILE"

  # Stop any existing containers
  docker compose down --remove-orphans 2>/dev/null || true

  # Ensure auth directory exists and clean up auth state from previous Grafana version
  mkdir -p playwright/.auth
  rm -f playwright/.auth/*.json 2>/dev/null || true

  # Start Grafana with specific version
  GRAFANA_VERSION=$VERSION GRAFANA_IMAGE=grafana-enterprise ANONYMOUS_AUTH_ENABLED=false docker compose up -d --build

  # Wait for Grafana to be ready
  echo "Waiting for Grafana $VERSION to start..."
  for i in {1..60}; do
    if curl -s http://localhost:3000/api/health | grep -q "ok"; then
      echo "Grafana $VERSION is ready"
      break
    fi
    sleep 2
  done

  # Give plugin additional time to load and register
  echo "Waiting for plugin to initialize..."
  sleep 10

  # Run tests
  echo "Running E2E tests..."
  echo "Running E2E tests..." >> "$RESULTS_FILE"
  echo "" >> "$RESULTS_FILE"

  set +e  # Temporarily disable exit on error
  CI=true npm run e2e 2>&1 | tee -a "$RESULTS_FILE"
  TEST_EXIT_CODE=${PIPESTATUS[0]}
  set -e  # Re-enable exit on error

  echo "" >> "$RESULTS_FILE"
  if [ $TEST_EXIT_CODE -eq 0 ]; then
    RESULTS+=("$VERSION: PASSED")
    echo "Grafana $VERSION: PASSED"
    echo "Result: PASSED" >> "$RESULTS_FILE"
  else
    RESULTS+=("$VERSION: FAILED")
    echo "Grafana $VERSION: FAILED"
    echo "Result: FAILED" >> "$RESULTS_FILE"
    FAILED=1
  fi

  # Stop Grafana
  docker compose down --remove-orphans
done

echo ""
echo "=========================================="
echo "SUMMARY"
echo "=========================================="
for RESULT in "${RESULTS[@]}"; do
  echo "$RESULT"
done
echo "=========================================="

# Write summary to file
echo "" >> "$RESULTS_FILE"
echo "========================================" >> "$RESULTS_FILE"
echo "SUMMARY" >> "$RESULTS_FILE"
echo "========================================" >> "$RESULTS_FILE"
for RESULT in "${RESULTS[@]}"; do
  echo "$RESULT" >> "$RESULTS_FILE"
done
echo "========================================" >> "$RESULTS_FILE"

# Add final status
if [ $FAILED -eq 0 ]; then
  echo "" >> "$RESULTS_FILE"
  echo "Overall Status: ALL TESTS PASSED" >> "$RESULTS_FILE"
else
  echo "" >> "$RESULTS_FILE"
  echo "Overall Status: SOME TESTS FAILED" >> "$RESULTS_FILE"
fi

echo ""
echo "Results saved to: $RESULTS_FILE"

exit $FAILED

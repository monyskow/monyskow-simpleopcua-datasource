#!/bin/bash
# Stop all Grafana test containers

# Versions are read from .grafana-versions at repo root.
# Lines 1-5: manually curated anchors. Line 6: latest-stable slot,
# auto-bumped weekly by .github/workflows/bump-grafana-latest.yml
mapfile -t VERSIONS < "$(git rev-parse --show-toplevel)/.grafana-versions"

echo "Stopping all Grafana test containers..."

for VERSION in "${VERSIONS[@]}"; do
  CONTAINER_NAME="grafana-${VERSION}"
  if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    docker rm -f "$CONTAINER_NAME" > /dev/null 2>&1
    echo "  ✓ Stopped $CONTAINER_NAME"
  fi
done

# Clean up network
docker network rm grafana-test-network 2>/dev/null || true

echo ""
echo "All Grafana test containers stopped."

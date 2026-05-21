#!/bin/bash
# Stop all Grafana test containers

# Versions are read from .grafana-versions at repo root (manually maintained).
# while-read loop instead of `mapfile` for bash 3.2 (macOS default) compatibility.
VERSIONS=()
while IFS= read -r line || [ -n "$line" ]; do
  [ -n "$line" ] && VERSIONS+=("$line")
done < "$(git rev-parse --show-toplevel)/.grafana-versions"

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

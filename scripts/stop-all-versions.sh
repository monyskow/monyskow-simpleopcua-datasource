#!/bin/bash
# Stop all Grafana test containers

VERSIONS=(
  "10.4.19"
  "11.1.13"
  "11.4.8"
  "12.1.5"
  "12.3.1"
)

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

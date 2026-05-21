#!/bin/bash
# Start all Grafana versions simultaneously for manual testing
# Each version runs on a different port

set -e

# Versions are read from .grafana-versions at repo root (manually maintained).
# while-read loop instead of `mapfile` for bash 3.2 (macOS default) compatibility.
VERSIONS=()
while IFS= read -r line || [ -n "$line" ]; do
  [ -n "$line" ] && VERSIONS+=("$line")
done < "$(git rev-parse --show-toplevel)/.grafana-versions"

# Base port - each version will use BASE_PORT + index
BASE_PORT=3000

echo "=========================================="
echo "Starting ${#VERSIONS[@]} Grafana versions for manual testing"
echo "=========================================="

# Build the plugin first
echo ""
echo "Building plugin..."
echo "Building frontend..."
npm run build
echo "Building backend..."
mage -v buildAll
echo "Plugin build complete"
echo ""

# Stop any existing containers from previous runs
echo "Cleaning up existing containers..."
for i in "${!VERSIONS[@]}"; do
  docker rm -f "grafana-${VERSIONS[$i]}" 2>/dev/null || true
done
docker network rm grafana-test-network 2>/dev/null || true

# Create a shared network
docker network create grafana-test-network 2>/dev/null || true

echo ""
echo "Starting Grafana containers..."
echo ""

# Start each version on a different port
for i in "${!VERSIONS[@]}"; do
  VERSION="${VERSIONS[$i]}"
  PORT=$((BASE_PORT + i))
  CONTAINER_NAME="grafana-${VERSION}"

  echo "Starting Grafana $VERSION on port $PORT..."

  docker run -d \
    --name "$CONTAINER_NAME" \
    --network grafana-test-network \
    -p "${PORT}:3000" \
    -v "$(pwd)/dist:/var/lib/grafana/plugins/monyskow-simpleopcua-datasource:ro" \
    -v "$(pwd)/provisioning:/etc/grafana/provisioning:ro" \
    -e GF_LOG_LEVEL=debug \
    -e GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=monyskow-simpleopcua-datasource \
    -e GF_AUTH_ANONYMOUS_ENABLED=false \
    -e GF_SECURITY_ADMIN_USER=admin \
    -e GF_SECURITY_ADMIN_PASSWORD=admin \
    "grafana/grafana-enterprise:${VERSION}" \
    > /dev/null

  echo "  ✓ Grafana $VERSION starting on http://localhost:$PORT"
done

echo ""
echo "Waiting for all instances to be ready..."
echo ""

# Wait for all instances to be healthy
for i in "${!VERSIONS[@]}"; do
  VERSION="${VERSIONS[$i]}"
  PORT=$((BASE_PORT + i))

  echo -n "Waiting for Grafana $VERSION (port $PORT)..."
  for j in {1..30}; do
    if curl -s "http://localhost:${PORT}/api/health" | grep -q "ok"; then
      echo " Ready!"
      break
    fi
    sleep 2
    echo -n "."
  done
done

echo ""
echo "=========================================="
echo "All Grafana instances are running!"
echo "=========================================="
echo ""
echo "Access URLs (login: admin / admin):"
echo ""
for i in "${!VERSIONS[@]}"; do
  VERSION="${VERSIONS[$i]}"
  PORT=$((BASE_PORT + i))
  echo "  Grafana $VERSION: http://localhost:$PORT"
done
echo ""
echo "=========================================="
echo "Useful pages to test:"
echo "=========================================="
echo ""
echo "For each version, try:"
echo "  - Plugin page:      /plugins/monyskow-simpleopcua-datasource"
echo "  - Data sources:     /connections/datasources"
echo "  - Config page:      /connections/datasources/edit/opcua-test-server"
echo "  - Explore:          /explore"
echo ""
echo "=========================================="
echo "To stop all instances, run:"
echo "  ./scripts/stop-all-versions.sh"
echo "=========================================="

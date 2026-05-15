#!/usr/bin/env bash
# Generate test client certificate/key for ProSys provisioning and render the
# datasources YAML from its template. All outputs are gitignored — run this
# before `docker compose -f docker-compose.prosys.yaml up`.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CERT_DIR="$HERE/provisioning/certs"
TEMPLATE="$HERE/provisioning/datasources/datasources-prosys.yaml.template"
OUTPUT="$HERE/provisioning/datasources/datasources-prosys.yaml"

CERT="$CERT_DIR/client-cert.pem"
KEY="$CERT_DIR/client-key.pem"

mkdir -p "$CERT_DIR"

if [[ ! -f "$CERT" || ! -f "$KEY" ]]; then
  echo "Generating test client certificate (10y validity) in $CERT_DIR..." >&2
  openssl req -x509 -newkey rsa:2048 -nodes \
    -keyout "$KEY" \
    -out "$CERT" \
    -days 3650 \
    -subj "/C=US/O=Grafana/CN=Grafana Simple OPC-UA Client (Test)" \
    -addext "subjectAltName=DNS:grafana-opcua-client,DNS:localhost,IP:127.0.0.1,IP:0:0:0:0:0:0:0:1,URI:urn:grafana-opcua-client:grafana:simpleopcua:client" \
    -addext "keyUsage=critical,digitalSignature,keyEncipherment,dataEncipherment" \
    -addext "extendedKeyUsage=clientAuth,serverAuth" \
    >/dev/null 2>&1
  chmod 600 "$KEY"
fi

if [[ ! -f "$TEMPLATE" ]]; then
  echo "Missing template: $TEMPLATE" >&2
  exit 1
fi

awk -v cert="$CERT" -v key="$KEY" '
  function emit(path,   line) {
    while ((getline line < path) > 0) print "  " line
    close(path)
  }
  /__CLIENT_CERT__/ { emit(cert); next }
  /__CLIENT_KEY__/  { emit(key);  next }
  { print }
' "$TEMPLATE" > "$OUTPUT"

echo "Wrote $OUTPUT" >&2

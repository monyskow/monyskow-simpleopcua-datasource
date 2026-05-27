# Development

How to build, run, and contribute to the Simple OPC-UA datasource plugin.

## Prerequisites

- Node.js 22+
- Go 1.22+
- Docker (for testing)

## Build and Run

```bash
# Install and build
npm install
npm run build
mage buildAll

# Start development environment (choose one):

# 1. Basic - Grafana + plugin (no provisioning)
docker compose up

# 2. Full - Complete test environment with OPC-UA simulators
docker compose -f docker-compose.full.yaml up

# 3. ProSys - For ProSys OPC-UA Simulator testing (requires ProSys on host)
docker compose -f docker-compose.prosys.yaml up

# Open http://localhost:3000 (or :3001 for ProSys)
```

## Docker Compose Configurations

| Configuration                | Datasources       | Dashboards | OPC-UA Servers | Port | Use Case                          |
| ---------------------------- | ----------------- | ---------- | -------------- | ---- | --------------------------------- |
| `docker-compose.yaml`        | None              | None       | None           | 3000 | Manual testing, clean slate       |
| `docker-compose.e2e.yaml`    | 1 (test)          | None       | None           | 3000 | E2E tests (used by `npm run e2e`) |
| `docker-compose.full.yaml`   | 14 (docker-based) | Yes        | 5 containers   | 3000 | Complete integration testing      |
| `docker-compose.prosys.yaml` | 14 (ProSys)       | Yes        | ProSys on host | 3001 | ProSys simulator testing          |

## OPC-UA Test Servers

### Docker-based Simulators (Included)

The `docker-compose.full.yaml` setup includes 5 containerized OPC-UA test servers using [Microsoft OPC-PLC](https://github.com/Azure-Samples/iot-edge-opc-plc) and [node-opcua](https://github.com/node-opcua/node-opcua). These servers are provided **solely for testing convenience** and are not part of the plugin itself.

**Available servers:**

- `opcua-nosecurity` (port 50000) - No security, anonymous only
- `opcua-secure-anon` (port 50001) - Security enabled, anonymous auth
- `opcua-secure-userpass` (port 50002) - Security enabled, username/password (user1/password1)
- `opcua-all-auth` (port 50003) - All authentication methods enabled
- `opcua-node` (port 4840) - Node-OPCUA server with Aes256_Sha256_RsaPss support

### ProSys OPC-UA Simulator (External)

[ProSys OPC-UA Simulation Server](https://www.prosysopc.com/products/opc-ua-simulation-server/) is a **third-party external tool** that can be used for testing. It is not included with this plugin.

To test with ProSys:

1. Download and install ProSys OPC-UA Simulation Server on your host machine
2. Start the ProSys simulator
3. Use `docker compose -f docker-compose.prosys.yaml up` to run Grafana configured to connect to ProSys via `host.docker.internal`

**Note:** ProSys is a separate commercial product with its own licensing. See the [ProSys website](https://www.prosysopc.com/) for details.

## Additional Commands

```bash
# Linting and validation
npm run lint               # ESLint
npm run lint:fix           # Auto-fix + Prettier
npm run typecheck          # TypeScript check

# Development mode
npm run dev                # Frontend watch mode
```

## Testing

```bash
# Unit tests
npm run test:ci              # Frontend
mage test                    # Backend

# E2E tests (against multiple Grafana versions)
npm run e2e                  # Run all tests
npx playwright test --ui     # Interactive mode

# Test specific Grafana version
GRAFANA_VERSION=10.4.0 npm run server
```

For full E2E testing documentation see [testing.md](testing.md).

## Multi-version Testing

The following scripts test the plugin against all supported Grafana versions. They build the plugin once, then iterate over each version.

- `scripts/e2e-all-versions.sh` — builds the plugin, starts each Grafana version sequentially via `docker compose`, runs `npm run e2e` against it, and writes a timestamped results file.
- `scripts/start-all-versions.sh` — builds the plugin, then starts all versions simultaneously as individual Docker containers on consecutive ports starting at 3000 for side-by-side manual comparison.
- `scripts/stop-all-versions.sh` — stops and removes all containers started by `start-all-versions.sh`.

## Architecture

**Frontend (`src/`):**

- `module.ts` - Plugin entry point
- `datasource.ts` - OpcuaDataSource (extends DataSourceWithBackend)
- `components/ConfigEditor/` - Data source configuration UI
- `components/QueryEditor/` - Query editor with node browser

**Backend (`pkg/`):**

- `plugin/datasource.go` - Query and health check handlers
- `plugin/resources.go` - HTTP handlers for /browse and /endpoints
- `plugin/opcua/` - OPC-UA client, auth, browsing, certificates

**Key Dependencies:**

- Go: `github.com/gopcua/opcua` - OPC-UA protocol
- Go: `github.com/grafana/grafana-plugin-sdk-go` - Plugin SDK

## Pre-PR Checklist

`lint:fix` → `typecheck` → `test:ci` → `mage test` → `mage buildAll` → `npm run build` → at minimum `npm run e2e` (matrix runs in CI).

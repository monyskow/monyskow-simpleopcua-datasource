# Simple OPC-UA Data Source for Grafana

A Grafana data source plugin for connecting to OPC-UA servers. Read industrial data directly in Grafana dashboards.

## Features

- **Easy Configuration**: Simple setup with endpoint URL, security settings, and authentication
- **Multiple Auth Methods**: Anonymous, Username/Password, and Certificate authentication
- **Node Browser**: Graphical tree browser for exploring OPC-UA address space
- **Template Variables**: Full support for Grafana template variables in Node IDs
- **Health Checks**: Built-in connection testing via Save & Test

## Requirements

- Grafana 10.4.0 or later
- Access to an OPC-UA server

## Installation

### From Grafana Plugin Catalog

1. In Grafana, go to **Configuration > Plugins**
2. Search for "Simple OPC-UA"
3. Click **Install**

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/monyskow/monyskow-simpleopcua-datasource/releases)
2. Extract to your Grafana plugins directory (usually `/var/lib/grafana/plugins`)
3. Restart Grafana

## Configuration

### Connection Settings

| Setting         | Description                  | Example                    |
| --------------- | ---------------------------- | -------------------------- |
| Endpoint URL    | OPC-UA server endpoint       | `opc.tcp://localhost:4840` |
| Security Policy | Encryption algorithm         | None, Basic256Sha256       |
| Security Mode   | Message security             | None, Sign, SignAndEncrypt |
| Timeout         | Connection timeout (seconds) | 10                         |

### Security

When using security modes other than "None" (Sign or SignAndEncrypt), the plugin provides two options:

#### Option 1: Auto-Generated Certificate (Recommended)

The plugin can automatically generate and manage client certificates:

1. Click **"Generate Certificate"** button in the datasource configuration
2. The plugin generates a self-signed client certificate
3. Click **"Save & Test"** to persist the certificate
4. The certificate is securely stored in Grafana's encrypted storage

The auto-generated certificate is valid for 3 years and persists across Grafana restarts.

#### Option 2: Provide Your Own Certificate

For advanced use cases, you can provide your own client certificate using the Certificate authentication method (see Authentication section below).

**Important:** For secure connections, you must either generate a certificate or provide your own. The OPC-UA server must trust the client certificate - consult your OPC-UA server documentation for adding trusted certificates.

**Supported Security Policies:**

- None (no encryption)
- Basic256Sha256
- Aes128_Sha256_RsaOaep
- Aes256_Sha256_RsaPss

**Supported Security Modes:**

- None
- Sign (messages are signed but not encrypted)
- SignAndEncrypt (messages are signed and encrypted)

### Authentication

#### Anonymous

No credentials required. Select "Anonymous" as the authentication method.

#### Username/Password

1. Select "Username / Password" as the authentication method
2. Enter your username and password
3. Credentials are stored securely using Grafana's secure JSON data

#### Certificate

For advanced use cases where you need to use a specific client certificate:

1. Select "Certificate" as the authentication method
2. Paste your PEM-encoded certificate and private key
3. Both are stored securely using Grafana's secure JSON data

Note: When using Certificate authentication, the provided certificate is used instead of the auto-generated one.

## Usage

### Building Queries

1. Add a new panel and select the OPC-UA data source
2. Click "Browse Nodes" to explore the OPC-UA address space
3. Click the + icon next to any Variable node to add it to your query
4. Optionally, set an alias for each node

### Using Template Variables

You can use Grafana template variables in Node IDs:

```
ns=2;s=${machine}/Temperature
```

## Development

### Prerequisites

- Node.js 22+
- Go 1.22+
- Docker (for testing)

### Quick Start

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

**Docker Compose Configurations:**

| Configuration                | Datasources       | Dashboards | OPC-UA Servers | Port | Use Case                          |
| ---------------------------- | ----------------- | ---------- | -------------- | ---- | --------------------------------- |
| `docker-compose.yaml`        | None              | None       | None           | 3000 | Manual testing, clean slate       |
| `docker-compose.e2e.yaml`    | 1 (test)          | None       | None           | 3000 | E2E tests (used by `npm run e2e`) |
| `docker-compose.full.yaml`   | 14 (docker-based) | Yes        | 5 containers   | 3000 | Complete integration testing      |
| `docker-compose.prosys.yaml` | 14 (ProSys)       | Yes        | ProSys on host | 3001 | ProSys simulator testing          |

### OPC-UA Test Servers

For development and testing, you can use OPC-UA simulators:

#### Docker-based Simulators (Included)

The `docker-compose.full.yaml` setup includes 5 containerized OPC-UA test servers using [Microsoft OPC-PLC](https://github.com/Azure-Samples/iot-edge-opc-plc) and [node-opcua](https://github.com/node-opcua/node-opcua). These servers are provided **solely for testing convenience** and are not part of the plugin itself.

**Available servers:**

- `opcua-nosecurity` (port 50000) - No security, anonymous only
- `opcua-secure-anon` (port 50001) - Security enabled, anonymous auth
- `opcua-secure-userpass` (port 50002) - Security enabled, username/password (user1/password1)
- `opcua-all-auth` (port 50003) - All authentication methods enabled
- `opcua-node` (port 4840) - Node-OPCUA server with Aes256_Sha256_RsaPss support

#### ProSys OPC-UA Simulator (External)

[ProSys OPC-UA Simulation Server](https://www.prosysopc.com/products/opc-ua-simulation-server/) is a **third-party external tool** that can be used for testing. It is not included with this plugin.

To test with ProSys:

1. Download and install ProSys OPC-UA Simulation Server on your host machine
2. Start the ProSys simulator
3. Use `docker compose -f docker-compose.prosys.yaml up` to run Grafana configured to connect to ProSys via `host.docker.internal`

**Note:** ProSys is a separate commercial product with its own licensing. See the [ProSys website](https://www.prosysopc.com/) for details.

### Additional Commands

```bash
# Linting and validation
npm run lint               # ESLint
npm run lint:fix           # Auto-fix + Prettier
npm run typecheck          # TypeScript check

# Development mode
npm run dev                # Frontend watch mode
```

### Testing

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

### Multi-version testing

The following scripts test the plugin against all supported Grafana versions (currently 10.4.19, 11.1.13, 11.4.8, 12.1.5, 12.3.1). They build the plugin once, then iterate over each version.

- `scripts/e2e-all-versions.sh` — builds the plugin, starts each Grafana version sequentially via `docker compose`, runs `npm run e2e` against it, and writes a timestamped results file.
- `scripts/start-all-versions.sh` — builds the plugin, then starts all versions simultaneously as individual Docker containers on consecutive ports starting at 3000 (e.g. 3000, 3001, 3002, ...) for side-by-side manual comparison.
- `scripts/stop-all-versions.sh` — stops and removes all containers started by `start-all-versions.sh`.

### Architecture

**Frontend (src/):**

- `module.ts` - Plugin entry point
- `datasource.ts` - OpcuaDataSource (extends DataSourceWithBackend)
- `components/ConfigEditor/` - Data source configuration UI
- `components/QueryEditor/` - Query editor with node browser

**Backend (pkg/):**

- `plugin/datasource.go` - Query and health check handlers
- `plugin/resources.go` - HTTP handlers for /browse and /endpoints
- `plugin/opcua/` - OPC-UA client, auth, browsing, certificates

**Key Dependencies:**

- Go: `github.com/gopcua/opcua` - OPC-UA protocol
- Go: `github.com/grafana/grafana-plugin-sdk-go` - Plugin SDK

## License

Apache License 2.0. See [LICENSE](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/LICENSE) for details.

## Support

- [GitHub Issues](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues)
- [Documentation](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/README.md)

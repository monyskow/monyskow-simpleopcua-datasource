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

| Setting | Description | Example |
|---------|-------------|---------|
| Endpoint URL | OPC-UA server endpoint | `opc.tcp://localhost:4840` |
| Security Policy | Encryption algorithm | None, Basic256Sha256 |
| Security Mode | Message security | None, Sign, SignAndEncrypt |
| Timeout | Connection timeout (seconds) | 10 |

### Authentication

#### Anonymous

No credentials required. Select "Anonymous" as the authentication method.

#### Username/Password

1. Select "Username / Password" as the authentication method
2. Enter your username and password
3. Credentials are stored securely using Grafana's secure JSON data

#### Certificate

1. Select "Certificate" as the authentication method
2. Paste your PEM-encoded certificate and private key
3. Both are stored securely using Grafana's secure JSON data

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
- Docker (for local testing)

### Building

```bash
# Install dependencies
npm install

# Build frontend
npm run build

# Build backend (all platforms)
mage buildAll

# Start development environment
docker compose up
```

### Testing

```bash
# Unit tests
npm run test:ci
mage test

# E2E tests
npm run e2e
```

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

## Support

- [GitHub Issues](https://github.com/monyskow/monyskow-simpleopcua-datasource/issues)
- [Documentation](https://github.com/monyskow/monyskow-simpleopcua-datasource/blob/main/README.md)

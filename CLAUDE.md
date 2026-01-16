# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Simple OPC-UA is a Grafana data source plugin for connecting to OPC-UA servers. It reads industrial data directly into Grafana dashboards.

**Plugin ID**: `monyskow-simpleopcua-datasource`
**Plugin Type**: Data source plugin with Go backend
**Grafana Compatibility**: >=10.4.0

## Common Commands

### Development

```bash
npm install              # Install frontend dependencies
go mod tidy              # Install Go dependencies
npm run dev              # Start webpack in watch mode (frontend)
mage -v watch            # Watch mode for Go backend (auto-rebuild)
npm run server           # Start local Grafana + OPC-UA server via Docker
```

### Testing

```bash
npm test                 # Run unit tests in watch mode
npm run test:ci          # Run all unit tests once (CI mode)
npm run test:coverage    # Run tests with coverage report
mage test                # Run Go backend tests
mage coverage            # Go tests with coverage
npm run e2e              # Run E2E tests (requires npm run server)
npm run typecheck        # TypeScript type checking
npm run lint             # Run ESLint
npm run lint:fix         # Auto-fix linting issues
```

### Building

```bash
npm run build            # Production frontend build
mage build:darwinARM64   # Build backend for macOS ARM
mage build:linuxARM64    # Build backend for Linux ARM
mage buildAll            # Build for all platforms
npm run sign             # Sign plugin for distribution
```

## Architecture

### Backend (Go)

**pkg/main.go** - Plugin entry point, starts the gRPC server using grafana-plugin-sdk-go.

**pkg/plugin/datasource.go** - Core implementation:

- Implements `QueryDataHandler` for data queries
- Implements `CheckHealthHandler` for Save & Test
- Implements `CallResourceHandler` for browse API
- Manages OPC-UA client lifecycle

**pkg/plugin/models/settings.go** - Data source configuration:

- Parses jsonData and secureJsonData from Grafana
- Supports Anonymous, Username/Password, Certificate auth
- Security policy and mode configuration

**pkg/plugin/models/query.go** - Query model:

- Represents nodes to read from OPC-UA server
- Each node has nodeId, displayName, and optional alias

**pkg/plugin/opcua/client.go** - OPC-UA client wrapper:

- Uses gopcua library for OPC-UA protocol
- Connection management with auto-reconnect
- ReadNodes for batch reading values
- ReadServerState for health checks

**pkg/plugin/opcua/auth.go** - Authentication:

- Anonymous auth
- Username/Password auth
- Certificate auth with PEM parsing

**pkg/plugin/opcua/browse.go** - Node browsing:

- Traverses OPC-UA address space
- Returns nodes with metadata (class, hasChildren)
- GetEndpoints for server discovery

**pkg/plugin/resources.go** - HTTP resource handlers:

- `/browse` - Browse OPC-UA nodes
- `/endpoints` - Get server endpoints

### Frontend (TypeScript/React)

**src/module.ts** - Plugin entry point, registers DataSource, ConfigEditor, QueryEditor.

**src/types.ts** - TypeScript interfaces:

- `OpcuaDataSourceOptions` - Connection and auth settings
- `OpcuaSecureJsonData` - Encrypted credentials
- `OpcuaQuery` - Query model with nodes array
- `OpcuaBrowseNode` - Node from browsing

**src/datasource.ts** - DataSource class:

- Extends `DataSourceWithBackend` for backend queries
- `applyTemplateVariables` - Template variable substitution
- `browseNodes` - Calls backend browse API
- `getEndpoints` - Calls backend endpoints API

**src/components/ConfigEditor/ConfigEditor.tsx** - Configuration UI:

- Endpoint URL input
- Security policy/mode selectors
- Authentication method selection
- Credential inputs (username/password or certificate)

**src/components/QueryEditor/QueryEditor.tsx** - Query builder:

- Node list management (add/remove/update)
- Integration with NodeBrowser
- Alias configuration

**src/components/QueryEditor/NodeBrowser.tsx** - Tree browser:

- Loads OPC-UA address space from backend
- Expandable tree with lazy loading
- Click to add Variable nodes to query

### File Structure

```
monyskow-simpleopcua-datasource/
├── pkg/                          # Go backend
│   ├── main.go                   # Entry point
│   └── plugin/
│       ├── datasource.go         # Core handlers
│       ├── resources.go          # HTTP API
│       ├── models/
│       │   ├── settings.go       # Config parsing
│       │   └── query.go          # Query model
│       └── opcua/
│           ├── client.go         # OPC-UA client
│           ├── auth.go           # Auth options
│           └── browse.go         # Node browsing
├── src/                          # Frontend
│   ├── module.ts                 # Plugin registration
│   ├── datasource.ts             # DataSource class
│   ├── types.ts                  # TypeScript types
│   └── components/
│       ├── ConfigEditor/
│       │   └── ConfigEditor.tsx  # Config UI
│       └── QueryEditor/
│           ├── QueryEditor.tsx   # Query builder
│           └── NodeBrowser.tsx   # Node tree
├── provisioning/                 # Grafana provisioning
├── tests/                        # E2E tests
├── go.mod                        # Go dependencies
├── Magefile.go                   # Go build config
└── package.json                  # Node dependencies
```

## Key Implementation Details

### OPC-UA Client Management

- Client is created on first query or health check
- Auto-reconnect enabled with 5-second retry interval
- Client disposed when Grafana removes the data source instance
- Uses mutex for thread-safe connection management

### Authentication Flow

1. Settings parsed from Grafana's jsonData and secureJsonData
2. Auth options generated based on authMethod
3. Certificate auth: PEM parsing, PKCS1/PKCS8 private key support
4. Applied to gopcua client options

### Browse API

- Default starts from Objects folder (ns=0;i=85)
- Returns child nodes with node class and hasChildren flag
- Supports continuation points for large node sets
- Frontend lazy-loads children on expand

### Query Processing

1. Parse query JSON from frontend
2. Parse OPC-UA Node IDs
3. Build ReadRequest with AttributeIDValue
4. Execute Read via gopcua
5. Convert values to Grafana data frames

### Data Frame Format

- Single row per query (snapshot of current values)
- Time field with current timestamp
- One field per node, named by alias or displayName

## Testing Philosophy

**Unit Tests** (Jest + Go):

- Test TypeScript types and DataSource methods
- Test Go settings parsing and auth option generation
- Mock OPC-UA client for isolation

**E2E Tests** (Playwright):

- Test complete workflows with real Grafana
- Use provisioned OPC-UA test server
- Test config editor, query editor, node browser

## Development Workflow

1. Start OPC-UA test server: `docker compose -f docker-compose.opcua.yaml up -d`
2. Start Grafana: `npm run server`
3. Build frontend: `npm run dev`
4. Build backend: `mage -v watch`
5. Access Grafana at http://localhost:3000 (admin/admin)
6. Add "OPC-UA Test Server" data source (pre-provisioned)

## Critical Gotchas

### gopcua API

- Use `result.Status` not `result.StatusCode` for DataValue status
- Node IDs must be properly formatted: `ns=<namespace>;s=<string>` or `ns=<namespace>;i=<integer>`
- Security options must match server requirements

### Grafana Plugin SDK

- Health check must return CheckHealthResult, not error
- Resource handlers use httpadapter for HTTP-like API
- secureJsonData is decrypted in settings.DecryptedSecureJSONData

### Frontend

- Use `DataSourceWithBackend` for backend-powered data sources
- `getResource` calls backend CallResource handler
- Template variables replaced via `getTemplateSrv().replace()`

## CI/CD

GitHub Actions should:

- Build and lint on every PR
- Run unit tests (frontend + backend)
- Run E2E tests against Grafana version matrix
- Sign and release on version tag

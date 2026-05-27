# Configuration

How to configure the Simple OPC-UA datasource in Grafana.

## Connection Settings

| Setting         | Description                  | Example                    |
| --------------- | ---------------------------- | -------------------------- |
| Endpoint URL    | OPC-UA server endpoint       | `opc.tcp://localhost:4840` |
| Security Policy | Encryption algorithm         | None, Basic256Sha256       |
| Security Mode   | Message security             | None, Sign, SignAndEncrypt |
| Timeout         | Connection timeout (seconds) | 10                         |

## Security

When using security modes other than "None" (Sign or SignAndEncrypt), the plugin needs a client certificate. Two options are available:

### Option 1: Auto-Generated Certificate (Recommended)

The plugin can automatically generate and manage client certificates:

1. Click **"Generate Certificate"** button in the datasource configuration
2. The plugin generates a self-signed client certificate
3. Click **"Save & Test"** to persist the certificate
4. The certificate is securely stored in Grafana's encrypted storage

The auto-generated certificate is valid for 3 years and persists across Grafana restarts.

### Option 2: Provide Your Own Certificate

For advanced use cases, you can provide your own client certificate using the Certificate authentication method (see Authentication section below).

**Important:** For secure connections, you must either generate a certificate or provide your own. The OPC-UA server must trust the client certificate — consult your OPC-UA server documentation for adding trusted certificates.

### Supported Security Policies

- None (no encryption)
- Basic256Sha256
- Aes128_Sha256_RsaOaep
- Aes256_Sha256_RsaPss

### Supported Security Modes

- None
- Sign (messages are signed but not encrypted)
- SignAndEncrypt (messages are signed and encrypted)

## Authentication

### Anonymous

No credentials required. Select "Anonymous" as the authentication method.

### Username/Password

1. Select "Username / Password" as the authentication method
2. Enter your username and password
3. Credentials are stored securely using Grafana's secure JSON data

### Certificate

For advanced use cases where you need to use a specific client certificate:

1. Select "Certificate" as the authentication method
2. Paste your PEM-encoded certificate and private key
3. Both are stored securely using Grafana's secure JSON data

Note: When using Certificate authentication, the provided certificate is used instead of the auto-generated one.

const {
  OPCUAServer,
  Variant,
  DataType,
  SecurityPolicy,
  MessageSecurityMode,
  UserTokenType,
  OPCUACertificateManager,
} = require('node-opcua');

const path = require('path');
const fs = require('fs');

// Configuration from environment variables
const PORT = parseInt(process.env.PORT || '4840', 10);
const HOSTNAME = process.env.HOSTNAME || '0.0.0.0';
const SERVER_NAME = process.env.SERVER_NAME || 'OPCUATestServer';

// User credentials for testing
const USERS = {
  user1: 'password1',
  admin: 'admin123',
};

// PKI directory for certificates
const PKI_DIR = path.join(__dirname, 'pki');

// Ensure PKI directory exists
if (!fs.existsSync(PKI_DIR)) {
  fs.mkdirSync(PKI_DIR, { recursive: true });
}

async function createServer() {
  console.log(`Starting OPC UA server on port ${PORT}...`);

  // Create certificate manager with auto-accept
  const serverCertificateManager = new OPCUACertificateManager({
    automaticallyAcceptUnknownCertificate: true,
    rootFolder: PKI_DIR,
    name: 'PKI',
  });
  await serverCertificateManager.initialize();

  // Separate certificate manager for X509 user-identity tokens; auto-accepts any
  // client cert so cert-based authentication round-trips work in tests without
  // pre-provisioned trust. Required so the server advertises X509IdentityToken
  // in its endpoint descriptions (cert auth combinations from #18 AC).
  const userCertificateManager = new OPCUACertificateManager({
    automaticallyAcceptUnknownCertificate: true,
    rootFolder: path.join(PKI_DIR, 'user'),
    name: 'UserPKI',
  });
  await userCertificateManager.initialize();

  const server = new OPCUAServer({
    port: PORT,
    hostname: HOSTNAME,

    // Server info
    serverInfo: {
      applicationUri: `urn:${SERVER_NAME}`,
      productUri: 'urn:GrafanaOPCUATestServer',
      applicationName: { text: 'Grafana OPC UA Test Server' },
    },

    // Build info
    buildInfo: {
      productName: 'Grafana OPC UA Test Server',
      buildNumber: '1.0.0',
      buildDate: new Date(),
    },

    // Enable all security policies including Aes256_Sha256_RsaPss
    securityPolicies: [
      SecurityPolicy.None,
      SecurityPolicy.Basic256Sha256,
      SecurityPolicy.Aes128_Sha256_RsaOaep,
      SecurityPolicy.Aes256_Sha256_RsaPss,
    ],

    // Security modes
    securityModes: [MessageSecurityMode.None, MessageSecurityMode.Sign, MessageSecurityMode.SignAndEncrypt],

    // User management
    userManager: {
      isValidUser: (userName, password) => {
        console.log(`Authentication attempt: user=${userName}`);
        if (USERS[userName] && USERS[userName] === password) {
          console.log(`Authentication successful for user: ${userName}`);
          return true;
        }
        console.log(`Authentication failed for user: ${userName}`);
        return false;
      },
    },

    // Allow anonymous access
    allowAnonymous: true,

    // Certificate manager
    serverCertificateManager,
    userCertificateManager,

    // Resource path for self-signed certificate generation
    resourcePath: '/UA/Server',
  });

  // Initialize the server (generates certificates if needed)
  await server.initialize();

  console.log('Server initialized. Creating address space...');

  // Get address space
  const addressSpace = server.engine.addressSpace;
  const namespace = addressSpace.getOwnNamespace();

  // Create a folder for our test variables
  const testFolder = namespace.addFolder('ObjectsFolder', {
    browseName: 'TestVariables',
  });

  // Add test variables
  let counter = 0;
  let temperature = 20.0;
  let status = true;

  // Counter variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'Counter',
    dataType: 'Int32',
    value: {
      get: () => new Variant({ dataType: DataType.Int32, value: counter }),
    },
  });

  // Temperature variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'Temperature',
    dataType: 'Double',
    value: {
      get: () => new Variant({ dataType: DataType.Double, value: temperature }),
    },
  });

  // Status variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'Status',
    dataType: 'Boolean',
    value: {
      get: () => new Variant({ dataType: DataType.Boolean, value: status }),
    },
  });

  // Random number variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'RandomNumber',
    dataType: 'Double',
    value: {
      get: () => new Variant({ dataType: DataType.Double, value: Math.random() * 100 }),
    },
  });

  // Timestamp variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'CurrentTime',
    dataType: 'DateTime',
    value: {
      get: () => new Variant({ dataType: DataType.DateTime, value: new Date() }),
    },
  });

  // String variable
  namespace.addVariable({
    componentOf: testFolder,
    browseName: 'ServerMessage',
    dataType: 'String',
    value: {
      get: () => new Variant({ dataType: DataType.String, value: `Server running since ${server.startTime}` }),
    },
  });

  // Update counter and temperature periodically
  setInterval(() => {
    counter++;
    temperature = 20 + Math.sin(counter / 10) * 5 + (Math.random() - 0.5);
    status = counter % 10 !== 0;
  }, 1000);

  // Start the server
  await server.start();

  console.log('='.repeat(60));
  console.log('OPC UA Test Server started successfully!');
  console.log('='.repeat(60));
  console.log(`Endpoint URL: ${server.getEndpointUrl()}`);
  console.log(`Port: ${PORT}`);
  console.log('Security Policies:');
  console.log('  - None');
  console.log('  - Basic256Sha256');
  console.log('  - Aes128_Sha256_RsaOaep');
  console.log('  - Aes256_Sha256_RsaPss');
  console.log('Security Modes:');
  console.log('  - None');
  console.log('  - Sign');
  console.log('  - SignAndEncrypt');
  console.log('Authentication:');
  console.log('  - Anonymous: allowed');
  console.log('  - Username/Password: user1/password1, admin/admin123');
  console.log('  - Certificate: any client cert auto-accepted (test/dev only)');
  console.log('='.repeat(60));
  console.log('Available endpoints:');

  const endpoints = server.endpoints[0].endpointDescriptions();
  endpoints.forEach((ep, i) => {
    const policyName = ep.securityPolicyUri.split('#')[1] || 'None';
    const modeName = MessageSecurityMode[ep.securityMode];
    const tokens = ep.userIdentityTokens.map((t) => UserTokenType[t.tokenType]).join(', ');
    console.log(`  ${i + 1}. ${policyName}/${modeName} - Tokens: [${tokens}]`);
  });
  console.log('='.repeat(60));

  // Graceful shutdown
  process.on('SIGINT', async () => {
    console.log('\nShutting down server...');
    await server.shutdown();
    console.log('Server shut down.');
    process.exit(0);
  });

  process.on('SIGTERM', async () => {
    console.log('\nShutting down server...');
    await server.shutdown();
    console.log('Server shut down.');
    process.exit(0);
  });
}

createServer().catch((err) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});

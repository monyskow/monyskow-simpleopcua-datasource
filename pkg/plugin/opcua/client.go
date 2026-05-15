package opcua

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// Client wraps the gopcua client with additional functionality
type Client struct {
	client    *opcua.Client
	settings  models.DataSourceSettings
	logger    log.Logger
	mu        sync.RWMutex
	connected bool
	certMgr   *CertificateManager // Reference to datasource's certificate manager
}

// NodeValue represents the result of reading a node
type NodeValue struct {
	NodeID          string
	DisplayName     string
	Value           interface{}
	StatusCode      ua.StatusCode
	SourceTimestamp time.Time
	ServerTimestamp time.Time
}

// NewClient creates a new OPC-UA client
func NewClient(settings models.DataSourceSettings, logger log.Logger, certMgr *CertificateManager) (*Client, error) {
	opts := []opcua.Option{
		opcua.RequestTimeout(time.Duration(settings.Timeout) * time.Second),
		opcua.AutoReconnect(true),
		opcua.ReconnectInterval(5 * time.Second),
	}

	// Check if endpoint discovery is needed:
	// - For any non-anonymous authentication (userpass, certificate)
	// - For any security policy/mode other than None
	needsEndpointDiscovery := settings.AuthMethod != models.AuthAnonymous ||
		(settings.SecurityPolicy != "" && settings.SecurityPolicy != "None") ||
		(settings.SecurityMode != "" && settings.SecurityMode != "None")

	// For connections requiring endpoint discovery, get server endpoints first
	if needsEndpointDiscovery {
		logger.Info("Fetching server endpoints",
			"authMethod", settings.AuthMethod,
			"securityPolicy", settings.SecurityPolicy,
			"securityMode", settings.SecurityMode)

		// Fetch endpoints to get server certificate
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Fetch endpoints without security - GetEndpoints should work without credentials
		endpoints, err := opcua.GetEndpoints(ctx, settings.Endpoint)
		if err != nil {
			logger.Warn("Failed to get endpoints, will try connection without endpoint discovery", "error", err)
		} else {
			// Log available endpoints for debugging
			logAvailableEndpoints(logger, endpoints, settings)

			// Validate that the requested security policy is supported
			if err := validatePolicySupport(endpoints, settings, logger); err != nil {
				return nil, err
			}

			// Find matching endpoint
			for _, ep := range endpoints {
				if matchesSecuritySettings(ep, settings) {
					logger.Info("Found matching endpoint, using SecurityFromEndpoint",
						"securityPolicy", ep.SecurityPolicyURI,
						"securityMode", ep.SecurityMode,
						"userTokenTypes", getUserTokenTypes(ep))

					// Determine the correct UserTokenType based on auth method
					userTokenType := ua.UserTokenTypeAnonymous
					switch settings.AuthMethod {
					case models.AuthUserPass:
						userTokenType = ua.UserTokenTypeUserName
					case models.AuthCertificate:
						userTokenType = ua.UserTokenTypeCertificate
					}

					// Check if this is a secure channel (not None/None)
					isSecureChannel := ep.SecurityMode != ua.MessageSecurityModeNone

					// Generate or use provided client certificate (only for secure channels)
					// This must be done BEFORE SecurityFromEndpoint
					if isSecureChannel {
						certOpts, err := getClientCertificateOptions(settings, certMgr, logger)
						if err != nil {
							return nil, fmt.Errorf("client certificate: %w", err)
						}
						opts = append(opts, certOpts...)
					}

					// Use SecurityFromEndpoint to automatically configure security
					// This sets: SecurityPolicy, SecurityMode, RemoteCertificate, and AuthPolicyID
					opts = append(opts, opcua.SecurityFromEndpoint(ep, userTokenType))

					// Add authentication options (username/password, certificate, etc.)
					authOpts, err := GetAuthOptions(settings)
					if err != nil {
						return nil, fmt.Errorf("auth options: %w", err)
					}
					opts = append(opts, authOpts...)

					// Log authentication details (without revealing password)
					if settings.AuthMethod == models.AuthUserPass {
						logger.Info("Using username/password authentication",
							"username", settings.Username,
							"hasPassword", len(settings.Password) > 0,
							"passwordLength", len(settings.Password))
					}

					c, err := opcua.NewClient(settings.Endpoint, opts...)
					if err != nil {
						return nil, fmt.Errorf("create client: %w", err)
					}

					return &Client{
						client:   c,
						settings: settings,
						logger:   logger,
						certMgr:  certMgr,
					}, nil
				}
			}
			logger.Warn("No matching endpoint found, falling back to manual security configuration")
		}
	}

	// Fallback: Add security options manually
	secOpts, err := getSecurityOptions(settings, certMgr, logger)
	if err != nil {
		return nil, fmt.Errorf("security options: %w", err)
	}
	opts = append(opts, secOpts...)

	// Add authentication options
	authOpts, err := GetAuthOptions(settings)
	if err != nil {
		return nil, fmt.Errorf("auth options: %w", err)
	}
	opts = append(opts, authOpts...)

	c, err := opcua.NewClient(settings.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Client{
		client:   c,
		settings: settings,
		logger:   logger,
		certMgr:  certMgr,
	}, nil
}

// matchesSecuritySettings checks if an endpoint matches the configured security settings
func matchesSecuritySettings(ep *ua.EndpointDescription, settings models.DataSourceSettings) bool {
	// For None/None security (unsecure channel), only check that endpoint has MessageSecurityModeNone
	if (settings.SecurityPolicy == "" || settings.SecurityPolicy == "None") &&
		(settings.SecurityMode == "" || settings.SecurityMode == "None") {
		if ep.SecurityMode != ua.MessageSecurityModeNone {
			return false
		}
		// For unsecure channel, don't check security policy URI - just need mode None
	} else {
		// Check security policy for secure channels
		if settings.SecurityPolicy != "" && settings.SecurityPolicy != "None" {
			expectedPolicy := "http://opcfoundation.org/UA/SecurityPolicy#" + settings.SecurityPolicy
			if ep.SecurityPolicyURI != expectedPolicy {
				return false
			}
		}

		// Check security mode for secure channels
		if settings.SecurityMode != "" && settings.SecurityMode != "None" {
			var expectedMode ua.MessageSecurityMode
			switch settings.SecurityMode {
			case "Sign":
				expectedMode = ua.MessageSecurityModeSign
			case "SignAndEncrypt":
				expectedMode = ua.MessageSecurityModeSignAndEncrypt
			default:
				expectedMode = ua.MessageSecurityModeNone
			}
			if ep.SecurityMode != expectedMode {
				return false
			}
		}
	}

	// Check if endpoint supports the required user token type
	requiredTokenType := getRequiredTokenType(settings.AuthMethod)
	if !endpointSupportsTokenType(ep, requiredTokenType) {
		return false
	}

	return true
}

// getRequiredTokenType returns the UserTokenType for the given auth method
func getRequiredTokenType(authMethod models.AuthMethod) ua.UserTokenType {
	switch authMethod {
	case models.AuthUserPass:
		return ua.UserTokenTypeUserName
	case models.AuthCertificate:
		return ua.UserTokenTypeCertificate
	default:
		return ua.UserTokenTypeAnonymous
	}
}

// validatePolicySupport checks if the requested security policy is supported by the server
func validatePolicySupport(endpoints []*ua.EndpointDescription, settings models.DataSourceSettings, _ log.Logger) error {
	// No validation needed for None security
	if settings.SecurityPolicy == "" || settings.SecurityPolicy == "None" {
		return nil
	}

	expectedPolicy := "http://opcfoundation.org/UA/SecurityPolicy#" + settings.SecurityPolicy

	// Check if any endpoint supports the requested policy
	for _, ep := range endpoints {
		if ep.SecurityPolicyURI == expectedPolicy {
			return nil // Policy is supported
		}
	}

	// Collect available policies for error message
	var available []string
	seen := make(map[string]bool)
	for _, ep := range endpoints {
		policyName := extractPolicyName(ep.SecurityPolicyURI)
		if !seen[policyName] {
			seen[policyName] = true
			available = append(available, policyName)
		}
	}

	return fmt.Errorf("security policy '%s' is not supported by this server. Available policies: %v", settings.SecurityPolicy, available)
}

// extractPolicyName extracts the policy name from a full security policy URI
func extractPolicyName(uri string) string {
	// Extract just the policy name from URIs like "http://opcfoundation.org/UA/SecurityPolicy#Basic256Sha256"
	const prefix = "http://opcfoundation.org/UA/SecurityPolicy#"
	if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
		return uri[len(prefix):]
	}
	return uri
}

// endpointSupportsTokenType checks if an endpoint supports the given user token type
func endpointSupportsTokenType(ep *ua.EndpointDescription, tokenType ua.UserTokenType) bool {
	for _, policy := range ep.UserIdentityTokens {
		if policy.TokenType == tokenType {
			return true
		}
	}
	return false
}

// getUserTokenTypes returns a string describing supported user token types for an endpoint
func getUserTokenTypes(ep *ua.EndpointDescription) string {
	var types []string
	for _, policy := range ep.UserIdentityTokens {
		types = append(types, fmt.Sprintf("%s(policy=%s)", tokenTypeName(policy.TokenType), policy.PolicyID))
	}
	if len(types) == 0 {
		return "none"
	}
	return fmt.Sprintf("[%s]", joinStrings(types, ", "))
}

// tokenTypeName returns a human-readable name for a UserTokenType
func tokenTypeName(t ua.UserTokenType) string {
	switch t {
	case ua.UserTokenTypeAnonymous:
		return "Anonymous"
	case ua.UserTokenTypeUserName:
		return "UserName"
	case ua.UserTokenTypeCertificate:
		return "Certificate"
	case ua.UserTokenTypeIssuedToken:
		return "IssuedToken"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// joinStrings joins strings with a separator (simple helper to avoid importing strings package)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// logAvailableEndpoints logs information about available server endpoints for debugging
func logAvailableEndpoints(logger log.Logger, endpoints []*ua.EndpointDescription, settings models.DataSourceSettings) {
	requiredTokenType := getRequiredTokenType(settings.AuthMethod)
	logger.Info("Server endpoints discovery",
		"totalEndpoints", len(endpoints),
		"requiredSecurityPolicy", settings.SecurityPolicy,
		"requiredSecurityMode", settings.SecurityMode,
		"requiredAuthMethod", settings.AuthMethod,
		"requiredTokenType", tokenTypeName(requiredTokenType),
	)

	for i, ep := range endpoints {
		policyMatch := true
		modeMatch := true
		tokenMatch := endpointSupportsTokenType(ep, requiredTokenType)

		// For None/None security, match any endpoint with MessageSecurityModeNone
		if (settings.SecurityPolicy == "" || settings.SecurityPolicy == "None") &&
			(settings.SecurityMode == "" || settings.SecurityMode == "None") {
			policyMatch = true // Don't check policy for None security
			modeMatch = ep.SecurityMode == ua.MessageSecurityModeNone
		} else {
			if settings.SecurityPolicy != "" && settings.SecurityPolicy != "None" {
				expectedPolicy := "http://opcfoundation.org/UA/SecurityPolicy#" + settings.SecurityPolicy
				policyMatch = ep.SecurityPolicyURI == expectedPolicy
			}

			if settings.SecurityMode != "" && settings.SecurityMode != "None" {
				var expectedMode ua.MessageSecurityMode
				switch settings.SecurityMode {
				case "Sign":
					expectedMode = ua.MessageSecurityModeSign
				case "SignAndEncrypt":
					expectedMode = ua.MessageSecurityModeSignAndEncrypt
				}
				modeMatch = ep.SecurityMode == expectedMode
			}
		}

		logger.Info("Endpoint",
			"index", i,
			"securityPolicy", ep.SecurityPolicyURI,
			"securityMode", ep.SecurityMode,
			"userTokens", getUserTokenTypes(ep),
			"policyMatch", policyMatch,
			"modeMatch", modeMatch,
			"tokenMatch", tokenMatch,
			"overallMatch", policyMatch && modeMatch && tokenMatch,
		)

		// Log detailed token policy information for UserName tokens (diagnostic)
		for _, policy := range ep.UserIdentityTokens {
			if policy.TokenType == ua.UserTokenTypeUserName {
				logger.Info("UserName token policy detail",
					"endpointIndex", i,
					"policyId", policy.PolicyID,
					"securityPolicyURI", policy.SecurityPolicyURI,
					"issuedTokenType", policy.IssuedTokenType)
			}
		}
	}
}

// Connect establishes connection to the OPC-UA server
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	if err := c.client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	c.connected = true
	c.logger.Info("Connected to OPC-UA server", "endpoint", c.settings.Endpoint)
	return nil
}

// Close closes the connection
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	err := c.client.Close(ctx)
	c.connected = false
	c.logger.Info("Disconnected from OPC-UA server")
	if err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// ReadNodes reads values from multiple nodes
func (c *Client) ReadNodes(ctx context.Context, nodes []models.NodeQuery) ([]NodeValue, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	if len(nodes) == 0 {
		return []NodeValue{}, nil
	}

	// Parse node IDs
	nodeIDs := make([]*ua.NodeID, len(nodes))
	for i, n := range nodes {
		id, err := ua.ParseNodeID(n.NodeID)
		if err != nil {
			return nil, fmt.Errorf("parse node %s: %w", n.NodeID, err)
		}
		nodeIDs[i] = id
	}

	// Build read request
	req := &ua.ReadRequest{
		MaxAge:             0,
		NodesToRead:        make([]*ua.ReadValueID, len(nodeIDs)),
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	for i, id := range nodeIDs {
		req.NodesToRead[i] = &ua.ReadValueID{
			NodeID:      id,
			AttributeID: ua.AttributeIDValue,
		}
	}

	// Execute read
	resp, err := client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	// Process results
	values := make([]NodeValue, len(resp.Results))
	for i, result := range resp.Results {
		var value interface{}
		if result.Value != nil {
			value = result.Value.Value()
		}

		values[i] = NodeValue{
			NodeID:          nodes[i].NodeID,
			DisplayName:     nodes[i].DisplayName,
			Value:           value,
			StatusCode:      result.Status,
			SourceTimestamp: result.SourceTimestamp,
			ServerTimestamp: result.ServerTimestamp,
		}
	}

	return values, nil
}

// ReadServerState reads the server state node to verify connectivity
func (c *Client) ReadServerState(ctx context.Context) (ua.ServerState, error) {
	nodes := []models.NodeQuery{{
		NodeID:      "ns=0;i=2259", // Server_ServerStatus_State
		DisplayName: "ServerState",
	}}

	values, err := c.ReadNodes(ctx, nodes)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("no server state returned")
	}

	if values[0].StatusCode != ua.StatusOK {
		return 0, fmt.Errorf("bad status: %v", values[0].StatusCode)
	}

	state, ok := values[0].Value.(int32)
	if !ok {
		return 0, fmt.Errorf("unexpected server state type: %T", values[0].Value)
	}

	return ua.ServerState(state), nil
}

// loadClientCert decodes and parses a PEM-encoded certificate and RSA private key,
// validates that they form a matching pair, and returns the parsed values.
func loadClientCert(certPEM, keyPEM []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "RSA PRIVATE KEY" {
		return nil, nil, fmt.Errorf("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if err := validateCertKeyPair(cert, key); err != nil {
		return nil, nil, fmt.Errorf("certificate and key validation failed: %w", err)
	}

	return cert, key, nil
}

// getClientCertificateOptions returns client certificate and key options
func getClientCertificateOptions(settings models.DataSourceSettings, certMgr *CertificateManager, logger log.Logger) ([]opcua.Option, error) {
	var opts []opcua.Option

	// If user provided a certificate for certificate auth, that's handled by GetAuthOptions
	if len(settings.Certificate) > 0 {
		return opts, nil
	}

	var certPEM, keyPEM []byte
	var err error

	// Priority 1: Use stored client certificate from secureJsonData (persists across restarts)
	if len(settings.ClientCert) > 0 && len(settings.ClientKey) > 0 {
		logger.Info("Using stored client certificate from secureJsonData")
		certPEM = settings.ClientCert
		keyPEM = settings.ClientKey
	} else {
		// Priority 2: Generate new certificate (only for initial setup or if not stored)
		logger.Info("Security enabled, generating client certificate (store it via 'Generate Certificate' button for persistence)")
		certPEM, keyPEM, err = certMgr.GetOrCreate()
		if err != nil {
			return nil, fmt.Errorf("get or create certificate: %w", err)
		}
	}

	cert, key, err := loadClientCert(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	appURI := extractApplicationURI(cert)
	logger.Info("Extracted Application URI from certificate", "applicationURI", appURI)

	opts = append(opts,
		opcua.Certificate(cert.Raw),
		opcua.PrivateKey(key),
		opcua.ApplicationURI(appURI),
	)

	logger.Info("Using auto-generated certificate",
		"subject", cert.Subject.CommonName,
		"notAfter", cert.NotAfter,
		"applicationURI", appURI,
	)

	return opts, nil
}

// getSecurityOptions returns security-related client options
func getSecurityOptions(settings models.DataSourceSettings, certMgr *CertificateManager, logger log.Logger) ([]opcua.Option, error) {
	var opts []opcua.Option

	// Check if security is enabled
	securityEnabled := (settings.SecurityPolicy != "" && settings.SecurityPolicy != "None") ||
		(settings.SecurityMode != "" && settings.SecurityMode != "None")

	if settings.SecurityPolicy != "" && settings.SecurityPolicy != "None" {
		opts = append(opts, opcua.SecurityPolicy(settings.SecurityPolicy))
	}

	if settings.SecurityMode != "" && settings.SecurityMode != "None" {
		opts = append(opts, opcua.SecurityModeString(settings.SecurityMode))
	}

	// If security is enabled and user didn't provide a certificate, use stored or auto-generated one
	if securityEnabled && len(settings.Certificate) == 0 {
		var certPEM, keyPEM []byte
		var err error

		// Priority 1: Use stored client certificate from secureJsonData (persists across restarts)
		if len(settings.ClientCert) > 0 && len(settings.ClientKey) > 0 {
			logger.Info("Using stored client certificate from secureJsonData")
			certPEM = settings.ClientCert
			keyPEM = settings.ClientKey
		} else {
			// Priority 2: Generate new certificate (only for initial setup or if not stored)
			logger.Info("Security enabled, generating client certificate (store it via 'Generate Certificate' button for persistence)")
			certPEM, keyPEM, err = certMgr.GetOrCreate()
			if err != nil {
				return nil, fmt.Errorf("get or create certificate: %w", err)
			}
		}

		cert, key, err := loadClientCert(certPEM, keyPEM)
		if err != nil {
			return nil, err
		}

		appURI := extractApplicationURI(cert)

		opts = append(opts,
			opcua.Certificate(cert.Raw),
			opcua.PrivateKey(key),
			opcua.ApplicationURI(appURI),
		)

		logger.Info("Using auto-generated certificate",
			"subject", cert.Subject.CommonName,
			"notAfter", cert.NotAfter,
			"applicationURI", appURI,
		)
	}

	return opts, nil
}

// validateCertKeyPair validates that a certificate and private key are a matching pair
func validateCertKeyPair(cert *x509.Certificate, key *rsa.PrivateKey) error {
	// Compare public key from certificate with public key derived from private key
	certPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("certificate public key is not RSA")
	}

	if certPubKey.N.Cmp(key.N) != 0 || certPubKey.E != key.E {
		return fmt.Errorf("certificate and private key do not match")
	}

	return nil
}

// extractApplicationURI extracts the application URI from the certificate's Subject Alternative Names
func extractApplicationURI(cert *x509.Certificate) string {
	// Extract URI from certificate's SAN (Subject Alternative Names)
	for _, uri := range cert.URIs {
		return uri.String()
	}
	// Fallback to a default URI if none found
	return "urn:grafana:simpleopcua:client"
}

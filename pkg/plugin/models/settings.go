package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// AuthMethod represents the authentication method to use
type AuthMethod string

const (
	AuthAnonymous   AuthMethod = "anonymous"
	AuthUserPass    AuthMethod = "userpass"
	AuthCertificate AuthMethod = "certificate"
)

// DataSourceSettings contains all configuration for the OPC-UA data source
type DataSourceSettings struct {
	// Connection settings (from jsonData)
	Endpoint       string     `json:"endpoint"`       // opc.tcp://host:port/path
	SecurityPolicy string     `json:"securityPolicy"` // None, Basic128Rsa15, Basic256, Basic256Sha256, etc.
	SecurityMode   string     `json:"securityMode"`   // None, Sign, SignAndEncrypt
	AuthMethod     AuthMethod `json:"authMethod"`     // anonymous, userpass, certificate
	Timeout        int        `json:"timeout"`        // Connection timeout in seconds
	// True when a client certificate has been generated and saved. Stored in jsonData
	// rather than secureJsonFields so provisioning re-apply does not clobber it.
	ClientCertConfigured bool `json:"clientCertConfigured"`

	// Secure settings (from secureJsonData - decrypted)
	Username    string `json:"-"`
	Password    string `json:"-"`
	Certificate []byte `json:"-"`
	PrivateKey  []byte `json:"-"`

	// Auto-generated client certificate for secure connections (stored in secureJsonData)
	ClientCert []byte `json:"-"`
	ClientKey  []byte `json:"-"`
}

// ParseSettings extracts settings from Grafana data source instance settings
func ParseSettings(settings backend.DataSourceInstanceSettings) (DataSourceSettings, error) {
	var ds DataSourceSettings

	// Parse jsonData
	if err := json.Unmarshal(settings.JSONData, &ds); err != nil {
		return ds, fmt.Errorf("failed to parse jsonData: %w", err)
	}

	// Set defaults
	if ds.Timeout <= 0 {
		ds.Timeout = 10
	}
	if ds.AuthMethod == "" {
		ds.AuthMethod = AuthAnonymous
	}
	if ds.SecurityPolicy == "" {
		ds.SecurityPolicy = "None"
	}
	if ds.SecurityMode == "" {
		ds.SecurityMode = "None"
	}

	// Extract secure data
	if val, ok := settings.DecryptedSecureJSONData["username"]; ok {
		ds.Username = val
	}
	if val, ok := settings.DecryptedSecureJSONData["password"]; ok {
		ds.Password = val
	}
	if val, ok := settings.DecryptedSecureJSONData["certificate"]; ok {
		ds.Certificate = []byte(val)
	}
	if val, ok := settings.DecryptedSecureJSONData["privateKey"]; ok {
		ds.PrivateKey = []byte(val)
	}
	if val, ok := settings.DecryptedSecureJSONData["clientCert"]; ok {
		ds.ClientCert = []byte(val)
	}
	if val, ok := settings.DecryptedSecureJSONData["clientKey"]; ok {
		ds.ClientKey = []byte(val)
	}

	// Backfill: datasources saved before clientCertConfigured was introduced have the
	// cert bytes in secureJsonData but the flag absent from jsonData. Treat them as
	// configured so backend code paths that inspect the flag behave consistently.
	if !ds.ClientCertConfigured && len(ds.ClientCert) > 0 && len(ds.ClientKey) > 0 {
		ds.ClientCertConfigured = true
	}

	return ds, nil
}

// Validate checks if the settings are valid
func (s *DataSourceSettings) Validate() error {
	if s.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	switch s.AuthMethod {
	case AuthAnonymous:
		// No additional validation needed
	case AuthUserPass:
		if s.Username == "" {
			return fmt.Errorf("username is required for username/password authentication")
		}
	case AuthCertificate:
		if len(s.Certificate) == 0 {
			return fmt.Errorf("certificate is required for certificate authentication")
		}
		if len(s.PrivateKey) == 0 {
			return fmt.Errorf("private key is required for certificate authentication")
		}
	default:
		return fmt.Errorf("unsupported auth method: %s", s.AuthMethod)
	}

	return nil
}

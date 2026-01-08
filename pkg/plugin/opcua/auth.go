package opcua

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/gopcua/opcua"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// GetAuthOptions returns authentication options based on the settings
func GetAuthOptions(settings models.DataSourceSettings) ([]opcua.Option, error) {
	switch settings.AuthMethod {
	case models.AuthAnonymous:
		return []opcua.Option{opcua.AuthAnonymous()}, nil

	case models.AuthUserPass:
		if settings.Username == "" {
			return nil, fmt.Errorf("username required for userpass auth")
		}
		return []opcua.Option{
			opcua.AuthUsername(settings.Username, settings.Password),
		}, nil

	case models.AuthCertificate:
		return getCertificateAuthOptions(settings)

	default:
		return nil, fmt.Errorf("unsupported auth method: %s", settings.AuthMethod)
	}
}

// getCertificateAuthOptions parses and returns certificate authentication options
func getCertificateAuthOptions(settings models.DataSourceSettings) ([]opcua.Option, error) {
	if len(settings.Certificate) == 0 {
		return nil, fmt.Errorf("certificate required for certificate auth")
	}
	if len(settings.PrivateKey) == 0 {
		return nil, fmt.Errorf("private key required for certificate auth")
	}

	// Parse certificate
	certBlock, _ := pem.Decode(settings.Certificate)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	key, err := parsePrivateKey(settings.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return []opcua.Option{
		opcua.AuthCertificate(certBlock.Bytes),
		opcua.AuthPrivateKey(key),
		opcua.Certificate(certBlock.Bytes),
		opcua.PrivateKey(key),
		opcua.ApplicationURI(getApplicationURI(cert)),
	}, nil
}

// parsePrivateKey attempts to parse a PEM-encoded private key
func parsePrivateKey(keyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	// Try PKCS1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS8
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key (tried PKCS1 and PKCS8): %w", err)
	}

	rsaKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}

// getApplicationURI extracts the application URI from the certificate
func getApplicationURI(cert *x509.Certificate) string {
	for _, uri := range cert.URIs {
		return uri.String()
	}
	// Fallback to a default URI
	return "urn:grafana:simpleopcua:client"
}

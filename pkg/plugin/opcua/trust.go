package opcua

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

const (
	trustedCertsDir = "trusted"
	rejectedCertsDir = "rejected"
)

// TrustStore manages server certificate trust
type TrustStore struct {
	dataDir string
	logger  log.Logger
}

// NewTrustStore creates a new trust store
func NewTrustStore(dataDir string, logger log.Logger) *TrustStore {
	return &TrustStore{
		dataDir: dataDir,
		logger:  logger,
	}
}

// TrustServerCertificate is called during OPC-UA connection to validate server certificates
// For now, we auto-trust all server certificates (similar to Prosys client behavior)
// In the future, this could be enhanced to check against a list of trusted certificates
func (ts *TrustStore) TrustServerCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("nil certificate")
	}

	// Log the certificate being trusted
	ts.logger.Info("Trusting server certificate",
		"subject", cert.Subject.CommonName,
		"issuer", cert.Issuer.CommonName,
		"notBefore", cert.NotBefore,
		"notAfter", cert.NotAfter,
	)

	// Save certificate to trusted store for future reference
	if err := ts.saveTrustedCert(cert); err != nil {
		// Log but don't fail - we still want to allow the connection
		ts.logger.Warn("Failed to save trusted certificate", "error", err)
	}

	return nil
}

// saveTrustedCert saves a server certificate to the trusted certificates directory
func (ts *TrustStore) saveTrustedCert(cert *x509.Certificate) error {
	trustedDir := filepath.Join(ts.dataDir, trustedCertsDir)
	if err := os.MkdirAll(trustedDir, 0700); err != nil {
		return fmt.Errorf("create trusted certs directory: %w", err)
	}

	// Generate filename from certificate fingerprint
	fingerprint := fmt.Sprintf("%x", cert.SerialNumber)
	certPath := filepath.Join(trustedDir, fingerprint+".pem")

	// Check if already saved
	if _, err := os.Stat(certPath); err == nil {
		return nil // Already exists
	}

	// Encode and save
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		return fmt.Errorf("write certificate: %w", err)
	}

	ts.logger.Debug("Saved trusted certificate", "path", certPath)
	return nil
}

// IsCertificateTrusted checks if a certificate is in the trusted store
func (ts *TrustStore) IsCertificateTrusted(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}

	trustedDir := filepath.Join(ts.dataDir, trustedCertsDir)
	fingerprint := fmt.Sprintf("%x", cert.SerialNumber)
	certPath := filepath.Join(trustedDir, fingerprint+".pem")

	_, err := os.Stat(certPath)
	return err == nil
}

// GetTrustedCertificates returns all trusted server certificates
func (ts *TrustStore) GetTrustedCertificates() ([]*x509.Certificate, error) {
	trustedDir := filepath.Join(ts.dataDir, trustedCertsDir)

	// Check if directory exists
	if _, err := os.Stat(trustedDir); os.IsNotExist(err) {
		return []*x509.Certificate{}, nil
	}

	entries, err := os.ReadDir(trustedDir)
	if err != nil {
		return nil, fmt.Errorf("read trusted certs directory: %w", err)
	}

	var certs []*x509.Certificate
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pem" {
			continue
		}

		certPath := filepath.Join(trustedDir, entry.Name())
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			ts.logger.Warn("Failed to read certificate", "path", certPath, "error", err)
			continue
		}

		block, _ := pem.Decode(certPEM)
		if block == nil {
			ts.logger.Warn("Failed to decode certificate PEM", "path", certPath)
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			ts.logger.Warn("Failed to parse certificate", "path", certPath, "error", err)
			continue
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

// ClearTrustedCertificates removes all trusted certificates
func (ts *TrustStore) ClearTrustedCertificates() error {
	trustedDir := filepath.Join(ts.dataDir, trustedCertsDir)
	if err := os.RemoveAll(trustedDir); err != nil {
		return fmt.Errorf("remove trusted certs directory: %w", err)
	}
	return nil
}

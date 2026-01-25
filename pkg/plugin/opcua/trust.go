package opcua

import (
	"crypto/x509"
	"fmt"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// TrustStore manages server certificate trust in-memory
type TrustStore struct {
	logger  log.Logger
	trusted map[string]*x509.Certificate // fingerprint -> cert (in-memory only)
	mu      sync.RWMutex
}

// NewTrustStore creates a new trust store
func NewTrustStore(logger log.Logger) *TrustStore {
	return &TrustStore{
		logger:  logger,
		trusted: make(map[string]*x509.Certificate),
	}
}

// TrustServerCertificate is called during OPC-UA connection to validate server certificates
// Automatically trusts all server certificates and caches them in memory for the session
func (ts *TrustStore) TrustServerCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("nil certificate")
	}

	// Generate fingerprint for tracking
	fingerprint := fmt.Sprintf("%x", cert.SerialNumber)

	// Log the certificate being trusted
	ts.logger.Info("Trusting server certificate",
		"subject", cert.Subject.CommonName,
		"issuer", cert.Issuer.CommonName,
		"notBefore", cert.NotBefore,
		"notAfter", cert.NotAfter,
		"fingerprint", fingerprint,
	)

	// Store in memory for session tracking
	ts.mu.Lock()
	ts.trusted[fingerprint] = cert
	ts.mu.Unlock()

	ts.logger.Debug("Cached server certificate in memory", "fingerprint", fingerprint)

	return nil
}

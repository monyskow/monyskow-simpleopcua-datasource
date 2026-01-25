package opcua

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Package-level certificate registry
// Key: datasource UID, Value: CertificateManager
var (
	certRegistry   = make(map[string]*CertificateManager)
	certRegistryMu sync.RWMutex
)

// GetCertificateManager returns an existing CertificateManager for the given UID,
// or creates a new one if it doesn't exist. This ensures certificate persistence
// across datasource instance recreations.
func GetCertificateManager(uid string, logger log.Logger) *CertificateManager {
	// Fast path: check if manager exists (read lock)
	certRegistryMu.RLock()
	if mgr, ok := certRegistry[uid]; ok {
		certRegistryMu.RUnlock()
		logger.Debug("Using existing certificate manager from registry", "uid", uid)
		return mgr
	}
	certRegistryMu.RUnlock()

	// Slow path: create new manager (write lock)
	certRegistryMu.Lock()
	defer certRegistryMu.Unlock()

	// Double-check after acquiring write lock
	if mgr, ok := certRegistry[uid]; ok {
		logger.Debug("Using existing certificate manager (double-check)", "uid", uid)
		return mgr
	}

	// Create new manager
	mgr := NewCertificateManager(logger)
	certRegistry[uid] = mgr
	logger.Info("Created new certificate manager for datasource", "uid", uid)
	return mgr
}

// RemoveCertificateManager removes the CertificateManager for a given UID.
// Used for testing cleanup.
func RemoveCertificateManager(uid string) {
	certRegistryMu.Lock()
	defer certRegistryMu.Unlock()
	delete(certRegistry, uid)
}

const (
	keyBits      = 2048
	certValidity = 365 * 24 * time.Hour * 3 // 3 years
	hostname     = "grafana-opcua-client"
)

// CertificateManager manages client certificates for OPC-UA connections in-memory
type CertificateManager struct {
	logger   log.Logger
	certPEM  []byte
	keyPEM   []byte
	notAfter time.Time
	mu       sync.RWMutex
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(logger log.Logger) *CertificateManager {
	return &CertificateManager{
		logger: logger,
	}
}

// GetOrCreate returns existing certificate from memory or generates a new one
// Certificates are cached in memory and regenerated on plugin restart
func (cm *CertificateManager) GetOrCreate() (certPEM, keyPEM []byte, err error) {
	// Check if we have a valid certificate in memory
	cm.mu.RLock()
	if cm.certPEM != nil && cm.keyPEM != nil {
		// Check if certificate is still valid (not expiring within 30 days)
		if time.Now().Add(30 * 24 * time.Hour).Before(cm.notAfter) {
			certPEM = cm.certPEM
			keyPEM = cm.keyPEM
			cm.mu.RUnlock()
			cm.logger.Debug("Using cached certificate from memory")
			return certPEM, keyPEM, nil
		}
	}
	cm.mu.RUnlock()

	// Need to generate new certificate
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock
	if cm.certPEM != nil && cm.keyPEM != nil {
		if time.Now().Add(30 * 24 * time.Hour).Before(cm.notAfter) {
			cm.logger.Debug("Using cached certificate from memory (double-check)")
			return cm.certPEM, cm.keyPEM, nil
		}
	}

	cm.logger.Info("Generating new in-memory OPC-UA client certificate")
	certPEM, keyPEM, notAfter, err := generateCertificate()
	if err != nil {
		return nil, nil, fmt.Errorf("generate certificate: %w", err)
	}

	// Cache in memory
	cm.certPEM = certPEM
	cm.keyPEM = keyPEM
	cm.notAfter = notAfter

	cm.logger.Info("Generated new OPC-UA client certificate", "validUntil", notAfter.Format(time.RFC3339))
	return certPEM, keyPEM, nil
}

// GetApplicationURI returns the application URI for the generated certificate
func (cm *CertificateManager) GetApplicationURI() string {
	return fmt.Sprintf("urn:%s:grafana:simpleopcua:client", hostname)
}

// generateCertificate creates a new self-signed certificate for OPC-UA client
func generateCertificate() (certPEM, keyPEM []byte, notAfter time.Time, err error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("generate RSA key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("generate serial number: %w", err)
	}

	applicationURI := fmt.Sprintf("urn:%s:grafana:simpleopcua:client", hostname)
	notBefore := time.Now().Add(-1 * time.Hour) // Valid from 1 hour ago
	notAfter = time.Now().Add(certValidity)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Grafana Simple OPC-UA Client",
			Organization: []string{"Grafana"},
			Country:      []string{"US"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Add Subject Alternative Names
	// Add application URI
	appURI, _ := url.Parse(applicationURI)
	if appURI != nil {
		template.URIs = append(template.URIs, appURI)
	}

	// Add DNS names
	template.DNSNames = []string{hostname, "localhost"}

	// Add IP addresses
	template.IPAddresses = []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("::1"),
	}

	// Try to get local IP addresses
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					template.IPAddresses = append(template.IPAddresses, ipnet.IP)
				}
			}
		}
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("create certificate: %w", err)
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM, notAfter, nil
}

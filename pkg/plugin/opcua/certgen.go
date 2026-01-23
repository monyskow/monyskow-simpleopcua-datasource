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
	"os"
	"path/filepath"
	"time"
)

const (
	certFileName = "client_cert.pem"
	keyFileName  = "client_key.pem"
	keyBits      = 2048
	certValidity = 365 * 24 * time.Hour * 3 // 3 years
)

// CertificateStore manages client certificates for OPC-UA connections
type CertificateStore struct {
	dataDir string
}

// NewCertificateStore creates a new certificate store
func NewCertificateStore(dataDir string) *CertificateStore {
	return &CertificateStore{
		dataDir: dataDir,
	}
}

// GetOrCreateCertificate returns existing certificate or generates a new one
func (cs *CertificateStore) GetOrCreateCertificate() (certPEM, keyPEM []byte, err error) {
	certPath := filepath.Join(cs.dataDir, certFileName)
	keyPath := filepath.Join(cs.dataDir, keyFileName)

	// Check if certificate exists and is valid
	if cs.certificateExists(certPath, keyPath) {
		certPEM, err = os.ReadFile(certPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read certificate: %w", err)
		}
		keyPEM, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read private key: %w", err)
		}

		// Verify certificate is still valid
		if cs.isCertificateValid(certPEM) {
			return certPEM, keyPEM, nil
		}
	}

	// Generate new certificate
	certPEM, keyPEM, err = cs.generateCertificate()
	if err != nil {
		return nil, nil, fmt.Errorf("generate certificate: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(cs.dataDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("create data directory: %w", err)
	}

	// Save certificate and key
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		return nil, nil, fmt.Errorf("save certificate: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, nil, fmt.Errorf("save private key: %w", err)
	}

	return certPEM, keyPEM, nil
}

// certificateExists checks if certificate and key files exist
func (cs *CertificateStore) certificateExists(certPath, keyPath string) bool {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// isCertificateValid checks if the certificate is not expired
func (cs *CertificateStore) isCertificateValid(certPEM []byte) bool {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	// Check if certificate expires within 30 days
	return time.Now().Add(30 * 24 * time.Hour).Before(cert.NotAfter)
}

// generateCertificate creates a new self-signed certificate for OPC-UA client
func (cs *CertificateStore) generateCertificate() (certPEM, keyPEM []byte, err error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, fmt.Errorf("generate RSA key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial number: %w", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "grafana-opcua-client"
	}

	applicationURI := fmt.Sprintf("urn:%s:grafana:simpleopcua:client", hostname)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Grafana Simple OPC-UA Client",
			Organization: []string{"Grafana"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour), // Valid from 1 hour ago
		NotAfter:              time.Now().Add(certValidity),
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
		return nil, nil, fmt.Errorf("create certificate: %w", err)
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

	return certPEM, keyPEM, nil
}

// GetApplicationURI returns the application URI for the generated certificate
func (cs *CertificateStore) GetApplicationURI() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "grafana-opcua-client"
	}
	return fmt.Sprintf("urn:%s:grafana:simpleopcua:client", hostname)
}

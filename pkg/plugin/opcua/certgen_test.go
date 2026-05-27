package opcua

import (
	"crypto/x509"
	"encoding/pem"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificateGeneration(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)

	// Generate certificate
	certPEM, keyPEM, err := cm.GetOrCreate()
	require.NoError(t, err)
	assert.NotEmpty(t, certPEM)
	assert.NotEmpty(t, keyPEM)

	// Should use cached version on second call
	certPEM2, keyPEM2, err := cm.GetOrCreate()
	require.NoError(t, err)
	assert.Equal(t, certPEM, certPEM2, "should return same certificate from cache")
	assert.Equal(t, keyPEM, keyPEM2, "should return same key from cache")
}

func TestCertificateValidity(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)
	certPEM, _, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Parse and validate
	block, _ := pem.Decode(certPEM)
	require.NotNil(t, block, "should decode PEM block")
	assert.Equal(t, "CERTIFICATE", block.Type)

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	// Verify certificate fields
	assert.Equal(t, "Grafana Simple OPC-UA Client", cert.Subject.CommonName)
	assert.Contains(t, cert.Subject.Organization, "Grafana")
	assert.Contains(t, cert.Subject.Country, "US")

	// Verify validity period
	assert.True(t, cert.NotBefore.Before(time.Now()), "certificate should be valid now")
	assert.True(t, cert.NotAfter.After(time.Now()), "certificate should not be expired")

	// Verify it's valid for at least 2 years (we set 3 years validity)
	expectedExpiry := time.Now().Add(2 * 365 * 24 * time.Hour)
	assert.True(t, cert.NotAfter.After(expectedExpiry), "certificate should be valid for at least 2 years")

	// Verify key usage
	assert.True(t, cert.KeyUsage&x509.KeyUsageDigitalSignature != 0, "should have digital signature usage")
	assert.True(t, cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0, "should have key encipherment usage")

	// Verify extended key usage
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)

	// Verify it's not a CA
	assert.False(t, cert.IsCA, "should not be a CA certificate")

	// Verify DNS names
	assert.Contains(t, cert.DNSNames, hostname)
	assert.Contains(t, cert.DNSNames, "localhost")

	// Verify Application URI
	assert.NotEmpty(t, cert.URIs, "should have application URI")
	if len(cert.URIs) > 0 {
		assert.Contains(t, cert.URIs[0].String(), "urn:"+hostname+":grafana:simpleopcua:client")
	}
}

func TestKeyValidity(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)
	_, keyPEM, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Parse and validate key
	block, _ := pem.Decode(keyPEM)
	require.NotNil(t, block, "should decode PEM block")
	assert.Equal(t, "RSA PRIVATE KEY", block.Type)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	// Verify key size
	assert.Equal(t, keyBits, key.N.BitLen(), "key should have correct bit size")
}

func TestCertificateExpiration(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)

	// Generate initial certificate
	certPEM1, keyPEM1, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Manually set certificate as expired (simulate expiration)
	cm.mu.Lock()
	cm.notAfter = time.Now().Add(-1 * time.Hour) // Set to past
	cm.mu.Unlock()

	// Should regenerate certificate when expired
	certPEM2, keyPEM2, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Verify new certificate was generated
	assert.NotEqual(t, certPEM1, certPEM2, "should generate new certificate when expired")
	assert.NotEqual(t, keyPEM1, keyPEM2, "should generate new key when expired")

	// Verify new certificate is valid
	block2, _ := pem.Decode(certPEM2)
	cert2, err := x509.ParseCertificate(block2.Bytes)
	require.NoError(t, err)
	assert.True(t, cert2.NotAfter.After(time.Now()), "new certificate should be valid")
}

func TestConcurrentAccess(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)

	// Launch multiple goroutines to access certificate concurrently
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([][]byte, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			certPEM, _, err := cm.GetOrCreate()
			results[index] = certPEM
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all succeeded
	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errors[i], "goroutine %d should succeed", i)
		assert.NotEmpty(t, results[i], "goroutine %d should get certificate", i)
	}

	// Verify all got the same certificate (from cache)
	for i := 1; i < numGoroutines; i++ {
		assert.Equal(t, results[0], results[i], "all goroutines should get same certificate")
	}
}

func TestApplicationURI(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)

	expectedURI := "urn:" + hostname + ":grafana:simpleopcua:client"
	actualURI := cm.GetApplicationURI()

	assert.Equal(t, expectedURI, actualURI)
}

func TestPEMEncoding(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)
	certPEM, keyPEM, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Verify certificate PEM
	certBlock, rest := pem.Decode(certPEM)
	assert.NotNil(t, certBlock, "should decode certificate PEM")
	assert.Empty(t, rest, "should have no trailing data")
	assert.Equal(t, "CERTIFICATE", certBlock.Type)

	// Verify key PEM
	keyBlock, rest := pem.Decode(keyPEM)
	assert.NotNil(t, keyBlock, "should decode key PEM")
	assert.Empty(t, rest, "should have no trailing data")
	assert.Equal(t, "RSA PRIVATE KEY", keyBlock.Type)
}

func TestCertificateRenewal(t *testing.T) {
	cm := NewCertificateManager(log.DefaultLogger)

	// Get initial certificate
	certPEM1, _, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Set certificate to expire in 25 days (less than 30 day threshold)
	cm.mu.Lock()
	cm.notAfter = time.Now().Add(25 * 24 * time.Hour)
	cm.mu.Unlock()

	// Should regenerate certificate when expiring soon
	certPEM2, _, err := cm.GetOrCreate()
	require.NoError(t, err)

	// Should have generated new certificate
	assert.NotEqual(t, certPEM1, certPEM2, "should regenerate certificate when expiring soon")
}

func TestCertificateReuseAcrossClients(t *testing.T) {
	logger := log.DefaultLogger
	certMgr := NewCertificateManager(logger)

	// First certificate generation
	certPEM1, keyPEM1, err := certMgr.GetOrCreate()
	require.NoError(t, err)

	// Simulate creating a second client with same CertificateManager
	// (as would happen on reconnection)
	certPEM2, keyPEM2, err := certMgr.GetOrCreate()
	require.NoError(t, err)

	// Should be exactly the same certificate
	assert.Equal(t, certPEM1, certPEM2, "certificate should be reused")
	assert.Equal(t, keyPEM1, keyPEM2, "key should be reused")
}

func TestCertificateRegistryPersistence(t *testing.T) {
	logger := log.DefaultLogger
	uid := "test-ds-uid-persistence"

	// Clean up from any previous test runs
	RemoveCertificateManager(uid)

	// Get certificate manager (creates new)
	mgr1 := GetCertificateManager(uid, logger)
	require.NotNil(t, mgr1)

	// Generate a certificate
	certPEM1, _, err := mgr1.GetOrCreate()
	require.NoError(t, err)

	// Get certificate manager again (simulates "Save & Test" creating new instance)
	mgr2 := GetCertificateManager(uid, logger)
	require.Same(t, mgr1, mgr2, "should return same manager instance")

	// Certificate should be the same
	certPEM2, _, err := mgr2.GetOrCreate()
	require.NoError(t, err)
	assert.Equal(t, certPEM1, certPEM2, "certificate should persist across instance recreations")

	// Cleanup
	RemoveCertificateManager(uid)
}

func TestCertificateRegistrySeparation(t *testing.T) {
	logger := log.DefaultLogger
	uid1 := "ds-uid-1"
	uid2 := "ds-uid-2"

	// Clean up
	RemoveCertificateManager(uid1)
	RemoveCertificateManager(uid2)

	// Get managers for different UIDs
	mgr1 := GetCertificateManager(uid1, logger)
	mgr2 := GetCertificateManager(uid2, logger)

	// Should be different instances
	require.NotSame(t, mgr1, mgr2, "different UIDs should have different managers")

	// Certificates should be different
	cert1, _, _ := mgr1.GetOrCreate()
	cert2, _, _ := mgr2.GetOrCreate()
	assert.NotEqual(t, cert1, cert2, "different datasources should have different certificates")

	// Cleanup
	RemoveCertificateManager(uid1)
	RemoveCertificateManager(uid2)
}

func TestGenerateCertificate_KeyUsageIncludesNonRepudiation(t *testing.T) {
	// Regression test for #68: ProSys (and other OPC-UA strict-compliant servers)
	// reject client certs without nonRepudiation (Go: ContentCommitment).
	// See: OPC-UA Part 6 §6.2.2.
	certPEM, _, _, err := generateCertificate()
	if err != nil {
		t.Fatalf("generateCertificate: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("pem.Decode returned nil")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("x509.ParseCertificate: %v", err)
	}
	if cert.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		t.Errorf("KeyUsage missing ContentCommitment (nonRepudiation); got %b", cert.KeyUsage)
	}
	// Sanity — remaining required flags must also be present.
	required := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment
	if cert.KeyUsage&required != required {
		t.Errorf("KeyUsage missing one of DigitalSignature|KeyEncipherment|DataEncipherment; got %b", cert.KeyUsage)
	}
}

func TestCertificateRegistryConcurrentAccess(t *testing.T) {
	logger := log.DefaultLogger
	uid := "concurrent-test-uid"

	// Clean up
	RemoveCertificateManager(uid)

	var wg sync.WaitGroup
	const numGoroutines = 20
	managers := make([]*CertificateManager, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			managers[idx] = GetCertificateManager(uid, logger)
		}(i)
	}
	wg.Wait()

	// All should be the same instance
	for i := 1; i < numGoroutines; i++ {
		require.Same(t, managers[0], managers[i], "all goroutines should get same manager")
	}

	// Cleanup
	RemoveCertificateManager(uid)
}

package opcua

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// ---------------------------------------------------------------------------
// Fix #3 — loadClientCert helper
// ---------------------------------------------------------------------------

// generateTestCertAndKey creates a minimal self-signed RSA cert + matching key
// for use in tests.  It mirrors the structure of generateCertificate() but is
// intentionally smaller (1024-bit) so the tests run quickly.
func generateTestCertAndKey(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 64))
	require.NoError(t, err)

	appURI, err := url.Parse("urn:test:grafana:opcua:client")
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "test-opcua-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		URIs:         []*url.URL{appURI},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM
}

func TestLoadClientCert_ValidPair_ReturnsParsedValues(t *testing.T) {
	certPEM, keyPEM := generateTestCertAndKey(t)

	cert, key, err := loadClientCert(certPEM, keyPEM)

	require.NoError(t, err)
	require.NotNil(t, cert)
	require.NotNil(t, key)
	assert.Equal(t, "test-opcua-client", cert.Subject.CommonName)
	// Confirm the returned key's public component matches the cert's public key.
	certPub := cert.PublicKey.(*rsa.PublicKey)
	assert.Equal(t, 0, certPub.N.Cmp(key.N), "public key modulus must match cert")
	assert.Equal(t, certPub.E, key.E, "public key exponent must match cert")
}

func TestLoadClientCert_MismatchedKey_ReturnsValidationError(t *testing.T) {
	certPEM, _ := generateTestCertAndKey(t)
	// Generate a completely different key — the public key won't match the cert.
	_, differentKeyPEM := generateTestCertAndKey(t)

	_, _, err := loadClientCert(certPEM, differentKeyPEM)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "certificate and key validation failed",
		"error message should reference the validation failure")
}

func TestLoadClientCert_GarbageCertPEM_ReturnsError(t *testing.T) {
	_, keyPEM := generateTestCertAndKey(t)

	_, _, err := loadClientCert([]byte("not a pem block"), keyPEM)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode", "should report decode failure")
}

func TestLoadClientCert_GarbageKeyPEM_ReturnsError(t *testing.T) {
	certPEM, _ := generateTestCertAndKey(t)

	_, _, err := loadClientCert(certPEM, []byte("not a pem block"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode", "should report key decode failure")
}

func TestLoadClientCert_WrongPEMType_ReturnsError(t *testing.T) {
	// Encode valid DER bytes under the wrong PEM type header.
	_, keyPEM := generateTestCertAndKey(t)
	wrongTypePEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("irrelevant")})

	_, _, err := loadClientCert(wrongTypePEM, keyPEM)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode certificate PEM")
}

func TestLoadClientCert_EmptyInputs_DoNotPanic(t *testing.T) {
	// Regression: nil/empty slices must not panic — they should return an error.
	require.NotPanics(t, func() {
		_, _, err := loadClientCert(nil, nil)
		assert.Error(t, err)
	})

	require.NotPanics(t, func() {
		_, _, err := loadClientCert([]byte{}, []byte{})
		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// Fix #2 — Close(ctx context.Context)
// ---------------------------------------------------------------------------

// newDisconnectedClient builds a *Client with connected=false and a nil inner
// client.  This is sufficient to exercise the not-connected early-return paths.
func newDisconnectedClient() *Client {
	return &Client{
		client:    nil,
		connected: false,
		logger:    log.DefaultLogger,
		settings:  models.DataSourceSettings{Endpoint: "opc.tcp://localhost:4840"},
	}
}

func TestClose_NotConnected_ReturnsNilImmediately(t *testing.T) {
	c := newDisconnectedClient()

	err := c.Close(context.Background())

	assert.NoError(t, err, "Close on an unconnected client must return nil")
}

func TestClose_NotConnected_IsIdempotent(t *testing.T) {
	c := newDisconnectedClient()

	// Calling Close multiple times must not panic or error.
	require.NoError(t, c.Close(context.Background()))
	require.NoError(t, c.Close(context.Background()))
}

func TestClose_NotConnected_RespectsAlreadyCancelledCtx(t *testing.T) {
	c := newDisconnectedClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	// The early-return path never inspects ctx, so this must still succeed.
	err := c.Close(ctx)
	assert.NoError(t, err, "not-connected early return should not care about ctx state")
}

// ---------------------------------------------------------------------------
// Fix #1 — Lock scope narrowed in Browse + ReadNodes
// ---------------------------------------------------------------------------

// TestReadNodes_NotConnected_ReturnsErrorWithoutPanic verifies that the
// not-connected early-return correctly releases the read lock and doesn't
// panic with a nil inner client.
func TestReadNodes_NotConnected_ReturnsErrorWithoutPanic(t *testing.T) {
	c := newDisconnectedClient()

	nodes := []models.NodeQuery{{NodeID: "ns=1;i=1000", DisplayName: "TestNode"}}

	require.NotPanics(t, func() {
		_, err := c.ReadNodes(context.Background(), nodes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

func TestReadNodes_NotConnected_NilNodes_StillErrorsBeforeProcessing(t *testing.T) {
	// The not-connected guard fires before the empty-slice check, so a nil node
	// list still returns an error rather than an empty slice.
	c := newDisconnectedClient()

	_, err := c.ReadNodes(context.Background(), nil)
	assert.Error(t, err, "should fail on not-connected even with nil node list")
	assert.Contains(t, err.Error(), "not connected")
}

func TestBrowse_NotConnected_ReturnsErrorWithoutPanic(t *testing.T) {
	c := newDisconnectedClient()

	require.NotPanics(t, func() {
		_, err := c.Browse(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not connected")
	})
}

// TestConcurrentBrowseRead_NoStarvation_RaceClean spawns 100 goroutines each
// calling Browse and ReadNodes concurrently against a disconnected client.
// This exercises the narrowed lock path under -race.
//
// Limitation: because there is no injectable interface for the inner gopcua
// client, we cannot simulate a slow network call to prove the OLD code would
// deadlock.  This is a smoke test: it confirms (a) no race detector violations,
// (b) all goroutines complete within the timeout, and (c) the lock is released
// correctly on the not-connected path.  A true regression test would require a
// seam (interface injection) that does not currently exist in the production code.
func TestConcurrentBrowseRead_NoStarvation_RaceClean(t *testing.T) {
	const numGoroutines = 100
	c := newDisconnectedClient()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	browseErrs := make([]error, numGoroutines)
	readErrs := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		idx := i
		go func() {
			defer wg.Done()
			_, err := c.Browse(ctx, "ns=0;i=85")
			browseErrs[idx] = err
		}()
		go func() {
			defer wg.Done()
			_, err := c.ReadNodes(ctx, []models.NodeQuery{{NodeID: "ns=1;i=1"}})
			readErrs[idx] = err
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed in time — no starvation.
	case <-ctx.Done():
		t.Fatal("goroutines did not complete within 5s — possible starvation or deadlock")
	}

	// Every goroutine must have received the expected "not connected" error.
	for i, err := range browseErrs {
		require.Error(t, err, "Browse goroutine %d should return an error", i)
		assert.Contains(t, err.Error(), "not connected", "Browse goroutine %d unexpected error: %v", i, err)
	}
	for i, err := range readErrs {
		require.Error(t, err, "ReadNodes goroutine %d should return an error", i)
		assert.Contains(t, err.Error(), "not connected", "ReadNodes goroutine %d unexpected error: %v", i, err)
	}
}

// TestIsConnected_AfterClose_ReturnsFalse verifies that IsConnected (which uses
// mu.RLock) is not blocked by any lingering write lock from Close.
func TestIsConnected_AfterClose_ReturnsFalse(t *testing.T) {
	c := newDisconnectedClient()

	_ = c.Close(context.Background())

	assert.False(t, c.IsConnected(), "IsConnected must return false after Close on unconnected client")
}

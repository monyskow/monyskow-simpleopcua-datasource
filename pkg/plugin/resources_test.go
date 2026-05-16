package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/opcua"
)

// newDatasourceForResourceTests returns the minimal Datasource needed to call
// handleGenerateCertificate.  It only accesses d.certMgr and d.logger.
func newDatasourceForResourceTests() *Datasource {
	return &Datasource{
		certMgr: opcua.NewCertificateManager(log.DefaultLogger),
		logger:  log.DefaultLogger,
	}
}

// ---------------------------------------------------------------------------
// Fix #1 — HTTP encode-first (encode to buffer before writing headers)
//
// The bug: when json.Encoder wrote directly to http.ResponseWriter, the
// 200 + Content-Type header was already sent before a potential encode error.
// Fix: encode to bytes.Buffer first; only set headers and write body if encode
// succeeds.
//
// Happy-path test: Content-Type and 200 are set correctly, body is valid JSON.
// The encode-fail path cannot be triggered without injecting a failing encoder —
// this limitation is noted in the test output.
// ---------------------------------------------------------------------------

func TestHandleGenerateCertificate_HappyPath_Returns200WithJSONContentType(t *testing.T) {
	// Fix #1: headers must only be committed after a successful encode.
	d := newDatasourceForResourceTests()

	req := httptest.NewRequest(http.MethodGet, "/generate-certificate", nil)
	w := httptest.NewRecorder()

	d.handleGenerateCertificate(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"successful certificate generation must return 200")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"),
		"Content-Type must be set to application/json only after successful encode")
}

func TestHandleGenerateCertificate_HappyPath_BodyIsValidCertificateResponse(t *testing.T) {
	d := newDatasourceForResourceTests()

	req := httptest.NewRequest(http.MethodGet, "/generate-certificate", nil)
	w := httptest.NewRecorder()

	d.handleGenerateCertificate(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var cr CertificateResponse
	err := json.Unmarshal(w.Body.Bytes(), &cr)
	require.NoError(t, err, "response body must be valid JSON decodable into CertificateResponse")
	assert.NotEmpty(t, cr.ClientCert, "ClientCert field must be non-empty")
	assert.NotEmpty(t, cr.ClientKey, "ClientKey field must be non-empty")
}

func TestHandleGenerateCertificate_HappyPath_CertAndKeyArePEMEncoded(t *testing.T) {
	// Verify the returned strings actually look like PEM (start with "-----BEGIN").
	d := newDatasourceForResourceTests()

	req := httptest.NewRequest(http.MethodGet, "/generate-certificate", nil)
	w := httptest.NewRecorder()

	d.handleGenerateCertificate(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var cr CertificateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &cr))

	assert.Contains(t, cr.ClientCert, "-----BEGIN CERTIFICATE-----",
		"ClientCert must be PEM-encoded")
	assert.Contains(t, cr.ClientKey, "-----BEGIN RSA PRIVATE KEY-----",
		"ClientKey must be PEM-encoded RSA key")
}

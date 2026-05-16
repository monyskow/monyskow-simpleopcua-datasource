package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/opcua"
)

// CallResource handles HTTP-like requests from the frontend
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return httpadapter.New(d.resourceHandler()).CallResource(ctx, req, sender)
}

// resourceHandler returns the HTTP handler for resource requests
func (d *Datasource) resourceHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/browse", d.handleBrowse)
	mux.HandleFunc("/endpoints", d.handleGetEndpoints)
	mux.HandleFunc("/generate-certificate", d.handleGenerateCertificate)

	return mux
}

// handleBrowse handles the browse nodes request
func (d *Datasource) handleBrowse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get node ID from query parameter
	nodeID := r.URL.Query().Get("nodeId")

	// Get or create client
	client, err := d.getOrCreateClient(ctx)
	if err != nil {
		d.logger.Error("Failed to get client for browse", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Browse nodes
	nodes, err := client.Browse(ctx, nodeID)
	if err != nil {
		d.logger.Error("Failed to browse nodes", "nodeId", nodeID, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(nodes); err != nil {
		d.logger.Error("Failed to encode browse response", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes()) //nolint:errcheck
}

// handleGetEndpoints handles the get endpoints request
func (d *Datasource) handleGetEndpoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get endpoints from server
	endpoints, err := opcua.GetEndpoints(ctx, d.settings.Endpoint)
	if err != nil {
		d.logger.Error("Failed to get endpoints", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(endpoints); err != nil {
		d.logger.Error("Failed to encode endpoints response", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes()) //nolint:errcheck
}

// CertificateResponse is the response from the generate-certificate endpoint
type CertificateResponse struct {
	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
}

// handleGenerateCertificate generates a new client certificate and returns it
// The frontend should store this in secureJsonData for persistence
func (d *Datasource) handleGenerateCertificate(w http.ResponseWriter, r *http.Request) {
	d.logger.Info("Generating new client certificate for secure connection")

	// Generate new certificate using the certificate manager
	certPEM, keyPEM, err := d.certMgr.GetOrCreate()
	if err != nil {
		d.logger.Error("Failed to generate certificate", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return certificate and key as JSON
	response := CertificateResponse{
		ClientCert: string(certPEM),
		ClientKey:  string(keyPEM),
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(response); err != nil {
		d.logger.Error("Failed to encode certificate response", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes()) //nolint:errcheck

	d.logger.Info("Certificate generated successfully")
}

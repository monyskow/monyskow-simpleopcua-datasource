package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/opcua"
)

const pluginID = "monyskow-simpleopcua-datasource"

// Ensure Datasource implements required interfaces
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is the backend implementation for the OPC-UA data source
type Datasource struct {
	settings models.DataSourceSettings
	client   *opcua.Client
	logger   log.Logger
	mu       sync.RWMutex
	certMgr  *opcua.CertificateManager // One certificate manager per datasource
	uid      string                    // Datasource UID for logging
}

// NewDatasource creates a new datasource instance
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger := log.DefaultLogger.With("pluginId", pluginID, "uid", settings.UID)

	dsSettings, err := models.ParseSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	logger.Info("Creating new datasource instance",
		"endpoint", dsSettings.Endpoint,
		"uid", settings.UID,
		"hasClientCert", len(dsSettings.ClientCert) > 0,
		"hasClientKey", len(dsSettings.ClientKey) > 0,
		"clientCertLen", len(dsSettings.ClientCert),
	)

	// Get or create certificate manager from GLOBAL registry (persists across instances)
	certMgr := opcua.GetCertificateManager(settings.UID, logger)

	return &Datasource{
		settings: dsSettings,
		logger:   logger,
		certMgr:  certMgr,
		uid:      settings.UID,
	}, nil
}

// Dispose cleans up datasource resources when the instance is disposed
func (d *Datasource) Dispose() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client != nil {
		if err := d.client.Close(); err != nil {
			d.logger.Error("Error closing client", "error", err)
		}
		d.client = nil
	}

	// NOTE: Do NOT remove certificate manager from registry here.
	// Dispose() is called on instance recreation, not datasource deletion.
	// Certificate should persist across "Save & Test" clicks.

	d.logger.Info("Datasource instance disposed", "uid", d.uid)
}

// getOrCreateClient returns an existing client or creates a new one
func (d *Datasource) getOrCreateClient(ctx context.Context) (*opcua.Client, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Return existing connected client
	if d.client != nil && d.client.IsConnected() {
		return d.client, nil
	}

	// Close existing client if disconnected
	if d.client != nil {
		_ = d.client.Close() // Ignore error on cleanup
		d.client = nil
	}

	// Create new client
	client, err := opcua.NewClient(d.settings, d.logger, d.certMgr)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		_ = client.Close() // Ignore error on cleanup
		return nil, fmt.Errorf("connect: %w", err)
	}

	d.client = client
	return client, nil
}

// QueryData handles multiple queries and returns data frames
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	d.logger.Debug("QueryData called",
		"numQueries", len(req.Queries),
		"endpoint", d.settings.Endpoint)

	response := backend.NewQueryDataResponse()

	// Get or create client
	client, err := d.getOrCreateClient(ctx)
	if err != nil {
		d.logger.Error("Failed to get or create client",
			"endpoint", d.settings.Endpoint,
			"error", err)
		// Return error for all queries if we can't connect
		for _, q := range req.Queries {
			response.Responses[q.RefID] = backend.ErrDataResponse(
				backend.StatusBadGateway,
				fmt.Sprintf("connection failed: %s", err.Error()),
			)
		}
		return response, nil
	}

	d.logger.Debug("Client ready, processing queries",
		"endpoint", d.settings.Endpoint,
		"isConnected", client.IsConnected())

	// Process each query
	for _, q := range req.Queries {
		res := d.query(ctx, client, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// query processes a single query
func (d *Datasource) query(ctx context.Context, client *opcua.Client, query backend.DataQuery) backend.DataResponse {
	// Parse query
	var opcQuery models.OpcuaQuery
	if err := json.Unmarshal(query.JSON, &opcQuery); err != nil {
		d.logger.Error("Failed to parse query JSON",
			"refId", query.RefID,
			"error", err,
			"json", string(query.JSON))
		return backend.ErrDataResponse(
			backend.StatusBadRequest,
			fmt.Sprintf("invalid query: %s", err.Error()),
		)
	}

	d.logger.Debug("Parsed query",
		"refId", query.RefID,
		"numNodes", len(opcQuery.Nodes))

	// Skip empty queries
	if len(opcQuery.Nodes) == 0 {
		d.logger.Debug("Skipping empty query", "refId", query.RefID)
		return backend.DataResponse{}
	}

	// Log node IDs being read
	for i, node := range opcQuery.Nodes {
		d.logger.Debug("Reading node",
			"refId", query.RefID,
			"index", i,
			"nodeId", node.NodeID,
			"displayName", node.DisplayName)
	}

	// Read node values
	values, err := client.ReadNodes(ctx, opcQuery.Nodes)
	if err != nil {
		d.logger.Error("Failed to read nodes",
			"refId", query.RefID,
			"error", err)
		return backend.ErrDataResponse(
			backend.StatusInternal,
			fmt.Sprintf("read failed: %s", err.Error()),
		)
	}

	// Log read values
	for i, val := range values {
		d.logger.Debug("Read value",
			"refId", query.RefID,
			"index", i,
			"nodeId", val.NodeID,
			"value", val.Value,
			"status", val.StatusCode)
	}

	// Build data frame
	frame := d.buildDataFrame(opcQuery.Nodes, values)

	d.logger.Debug("Built data frame",
		"refId", query.RefID,
		"numFields", len(frame.Fields))

	return backend.DataResponse{
		Frames: data.Frames{frame},
	}
}

// buildDataFrame creates a data frame from node values
func (d *Datasource) buildDataFrame(nodes []models.NodeQuery, values []opcua.NodeValue) *data.Frame {
	frame := data.NewFrame("response")

	// Add timestamp field
	timestamps := make([]time.Time, 1)
	timestamps[0] = time.Now()
	frame.Fields = append(frame.Fields, data.NewField("time", nil, timestamps))

	// Add value fields
	for i, node := range nodes {
		fieldName := node.GetFieldName()

		if i < len(values) {
			value := values[i]
			field := d.createField(fieldName, value.Value)
			frame.Fields = append(frame.Fields, field)
		}
	}

	return frame
}

// createField creates a data field from a value
func (d *Datasource) createField(name string, value interface{}) *data.Field {
	switch v := value.(type) {
	case float64:
		return data.NewField(name, nil, []float64{v})
	case float32:
		return data.NewField(name, nil, []float64{float64(v)})
	case int64:
		return data.NewField(name, nil, []int64{v})
	case int32:
		return data.NewField(name, nil, []int64{int64(v)})
	case int16:
		return data.NewField(name, nil, []int64{int64(v)})
	case int8:
		return data.NewField(name, nil, []int64{int64(v)})
	case int:
		return data.NewField(name, nil, []int64{int64(v)})
	case uint64:
		return data.NewField(name, nil, []uint64{v})
	case uint32:
		return data.NewField(name, nil, []uint64{uint64(v)})
	case uint16:
		return data.NewField(name, nil, []uint64{uint64(v)})
	case uint8:
		return data.NewField(name, nil, []uint64{uint64(v)})
	case bool:
		return data.NewField(name, nil, []bool{v})
	case string:
		return data.NewField(name, nil, []string{v})
	case nil:
		return data.NewField(name, nil, []*float64{nil})
	default:
		// Convert unknown types to string
		return data.NewField(name, nil, []string{fmt.Sprintf("%v", v)})
	}
}

// CheckHealth validates the data source configuration
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	d.logger.Info("Checking health", "endpoint", d.settings.Endpoint)

	// Validate settings
	if err := d.settings.Validate(); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Invalid configuration: %s", err.Error()),
		}, nil
	}

	// Try to connect
	client, err := d.getOrCreateClient(ctx)
	if err != nil {
		d.logger.Error("Health check failed during connection", "error", err)
		errMsg := formatConnectionError(err, d.settings)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: errMsg,
		}, nil
	}

	d.logger.Info("Client created successfully, reading server state")

	// Try to read server state
	state, err := client.ReadServerState(ctx)
	if err != nil {
		d.logger.Error("Health check failed reading server state", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to read server status: %s", err.Error()),
		}, nil
	}

	d.logger.Info("Health check successful", "state", state)
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Successfully connected to OPC-UA server (state: %v)", state),
	}, nil
}

// formatConnectionError provides detailed error messages for common OPC-UA connection errors
func formatConnectionError(err error, settings models.DataSourceSettings) string {
	errStr := err.Error()

	// Check for common OPC-UA status codes and provide helpful messages
	if contains(errStr, "StatusBadIdentityTokenRejected") || contains(errStr, "0x80210000") {
		hint := "The server rejected the identity token. "
		switch settings.AuthMethod {
		case models.AuthUserPass:
			hint += "Possible causes: (1) Wrong username or password, (2) User account is disabled or locked, (3) Server requires encrypted password but security mode is 'None' - try using 'Sign' or 'SignAndEncrypt' security mode"
		case models.AuthCertificate:
			hint += "Possible causes: (1) Client certificate not trusted by server, (2) Certificate has expired, (3) Certificate doesn't match server requirements"
		default:
			hint += "Possible causes: (1) Server doesn't allow anonymous access, (2) Try using username/password authentication"
		}
		return fmt.Sprintf("Connection failed: %s. %s", errStr, hint)
	}

	if contains(errStr, "StatusBadUserAccessDenied") || contains(errStr, "0x801F0000") {
		return fmt.Sprintf("Connection failed: %s. The user does not have permission to access the server. Check user permissions on the OPC-UA server.", errStr)
	}

	if contains(errStr, "StatusBadSecurityModeRejected") || contains(errStr, "0x80310000") {
		return fmt.Sprintf("Connection failed: %s. The server rejected the security mode. Try a different security mode (None, Sign, or SignAndEncrypt).", errStr)
	}

	if contains(errStr, "StatusBadSecurityPolicyRejected") || contains(errStr, "0x80550000") {
		return fmt.Sprintf("Connection failed: %s. The server rejected the security policy. Check which security policies the server supports.", errStr)
	}

	if contains(errStr, "StatusBadCertificateUntrusted") || contains(errStr, "0x801A0000") {
		return fmt.Sprintf("Connection failed: %s. The server doesn't trust the client certificate. You may need to add the client certificate to the server's trusted certificates list.", errStr)
	}

	if contains(errStr, "StatusBadTimeout") || contains(errStr, "0x800A0000") {
		return fmt.Sprintf("Connection failed: %s. Connection timed out. Check if the server is reachable and the endpoint URL is correct.", errStr)
	}

	if contains(errStr, "no matching endpoint") {
		return fmt.Sprintf("Connection failed: %s. No server endpoint matches the configured security policy, mode, and authentication method. Check Grafana server logs for available endpoints.", errStr)
	}

	// Default: return the original error
	return fmt.Sprintf("Connection failed: %s", errStr)
}

// contains checks if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

// findSubstring checks if substr exists in s
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

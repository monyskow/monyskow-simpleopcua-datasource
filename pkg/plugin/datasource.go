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
}

// NewDatasource creates a new datasource instance
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger := log.DefaultLogger.With("pluginId", "monyskow-simpleopcua-datasource")

	dsSettings, err := models.ParseSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	logger.Info("Creating new datasource instance", "endpoint", dsSettings.Endpoint)

	return &Datasource{
		settings: dsSettings,
		logger:   logger,
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

	d.logger.Info("Datasource instance disposed")
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
	client, err := opcua.NewClient(d.settings, d.logger)
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
	response := backend.NewQueryDataResponse()

	// Get or create client
	client, err := d.getOrCreateClient(ctx)
	if err != nil {
		// Return error for all queries if we can't connect
		for _, q := range req.Queries {
			response.Responses[q.RefID] = backend.ErrDataResponse(
				backend.StatusBadGateway,
				fmt.Sprintf("connection failed: %s", err.Error()),
			)
		}
		return response, nil
	}

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
		return backend.ErrDataResponse(
			backend.StatusBadRequest,
			fmt.Sprintf("invalid query: %s", err.Error()),
		)
	}

	// Skip empty queries
	if len(opcQuery.Nodes) == 0 {
		return backend.DataResponse{}
	}

	// Read node values
	values, err := client.ReadNodes(ctx, opcQuery.Nodes)
	if err != nil {
		return backend.ErrDataResponse(
			backend.StatusInternal,
			fmt.Sprintf("read failed: %s", err.Error()),
		)
	}

	// Build data frame
	frame := d.buildDataFrame(opcQuery.Nodes, values)

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
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection failed: %s", err.Error()),
		}, nil
	}

	// Try to read server state
	state, err := client.ReadServerState(ctx)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to read server status: %s", err.Error()),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Successfully connected to OPC-UA server (state: %v)", state),
	}, nil
}

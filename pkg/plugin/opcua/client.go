package opcua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// Client wraps the gopcua client with additional functionality
type Client struct {
	client    *opcua.Client
	settings  models.DataSourceSettings
	logger    log.Logger
	mu        sync.RWMutex
	connected bool
}

// NodeValue represents the result of reading a node
type NodeValue struct {
	NodeID          string
	DisplayName     string
	Value           interface{}
	StatusCode      ua.StatusCode
	SourceTimestamp time.Time
	ServerTimestamp time.Time
}

// NewClient creates a new OPC-UA client
func NewClient(settings models.DataSourceSettings, logger log.Logger) (*Client, error) {
	opts := []opcua.Option{
		opcua.RequestTimeout(time.Duration(settings.Timeout) * time.Second),
		opcua.AutoReconnect(true),
		opcua.ReconnectInterval(5 * time.Second),
	}

	// Add security options
	secOpts := getSecurityOptions(settings)
	opts = append(opts, secOpts...)

	// Add authentication options
	authOpts, err := GetAuthOptions(settings)
	if err != nil {
		return nil, fmt.Errorf("auth options: %w", err)
	}
	opts = append(opts, authOpts...)

	c, err := opcua.NewClient(settings.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Client{
		client:   c,
		settings: settings,
		logger:   logger,
	}, nil
}

// Connect establishes connection to the OPC-UA server
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	if err := c.client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	c.connected = true
	c.logger.Info("Connected to OPC-UA server", "endpoint", c.settings.Endpoint)
	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	if err := c.client.Close(context.Background()); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	c.connected = false
	c.logger.Info("Disconnected from OPC-UA server")
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// ReadNodes reads values from multiple nodes
func (c *Client) ReadNodes(ctx context.Context, nodes []models.NodeQuery) ([]NodeValue, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	if len(nodes) == 0 {
		return []NodeValue{}, nil
	}

	// Parse node IDs
	nodeIDs := make([]*ua.NodeID, len(nodes))
	for i, n := range nodes {
		id, err := ua.ParseNodeID(n.NodeID)
		if err != nil {
			return nil, fmt.Errorf("parse node %s: %w", n.NodeID, err)
		}
		nodeIDs[i] = id
	}

	// Build read request
	req := &ua.ReadRequest{
		MaxAge:             0,
		NodesToRead:        make([]*ua.ReadValueID, len(nodeIDs)),
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	for i, id := range nodeIDs {
		req.NodesToRead[i] = &ua.ReadValueID{
			NodeID:      id,
			AttributeID: ua.AttributeIDValue,
		}
	}

	// Execute read
	resp, err := c.client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	// Process results
	values := make([]NodeValue, len(resp.Results))
	for i, result := range resp.Results {
		var value interface{}
		if result.Value != nil {
			value = result.Value.Value()
		}

		values[i] = NodeValue{
			NodeID:          nodes[i].NodeID,
			DisplayName:     nodes[i].DisplayName,
			Value:           value,
			StatusCode:      result.Status,
			SourceTimestamp: result.SourceTimestamp,
			ServerTimestamp: result.ServerTimestamp,
		}
	}

	return values, nil
}

// ReadServerState reads the server state node to verify connectivity
func (c *Client) ReadServerState(ctx context.Context) (ua.ServerState, error) {
	nodes := []models.NodeQuery{{
		NodeID:      "ns=0;i=2259", // Server_ServerStatus_State
		DisplayName: "ServerState",
	}}

	values, err := c.ReadNodes(ctx, nodes)
	if err != nil {
		return 0, err
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("no server state returned")
	}

	if values[0].StatusCode != ua.StatusOK {
		return 0, fmt.Errorf("bad status: %v", values[0].StatusCode)
	}

	state, ok := values[0].Value.(int32)
	if !ok {
		return 0, fmt.Errorf("unexpected server state type: %T", values[0].Value)
	}

	return ua.ServerState(state), nil
}

// getSecurityOptions returns security-related client options
func getSecurityOptions(settings models.DataSourceSettings) []opcua.Option {
	var opts []opcua.Option

	if settings.SecurityPolicy != "" && settings.SecurityPolicy != "None" {
		opts = append(opts, opcua.SecurityPolicy(settings.SecurityPolicy))
	}

	if settings.SecurityMode != "" && settings.SecurityMode != "None" {
		opts = append(opts, opcua.SecurityModeString(settings.SecurityMode))
	}

	return opts
}

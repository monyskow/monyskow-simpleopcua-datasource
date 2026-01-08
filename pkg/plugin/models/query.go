package models

// OpcuaQuery represents a query from the frontend
type OpcuaQuery struct {
	RefID         string      `json:"refId"`
	Nodes         []NodeQuery `json:"nodes"`
	MaxDataPoints int64       `json:"maxDataPoints,omitempty"`
}

// NodeQuery represents a single node to read
type NodeQuery struct {
	NodeID      string `json:"nodeId"`      // OPC-UA Node ID (e.g., "ns=2;s=MyVariable")
	DisplayName string `json:"displayName"` // Human-readable name
	Alias       string `json:"alias"`       // Optional alias for the field name
}

// GetFieldName returns the field name to use in the data frame
func (n *NodeQuery) GetFieldName() string {
	if n.Alias != "" {
		return n.Alias
	}
	if n.DisplayName != "" {
		return n.DisplayName
	}
	return n.NodeID
}

// BrowseNode represents a node returned from browsing
type BrowseNode struct {
	NodeID      string `json:"nodeId"`
	DisplayName string `json:"displayName"`
	BrowseName  string `json:"browseName"`
	NodeClass   string `json:"nodeClass"`
	HasChildren bool   `json:"hasChildren"`
}

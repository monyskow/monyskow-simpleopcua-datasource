package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// Browse returns child nodes of the given node
func (c *Client) Browse(ctx context.Context, nodeID string) ([]models.BrowseNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	// Parse node ID (default to Objects folder)
	var nid *ua.NodeID
	var err error
	if nodeID == "" {
		nid = ua.NewNumericNodeID(0, id.ObjectsFolder)
	} else {
		nid, err = ua.ParseNodeID(nodeID)
		if err != nil {
			return nil, fmt.Errorf("parse node id: %w", err)
		}
	}

	// Browse request
	req := &ua.BrowseRequest{
		View: &ua.ViewDescription{},
		NodesToBrowse: []*ua.BrowseDescription{{
			NodeID:          nid,
			BrowseDirection: ua.BrowseDirectionForward,
			ReferenceTypeID: ua.NewNumericNodeID(0, id.HierarchicalReferences),
			IncludeSubtypes: true,
			NodeClassMask:   uint32(ua.NodeClassVariable | ua.NodeClassObject),
			ResultMask:      uint32(ua.BrowseResultMaskAll),
		}},
	}

	resp, err := c.client.Browse(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("browse: %w", err)
	}

	if len(resp.Results) == 0 {
		return []models.BrowseNode{}, nil
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		return nil, fmt.Errorf("browse failed: %v", result.StatusCode)
	}

	// Handle continuation point if there are more results
	nodes := make([]models.BrowseNode, 0, len(result.References))
	nodes = append(nodes, referencesToNodes(result.References)...)

	// Browse next if there's a continuation point
	for len(result.ContinuationPoint) > 0 {
		nextReq := &ua.BrowseNextRequest{
			ContinuationPoints:        [][]byte{result.ContinuationPoint},
			ReleaseContinuationPoints: false,
		}

		nextResp, err := c.client.BrowseNext(ctx, nextReq)
		if err != nil {
			return nil, fmt.Errorf("browse next: %w", err)
		}

		if len(nextResp.Results) == 0 {
			break
		}

		result = nextResp.Results[0]
		if result.StatusCode != ua.StatusOK {
			break
		}

		nodes = append(nodes, referencesToNodes(result.References)...)
	}

	return nodes, nil
}

// referencesToNodes converts OPC-UA references to BrowseNode models
func referencesToNodes(refs []*ua.ReferenceDescription) []models.BrowseNode {
	nodes := make([]models.BrowseNode, len(refs))
	for i, ref := range refs {
		nodeClass := nodeClassToString(ref.NodeClass)
		hasChildren := nodeClass == "Object" // Objects typically have children

		nodes[i] = models.BrowseNode{
			NodeID:      ref.NodeID.NodeID.String(),
			DisplayName: ref.DisplayName.Text,
			BrowseName:  ref.BrowseName.Name,
			NodeClass:   nodeClass,
			HasChildren: hasChildren,
		}
	}
	return nodes
}

// nodeClassToString converts a NodeClass to a string representation
func nodeClassToString(nc ua.NodeClass) string {
	switch nc {
	case ua.NodeClassObject:
		return "Object"
	case ua.NodeClassVariable:
		return "Variable"
	case ua.NodeClassMethod:
		return "Method"
	case ua.NodeClassObjectType:
		return "ObjectType"
	case ua.NodeClassVariableType:
		return "VariableType"
	case ua.NodeClassReferenceType:
		return "ReferenceType"
	case ua.NodeClassDataType:
		return "DataType"
	case ua.NodeClassView:
		return "View"
	default:
		return "Unknown"
	}
}

// GetEndpoints returns available endpoints from the OPC-UA server
func GetEndpoints(ctx context.Context, endpoint string) ([]EndpointInfo, error) {
	endpoints, err := opcua.GetEndpoints(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get endpoints: %w", err)
	}

	result := make([]EndpointInfo, len(endpoints))
	for i, ep := range endpoints {
		tokens := make([]string, len(ep.UserIdentityTokens))
		for j, token := range ep.UserIdentityTokens {
			tokens[j] = tokenTypeToString(token.TokenType)
		}

		result[i] = EndpointInfo{
			EndpointURL:        ep.EndpointURL,
			SecurityMode:       securityModeToString(ep.SecurityMode),
			SecurityPolicy:     ep.SecurityPolicyURI,
			UserIdentityTokens: tokens,
		}
	}

	return result, nil
}

// EndpointInfo represents information about an OPC-UA endpoint
type EndpointInfo struct {
	EndpointURL        string   `json:"endpointUrl"`
	SecurityMode       string   `json:"securityMode"`
	SecurityPolicy     string   `json:"securityPolicy"`
	UserIdentityTokens []string `json:"userIdentityTokens"`
}

// securityModeToString converts SecurityMode to string
func securityModeToString(mode ua.MessageSecurityMode) string {
	switch mode {
	case ua.MessageSecurityModeNone:
		return "None"
	case ua.MessageSecurityModeSign:
		return "Sign"
	case ua.MessageSecurityModeSignAndEncrypt:
		return "SignAndEncrypt"
	default:
		return "Unknown"
	}
}

// tokenTypeToString converts UserTokenType to string
func tokenTypeToString(tokenType ua.UserTokenType) string {
	switch tokenType {
	case ua.UserTokenTypeAnonymous:
		return "Anonymous"
	case ua.UserTokenTypeUserName:
		return "Username"
	case ua.UserTokenTypeCertificate:
		return "Certificate"
	case ua.UserTokenTypeIssuedToken:
		return "IssuedToken"
	default:
		return "Unknown"
	}
}

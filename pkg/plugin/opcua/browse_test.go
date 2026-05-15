package opcua

import (
	"testing"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/assert"

	"github.com/monyskow/monyskow-simpleopcua-datasource/pkg/plugin/models"
)

// ---------------------------------------------------------------------------
// nodeClassToString
// ---------------------------------------------------------------------------

func TestNodeClassToString_KnownClasses(t *testing.T) {
	cases := []struct {
		nc       ua.NodeClass
		expected string
	}{
		{ua.NodeClassObject, "Object"},
		{ua.NodeClassVariable, "Variable"},
		{ua.NodeClassMethod, "Method"},
		{ua.NodeClassObjectType, "ObjectType"},
		{ua.NodeClassVariableType, "VariableType"},
		{ua.NodeClassReferenceType, "ReferenceType"},
		{ua.NodeClassDataType, "DataType"},
		{ua.NodeClassView, "View"},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, nodeClassToString(tc.nc), "class %v", tc.nc)
	}
}

func TestNodeClassToString_UnknownClass_ReturnsUnknown(t *testing.T) {
	assert.Equal(t, "Unknown", nodeClassToString(ua.NodeClass(9999)))
}

// ---------------------------------------------------------------------------
// securityModeToString / tokenTypeToString
// ---------------------------------------------------------------------------

func TestSecurityModeToString(t *testing.T) {
	cases := []struct {
		mode     ua.MessageSecurityMode
		expected string
	}{
		{ua.MessageSecurityModeNone, "None"},
		{ua.MessageSecurityModeSign, "Sign"},
		{ua.MessageSecurityModeSignAndEncrypt, "SignAndEncrypt"},
		{ua.MessageSecurityMode(99), "Unknown"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, securityModeToString(tc.mode))
	}
}

func TestTokenTypeToString(t *testing.T) {
	cases := []struct {
		tt       ua.UserTokenType
		expected string
	}{
		{ua.UserTokenTypeAnonymous, "Anonymous"},
		{ua.UserTokenTypeUserName, "Username"},
		{ua.UserTokenTypeCertificate, "Certificate"},
		{ua.UserTokenTypeIssuedToken, "IssuedToken"},
		{ua.UserTokenType(99), "Unknown"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.expected, tokenTypeToString(tc.tt))
	}
}

// ---------------------------------------------------------------------------
// extractPolicyName
// ---------------------------------------------------------------------------

func TestExtractPolicyName_FullURI_ReturnsShortName(t *testing.T) {
	uri := "http://opcfoundation.org/UA/SecurityPolicy#Basic256Sha256"
	assert.Equal(t, "Basic256Sha256", extractPolicyName(uri))
}

func TestExtractPolicyName_NoPrefix_ReturnsURIAsIs(t *testing.T) {
	uri := "SomeOtherPolicy"
	assert.Equal(t, uri, extractPolicyName(uri))
}

func TestExtractPolicyName_EmptyString_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", extractPolicyName(""))
}

// ---------------------------------------------------------------------------
// endpointSupportsTokenType
// ---------------------------------------------------------------------------

func makeEndpointWithTokens(tokenTypes ...ua.UserTokenType) *ua.EndpointDescription {
	ep := &ua.EndpointDescription{}
	for _, tt := range tokenTypes {
		ep.UserIdentityTokens = append(ep.UserIdentityTokens, &ua.UserTokenPolicy{TokenType: tt})
	}
	return ep
}

func TestEndpointSupportsTokenType_Supported(t *testing.T) {
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous, ua.UserTokenTypeUserName)

	assert.True(t, endpointSupportsTokenType(ep, ua.UserTokenTypeAnonymous))
	assert.True(t, endpointSupportsTokenType(ep, ua.UserTokenTypeUserName))
}

func TestEndpointSupportsTokenType_NotSupported(t *testing.T) {
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)

	assert.False(t, endpointSupportsTokenType(ep, ua.UserTokenTypeCertificate))
}

func TestEndpointSupportsTokenType_EmptyTokens_ReturnsFalse(t *testing.T) {
	ep := &ua.EndpointDescription{}
	assert.False(t, endpointSupportsTokenType(ep, ua.UserTokenTypeAnonymous))
}

// ---------------------------------------------------------------------------
// matchesSecuritySettings
// ---------------------------------------------------------------------------

func TestMatchesSecuritySettings_NoneNone_MatchesNoneModeEndpoint(t *testing.T) {
	settings := models.DataSourceSettings{
		SecurityPolicy: "None",
		SecurityMode:   "None",
		AuthMethod:     models.AuthAnonymous,
	}
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)
	ep.SecurityMode = ua.MessageSecurityModeNone

	assert.True(t, matchesSecuritySettings(ep, settings))
}

func TestMatchesSecuritySettings_NoneNone_NoMatchOnSecureEndpoint(t *testing.T) {
	settings := models.DataSourceSettings{
		SecurityPolicy: "None",
		SecurityMode:   "None",
		AuthMethod:     models.AuthAnonymous,
	}
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)
	ep.SecurityMode = ua.MessageSecurityModeSign

	assert.False(t, matchesSecuritySettings(ep, settings))
}

func TestMatchesSecuritySettings_SecurePolicy_MatchesCorrectEndpoint(t *testing.T) {
	settings := models.DataSourceSettings{
		SecurityPolicy: "Basic256Sha256",
		SecurityMode:   "SignAndEncrypt",
		AuthMethod:     models.AuthAnonymous,
	}
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)
	ep.SecurityPolicyURI = "http://opcfoundation.org/UA/SecurityPolicy#Basic256Sha256"
	ep.SecurityMode = ua.MessageSecurityModeSignAndEncrypt

	assert.True(t, matchesSecuritySettings(ep, settings))
}

func TestMatchesSecuritySettings_SecurePolicy_MismatchPolicyURI_NoMatch(t *testing.T) {
	settings := models.DataSourceSettings{
		SecurityPolicy: "Basic256Sha256",
		SecurityMode:   "SignAndEncrypt",
		AuthMethod:     models.AuthAnonymous,
	}
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)
	ep.SecurityPolicyURI = "http://opcfoundation.org/UA/SecurityPolicy#Basic128Rsa15"
	ep.SecurityMode = ua.MessageSecurityModeSignAndEncrypt

	assert.False(t, matchesSecuritySettings(ep, settings))
}

func TestMatchesSecuritySettings_TokenTypeMismatch_NoMatch(t *testing.T) {
	settings := models.DataSourceSettings{
		SecurityPolicy: "None",
		SecurityMode:   "None",
		AuthMethod:     models.AuthUserPass,
	}
	// Endpoint only offers anonymous, not username.
	ep := makeEndpointWithTokens(ua.UserTokenTypeAnonymous)
	ep.SecurityMode = ua.MessageSecurityModeNone

	assert.False(t, matchesSecuritySettings(ep, settings))
}

// ---------------------------------------------------------------------------
// referencesToNodes
// ---------------------------------------------------------------------------

func TestReferencesToNodes_EmptyRefs_ReturnsEmptySlice(t *testing.T) {
	nodes := referencesToNodes(nil)
	assert.Empty(t, nodes)
}

func TestReferencesToNodes_SingleRef_ConvertsCorrectly(t *testing.T) {
	ref := &ua.ReferenceDescription{
		NodeClass:   ua.NodeClassVariable,
		DisplayName: &ua.LocalizedText{Text: "Pressure"},
		BrowseName:  &ua.QualifiedName{Name: "Pressure"},
		NodeID:      &ua.ExpandedNodeID{NodeID: ua.NewNumericNodeID(1, 1001)},
	}

	nodes := referencesToNodes([]*ua.ReferenceDescription{ref})

	assert.Len(t, nodes, 1)
	assert.Equal(t, "Variable", nodes[0].NodeClass)
	assert.Equal(t, "Pressure", nodes[0].DisplayName)
	assert.Equal(t, "Pressure", nodes[0].BrowseName)
	assert.False(t, nodes[0].HasChildren, "Variable nodes must not be marked as having children")
}

func TestReferencesToNodes_ObjectNode_HasChildrenTrue(t *testing.T) {
	ref := &ua.ReferenceDescription{
		NodeClass:   ua.NodeClassObject,
		DisplayName: &ua.LocalizedText{Text: "PLC1"},
		BrowseName:  &ua.QualifiedName{Name: "PLC1"},
		NodeID:      &ua.ExpandedNodeID{NodeID: ua.NewNumericNodeID(1, 2000)},
	}

	nodes := referencesToNodes([]*ua.ReferenceDescription{ref})

	assert.True(t, nodes[0].HasChildren, "Object nodes should be flagged as potentially having children")
}

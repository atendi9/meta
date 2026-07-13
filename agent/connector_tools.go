package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// ToolParameterBinding describes an agent-owned source for a tool parameter or
// body field, letting the platform fill the value instead of the model.
type ToolParameterBinding struct {
	// Kind is the binding source: "default" for a literal or "macro" for a
	// platform-provided value.
	Kind string `json:"kind"`
	// Value is the typed literal used when Kind is "default".
	Value string `json:"value,omitzero"`
	// Macro is the platform value used when Kind is "macro": one of
	// "WHATSAPP_PHONE_NUMBER", "WHATSAPP_IDENTITY_HASH" or
	// "WHATSAPP_CURRENT_STATUS_ID".
	Macro string `json:"macro,omitzero"`
}

// ToolParameterNode describes a single path, query or header parameter of a
// connector tool request.
type ToolParameterNode struct {
	// Type is the parameter type: "string", "integer", "number" or "boolean".
	Type string `json:"type"`
	// Description documents the parameter for the agent.
	Description string `json:"description,omitzero"`
	// Required reports whether the parameter must be provided.
	Required bool `json:"required,omitzero"`
	// Binding optionally sources the value from the platform instead of the agent.
	Binding *ToolParameterBinding `json:"binding,omitzero"`
}

// ToolBodyNode describes a single JSON body field of a connector tool request.
// Object nodes nest children in Properties and array nodes describe their
// element schema in Items.
type ToolBodyNode struct {
	// Type is the field type: "object", "array", "string", "integer",
	// "number" or "boolean".
	Type string `json:"type"`
	// Description documents the field for the agent.
	Description string `json:"description,omitzero"`
	// Required lists the required property names of an object node.
	Required []string `json:"required,omitzero"`
	// Properties holds the child fields of an object node.
	Properties map[string]ToolBodyNode `json:"properties,omitzero"`
	// Items describes the element schema of an array node.
	Items *ToolBodyNode `json:"items,omitzero"`
	// Binding optionally sources the value from the platform instead of the agent.
	Binding *ToolParameterBinding `json:"binding,omitzero"`
}

// ToolRequestBodyDefinition describes the JSON body of a connector tool request.
type ToolRequestBodyDefinition struct {
	// ContentType is the body content type; currently "application/json".
	ContentType string `json:"content_type"`
	// Params holds the top-level body fields keyed by name.
	Params map[string]ToolBodyNode `json:"params"`
	// Required lists the required top-level body field names.
	Required []string `json:"required,omitzero"`
}

// ToolRequestDefinition describes the outbound HTTP request a connector tool
// performs against the external service.
type ToolRequestDefinition struct {
	// Method is the HTTP method: "GET", "POST", "PUT", "DELETE" or "PATCH".
	Method string `json:"method"`
	// Path is the outbound request path template, including {placeholder} segments.
	Path string `json:"path"`
	// PathParameters holds the path-parameter schema keyed by placeholder name.
	PathParameters map[string]ToolParameterNode `json:"path_parameters,omitzero"`
	// QueryParameters holds the query-parameter schema keyed by outbound name.
	QueryParameters map[string]ToolParameterNode `json:"query_parameters,omitzero"`
	// Headers holds the header schema keyed by outbound header name.
	Headers map[string]ToolParameterNode `json:"headers,omitzero"`
	// Body optionally describes the request body.
	Body *ToolRequestBodyDefinition `json:"body,omitzero"`
}

// ToolUserAuthActionConfig configures how an authentication tool extracts token
// material from the external service's response.
type ToolUserAuthActionConfig struct {
	// UserActionToolType is the action type: "auth" or "refresh".
	UserActionToolType string `json:"user_action_tool_type"`
	// UserAuthTokenPath is the dot-path to the access token in the response.
	UserAuthTokenPath string `json:"user_auth_token_path"`
	// RefreshTokenPath is the dot-path to the refresh token in the response.
	RefreshTokenPath string `json:"refresh_token_path,omitzero"`
	// ExpiresAtPath is the dot-path to the token expiry in the response.
	ExpiresAtPath string `json:"expires_at_path,omitzero"`
	// ExpiresAtType is the expiry format: "absolute" or "relative_seconds".
	ExpiresAtType string `json:"expires_at_type,omitzero"`
}

// ConnectorToolRequest represents the body used to create or update a connector tool.
type ConnectorToolRequest struct {
	// Name is a stable tool key visible to the agent (for example "check_order_status").
	Name string `json:"name"`
	// Description tells the agent when to use the tool; the agent relies on this.
	Description string `json:"description"`
	// RequestDefinition specifies the outbound HTTP request.
	RequestDefinition ToolRequestDefinition `json:"request_definition"`
	// UserAuthRequired reports whether stored user auth is injected into the request.
	UserAuthRequired bool `json:"user_auth_required"`
	// UserAuthActionConfig configures auth-token extraction for auth tools.
	UserAuthActionConfig *ToolUserAuthActionConfig `json:"user_auth_action_config,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *ConnectorToolRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// ConnectorToolResponse represents a connector tool object.
type ConnectorToolResponse struct {
	// ID is the unique identifier for this tool.
	ID string `json:"id"`
	// Name is the stable tool key visible to the agent.
	Name string `json:"name"`
	// Description tells the agent when to use the tool.
	Description string `json:"description"`
	// RequestDefinition specifies the outbound HTTP request.
	RequestDefinition ToolRequestDefinition `json:"request_definition"`
	// UserAuthRequired reports whether stored user auth is injected into the request.
	UserAuthRequired bool `json:"user_auth_required"`
	// UserAuthActionConfig configures auth-token extraction for auth tools.
	UserAuthActionConfig *ToolUserAuthActionConfig `json:"user_auth_action_config,omitzero"`
}

// ToolRunRequest represents the body used to execute a connector tool.
type ToolRunRequest struct {
	// Input is the JSON-encoded input payload; defaults to an empty object.
	Input string `json:"input,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
func (r *ToolRunRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// ToolRunResponse represents the result of executing a connector tool.
type ToolRunResponse struct {
	// Output is the JSON-encoded response from the tool execution.
	Output string `json:"output"`
	// Status is the execution status: "success" or "error".
	Status string `json:"status"`
}

// ConnectorTools provides CRUD and execution access to the tools of a connector.
type ConnectorTools struct {
	o           *Onboard
	connectorID string
}

// Tools exposes the tool management operations for the given connector.
func (c *Connectors) Tools(connectorId string) *ConnectorTools {
	return &ConnectorTools{o: c.o, connectorID: connectorId}
}

// path builds a tools endpoint URL for the parent connector.
func (t *ConnectorTools) path(suffix string) string {
	return t.o.Config.URL("/agent_connectors/" + t.connectorID + "/tools" + suffix)
}

// Create adds a new tool to the connector.
func (t *ConnectorTools) Create(tool ConnectorToolRequest) (ConnectorToolResponse, error) {
	url := t.path("")
	headers := t.o.Client.Headers("application/json")
	var result ConnectorToolResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &tool,
	}
	res, err := t.o.Client.Post(url, opts)
	if err != nil {
		return ConnectorToolResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorToolResponse{}, err
	}
	return result, nil
}

// List retrieves the tools for the connector.
func (t *ConnectorTools) List() ([]ConnectorToolResponse, error) {
	url := t.path("")
	headers := t.o.Client.Headers("application/json")
	var result struct {
		Root []ConnectorToolResponse `json:"root"`
	}
	res, err := t.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return []ConnectorToolResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []ConnectorToolResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single tool by its identifier.
func (t *ConnectorTools) Get(toolId string) (ConnectorToolResponse, error) {
	url := t.path("/" + toolId)
	headers := t.o.Client.Headers("application/json")
	var result ConnectorToolResponse
	res, err := t.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return ConnectorToolResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorToolResponse{}, err
	}
	return result, nil
}

// Update modifies an existing tool identified by toolId.
func (t *ConnectorTools) Update(toolId string, tool ConnectorToolRequest) (ConnectorToolResponse, error) {
	url := t.path("/" + toolId)
	headers := t.o.Client.Headers("application/json")
	var result ConnectorToolResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &tool,
	}
	res, err := t.o.Client.Put(url, opts)
	if err != nil {
		return ConnectorToolResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorToolResponse{}, err
	}
	return result, nil
}

// Delete removes a tool by its identifier.
func (t *ConnectorTools) Delete(toolId string) error {
	url := t.path("/" + toolId)
	headers := t.o.Client.Headers("application/json")
	_, err := t.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

// Run executes a tool with the provided input and returns its result.
func (t *ConnectorTools) Run(toolId string, req ToolRunRequest) (ToolRunResponse, error) {
	url := t.path("/" + toolId + "/run")
	headers := t.o.Client.Headers("application/json")
	var result ToolRunResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &req,
	}
	res, err := t.o.Client.Post(url, opts)
	if err != nil {
		return ToolRunResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ToolRunResponse{}, err
	}
	return result, nil
}

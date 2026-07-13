package agent

import (
	"bytes"
	"io"
	"strconv"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// ConnectorAuthType identifies how a connector authenticates against the
// external service it integrates with.
type ConnectorAuthType string

func (t ConnectorAuthType) String() string {
	switch t {
	case ConnectorAuthOAuth2ClientCredentials,
		ConnectorAuthAPIKey,
		ConnectorAuthNone:
		return string(t)
	default:
		return string(ConnectorAuthNone)
	}
}

const (
	// ConnectorAuthOAuth2ClientCredentials authenticates using the OAuth2
	// client credentials grant.
	ConnectorAuthOAuth2ClientCredentials ConnectorAuthType = "OAUTH2_CLIENT_CREDENTIALS"
	// ConnectorAuthAPIKey authenticates using a static API key injected into
	// headers, query parameters or the request body.
	ConnectorAuthAPIKey ConnectorAuthType = "API_KEY"
	// ConnectorAuthNone performs no authentication.
	ConnectorAuthNone ConnectorAuthType = "NONE"
)

// APIKeyParam describes a single API-key field and how it is injected into an
// outbound request.
type APIKeyParam struct {
	// FieldName is the name of the header, query parameter or body field.
	FieldName string `json:"field_name"`
	// Value is the API-key value to inject.
	Value string `json:"value"`
	// Prefix is prepended to Value (for example "Bearer ").
	Prefix string `json:"prefix,omitzero"`
}

// APIKeyConfig holds the API-key material for a connector, grouped by where it
// is injected in the outbound request.
type APIKeyConfig struct {
	// Headers holds API keys injected as request headers.
	Headers []APIKeyParam `json:"headers,omitzero"`
	// QueryParams holds API keys injected as query parameters.
	QueryParams []APIKeyParam `json:"query_params,omitzero"`
	// BodyParams holds API keys injected as JSON body fields.
	BodyParams []APIKeyParam `json:"body_params,omitzero"`
}

// OAuth2ClientCredentialsConfig holds the OAuth2 client-credentials settings
// used to obtain access tokens for a connector.
type OAuth2ClientCredentialsConfig struct {
	// TokenURL is the OAuth token endpoint.
	TokenURL string `json:"token_url"`
	// ScopesToRequest lists the OAuth scopes requested with each token.
	ScopesToRequest []string `json:"scopes_to_request"`
	// TokenRequestContentType is the content type used for the token request,
	// either "application/x-www-form-urlencoded" or "application/json".
	TokenRequestContentType string `json:"token_request_content_type,omitzero"`
	// ClientID is the OAuth client identifier.
	ClientID string `json:"client_id"`
	// ClientSecret is the OAuth client secret.
	ClientSecret string `json:"client_secret"`
}

// ConnectorAuthConfig holds the authentication configuration of a connector.
// Exactly one field is populated, matching the connector's auth type.
type ConnectorAuthConfig struct {
	// OAuth2ClientCredentials holds the OAuth2 client-credentials settings.
	OAuth2ClientCredentials *OAuth2ClientCredentialsConfig `json:"oauth2_client_credentials,omitzero"`
	// APIKey holds the API-key settings.
	APIKey *APIKeyConfig `json:"api_key,omitzero"`
}

// UserAuthInjectionConfig describes how a stored per-user token is injected
// into outbound requests.
type UserAuthInjectionConfig struct {
	// Location is the placement of the token: "body", "headers", "path" or "query".
	Location string `json:"location"`
	// FieldName identifies the token field (for example "X-User-Token").
	FieldName string `json:"field_name"`
	// Prefix is prepended to the token value (for example "Bearer ").
	Prefix string `json:"prefix"`
}

// MTLSConfig describes the mutual-TLS certificate state of a connector.
type MTLSConfig struct {
	// HasCertificate reports whether a client certificate is configured.
	HasCertificate bool `json:"has_certificate"`
	// Fingerprint is the SHA-256 hash of the certificate.
	Fingerprint string `json:"fingerprint,omitzero"`
	// ExpiresAt is the certificate expiry as a Unix timestamp.
	ExpiresAt int64 `json:"expires_at,omitzero"`
	// Subject is the certificate subject distinguished name.
	Subject string `json:"subject,omitzero"`
	// ClientCertificate is the PEM-encoded public client certificate.
	ClientCertificate string `json:"client_certificate,omitzero"`
	// CACertificate is the PEM-encoded CA certificate chain.
	CACertificate string `json:"ca_certificate,omitzero"`
}

// ConnectionStatus reports the current connection state of a connector.
type ConnectionStatus struct {
	// Status is the connection state: "PENDING_OAUTH", "ACTIVE", "EXPIRED" or "ERROR".
	Status string `json:"status"`
	// ErrorMessage explains the failure when Status is "ERROR".
	ErrorMessage string `json:"error_message,omitzero"`
}

// ConnectorRequest represents the body used to create or update a connector.
type ConnectorRequest struct {
	// Name is a display name identifying the external service.
	Name string `json:"name"`
	// Description explains the connector's integration purpose and capabilities.
	Description string `json:"description"`
	// BaseURL is the external API base URL.
	BaseURL string `json:"base_url"`
	// AuthType selects the authentication scheme for the connector.
	AuthType ConnectorAuthType `json:"auth_type"`
	// AuthConfig holds the authentication configuration matching AuthType.
	AuthConfig *ConnectorAuthConfig `json:"auth_config,omitzero"`
	// UserAuthInjectionConfig configures per-user token injection.
	UserAuthInjectionConfig *UserAuthInjectionConfig `json:"user_auth_injection_config,omitzero"`
	// RequiresCertificate reports whether mutual TLS is required.
	RequiresCertificate bool `json:"requires_certificate,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *ConnectorRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// ConnectorResponse represents a connector object.
type ConnectorResponse struct {
	// ID is the unique identifier for this connector.
	ID string `json:"id"`
	// Name is the display name identifying the external service.
	Name string `json:"name"`
	// Description explains the connector's integration purpose and capabilities.
	Description string `json:"description,omitzero"`
	// BaseURL is the external API base URL.
	BaseURL string `json:"base_url,omitzero"`
	// AuthType is the authentication scheme configured for the connector.
	AuthType ConnectorAuthType `json:"auth_type,omitzero"`
	// AuthConfig holds the non-secret authentication configuration.
	AuthConfig *ConnectorAuthConfig `json:"auth_config,omitzero"`
	// MTLSConfig reports the mutual-TLS certificate state.
	MTLSConfig *MTLSConfig `json:"mtls_config,omitzero"`
	// ConnectionStatus reports the current connection state.
	ConnectionStatus *ConnectionStatus `json:"connection_status,omitzero"`
}

// UpsertAPIKeyRequest represents the body used to set or replace the API-key
// credentials of a connector.
type UpsertAPIKeyRequest struct {
	// APIKeyConfig holds the API-key material to store.
	APIKeyConfig APIKeyConfig `json:"api_key_config"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
func (r *UpsertAPIKeyRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// UpsertOAuthRequest represents the body used to set or replace the OAuth2
// client-credentials of a connector.
type UpsertOAuthRequest struct {
	// OAuthConfig holds the OAuth2 client-credentials to store.
	OAuthConfig OAuth2ClientCredentialsConfig `json:"oauth_config"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
func (r *UpsertOAuthRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// CertificateRequest represents the body used to set or replace the mutual-TLS
// certificate of a connector.
type CertificateRequest struct {
	// ClientCertificate is the PEM-encoded client certificate.
	ClientCertificate string `json:"client_certificate"`
	// ClientKey is the PEM-encoded private key (PKCS8, RSA or EC).
	ClientKey string `json:"client_key"`
	// CACertificate is the optional PEM-encoded CA certificate.
	CACertificate string `json:"ca_certificate,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
func (r *CertificateRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// ConnectorLogEntry represents a single connector error-log entry or an
// aggregated failure pattern, depending on the query mode.
type ConnectorLogEntry struct {
	// EventTime is the ISO 8601 UTC timestamp of the log entry (individual mode).
	EventTime string `json:"event_time,omitzero"`
	// FailureCodeName is a human-readable failure identifier.
	FailureCodeName string `json:"failure_code_name,omitzero"`
	// ErrorMessage describes the operation failure.
	ErrorMessage string `json:"error_message,omitzero"`
	// ToolName is the tool executed when the error occurred.
	ToolName string `json:"tool_name,omitzero"`
	// Occurrences is the failure-pattern frequency (summary mode).
	Occurrences int64 `json:"occurrences,omitzero"`
	// LastSeen is the most recent occurrence timestamp (summary mode).
	LastSeen string `json:"last_seen,omitzero"`
}

// ConnectorLogStats holds aggregate statistics over a connector log window.
type ConnectorLogStats struct {
	// StartCount is the number of started operations.
	StartCount int64 `json:"start_count"`
	// SuccessCount is the number of successful operations.
	SuccessCount int64 `json:"success_count"`
	// ExceptionCount is the number of failed operations.
	ExceptionCount int64 `json:"exception_count"`
	// SuccessRate is the fraction of successful operations.
	SuccessRate float64 `json:"success_rate"`
	// AvgLatencyS is the average latency in seconds.
	AvgLatencyS float64 `json:"avg_latency_s"`
	// P95LatencyS is the 95th-percentile latency in seconds.
	P95LatencyS float64 `json:"p95_latency_s"`
	// P99LatencyS is the 99th-percentile latency in seconds.
	P99LatencyS float64 `json:"p99_latency_s"`
	// TimeWindowSeconds is the span of the statistics window in seconds.
	TimeWindowSeconds int64 `json:"time_window_seconds"`
}

// ConnectorLogStatsResponse holds the result of a connector logs query.
type ConnectorLogStatsResponse struct {
	// Data holds the individual log entries or aggregated failure patterns.
	Data []ConnectorLogEntry `json:"data"`
	// Stats holds the optional aggregate statistics.
	Stats *ConnectorLogStats `json:"stats,omitzero"`
}

// ConnectorLogsOptions holds the optional filters for a connector logs query.
// The zero value requests the default window of the last 24 hours.
type ConnectorLogsOptions struct {
	// StartTime is the Unix timestamp lower bound; defaults to 24 hours ago.
	StartTime int64
	// EndTime is the Unix timestamp upper bound; defaults to now.
	EndTime int64
	// Limit caps the number of returned entries (1-1000); defaults to 100.
	Limit int64
	// ToolID filters logs to a specific operation.
	ToolID string
	// IncludeStats requests aggregate statistics in the response.
	IncludeStats bool
	// SummaryOnly returns aggregated failure patterns instead of entries.
	SummaryOnly bool
	// TopN caps the number of failure patterns (1-50); defaults to 10.
	TopN int64
}

func (o ConnectorLogsOptions) queryParams() []xhttp.HTTPData {
	var params []xhttp.HTTPData
	if o.StartTime != 0 {
		params = append(params, &xhttp.Data{Key: "start_time", Value: strconv.FormatInt(o.StartTime, 10)})
	}
	if o.EndTime != 0 {
		params = append(params, &xhttp.Data{Key: "end_time", Value: strconv.FormatInt(o.EndTime, 10)})
	}
	if o.Limit != 0 {
		params = append(params, &xhttp.Data{Key: "limit", Value: strconv.FormatInt(o.Limit, 10)})
	}
	if o.ToolID != "" {
		params = append(params, &xhttp.Data{Key: "tool_id", Value: o.ToolID})
	}
	if o.IncludeStats {
		params = append(params, &xhttp.Data{Key: "include_stats", Value: "true"})
	}
	if o.SummaryOnly {
		params = append(params, &xhttp.Data{Key: "summary_only", Value: "true"})
	}
	if o.TopN != 0 {
		params = append(params, &xhttp.Data{Key: "top_n", Value: strconv.FormatInt(o.TopN, 10)})
	}
	return params
}

// Connectors provides CRUD and credential-management access to the omni-channel
// connectors of an entity.
type Connectors struct {
	o *Onboard
}

// Connectors exposes the connector management operations for the configured entity.
func (c *Configurator) Connectors() *Connectors {
	return &Connectors{o: c.Client.Onboard}
}

// Create adds a new connector to the specified entity.
func (c *Connectors) Create(connector ConnectorRequest) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors")
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &connector,
	}
	res, err := c.o.Client.Post(url, opts)
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// List retrieves the connectors for the specified entity.
func (c *Connectors) List() ([]ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors")
	headers := c.o.Client.Headers("application/json")
	var result struct {
		Root []ConnectorResponse `json:"root"`
	}
	res, err := c.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return []ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []ConnectorResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single connector by its identifier.
func (c *Connectors) Get(connectorId string) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId)
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	res, err := c.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// Update modifies an existing connector identified by connectorId.
func (c *Connectors) Update(connectorId string, connector ConnectorRequest) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId)
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &connector,
	}
	res, err := c.o.Client.Put(url, opts)
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// Delete removes a connector by its identifier.
func (c *Connectors) Delete(connectorId string) error {
	url := c.o.Config.URL("/agent_connectors/" + connectorId)
	headers := c.o.Client.Headers("application/json")
	_, err := c.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

// UpsertAPIKey sets or replaces the API-key credentials of a connector.
func (c *Connectors) UpsertAPIKey(connectorId string, req UpsertAPIKeyRequest) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId + "/upsertApiKey")
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &req,
	}
	res, err := c.o.Client.Post(url, opts)
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// UpsertOAuth sets or replaces the OAuth2 client-credentials of a connector.
func (c *Connectors) UpsertOAuth(connectorId string, req UpsertOAuthRequest) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId + "/upsertOAuth")
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &req,
	}
	res, err := c.o.Client.Post(url, opts)
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// UpsertCertificate sets or replaces the mutual-TLS certificate of a connector.
func (c *Connectors) UpsertCertificate(connectorId string, req CertificateRequest) (ConnectorResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId + "/upsertCertificate")
	headers := c.o.Client.Headers("application/json")
	var result ConnectorResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &req,
	}
	res, err := c.o.Client.Post(url, opts)
	if err != nil {
		return ConnectorResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorResponse{}, err
	}
	return result, nil
}

// Logs retrieves the error logs and optional statistics for a connector.
// Passing the zero-value options requests the default window of the last 24 hours.
func (c *Connectors) Logs(connectorId string, opts ConnectorLogsOptions) (ConnectorLogStatsResponse, error) {
	url := c.o.Config.URL("/agent_connectors/" + connectorId + "/logs")
	headers := c.o.Client.Headers("application/json")
	var result ConnectorLogStatsResponse
	res, err := c.o.Client.Get(url, &xhttp.Options{
		Headers:     headers,
		QueryParams: opts.queryParams(),
	})
	if err != nil {
		return ConnectorLogStatsResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ConnectorLogStatsResponse{}, err
	}
	return result, nil
}

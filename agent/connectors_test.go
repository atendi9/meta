package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestConnectorAuthType_String(t *testing.T) {
	assert.Equal(t, "OAUTH2_CLIENT_CREDENTIALS", ConnectorAuthOAuth2ClientCredentials.String())
	assert.Equal(t, "API_KEY", ConnectorAuthAPIKey.String())
	assert.Equal(t, "NONE", ConnectorAuthNone.String())
	// Unknown types default to NONE.
	assert.Equal(t, "NONE", ConnectorAuthType("bogus").String())
}

func TestConnectorLogsOptions_QueryParams(t *testing.T) {
	// The zero value produces no query parameters (default window).
	assert.LengthSlice(t, 0, ConnectorLogsOptions{}.queryParams())

	opts := ConnectorLogsOptions{
		StartTime:    100,
		EndTime:      200,
		Limit:        50,
		ToolID:       "tool_1",
		IncludeStats: true,
		SummaryOnly:  true,
		TopN:         5,
	}
	params := opts.queryParams()
	got := map[string]string{}
	for _, p := range params {
		got[p.Data().Key] = p.Data().Value.(string)
	}
	assert.Equal(t, "100", got["start_time"])
	assert.Equal(t, "200", got["end_time"])
	assert.Equal(t, "50", got["limit"])
	assert.Equal(t, "tool_1", got["tool_id"])
	assert.Equal(t, "true", got["include_stats"])
	assert.Equal(t, "true", got["summary_only"])
	assert.Equal(t, "5", got["top_n"])
}

func TestConnectors_Create_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1","name":"CRM","auth_type":"API_KEY"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.Create(ConnectorRequest{
		Name:     "CRM",
		BaseURL:  "https://crm.example.com",
		AuthType: ConnectorAuthAPIKey,
	})

	assert.NoError(t, err)
	assert.Equal(t, "c1", conn.ID)
	assert.Equal(t, "CRM", conn.Name)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors", mock.Calls[0].URL)
}

func TestConnectors_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"c1","name":"CRM"}]}`)
	c := newTestConfigurator(mock).Connectors()

	conns, err := c.List()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, conns)
	assert.Equal(t, "c1", conns[0].ID)
}

func TestConnectors_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1","name":"CRM"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.Get("c1")

	assert.NoError(t, err)
	assert.Equal(t, "c1", conn.ID)
	assert.Equal(t, testBase+"/agent_connectors/c1", mock.Calls[0].URL)
}

func TestConnectors_Update_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1","name":"CRM2"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.Update("c1", ConnectorRequest{Name: "CRM2"})

	assert.NoError(t, err)
	assert.Equal(t, "CRM2", conn.Name)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1", mock.Calls[0].URL)
}

func TestConnectors_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	c := newTestConfigurator(mock).Connectors()

	err := c.Delete("c1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1", mock.Calls[0].URL)
}

func TestConnectors_UpsertAPIKey_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1","auth_type":"API_KEY"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.UpsertAPIKey("c1", UpsertAPIKeyRequest{
		APIKeyConfig: APIKeyConfig{
			Headers: []APIKeyParam{{FieldName: "X-Key", Value: "secret", Prefix: "Bearer "}},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "c1", conn.ID)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/upsertApiKey", mock.Calls[0].URL)
}

func TestConnectors_UpsertOAuth_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.UpsertOAuth("c1", UpsertOAuthRequest{
		OAuthConfig: OAuth2ClientCredentialsConfig{
			TokenURL:     "https://auth.example.com/token",
			ClientID:     "id",
			ClientSecret: "secret",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "c1", conn.ID)
	assert.Equal(t, testBase+"/agent_connectors/c1/upsertOAuth", mock.Calls[0].URL)
}

func TestConnectors_UpsertCertificate_Success(t *testing.T) {
	mock := mockOK(`{"id":"c1"}`)
	c := newTestConfigurator(mock).Connectors()

	conn, err := c.UpsertCertificate("c1", CertificateRequest{
		ClientCertificate: "-----BEGIN CERTIFICATE-----",
		ClientKey:         "-----BEGIN PRIVATE KEY-----",
	})

	assert.NoError(t, err)
	assert.Equal(t, "c1", conn.ID)
	assert.Equal(t, testBase+"/agent_connectors/c1/upsertCertificate", mock.Calls[0].URL)
}

func TestConnectors_Logs_Success(t *testing.T) {
	body := `{"data":[{"failure_code_name":"TIMEOUT","error_message":"boom","tool_name":"lookup"}],"stats":{"start_count":10,"success_count":8,"exception_count":2,"success_rate":0.8}}`
	mock := mockOK(body)
	c := newTestConfigurator(mock).Connectors()

	res, err := c.Logs("c1", ConnectorLogsOptions{IncludeStats: true})

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, res.Data)
	assert.Equal(t, "TIMEOUT", res.Data[0].FailureCodeName)
	assert.NotNil(t, res.Stats)
	assert.Equal(t, int64(10), res.Stats.StartCount)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/logs", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "include_stats", params[0].Data().Key)
}

func TestConnectors_TransportErrors(t *testing.T) {
	c := newTestConfigurator(mockErr(io.ErrClosedPipe)).Connectors()

	_, err := c.Create(ConnectorRequest{})
	assert.Error(t, err)
	_, err = c.List()
	assert.Error(t, err)
	_, err = c.Get("c1")
	assert.Error(t, err)
	_, err = c.Update("c1", ConnectorRequest{})
	assert.Error(t, err)
	err = c.Delete("c1")
	assert.Error(t, err)
	_, err = c.UpsertAPIKey("c1", UpsertAPIKeyRequest{})
	assert.Error(t, err)
	_, err = c.UpsertOAuth("c1", UpsertOAuthRequest{})
	assert.Error(t, err)
	_, err = c.UpsertCertificate("c1", CertificateRequest{})
	assert.Error(t, err)
	_, err = c.Logs("c1", ConnectorLogsOptions{})
	assert.Error(t, err)
}

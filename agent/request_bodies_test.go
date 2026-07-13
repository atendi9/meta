package agent

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// readJSON reads a lazily-marshaled request body and unmarshals it into a
// generic map, asserting the payload is valid JSON.
func readJSON(t *testing.T, r io.Reader) map[string]any {
	t.Helper()
	raw, err := io.ReadAll(r)
	assert.NoError(t, err)
	var out map[string]any
	assert.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestRequestBodies_MarshalToJSON(t *testing.T) {
	t.Run("SettingsRequest", func(t *testing.T) {
		body := readJSON(t, &SettingsRequest{AIAudience: string(AIAudiencePublic)})
		assert.Equal(t, "EVERYONE", body["ai_audience"].(string))
	})

	t.Run("SkillRequest", func(t *testing.T) {
		body := readJSON(t, &SkillRequest{Title: "greet", Skill: "say hi"})
		assert.Equal(t, "greet", body["title"].(string))
		assert.Equal(t, "say hi", body["skill"].(string))
	})

	t.Run("FAQRequest", func(t *testing.T) {
		body := readJSON(t, &FAQRequest{Question: "Q?", Answer: "A."})
		assert.Equal(t, "Q?", body["question"].(string))
		assert.Equal(t, "A.", body["answer"].(string))
	})

	t.Run("WebsiteRequest", func(t *testing.T) {
		body := readJSON(t, &WebsiteRequest{URL: "https://x.com"})
		assert.Equal(t, "https://x.com", body["url"].(string))
	})

	t.Run("ConnectorRequest", func(t *testing.T) {
		body := readJSON(t, &ConnectorRequest{Name: "CRM", BaseURL: "https://crm", AuthType: ConnectorAuthAPIKey})
		assert.Equal(t, "CRM", body["name"].(string))
		assert.Equal(t, "API_KEY", body["auth_type"].(string))
	})

	t.Run("UpsertAPIKeyRequest", func(t *testing.T) {
		body := readJSON(t, &UpsertAPIKeyRequest{APIKeyConfig: APIKeyConfig{
			Headers: []APIKeyParam{{FieldName: "X-Key", Value: "secret"}},
		}})
		_, ok := body["api_key_config"]
		assert.True(t, ok)
	})

	t.Run("UpsertOAuthRequest", func(t *testing.T) {
		body := readJSON(t, &UpsertOAuthRequest{OAuthConfig: OAuth2ClientCredentialsConfig{
			TokenURL: "https://auth", ClientID: "id", ClientSecret: "secret",
		}})
		_, ok := body["oauth_config"]
		assert.True(t, ok)
	})

	t.Run("CertificateRequest", func(t *testing.T) {
		body := readJSON(t, &CertificateRequest{ClientCertificate: "cert", ClientKey: "key"})
		assert.Equal(t, "cert", body["client_certificate"].(string))
		assert.Equal(t, "key", body["client_key"].(string))
	})

	t.Run("ConnectorToolRequest", func(t *testing.T) {
		body := readJSON(t, &ConnectorToolRequest{Name: "check_order", Description: "d"})
		assert.Equal(t, "check_order", body["name"].(string))
	})

	t.Run("ToolRunRequest", func(t *testing.T) {
		body := readJSON(t, &ToolRunRequest{Input: `{"id":"A1"}`})
		assert.Equal(t, `{"id":"A1"}`, body["input"].(string))
	})

	t.Run("BusinessInfoRequest", func(t *testing.T) {
		body := readJSON(t, &BusinessInfoRequest{BusinessDescription: "shoes"})
		assert.Equal(t, "shoes", body["business_description"].(string))
	})

	t.Run("AgentEventRequest", func(t *testing.T) {
		body := readJSON(t, &AgentEventRequest{To: "551199", Event: AgentEventPayload{Type: "x", Description: "d", Payload: "{}"}})
		assert.Equal(t, "551199", body["to"].(string))
	})

	t.Run("AgentTestRequest", func(t *testing.T) {
		body := readJSON(t, &AgentTestRequest{UserMsg: "hi"})
		assert.Equal(t, "hi", body["user_msg"].(string))
	})

	t.Run("ThreadControlRequest", func(t *testing.T) {
		body := readJSON(t, &ThreadControlRequest{MessagingProduct: "whatsapp", Action: "release", To: "551199"})
		assert.Equal(t, "release", body["action"].(string))
	})
}

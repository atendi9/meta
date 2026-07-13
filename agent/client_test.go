package agent

import (
	"context"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNew(t *testing.T) {
	c := New(context.Background(), "entity_1", "token_1")

	assert.Equal(t, "entity_1", c.entityId)
	assert.Equal(t, APIVersion, c.ApiVersion)
	assert.Equal(t, BaseURL, c.BaseUrl)
	assert.Equal(t, "token_1", c.AccessToken)
	assert.NotNil(t, c.Onboard)
	assert.NotNil(t, c.AllowList)

	// The onboarding config mirrors the client configuration.
	assert.Equal(t, "entity_1", c.Onboard.Config.EntityID)
	assert.Equal(t, BaseURL, c.Onboard.Config.BaseURL)
	assert.Equal(t, APIVersion, c.Onboard.Config.ApiVersion)

	// AllowList and Onboard share the same client pointer.
	assert.True(t, c.AllowList.o == c.Onboard)
	assert.True(t, c.Onboard.Client == c)
}

func TestConfigure(t *testing.T) {
	c := New(context.Background(), "entity_1", "token_1")

	cfg := Configure(c)

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Client == c)
}

func TestClient_Headers(t *testing.T) {
	c := New(context.Background(), "entity_1", "token_1")

	headers := c.Headers("application/json")

	assert.LengthSlice(t, 3, headers)
	assert.Equal(t, "Content-Type", headers[0].Data().Key)
	assert.Equal(t, "application/json", headers[0].Data().Value.(string))
	assert.Equal(t, "Authorization", headers[1].Data().Key)
	assert.Equal(t, "Bearer token_1", headers[1].Data().Value.(string))
	assert.Equal(t, "X-API-Version", headers[2].Data().Key)
	assert.Equal(t, APIVersion, headers[2].Data().Value.(string))
}

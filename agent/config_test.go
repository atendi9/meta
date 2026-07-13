package agent

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestConfig_URL(t *testing.T) {
	c := Config{
		EntityID:   "123",
		BaseURL:    "https://api.facebook.com",
		ApiVersion: "2.0.0",
	}

	assert.Equal(t, "https://api.facebook.com/123/agent_eligibility", c.URL("/agent_eligibility"))
	assert.Equal(t, "https://api.facebook.com/123", c.URL(""))
	assert.Equal(t, "https://api.facebook.com/123/agent_config/skills/abc", c.URL("/agent_config/skills/abc"))
}

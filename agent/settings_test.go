package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestOnboard_Settings_Success(t *testing.T) {
	body := `{"root":[{"agent_id":"a1","channel":"whatsapp","rollout":{"enabled":true},"ai_audience":"EVERYONE"}]}`
	mock := mockOK(body)
	c := newTestClient(mock)

	settings, err := c.Onboard.Settings()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, settings)
	assert.Equal(t, "a1", settings[0].AgentID)
	assert.True(t, settings[0].Rollout.Enabled)
	assert.Equal(t, "EVERYONE", settings[0].AIAudience)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/settings", mock.Calls[0].URL)
	assert.LengthSlice(t, 0, mock.Calls[0].Options.Q())
}

func TestOnboard_Settings_WithAgentID(t *testing.T) {
	mock := mockOK(`{"root":[]}`)
	c := newTestClient(mock)

	_, err := c.Onboard.Settings("agent_42")

	assert.NoError(t, err)
	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_42", params[0].Data().Value.(string))
}

func TestOnboard_Settings_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrUnexpectedEOF))

	settings, err := c.Onboard.Settings()

	assert.Error(t, err)
	assert.LengthSlice(t, 0, settings)
}

func TestOnboard_UpdateSettings_Success(t *testing.T) {
	mock := mockOK(`{"agent_id":"a1","channel":"whatsapp","rollout":{"enabled":false}}`)
	c := newTestClient(mock)

	req := SettingsRequest{
		Rollout:    SettingsRollout{Enabled: false},
		AIAudience: string(AIAudiencePrivate),
	}
	res, err := c.Onboard.UpdateSettings(req, "agent_42")

	assert.NoError(t, err)
	assert.Equal(t, "a1", res.AgentID)
	assert.False(t, res.Rollout.Enabled)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/settings", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_42", params[0].Data().Value.(string))
}

func TestOnboard_UpdateSettings_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrClosedPipe))

	_, err := c.Onboard.UpdateSettings(SettingsRequest{})

	assert.Error(t, err)
}

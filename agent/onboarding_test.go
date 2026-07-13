package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestOnboardingChannel_String(t *testing.T) {
	cases := map[OnboardingChannel]string{
		OnboardingEmailChannel:     "email",
		OnboardingWhatsappChannel:  "whatsapp",
		OnboardingInstagramChannel: "instagram",
		OnboardingMessengerChannel: "messenger",
		OnboardingLineChannel:      "line",
		OnboardingSMSChannel:       "sms",
		OnboardingTiktokChannel:    "tiktok",
		OnboardingWebchatChannel:   "webchat",
	}
	for channel, want := range cases {
		assert.Equal(t, want, channel.String())
	}

	// Unknown values collapse to the "unknown" channel.
	assert.Equal(t, "unknown", OnboardingChannel("bogus").String())
	assert.Equal(t, "unknown", OnboardingUnknownChannel.String())
}

func TestOnboard_Onboarding_Success(t *testing.T) {
	mock := mockOK(`{"agent_id": "agent_123"}`)
	c := newTestClient(mock)

	agentID, err := c.Onboard.Onboarding(OnboardingWhatsappChannel)

	assert.NoError(t, err)
	assert.Equal(t, "agent_123", agentID)
	assert.LengthSlice(t, 1, mock.Calls)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_onboarding", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "channel", params[0].Data().Key)
	assert.Equal(t, "whatsapp", params[0].Data().Value.(string))
}

func TestOnboard_Onboarding_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrClosedPipe))

	agentID, err := c.Onboard.Onboarding(OnboardingWhatsappChannel)

	assert.Error(t, err)
	assert.Equal(t, "", agentID)
}

func TestOnboard_Onboarding_DecodeError(t *testing.T) {
	c := newTestClient(mockOK(`not-json`))

	agentID, err := c.Onboard.Onboarding(OnboardingEmailChannel)

	assert.Error(t, err)
	assert.Equal(t, "", agentID)
}

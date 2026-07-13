package agent

import (
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// Onboarding represents the result of an agent onboarding request.
type Onboarding struct {
	AgentID string `json:"agent_id"`
}

// OnboardingChannel identifies the messaging channel used during onboarding.
type OnboardingChannel string

func (ch OnboardingChannel) String() string {
	switch ch {
	case OnboardingEmailChannel,
		OnboardingInstagramChannel,
		OnboardingMessengerChannel,
		OnboardingWhatsappChannel,
		OnboardingLineChannel,
		OnboardingSMSChannel,
		OnboardingTiktokChannel,
		OnboardingWebchatChannel:
		return string(ch)
	default:
		return string(OnboardingUnknownChannel)
	}
}

const (
	OnboardingEmailChannel     OnboardingChannel = "email"
	OnboardingWhatsappChannel  OnboardingChannel = "whatsapp"
	OnboardingInstagramChannel OnboardingChannel = "instagram"
	OnboardingMessengerChannel OnboardingChannel = "messenger"
	OnboardingLineChannel      OnboardingChannel = "line"
	OnboardingSMSChannel       OnboardingChannel = "sms"
	OnboardingTiktokChannel    OnboardingChannel = "tiktok"
	OnboardingWebchatChannel   OnboardingChannel = "webchat"
	OnboardingUnknownChannel   OnboardingChannel = "unknown"
)

// Onboarding prepare the agent on a phone number.
//   - Trigger AI agent onboarding for the specified entity and channel.
//   - Creates the necessary entities and schedules async jobs for data preparation.
func (o *Onboard) Onboarding(channel OnboardingChannel) (agentId string, err error) {
	url := o.Config.URL("/agent_onboarding")
	headers := o.Client.Headers("application/json")
	var result Onboarding
	res, err := o.Client.Post(url, &xhttp.Options{
		Headers: headers,
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "channel", Value: channel.String()},
		},
	})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return "", err
	}
	return result.AgentID, nil
}

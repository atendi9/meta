package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// SettingsRollout represents the rollout configuration.
type SettingsRollout struct {
	// Enabled: Whether the AI agent is currently enabled.
	Enabled bool `json:"enabled"`
}

// SettingsHandoff represents settings for handing over to a human.
type SettingsHandoff struct {
	// Enabled: Whether handoff to a human agent is enabled.
	Enabled bool   `json:"enabled"`
	Message string `json:"message,omitzero"`
}

// SettingsFollowup represents settings for following up with inactive users.
type SettingsFollowup struct {
	// Enabled: Whether followup is enabled.
	Enabled                   bool   `json:"enabled"`
	FollowupIntervalInSeconds int    `json:"followup_interval_in_seconds,omitzero"`
	Message                   string `json:"message,omitzero"`
}

// AIAudience controls which consumers the AI agent responds to.
type AIAudience string

const (
	// AIAudiencePrivate restricts responses to phone numbers in the allowlist.
	//   - Only supported for WhatsApp entities.
	AIAudiencePrivate AIAudience = "ALLOWLISTED_ONLY"
	// AIAudiencePublic responds to all consumers (default).
	AIAudiencePublic AIAudience = "EVERYONE"
)

// SettingsRequest represents the body used in PUT requests.
type SettingsRequest struct {
	Rollout    SettingsRollout  `json:"rollout,omitzero"`
	Handoff    SettingsHandoff  `json:"handoff,omitzero"`
	Followup   SettingsFollowup `json:"followup,omitzero"`
	AIAudience string           `json:"ai_audience,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *SettingsRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// SettingsResponse represents the response object.
type SettingsResponse struct {
	AgentID    string            `json:"agent_id"`
	Channel    OnboardingChannel `json:"channel"`
	Rollout    SettingsRollout   `json:"rollout"`
	Handoff    SettingsHandoff   `json:"handoff,omitzero"`
	Followup   SettingsFollowup  `json:"followup,omitzero"`
	AIAudience string            `json:"ai_audience,omitzero"`
}

// Settings retrieve the current AI settings for the specified entity.
func (o *Onboard) Settings(agentId ...string) ([]SettingsResponse, error) {
	url := o.Config.URL("/agent_config/settings")
	headers := o.Client.Headers("application/json")
	var result struct {
		Root []SettingsResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers: headers,
	}
	if len(agentId) > 0 {
		opts.QueryParams = []xhttp.HTTPData{
			&xhttp.Data{Key: "agent_id", Value: agentId[0]},
		}
	}
	res, err := o.Client.Get(url, opts)
	if err != nil {
		return []SettingsResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []SettingsResponse{}, err
	}
	return result.Root, nil
}

// UpdateSettings update the AI settings for the specified entity.
func (o *Onboard) UpdateSettings(config SettingsRequest, agentId ...string) (SettingsResponse, error) {
	url := o.Config.URL("/agent_config/settings")
	headers := o.Client.Headers("application/json")
	var result SettingsResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &config,
	}
	if len(agentId) > 0 {
		opts.QueryParams = []xhttp.HTTPData{
			&xhttp.Data{Key: "agent_id", Value: agentId[0]},
		}
	}
	res, err := o.Client.Put(url, opts)
	if err != nil {
		return SettingsResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return SettingsResponse{}, err
	}
	return result, nil
}

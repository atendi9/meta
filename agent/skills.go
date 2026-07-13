package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// SkillRequest represents the body used to create or update a skill.
type SkillRequest struct {
	// Title is a human-readable name for the skill.
	//   - Max 64 characters, lowercase letters, numbers and hyphens only.
	Title string `json:"title,omitzero"`
	// Description is the trigger or context telling the AI when to apply this skill.
	//   - Max 1024 characters.
	Description string `json:"description,omitzero"`
	// Skill holds the instructions for the AI.
	//   - Max 20,000 characters.
	Skill string `json:"skill"`
	// Metadata holds arbitrary key-value pairs for additional context.
	Metadata map[string]any `json:"metadata,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *SkillRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// SkillResponse represents a skill object.
type SkillResponse struct {
	// ID is the unique identifier for this skill.
	ID string `json:"id"`
	// Title is the optional human-readable name of the skill.
	Title string `json:"title,omitzero"`
	// Description tells when the AI applies this skill.
	Description string `json:"description,omitzero"`
	// Skill holds the AI instructions.
	Skill string `json:"skill"`
	// Channel is the messaging channel the skill belongs to.
	Channel OnboardingChannel `json:"channel"`
	// CreatedAt is the creation timestamp.
	CreatedAt int64 `json:"created_at,omitzero"`
	// Metadata holds arbitrary key-value pairs for additional context.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// Skills provides CRUD access to the agent skills of an entity.
type Skills struct {
	o *Onboard
}

// Skills exposes the skill management operations for the configured entity.
func (c *Configurator) Skills() *Skills {
	return &Skills{o: c.Client.Onboard}
}

// agentIDParam builds the optional agent_id query parameter.
func agentIDParam(agentId []string) []xhttp.HTTPData {
	if len(agentId) == 0 {
		return nil
	}
	return []xhttp.HTTPData{
		&xhttp.Data{Key: "agent_id", Value: agentId[0]},
	}
}

// Create adds a new skill to the specified entity.
func (s *Skills) Create(skill SkillRequest, agentId ...string) (SkillResponse, error) {
	url := s.o.Config.URL("/agent_config/skills")
	headers := s.o.Client.Headers("application/json")
	var result SkillResponse
	opts := &xhttp.Options{
		Headers:     headers,
		Body:        &skill,
		QueryParams: agentIDParam(agentId),
	}
	res, err := s.o.Client.Post(url, opts)
	if err != nil {
		return SkillResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return SkillResponse{}, err
	}
	return result, nil
}

// List retrieves the skills for the specified entity.
func (s *Skills) List(agentId ...string) ([]SkillResponse, error) {
	url := s.o.Config.URL("/agent_config/skills")
	headers := s.o.Client.Headers("application/json")
	var result struct {
		Root []SkillResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers:     headers,
		QueryParams: agentIDParam(agentId),
	}
	res, err := s.o.Client.Get(url, opts)
	if err != nil {
		return []SkillResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []SkillResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single skill by its identifier.
func (s *Skills) Get(skillId string) (SkillResponse, error) {
	url := s.o.Config.URL("/agent_config/skills/" + skillId)
	headers := s.o.Client.Headers("application/json")
	var result SkillResponse
	res, err := s.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return SkillResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return SkillResponse{}, err
	}
	return result, nil
}

// Update modifies an existing skill identified by skillId.
func (s *Skills) Update(skillId string, skill SkillRequest) (SkillResponse, error) {
	url := s.o.Config.URL("/agent_config/skills/" + skillId)
	headers := s.o.Client.Headers("application/json")
	var result SkillResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &skill,
	}
	res, err := s.o.Client.Put(url, opts)
	if err != nil {
		return SkillResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return SkillResponse{}, err
	}
	return result, nil
}

// Delete removes a skill by its identifier.
func (s *Skills) Delete(skillId string) error {
	url := s.o.Config.URL("/agent_config/skills/" + skillId)
	headers := s.o.Client.Headers("application/json")
	_, err := s.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

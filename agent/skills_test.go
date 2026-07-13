package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAgentIDParam(t *testing.T) {
	assert.LengthSlice(t, 0, agentIDParam(nil))
	assert.LengthSlice(t, 0, agentIDParam([]string{}))

	params := agentIDParam([]string{"agent_9"})
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_9", params[0].Data().Value.(string))

	// Only the first value is used.
	params = agentIDParam([]string{"agent_9", "agent_10"})
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "agent_9", params[0].Data().Value.(string))
}

func TestSkills_Create_Success(t *testing.T) {
	mock := mockOK(`{"id":"s1","title":"greet","skill":"say hi","channel":"whatsapp"}`)
	s := newTestConfigurator(mock).Skills()

	skill, err := s.Create(SkillRequest{Title: "greet", Skill: "say hi"})

	assert.NoError(t, err)
	assert.Equal(t, "s1", skill.ID)
	assert.Equal(t, "greet", skill.Title)
	assert.Equal(t, OnboardingWhatsappChannel, skill.Channel)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/skills", mock.Calls[0].URL)
}

func TestSkills_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"s1","skill":"say hi"}]}`)
	s := newTestConfigurator(mock).Skills()

	skills, err := s.List("agent_1")

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, skills)
	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "agent_1", params[0].Data().Value.(string))
}

func TestSkills_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"s1","skill":"say hi"}`)
	s := newTestConfigurator(mock).Skills()

	skill, err := s.Get("s1")

	assert.NoError(t, err)
	assert.Equal(t, "s1", skill.ID)
	assert.Equal(t, testBase+"/agent_config/skills/s1", mock.Calls[0].URL)
}

func TestSkills_Update_Success(t *testing.T) {
	mock := mockOK(`{"id":"s1","skill":"say hello"}`)
	s := newTestConfigurator(mock).Skills()

	skill, err := s.Update("s1", SkillRequest{Skill: "say hello"})

	assert.NoError(t, err)
	assert.Equal(t, "say hello", skill.Skill)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/skills/s1", mock.Calls[0].URL)
}

func TestSkills_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	s := newTestConfigurator(mock).Skills()

	err := s.Delete("s1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/skills/s1", mock.Calls[0].URL)
}

func TestSkills_TransportErrors(t *testing.T) {
	s := newTestConfigurator(mockErr(io.ErrClosedPipe)).Skills()

	_, err := s.Create(SkillRequest{})
	assert.Error(t, err)
	_, err = s.List()
	assert.Error(t, err)
	_, err = s.Get("s1")
	assert.Error(t, err)
	_, err = s.Update("s1", SkillRequest{})
	assert.Error(t, err)
	err = s.Delete("s1")
	assert.Error(t, err)
}

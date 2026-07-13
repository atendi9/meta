package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAgentEval_Cases_Success(t *testing.T) {
	body := `{"eval_cases":[{"id":"c1","scenario":"greeting","categories":["support"],"max_turns":5,"success_criteria":["polite"]}]}`
	mock := mockOK(body)
	ev := newTestConfigurator(mock).Operate().AgentEval()

	cases, err := ev.Cases()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, cases)
	assert.Equal(t, "c1", cases[0].ID)
	assert.Equal(t, "greeting", cases[0].Scenario)
	assert.Equal(t, 5, cases[0].MaxTurns)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent-eval/cases", mock.Calls[0].URL)
}

func TestAgentEval_Details_Success(t *testing.T) {
	body := `{"evaluations":[{"id":"e1","score":80,"eval_case_id":"c1"}]}`
	mock := mockOK(body)
	ev := newTestConfigurator(mock).Operate().AgentEval()

	details, err := ev.Details("e1", "e2")

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, details)
	assert.Equal(t, "e1", details[0].ID)
	assert.Equal(t, 80, details[0].Score)
	assert.Equal(t, testBase+"/agent-eval/details", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "eval_ids", params[0].Data().Key)
	assert.Equal(t, "e1,e2", params[0].Data().Value.(string))
}

func TestAgentEval_Summary_Success(t *testing.T) {
	body := `{"insights":[{"id":"s1","avg_conversation_score":0.9,"avg_turn_score":0.8,"summary":"good"}]}`
	mock := mockOK(body)
	ev := newTestConfigurator(mock).Operate().AgentEval()

	insights, err := ev.Summary("s1")

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, insights)
	assert.Equal(t, "s1", insights[0].ID)
	assert.Equal(t, "good", insights[0].Summary)
	assert.Equal(t, testBase+"/agent-eval/summary", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "summary_ids", params[0].Data().Key)
	assert.Equal(t, "s1", params[0].Data().Value.(string))
}

func TestAgentEval_RunStatus_Success(t *testing.T) {
	body := `{"status":"RUNNING","progress":{"completed":2,"total":10,"current_stage":"scoring"},"result":{"summary_id":"s1"}}`
	mock := mockOK(body)
	ev := newTestConfigurator(mock).Operate().AgentEval()

	status, err := ev.RunStatus("job_1")

	assert.NoError(t, err)
	assert.Equal(t, "RUNNING", status.Status)
	assert.Equal(t, 2, status.Progress.Completed)
	assert.Equal(t, 10, status.Progress.Total)
	assert.Equal(t, "s1", status.Result.SummaryID)
	assert.Equal(t, testBase+"/agent-eval/run", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "job_id", params[0].Data().Key)
	assert.Equal(t, "job_1", params[0].Data().Value.(string))
}

func TestAgentEval_Run_Success(t *testing.T) {
	mock := mockOK(`{"job_id":"job_1","status":"QUEUED"}`)
	ev := newTestConfigurator(mock).Operate().AgentEval()

	res, err := ev.Run("c1", "c2")

	assert.NoError(t, err)
	assert.Equal(t, "job_1", res.JobID)
	assert.Equal(t, "QUEUED", res.Status)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent-eval/run", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "eval_case_ids", params[0].Data().Key)
	assert.Equal(t, "c1,c2", params[0].Data().Value.(string))

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.Equal(t, "{}", string(sent))
}

func TestAgentEval_TransportErrors(t *testing.T) {
	ev := newTestConfigurator(mockErr(io.ErrClosedPipe)).Operate().AgentEval()

	_, err := ev.Cases()
	assert.Error(t, err)
	_, err = ev.Details("e1")
	assert.Error(t, err)
	_, err = ev.Summary("s1")
	assert.Error(t, err)
	_, err = ev.RunStatus("job_1")
	assert.Error(t, err)
	_, err = ev.Run("c1")
	assert.Error(t, err)
}

func TestAgentEval_DecodeError(t *testing.T) {
	ev := newTestConfigurator(mockOK(`not-json`)).Operate().AgentEval()

	_, err := ev.Cases()

	assert.Error(t, err)
}

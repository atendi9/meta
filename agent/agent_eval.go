package agent

import (
	"bytes"
	"strings"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// EvalCase describes a scenario used to evaluate the agent.
type EvalCase struct {
	// ID is the unique identifier of the evaluation case.
	ID string `json:"id"`
	// Scenario describes the situation the agent is evaluated against.
	Scenario string `json:"scenario"`
	// Categories groups the case by topic.
	Categories []string `json:"categories,omitzero"`
	// MaxTurns is the maximum number of conversation turns.
	MaxTurns int `json:"max_turns,omitzero"`
	// SuccessCriteria lists the conditions for a successful evaluation.
	SuccessCriteria []string `json:"success_criteria,omitzero"`
}

// EvalDetail holds the result of a single evaluation.
type EvalDetail struct {
	// ID is the unique identifier of the evaluation.
	ID string `json:"id"`
	// Score is the evaluation score.
	Score int `json:"score,omitzero"`
	// PerTurnLabels holds a JSON array of per-turn labels.
	PerTurnLabels string `json:"per_turn_labels,omitzero"`
	// Reasons holds a JSON array of reasons for the score.
	Reasons string `json:"reasons,omitzero"`
	// CustomSuccessCriteria holds a JSON array of custom success criteria.
	CustomSuccessCriteria string `json:"custom_success_criteria,omitzero"`
	// EvalCaseID is the identifier of the evaluation case.
	EvalCaseID string `json:"eval_case_id,omitzero"`
	// Transcript holds the conversation transcript as a JSON object.
	Transcript string `json:"transcript,omitzero"`
	// CreationTime is the creation timestamp.
	CreationTime int64 `json:"creation_time,omitzero"`
	// UpdateTime is the last-update timestamp.
	UpdateTime int64 `json:"update_time,omitzero"`
}

// EvalSummary holds aggregated insights across a set of evaluations.
type EvalSummary struct {
	// ID is the unique identifier of the summary.
	ID string `json:"id"`
	// AvgConversationScore is the average conversation-level score.
	AvgConversationScore float64 `json:"avg_conversation_score,omitzero"`
	// AvgTurnScore is the average turn-level score.
	AvgTurnScore float64 `json:"avg_turn_score,omitzero"`
	// Summary is a human-readable summary of the results.
	Summary string `json:"summary,omitzero"`
	// Highlights holds a JSON array of notable results.
	Highlights string `json:"highlights,omitzero"`
	// TopFailureCategories holds a JSON array of the most common failure
	// categories.
	TopFailureCategories string `json:"top_failure_categories,omitzero"`
	// EvalIDsByScore holds a JSON object mapping scores to evaluation IDs.
	EvalIDsByScore string `json:"eval_ids_by_score,omitzero"`
	// CreationTime is the creation timestamp.
	CreationTime int64 `json:"creation_time,omitzero"`
	// UpdateTime is the last-update timestamp.
	UpdateTime int64 `json:"update_time,omitzero"`
}

// EvalRunProgress reports how far along an evaluation job is.
type EvalRunProgress struct {
	// Completed is the number of cases evaluated so far.
	Completed int `json:"completed,omitzero"`
	// Total is the total number of cases in the job.
	Total int `json:"total,omitzero"`
	// CurrentStage names the stage currently being processed.
	CurrentStage string `json:"current_stage,omitzero"`
}

// EvalRunResult holds the aggregated outcome of a completed evaluation job.
type EvalRunResult struct {
	// SummaryID identifies the generated summary.
	SummaryID string `json:"summary_id,omitzero"`
	// AvgConversationScore is the average conversation-level score.
	AvgConversationScore float64 `json:"avg_conversation_score,omitzero"`
	// AvgTurnScore is the average turn-level score.
	AvgTurnScore float64 `json:"avg_turn_score,omitzero"`
	// Summary is a human-readable summary of the results.
	Summary string `json:"summary,omitzero"`
	// Highlights holds a JSON array of notable results.
	Highlights string `json:"highlights,omitzero"`
	// TopFailureCategories holds a JSON array of the most common failure
	// categories.
	TopFailureCategories string `json:"top_failure_categories,omitzero"`
	// EvalIDsByScore holds a JSON object mapping scores to evaluation IDs.
	EvalIDsByScore string `json:"eval_ids_by_score,omitzero"`
	// CreationTime is the creation timestamp.
	CreationTime int64 `json:"creation_time,omitzero"`
	// UpdateTime is the last-update timestamp.
	UpdateTime int64 `json:"update_time,omitzero"`
}

// EvalRunError describes why an evaluation job failed.
type EvalRunError struct {
	// Code is the error code.
	Code string `json:"code,omitzero"`
	// Message is a human-readable error message.
	Message string `json:"message,omitzero"`
	// FailedCaseIDs lists the evaluation cases that failed.
	FailedCaseIDs []string `json:"failed_case_ids,omitzero"`
}

// EvalRunStatus reports the status of an evaluation job.
type EvalRunStatus struct {
	// Status is one of "QUEUED", "RUNNING", "COMPLETED" or "FAILED".
	Status string `json:"status"`
	// Progress reports how far along the job is.
	Progress EvalRunProgress `json:"progress,omitzero"`
	// Result holds the aggregated outcome when the job is complete.
	Result EvalRunResult `json:"result,omitzero"`
	// Error describes why the job failed, when applicable.
	Error EvalRunError `json:"error,omitzero"`
}

// EvalRunResponse acknowledges a newly started evaluation job.
type EvalRunResponse struct {
	// JobID identifies the started job, used to poll its status.
	JobID string `json:"job_id"`
	// Status is the initial job status, typically "QUEUED".
	Status string `json:"status"`
}

// AgentEval measures agent effectiveness through evaluation cases and jobs.
type AgentEval struct {
	o *Onboard
}

// AgentEval exposes the agent evaluation operations for the configured entity.
func (op *Operator) AgentEval() *AgentEval {
	return &AgentEval{o: op.o}
}

// Cases lists the available evaluation cases for the entity.
func (e *AgentEval) Cases() ([]EvalCase, error) {
	url := e.o.Config.URL("/agent-eval/cases")
	headers := e.o.Client.Headers("application/json")
	var result struct {
		EvalCases []EvalCase `json:"eval_cases"`
	}
	res, err := e.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return nil, err
	}
	return result.EvalCases, nil
}

// Details retrieves the individual results for the given evaluation IDs.
func (e *AgentEval) Details(evalIDs ...string) ([]EvalDetail, error) {
	url := e.o.Config.URL("/agent-eval/details")
	headers := e.o.Client.Headers("application/json")
	var result struct {
		Evaluations []EvalDetail `json:"evaluations"`
	}
	opts := &xhttp.Options{
		Headers: headers,
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "eval_ids", Value: strings.Join(evalIDs, ",")},
		},
	}
	res, err := e.o.Client.Get(url, opts)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return nil, err
	}
	return result.Evaluations, nil
}

// Summary retrieves the aggregated insights for the given summary IDs.
func (e *AgentEval) Summary(summaryIDs ...string) ([]EvalSummary, error) {
	url := e.o.Config.URL("/agent-eval/summary")
	headers := e.o.Client.Headers("application/json")
	var result struct {
		Insights []EvalSummary `json:"insights"`
	}
	opts := &xhttp.Options{
		Headers: headers,
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "summary_ids", Value: strings.Join(summaryIDs, ",")},
		},
	}
	res, err := e.o.Client.Get(url, opts)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return nil, err
	}
	return result.Insights, nil
}

// RunStatus retrieves the status of the evaluation job identified by jobID.
func (e *AgentEval) RunStatus(jobID string) (EvalRunStatus, error) {
	url := e.o.Config.URL("/agent-eval/run")
	headers := e.o.Client.Headers("application/json")
	var result EvalRunStatus
	opts := &xhttp.Options{
		Headers: headers,
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "job_id", Value: jobID},
		},
	}
	res, err := e.o.Client.Get(url, opts)
	if err != nil {
		return EvalRunStatus{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return EvalRunStatus{}, err
	}
	return result, nil
}

// Run starts an evaluation job for the given evaluation case IDs and returns
// the created job.
func (e *AgentEval) Run(evalCaseIDs ...string) (EvalRunResponse, error) {
	url := e.o.Config.URL("/agent-eval/run")
	headers := e.o.Client.Headers("application/json")
	var result EvalRunResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    bytes.NewReader([]byte("{}")),
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "eval_case_ids", Value: strings.Join(evalCaseIDs, ",")},
		},
	}
	res, err := e.o.Client.Post(url, opts)
	if err != nil {
		return EvalRunResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return EvalRunResponse{}, err
	}
	return result, nil
}

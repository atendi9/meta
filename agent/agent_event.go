package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// AgentEventPayload describes the business occurrence that triggers an agent
// response.
type AgentEventPayload struct {
	// Type is the partner-defined event identifier.
	//   - Max 256 characters.
	Type string `json:"type"`
	// Description is a human-readable description of the event.
	//   - Max 1024 characters.
	Description string `json:"description"`
	// Payload holds the event data as a JSON string.
	//   - Max 4096 characters.
	Payload string `json:"payload"`
}

// AgentEventRequest represents the body used to trigger an agent event.
type AgentEventRequest struct {
	// To is the consumer phone number in E.164 format.
	To string `json:"to"`
	// Event holds the event details.
	Event AgentEventPayload `json:"event"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *AgentEventRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// AgentEventResponse represents the acknowledgment returned when an agent event
// is accepted for asynchronous processing.
type AgentEventResponse struct {
	// Status is "accepted" when the event is queued.
	Status string `json:"status"`
	// AgentEventID identifies the submitted event, used to poll its status.
	AgentEventID string `json:"agent_event_id"`
}

// AgentEventStatus represents the processing status of a submitted agent event.
type AgentEventStatus struct {
	// Status is one of "request_received", "processing", "sent", "failed",
	// "skipped" or "success".
	Status string `json:"status"`
	// EventType is the partner-defined identifier of the event.
	EventType string `json:"event_type,omitzero"`
	// ErrorMessage is present when Status is "failed".
	ErrorMessage string `json:"error_message,omitzero"`
	// SkippedReason is present when Status is "skipped".
	SkippedReason string `json:"skipped_reason,omitzero"`
	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at,omitzero"`
	// UpdatedAt is the ISO 8601 last-update timestamp.
	UpdatedAt string `json:"updated_at,omitzero"`
}

// AgentEvents triggers agent responses from business occurrences and reports
// their processing status.
type AgentEvents struct {
	o *Onboard
}

// AgentEvent exposes the agent event operations for the configured entity.
func (op *Operator) AgentEvent() *AgentEvents {
	return &AgentEvents{o: op.o}
}

// Send triggers an agent event asynchronously and returns its acknowledgment.
func (a *AgentEvents) Send(event AgentEventRequest) (AgentEventResponse, error) {
	url := a.o.Config.URL("/agent_event")
	headers := a.o.Client.Headers("application/json")
	var result AgentEventResponse
	res, err := a.o.Client.Post(url, &xhttp.Options{Headers: headers, Body: &event})
	if err != nil {
		return AgentEventResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return AgentEventResponse{}, err
	}
	return result, nil
}

// Get retrieves the processing status of a previously submitted agent event.
func (a *AgentEvents) Get(agentEventID string) (AgentEventStatus, error) {
	url := a.o.Config.URL("/agent_event/" + agentEventID)
	headers := a.o.Client.Headers("application/json")
	var result AgentEventStatus
	res, err := a.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return AgentEventStatus{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return AgentEventStatus{}, err
	}
	return result, nil
}

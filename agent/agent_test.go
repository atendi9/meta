package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// AgentTestRequest represents the body used to send a test message to the AI
// agent.
type AgentTestRequest struct {
	// UserMsg is the text content of the test message to send to the AI agent.
	UserMsg string `json:"user_msg"`
	// ConversationID continues a multi-turn test conversation. Leave empty to
	// start a new conversation; the returned ConversationID can be supplied on
	// subsequent calls.
	ConversationID string `json:"conversation_id,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *AgentTestRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// AgentTestResponse represents the AI agent reply to a test message.
type AgentTestResponse struct {
	// MessageID uniquely identifies the message exchange.
	MessageID string `json:"message_id"`
	// AgentResponse is the AI agent response text.
	AgentResponse string `json:"agent_response"`
	// ConversationID identifies the test conversation for follow-up requests.
	ConversationID string `json:"conversation_id"`
	// Timestamp is the Unix timestamp of the response generation.
	Timestamp int64 `json:"timestamp,omitzero"`
	// HandoffReason is present when the agent transfers to human support.
	HandoffReason string `json:"handoff_reason,omitzero"`
	// NoResponseReason is present when no response is generated, for example
	// "ELIGIBILITY_CHECK_FAILED".
	NoResponseReason string `json:"no_response_reason,omitzero"`
	// QuickReplies holds suggested reply messages.
	QuickReplies []string `json:"quick_replies,omitzero"`
	// ProductVariantIDs holds referenced product variant IDs.
	ProductVariantIDs []string `json:"product_variant_ids,omitzero"`
}

// AgentTest validates agent behavior by submitting test messages and returning
// the agent's response.
type AgentTest struct {
	o *Onboard
}

// AgentTest exposes the agent test operations for the configured entity.
func (op *Operator) AgentTest() *AgentTest {
	return &AgentTest{o: op.o}
}

// Send submits a test message to the AI agent and returns its response.
func (t *AgentTest) Send(msg AgentTestRequest) (AgentTestResponse, error) {
	url := t.o.Config.URL("/agent_test")
	headers := t.o.Client.Headers("application/json")
	var result AgentTestResponse
	res, err := t.o.Client.Post(url, &xhttp.Options{Headers: headers, Body: &msg})
	if err != nil {
		return AgentTestResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return AgentTestResponse{}, err
	}
	return result, nil
}

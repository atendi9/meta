package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAgentEvent_Send_Success(t *testing.T) {
	mock := mockOK(`{"status":"accepted","agent_event_id":"evt_123"}`)
	ev := newTestConfigurator(mock).Operate().AgentEvent()

	res, err := ev.Send(AgentEventRequest{
		To: "5511999999999",
		Event: AgentEventPayload{
			Type:        "order_shipped",
			Description: "The customer order has shipped",
			Payload:     `{"order_id":"A1"}`,
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "accepted", res.Status)
	assert.Equal(t, "evt_123", res.AgentEventID)
	assert.LengthSlice(t, 1, mock.Calls)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_event", mock.Calls[0].URL)

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	body := string(sent)
	assert.True(t, contains(body, `"to": "5511999999999"`))
	assert.True(t, contains(body, `"type": "order_shipped"`))
}

func TestAgentEvent_Send_TransportError(t *testing.T) {
	ev := newTestConfigurator(mockErr(io.ErrClosedPipe)).Operate().AgentEvent()

	_, err := ev.Send(AgentEventRequest{To: "5511999999999"})

	assert.Error(t, err)
}

func TestAgentEvent_Get_Success(t *testing.T) {
	body := `{"status":"failed","event_type":"order_shipped","error_message":"boom","created_at":"2026-07-13T00:00:00Z","updated_at":"2026-07-13T00:01:00Z"}`
	mock := mockOK(body)
	ev := newTestConfigurator(mock).Operate().AgentEvent()

	status, err := ev.Get("evt_123")

	assert.NoError(t, err)
	assert.Equal(t, "failed", status.Status)
	assert.Equal(t, "order_shipped", status.EventType)
	assert.Equal(t, "boom", status.ErrorMessage)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_event/evt_123", mock.Calls[0].URL)
}

func TestAgentEvent_Get_TransportError(t *testing.T) {
	ev := newTestConfigurator(mockErr(io.ErrUnexpectedEOF)).Operate().AgentEvent()

	_, err := ev.Get("evt_123")

	assert.Error(t, err)
}

func TestAgentEvent_Get_DecodeError(t *testing.T) {
	ev := newTestConfigurator(mockOK(`not-json`)).Operate().AgentEvent()

	_, err := ev.Get("evt_123")

	assert.Error(t, err)
}

package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAgentTest_Send_Success(t *testing.T) {
	body := `{
		"message_id":"m1",
		"agent_response":"Hello there!",
		"conversation_id":"conv_1",
		"timestamp":1752374400,
		"quick_replies":["Yes","No"],
		"product_variant_ids":["v1"]
	}`
	mock := mockOK(body)
	at := newTestConfigurator(mock).Operate().AgentTest()

	res, err := at.Send(AgentTestRequest{UserMsg: "hi", ConversationID: "conv_1"})

	assert.NoError(t, err)
	assert.Equal(t, "m1", res.MessageID)
	assert.Equal(t, "Hello there!", res.AgentResponse)
	assert.Equal(t, "conv_1", res.ConversationID)
	assert.Equal(t, int64(1752374400), res.Timestamp)
	assert.LengthSlice(t, 2, res.QuickReplies)
	assert.LengthSlice(t, 1, res.ProductVariantIDs)

	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_test", mock.Calls[0].URL)

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.True(t, contains(string(sent), `"user_msg": "hi"`))
}

func TestAgentTest_Send_NoResponseReason(t *testing.T) {
	mock := mockOK(`{"message_id":"m1","no_response_reason":"ELIGIBILITY_CHECK_FAILED"}`)
	at := newTestConfigurator(mock).Operate().AgentTest()

	res, err := at.Send(AgentTestRequest{UserMsg: "hi"})

	assert.NoError(t, err)
	assert.Equal(t, "ELIGIBILITY_CHECK_FAILED", res.NoResponseReason)
}

func TestAgentTest_Send_TransportError(t *testing.T) {
	at := newTestConfigurator(mockErr(io.ErrUnexpectedEOF)).Operate().AgentTest()

	_, err := at.Send(AgentTestRequest{UserMsg: "hi"})

	assert.Error(t, err)
}

func TestAgentTest_Send_DecodeError(t *testing.T) {
	at := newTestConfigurator(mockOK(`not-json`)).Operate().AgentTest()

	_, err := at.Send(AgentTestRequest{UserMsg: "hi"})

	assert.Error(t, err)
}

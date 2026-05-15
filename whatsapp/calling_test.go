package whatsapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

type mockEventEmitter struct {
	events map[string]any
}

func newMockEventEmitter() *mockEventEmitter {
	return &mockEventEmitter{events: make(map[string]any)}
}

func (m *mockEventEmitter) Emit(eventName string, data any) {
	m.events[eventName] = data
}

func TestNewCall(t *testing.T) {
	api := &Client{}
	emitter := newMockEventEmitter()

	call := NewCall(api, "app_123", emitter)

	assert.Equal(t, "app_123", call.appId)
	assert.Equal(t, api, call.api)
	assert.Equal(t, fmt.Sprint(emitter), fmt.Sprint(call.io))

	logTriggered := false
	customLogger := func(msg string) {
		logTriggered = true
	}

	callWithLogger := NewCall(api, "app_123", emitter, customLogger)
	callWithLogger.logger("test")
	assert.True(t, logTriggered)
}

func TestAcceptCall(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
			BaseUrl:    "https://graph.facebook.com",
			ApiVersion: "v19.0",
		},
	}

	emitter := newMockEventEmitter()
	call := NewCall(api, "app_123", emitter)

	result := call.AcceptCall("call_abc123", "v=0\r\n...")

	assert.Equal(t, "call_abc123", result)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
	expectedURL := api.Endpoint(api.senderID + "/calls")
	assert.Equal(t, expectedURL, mockClient.Calls[0].URL)
}

func TestInitiateOutboundCall_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"call_id": "call_999"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
			BaseUrl:    "https://graph.facebook.com",
			ApiVersion: "v19.0",
		},
	}

	emitter := newMockEventEmitter()
	call := NewCall(api, "app_123", emitter)

	result, err := call.InitiateOutboundCall("5511999999999", "v=0\r\n...")

	assert.NoError(t, err)
	assert.Equal(t, "call_999", result.CallId)
	assert.Equal(t, "app_123", result.AppId)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func TestInitiateOutboundCall_FallbackId(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "call_888"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	call := NewCall(api, "app_123", newMockEventEmitter())
	result, err := call.InitiateOutboundCall("5511999999999", "v=0\r\n...")

	assert.NoError(t, err)
	assert.Equal(t, "call_888", result.CallId)
}

func TestInitiateOutboundCall_HTTPError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, errors.New("network error"))

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	call := NewCall(api, "app_123", newMockEventEmitter())
	_, err := call.InitiateOutboundCall("5511999999999", "v=0\r\n...")

	assert.Error(t, err)
}

func TestInitiateOutboundCall_NoIdReturned(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	call := NewCall(api, "app_123", newMockEventEmitter())
	_, err := call.InitiateOutboundCall("5511999999999", "v=0\r\n...")

	assert.Error(t, err)
	assert.Equal(t, ErrCannotGetCallId.Error(), err.Error())
}

func TestInitiateOutboundCall_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`not-json`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	call := NewCall(api, "app_123", newMockEventEmitter())
	_, err := call.InitiateOutboundCall("5511999999999", "v=0\r\n...")

	assert.Error(t, err)
}

func TestWebhookPreAccept_Ringing(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
			BaseUrl:    "https://graph.facebook.com",
			ApiVersion: "v19.0",
		},
	}

	emitter := newMockEventEmitter()
	call := NewCall(api, "app_123", emitter)

	sdp := "a=setup:actpass"
	call.WebhookPreAccept("call_111", CallData{Status: "ringing"}, sdp)

	assert.LengthSlice(t, 1, mockClient.Calls)
	
	_, ok := emitter.events["incoming_call"]
	assert.True(t, ok)
	
	emittedData := emitter.events["incoming_call"].(xjson.JSON)
	assert.Equal(t, "app_123", emittedData["appId"])
	assert.Equal(t, "call_111", emittedData["callId"])
}

func TestWebhookPreAccept_Active(t *testing.T) {
	api := &Client{}
	emitter := newMockEventEmitter()
	call := NewCall(api, "app_123", emitter)

	call.WebhookPreAccept("call_222", CallData{Status: "active"}, "answer_sdp_here")

	_, ok := emitter.events["outbound_call_answered"]
	assert.True(t, ok)

	emittedData := emitter.events["outbound_call_answered"].(xjson.JSON)
	assert.Equal(t, "call_222", emittedData["callId"])
	assert.Equal(t, "answer_sdp_here", emittedData["answerSdp"])
}

func TestWebhookPreAccept_EndedAndCanceled(t *testing.T) {
	api := &Client{}
	
	emitterEnded := newMockEventEmitter()
	callEnded := NewCall(api, "app_123", emitterEnded)
	callEnded.WebhookPreAccept("call_333", CallData{Status: "ended"}, "")

	_, okEnded := emitterEnded.events["call_ended"]
	assert.True(t, okEnded)
	assert.Equal(t, "call_333", emitterEnded.events["call_ended"].(xjson.JSON)["callId"])

	emitterCanceled := newMockEventEmitter()
	callCanceled := NewCall(api, "app_123", emitterCanceled)
	callCanceled.WebhookPreAccept("call_444", CallData{Status: "canceled"}, "")

	_, okCanceled := emitterCanceled.events["call_ended"]
	assert.True(t, okCanceled)
	assert.Equal(t, "call_444", emitterCanceled.events["call_ended"].(xjson.JSON)["callId"])
}
package whatsapp

import (
	"errors"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta/xhttp"
)

func TestWebhookSubscribe_Success(t *testing.T) {
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       nil,
	}
	mock := xhttp.NewMockClient(mockResponse, nil)

	client := Default("sender123", "token_abc")
	client.HttpClient = mock

	wabaID := "waba456"
	pin := WebhookPin{
		Register:  "123456",
		Subscribe: "654321",
	}

	err := WebhookSubscribe(client, wabaID, pin)
	assert.NoError(t, err)

	assert.LengthSlice(t, 2, mock.Calls)
}

func TestWebhookSubscribe_RegisterError(t *testing.T) {
	expectedErr := errors.New("failed to register")
	mock := xhttp.NewMockClient(nil, expectedErr)

	client := Default("sender123", "token_abc")
	client.HttpClient = mock

	pin := WebhookPin{Register: "123", Subscribe: "456"}

	err := WebhookSubscribe(client, "waba456", pin)
	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
	assert.LengthSlice(t, 1, mock.Calls)
}

func TestWebhookSubscribe_SubscribeError(t *testing.T) {
	expectedErr := errors.New("failed to subscribe")
	successRes := &http.Response{StatusCode: http.StatusOK}

	mock := xhttp.NewMockClient(successRes, nil)

	client := Default("sender123", "token_abc")
	client.HttpClient = mock

	wabaID := "waba456"
	pin := WebhookPin{Register: "123", Subscribe: "456"}

	registerURL := client.Endpoint(client.senderID + "/register")
	subscribeURL := client.Endpoint(wabaID + "/subscribed_apps")
	mock.MapResponse(http.MethodPost, subscribeURL, nil, expectedErr)

	err := WebhookSubscribe(client, wabaID, pin)

	assert.Error(t, err)
	assert.Equal(t, expectedErr.Error(), err.Error())

	assert.LengthSlice(t, 2, mock.Calls)

	assert.Equal(t, registerURL, mock.Calls[0].URL)
	assert.Equal(t, subscribeURL, mock.Calls[1].URL)
}

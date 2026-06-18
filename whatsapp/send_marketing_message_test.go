package whatsapp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestBuildMarketingMessage(t *testing.T) {
	t.Run("should keep a phone number in the to field", func(t *testing.T) {
		payload := []byte(`{"messaging_product":"whatsapp","to":"5511999999999","type":"template","template":{"name":"promo"}}`)

		message, err := BuildMarketingMessage(payload)
		assert.NoError(t, err)
		assert.Equal(t, "5511999999999", message.To)
		assert.Equal(t, "", message.Recipient)
		assert.Equal(t, "template", message.Type)
		assert.Equal(t, "promo", message.Template["name"])
	})

	t.Run("should move a BSUID from to into recipient", func(t *testing.T) {
		payload := []byte(`{"messaging_product":"whatsapp","to":"US.13491208655302741918","type":"template"}`)

		message, err := BuildMarketingMessage(payload)
		assert.NoError(t, err)
		assert.Equal(t, "", message.To)
		assert.Equal(t, "US.13491208655302741918", message.Recipient)
	})

	t.Run("should preserve a recipient already provided as a BSUID", func(t *testing.T) {
		payload := []byte(`{"messaging_product":"whatsapp","recipient":"US.ENT.11815799212886844830","type":"template"}`)

		message, err := BuildMarketingMessage(payload)
		assert.NoError(t, err)
		assert.Equal(t, "", message.To)
		assert.Equal(t, "US.ENT.11815799212886844830", message.Recipient)
	})

	t.Run("should return an error for an invalid payload", func(t *testing.T) {
		message, err := BuildMarketingMessage([]byte(`{invalid_json}`))
		assert.Error(t, err)
		ok := message == nil
		assert.True(t, ok)
	})

	t.Run("should rebuild a sendable payload routing the BSUID", func(t *testing.T) {
		payload := []byte(`{"messaging_product":"whatsapp","to":"BR.5511999999999","type":"template"}`)

		message, err := BuildMarketingMessage(payload)
		assert.NoError(t, err)

		rebuilt, err := BuildMarketingMessage(message.Bytes())
		assert.NoError(t, err)
		assert.Equal(t, "BR.5511999999999", rebuilt.Recipient)
		assert.Equal(t, "", rebuilt.To)
	})
}

func TestSendMarketingMessage_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages": [{"id": "wamid.123456789"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	payload := []byte(`{"messaging_product": "whatsapp", "to": "5511999999999", "type": "template"}`)

	res, err := SendMarketingMessage(api, payload)
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)

	ok := len(res.FirstId()) > 0
	assert.True(t, ok)
}

func TestSendMarketingMessage_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	payload := []byte(`{"messaging_product": "whatsapp"}`)

	res, err := SendMarketingMessage(api, payload)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprint(&SendMessageResponse{}), fmt.Sprint(res))
}

func TestSendMarketingMessage_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{invalid_json}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	payload := []byte(`{"messaging_product": "whatsapp"}`)

	res, err := SendMarketingMessage(api, payload)

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprint(&SendMessageResponse{}), fmt.Sprint(res))
}

func TestSendMarketingMessage_NotSentError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages": []}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	payload := []byte(`{"messaging_product": "whatsapp"}`)

	res, err := SendMarketingMessage(api, payload)
	assert.Error(t, err)
	assert.Equal(t, ErrMessageNotSent, err)

	ok := res != nil
	assert.True(t, ok)
}

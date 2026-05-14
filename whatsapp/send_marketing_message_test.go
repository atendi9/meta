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

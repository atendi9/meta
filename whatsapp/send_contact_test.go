package whatsapp

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestSendContact_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[{"id": "wamid_contact_123"}]}`)),
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

	receiver := "5511999999999"
	name := "John Doe da Silva"
	phone := "+55 11 98888-8888"

	id, err := api.SendContact(receiver, name, phone)
	assert.NoError(t, err)
	assert.Equal(t, "wamid_contact_123", id)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestSendContact_InvalidName(t *testing.T) {
	api := &Client{}
	id, err := api.SendContact("123", "John", "555-5555")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidContactName, err)
	assert.Equal(t, "", id)
}

func TestSendContact_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	id, err := api.SendContact("5511999999999", "John Doe", "555-0000")

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

func TestSendContact_EmptyResponse(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	id, err := api.SendContact("5511999999999", "John Doe", "555-1234")
	assert.Equal(t, err, ErrMessageNotSent)
	assert.Equal(t, "", id)
}

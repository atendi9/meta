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

func TestSendInteractiveList_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[{"id": "wamid_list_123"}]}`)),
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

	header := Header{"to": "5511999999999"}
	opts := SendInteractiveListOpts{
		Header:     "Our Menu",
		Message:    "Please select an option below:",
		Footer:     "Thank you for choosing us",
		ButtonText: "View Options",
		Rows: map[string][]InteractiveListRow{
			"Food": {
				{Id: "1", Title: "Burger", Description: "Delicious beef burger"},
				{Id: "2", Title: "Pizza", Description: "Cheese and tomato"},
			},
		},
	}

	id, err := api.SendInteractiveList(header, opts)

	assert.NoError(t, err)
	assert.Equal(t, "wamid_list_123", id)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestSendInteractiveList_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	header := Header{"to": "5511999999999"}
	opts := SendInteractiveListOpts{
		Header: "Test Header",
	}

	id, err := api.SendInteractiveList(header, opts)

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

func TestSendInteractiveList_EmptyResponse(t *testing.T) {
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

	header := Header{"to": "5511999999999"}
	opts := SendInteractiveListOpts{
		Header: "Test Header",
	}

	id, err := api.SendInteractiveList(header, opts)

	assert.Equal(t, ErrMessageNotSent, err)
	assert.Equal(t, "", id)
}

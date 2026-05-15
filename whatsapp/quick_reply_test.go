package whatsapp

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

func TestQuickReply(t *testing.T) {

	t.Run("GenerateQuickReplyButton", func(t *testing.T) {
		btn := GenerateQuickReplyButton("btn_1", "Accept")

		assert.Equal(t, "reply", btn.Type)
		assert.Equal(t, "btn_1", btn.Reply.Id)
		assert.Equal(t, "Accept", btn.Reply.Title)
	})

	t.Run("SendQuickReplyMessage_NoButtonsError", func(t *testing.T) {
		client := &Client{}
		h := make(Header)

		id, err := client.SendQuickReplyMessage(h, "Enterprise", "Message body", nil)

		assert.Error(t, err)
		assert.Equal(t, ErrAtLeastOneButtonRequired, err)
		assert.Equal(t, "", id)
	})

	t.Run("SendQuickReplyMessage_RequestError", func(t *testing.T) {
		mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

		client := &Client{
			GraphAPIClient: meta.GraphAPIClient{
				HttpClient: mockClient,
			},
		}

		h := make(Header)
		buttons := []QuickReplyButton{
			GenerateQuickReplyButton("btn_1", "Option 1"),
		}

		id, err := client.SendQuickReplyMessage(h, "Enterprise", "Message body", buttons)

		assert.Error(t, err)
		assert.Equal(t, "", id)
	})

	t.Run("SendQuickReplyMessage_Success", func(t *testing.T) {
		mockRes := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"messages": [{"id": "wamid.12345"}]}`)),
		}
		mockClient := xhttp.NewMockClient(mockRes, nil)

		client := &Client{
			GraphAPIClient: meta.GraphAPIClient{
				HttpClient: mockClient,
			},
		}

		h := make(Header)
		buttons := []QuickReplyButton{
			GenerateQuickReplyButton("btn_1", "Option 1"),
			GenerateQuickReplyButton("btn_2", "Option 2"),
		}

		id, err := client.SendQuickReplyMessage(h, "Enterprise", "Message body", buttons)

		assert.NoError(t, err)
		assert.Equal(t, "wamid.12345", id)
		assert.LengthSlice(t, 1, mockClient.Calls)
		
		// Verifying if the header was mutated properly
		assert.Equal(t, "interactive", h["type"])
	})

	t.Run("SendQuickReplyMessage_TooManyButtonsError", func(t *testing.T) {
		mockClient := xhttp.NewMockClient(nil, nil)

		client := &Client{
			GraphAPIClient: meta.GraphAPIClient{
				HttpClient: mockClient,
			},
		}

		h := make(Header)

		buttons := []QuickReplyButton{
			GenerateQuickReplyButton("btn_1", "Option 1"),
			GenerateQuickReplyButton("btn_2", "Option 2"),
			GenerateQuickReplyButton("btn_3", "Option 3"),
			GenerateQuickReplyButton("btn_4", "Option 4"),
		}

		id, err := client.SendQuickReplyMessage(h, "Enterprise", "Message body", buttons)

		assert.Error(t, err)
		assert.Equal(t, ErrTooManyButtons, err)
		assert.Equal(t, "", id)
		// No request should be made when validation fails.
		assert.LengthSlice(t, 0, mockClient.Calls)
	})

	t.Run("SendQuickReplyMessage_ThreeButtonsAllowed", func(t *testing.T) {
		mockRes := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"messages": [{"id": "wamid.12345"}]}`)),
		}
		mockClient := xhttp.NewMockClient(mockRes, nil)

		client := &Client{
			GraphAPIClient: meta.GraphAPIClient{
				HttpClient: mockClient,
			},
		}

		h := make(Header)

		buttons := []QuickReplyButton{
			GenerateQuickReplyButton("btn_1", "Option 1"),
			GenerateQuickReplyButton("btn_2", "Option 2"),
			GenerateQuickReplyButton("btn_3", "Option 3"),
		}

		id, err := client.SendQuickReplyMessage(h, "Enterprise", "Message body", buttons)

		assert.NoError(t, err)
		assert.Equal(t, "wamid.12345", id)

		interactiveData, ok := h["interactive"].(xjson.JSON)
		assert.True(t, ok)

		actionData, ok := interactiveData["action"].(xjson.JSON)
		assert.True(t, ok)

		finalButtons, ok := actionData["buttons"].([]QuickReplyButton)
		assert.True(t, ok)

		assert.LengthSlice(t, 3, finalButtons)
	})
}
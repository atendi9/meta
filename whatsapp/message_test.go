package whatsapp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

func TestSendMessage(t *testing.T) {
	mockHTTP := xhttp.NewMockClient(nil, nil)
	client := &Client{
		senderID: "12345",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockHTTP,
			BaseUrl:    "https://graph.facebook.com",
			ApiVersion: "v20.0",
		},
	}

	t.Run("MessageHeader", func(t *testing.T) {
		h := MessageHeader("5511999999999", "text")
		assert.Equal(t, "whatsapp", h["messaging_product"])
		assert.Equal(t, "individual", h["recipient_type"])
		assert.Equal(t, "5511999999999", h["to"])
		assert.Equal(t, "text", h["type"])
		assert.False(t, h["context"] != nil)

		hReply := MessageHeader("5511999999999", "text", "original_msg_id")
		ctx, ok := hReply["context"].(xjson.JSON)
		assert.True(t, ok)
		assert.Equal(t, "original_msg_id", ctx["message_id"])
	})

	t.Run("TextMessage", func(t *testing.T) {
		h := MessageHeader("123", "text")
		msg := TextMessage(h, "Hello World")
		textBody, ok := msg["text"].(xjson.JSON)
		assert.True(t, ok)
		assert.Equal(t, "Hello World", textBody["body"])
	})

	t.Run("MediaMessages", func(t *testing.T) {
		h := MessageHeader("123", "media")
		media := Media{Id: "media_01", Caption: "Check this out"}

		t.Run("AudioMessage", func(t *testing.T) {
			msg := AudioMessage(h, media)
			audio, ok := msg["audio"].(Media)
			assert.True(t, ok)
			assert.Equal(t, media, audio)
		})

		t.Run("DocumentMessage", func(t *testing.T) {
			msg := DocumentMessage(h, media)
			document, ok := msg["document"].(Media)
			assert.True(t, ok)
			assert.Equal(t, media, document)
		})

		t.Run("ImageMessage", func(t *testing.T) {
			msg := ImageMessage(h, media)
			image, ok := msg["image"].(Media)
			assert.True(t, ok)
			assert.Equal(t, media, image)
		})

		t.Run("VideoMessage", func(t *testing.T) {
			msg := VideoMessage(h, media)
			video, ok := msg["video"].(Media)
			assert.True(t, ok)
			assert.Equal(t, media, video)
		})

		t.Run("StickerMessage", func(t *testing.T) {
			msg := StickerMessage(h, media)
			sticker, ok := msg["sticker"].(Media)
			assert.True(t, ok)
			assert.Equal(t, media, sticker)
		})
	})

	t.Run("Reaction", func(t *testing.T) {
		h := MessageHeader("123", "reaction")
		reaction := ReactionBody{MessageId: "msg_1", Emoji: "🚀"}
		msg := Reaction(h, reaction)
		reactionEmoji, ok := msg["reaction"].(ReactionBody)
		assert.True(t, ok)
		assert.Equal(t, reaction, reactionEmoji)
	})

	t.Run("MessagesSent_Methods", func(t *testing.T) {
		ms := MessagesSent{{Id: "a"}, {Id: "b"}}
		assert.Equal(t, 2, ms.Len())
		assert.Equal(t, "a", ms.FirstId())

		empty := MessagesSent{}
		assert.Equal(t, 0, empty.Len())
		assert.Equal(t, "", empty.FirstId())
	})

	t.Run("MessagesEndpointRequest", func(t *testing.T) {
		mockJSON := `{"messages": [{"id": "wa_id_123"}]}`
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(mockJSON)),
		}

		res, err := MessagesEndpointRequest(client, []byte(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, "wa_id_123", res.FirstId())

		mockEmptyJSON := `{"messages": []}`
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(mockEmptyJSON)),
		}

		_, err = MessagesEndpointRequest(client, []byte(`{}`))
		assert.Error(t, err)
		assert.Equal(t, ErrMessageNotSent, err)

		mockHTTP.DefaultResponse = nil
		mockHTTP.DefaultErr = errors.New("network fail")

		_, err = MessagesEndpointRequest(client, []byte(`{}`))
		assert.Error(t, err)
		assert.Equal(t, "network fail", err.Error())

		mockHTTP.DefaultErr = nil
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`not-json`)),
		}

		_, err = MessagesEndpointRequest(client, []byte(`{}`))
		assert.Error(t, err)
	})

	t.Run("SendMessage", func(t *testing.T) {
		mockJSON := `{"messages": [{"id": "sent_ok"}]}`
		mockHTTP.DefaultErr = nil
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(mockJSON)),
		}

		h := MessageHeader("123", "text")
		msg := TextMessage(h, "Final Test")

		res, err := SendMessage(client, msg)
		assert.NoError(t, err)
		assert.Equal(t, "sent_ok", res.FirstId())
	})
}

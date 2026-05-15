package whatsapp

import (
	"errors"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// Header is an alias for [xjson.JSON] used to define message headers.
type Header = xjson.JSON

// Message is an alias for [xjson.JSON] representing the final message payload.
type Message = xjson.JSON

// MessageHeader creates a basic [Header] for a WhatsApp message.
// It sets the recipient number, message type, and optionally a reply context if replyId is provided.
func MessageHeader(
	receiverNumber string,
	msgType string,
	replyId ...string,
) Header {
	header := xjson.JSON{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                receiverNumber,
		"type":              msgType,
	}
	if len(replyId) > 0 {
		header["context"] = xjson.JSON{
			"message_id": replyId[0],
		}
	}
	return header
}

// SendMessage sends a [Message] using the provided [Client].
func SendMessage(
	whats *Client,
	body Message,
) (*SendMessageResponse, error) {
	res, err := MessagesEndpointRequest(whats, body.Bytes())
	return &res, err
}

// ErrMessageNotSent is returned when the API response does not contain a valid message ID.
var ErrMessageNotSent = errors.New("message not sent")

// MessagesEndpointRequest performs a request to the messages endpoint.
func MessagesEndpointRequest(
	whats *Client,
	b []byte,
) (SendMessageResponse, error) {
	url := whats.Endpoint(whats.senderID + "/messages")
	res, err := whats.Post(url, &xhttp.Options{
		Body:    whats.Reader(b),
		Headers: whats.Headers("application/json"),
	})
	if err != nil {
		return SendMessageResponse{}, err
	}
	defer res.Body.Close()

	var response SendMessageResponse
	if err := xjson.Decode(res.Body, &response); err != nil {
		return SendMessageResponse{}, err
	}

	if !response.Ok() {
		return response, ErrMessageNotSent
	}
	return response, nil
}

// TextMessage appends a text body to the provided [Header] and returns it as a [Message].
func TextMessage(h Header, body string) Message {
	textType := "text"
	h["type"] = textType
	h[textType] = xjson.JSON{
		"body": body,
	}
	return h
}

// Media represents the media object structure for WhatsApp messages (Audio, Image, Video, etc).
type Media struct {
	Id       string `json:"id"`
	Filename string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// AudioMessage attaches an [Media] audio object to the [Header].
func AudioMessage(h Header, audio Media) Message {
	audioType := "audio"
	h["type"] = audioType
	h[audioType] = audio
	return h
}

// DocumentMessage attaches an [Media] document object to the [Header].
func DocumentMessage(h Header, document Media) Message {
	documentType := "document"
	h["type"] = documentType
	h[documentType] = document
	return h
}

// ImageMessage attaches an [Media] image object to the [Header].
func ImageMessage(h Header, image Media) Message {
	imageType := "image"
	h["type"] = imageType
	h[imageType] = image
	return h
}

// VideoMessage attaches an [Media] video object to the [Header].
func VideoMessage(h Header, video Media) Message {
	videoType := "video"
	h["type"] = videoType
	h[videoType] = video
	return h
}

// StickerMessage attaches an [Media] sticker object to the [Header].
func StickerMessage(h Header, sticker Media) Message {
	stickerType := "sticker"
	h["type"] = stickerType
	h[stickerType] = sticker
	return h
}

// Reaction attaches a [ReactionBody] to the [Header].
func Reaction(h Header, reaction ReactionBody) Message {
	reactionType := "reaction"
	h["type"] = reactionType
	h[reactionType] = reaction
	return h
}

// ReactionBody defines the structure for sending an emoji reaction to a specific message.
type ReactionBody struct {
	MessageId string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

// MessageSent represents the basic information of a message successfully sent by the API.
type MessageSent struct {
	Id string `json:"id"`
}

// MessagesSent is a slice of [MessageSent].
type MessagesSent []MessageSent

// Len returns the number of messages in the [MessagesSent] slice.
func (m MessagesSent) Len() int {
	return len(m)
}

// FirstId returns the ID of the first message in the slice, or an empty string if empty.
func (m MessagesSent) FirstId() string {
	if m.Len() > 0 {
		return m[0].Id
	}
	return ""
}

// SendMessageResponse represents the API response containing the list of sent messages.
type SendMessageResponse struct {
	MessagesSent `json:"messages"`
	Success      bool `json:"success"`
}

func (r SendMessageResponse) Ok() bool {
	return len(r.FirstId()) > 0 || r.Success
}

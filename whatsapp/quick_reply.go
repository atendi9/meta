package whatsapp

import (
	"errors"

	"github.com/atendi9/meta/xhttp/xjson"
)

// QuickReplyButton represents a quick reply button used in an interactive message.
type QuickReplyButton struct {
	Type  string      `json:"type"`
	Reply ReplyButton `json:"reply"`
}

// ReplyButton contains the unique identifier and the title text for a quick reply button.
type ReplyButton struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

// ErrAtLeastOneButtonRequired is returned when trying to send an interactive message without buttons.
var ErrAtLeastOneButtonRequired = errors.New("at least one button is required")

// ErrTooManyButtons is returned when more than the Meta-allowed number of
// quick reply buttons (3) is provided.
var ErrTooManyButtons = errors.New("a maximum of 3 quick reply buttons is allowed")

// SendQuickReplyMessage sends an interactive message containing quick reply buttons.
func (api *Client) SendQuickReplyMessage(
	h Header,
	enterpriseName, message string,
	buttons []QuickReplyButton,
) (id string, err error) {
	if len(buttons) == 0 {
		return "", ErrAtLeastOneButtonRequired
	}
	if len(buttons) > 3 {
		return "", ErrTooManyButtons
	}
	interactiveType := "interactive"
	h["type"] = interactiveType
	h[interactiveType] = xjson.JSON{
		"type": "button",
		"body": xjson.JSON{
			"text": message,
		},
		"footer": xjson.JSON{
			"text": enterpriseName,
		},
		"action": xjson.JSON{
			"buttons": buttons,
		},
	}
	res, err := MessagesEndpointRequest(api, h.Bytes())
	if err != nil {
		return "", err
	}
	return res.FirstId(), nil
}

// GenerateQuickReplyButton is a helper function that creates and returns a new [QuickReplyButton].
func GenerateQuickReplyButton(
	id, title string,
) QuickReplyButton {
	qrb := QuickReplyButton{
		Type: "reply",
		Reply: ReplyButton{
			Id:    id,
			Title: title,
		},
	}
	return qrb
}

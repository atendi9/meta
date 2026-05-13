package whatsapp

import (
	"github.com/atendi9/meta/xhttp/xjson"
)

func TypingIndicator(
	api *Client,
	messageId string,
) error {
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        messageId,
		"typing_indicator": xjson.JSON{
			"type": "text",
		},
	}.Bytes()
	_, err := MessagesEndpointRequest(api, body)
	return err
}

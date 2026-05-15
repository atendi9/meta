package whatsapp

import (
	"github.com/atendi9/meta/xhttp/xjson"
)

// MarkMessageAsRead sends a request using the provided [Client] to update the message status to read.
func MarkMessageAsRead(
	api *Client,
	messageID string,
) error {
	_, err := MessagesEndpointRequest(api, xjson.JSON{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        messageID,
	}.Bytes())
	return err
}

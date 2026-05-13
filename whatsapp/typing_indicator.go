package whatsapp

import (
	"github.com/atendi9/meta/xhttp/xjson"
)

// TypingIndicator sends a request to the WhatsApp Messages API to mark a specific message
// as read and display a typing indicator to the user.
//
// The typing indicator informs the user that a response is being prepared, which is 
// a best practice when the response takes a few seconds to be generated.
//
// The indicator is automatically removed after a reply is sent or after 25 seconds, 
// whichever occurs first. It should only be triggered if a response is intended 
// to be sent to avoid a poor user experience.
//
// Parameters:
//   - api: [Client] The WhatsApp API client instance.
//   - messageId: [string] The unique ID of the message to be marked as read.
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
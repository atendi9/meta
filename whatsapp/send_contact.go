package whatsapp

import (
	"errors"
	"strings"

	"github.com/atendi9/meta/xhttp/xjson"
)

// ErrInvalidContactName is returned when the contact name does not contain at least a first and last name.
var ErrInvalidContactName = errors.New("invalid contact name")

// SendContact sends a contact card to a specific WhatsApp number.
//
// 	- It uses [xjson.JSON] to construct the request body and interacts with the WhatsApp Messages API.
//	- Returns the message ID or an error if the validation or the [MessagesEndpointRequest] fails.
func (c *Client) SendContact(
	receiverNumber,
	contactName,
	contactPhone string,
) (id string, err error) {
	contactNameParts := strings.Split(contactName, " ")
	if len(contactNameParts) < 2 {
		return "", ErrInvalidContactName
	}

	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"to":                receiverNumber,
		"type":              "contacts",
		"contacts": []xjson.JSON{
			{
				"name": xjson.JSON{
					"formatted_name": contactName,
					"first_name":     contactNameParts[0],
					"last_name":      contactNameParts[1],
				},
				"phones": []xjson.JSON{
					{
						"phone": contactPhone,
					},
				},
			},
		},
	}.Bytes()

	res, err := MessagesEndpointRequest(c, body)
	return res.FirstId(), err
}
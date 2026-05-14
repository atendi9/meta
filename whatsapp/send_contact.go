package whatsapp

import (
	"errors"
	"strings"
	"unicode"

	"github.com/atendi9/meta/xhttp/xjson"
)

// ErrInvalidContactName is returned when the contact name does not contain at least a first and last name.
var ErrInvalidContactName = errors.New("invalid contact name")

// SendContact sends a contact card to a specific WhatsApp number.
//
//   - It uses [xjson.JSON] to construct the request body and interacts with the WhatsApp Messages API.
//   - Returns the message ID or an error if the validation or the [MessagesEndpointRequest] fails.
func (c *Client) SendContact(
	receiverNumber,
	contactName,
	contactPhone string,
) (id string, err error) {
	contactNameParts := splitContactName(contactName)
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

// splitContactName splits a name into first/last name resiliently.
func splitContactName(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		return parts
	}
	if camel := splitCamelCase(name); len(camel) >= 2 {
		return camel
	}
	if snake := splitSnakeCase(name); len(snake) >= 2 {
		return snake
	}
	return []string{name}
}

// splitCamelCase divides a string whenever an uppercase letter follows a lowercase one.
//   - Example: "ValRamos" -> ["Val", "Ramos"] or "JoaoSilvaSauro" -> ["Joao", "Silva", "Sauro"].
//   - Consecutive uppercase letters are kept together.
func splitCamelCase(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	var current []rune
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runes[i-1]) {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = current[:0]
			}
		}
		current = append(current, r)
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	return parts
}

// splitSnakeCase divides a string by underscores, keeping only the non-empty parts.
//   - Example: "val_ramos" -> ["val", "ramos"].
func splitSnakeCase(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	split := strings.Split(s, "_")
	for _, p := range split {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

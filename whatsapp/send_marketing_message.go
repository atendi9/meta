package whatsapp

import (
	"bytes"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// MarketingMessage models a WhatsApp Marketing Messages API payload. It mirrors
// the documented fields a marketing request may carry: a recipient addressed by
// phone number ([MarketingMessage.To]) or BSUID ([MarketingMessage.Recipient]),
// the message type, and the free-form template content.
type MarketingMessage struct {
	MessagingProduct string     `json:"messaging_product,omitempty"`
	RecipientType    string     `json:"recipient_type,omitempty"`
	To               string     `json:"to,omitempty"`
	Recipient        string     `json:"recipient,omitempty"`
	Type             string     `json:"type,omitempty"`
	Template         xjson.JSON `json:"template,omitempty"`
}

// BuildMarketingMessage decodes a raw marketing message payload into a
// [MarketingMessage] and normalizes its recipient. When the "to" field holds a
// BSUID (for example "US.13491208655302741918"), the value is moved to the
// "recipient" field as required by the Marketing Messages API, so existing
// callers may keep placing either a phone number or a BSUID in "to" without
// changing their payload. A value already supplied in "recipient" is preserved.
func BuildMarketingMessage(body []byte) (*MarketingMessage, error) {
	var message MarketingMessage
	if err := xjson.Decode(bytes.NewReader(body), &message); err != nil {
		return nil, err
	}
	message.routeRecipient()
	return &message, nil
}

// routeRecipient moves a BSUID found in [MarketingMessage.To] into
// [MarketingMessage.Recipient], leaving phone numbers untouched.
func (m *MarketingMessage) routeRecipient() {
	if m.To != "" && IsBSUID(m.To) {
		m.Recipient = m.To
		m.To = ""
	}
}

// Bytes marshals the [MarketingMessage] into the JSON payload expected by
// [SendMarketingMessage].
func (m *MarketingMessage) Bytes() []byte {
	return xjson.Bytes(m)
}

// SendMarketingMessage sends a marketing message using the WhatsApp API.
//   - It requires a [Client] and a byte slice representing the JSON payload.
//   - It returns a [SendMessageResponse] and an error if the HTTP request fails,
//   - if the response cannot be decoded, or if the message could not be sent.
func SendMarketingMessage(
	api *Client,
	body []byte,
) (*SendMessageResponse, error) {
	url := api.Endpoint(api.senderID + "/marketing_messages")
	msg, err := BuildMarketingMessage(body)
	if err != nil {
		return &SendMessageResponse{}, err
	}
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers("application/json"),
		Body:    api.Reader(msg.Bytes()),
	})
	if err != nil {
		return &SendMessageResponse{}, err
	}
	defer res.Body.Close()

	var response SendMessageResponse
	if err := xjson.Decode(res.Body, &response); err != nil {
		return &SendMessageResponse{}, err
	}

	if len(response.FirstId()) == 0 {
		return &response, ErrMessageNotSent
	}
	return &response, nil
}

package whatsapp

import (
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// SendMarketingMessage sends a marketing message using the WhatsApp API.
//   - It requires a [Client] and a byte slice representing the JSON payload.
//   - It returns a [SendMessageResponse] and an error if the HTTP request fails,
//   - if the response cannot be decoded, or if the message could not be sent.
func SendMarketingMessage(
	api *Client,
	body []byte,
) (*SendMessageResponse, error) {
	url := api.Endpoint(api.senderID + "/marketing_messages")
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers("application/json"),
		Body:    api.Reader(body),
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

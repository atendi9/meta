package whatsapp

import (
	"fmt"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// WebhookPin defines the necessary credentials for the registration 
// and subscription steps of the WhatsApp Webhook.
type WebhookPin struct {
	Register  string
	Subscribe string
}

// WebhookSubscribe registers the sender and subscribes the application 
// to the specified WhatsApp Business Account (wabaId).
//
// 	- It uses the provided [Client] to perform the requests and requires 
// 	a [WebhookPin] containing the registration and subscription 2FA pins.
func WebhookSubscribe(
	whats *Client,
	wabaId string,
	pin WebhookPin,
) error {
	registerEndpoint := whats.Endpoint(fmt.Sprintf("%s/register", whats.SenderID()))
	subscribeEndpoint := whats.Endpoint(fmt.Sprintf("%s/subscribed_apps", wabaId))

	buf := xjson.JSON{
		"messaging_product": "whatsapp",
		"pin":               pin.Register,
	}.Buffer()

	registerRes, err := whats.Post(registerEndpoint, &xhttp.Options{
		Headers: whats.Headers("application/json"),
		Body:    buf,
	})
	if err != nil {
		return err
	}
	drainAndClose(registerRes)

	buf = xjson.JSON{
		"messaging_product": "whatsapp",
		"pin":               pin.Subscribe,
	}.Buffer()

	subscribeRes, err := whats.Post(subscribeEndpoint, &xhttp.Options{
		Headers: whats.Headers("application/json"),
		Body:    buf,
	})
	if err != nil {
		return err
	}
	drainAndClose(subscribeRes)

	return nil
}
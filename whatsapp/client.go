// Package whatsapp provides a client and utilities for interacting with the WhatsApp Business API.
package whatsapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

// BaseUrl represents the default Graph API base URL used by the client.
const BaseUrl = meta.GraphAPIBaseUrl

// Client represents the WhatsApp API client.
// It encapsulates the sender ID and inherits from [meta.GraphAPIClient] to handle API requests.
type Client struct {
	senderID string
	meta.GraphAPIClient
}

// Default creates and returns a new [Client] with the provided senderId and accessToken,
// using a background [context.Context] and the default Graph API version.
func Default(senderId, accessToken string) *Client {
	return New(
		context.Background(),
		meta.GraphAPIVersion,
		accessToken,
		senderId,
	)
}

// New creates and returns a new [Client] with the specified [context.Context],
// API version, access token, and sender ID.
func New(
	ctx context.Context,
	apiVersion,
	accessToken,
	senderID string,
) *Client {
	return &Client{
		senderID: senderID,
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  xhttp.New(ctx),
			ApiVersion:  apiVersion,
			BaseUrl:     BaseUrl,
			AccessToken: accessToken,
		},
	}
}

// SenderID returns the current sender ID configured in the [Client].
func (c *Client) SenderID() string {
	return c.senderID
}

// ChangeSenderId updates the sender ID of the [Client] to the provided id.
func (c *Client) ChangeSenderId(id string) {
	c.senderID = id
}

// Endpoint constructs and returns the full API URL string for a given endpoint.
func (c *Client) Endpoint(endpoint string) string {
	return fmt.Sprintf("%s/%s/%s", c.BaseUrl, c.ApiVersion, endpoint)
}

// GenerateWhatsappLink creates a direct WhatsApp API link with the provided phone number
// and URL-escaped text.
func GenerateWhatsappLink(phone, text string) string {
	escapedText := url.QueryEscape(text)
	return fmt.Sprintf("https://api.whatsapp.com/send?phone=%s&text=%s", phone, escapedText)
}

// drainAndClose drains and closes an HTTP response body so the underlying
// connection can be reused by the keep-alive pool. It tolerates a nil
// response or a nil body, so callers can invoke it unconditionally.
func drainAndClose(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()
}

// Headers returns a slice of [xhttp.HTTPData] containing the standard headers needed for API requests,
// including the specified Content-Type and the Bearer Authorization token.
func (c *Client) Headers(contentType string) []xhttp.HTTPData {
	headers := []xhttp.HTTPData{&xhttp.Header{

		Key:   "Content-Type",
		Value: contentType,
	},
		&xhttp.Header{

			Key:   "Authorization",
			Value: fmt.Sprintf("Bearer %s", c.AccessToken),
		},
	}
	return headers
}

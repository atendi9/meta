package agent

import (
	"context"
	"fmt"

	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

type Client struct {
	entityId string
	meta.GraphAPIClient
	Onboard   *Onboard
	AllowList *AllowList
}

const (
	APIVersion = "2.0.0"
	BaseURL    = "https://api.facebook.com"
)

func New(
	ctx context.Context,
	entityId,
	accessToken string,
) *Client {
	c := &Client{
		entityId: entityId,
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  xhttp.New(ctx),
			ApiVersion:  APIVersion,
			BaseUrl:     BaseURL,
			AccessToken: accessToken,
		},
	}
	c.Onboard = &Onboard{
		Client: c,
		Config: Config{
			EntityID:   c.entityId,
			BaseURL:    c.BaseUrl,
			ApiVersion: c.ApiVersion,
		},
	}
	c.AllowList = &AllowList{
		o: c.Onboard,
	}
	return c
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
		&xhttp.Header{
			Key:   "X-API-Version",
			Value: c.ApiVersion,
		},
	}
	return headers
}

// Package meta provides the core client and utilities for interacting with the Meta Graph API.
package meta

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
)

const (
	// GraphAPIVersion defines the default API version used by [GraphAPIClient].
	GraphAPIVersion = "v24.0"

	// GraphAPIBaseUrl defines the base URL for the Graph API used by [GraphAPIClient].
	GraphAPIBaseUrl = "https://graph.facebook.com"
)

type (
	// HttpClient is an alias for the external [xhttp.HTTPClient].
	HttpClient = xhttp.HTTPClient

	// GraphAPIClient represents the main client structure for Meta Graph API requests.
	// It embeds [HttpClient] to handle HTTP communications.
	GraphAPIClient struct {
		HttpClient
		ApiVersion  string
		BaseUrl     string
		AccessToken string
	}
)

// Buffer creates and returns a new [bytes.Buffer] from the provided byte slice.
func (c GraphAPIClient) Buffer(buf []byte) *bytes.Buffer {
	return bytes.NewBuffer(buf)
}

// Reader returns an [io.Reader] wrapping the provided byte slice.
func (c GraphAPIClient) Reader(buf []byte) io.Reader {
	return c.Buffer(buf)
}

// Writer returns an [io.Writer] wrapping the provided byte slice.
func (c GraphAPIClient) Writer(buf []byte) io.Writer {
	return c.Buffer(buf)
}

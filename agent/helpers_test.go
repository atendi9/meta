package agent

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/atendi9/meta/xhttp"
)

// contains reports whether s contains substr; a small alias keeping the test
// assertions terse.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

const (
	testEntityID    = "123456789"
	testAccessToken = "valid_token"
	// testBase is the URL prefix produced by Config.URL for the test entity.
	testBase = BaseURL + "/" + testEntityID
)

// newTestClient builds an agent [Client] whose HTTP transport is the provided
// mock, allowing endpoints to be exercised without real network calls. Because
// New wires Onboard.Client to the returned Client, swapping the embedded
// HttpClient afterwards makes every sub-accessor use the mock too.
func newTestClient(mock xhttp.HTTPClient) *Client {
	c := New(context.Background(), testEntityID, testAccessToken)
	c.HttpClient = mock
	return c
}

// newTestConfigurator builds a [Configurator] backed by the provided mock.
func newTestConfigurator(mock xhttp.HTTPClient) *Configurator {
	return Configure(newTestClient(mock))
}

// okResponse builds a 200 response whose body is the given JSON string.
func okResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

// mockOK returns a mock client that answers every call with a 200 and body.
func mockOK(body string) *xhttp.MockClient {
	return xhttp.NewMockClient(okResponse(body), nil)
}

// mockErr returns a mock client that fails every call with err.
func mockErr(err error) *xhttp.MockClient {
	return xhttp.NewMockClient(nil, err)
}

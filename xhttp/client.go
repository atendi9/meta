package xhttp

import (
	"context"
	"net/http"
)

// HTTPClient defines the interface for an HTTP client.
type HTTPClient interface {
	// Get executes an HTTP GET request to the given URL with the provided [HTTPOptions].
	Get(url string, options HTTPOptions) (*http.Response, error)
	// Post executes an HTTP POST request to the given URL with the provided [HTTPOptions].
	Post(url string, options HTTPOptions) (*http.Response, error)
	// Put executes an HTTP PUT request to the given URL with the provided [HTTPOptions].
	Put(url string, options HTTPOptions) (*http.Response, error)
	// Patch executes an HTTP PATCH request to the given URL with the provided [HTTPOptions].
	Patch(url string, options HTTPOptions) (*http.Response, error)
	// Delete executes an HTTP DELETE request to the given URL with the provided [HTTPOptions].
	Delete(url string, options HTTPOptions) (*http.Response, error)
	// Options executes an HTTP OPTIONS request to the given URL with the provided [HTTPOptions].
	Options(url string, options HTTPOptions) (*http.Response, error)
	// Head executes an HTTP HEAD request to the given URL with the provided [HTTPOptions].
	Head(url string, options HTTPOptions) (*http.Response, error)
	// Connect executes an HTTP CONNECT request to the given URL with the provided [HTTPOptions].
	Connect(url string, options HTTPOptions) (*http.Response, error)
	// Trace executes an HTTP TRACE request to the given URL with the provided [HTTPOptions].
	Trace(url string, options HTTPOptions) (*http.Response, error)
}

// Client represents the HTTP client implementation.
// It holds a default [context.Context] and a pointer to an [http.Client].
type Client struct {
	ctx    context.Context
	client *http.Client
}

// Context returns the default [context.Context] associated with the [Client].
func (c *Client) Context() context.Context {
	return c.ctx
}

// New initializes and returns a new [HTTPClient].
// It accepts a [context.Context] and an optional variadic [http.Client] config.
func New(
	ctx context.Context,
	config ...http.Client,
) HTTPClient {
	client := &http.Client{}
	if len(config) > 0 {
		client = &config[0]
	}
	return &Client{ctx: ctx, client: client}
}

// Get executes an HTTP GET request using the underlying [Client].
func (c *Client) Get(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodGet, url, options)
}

// Post executes an HTTP POST request using the underlying [Client].
func (c *Client) Post(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodPost, url, options)
}

// Put executes an HTTP PUT request using the underlying [Client].
func (c *Client) Put(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodPut, url, options)
}

// Patch executes an HTTP PATCH request using the underlying [Client].
func (c *Client) Patch(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodPatch, url, options)
}

// Delete executes an HTTP DELETE request using the underlying [Client].
func (c *Client) Delete(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodDelete, url, options)
}

// Options executes an HTTP OPTIONS request using the underlying [Client].
func (c *Client) Options(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodOptions, url, options)
}

// Head executes an HTTP HEAD request using the underlying [Client].
func (c *Client) Head(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodHead, url, options)
}

// Connect executes an HTTP CONNECT request using the underlying [Client].
func (c *Client) Connect(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodConnect, url, options)
}

// Trace executes an HTTP TRACE request using the underlying [Client].
func (c *Client) Trace(url string, options HTTPOptions) (*http.Response, error) {
	return c.executeRequest(http.MethodTrace, url, options)
}

// executeRequest prepares and executes the HTTP request.
// It applies query parameters, headers, cookies, basic authentication, and custom context
// from the [HTTPOptions] to the [http.Request].
func (c *Client) executeRequest(method, url string, options HTTPOptions) (*http.Response, error) {
	var (
		headers     = options.H()
		body        = options.B()
		queryParams = options.Q()
		cookies     = options.C()
		ctx         = options.Ctx()
		auth        = options.Auth()
	)

	url = SetQueryParams(url, queryParams...)

	// Use custom context if provided, otherwise fallback to the client's default context.
	reqCtx := c.ctx
	if ctx != nil {
		reqCtx = ctx
	}

	req, err := http.NewRequestWithContext(reqCtx, method, url, body)
	if err != nil {
		return nil, NewRequestError(err)
	}

	SetHeaders(req, headers)
	SetCookies(req, cookies)
	SetBasicAuth(req, auth)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, NewRequestError(err)
	}
	if res.StatusCode >= http.StatusBadRequest {
		return res, &RequestError{
			StatusCode: res.StatusCode,
			Payload:    res.Body,
		}
	}

	return res, nil
}

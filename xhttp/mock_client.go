package xhttp

import (
	"net/http"
)

// MockClient is a mock implementation of the [HTTPClient] interface for testing purposes.
// It records the arguments of the last call to allow verification in tests.
type MockClient struct {
	CalledURL     string
	CalledMethod  string
	CalledOptions HTTPOptions
	Response      *http.Response
	Err           error
}

// NewMockClient returns a mock implementation of the [HTTPClient] interface for testing purposes.
func NewMockClient(response *http.Response, err error) *MockClient {
	return &MockClient{Response: response, Err: err}
}

// Get records the GET request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Get(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodGet, url, options)
}

// Post records the POST request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Post(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPost, url, options)
}

// Put records the PUT request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Put(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPut, url, options)
}

// Patch records the PATCH request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Patch(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPatch, url, options)
}

// Delete records the DELETE request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Delete(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodDelete, url, options)
}

// Options records the OPTIONS request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Options(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodOptions, url, options)
}

// Head records the HEAD request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Head(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodHead, url, options)
}

// Connect records the CONNECT request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Connect(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodConnect, url, options)
}

// Trace records the TRACE request and returns the pre-defined [http.Response] and error.
func (m *MockClient) Trace(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodTrace, url, options)
}

// recordCall populates the [MockClient] fields with the call data.
func (m *MockClient) recordCall(method, url string, options HTTPOptions) (*http.Response, error) {
	m.CalledMethod = method
	m.CalledURL = url
	m.CalledOptions = options
	return m.Response, m.Err
}

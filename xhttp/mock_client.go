package xhttp

import (
	"fmt"
	"net/http"
)

// Call represents the arguments of an HTTP call recorded by the [MockClient].
type Call struct {
	Method  string
	URL     string
	Options HTTPOptions
}

// mockResponse encapsulates the expected return values for a mocked call.
type mockResponse struct {
	res *http.Response
	err error
}

// MockClient is a mock implementation of the [HTTPClient] interface.
// It records all calls in a history slice and allows O(1) response mapping.
type MockClient struct {
	Calls           []Call
	mappedResponses map[string]mockResponse
	DefaultResponse *http.Response
	DefaultErr      error
}

// NewMockClient returns a new instance of [MockClient].
// If a default response or error is provided, it will be used when no specific mapping exists.
func NewMockClient(defaultRes *http.Response, defaultErr error) *MockClient {
	return &MockClient{
		Calls:           []Call{},
		mappedResponses: make(map[string]mockResponse),
		DefaultResponse: defaultRes,
		DefaultErr:      defaultErr,
	}
}

// MapResponse sets a specific response for a method and URL combination (O(1) lookup).
func (m *MockClient) MapResponse(method, url string, res *http.Response, err error) {
	key := fmt.Sprintf("%s:%s", method, url)
	m.mappedResponses[key] = mockResponse{res: res, err: err}
}

// Clear resets the history of calls recorded by the [MockClient].
func (m *MockClient) Clear() {
	m.Calls = []Call{}
}

func (m *MockClient) Get(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodGet, url, options)
}

func (m *MockClient) Post(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPost, url, options)
}

func (m *MockClient) Put(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPut, url, options)
}

func (m *MockClient) Patch(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodPatch, url, options)
}

func (m *MockClient) Delete(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodDelete, url, options)
}

func (m *MockClient) Options(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodOptions, url, options)
}

func (m *MockClient) Head(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodHead, url, options)
}

func (m *MockClient) Connect(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodConnect, url, options)
}

func (m *MockClient) Trace(url string, options HTTPOptions) (*http.Response, error) {
	return m.recordCall(http.MethodTrace, url, options)
}

// recordCall logs the call into the history and returns the mapped or default response.
func (m *MockClient) recordCall(method, url string, options HTTPOptions) (*http.Response, error) {
	m.Calls = append(m.Calls, Call{
		Method:  method,
		URL:     url,
		Options: options,
	})

	key := fmt.Sprintf("%s:%s", method, url)
	if mapped, ok := m.mappedResponses[key]; ok {
		return mapped.res, mapped.err
	}

	return m.DefaultResponse, m.DefaultErr
}

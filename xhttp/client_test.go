package xhttp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Method", r.Method)

		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	defaultClient := New(ctx)
	opts := mockOptions{}

	t.Run("New_DefaultClient", func(t *testing.T) {
		c := New(ctx)
		cImpl, ok := c.(*Client)

		assert.True(t, ok)
		assert.Equal(t, ctx, cImpl.ctx)
		assert.True(t, cImpl.client != nil)
	})

	t.Run("New_WithCustomConfig", func(t *testing.T) {
		customHTTP := http.Client{Timeout: 10}
		c := New(ctx, customHTTP)
		cImpl, ok := c.(*Client)

		assert.True(t, ok)
		assert.Equal(t, customHTTP.Timeout, cImpl.client.Timeout)
	})

	t.Run("Context", func(t *testing.T) {
		c := New(ctx)
		cImpl, ok := c.(*Client)

		assert.True(t, ok)
		assert.Equal(t, ctx, cImpl.Context())
	})

	methods := []struct {
		name       string
		methodCall func(url string, options HTTPOptions) (*http.Response, error)
		expected   string
	}{
		{"Get", defaultClient.Get, http.MethodGet},
		{"Post", defaultClient.Post, http.MethodPost},
		{"Put", defaultClient.Put, http.MethodPut},
		{"Patch", defaultClient.Patch, http.MethodPatch},
		{"Delete", defaultClient.Delete, http.MethodDelete},
		{"Options", defaultClient.Options, http.MethodOptions},
		{"Head", defaultClient.Head, http.MethodHead},
		{"Connect", defaultClient.Connect, http.MethodConnect},
		{"Trace", defaultClient.Trace, http.MethodTrace},
	}

	for _, tc := range methods {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.methodCall(ts.URL, opts)

			assert.NoError(t, err)
			assert.True(t, res != nil)

			assert.Equal(t, tc.expected, res.Request.Method)
		})
	}

	t.Run("executeRequest_InvalidURL", func(t *testing.T) {
		cImpl, ok := defaultClient.(*Client)
		assert.True(t, ok)

		res, err := cImpl.executeRequest(http.MethodGet, "http://%", opts)

		assert.Error(t, err)
		assert.False(t, res != nil)
	})

	t.Run("executeRequest_HTTPErrorStatus", func(t *testing.T) {
		cImpl, ok := defaultClient.(*Client)
		assert.True(t, ok)

		res, err := cImpl.executeRequest(http.MethodGet, ts.URL+"/error", opts)

		assert.Error(t, err)
		assert.True(t, res != nil)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("executeRequest_CustomContextOverride", func(t *testing.T) {
		cImpl, ok := defaultClient.(*Client)
		assert.True(t, ok)

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		customOpts := mockOptionsWithContext{ctx: canceledCtx}

		res, err := cImpl.executeRequest(http.MethodGet, ts.URL, customOpts)
		assert.Error(t, err)
		assert.False(t, res != nil)
	})
}

// mockOptionsWithContext is a helper to test custom context overrides.
type mockOptionsWithContext struct {
	mockOptions
	ctx context.Context
}

// Ctx overrides the base mock to return a custom context.
func (m mockOptionsWithContext) Ctx() context.Context {
	return m.ctx
}

// mockOptions provides a dummy implementation of the [HTTPOptions] interface for testing purposes.
type mockOptions struct{}

// H returns the mocked HTTP headers.
func (m mockOptions) H() []HTTPData { return nil }

// B returns the mocked HTTP body.
func (m mockOptions) B() io.Reader { return nil }

// Q returns the mocked query parameters.
func (m mockOptions) Q() []HTTPData { return nil }

// C returns the mocked HTTP cookies.
func (m mockOptions) C() []*http.Cookie { return nil }

// Ctx returns the mocked [context.Context].
func (m mockOptions) Ctx() context.Context { return nil }

// Auth returns the mocked basic authentication credentials.
func (m mockOptions) Auth() *BasicAuth { return &BasicAuth{"", ""} }

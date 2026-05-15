package xhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestMockClient_Validation(t *testing.T) {
	mock := NewMockClient(
		&http.Response{StatusCode: http.StatusOK},
		nil,
	)
	ctx := context.Background()
	body := "Hello World"
	cookie := &http.Cookie{Name: "sla", Value: "sla234"}
	query := &QueryParam{Key: "page", Value: 1}
	header := &Header{Key: "Content-Type", Value: "application/json"}
	auth := &BasicAuth{
		Username: "admin",
		Password: "password123",
	}
	opts := &Options{
		Headers: []HTTPData{
			header,
		},
		QueryParams: []HTTPData{
			query,
		},
		BasicAuth: auth,
		Body:      bytes.NewBufferString(body),
		Cookies:   []*http.Cookie{cookie},
		Context:   ctx,
	}

	targetURL := "https://api.example.com/v1/users"

	res, err := mock.Post(targetURL, opts)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	call := mock.Calls[0]
	assert.Equal(t, targetURL, call.URL)
	assert.Equal(t, http.MethodPost, call.Method)

	headers := call.Options.H()
	assert.LengthSlice(t, 1, headers)
	assert.Equal(t, header.Key, headers[0].Data().Key)
	assert.Equal(t, header.Value, headers[0].Data().Value)

	v := call.Options.Q()
	assert.LengthSlice(t, 1, v)
	assert.Equal(t, query.Key, v[0].Data().Key)
	assert.Equal(t, query.Value, v[0].Data().Value)

	basicAuth := call.Options.Auth()
	assert.Equal(t, auth.Username, basicAuth.Username)
	assert.Equal(t, auth.Password, basicAuth.Password)

	assert.Equal(t, ctx, call.Options.Ctx())
	b, _ := io.ReadAll(call.Options.B())
	assert.Equal(t, body, string(b))
	cookies := call.Options.C()
	assert.LengthSlice(t, 1, cookies)
	assert.Equal(t, cookie, cookies[0])
}


func TestMockClient_Methods(t *testing.T) {
	mock := NewMockClient(&http.Response{StatusCode: http.StatusOK}, nil)
	targetURL := "https://api.example.com"
	opts := &Options{}

	t.Run("Get method", func(t *testing.T) {
		res, err := mock.Get(targetURL, opts)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
		assert.Equal(t, targetURL, mock.Calls[0].URL)
	})

	t.Run("Put method", func(t *testing.T) {
		_, _ = mock.Put(targetURL, opts)
		assert.Equal(t, http.MethodPut, mock.Calls[1].Method)
	})

	t.Run("Patch method", func(t *testing.T) {
		_, _ = mock.Patch(targetURL, opts)
		assert.Equal(t, http.MethodPatch, mock.Calls[2].Method)
	})

	t.Run("Delete method", func(t *testing.T) {
		_, _ = mock.Delete(targetURL, opts)
		assert.Equal(t, http.MethodDelete, mock.Calls[3].Method)
	})

	t.Run("Options method", func(t *testing.T) {
		_, _ = mock.Options(targetURL, opts)
		assert.Equal(t, http.MethodOptions, mock.Calls[4].Method)
	})

	t.Run("Head method", func(t *testing.T) {
		_, _ = mock.Head(targetURL, opts)
		assert.Equal(t, http.MethodHead, mock.Calls[5].Method)
	})

	t.Run("Connect method", func(t *testing.T) {
		_, _ = mock.Connect(targetURL, opts)
		assert.Equal(t, http.MethodConnect, mock.Calls[6].Method)
	})

	t.Run("Trace method", func(t *testing.T) {
		_, _ = mock.Trace(targetURL, opts)
		assert.Equal(t, http.MethodTrace, mock.Calls[7].Method)
	})
}


func TestMockClient_Error(t *testing.T) {
	expectedErr := http.ErrHandlerTimeout
	mock := NewMockClient(nil, expectedErr)

	_, err := mock.Get("https://api.example.com", nil)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMockClient_MapResponse(t *testing.T) {
	defaultRes := &http.Response{StatusCode: http.StatusOK}
	mock := NewMockClient(defaultRes, nil)
	opts := &Options{}

	mappedURL := "https://api.example.com/v1/users"
	mappedRes := &http.Response{StatusCode: http.StatusCreated}
	mappedErr := http.ErrServerClosed
	mock.MapResponse(http.MethodPost, mappedURL, mappedRes, mappedErr)

	t.Run("returns mapped response for matching method and URL", func(t *testing.T) {
		res, err := mock.Post(mappedURL, opts)
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, mappedErr, err)
	})

	t.Run("falls back to default response when no mapping matches", func(t *testing.T) {
		res, err := mock.Get(mappedURL, opts)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("falls back to default when method differs", func(t *testing.T) {
		res, err := mock.Put(mappedURL, opts)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestMockClient_Clear(t *testing.T) {
	mock := NewMockClient(&http.Response{StatusCode: http.StatusOK}, nil)
	opts := &Options{}

	_, _ = mock.Get("https://api.example.com", opts)
	_, _ = mock.Post("https://api.example.com", opts)
	assert.LengthSlice(t, 2, mock.Calls)

	mock.Clear()
	assert.LengthSlice(t, 0, mock.Calls)
}
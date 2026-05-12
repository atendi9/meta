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

	assert.Equal(t, targetURL, mock.CalledURL)
	assert.Equal(t, http.MethodPost, mock.CalledMethod)

	headers := mock.CalledOptions.H()
	assert.LengthSlice(t, 1, headers)
	assert.Equal(t, header.Key, headers[0].Data().Key)
	assert.Equal(t, header.Value, headers[0].Data().Value)

	v := mock.CalledOptions.Q()
	assert.LengthSlice(t, 1, v)
	assert.Equal(t, query.Key, v[0].Data().Key)
	assert.Equal(t, query.Value, v[0].Data().Value)

	basicAuth := mock.CalledOptions.Auth()
	assert.Equal(t, auth.Username, basicAuth.Username)
	assert.Equal(t, auth.Password, basicAuth.Password)

	assert.Equal(t, ctx, mock.CalledOptions.Ctx())
	b, _ := io.ReadAll(mock.CalledOptions.B())
	assert.Equal(t, body, string(b))
	cookies := mock.CalledOptions.C()
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
		assert.Equal(t, http.MethodGet, mock.CalledMethod)
		assert.Equal(t, targetURL, mock.CalledURL)
	})

	t.Run("Put method", func(t *testing.T) {
		_, _ = mock.Put(targetURL, opts)
		assert.Equal(t, http.MethodPut, mock.CalledMethod)
	})

	t.Run("Patch method", func(t *testing.T) {
		_, _ = mock.Patch(targetURL, opts)
		assert.Equal(t, http.MethodPatch, mock.CalledMethod)
	})

	t.Run("Delete method", func(t *testing.T) {
		_, _ = mock.Delete(targetURL, opts)
		assert.Equal(t, http.MethodDelete, mock.CalledMethod)
	})

	t.Run("Options method", func(t *testing.T) {
		_, _ = mock.Options(targetURL, opts)
		assert.Equal(t, http.MethodOptions, mock.CalledMethod)
	})

	t.Run("Head method", func(t *testing.T) {
		_, _ = mock.Head(targetURL, opts)
		assert.Equal(t, http.MethodHead, mock.CalledMethod)
	})

	t.Run("Connect method", func(t *testing.T) {
		_, _ = mock.Connect(targetURL, opts)
		assert.Equal(t, http.MethodConnect, mock.CalledMethod)
	})

	t.Run("Trace method", func(t *testing.T) {
		_, _ = mock.Trace(targetURL, opts)
		assert.Equal(t, http.MethodTrace, mock.CalledMethod)
	})
}

func TestMockClient_Error(t *testing.T) {
	expectedErr := http.ErrHandlerTimeout
	mock := NewMockClient(nil, expectedErr)

	_, err := mock.Get("https://api.example.com", nil)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
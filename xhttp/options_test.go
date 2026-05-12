package xhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestOptions(t *testing.T) {
	t.Run("Options.H", func(t *testing.T) {
		expectedHeaders := []HTTPData{&Header{Key: "Content-Type", Value: "application/json"}}
		opts := &Options{Headers: expectedHeaders}
		result := opts.H()

		assert.LengthSlice(t, 1, result)
		assert.Equal(t, expectedHeaders[0], result[0])
	})

	t.Run("Options.Q", func(t *testing.T) {
		expectedQueryParams := []HTTPData{&QueryParam{Key: "page", Value: 1}}
		opts := &Options{QueryParams: expectedQueryParams}
		result := opts.Q()

		assert.LengthSlice(t, 1, result)
		assert.Equal(t, expectedQueryParams[0], result[0])
	})

	t.Run("Options.B", func(t *testing.T) {
		expectedContent := "body content"
		expectedBody := bytes.NewBufferString(expectedContent)
		opts := &Options{Body: expectedBody}
		result := opts.B()

		b, err := io.ReadAll(result)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(b))
	})

	t.Run("Options.C", func(t *testing.T) {
		expectedCookies := []*http.Cookie{{Name: "session", Value: "12345"}}
		opts := &Options{Cookies: expectedCookies}
		result := opts.C()

		assert.LengthSlice(t, 1, result)
		assert.Equal(t, expectedCookies[0], result[0])
	})

	t.Run("Options.Ctx", func(t *testing.T) {
		expectedCtx := context.WithValue(context.Background(), "request_id", "abc-123")
		opts := &Options{Context: expectedCtx}
		result := opts.Ctx()

		assert.Equal(t, expectedCtx, result)
	})

	t.Run("Options.Auth", func(t *testing.T) {
		expectedAuth := &BasicAuth{Username: "admin", Password: "password123"}
		opts := &Options{BasicAuth: expectedAuth}
		result := opts.Auth()

		assert.Equal(t, expectedAuth, result)
	})

	t.Run("NewData", func(t *testing.T) {
		d := &Header{}
		NewData(d, "Authorization", "Bearer token")

		assert.Equal(t, "Authorization", d.Key)
		assert.Equal(t, "Bearer token", d.Value)
	})

	t.Run("SetQueryParams", func(t *testing.T) {
		url := "https://api.example.com/users"

		// Test with no params
		resultEmpty := SetQueryParams(url)
		assert.Equal(t, url, resultEmpty)

		// Test with params
		params := []HTTPData{
			&QueryParam{Key: "role", Value: "admin"},
			&QueryParam{Key: "search query", Value: "John Doe"},
		}
		expectedURL := "https://api.example.com/users?role=admin&search+query=John+Doe"
		result := SetQueryParams(url, params...)

		assert.Equal(t, expectedURL, result)
	})

	t.Run("SetHeaders", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
		assert.NoError(t, err)

		headers := []HTTPData{
			&Header{Key: "Accept", Value: "application/json"},
			&Header{Key: "X-Custom-ID", Value: 42},
		}
		SetHeaders(req, headers)

		assert.Equal(t, "application/json", req.Header.Get("Accept"))
		assert.Equal(t, "42", req.Header.Get("X-Custom-ID"))
	})

	t.Run("SetCookies", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
		assert.NoError(t, err)

		cookies := []*http.Cookie{
			{Name: "token", Value: "abcde"},
			{Name: "theme", Value: "dark"},
		}
		SetCookies(req, cookies)

		reqCookies := req.Cookies()
		assert.LengthSlice(t, 2, reqCookies)
		assert.Equal(t, "abcde", reqCookies[0].Value)
		assert.Equal(t, "dark", reqCookies[1].Value)
	})

	t.Run("SetBasicAuth", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://api.example.com", nil)
		assert.NoError(t, err)

		auth := &BasicAuth{Username: "user", Password: "secretpassword"}
		SetBasicAuth(req, auth)

		username, password, ok := req.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", username)
		assert.Equal(t, "secretpassword", password)
	})

	t.Run("Data.Data", func(t *testing.T) {
		d := &Header{Key: "Origin", Value: "https://example.com"}
		result := d.Data()

		assert.Equal(t, "Origin", result.Key)
		assert.Equal(t, "https://example.com", result.Value)
	})

	t.Run("Data.Set", func(t *testing.T) {
		d := &Header{}
		d.Set("Host", "localhost:8080")

		assert.Equal(t, "Host", d.Key)
		assert.Equal(t, "localhost:8080", d.Value)
	})
}

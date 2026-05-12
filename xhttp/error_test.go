package xhttp

import (
	"bytes"
	"errors"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta/xhttp/xjson"
)

func TestError(t *testing.T) {
	t.Run("NewRequestError", func(t *testing.T) {
		originalErr := errors.New("connection failed")

		reqErr := NewRequestError(originalErr)

		buf, ok := reqErr.Payload.(*bytes.Buffer)
		assert.True(t, ok)

		assert.Equal(t, "connection failed", buf.String())

		assert.Equal(t, 0, reqErr.StatusCode)
	})

	t.Run("Error", func(t *testing.T) {
		expectedStr := "{\n  \"message\": \"internal server error\"\n}"

		reqErr := NewRequestError(errors.New(xjson.JSON{"message": "internal server error"}.Buffer().String()))

		result := reqErr.Error()

		assert.Equal(t, expectedStr, result)
	})

	t.Run("JSON", func(t *testing.T) {
		expectedStr := "{\n  \"detail\": \"not found\"\n}"

		reqErr := &RequestError{
			StatusCode: 404,
			Payload:    bytes.NewBufferString(expectedStr),
		}

		jsonResult := reqErr.JSON()

		assert.Equal(t, expectedStr, jsonResult.String())
	})
}

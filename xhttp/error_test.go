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

	t.Run("Error_JSONPayload", func(t *testing.T) {
		expectedStr := "{\n  \"message\": \"internal server error\"\n}"

		reqErr := &RequestError{
			Payload: xjson.JSON{"message": "internal server error"}.Buffer(),
		}

		result := reqErr.Error()

		assert.Equal(t, expectedStr, result)
	})

	t.Run("Error_PreservesTransportMessage", func(t *testing.T) {
		// A plain (non-JSON) transport error message must be preserved
		// verbatim instead of being decoded into an empty object.
		reqErr := NewRequestError(errors.New("connection failed"))

		assert.Equal(t, "connection failed", reqErr.Error())
		// Calling Error again must still return the same message,
		// since the payload reader is restored after reading.
		assert.Equal(t, "connection failed", reqErr.Error())
	})

	t.Run("Error_NilPayload", func(t *testing.T) {
		reqErr := &RequestError{StatusCode: 500}

		assert.Equal(t, "", reqErr.Error())
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

	t.Run("JSON_NilPayload", func(t *testing.T) {
		reqErr := &RequestError{StatusCode: 500}

		jsonResult := reqErr.JSON()

		assert.Equal(t, true, jsonResult == nil)
	})
}

package xhttp

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp/xjson"
)

// RequestError represents an error that occurred during an HTTP request.
// It contains the HTTP status code and a payload holding the error details.
type RequestError struct {
	StatusCode int
	Payload    io.Reader
}

// NewRequestError creates a new [RequestError] from a standard Go error.
// The provided error's message is converted into the RequestError payload.
func NewRequestError(err error) *RequestError {
	return &RequestError{
		Payload: bytes.NewBufferString(err.Error()),
	}
}

// Error implements the standard error interface.
// If the payload holds a valid JSON object, it returns its formatted
// representation. Otherwise (e.g. a transport-level error message), it
// returns the raw payload content so the original message is preserved.
func (err *RequestError) Error() string {
	if err.Payload == nil {
		return ""
	}
	raw, _ := io.ReadAll(err.Payload)
	// Restore the payload so the reader can be consumed again.
	err.Payload = bytes.NewReader(raw)
	var payload xjson.JSON
	if jsonErr := xjson.Decode(bytes.NewReader(raw), &payload); jsonErr == nil {
		return payload.String()
	}
	return string(raw)
}

// JSON reads the error Payload and decodes it into an [xjson.JSON] object.
func (err *RequestError) JSON() (payload xjson.JSON) {
	if err.Payload == nil {
		return
	}
	_ = xjson.Decode(err.Payload, &payload)
	return
}
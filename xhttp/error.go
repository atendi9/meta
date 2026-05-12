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
// It returns the string representation of the decoded JSON payload.
func (err *RequestError) Error() string {
	return err.JSON().String()
}

// JSON reads the error Payload and decodes it into an [xjson.JSON] object.
func (err *RequestError) JSON() (payload xjson.JSON) {
	_ = xjson.Decode(err.Payload, &payload)
	return
}
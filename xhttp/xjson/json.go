// Package xjson provides utility functions and types for handling JSON data.
package xjson

import (
	"bytes"
	"encoding/json"
)

// JSON represents a generic JSON object as a map of strings to arbitrary values.
type JSON map[string]any

// Buffer marshals the JSON object and returns it as a new [bytes.Buffer] pointer.
func (j JSON) Buffer() *bytes.Buffer {
	return bytes.NewBuffer(j.Bytes())
}

// Bytes marshals the JSON object and returns its byte slice representation.
// It silently ignores any marshaling errors.
func (j JSON) Bytes() []byte {
	b, _ := json.Marshal(j)
	return b
}

// Decode reads the JSON object and decodes it into the provided value v.
func (j JSON) Decode(v any) error {
	return Decode(j.Buffer(), v)
}

// String marshals the JSON object with indentation, returning it as a formatted string.
func (j JSON) String() string {
	b, _ := json.MarshalIndent(j, "", "  ")
	return string(b)
}

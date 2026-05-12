package xjson

import (
	"encoding/json"
	"io"
)

// Decode reads JSON-encoded data from the provided [io.Reader] and stores it in the value pointed to by v.
func Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
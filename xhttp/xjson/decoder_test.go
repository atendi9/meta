package xjson

import (
	"bytes"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestDecoder(t *testing.T) {
	var payload JSON
	buf := bytes.NewBufferString("nojson")
	err := Decode(buf, &payload)
	assert.Error(t, err)
	buf = JSON{"message": "Hello World"}.Buffer()
	err = Decode(buf, &payload)
	assert.NoError(t, err)
}

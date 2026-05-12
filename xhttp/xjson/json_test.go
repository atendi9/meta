package xjson

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestJSON_String(t *testing.T) {
	j := JSON{
		"name": "test",
		"id":   1,
	}

	expected := "{\n  \"id\": 1,\n  \"name\": \"test\"\n}"
	result := j.String()

	assert.Equal(t, expected, result)
}

func TestJSON_Bytes(t *testing.T) {
	j := JSON{"key": "value"}

	expected := []byte(`{"key":"value"}`)
	result := j.Bytes()

	assert.Equal(t, string(expected), string(result))
}

func TestJSON_Buffer(t *testing.T) {
	j := JSON{"active": true}

	expected := `{"active":true}`
	result := j.Buffer().String()

	assert.Equal(t, expected, result)
}

func TestJSON_Decode(t *testing.T) {
	j := JSON{"score": 100}

	type payload struct {
		Score int `json:"score"`
	}

	var result payload
	err := j.Decode(&result)

	if err != nil {
		t.Fatalf("unexpected error during decode: %v", err)
	}

	expected := payload{Score: 100}

	assert.Equal(t, expected, result)
}

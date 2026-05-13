package uuid

import (
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestNewV7_RFCCompliance(t *testing.T) {
	id, err := NewV7()
	assert.NoError(t, err)
	t.Run("Verify Version (Bits 4-7 of byte 6 must be 0111 / 7)", func(t *testing.T) {
		version := id[6] >> 4
		assert.Equal(t, byte(7), version)
	})
	t.Run("Verify Variant (Bits 6-7 of byte 8 must be 10 / 2)", func(t *testing.T) {
		variant := id[8] >> 6
		assert.Equal(t, byte(2), variant)
	})
}

func TestNewV7_Timestamp(t *testing.T) {
	before := time.Now().UnixMilli()

	id, err := NewV7()
	assert.NoError(t, err)

	after := time.Now().UnixMilli()

	var extractedTime int64
	extractedTime |= int64(id[0]) << 40
	extractedTime |= int64(id[1]) << 32
	extractedTime |= int64(id[2]) << 24
	extractedTime |= int64(id[3]) << 16
	extractedTime |= int64(id[4]) << 8
	extractedTime |= int64(id[5])

	assert.True(t, extractedTime >= before)
	assert.True(t, extractedTime <= after)
}

func TestNewV7_StringFormat(t *testing.T) {
	id, err := NewV7()
	assert.NoError(t, err)

	str := id.String()

	assert.Equal(t, 36, len(str))
	assert.Equal(t, byte('-'), str[8])
	assert.Equal(t, byte('-'), str[13])
	assert.Equal(t, byte('-'), str[18])
	assert.Equal(t, byte('-'), str[23])
}

func TestNewV7_Uniqueness(t *testing.T) {
	iterations := 10000
	generated := make(map[string]any, iterations)

	for range iterations {
		id, err := NewV7()
		assert.NoError(t, err)

		str := id.String()
		_, exists := generated[str]

		assert.False(t, exists)
		generated[str] = struct{}{}
	}

	assert.LengthMap(t, iterations, generated)
}

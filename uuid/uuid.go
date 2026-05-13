package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// UUID represents a 128-bit Universally Unique Identifier.
type UUID [16]byte

// String returns the string representation of the UUID in the standard 8-4-4-4-12 format.
func (u UUID) String() string {
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

// NewV7 generates a new UUID version 7 as defined by RFC 9562.
func NewV7() (UUID, error) {
	var uuid UUID

	now := time.Now().UnixMilli()

	uuid[0] = byte(now >> 40)
	uuid[1] = byte(now >> 32)
	uuid[2] = byte(now >> 24)
	uuid[3] = byte(now >> 16)
	uuid[4] = byte(now >> 8)
	uuid[5] = byte(now)

	_, err := rand.Read(uuid[6:])
	if err != nil {
		return uuid, err
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x70
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return uuid, nil
}

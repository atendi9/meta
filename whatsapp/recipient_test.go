package whatsapp

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta/xhttp/xjson"
)

func TestIsBSUID(t *testing.T) {
	t.Run("should recognize a standard BSUID", func(t *testing.T) {
		assert.True(t, IsBSUID("US.13491208655302741918"))
	})

	t.Run("should recognize a Brazilian BSUID", func(t *testing.T) {
		assert.True(t, IsBSUID("BR.5511999999999"))
	})

	t.Run("should recognize a parent BSUID", func(t *testing.T) {
		assert.True(t, IsBSUID("US.ENT.11815799212886844830"))
	})

	t.Run("should reject a plain phone number", func(t *testing.T) {
		assert.False(t, IsBSUID("5511999999999"))
	})

	t.Run("should reject a phone number with a plus prefix", func(t *testing.T) {
		assert.False(t, IsBSUID("+16505551234"))
	})

	t.Run("should reject an empty string", func(t *testing.T) {
		assert.False(t, IsBSUID(""))
	})

	t.Run("should reject a single-letter country code", func(t *testing.T) {
		assert.False(t, IsBSUID("U.13491208655302741918"))
	})

	t.Run("should reject a country code without an identifier", func(t *testing.T) {
		assert.False(t, IsBSUID("US."))
	})
}

func TestSetRecipient(t *testing.T) {
	t.Run("should route a phone number to the to field", func(t *testing.T) {
		payload := xjson.JSON{}
		SetRecipient(payload, "5511999999999")
		assert.Equal(t, "5511999999999", payload["to"])
		assert.False(t, payload["recipient"] != nil)
	})

	t.Run("should route a BSUID to the recipient field", func(t *testing.T) {
		payload := xjson.JSON{}
		SetRecipient(payload, "US.13491208655302741918")
		assert.Equal(t, "US.13491208655302741918", payload["recipient"])
		assert.False(t, payload["to"] != nil)
	})

	t.Run("should route a parent BSUID to the recipient field", func(t *testing.T) {
		payload := xjson.JSON{}
		SetRecipient(payload, "US.ENT.11815799212886844830")
		assert.Equal(t, "US.ENT.11815799212886844830", payload["recipient"])
		assert.False(t, payload["to"] != nil)
	})
}

func TestMessageHeaderRecipientRouting(t *testing.T) {
	t.Run("should keep using to for phone numbers", func(t *testing.T) {
		h := MessageHeader("5511999999999", "text")
		assert.Equal(t, "5511999999999", h["to"])
		assert.False(t, h["recipient"] != nil)
	})

	t.Run("should use recipient for a BSUID", func(t *testing.T) {
		h := MessageHeader("US.13491208655302741918", "text")
		assert.Equal(t, "US.13491208655302741918", h["recipient"])
		assert.False(t, h["to"] != nil)
	})
}

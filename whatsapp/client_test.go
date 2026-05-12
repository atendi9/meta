package whatsapp

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
)

func TestClient(t *testing.T) {
	senderID := "123456789"
	token := "access_token_123"

	t.Run("Default", func(t *testing.T) {
		client := Default(senderID, token)

		assert.Equal(t, senderID, client.senderID)
		assert.Equal(t, token, client.AccessToken)
		assert.Equal(t, meta.GraphAPIVersion, client.ApiVersion)
		assert.Equal(t, BaseUrl, client.BaseUrl)
	})

	t.Run("New", func(t *testing.T) {
		ctx := context.TODO()
		version := "v20.0"
		client := New(ctx, version, token, senderID)

		assert.Equal(t, senderID, client.senderID)
		assert.Equal(t, version, client.ApiVersion)
		assert.Equal(t, token, client.AccessToken)
	})

	t.Run("SenderID", func(t *testing.T) {
		client := Default(senderID, token)
		result := client.SenderID()

		assert.Equal(t, senderID, result)
	})

	t.Run("ChangeSenderId", func(t *testing.T) {
		newID := "987654321"
		client := Default(senderID, token)
		client.ChangeSenderId(newID)

		assert.Equal(t, newID, client.senderID)
	})

	t.Run("Endpoint", func(t *testing.T) {
		client := Default(senderID, token)
		endpoint := "messages"
		expected := fmt.Sprintf("%s/%s/%s", BaseUrl, meta.GraphAPIVersion, endpoint)

		result := client.Endpoint(endpoint)

		assert.Equal(t, expected, result)
	})

	t.Run("GenerateWhatsappLink", func(t *testing.T) {
		phone := "5511999999999"
		text := "Hello World!"
		expected := fmt.Sprintf("https://api.whatsapp.com/send?phone=%s&text=%s", phone, url.QueryEscape(text))

		result := GenerateWhatsappLink(phone, text)

		assert.Equal(t, expected, result)
	})

	t.Run("Headers", func(t *testing.T) {
		client := Default(senderID, token)
		contentType := "application/json"
		headers := client.Headers(contentType)

		assert.LengthSlice(t, 2, headers)
		expectedAuth := fmt.Sprintf("Bearer %s", token)

		assert.True(t, headers != nil)
		assert.Equal(t, expectedAuth, fmt.Sprintf("Bearer %s", client.AccessToken))
	})
}

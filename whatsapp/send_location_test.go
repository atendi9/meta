package whatsapp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestSendLocation(t *testing.T) {
	mockHTTP := xhttp.NewMockClient(nil, nil)
	client := &Client{
		senderID: "12345",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockHTTP,
			BaseUrl:    "https://graph.facebook.com",
			ApiVersion: "v20.0",
		},
	}

	googlePlex := Location{
		Latitude:  37.422,
		Longitude: -122.084,
		Name:      "Googleplex",
		Address:   "1600 Amphitheatre Pkwy, Mountain View, CA 94043",
	}

	t.Run("SuccessfulLocationSend", func(t *testing.T) {
		mockJSON := `{"messages": [{"id": "wa_id_googleplex_123"}]}`
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(mockJSON)),
		}

		h := MessageHeader("5511999999999", "location")
		id, err := client.SendLocation(h, googlePlex)

		assert.NoError(t, err)
		assert.Equal(t, "wa_id_googleplex_123", id)

		assert.Equal(t, "location", h["type"])
		locData, ok := h["location"].(Location)
		assert.True(t, ok)
		assert.Equal(t, googlePlex.Name, locData.Name)
		assert.Equal(t, googlePlex.Latitude, locData.Latitude)
		assert.Equal(t, googlePlex.Longitude, locData.Longitude)
		assert.Equal(t, googlePlex.Address, locData.Address)
	})

	t.Run("APIErrorHandling", func(t *testing.T) {
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(strings.NewReader(`{"error": {"message": "invalid parameters"}}`)),
		}

		h := MessageHeader("5511999999999", "location")
		id, err := client.SendLocation(h, googlePlex)

		assert.Error(t, err)
		assert.Equal(t, "", id)
	})

	t.Run("EmptyResponseValidation", func(t *testing.T) {
		mockHTTP.DefaultResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"messages": []}`)),
		}

		h := MessageHeader("5511999999999", "location")
		_, err := client.SendLocation(h, googlePlex)
		assert.Error(t, err)
	})
}

package whatsapp

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAuthTemplate(t *testing.T) {
	expected := "{\n  \"messaging_product\": \"whatsapp\",\n  \"recipient_type\": \"individual\",\n  \"template\": {\n    \"components\": [\n      {\n        \"parameters\": [\n          {\n            \"text\": \"api_token\",\n            \"type\": \"text\"\n          }\n        ],\n        \"type\": \"body\"\n      },\n      {\n        \"index\": 0,\n        \"parameters\": [\n          {\n            \"text\": \"api_token\",\n            \"type\": \"text\"\n          }\n        ],\n        \"sub_type\": \"url\",\n        \"type\": \"button\"\n      }\n    ],\n    \"language\": {\n      \"code\": \"pt_BR\"\n    },\n    \"name\": \"auth_test\"\n  },\n  \"to\": \"55819999999\",\n  \"type\": \"template\"\n}"
	tmpl := AuthTemplate(
		"auth_test",
		"55819999999",
		"api_token",
		PortugueseBrazil,
	)
	assert.Equal(t, expected, tmpl.String())
}

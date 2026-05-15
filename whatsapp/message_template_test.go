package whatsapp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

func TestMessageTemplate(t *testing.T) {
	h := Header{"from": "123456789"}
	template := xjson.JSON{
		"name": "hello_world",
		"language": xjson.JSON{
			"code": "en_US",
		},
	}

	result := MessageTemplate(h, template)

	assert.Equal(t, "template", result["type"])
	templateContent, ok := result["template"].(xjson.JSON)
	assert.True(t, ok)
	assert.Equal(t, template.String(), templateContent.String())
	assert.LengthMap(t, 3, result)
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[{"id": "wamid_39178d1uh83"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}
	res, err := SendMessage(api, result)
	assert.NoError(t, err)
	assert.Equal(t, "wamid_39178d1uh83", res.FirstId())
}

func TestCreateMessageTemplate_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "987654321"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	fields := TemplateFields{
		Name:     "seasonal_promotion",
		Category: MARKETING,
		Lang:     English,
		Components: []xjson.JSON{
			{
				"type":   "HEADER",
				"format": "TEXT",
				"text":   "Summer Sale!",
			},
			{
				"type": "BODY",
				"text": "Hello {{1}}, shop now and get 20% off!",
				"example": xjson.JSON{
					"body_text": [][]string{{"John"}},
				},
			},
			{
				"type": "FOOTER",
				"text": "Terms and conditions apply.",
			},
		},
	}

	err := CreateMessageTemplate(api, fields)
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)

	expectedURL := api.Endpoint(api.senderID + "/message_templates")
	assert.Equal(t, expectedURL, mockClient.Calls[0].URL)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestCreateMessageTemplate_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fields := TemplateFields{
		Name:     "auth_code",
		Category: AUTHENTICATION,
		Lang:     PortugueseBrazil,
		Components: []xjson.JSON{
			{
				"type":       "BODY",
				"add_safety": true,
			},
			{
				"type": "BUTTONS",
				"buttons": []xjson.JSON{
					{
						"type":     "OTP",
						"otp_type": "COPY_CODE",
						"text":     "Copy Code",
					},
				},
			},
		},
	}

	err := CreateMessageTemplate(api, fields)
	assert.Error(t, err)
}

func TestCreateMessageTemplate_EmptyBody(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fields := TemplateFields{
		Name:     "simple_utility",
		Category: UTILITY,
		Lang:     EnglishUK,
		Components: []xjson.JSON{
			{
				"type": "BODY",
				"text": "Your order has been shipped.",
			},
		},
	}

	err := CreateMessageTemplate(api, fields)
	assert.NoError(t, err)
}

// errorReadCloser is an [io.ReadCloser] whose Read always fails, used to
// exercise the io.ReadAll error path when consuming an HTTP response body.
type errorReadCloser struct{}

func (errorReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("body read failure")
}

func (errorReadCloser) Close() error { return nil }

func TestCreateMessageTemplate_BodyReadError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       errorReadCloser{},
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fields := TemplateFields{
		Name:     "simple_utility",
		Category: UTILITY,
		Lang:     EnglishUK,
		Components: []xjson.JSON{
			{
				"type": "BODY",
				"text": "Your order has been shipped.",
			},
		},
	}

	err := CreateMessageTemplate(api, fields)
	assert.Error(t, err)
}

func TestDeleteMessageTemplate_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	templateName := "old_promotion_template"
	err := DeleteMessageTemplate(api, templateName)

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodDelete, mockClient.Calls[0].Method)

	expectedURL := api.Endpoint(api.senderID + "/message_templates")
	assert.Equal(t, expectedURL, mockClient.Calls[0].URL)
}

func TestDeleteMessageTemplate_Error(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrClosedPipe)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	err := DeleteMessageTemplate(api, "any_template")
	assert.Error(t, err)
}

func TestTemplateStatusFuncs(t *testing.T) {
	assert.Equal(t, status("PENDING"), Pending())
	assert.Equal(t, status("REJECTED"), Rejected())
	assert.Equal(t, status("APPROVED"), Aproved())
}

func TestGetTemplates_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": [{"id": "1", "name": "t1", "status": "APPROVED"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
		},
	}

	res, err := GetTemplates[TemplateStatus](api, nil)
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, res.Data)
	assert.Equal(t, "1", res.Data[0].Id)
	assert.Equal(t, "t1", res.Data[0].Name)
	assert.Equal(t, Aproved(), res.Data[0].Status)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodGet, mockClient.Calls[0].Method)
}

func TestGetTemplates_HTTPError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	res, err := GetTemplates[TemplateStatus](api, nil)
	assert.Error(t, err)
	assert.LengthSlice(t, 0, res.Data)
}

func TestGetTemplates_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{invalid json`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
		},
	}

	res, err := GetTemplates[TemplateStatus](api, nil)
	assert.Error(t, err)
	assert.LengthSlice(t, 0, res.Data)
}

func TestGetTemplateStatus_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": [{"id": "1", "name": "hello_world", "status": "PENDING"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
		},
	}

	res, err := GetTemplateStatus(api, "hello_world")
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, res.Data)
	assert.Equal(t, "hello_world", res.Data[0].Name)
	assert.Equal(t, Pending(), res.Data[0].Status)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodGet, mockClient.Calls[0].Method)
}

func TestGetJSONTemplates_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data": [{"name": "template_json", "language": "en_US"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "123",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
		},
	}

	res, err := GetJSONTemplates(api)
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, res.Data)
	assert.Equal(t, "template_json", res.Data[0]["name"])
	assert.Equal(t, "en_US", res.Data[0]["language"])
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodGet, mockClient.Calls[0].Method)
}
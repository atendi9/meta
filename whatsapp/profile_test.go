package whatsapp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestProfile_Info_Success(t *testing.T) {
	responseBody := `{
		"data": [
			{
				"about": "Tech Solutions",
				"email": "contact@tech.com",
				"vertical": "PROF_SERVICES",
				"websites": ["https://tech.com"]
			}
		]
	}`

	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	profile := &Profile{
		PhoneNumberID: "123456789",
		client:        api,
	}

	result, err := profile.Info("My Business Account")

	assert.NoError(t, err)
	assert.Equal(t, "Tech Solutions", result.About)
	assert.Equal(t, "contact@tech.com", result.Email)
	assert.Equal(t, "PROF_SERVICES", result.Vertical)
	assert.Equal(t, "My Business Account", result.Name)
	assert.LengthSlice(t, 1, result.Websites)

	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodGet, mockClient.Calls[0].Method)
}

func TestProfile_Info_Error(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	profile := &Profile{
		PhoneNumberID: "123456789",
		client:        api,
	}

	result, err := profile.Info("My Business Account")

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprint(InfoData{}), fmt.Sprint(result))
}

func TestProfile_ChangeAbout_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	profile := &Profile{
		PhoneNumberID: "123456789",
		client:        api,
	}

	err := profile.ChangeAbout("We provide the best services.")

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestProfile_ChangeWebsites_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	profile := &Profile{
		PhoneNumberID: "987654321",
		client:        api,
	}

	websites := []string{"https://example.com", "https://shop.example.com"}
	err := profile.ChangeWebsites(websites)

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func TestProfile_ChangeVertical_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	profile := &Profile{
		PhoneNumberID: "555555555",
		client:        api,
	}

	err := profile.ChangeVertical(FINANCE)

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func TestProfile_ChangeEmail_Error(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrClosedPipe)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	profile := &Profile{
		PhoneNumberID: "111222333",
		client:        api,
	}

	err := profile.ChangeEmail("invalid_email")

	assert.Error(t, err)
}

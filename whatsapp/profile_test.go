package whatsapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

// newProfileWithMock builds a [Profile] whose underlying [Client] uses the
// provided mocked HTTP client, allowing the profile endpoints to be exercised
// without performing real network calls.
func newProfileWithMock(mock xhttp.HTTPClient) *Profile {
	return &Profile{
		PhoneNumberID: "1234567890",
		client: &Client{
			senderID: "1234567890",
			GraphAPIClient: meta.GraphAPIClient{
				HttpClient:  mock,
				ApiVersion:  "v24.0",
				BaseUrl:     "https://graph.facebook.com",
				AccessToken: "valid_token",
			},
		},
	}
}

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

func TestNewProfile(t *testing.T) {
	p := NewProfile(context.Background(), "1234567890", "token")

	assert.Equal(t, "1234567890", p.PhoneNumberID)
	assert.NotNil(t, p.client)
}

func TestProfile_Info_EmptyData(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data":[]}`)),
	}
	p := newProfileWithMock(xhttp.NewMockClient(mockRes, nil))

	result, err := p.Info("Acme")

	assert.NoError(t, err)
	assert.Equal(t, "", result.Name)
}

func TestProfile_Info_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`not-json`)),
	}
	p := newProfileWithMock(xhttp.NewMockClient(mockRes, nil))

	_, err := p.Info("Acme")

	assert.Error(t, err)
}

func TestProfile_ChangeDescription_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success":true}`)),
	}
	p := newProfileWithMock(xhttp.NewMockClient(mockRes, nil))

	err := p.ChangeDescription("new description")

	assert.NoError(t, err)
}

func TestProfile_ChangeAddress_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success":true}`)),
	}
	p := newProfileWithMock(xhttp.NewMockClient(mockRes, nil))

	err := p.ChangeAddress("Rua das Flores, 123")

	assert.NoError(t, err)
}

func TestProfile_ChangeProfilePicture_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       newReusableBody([]byte(`{"id":"session_123","h":"handle_abc"}`)),
	}
	p := newProfileWithMock(xhttp.NewMockClient(mockRes, nil))

	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x01}
	fileHeader := createMockFileHeader(t, "logo.png", pngBytes)

	data, err := p.ChangeProfilePicture("app_123", fileHeader)

	assert.NoError(t, err)
	assert.Equal(t, "logo.png", data.Name)
	assert.Equal(t, int64(len(pngBytes)), data.Size)
}

func TestProfile_ChangeProfilePicture_UploadError(t *testing.T) {
	p := newProfileWithMock(xhttp.NewMockClient(nil, errors.New("upload failed")))

	fileHeader := createMockFileHeader(t, "logo.png", []byte("content"))

	_, err := p.ChangeProfilePicture("app_123", fileHeader)

	assert.Error(t, err)
}

func TestNewProfileVertical(t *testing.T) {
	assert.Equal(t, RETAIL, NewProfileVertical("RETAIL"))
	assert.Equal(t, HEALTH, NewProfileVertical("HEALTH"))
	assert.Equal(t, OTHER, NewProfileVertical("NOT_A_VERTICAL"))
}

func TestIsValidProfileVertical(t *testing.T) {
	valid := []string{
		"ALCOHOL", "APPAREL", "AUTO", "BEAUTY", "EDU", "ENTERTAIN",
		"EVENT_PLAN", "FINANCE", "GOVT", "GROCERY", "HEALTH", "HOTEL",
		"NONPROFIT", "ONLINE_GAMBLING", "OTC_DRUGS", "OTHER",
		"PHYSICAL_GAMBLING", "PROF_SERVICES", "RESTAURANT", "RETAIL", "TRAVEL",
	}
	for _, v := range valid {
		assert.True(t, IsValidProfileVertical(v))
	}

	assert.False(t, IsValidProfileVertical("INVALID"))
	assert.False(t, IsValidProfileVertical(""))
}

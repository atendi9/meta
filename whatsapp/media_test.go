package whatsapp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestIsValidMediaUploadType(t *testing.T) {
	validTypes := []string{
		"image/jpeg", "image/png", "image/webp",
		"video/mp4", "video/3gpp",
		"audio/aac", "audio/mpeg", "audio/ogg",
		"text/plain", "application/pdf",
	}
	for _, mime := range validTypes {
		assert.True(t, IsValidMediaUploadType(mime))
	}

	invalidTypes := []string{
		"application/json",
		"text/html",
		"image/gif",
		"invalid/type",
	}
	for _, mime := range invalidTypes {
		assert.False(t, IsValidMediaUploadType(mime))
	}
}

func TestGetMediaType(t *testing.T) {
	assert.Equal(t, "image", getMediaType("image/jpeg"))
	assert.Equal(t, "video", getMediaType("video/mp4"))
	assert.Equal(t, "audio", getMediaType("audio/ogg"))
	assert.Equal(t, "document", getMediaType("application/pdf"))
	assert.Equal(t, "document", getMediaType("text/plain"))
}

func TestMultipartWriter(t *testing.T) {
	var body bytes.Buffer
	writer := NewMultipartWriter(&body)

	err := writer.WriteField("messaging_product", "whatsapp")
	assert.NoError(t, err)

	header := textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="file"; filename="test.txt"`},
		"Content-Type":        []string{"text/plain"},
	}
	part, err := writer.CreatePart(header)
	assert.NoError(t, err)

	_, err = part.Write([]byte("fake file content"))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	assert.True(t, writer.Value() != nil)
	assert.True(t, len(body.Bytes()) > 0)
}

func TestFileWriter(t *testing.T) {
	api := &Client{}
	var body bytes.Buffer
	fileContent := bytes.NewReader([]byte("dummy content"))

	writer, err := api.fileWriter(&body, fileContent, "text/plain", "dummy.txt")
	assert.NoError(t, err)
	assert.True(t, writer != nil)
}

func TestGenerateMediaID_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "media_123456"}`)),
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

	fileContent := bytes.NewReader([]byte("image content"))
	id, err := api.GenerateMediaID("image/png", "test.png", fileContent)

	assert.NoError(t, err)
	assert.Equal(t, "media_123456", id)
	assert.LengthSlice(t, 1, mockClient.Calls)

	expectedURL := api.Endpoint(api.senderID + "/media")
	assert.Equal(t, expectedURL, mockClient.Calls[0].URL)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestGenerateMediaID_InvalidMimeType(t *testing.T) {
	api := &Client{}
	fileContent := bytes.NewReader([]byte("content"))

	id, err := api.GenerateMediaID("application/json", "test.json", fileContent)

	assert.Error(t, err)
	assert.Equal(t, "", id)
	assert.Equal(t, ErrInvalidMimeType, err)
}

func TestGenerateMediaID_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, errors.New("network error"))

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fileContent := bytes.NewReader([]byte("video content"))
	id, err := api.GenerateMediaID("video/mp4", "video.mp4", fileContent)

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

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

func TestIsValidMediaUploadType_CSV(t *testing.T) {
	// text/csv was missing from the allowlist before the MIME fix.
	assert.True(t, IsValidMediaUploadType("text/csv"))
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

// TestGenerateMediaID_NormalizesMimeType covers the media upload MIME bug:
// generic, mobile-specific, and transcoder-mangled MIME types must be
// normalized to a Meta-accepted type instead of being rejected outright.
func TestGenerateMediaID_NormalizesMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		filePath string
		content  []byte
	}{
		{"octet-stream pdf by extension", "application/octet-stream", "doc.pdf", []byte("%PDF-1.7")},
		{"octet-stream pdf by content", "application/octet-stream", "blob", []byte("%PDF-1.7")},
		{"mobile m4a audio", "audio/x-m4a", "voice.m4a", []byte("audio bytes")},
		{"mobile opus audio", "audio/opus", "voice.opus", []byte("audio bytes")},
		{"ffmpeg webm audio", "audio/webm", "out", []byte("audio bytes")},
		{"3gpp audio", "audio/3gpp", "voice.3gp", []byte("audio bytes")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRes := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id": "media_ok"}`)),
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

			id, err := api.GenerateMediaID(tc.mimeType, tc.filePath, bytes.NewReader(tc.content))

			assert.NoError(t, err)
			assert.Equal(t, "media_ok", id)
		})
	}
}

func TestGenerateMediaID_ReadError(t *testing.T) {
	api := &Client{senderID: "10987654321"}

	id, err := api.GenerateMediaID("image/png", "test.png", errReader{})

	assert.Error(t, err)
	assert.Equal(t, "", id)
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

func TestGenerateMediaID_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`not-json`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fileContent := bytes.NewReader([]byte("image content"))
	id, err := api.GenerateMediaID("image/png", "test.png", fileContent)

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

func TestGenerateMediaID_FileWriterError(t *testing.T) {
	api := &Client{senderID: "10987654321"}

	_, err := api.GenerateMediaID("image/png", "test.png", errReader{})

	assert.Error(t, err)
}

// errReader is an [io.Reader] that always fails, used to exercise the
// io.Copy error path inside fileWriter.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failure")
}

// failWriter is an [io.Writer] that fails on the Nth Write call (1-indexed),
// allowing prior calls through. It is used to exercise the multipart writer
// error branches in fileWriter at precise points.
type failWriter struct {
	calls  int
	failOn int
}

func (w *failWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls >= w.failOn {
		return 0, errors.New("write failure")
	}
	return len(p), nil
}

func TestFileWriter_CopyError(t *testing.T) {
	api := &Client{}
	var body bytes.Buffer

	_, err := api.fileWriter(&body, errReader{}, "text/plain", "dummy.txt")

	assert.Error(t, err)
}

func TestFileWriter_CreatePartError(t *testing.T) {
	api := &Client{}

	_, err := api.fileWriter(&failWriter{failOn: 1}, bytes.NewReader([]byte("data")), "text/plain", "dummy.txt")

	assert.Error(t, err)
}

func TestFileWriter_WriteFieldErrors(t *testing.T) {
	api := &Client{}
	fileContent := []byte("dummy content")

	// Each subsequent Write call corresponds to a later stage of the multipart
	// payload: file part, first WriteField, second WriteField, and Close.
	for _, failOn := range []int{2, 3, 4, 5, 6, 7} {
		_, err := api.fileWriter(
			&failWriter{failOn: failOn},
			bytes.NewReader(fileContent),
			"text/plain",
			"dummy.txt",
		)
		assert.Error(t, err)
	}
}

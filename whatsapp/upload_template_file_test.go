package whatsapp

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

// reusableBody represents an [io.ReadCloser] that resets its buffer
// upon closing, allowing it to be read multiple times across different HTTP requests.
type reusableBody struct {
	data []byte
	io.Reader
}

// newReusableBody creates a new instance of [reusableBody] with the given data.
func newReusableBody(data []byte) io.ReadCloser {
	return &reusableBody{
		data:   data,
		Reader: bytes.NewReader(data),
	}
}

// Close resets the reader back to the original byte slice and returns nil.
func (r *reusableBody) Close() error {
	r.Reader = bytes.NewReader(r.data)
	return nil
}

func TestUploadTemplateFile_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       newReusableBody([]byte(`{"id": "session_123", "h": "hash_abc"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fileContent := []byte("Hello World")
	fileHeader := createMockFileHeader(t, "document.txt", fileContent)

	result, err := UploadTemplateFile(api, "app_123", fileHeader)
	assert.NoError(t, err)
	assert.Equal(t, "hash_abc", result.Id)
	assert.Equal(t, "session_123", result.Session)

	assert.LengthSlice(t, 2, mockClient.Calls)
}

func TestUploadTemplateFile_FileOpenError(t *testing.T) {
	api := &Client{}
	invalidHeader := &multipart.FileHeader{}

	_, err := UploadTemplateFile(api, "app_123", invalidHeader)

	assert.Error(t, err)
}

func TestStartUploadSession_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "session_12345"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	session, err := api.startUploadSession("app_123", "test.pdf", 1024, "application/pdf")

	assert.NoError(t, err)
	assert.Equal(t, "session_12345", session.Id)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func TestStartUploadSession_Error(t *testing.T) {
	expectedErr := errors.New("network timeout")
	mockClient := xhttp.NewMockClient(nil, expectedErr)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	_, err := api.startUploadSession("app_123", "test.pdf", 1024, "application/pdf")

	assert.Error(t, err)
}

func TestGenerateFileHandle_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"h": "hash_xyz789"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fileContent := []byte("testandooo")
	session := UploadSession{Id: "session_123"}

	handle, err := api.generateFileHandle(session, fileContent, "test.txt")

	assert.NoError(t, err)
	assert.Equal(t, "hash_xyz789", handle.H)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func createMockFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(t, err)

	_, err = part.Write(content)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(CalculateMaxMemory(len(content)))
	assert.NoError(t, err)

	_, header, err := req.FormFile("file")
	assert.NoError(t, err)

	return header
}

// Megabyte represents a single megabyte in bytes.
const Megabyte int64 = 1 << 20

// DefaultMaxMemory represents the fallback memory limit (10 MB).
const DefaultMaxMemory int64 = 10 * Megabyte

// CalculateMaxMemory determines the optimal memory allocation for a [net/http.Request].
// It receives the byte length as an [int] and returns the corresponding [int64] value,
// adding a 5% safety buffer for multipart headers and boundaries.
// If the provided length is zero or negative, it safely defaults to 10 MB.
func CalculateMaxMemory(byteLength int) int64 {
	if byteLength <= 0 {
		return DefaultMaxMemory
	}

	buffer := float64(byteLength) * 0.05

	return int64(byteLength) + int64(buffer)
}

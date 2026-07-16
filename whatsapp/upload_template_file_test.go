package whatsapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

// headerValue returns the value of the named header recorded in the call, or an
// empty string when the call carries no such header.
func headerValue(call xhttp.Call, key string) string {
	for _, header := range call.Options.H() {
		if data := header.Data(); data.Key == key {
			if value, ok := data.Value.(string); ok {
				return value
			}
			return fmt.Sprint(data.Value)
		}
	}
	return ""
}

// queryParamValue returns the value of the named query param recorded in the
// call, or an empty string when the call carries no such param.
func queryParamValue(call xhttp.Call, key string) string {
	for _, param := range call.Options.Q() {
		if data := param.Data(); data.Key == key {
			return fmt.Sprint(data.Value)
		}
	}
	return ""
}

// callBody drains and returns the request body recorded in the call.
func callBody(t *testing.T, call xhttp.Call) []byte {
	t.Helper()
	body, err := io.ReadAll(call.Options.B())
	assert.NoError(t, err)
	return body
}

// pngBytes is a minimal PNG: the 8-byte signature followed by an IHDR chunk.
var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

// jpegBytes is a minimal JPEG: the SOI marker followed by an APP0/JFIF segment.
var jpegBytes = []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00")

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

	session := UploadSession{Id: "session_123"}

	handle, err := api.generateFileHandle(session, jpegBytes, "image/jpeg")

	assert.NoError(t, err)
	assert.Equal(t, "hash_xyz789", handle.H)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

// TestGenerateFileHandle_SendsRawBody covers the contract of Meta's Resumable
// Upload API: the body is the file's bytes and nothing else. Wrapping them in a
// multipart envelope still yields a handle, because Meta does not inspect the
// content at this step, and every later use of that handle fails.
func TestGenerateFileHandle_SendsRawBody(t *testing.T) {
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

	session := UploadSession{Id: "session_123"}

	_, err := api.generateFileHandle(session, jpegBytes, "image/jpeg")
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)

	call := mockClient.Calls[0]
	body := callBody(t, call)
	assert.Equal(t, true, bytes.Equal(jpegBytes, body))

	contentType := headerValue(call, "Content-Type")
	assert.Equal(t, "image/jpeg", contentType)
	assert.Equal(t, false, strings.Contains(contentType, "multipart/"))
	assert.Equal(t, "0", headerValue(call, "file_offset"))
}

// TestGenerateFileHandle_PreservesRealMimeType guards against announcing a
// class default instead of the file's own type: with a raw body the
// Content-Type is the only type Meta sees, so labeling a PNG as image/jpeg
// makes it reject the handle later.
func TestGenerateFileHandle_PreservesRealMimeType(t *testing.T) {
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

	session := UploadSession{Id: "session_123"}

	_, err := api.generateFileHandle(session, pngBytes, "image/png")
	assert.NoError(t, err)

	assert.Equal(t, "image/png", headerValue(mockClient.Calls[0], "Content-Type"))
}

func TestStartUploadSession_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`not-json`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	_, err := api.startUploadSession("app_123", "test.pdf", 1024, "application/pdf")

	assert.Error(t, err)
}

func TestGenerateFileHandle_InvalidFile(t *testing.T) {
	mockClient := xhttp.NewMockClient(&http.Response{StatusCode: http.StatusOK}, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	// Bytes that resolve to no accepted MIME type are rejected before any
	// request is made.
	fileContent := []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	session := UploadSession{Id: "session_123"}

	_, err := api.generateFileHandle(
		session,
		fileContent,
		resolveUploadMimeType("test.bin", fileContent),
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidFile)
	assert.LengthSlice(t, 0, mockClient.Calls)
}

func TestGenerateFileHandle_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, errors.New("network down"))

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	session := UploadSession{Id: "session_123"}

	_, err := api.generateFileHandle(session, pngBytes, "image/png")

	assert.Error(t, err)
	assert.Equal(t, false, errors.Is(err, ErrInvalidFile))
}

func TestGenerateFileHandle_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`not-json`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	session := UploadSession{Id: "session_123"}

	_, err := api.generateFileHandle(session, pngBytes, "image/png")

	assert.Error(t, err)
	assert.Equal(t, false, errors.Is(err, ErrInvalidFile))
}

func TestUploadTemplateFile_GenerateFileHandleError(t *testing.T) {
	// The session call succeeds, but the file bytes resolve to the default
	// content type, so generateFileHandle rejects them with ErrInvalidFile.
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       newReusableBody([]byte(`{"id":"session_123"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	fileHeader := createMockFileHeader(t, "data.bin", []byte{0x00, 0x01, 0x02, 0x03})

	_, err := UploadTemplateFile(api, "app_123", fileHeader)

	assert.Error(t, err)
}

// TestUploadTemplateFile_SessionAndUploadAgree covers a routine mislabel: a
// browser derives a file's declared type from its name, so a PNG saved as
// .jpeg arrives named .jpeg. The session's file_type and the upload's
// Content-Type must still describe the same thing, and it must be what the
// bytes actually are, since Meta validates the handle against them.
func TestUploadTemplateFile_SessionAndUploadAgree(t *testing.T) {
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

	fileHeader := createMockFileHeader(t, "picture.jpeg", pngBytes)

	_, err := UploadTemplateFile(api, "app_123", fileHeader)
	assert.NoError(t, err)
	assert.LengthSlice(t, 2, mockClient.Calls)

	sessionType := queryParamValue(mockClient.Calls[0], "file_type")
	uploadType := headerValue(mockClient.Calls[1], "Content-Type")

	assert.Equal(t, "image/png", sessionType)
	assert.Equal(t, "image/png", uploadType)
}

// TestResolveUploadMimeType covers the two signals the resolver combines: exact
// image magic bytes, which outrank a lying extension, and the extension itself,
// which is the only thing that can tell apart formats whose bytes are
// ambiguous or that no sniffer recognizes.
func TestResolveUploadMimeType(t *testing.T) {
	zipHeader := []byte("PK\x03\x04")
	testCases := []struct {
		name     string
		fileName string
		content  []byte
		expected string
	}{
		{"jpeg", "photo.jpeg", jpegBytes, "image/jpeg"},
		{"png", "shot.png", pngBytes, "image/png"},
		{"png named jpeg", "photo.jpeg", pngBytes, "image/png"},
		{"jpeg named png", "photo.png", jpegBytes, "image/jpeg"},
		{"pdf", "report.pdf", []byte("%PDF-1.7\n"), "application/pdf"},
		{
			"docx",
			"contract.docx",
			append(zipHeader, []byte("word/document.xml")...),
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{"text", "notes.txt", []byte("Hello World"), "text/plain"},
		{"unknown", "data.bin", []byte{0x00, 0x01, 0x02, 0x03}, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mimeType := resolveUploadMimeType(testCase.fileName, testCase.content)
			assert.Equal(t, testCase.expected, mimeType)
		})
	}
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

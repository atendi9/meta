// Package whatsapp provides client functionalities to interact with the WhatsApp API.
package whatsapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// ErrInvalidMimeType indicates that the provided MIME type is not supported for media uploads.
var ErrInvalidMimeType = errors.New("invalid mimeType")

// GenerateMediaID uploads a media file to the WhatsApp API and returns its ID.
//   - It reads the file data from an [io.Reader], constructs the multipart form,
//     and sends an HTTP POST request configured with [xhttp.Options].
//   - The provided mimeType is normalized via [NormalizeMediaMimeType] before
//     validation, so generic (application/octet-stream), mobile-specific
//     (audio/x-m4a, audio/opus), or transcoder-mangled types resolve to a MIME
//     type the WhatsApp Cloud API accepts.
//   - On success, it unmarshals the response into a [Media] type and returns the media ID.
func (api *Client) GenerateMediaID(
	mimeType string,
	filePath string,
	fileBytes io.Reader,
) (string, error) {
	emptyMediaId := ""
	content, err := io.ReadAll(fileBytes)
	if err != nil {
		return emptyMediaId, err
	}
	mimeType = NormalizeMediaMimeType(mimeType, filePath, content)
	if !IsValidMediaUploadType(mimeType) {
		return emptyMediaId, ErrInvalidMimeType
	}
	media := Media{}
	url := api.Endpoint(api.senderID + "/media")
	body := api.Buffer(make([]byte, 0))
	writer, err := api.fileWriter(body, bytes.NewReader(content), mimeType, filePath)
	if err != nil {
		return emptyMediaId, err
	}
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers(writer.FormDataContentType()),
		Body:    body,
	})
	if err != nil {
		return emptyMediaId, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &media); err != nil {
		return emptyMediaId, err
	}

	return media.Id, nil
}

// fileWriter creates and prepares a multipart payload for the media upload.
//   - It uses an [io.Writer] for the request body and an [io.Reader] for the file content,
//     returning a built [multipart.Writer] ready for the request payload.
func (c *Client) fileWriter(
	body io.Writer,
	fileBytes io.Reader,
	mimeType,
	filePath string,
) (*multipart.Writer, error) {
	writer := NewMultipartWriter(body)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath))},
		"Content-Type":        []string{mimeType},
	})
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, fileBytes)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteField("messaging_product", "whatsapp"); err != nil {
		return nil, err
	}

	if err := writer.WriteField("type", getMediaType(mimeType)); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}
	return writer.Value(), nil
}

// getMediaType determines the target WhatsApp media classification based on the provided MIME type.
//   - It maps standard MIME types to specific string categories like image, video, audio, or document.
func getMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	default:
		return "document"
	}
}

// IsValidMediaUploadType checks if the provided mimeType is supported by the WhatsApp API.
//   - It validates against standard formats such as images (jpeg, png, webp),
//     videos (mp4, 3gp), audio (aac, mp4, amr, mpeg, ogg), and various documents
//     (txt, csv, pdf, doc/docx, xls/xlsx, ppt/pptx).
//   - Callers should normalize the MIME type via [NormalizeMediaMimeType] first;
//     this function performs only the final allowlist check.
func IsValidMediaUploadType(mimeType string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	isValid := func(mimeType, validMimeType string) bool { return strings.Contains(mimeType, validMimeType) }
	switch {
	case isValid(mimeType, "image/jpeg"), isValid(mimeType, "image/png"), isValid(mimeType, "image/webp"):
		return true
	case isValid(mimeType, "video/mp4"), isValid(mimeType, "video/3gpp"), isValid(mimeType, "video/3gp"):
		return true
	case isValid(mimeType, "audio/aac"), isValid(mimeType, "audio/mp4"), isValid(mimeType, "audio/amr"), isValid(mimeType, "audio/mpeg"), isValid(mimeType, "audio/ogg"):
		return true
	case isValid(mimeType, "text/plain"), isValid(mimeType, "text/csv"), isValid(mimeType, "application/pdf"),
		isValid(mimeType, "application/msword"), isValid(mimeType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"),
		isValid(mimeType, "application/vnd.ms-excel"), isValid(mimeType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"),
		isValid(mimeType, "application/vnd.ms-powerpoint"), isValid(mimeType, "application/vnd.openxmlformats-officedocument.presentationml.presentation"):
		return true
	default:
		return false
	}
}

// MultipartWriter is a custom wrapper around the standard [multipart.Writer]
// to simplify the creation of multipart requests.
type MultipartWriter struct {
	writer *multipart.Writer
}

// NewMultipartWriter initializes and returns a new [MultipartWriter]
// that writes its output to the provided [io.Writer].
func NewMultipartWriter(body io.Writer) *MultipartWriter {
	return &MultipartWriter{
		writer: multipart.NewWriter(body),
	}
}

// CreatePart creates a new part in the multipart payload using the given [textproto.MIMEHeader].
// It returns an [io.Writer] to which the contents of the part can be written.
func (w *MultipartWriter) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	return w.writer.CreatePart(header)
}

// WriteField calls the underlying writer to append a standard form field
// with the specified field name and value.
func (w *MultipartWriter) WriteField(fieldname, value string) error {
	return w.writer.WriteField(fieldname, value)
}

// Close finishes the multipart message by writing the trailing boundary end line.
func (w *MultipartWriter) Close() error {
	return w.writer.Close()
}

// Value returns the underlying [multipart.Writer] instance used by the wrapper.
func (w *MultipartWriter) Value() *multipart.Writer {
	return w.writer
}

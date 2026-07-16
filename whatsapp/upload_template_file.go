package whatsapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// UploadedTemplateFile represents a file that has been successfully uploaded
// and is ready to be used as a template media.
type UploadedTemplateFile struct {
	Id       string `json:"id"`
	Session  string `json:"session"`
	fileName string // internal use only
	fileSize int64  // internal use only
	mimeType string // internal use only
}

// UploadTemplateFile handles the entire upload workflow for a media template file.
// It opens the provided [multipart.FileHeader], creates an [UploadSession] using
// the [Client], and then uploads the bytes to generate a file handle. It returns
// an [UploadedTemplateFile] containing the ID and Session.
func UploadTemplateFile(
	client *Client,
	appId string,
	mediaFile *multipart.FileHeader,
) (UploadedTemplateFile, error) {
	f, err := mediaFile.Open()
	if err != nil {
		return UploadedTemplateFile{}, err
	}
	fileBytes, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return UploadedTemplateFile{}, err
	}
	var (
		fileName = mediaFile.Filename
		fileSize = len(fileBytes)
		mimeType = resolveUploadMimeType(fileName, fileBytes)
	)

	session, err := client.startUploadSession(
		appId,
		fileName,
		int64(fileSize),
		mimeType,
	)
	if err != nil {
		return UploadedTemplateFile{}, err
	}

	// The session and the upload must announce the same type: Meta records the
	// session's file_type and validates it against the bytes it receives.
	fileHandle, err := client.generateFileHandle(session, fileBytes, mimeType)
	if err != nil {
		return UploadedTemplateFile{}, err
	}

	return UploadedTemplateFile{
		Id:       fileHandle.H,
		Session:  session.Id,
		fileName: fileName,
		fileSize: int64(fileSize),
		mimeType: mimeType,
	}, nil
}

// startUploadSession initializes a new upload session for a given file.
// It makes a POST request to the API and returns an [UploadSession] containing
// the session ID needed for the actual file upload.
func (c *Client) startUploadSession(
	appId string,
	fileName string,
	fileLength int64,
	fileType string,
) (UploadSession, error) {
	url := c.Endpoint(fmt.Sprintf("%s/uploads", appId))
	queryParam := func(key string, value any) *xhttp.QueryParam {
		return &xhttp.QueryParam{Key: key, Value: value}
	}
	res, err := c.Post(url, &xhttp.Options{
		Headers: c.Headers("application/json"),
		QueryParams: []xhttp.HTTPData{
			queryParam("file_name", fileName),
			queryParam("file_length", fileLength),
			queryParam("file_type", fileType),
		},
	})
	if err != nil {
		return UploadSession{}, err
	}
	session := UploadSession{}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &session); err != nil {
		return UploadSession{}, err
	}
	return session, nil
}

// UploadSession holds the data returned when a session is initialized,
// primarily the session ID used for further upload steps.
type UploadSession struct {
	Id string `json:"id"`
}

// UploadedFileHandle contains the handle hash returned by the API once
// the file payload has been completely uploaded.
type UploadedFileHandle struct {
	H string `json:"h"`
}

// resolveUploadMimeType determines the MIME type to announce for a file that is
// being uploaded, from its name and its bytes.
//
// The file name alone is not trustworthy: browsers derive a file's declared type
// from its extension, so a mislabeled name makes the upload announce a type the
// bytes contradict, and Meta rejects it. The bytes alone are not enough either,
// since magic numbers cannot tell a .xls from a .ppt. So the extension resolves
// the type, and an exact image signature overrides it.
func resolveUploadMimeType(fileName string, content []byte) string {
	if magic := imageMagicMimeType(content); magic != "" {
		return magic
	}
	return NormalizeMediaMimeType("", fileName, content)
}

// ErrInvalidFile is returned when the provided file bytes do not resolve to a
// MIME type the WhatsApp Cloud API accepts for uploads.
var ErrInvalidFile = errors.New("invalid file")

// generateFileHandle uploads the file's binary content using an existing [UploadSession].
// It returns an [UploadedFileHandle] which contains the file hash necessary to use
// the media in WhatsApp templates.
//
// The Resumable Upload API expects the raw bytes as the request body (the doc's
// curl uses --data-binary), not a multipart envelope: Meta stores whatever the
// body carries and later validates it by magic number, so an envelope makes the
// upload succeed and every later use of the handle fail.
func (c *Client) generateFileHandle(
	session UploadSession,
	fileContent []byte,
	mimeType string,
) (UploadedFileHandle, error) {
	if !IsValidMediaUploadType(mimeType) {
		return UploadedFileHandle{}, ErrInvalidFile
	}
	url := c.Endpoint(session.Id)
	headers := append(c.Headers(mimeType), &xhttp.Header{
		Key:   "file_offset",
		Value: 0,
	})

	res, err := c.Post(url, &xhttp.Options{
		Headers: headers,
		Body:    bytes.NewReader(fileContent),
	})
	if err != nil {
		return UploadedFileHandle{}, err
	}
	defer res.Body.Close()
	var response UploadedFileHandle
	if err := xjson.Decode(res.Body, &response); err != nil {
		return UploadedFileHandle{}, err
	}
	return response, nil
}

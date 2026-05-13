package whatsapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"

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
		fileName  = mediaFile.Filename
		fileSize  = len(fileBytes)
		extension = filepath.Ext(fileName)
		mimeType  = mime.TypeByExtension(extension)
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

	fileHandle, err := client.generateFileHandle(
		session,
		fileBytes,
		fileName,
	)
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

// ErrInvalidFile is returned when the content type of the provided file
// bytes cannot be properly detected or resolves to the default content type.
var ErrInvalidFile = errors.New("invalid file")

// generateFileHandle uploads the file's binary content using an existing [UploadSession].
// It returns an [UploadedFileHandle] which contains the file hash necessary to use
// the media in WhatsApp templates.
func (c *Client) generateFileHandle(
	session UploadSession,
	fileContent []byte,
	fileName string,
) (UploadedFileHandle, error) {
	url := c.Endpoint(session.Id)
	body := c.Buffer(make([]byte, 0))
	mt := NewMessageType(http.DetectContentType(fileContent))
	mimeType := MimeTypeWithMessageType(mt)
	if mimeType == DefaultContentType {
		return UploadedFileHandle{}, ErrInvalidFile
	}
	writer, err := c.fileWriter(
		body,
		bytes.NewReader(fileContent),
		mimeType,
		GenerateFileName(fileName),
	)
	if err != nil {
		return UploadedFileHandle{}, err
	}
	headers := c.Headers(writer.FormDataContentType())
	headers = append(headers, &xhttp.Header{
		Key:   "file_offset",
		Value: 0,
	})

	res, err := c.Post(url, &xhttp.Options{
		Headers: headers,
		Body:    body,
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

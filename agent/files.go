package agent

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// FileRequest represents the multipart payload used to upload a knowledge file.
type FileRequest struct {
	// FileName is the designation for the uploaded file. Required.
	FileName string
	// File holds the file contents to upload. Required.
	//   - Maximum 100MB.
	//   - Accepted types: PDF, DOC, DOCX, PNG, JPG, JPEG, CSV, XLSX.
	File io.Reader

	// body holds the lazily-built multipart payload consumed by Read.
	body io.Reader
	// contentType holds the multipart Content-Type, including the boundary,
	// generated while building the body.
	contentType string
}

// Read implements [io.Reader], streaming the multipart-encoded request body.
// The payload is built lazily on the first call so the request can be passed
// directly as the body of an HTTP request. After the first call, ContentType
// returns the Content-Type header value matching the generated boundary.
func (r *FileRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		if err := r.build(); err != nil {
			return 0, err
		}
	}
	return r.body.Read(p)
}

// build encodes the request into a multipart/form-data body and records the
// matching Content-Type header value.
func (r *FileRequest) build() error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("file_name", r.FileName); err != nil {
		return err
	}
	part, err := writer.CreateFormFile("file", r.FileName)
	if err != nil {
		return err
	}
	if r.File != nil {
		if _, err := io.Copy(part, r.File); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	r.body = &buf
	r.contentType = writer.FormDataContentType()
	return nil
}

// ContentType returns the multipart Content-Type header value, including the
// boundary. The body is built on demand if it has not been read yet.
func (r *FileRequest) ContentType() (string, error) {
	if r.body == nil {
		if err := r.build(); err != nil {
			return "", err
		}
	}
	return r.contentType, nil
}

// FileResponse represents a knowledge file object.
type FileResponse struct {
	// ID is the unique identifier for this file.
	ID string `json:"id"`
	// FileName is the designation of the uploaded file.
	FileName string `json:"file_name"`
}

// Files provides access to the agent knowledge files of an entity.
type Files struct {
	o *Onboard
}

// Files exposes the knowledge file management operations for the configured entity.
func (c *Configurator) Files() *Files {
	return &Files{o: c.Client.Onboard}
}

// Upload adds a new knowledge file to the specified entity.
func (f *Files) Upload(file FileRequest, agentId ...string) (FileResponse, error) {
	contentType, err := file.ContentType()
	if err != nil {
		return FileResponse{}, err
	}
	url := f.o.Config.URL("/agent_config/files")
	headers := f.o.Client.Headers(contentType)
	var result FileResponse
	opts := &xhttp.Options{
		Headers:     headers,
		Body:        &file,
		QueryParams: agentIDParam(agentId),
	}
	res, err := f.o.Client.Post(url, opts)
	if err != nil {
		return FileResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return FileResponse{}, err
	}
	return result, nil
}

// List retrieves the knowledge files for the specified entity.
func (f *Files) List(agentId ...string) ([]FileResponse, error) {
	url := f.o.Config.URL("/agent_config/files")
	headers := f.o.Client.Headers("application/json")
	var result struct {
		Root []FileResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers:     headers,
		QueryParams: agentIDParam(agentId),
	}
	res, err := f.o.Client.Get(url, opts)
	if err != nil {
		return []FileResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []FileResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single knowledge file by its identifier.
func (f *Files) Get(fileId string) (FileResponse, error) {
	url := f.o.Config.URL("/agent_config/files/" + fileId)
	headers := f.o.Client.Headers("application/json")
	var result FileResponse
	res, err := f.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return FileResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return FileResponse{}, err
	}
	return result, nil
}

// Delete removes a knowledge file by its identifier.
func (f *Files) Delete(fileId string) error {
	url := f.o.Config.URL("/agent_config/files/" + fileId)
	headers := f.o.Client.Headers("application/json")
	_, err := f.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

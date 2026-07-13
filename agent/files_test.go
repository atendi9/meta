package agent

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestFileRequest_ContentTypeAndBody(t *testing.T) {
	req := FileRequest{FileName: "doc.pdf", File: strings.NewReader("hello")}

	contentType, err := req.ContentType()
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(contentType, "multipart/form-data; boundary="))

	body, err := io.ReadAll(&req)
	assert.NoError(t, err)
	payload := string(body)
	assert.True(t, contains(payload, `name="file_name"`))
	assert.True(t, contains(payload, "doc.pdf"))
	assert.True(t, contains(payload, `name="file"`))
	assert.True(t, contains(payload, "hello"))
}

func TestFiles_Upload_Success(t *testing.T) {
	mock := mockOK(`{"id":"file_1","file_name":"doc.pdf"}`)
	f := newTestConfigurator(mock).Files()

	res, err := f.Upload(FileRequest{FileName: "doc.pdf", File: strings.NewReader("hello")}, "agent_1")

	assert.NoError(t, err)
	assert.Equal(t, "file_1", res.ID)
	assert.Equal(t, "doc.pdf", res.FileName)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/files", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_1", params[0].Data().Value.(string))

	// The multipart Content-Type header carries a boundary.
	headers := mock.Calls[0].Options.H()
	assert.True(t, strings.HasPrefix(headers[0].Data().Value.(string), "multipart/form-data; boundary="))
}

func TestFiles_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"file_1","file_name":"doc.pdf"}]}`)
	f := newTestConfigurator(mock).Files()

	files, err := f.List()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, files)
	assert.Equal(t, "file_1", files[0].ID)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
}

func TestFiles_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"file_1","file_name":"doc.pdf"}`)
	f := newTestConfigurator(mock).Files()

	res, err := f.Get("file_1")

	assert.NoError(t, err)
	assert.Equal(t, "file_1", res.ID)
	assert.Equal(t, testBase+"/agent_config/files/file_1", mock.Calls[0].URL)
}

func TestFiles_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	f := newTestConfigurator(mock).Files()

	err := f.Delete("file_1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/files/file_1", mock.Calls[0].URL)
}

func TestFiles_TransportErrors(t *testing.T) {
	f := newTestConfigurator(mockErr(io.ErrClosedPipe)).Files()

	_, err := f.Upload(FileRequest{FileName: "doc.pdf", File: strings.NewReader("hi")})
	assert.Error(t, err)
	_, err = f.List()
	assert.Error(t, err)
	_, err = f.Get("file_1")
	assert.Error(t, err)
	err = f.Delete("file_1")
	assert.Error(t, err)
}

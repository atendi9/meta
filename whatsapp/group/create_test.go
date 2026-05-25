package group

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta/whatsapp"
	"github.com/atendi9/meta/xhttp"
)

func TestCreate_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": "12345", "sub_code": "success"}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := whatsapp.Default("10987654321", "valid_token")
	api.HttpClient = mockClient

	def := Definition{
		Name:             "Dev Team",
		Description:      "Group for backend discussions",
		JoinApprovalMode: AutoApprove,
	}

	res, err := Create(api, def)

	assert.NoError(t, err)
	assert.Equal(t, mockRes, res)
	assert.LengthSlice(t, 1, mockClient.Calls)
}

func TestCreate_Error(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := whatsapp.Default("10987654321", "valid_token")
	api.HttpClient = mockClient

	def := Definition{
		Name:             "Dev Team",
		Description:      "Group for backend discussions",
		JoinApprovalMode: ApprovalRequired,
	}

	res, err := Create(api, def)

	assert.Error(t, err)
	assert.Equal(t, (*http.Response)(nil), res)
}

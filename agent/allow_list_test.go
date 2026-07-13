package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestAllowList_Get_Success(t *testing.T) {
	body := `{"root":[{"id":"e1","consumer_phone_number":"5511999999999"}]}`
	mock := mockOK(body)
	c := newTestClient(mock)

	entries, err := c.AllowList.Get()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, entries)
	assert.Equal(t, "e1", entries[0].EntryId)
	assert.Equal(t, "5511999999999", entries[0].PhoneNumber)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/allowlist", mock.Calls[0].URL)
}

func TestAllowList_Get_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrUnexpectedEOF))

	entries, err := c.AllowList.Get()

	assert.Error(t, err)
	assert.LengthSlice(t, 0, entries)
}

func TestAllowList_Add_Success(t *testing.T) {
	mock := mockOK(`{"id":"e1","consumer_phone_number":"5511999999999"}`)
	c := newTestClient(mock)

	entry, err := c.AllowList.Add("5511999999999")

	assert.NoError(t, err)
	assert.Equal(t, "e1", entry.EntryId)
	assert.Equal(t, "5511999999999", entry.PhoneNumber)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/allowlist", mock.Calls[0].URL)

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.True(t, contains(string(sent), `"consumer_phone_number":"5511999999999"`))
}

func TestAllowList_Add_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrClosedPipe))

	_, err := c.AllowList.Add("5511999999999")

	assert.Error(t, err)
}

func TestAllowList_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	c := newTestClient(mock)

	err := c.AllowList.Delete("e1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/allowlist/e1", mock.Calls[0].URL)
}

func TestAllowList_Delete_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrClosedPipe))

	err := c.AllowList.Delete("e1")

	assert.Error(t, err)
}

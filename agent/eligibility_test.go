package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestOnboard_Eligibility_Success(t *testing.T) {
	mock := mockOK(`{"is_eligible": true}`)
	c := newTestClient(mock)

	ok, err := c.Onboard.Eligibility()

	assert.NoError(t, err)
	assert.True(t, ok)
	assert.LengthSlice(t, 1, mock.Calls)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_eligibility", mock.Calls[0].URL)
}

func TestOnboard_Eligibility_NotEligible(t *testing.T) {
	c := newTestClient(mockOK(`{"is_eligible": false}`))

	ok, err := c.Onboard.Eligibility()

	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestOnboard_Eligibility_TransportError(t *testing.T) {
	c := newTestClient(mockErr(io.ErrUnexpectedEOF))

	ok, err := c.Onboard.Eligibility()

	assert.Error(t, err)
	assert.False(t, ok)
}

func TestOnboard_Eligibility_DecodeError(t *testing.T) {
	c := newTestClient(mockOK(`not-json`))

	ok, err := c.Onboard.Eligibility()

	assert.Error(t, err)
	assert.False(t, ok)
}

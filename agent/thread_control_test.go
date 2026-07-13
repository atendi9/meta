package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestThreadControlAction_String(t *testing.T) {
	assert.Equal(t, "release", ThreadControlRelease.String())
	assert.Equal(t, "pass", ThreadControlPass.String())
	// Unknown actions default to release.
	assert.Equal(t, "release", ThreadControlAction("bogus").String())
}

func TestThreadControl_Release_Success(t *testing.T) {
	mock := mockOK(`{"messaging_product":"whatsapp"}`)
	tc := newTestConfigurator(mock).Operate().ThreadControl()

	res, err := tc.Release("5511999999999")

	assert.NoError(t, err)
	assert.Equal(t, "whatsapp", res.MessagingProduct)
	assert.LengthSlice(t, 1, mock.Calls)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, BaseURL+"/business/whatsapp/phone_numbers/"+testEntityID+"/thread_control", mock.Calls[0].URL)
}

func TestThreadControl_Release_Body(t *testing.T) {
	mock := mockOK(`{"messaging_product":"whatsapp"}`)
	tc := newTestConfigurator(mock).Operate().ThreadControl()

	_, err := tc.Release("5511999999999")
	assert.NoError(t, err)

	sent, err := io.ReadAll(mock.Calls[0].Options.B())
	assert.NoError(t, err)
	body := string(sent)
	assert.True(t, contains(body, `"messaging_product": "whatsapp"`))
	assert.True(t, contains(body, `"action": "release"`))
	assert.True(t, contains(body, `"to": "5511999999999"`))
}

func TestThreadControl_Pass_Success(t *testing.T) {
	mock := mockOK(`{"messaging_product":"whatsapp"}`)
	tc := newTestConfigurator(mock).Operate().ThreadControl()

	_, err := tc.Pass("5511999999999")

	assert.NoError(t, err)
	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.True(t, contains(string(sent), `"action": "pass"`))
}

func TestThreadControl_PhoneNumberOverride(t *testing.T) {
	mock := mockOK(`{"messaging_product":"whatsapp"}`)
	tc := newTestConfigurator(mock).Operate().ThreadControl()

	_, err := tc.Release("5511999999999", "999888777")

	assert.NoError(t, err)
	assert.Equal(t, BaseURL+"/business/whatsapp/phone_numbers/999888777/thread_control", mock.Calls[0].URL)
}

func TestThreadControl_Set_DefaultsMessagingProduct(t *testing.T) {
	mock := mockOK(`{"messaging_product":"whatsapp"}`)
	tc := newTestConfigurator(mock).Operate().ThreadControl()

	_, err := tc.Set(ThreadControlRequest{Action: string(ThreadControlRelease), To: "5511999999999"})

	assert.NoError(t, err)
	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.True(t, contains(string(sent), `"messaging_product": "whatsapp"`))
}

func TestThreadControl_TransportError(t *testing.T) {
	tc := newTestConfigurator(mockErr(io.ErrClosedPipe)).Operate().ThreadControl()

	_, err := tc.Release("5511999999999")

	assert.Error(t, err)
}

func TestThreadControl_DecodeError(t *testing.T) {
	tc := newTestConfigurator(mockOK(`not-json`)).Operate().ThreadControl()

	_, err := tc.Release("5511999999999")

	assert.Error(t, err)
}

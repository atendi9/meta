package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestBusinessInfo_Get_Success(t *testing.T) {
	body := `{"business_description":"We sell shoes","payment_method":"card","contact_info":{"email":"a@b.com","address":"Main St","hours_of_operation":"9-5"}}`
	mock := mockOK(body)
	b := newTestConfigurator(mock).BusinessInfo()

	info, err := b.Get()

	assert.NoError(t, err)
	assert.Equal(t, "We sell shoes", info.BusinessDescription)
	assert.Equal(t, "card", info.PaymentMethod)
	assert.Equal(t, "a@b.com", info.ContactInfo.Email)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/business_info", mock.Calls[0].URL)
}

func TestBusinessInfo_Get_TransportError(t *testing.T) {
	b := newTestConfigurator(mockErr(io.ErrUnexpectedEOF)).BusinessInfo()

	_, err := b.Get()

	assert.Error(t, err)
}

func TestBusinessInfo_Update_Success(t *testing.T) {
	mock := mockOK(`{"business_description":"Updated","return_policy":"30 days"}`)
	b := newTestConfigurator(mock).BusinessInfo()

	info, err := b.Update(BusinessInfoRequest{
		BusinessDescription: "Updated",
		ReturnPolicy:        "30 days",
		ContactInfo:         ContactInfo{Email: "a@b.com"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Updated", info.BusinessDescription)
	assert.Equal(t, "30 days", info.ReturnPolicy)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	body := string(sent)
	assert.True(t, contains(body, `"business_description": "Updated"`))
	assert.True(t, contains(body, `"return_policy": "30 days"`))
}

func TestBusinessInfo_Update_TransportError(t *testing.T) {
	b := newTestConfigurator(mockErr(io.ErrClosedPipe)).BusinessInfo()

	_, err := b.Update(BusinessInfoRequest{})

	assert.Error(t, err)
}

func TestBusinessInfo_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	b := newTestConfigurator(mock).BusinessInfo()

	_, err := b.Delete()

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/business_info", mock.Calls[0].URL)
}

func TestBusinessInfo_Delete_TransportError(t *testing.T) {
	b := newTestConfigurator(mockErr(io.ErrClosedPipe)).BusinessInfo()

	_, err := b.Delete()

	assert.Error(t, err)
}

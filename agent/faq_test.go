package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestFAQs_Create_Success(t *testing.T) {
	mock := mockOK(`{"id":"f1","question":"Q?","answer":"A."}`)
	f := newTestConfigurator(mock).FAQs()

	faq, err := f.Create(FAQRequest{Question: "Q?", Answer: "A."}, "agent_1")

	assert.NoError(t, err)
	assert.Equal(t, "f1", faq.ID)
	assert.Equal(t, "Q?", faq.Question)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/faq", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.LengthSlice(t, 1, params)
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_1", params[0].Data().Value.(string))
}

func TestFAQs_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"f1","question":"Q?","answer":"A."}]}`)
	f := newTestConfigurator(mock).FAQs()

	faqs, err := f.List()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, faqs)
	assert.Equal(t, "f1", faqs[0].ID)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.LengthSlice(t, 0, mock.Calls[0].Options.Q())
}

func TestFAQs_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"f1","question":"Q?","answer":"A."}`)
	f := newTestConfigurator(mock).FAQs()

	faq, err := f.Get("f1")

	assert.NoError(t, err)
	assert.Equal(t, "f1", faq.ID)
	assert.Equal(t, testBase+"/agent_config/faq/f1", mock.Calls[0].URL)
}

func TestFAQs_Update_Success(t *testing.T) {
	mock := mockOK(`{"id":"f1","question":"Q2?","answer":"A2."}`)
	f := newTestConfigurator(mock).FAQs()

	faq, err := f.Update("f1", FAQRequest{Question: "Q2?", Answer: "A2."})

	assert.NoError(t, err)
	assert.Equal(t, "Q2?", faq.Question)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/faq/f1", mock.Calls[0].URL)
}

func TestFAQs_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	f := newTestConfigurator(mock).FAQs()

	err := f.Delete("f1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/faq/f1", mock.Calls[0].URL)
}

func TestFAQs_TransportErrors(t *testing.T) {
	f := newTestConfigurator(mockErr(io.ErrClosedPipe)).FAQs()

	_, err := f.Create(FAQRequest{})
	assert.Error(t, err)
	_, err = f.List()
	assert.Error(t, err)
	_, err = f.Get("f1")
	assert.Error(t, err)
	_, err = f.Update("f1", FAQRequest{})
	assert.Error(t, err)
	err = f.Delete("f1")
	assert.Error(t, err)
}

package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestWebsites_Create_Success(t *testing.T) {
	mock := mockOK(`{"id":"w1","url":"https://x.com","crawl_status":"pending"}`)
	w := newTestConfigurator(mock).Websites()

	site, err := w.Create(WebsiteRequest{URL: "https://x.com"}, "agent_1")

	assert.NoError(t, err)
	assert.Equal(t, "w1", site.ID)
	assert.Equal(t, "https://x.com", site.URL)
	assert.Equal(t, "pending", site.CrawlStatus)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/websites", mock.Calls[0].URL)

	params := mock.Calls[0].Options.Q()
	assert.Equal(t, "agent_id", params[0].Data().Key)
	assert.Equal(t, "agent_1", params[0].Data().Value.(string))
}

func TestWebsites_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"w1","url":"https://x.com","pages_crawled":3}]}`)
	w := newTestConfigurator(mock).Websites()

	sites, err := w.List()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, sites)
	assert.Equal(t, int64(3), sites[0].PagesCrawled)
}

func TestWebsites_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"w1","url":"https://x.com"}`)
	w := newTestConfigurator(mock).Websites()

	site, err := w.Get("w1")

	assert.NoError(t, err)
	assert.Equal(t, "w1", site.ID)
	assert.Equal(t, testBase+"/agent_config/websites/w1", mock.Calls[0].URL)
}

func TestWebsites_Update_Success(t *testing.T) {
	mock := mockOK(`{"id":"w1","url":"https://y.com"}`)
	w := newTestConfigurator(mock).Websites()

	site, err := w.Update("w1", WebsiteRequest{URL: "https://y.com"})

	assert.NoError(t, err)
	assert.Equal(t, "https://y.com", site.URL)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/websites/w1", mock.Calls[0].URL)
}

func TestWebsites_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	w := newTestConfigurator(mock).Websites()

	err := w.Delete("w1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_config/websites/w1", mock.Calls[0].URL)
}

func TestWebsites_TransportErrors(t *testing.T) {
	w := newTestConfigurator(mockErr(io.ErrClosedPipe)).Websites()

	_, err := w.Create(WebsiteRequest{})
	assert.Error(t, err)
	_, err = w.List()
	assert.Error(t, err)
	_, err = w.Get("w1")
	assert.Error(t, err)
	_, err = w.Update("w1", WebsiteRequest{})
	assert.Error(t, err)
	err = w.Delete("w1")
	assert.Error(t, err)
}

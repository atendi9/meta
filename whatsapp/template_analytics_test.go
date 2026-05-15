package whatsapp

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestTemplateAnalytics(t *testing.T) {
	client := &Client{}
	api := TemplateAnalytics(client)

	assert.Equal(t, client, api.client)
}

func TestTmplAnalytics_Enable_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	})

	ok := api.Enable()

	assert.True(t, ok)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestTmplAnalytics_Enable_Error(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	})

	ok := api.Enable()

	assert.False(t, ok)
}

func TestTmplAnalytics_WithInterval_Success(t *testing.T) {
	responseBody := `{
		"data": [
			{
				"granularity": "daily",
				"product_type": "marketing",
				"data_points": [
					{
						"template_id": "tmpl_123",
						"sent": 100,
						"delivered": 90,
						"read": 80
					}
				]
			}
		],
		"paging": {
			"next": "https://graph.facebook.com/next_page",
			"cursors": {
				"after": "after_token",
				"before": "before_token"
			}
		}
	}`

	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	})

	interval := TmplAnalyticsInterval{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	res, err := api.WithInterval([]string{"tmpl_123"}, interval)

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodGet, mockClient.Calls[0].Method)

	assert.LengthSlice(t, 1, res.Data)
	assert.Equal(t, "marketing", res.Data[0].ProductType)
	assert.LengthSlice(t, 1, res.Data[0].DataPoints)
	assert.Equal(t, "tmpl_123", res.Data[0].DataPoints[0].TemplateID)
	assert.Equal(t, 100, res.Data[0].DataPoints[0].Sent)
}

func TestTmplAnalytics_WithInterval_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	})

	interval := TmplAnalyticsInterval{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	res, err := api.WithInterval([]string{"tmpl_123"}, interval)

	assert.Error(t, err)
	assert.Equal(t, (*TmplAnalyticsResponse)(nil), res)
}

func TestTmplAnalytics_WithInterval_DecodeError(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{invalid_json}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	})

	interval := TmplAnalyticsInterval{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	res, err := api.WithInterval([]string{"tmpl_123"}, interval)

	assert.Error(t, err)
	assert.Equal(t, (*TmplAnalyticsResponse)(nil), res)
}

func TestTmplAnalytics_WithInterval_WithPaging(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"data":[]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := TemplateAnalytics(&Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	})

	interval := TmplAnalyticsInterval{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	paging := Paging{Cursors: Cursors{After: "after_token"}}

	res, err := api.WithInterval([]string{"tmpl_123"}, interval, paging)

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.True(t, len(mockClient.Calls[0].URL) > 0)
	_ = res
}

func TestTmplAnalytics_urlWithPaging(t *testing.T) {
	api := TemplateAnalytics(&Client{})
	baseURL := "https://api.whatsapp.com/v1/analytics"

	pagingNext := Paging{Next: "https://api.whatsapp.com/v1/analytics/next"}
	urlNext := api.urlWithPaging(baseURL, pagingNext)
	assert.Equal(t, "https://api.whatsapp.com/v1/analytics/next", urlNext)

	pagingAfter := Paging{Cursors: Cursors{After: "token_after"}}
	urlAfter := api.urlWithPaging(baseURL, pagingAfter)
	assert.Equal(t, "https://api.whatsapp.com/v1/analytics&after=token_after", urlAfter)

	pagingBefore := Paging{Cursors: Cursors{Before: "token_before"}}
	urlBefore := api.urlWithPaging(baseURL, pagingBefore)
	assert.Equal(t, "https://api.whatsapp.com/v1/analytics&before=token_before", urlBefore)

	pagingBoth := Paging{Cursors: Cursors{After: "token_after", Before: "token_before"}}
	urlBoth := api.urlWithPaging(baseURL, pagingBoth)
	assert.Equal(t, "https://api.whatsapp.com/v1/analytics&after=token_after&before=token_before", urlBoth)
}

func TestTmplAnalyticsResponse_Bytes(t *testing.T) {
	response := &TmplAnalyticsResponse{
		Paging: Paging{
			Next: "https://next.page.com",
		},
	}

	b := response.Bytes()
	
	ok := bytes.Contains(b, []byte("https://next.page.com"))
	assert.True(t, ok)
}
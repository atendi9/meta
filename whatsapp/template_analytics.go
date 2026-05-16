package whatsapp

import (
	"fmt"
	"time"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// TmplAnalytics handles template-related analytical operations.
type TmplAnalytics struct {
	client *Client
}

// TemplateAnalytics creates a new instance of [TmplAnalytics].
func TemplateAnalytics(
	whats *Client,
) *TmplAnalytics {
	return &TmplAnalytics{whats}
}

// Enable activates insights for the configured sender ID.
// Returns true if the operation was successful.
func (api *TmplAnalytics) Enable() bool {
	whats := api.client
	url := xhttp.SetQueryParams(
		whats.Endpoint(whats.senderID), 
		api.queryParam("is_enabled_for_insights", true),
	)
	res, err := whats.Post(url, &xhttp.Options{
		Headers: whats.Headers("application/json"),
	})
	if err != nil {
		return false
	}
	drainAndClose(res)
	return true
}

// TmplAnalyticsInterval defines a time range for analytical queries.
type TmplAnalyticsInterval struct {
	Start, End time.Time
}

// WithInterval retrieves template analytics metrics within a specific time interval.
// It supports filtering by template IDs and handles pagination via [Paging].
func (api *TmplAnalytics) WithInterval(
	templateIDs []string,
	interval TmplAnalyticsInterval,
	paging ...Paging,
) (*TmplAnalyticsResponse, error) {
	whats := api.client
	startTs := interval.Start.Unix()
	endTs := interval.End.Unix()
	url := whats.Endpoint(fmt.Sprintf("%s/template_analytics", whats.senderID))
	url = xhttp.SetQueryParams(
		url,
		api.queryParam("start", startTs),
		api.queryParam("end", endTs),
		api.queryParam("granularity", "daily"),
		api.queryParam("metric_types", "cost,clicked,delivered,read,sent"),
		api.queryParam("template_ids", templateIDs),
	)
	if len(paging) > 0 {
		url = api.urlWithPaging(url, paging[0])
	}
	resp, err := whats.Get(url, &xhttp.Options{
		Headers: whats.Headers("application/json"),
	})
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	response := new(TmplAnalyticsResponse)
	if err := xjson.Decode(resp.Body, response); err != nil {
		return nil, err
	}
	return response, nil
}

// queryParam is a helper to build [xhttp.QueryParam].
func (api *TmplAnalytics) queryParam(key string, value any) (query *xhttp.QueryParam) {
	query = new(xhttp.QueryParam)
	xhttp.NewData(query, key, value)
	return
}

// urlWithPaging injects pagination cursors into the request URL.
func (api *TmplAnalytics) urlWithPaging(
	url string,
	paging Paging,
) string {
	if len(paging.Next) > 0 {
		return paging.Next
	}
	cursors := paging.Cursors
	after, before := cursors.After, cursors.Before
	if len(after) > 0 {
		url += fmt.Sprintf("&after=%s", after)
	}
	if len(before) > 0 {
		url += fmt.Sprintf("&before=%s", before)
	}
	return url
}

// TmplAnalyticsResponse represents the structure of the API response for analytics.
type TmplAnalyticsResponse struct {
	Data   []DataItem `json:"data"`
	Paging Paging     `json:"paging"`
}

// Bytes serializes [TmplAnalyticsResponse] into a JSON byte slice using [xjson.Bytes].
func (data *TmplAnalyticsResponse) Bytes() []byte {
	return xjson.Bytes(data)
}

// DataItem contains metrics grouped by product type and granularity.
type DataItem struct {
	DataPoints  []DataPoint `json:"data_points"`
	Granularity string      `json:"granularity"`
	ProductType string      `json:"product_type"`
}

// DataPoint holds the actual metrics for a specific template during a time window.
type DataPoint struct {
	Clicked    []Clicked `json:"clicked"`
	Cost       []Cost    `json:"cost"`
	Delivered  int       `json:"delivered"`
	End        int64     `json:"end"`
	Read       int       `json:"read"`
	Sent       int       `json:"sent"`
	Start      int64     `json:"start"`
	TemplateID string    `json:"template_id"`
}

// Clicked contains interaction details for template buttons.
type Clicked struct {
	ButtonContent string `json:"button_content"`
	Count         int    `json:"count"`
	Type          string `json:"type"`
}

// Cost represents the cost metrics associated with the message.
type Cost struct {
	Type string `json:"type"`
}

// Paging contains navigation metadata for paginated results.
type Paging struct {
	Cursors Cursors `json:"cursors"`
	Next    string  `json:"next"`
}

// Cursors holds the pointers for the next and previous pages.
type Cursors struct {
	After  string `json:"after"`
	Before string `json:"before"`
}

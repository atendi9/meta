package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// WebsiteRequest represents the body used to create or update a knowledge website.
type WebsiteRequest struct {
	// URL is the website address to crawl for agent knowledge.
	URL string `json:"url"`
	// Metadata holds arbitrary key-value pairs for additional context.
	Metadata map[string]any `json:"metadata,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *WebsiteRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// WebsiteResponse represents a knowledge website object.
type WebsiteResponse struct {
	// ID is the unique identifier for this website crawl entry.
	ID string `json:"id"`
	// URL is the website address being crawled.
	URL string `json:"url"`
	// CrawlStatus is the crawl state: "pending", "in_progress", "completed" or "failed".
	CrawlStatus string `json:"crawl_status,omitzero"`
	// PagesCrawled is the count of successfully crawled pages.
	PagesCrawled int64 `json:"pages_crawled,omitzero"`
	// LastCrawledAt is the Unix timestamp of the last successful crawl.
	LastCrawledAt int64 `json:"last_crawled_at,omitzero"`
	// CreatedAt is the creation timestamp.
	CreatedAt int64 `json:"created_at,omitzero"`
}

// Websites provides CRUD access to the agent knowledge websites of an entity.
type Websites struct {
	o *Onboard
}

// Websites exposes the knowledge website management operations for the configured entity.
func (c *Configurator) Websites() *Websites {
	return &Websites{o: c.Client.Onboard}
}

// Create adds a new knowledge website to the specified entity.
func (w *Websites) Create(website WebsiteRequest, agentId ...string) (WebsiteResponse, error) {
	url := w.o.Config.URL("/agent_config/websites")
	headers := w.o.Client.Headers("application/json")
	var result WebsiteResponse
	opts := &xhttp.Options{
		Headers:     headers,
		Body:        &website,
		QueryParams: agentIDParam(agentId),
	}
	res, err := w.o.Client.Post(url, opts)
	if err != nil {
		return WebsiteResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return WebsiteResponse{}, err
	}
	return result, nil
}

// List retrieves the knowledge websites for the specified entity.
func (w *Websites) List(agentId ...string) ([]WebsiteResponse, error) {
	url := w.o.Config.URL("/agent_config/websites")
	headers := w.o.Client.Headers("application/json")
	var result struct {
		Root []WebsiteResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers:     headers,
		QueryParams: agentIDParam(agentId),
	}
	res, err := w.o.Client.Get(url, opts)
	if err != nil {
		return []WebsiteResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []WebsiteResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single knowledge website by its identifier.
func (w *Websites) Get(websiteId string) (WebsiteResponse, error) {
	url := w.o.Config.URL("/agent_config/websites/" + websiteId)
	headers := w.o.Client.Headers("application/json")
	var result WebsiteResponse
	res, err := w.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return WebsiteResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return WebsiteResponse{}, err
	}
	return result, nil
}

// Update modifies an existing knowledge website identified by websiteId.
func (w *Websites) Update(websiteId string, website WebsiteRequest) (WebsiteResponse, error) {
	url := w.o.Config.URL("/agent_config/websites/" + websiteId)
	headers := w.o.Client.Headers("application/json")
	var result WebsiteResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &website,
	}
	res, err := w.o.Client.Put(url, opts)
	if err != nil {
		return WebsiteResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return WebsiteResponse{}, err
	}
	return result, nil
}

// Delete removes a knowledge website by its identifier.
func (w *Websites) Delete(websiteId string) error {
	url := w.o.Config.URL("/agent_config/websites/" + websiteId)
	headers := w.o.Client.Headers("application/json")
	_, err := w.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

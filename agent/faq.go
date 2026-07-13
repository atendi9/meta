package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// FAQRequest represents the body used to create or update an FAQ entry.
type FAQRequest struct {
	// Question is a natural customer question.
	Question string `json:"question"`
	// Answer is a factual, self-contained response to the question.
	Answer string `json:"answer"`
	// Metadata holds arbitrary key-value pairs for additional context.
	Metadata map[string]any `json:"metadata,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *FAQRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// FAQResponse represents an FAQ object.
type FAQResponse struct {
	// ID is the unique identifier for this FAQ.
	ID string `json:"id"`
	// Question is the natural customer question.
	Question string `json:"question"`
	// Answer is the factual, self-contained response to the question.
	Answer string `json:"answer"`
	// CreatedAt is the creation timestamp.
	CreatedAt int64 `json:"created_at,omitzero"`
	// Metadata holds arbitrary key-value pairs for additional context.
	Metadata map[string]any `json:"metadata,omitzero"`
}

// FAQs provides CRUD access to the agent knowledge FAQs of an entity.
type FAQs struct {
	o *Onboard
}

// FAQs exposes the FAQ management operations for the configured entity.
func (c *Configurator) FAQs() *FAQs {
	return &FAQs{o: c.Client.Onboard}
}

// Create adds a new FAQ to the specified entity.
func (f *FAQs) Create(faq FAQRequest, agentId ...string) (FAQResponse, error) {
	url := f.o.Config.URL("/agent_config/faq")
	headers := f.o.Client.Headers("application/json")
	var result FAQResponse
	opts := &xhttp.Options{
		Headers:     headers,
		Body:        &faq,
		QueryParams: agentIDParam(agentId),
	}
	res, err := f.o.Client.Post(url, opts)
	if err != nil {
		return FAQResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return FAQResponse{}, err
	}
	return result, nil
}

// List retrieves the FAQs for the specified entity.
func (f *FAQs) List(agentId ...string) ([]FAQResponse, error) {
	url := f.o.Config.URL("/agent_config/faq")
	headers := f.o.Client.Headers("application/json")
	var result struct {
		Root []FAQResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers:     headers,
		QueryParams: agentIDParam(agentId),
	}
	res, err := f.o.Client.Get(url, opts)
	if err != nil {
		return []FAQResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []FAQResponse{}, err
	}
	return result.Root, nil
}

// Get retrieves a single FAQ by its identifier.
func (f *FAQs) Get(faqId string) (FAQResponse, error) {
	url := f.o.Config.URL("/agent_config/faq/" + faqId)
	headers := f.o.Client.Headers("application/json")
	var result FAQResponse
	res, err := f.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return FAQResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return FAQResponse{}, err
	}
	return result, nil
}

// Update modifies an existing FAQ identified by faqId.
func (f *FAQs) Update(faqId string, faq FAQRequest) (FAQResponse, error) {
	url := f.o.Config.URL("/agent_config/faq/" + faqId)
	headers := f.o.Client.Headers("application/json")
	var result FAQResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &faq,
	}
	res, err := f.o.Client.Put(url, opts)
	if err != nil {
		return FAQResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return FAQResponse{}, err
	}
	return result, nil
}

// Delete removes an FAQ by its identifier.
func (f *FAQs) Delete(faqId string) error {
	url := f.o.Config.URL("/agent_config/faq/" + faqId)
	headers := f.o.Client.Headers("application/json")
	_, err := f.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

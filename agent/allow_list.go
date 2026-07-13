package agent

import (
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// AllowlistResponse represents an allowlist entry.
type AllowlistResponse struct {
	// EntryId is the unique identifier for this allowlist entry.
	EntryId string `json:"id"`
	// PhoneNumber is the consumer WhatsApp phone number in E.164 format.
	PhoneNumber string `json:"consumer_phone_number"`
}

type AllowList struct {
	o *Onboard
}

// Get retrieves the current allowlist for the specified entity.
func (l *AllowList) Get() ([]AllowlistResponse, error) {
	url := l.o.Config.URL("/agent_config/allowlist")
	headers := l.o.Client.Headers("application/json")
	var result struct {
		Root []AllowlistResponse `json:"root"`
	}
	opts := &xhttp.Options{
		Headers: headers,
	}
	res, err := l.o.Client.Get(url, opts)
	if err != nil {
		return []AllowlistResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return []AllowlistResponse{}, err
	}
	return result.Root, nil
}

// Add appends a consumer phone number to the allowlist.
func (l *AllowList) Add(phoneNumber string) (AllowlistResponse, error) {
	payload := xjson.JSON{
		"consumer_phone_number": phoneNumber,
	}
	url := l.o.Config.URL("/agent_config/allowlist")
	headers := l.o.Client.Headers("application/json")
	var result AllowlistResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    payload.Buffer(),
	}
	res, err := l.o.Client.Post(url, opts)
	if err != nil {
		return AllowlistResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return AllowlistResponse{}, err
	}
	return result, nil
}

// Delete removes an allowlist entry by its identifier.
func (l *AllowList) Delete(entryId string) error {
	url := l.o.Config.URL("/agent_config/allowlist/" + entryId)
	headers := l.o.Client.Headers("application/json")
	_, err := l.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return err
	}
	return nil
}

package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// ThreadControlAction identifies the ownership transition to perform on a
// conversation thread.
type ThreadControlAction string

func (a ThreadControlAction) String() string {
	switch a {
	case ThreadControlRelease, ThreadControlPass:
		return string(a)
	default:
		return string(ThreadControlRelease)
	}
}

const (
	// ThreadControlRelease relinquishes control of the conversation back to the
	// Meta Business Agent.
	ThreadControlRelease ThreadControlAction = "release"
	// ThreadControlPass hands control to another participant.
	//   - Reserved for future use; the API currently only supports "release".
	ThreadControlPass ThreadControlAction = "pass"
)

// ThreadControlRequest represents the body used to change ownership of a
// conversation thread.
type ThreadControlRequest struct {
	// MessagingProduct must be "whatsapp". It is set automatically by the
	// client when left empty.
	MessagingProduct string `json:"messaging_product"`
	// Action is the ownership transition to perform.
	Action string `json:"action"`
	// To is the consumer phone number or WhatsApp ID.
	To string `json:"to,omitzero"`
	// Recipient is the business-scoped user ID.
	//   - Not yet functional; use To instead.
	Recipient string `json:"recipient,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *ThreadControlRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// ThreadControlResponse represents the thread control response.
type ThreadControlResponse struct {
	// MessagingProduct is always "whatsapp".
	MessagingProduct string `json:"messaging_product"`
}

// ThreadControl manages conversation ownership between the application and the
// Meta Business Agent for a WhatsApp Business phone number.
type ThreadControl struct {
	o *Onboard
}

// ThreadControl exposes the thread control operations for the configured entity.
func (op *Operator) ThreadControl() *ThreadControl {
	return &ThreadControl{o: op.o}
}

// phoneNumberID resolves the phone number ID to operate on, defaulting to the
// configured entity when no override is provided.
func (t *ThreadControl) phoneNumberID(override []string) string {
	if len(override) > 0 && override[0] != "" {
		return override[0]
	}
	return t.o.Config.EntityID
}

// url builds the thread control endpoint for the given phone number ID.
func (t *ThreadControl) url(phoneNumberID string) string {
	return t.o.Config.BaseURL + "/business/whatsapp/phone_numbers/" + phoneNumberID + "/thread_control"
}

// Release relinquishes control of the conversation with to back to the Meta
// Business Agent. An optional phoneNumberID overrides the configured entity.
func (t *ThreadControl) Release(to string, phoneNumberID ...string) (ThreadControlResponse, error) {
	return t.Set(ThreadControlRequest{Action: string(ThreadControlRelease), To: to}, phoneNumberID...)
}

// Pass hands control of the conversation with to to another participant.
//   - Reserved for future use; the API currently only supports Release.
func (t *ThreadControl) Pass(to string, phoneNumberID ...string) (ThreadControlResponse, error) {
	return t.Set(ThreadControlRequest{Action: string(ThreadControlPass), To: to}, phoneNumberID...)
}

// Set performs a thread control transition using the provided request. The
// MessagingProduct field defaults to "whatsapp" when empty. An optional
// phoneNumberID overrides the configured entity.
func (t *ThreadControl) Set(req ThreadControlRequest, phoneNumberID ...string) (ThreadControlResponse, error) {
	if req.MessagingProduct == "" {
		req.MessagingProduct = "whatsapp"
	}
	url := t.url(t.phoneNumberID(phoneNumberID))
	headers := t.o.Client.Headers("application/json")
	var result ThreadControlResponse
	res, err := t.o.Client.Post(url, &xhttp.Options{Headers: headers, Body: &req})
	if err != nil {
		return ThreadControlResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return ThreadControlResponse{}, err
	}
	return result, nil
}

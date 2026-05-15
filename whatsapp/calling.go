package whatsapp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// EventEmitter defines the interface for emitting events.
type EventEmitter interface {
	Emit(eventName string, data any)
}

// Call represents a WhatsApp calling API handler.
// It uses a [Client] to communicate with the API and an [EventEmitter] for real-time events.
type Call struct {
	api    *Client
	appId  string
	io     EventEmitter
	logger func(message string)
}

// NewCall creates and returns a new [Call] instance.
// It requires a [Client], an application ID, an [EventEmitter],
// and an optional logger function.
func NewCall(
	c *Client,
	appId string,
	eventEmitter EventEmitter,
	logger ...func(data string),
) *Call {
	callingAPI := &Call{
		api:    c,
		appId:  appId,
		logger: func(_ string) {},
		io:     eventEmitter,
	}
	if len(logger) > 0 {
		callingAPI.logger = logger[0]
	}
	return callingAPI
}

// AcceptCall accepts an incoming call using the provided ID and answer SDP.
// It sends an accept action payload formatted as [xjson.JSON] through the [Client].
func (c *Call) AcceptCall(id, answerSdp string) string {
	acceptPayload := xjson.JSON{
		"messaging_product": "whatsapp",
		"call_id":           id,
		"action":            "accept",
		"session":           xjson.JSON{"sdp_type": "answer", "sdp": answerSdp},
	}

	res, err := c.api.Post(
		c.api.Endpoint(c.api.senderID+"/calls"),
		&xhttp.Options{
			Headers: c.api.Headers("application/json"),
			Body:    acceptPayload.Buffer(),
		},
	)
	if err != nil {
		c.logger(fmt.Sprintf(`Failed to accept call %s: %s`, id, err.Error()))
		return id
	}
	res.Body.Close()

	return id
}

// InitCall holds the initialization data for an outbound call,
// containing the CallId and AppId.
type InitCall struct {
	CallId string `json:"callId"`
	AppId  string `json:"appId"`
}

// InviteCallResponse represents the API response payload when inviting a call.
type InviteCallResponse struct {
	Id     string `json:"id,omitempty"`
	CallId string `json:"call_id,omitempty"`
}

// ErrCannotGetCallId is returned when the call ID cannot be parsed
// or retrieved from the [InviteCallResponse].
var ErrCannotGetCallId = errors.New("cannot get call id")

// InitiateOutboundCall starts a new outbound call to the specified phone number
// using the provided offer SDP. It returns an [InitCall] or an error if the request fails.
func (c *Call) InitiateOutboundCall(phoneNumber, offerSdp string) (InitCall, error) {
	c.logger(fmt.Sprintf(`Initiating outbound call to %s using Phone Number ID: %s`, phoneNumber, c.api.senderID))

	invitePayload := xjson.JSON{
		"messaging_product": "whatsapp",
		"to":                phoneNumber,
		"session":           xjson.JSON{"sdp_type": "offer", "sdp": offerSdp},
	}

	res, err := c.api.Post(
		c.api.Endpoint(c.api.senderID+"/calls"),
		&xhttp.Options{
			Headers: c.api.Headers("application/json"),
			Body:    invitePayload.Buffer(),
		},
	)
	if err != nil {
		return InitCall{}, err
	}
	defer res.Body.Close()
	var response InviteCallResponse
	if err := xjson.Decode(res.Body, &response); err != nil {
		return InitCall{}, err
	}
	newCallId := ternary(
		len(response.CallId) > 0,
		response.CallId,
		response.Id,
	)
	return InitCall{CallId: newCallId, AppId: c.appId}, ternary(len(newCallId) == 0, ErrCannotGetCallId, nil)
}

// ternary is a generic helper function that returns trueValue if the condition is true,
// otherwise it returns falseValue.
func ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

// CallData holds the status information of a call event.
type CallData struct {
	Status string `json:"status"`
}

// WebhookPreAccept processes webhook events using the provided [CallData] to determine
// the call status. It handles pre-accepting incoming calls, answering outbound calls,
// and notifying when calls end or are canceled via the [EventEmitter].
func (c *Call) WebhookPreAccept(id string, data CallData, sdpPayload string) {
	if len(id) > 0 && data.Status == "ringing" && len(sdpPayload) > 0 {
		c.logger(fmt.Sprintf(`Incoming call detected. Call ID: %s`, id))

		sdpForPreAccept := replaceSdpPayload(sdpPayload)

		preAcceptPayload := xjson.JSON{
			"messaging_product": "whatsapp",
			"call_id":           id,
			"action":            "pre_accept",
			"session":           xjson.JSON{"sdp_type": "answer", "sdp": sdpForPreAccept},
		}

		res, err := c.api.Post(
			c.api.Endpoint(c.api.senderID+"/calls"),
			&xhttp.Options{
				Headers: c.api.Headers("application/json"),
				Body:    preAcceptPayload.Buffer(),
			},
		)
		if err != nil {
			c.logger(fmt.Sprintf(`Failed to pre-accept call %s: %s`, id, err.Error()))
			return
		}
		res.Body.Close()

		c.io.Emit("incoming_call", xjson.JSON{
			"appId":    c.appId,
			"callId":   id,
			"offerSdp": sdpPayload,
		})
	}

	if len(id) > 0 && data.Status == "active" && len(sdpPayload) > 0 {
		c.logger(fmt.Sprintf(`Outbound call answered by user. Call ID: %s`, id))
		c.io.Emit("outbound_call_answered", xjson.JSON{"callId": id, "answerSdp": sdpPayload})
	}

	if data.Status == "ended" || data.Status == "canceled" {
		c.logger(fmt.Sprintf(`Call terminated. Call ID: %s`, id))
		c.io.Emit("call_ended", xjson.JSON{"callId": id})
	}
}

// replaceSdpPayload replaces "setup:actpass" with "setup:passive"
// and "a=sendrecv" with "a=recvonly" in the given SDP payload.
func replaceSdpPayload(sdpPayload string) string {
	sdpForPreAccept := strings.ReplaceAll(sdpPayload, "setup:actpass", "setup:passive")
	return strings.ReplaceAll(sdpForPreAccept, "a=sendrecv", "a=recvonly")
}

// Package group provides abstractions and functions to manage WhatsApp groups
// through the Meta API.
package group

import (
	"fmt"
	"net/http"

	"github.com/atendi9/meta/whatsapp"
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// JoinApprovalMode represents the setting that defines how new members
// are approved to join a specific group.
type JoinApprovalMode string

const (
	// AutoApprove allows members to join without confirmation.
	AutoApprove JoinApprovalMode = "auto_approve"
	// ApprovalRequired mandates admin confirmation before joining.
	ApprovalRequired JoinApprovalMode = "approval_required"
)

// Definition holds the configuration required to create or configure a group.
type Definition struct {
	// Name is the title or subject of the group.
	Name string
	// Description details the purpose or guidelines of the group.
	Description string
	// JoinApprovalMode determines the approval process for new participants.
	JoinApprovalMode JoinApprovalMode
}

// Create sends a request to the Meta API to provision a new group based on
// the provided [Definition] layout using the configured [*whatsapp.Client].
//
// It returns a standard [*http.Response] or an error if the operation fails.
func Create(
	api *whatsapp.Client,
	def Definition,
) (*http.Response, error) {
	url := api.Endpoint(api.SenderID() + "/groups")
	fmt.Println(url)
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers("application/json"),
		Body: xjson.JSON{
			"messaging_product":  "whatsapp",
			"subject":            def.Name,
			"description":        def.Description,
			"join_approval_mode": string(def.JoinApprovalMode),
		}.Buffer(),
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

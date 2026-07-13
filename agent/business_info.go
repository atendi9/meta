package agent

import (
	"bytes"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// ContactInfo holds the business contact details.
type ContactInfo struct {
	// Email is the business email address.
	Email string `json:"email,omitzero"`
	// Address is the physical business location.
	Address string `json:"address,omitzero"`
	// HoursOfOperation describes the business operating hours.
	HoursOfOperation string `json:"hours_of_operation,omitzero"`
}

// BusinessInfoRequest represents the body used to create or replace the
// business information of an entity.
type BusinessInfoRequest struct {
	// BusinessDescription holds general information about the business.
	BusinessDescription string `json:"business_description,omitzero"`
	// PaymentMethod describes the accepted payment methods.
	PaymentMethod string `json:"payment_method,omitzero"`
	// ReturnPolicy describes the company return policy.
	ReturnPolicy string `json:"return_policy,omitzero"`
	// PurchaseInfo holds information about how to make a purchase.
	PurchaseInfo string `json:"purchase_info,omitzero"`
	// DeliveryAndShipping holds details about delivery and shipping.
	DeliveryAndShipping string `json:"delivery_and_shipping,omitzero"`
	// ContactInfo holds the business contact details.
	ContactInfo ContactInfo `json:"contact_info,omitzero"`

	// body holds the lazily-marshaled JSON payload consumed by Read.
	body io.Reader
}

// Read implements [io.Reader], streaming the JSON-encoded request body.
// The payload is marshaled lazily on the first call so the request can be
// passed directly as the body of an HTTP request.
func (r *BusinessInfoRequest) Read(p []byte) (int, error) {
	if r.body == nil {
		r.body = bytes.NewReader(xjson.Bytes(r))
	}
	return r.body.Read(p)
}

// BusinessInfoResponse represents the business information object.
type BusinessInfoResponse struct {
	// BusinessDescription holds general information about the business.
	BusinessDescription string `json:"business_description,omitzero"`
	// PaymentMethod describes the accepted payment methods.
	PaymentMethod string `json:"payment_method,omitzero"`
	// ReturnPolicy describes the company return policy.
	ReturnPolicy string `json:"return_policy,omitzero"`
	// PurchaseInfo holds information about how to make a purchase.
	PurchaseInfo string `json:"purchase_info,omitzero"`
	// DeliveryAndShipping holds details about delivery and shipping.
	DeliveryAndShipping string `json:"delivery_and_shipping,omitzero"`
	// ContactInfo holds the business contact details.
	ContactInfo ContactInfo `json:"contact_info,omitzero"`
}

// BusinessInfo provides access to the agent knowledge business information of an entity.
type BusinessInfo struct {
	o *Onboard
}

// BusinessInfo exposes the business information operations for the configured entity.
func (c *Configurator) BusinessInfo() *BusinessInfo {
	return &BusinessInfo{o: c.Client.Onboard}
}

// Get retrieves the current business information for the specified entity.
// Empty values are returned when no information has been configured.
func (b *BusinessInfo) Get() (BusinessInfoResponse, error) {
	url := b.o.Config.URL("/agent_config/business_info")
	headers := b.o.Client.Headers("application/json")
	var result BusinessInfoResponse
	res, err := b.o.Client.Get(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return BusinessInfoResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return BusinessInfoResponse{}, err
	}
	return result, nil
}

// Update creates or replaces the business information for the specified entity,
// overwriting all fields with the provided values.
func (b *BusinessInfo) Update(info BusinessInfoRequest) (BusinessInfoResponse, error) {
	url := b.o.Config.URL("/agent_config/business_info")
	headers := b.o.Client.Headers("application/json")
	var result BusinessInfoResponse
	opts := &xhttp.Options{
		Headers: headers,
		Body:    &info,
	}
	res, err := b.o.Client.Put(url, opts)
	if err != nil {
		return BusinessInfoResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return BusinessInfoResponse{}, err
	}
	return result, nil
}

// Delete resets the business information to empty default values.
func (b *BusinessInfo) Delete() (BusinessInfoResponse, error) {
	url := b.o.Config.URL("/agent_config/business_info")
	headers := b.o.Client.Headers("application/json")
	var result BusinessInfoResponse
	res, err := b.o.Client.Delete(url, &xhttp.Options{Headers: headers})
	if err != nil {
		return BusinessInfoResponse{}, err
	}
	defer res.Body.Close()
	if err := xjson.Decode(res.Body, &result); err != nil {
		return BusinessInfoResponse{}, err
	}
	return result, nil
}

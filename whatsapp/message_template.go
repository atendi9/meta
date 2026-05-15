package whatsapp

import (
	"fmt"
	"io"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// MessageTemplate creates a [Message] by configuring the [Header] with the provided [xjson.JSON] template.
// It sets the message type to "template" and assigns the template data to the header.
func MessageTemplate(
	h Header,
	template xjson.JSON,
) Message {
	templateType := "template"
	h["type"] = templateType
	h[templateType] = template
	return h
}

// Category defines the classification of the message template according to Meta's requirements.
type Category string

const (
	// MARKETING templates include promotional offers, product announcements, and more.
	MARKETING Category = "MARKETING"
	// UTILITY templates facilitate a specific, agreed-upon transaction or update.
	UTILITY Category = "UTILITY"
	// AUTHENTICATION templates enable businesses to authenticate users with one-time passcodes.
	AUTHENTICATION Category = "AUTHENTICATION"
)

// Lang represents the supported language codes for WhatsApp message templates.
type Lang string

const (
	English           Lang = "en_US"
	Spanish           Lang = "es_ES"
	EnglishUK         Lang = "en_GB"
	PortugueseBrazil  Lang = "pt_BR"
	French            Lang = "fr"
	German            Lang = "de"
	ChineseSimplified Lang = "zh_CN"
	Arabic            Lang = "ar"
	Hindi             Lang = "hi"
	Japanese          Lang = "ja"
)

// CMessageTemplate represents the request structure used to create a new message template on the Meta platform.
type CMessageTemplate struct {
	Name                string       `json:"name"`
	Language            Lang         `json:"language"`
	Category            Category     `json:"category"`
	AllowCategoryChange bool         `json:"allow_category_change"`
	Components          []xjson.JSON `json:"components"`
}

// CreateTemplateResponse contains the identification data returned by the API after a successful template creation.
type CreateTemplateResponse struct {
	Id string `json:"id"`
}

// TemplateFields holds the necessary input data to define a new message template.
type TemplateFields struct {
	Name       string
	Category   Category
	Lang       Lang
	Components []xjson.JSON
}

// CreateMessageTemplate performs an HTTP POST request to the Meta API to create a new message template.
// It uses the [Client] configuration and the provided [TemplateFields] to build the request body.
func CreateMessageTemplate(
	api *Client,
	field TemplateFields,
) error {
	url := api.Endpoint(api.senderID + "/message_templates")

	body := CMessageTemplate{
		Name:                field.Name,
		Language:            field.Lang,
		Category:            field.Category,
		AllowCategoryChange: true,
		Components:          field.Components,
	}
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers("application/json"),
		Body:    api.Reader(xjson.Bytes(body)),
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return nil
}

func DeleteMessageTemplate(api *Client, name string) error {
	url := api.Endpoint(api.senderID + "/message_templates")
	res, err := api.Delete(url, &xhttp.Options{
		QueryParams: []xhttp.HTTPData{
			&xhttp.QueryParam{Key: "name", Value: name},
		},
		Headers: api.Headers("application/json"),
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// Templates represents a generic structure for holding WhatsApp templates data.
type Templates[T any] struct {
	Data []T `json:"data"`
}

// TemplateStatus represents the status information of a WhatsApp template.
type TemplateStatus struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status status `json:"status"`
}

// status represents the possible states of a [TemplateStatus].
type status string

// Pending returns a [status] indicating that the template is pending approval.
func Pending() status {
	return "PENDING"
}

// Rejected returns a [status] indicating that the template has been rejected.
func Rejected() status {
	return "REJECTED"
}

// Aproved returns a [status] indicating that the template has been approved.
func Aproved() status {
	return "APPROVED"
}

// GetTemplateStatus retrieves the [TemplateStatus] for a specific template by its name using the provided [Client].
//   - It returns a [Templates] containing the [TemplateStatus] data.
func GetTemplateStatus(client *Client, name string) (Templates[TemplateStatus], error) {
	return GetTemplates[TemplateStatus](client, &xhttp.Options{
		QueryParams: []xhttp.HTTPData{
			&xhttp.Data{Key: "fields", Value: "name,status"},
			&xhttp.Data{Key: "name", Value: name},
		},
		Headers: client.Headers("application/json"),
	})
}

// GetJSONTemplates retrieves all message templates as raw [xjson.JSON] data using the provided [Client].
//   - It returns a [Templates] containing the [xjson.JSON] data.
func GetJSONTemplates(client *Client) (Templates[xjson.JSON], error) {
	return GetTemplates[xjson.JSON](client, &xhttp.Options{
		Headers: client.Headers("application/json"),
	})
}

// GetTemplates retrieves a list of message templates using the provided [Client] and [xhttp.Options].
func GetTemplates[T any](client *Client, options *xhttp.Options) (Templates[T], error) {
	url := client.Endpoint(client.senderID + "/message_templates")
	emptySlice := Templates[T]{Data: []T{}}
	res, err := client.HttpClient.Get(url, options)
	if err != nil {
		return emptySlice, err
	}
	defer res.Body.Close()
	var messageTemplates Templates[T]
	if err := xjson.Decode(res.Body, &messageTemplates); err != nil {
		return emptySlice, fmt.Errorf("decode error: %v", err)
	}
	return messageTemplates, nil
}

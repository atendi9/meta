package whatsapp

import (
	"encoding/json"
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

	b, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	res, err := api.Post(url, &xhttp.Options{
		Headers: api.Headers("application/json"),
		Body:    api.Reader(b),
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

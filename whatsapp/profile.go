package whatsapp

import (
	"context"
	"mime/multipart"

	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

// Profile represents a WhatsApp business profile.
type Profile struct {
	PhoneNumberID string
	client        *Client
}

// NewProfile creates and returns a new [Profile].
//   - It initializes the profile using the provided phone number ID and access token,
//     internally calling [Default] to set up the [Client].
func NewProfile(
	ctx context.Context,
	phoneNumberId,
	accessToken string,
) *Profile {
	return &Profile{
		PhoneNumberID: phoneNumberId,
		client:        Default(phoneNumberId, accessToken),
	}
}

// InfoResponse represents the API response containing a list of [InfoData].
type InfoResponse struct {
	Data []InfoData `json:"data"`
}

// InfoData contains the details of a WhatsApp business profile, such as email, address, and description.
type InfoData struct {
	About        string   `json:"about"`
	Address      string   `json:"address"`
	Email        string   `json:"email"`
	Websites     []string `json:"websites"`
	Description  string   `json:"description"`
	ProfilePhoto string   `json:"profile_picture_url"`
	Vertical     string   `json:"vertical"`
	Name         string   `json:"name,omitempty"`
}

// Info retrieves the business profile information and returns an [InfoData].
//   - It fetches details including the about text, address, description, email, profile picture URL, websites, and vertical.
func (p *Profile) Info(accountName string) (InfoData, error) {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	fields := "about,address,description,email,profile_picture_url,websites,vertical"
	fieldsQueryParam := &xhttp.Data{Key: "fields", Value: fields}

	res, err := p.client.Get(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		QueryParams: []xhttp.HTTPData{
			fieldsQueryParam,
		},
	})
	if err != nil {
		return InfoData{}, err
	}
	defer res.Body.Close()

	var whatsInfoRes InfoResponse
	if err := xjson.Decode(res.Body, &whatsInfoRes); err != nil {
		return InfoData{}, err
	}

	data := InfoData{}
	if len(whatsInfoRes.Data) > 0 {
		data = whatsInfoRes.Data[0]
		data.Name = accountName
	}
	return data, nil
}

// ProfilePictureData represents the data returned after successfully updating a profile picture.
type ProfilePictureData struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// ChangeProfilePicture updates the business profile picture.
//   - It first uploads the media file using [UploadTemplateFile], and then assigns it to the profile.
func (p *Profile) ChangeProfilePicture(
	appId string,
	mediaFile *multipart.FileHeader,
) (ProfilePictureData, error) {
	fileHandle, err := UploadTemplateFile(p.client, appId, mediaFile)
	if err != nil {
		return ProfilePictureData{}, err
	}

	fileHandleId := fileHandle.Id
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product":      "whatsapp",
		"profile_picture_handle": fileHandleId,
	}.Buffer()

	_, err = p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})

	data := ProfilePictureData{
		Name: fileHandle.fileName,
		Size: fileHandle.fileSize,
		Type: fileHandle.mimeType,
	}
	return data, err
}

// ChangeAbout updates the "about" text of the business profile.
func (p *Profile) ChangeAbout(about string) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"about":             about,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ChangeDescription updates the description of the business profile.
func (p *Profile) ChangeDescription(description string) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"description":       description,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ChangeWebsites updates the list of websites associated with the business profile.
func (p *Profile) ChangeWebsites(websites []string) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"websites":          websites,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ChangeEmail updates the contact email of the business profile.
func (p *Profile) ChangeEmail(email string) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"email":             email,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ChangeAddress updates the physical address of the business profile.
func (p *Profile) ChangeAddress(address string) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"address":           address,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ChangeVertical updates the industry vertical category of the business profile using a [ProfileVertical].
func (p *Profile) ChangeVertical(vertical ProfileVertical) error {
	url := p.client.Endpoint(p.PhoneNumberID + "/_business_profile")
	body := xjson.JSON{
		"messaging_product": "whatsapp",
		"vertical":          vertical,
	}.Buffer()

	_, err := p.client.Post(url, &xhttp.Options{
		Headers: p.client.Headers("application/json"),
		Body:    body,
	})
	return err
}

// ProfileVertical represents the industry vertical category for a business profile.
type ProfileVertical string

// NewProfileVertical creates and returns a new [ProfileVertical] from the provided string v.
//   - If the given string does not match any valid [ProfileVertical], it defaults to returning [OTHER].
func NewProfileVertical(v string) ProfileVertical {
	if !IsValidProfileVertical(v) {
		return OTHER
	}
	return ProfileVertical(v)
}

const (
	// ALCOHOL represents the alcohol industry [ProfileVertical].
	ALCOHOL ProfileVertical = "ALCOHOL"
	// APPAREL represents the apparel and clothing [ProfileVertical].
	APPAREL ProfileVertical = "APPAREL"
	// AUTO represents the automotive industry [ProfileVertical].
	AUTO ProfileVertical = "AUTO"
	// BEAUTY represents the beauty and personal care [ProfileVertical].
	BEAUTY ProfileVertical = "BEAUTY"
	// EDU represents the education sector [ProfileVertical].
	EDU ProfileVertical = "EDU"
	// ENTERTAIN represents the entertainment industry [ProfileVertical].
	ENTERTAIN ProfileVertical = "ENTERTAIN"
	// EVENT_PLAN represents the event planning [ProfileVertical].
	EVENT_PLAN ProfileVertical = "EVENT_PLAN"
	// FINANCE represents the financial services [ProfileVertical].
	FINANCE ProfileVertical = "FINANCE"
	// GOVT represents government and public administration [ProfileVertical].
	GOVT ProfileVertical = "GOVT"
	// GROCERY represents the grocery and food retail [ProfileVertical].
	GROCERY ProfileVertical = "GROCERY"
	// HEALTH represents the healthcare and medical [ProfileVertical].
	HEALTH ProfileVertical = "HEALTH"
	// HOTEL represents the hotel and lodging [ProfileVertical].
	HOTEL ProfileVertical = "HOTEL"
	// NONPROFIT represents non-profit organizations [ProfileVertical].
	NONPROFIT ProfileVertical = "NONPROFIT"
	// ONLINE_GAMBLING represents the online gambling [ProfileVertical].
	ONLINE_GAMBLING ProfileVertical = "ONLINE_GAMBLING"
	// OTC_DRUGS represents over-the-counter drugs and pharmacies [ProfileVertical].
	OTC_DRUGS ProfileVertical = "OTC_DRUGS"
	// OTHER represents any other unspecified [ProfileVertical].
	OTHER ProfileVertical = "OTHER"
	// PHYSICAL_GAMBLING represents physical gambling and casinos [ProfileVertical].
	PHYSICAL_GAMBLING ProfileVertical = "PHYSICAL_GAMBLING"
	// PROF_SERVICES represents professional services [ProfileVertical].
	PROF_SERVICES ProfileVertical = "PROF_SERVICES"
	// RESTAURANT represents the restaurant and food service [ProfileVertical].
	RESTAURANT ProfileVertical = "RESTAURANT"
	// RETAIL represents general retail [ProfileVertical].
	RETAIL ProfileVertical = "RETAIL"
	// TRAVEL represents the travel and tourism [ProfileVertical].
	TRAVEL ProfileVertical = "TRAVEL"
)

// IsValidProfileVertical checks whether the provided string v corresponds to a supported [ProfileVertical].
//   - It returns true if it is a valid vertical, or false otherwise.
func IsValidProfileVertical(v string) bool {
	switch ProfileVertical(v) {
	case ALCOHOL,
		APPAREL,
		AUTO,
		BEAUTY,
		EDU,
		ENTERTAIN,
		EVENT_PLAN,
		FINANCE,
		GOVT,
		GROCERY,
		HEALTH,
		HOTEL,
		NONPROFIT,
		ONLINE_GAMBLING,
		OTC_DRUGS,
		OTHER,
		PHYSICAL_GAMBLING,
		PROF_SERVICES,
		RESTAURANT,
		RETAIL,
		TRAVEL:
		return true
	default:
		return false
	}
}

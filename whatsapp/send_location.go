package whatsapp

// Location represents a geographic location with its coordinates and optional descriptive metadata.
type Location struct {
	// Latitude is the north-south coordinate of the location.
	Latitude float64 `json:"latitude"`
	// Longitude is the east-west coordinate of the location.
	Longitude float64 `json:"longitude"`
	// Name is the name of the location (e.g., "Googleplex").
	Name string `json:"name,omitempty"`
	// Address is the physical address of the location.
	Address string `json:"address,omitempty"`
}

// SendLocation sends a location message through the WhatsApp API.
// 	- It receives a [Header] and a [Location] as parameters.
// 	- Returns the ID of the sent message or an error if the request fails.
func (api *Client) SendLocation(
	h Header,
	location Location,
) (id string, err error) {
	locationType := "location"
	h["type"] = locationType
	h[locationType] = location

	res, err := MessagesEndpointRequest(api, h.Bytes())
	return res.FirstId(), err
}

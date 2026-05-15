package whatsapp

import (
	"sort"

	"github.com/atendi9/meta/xhttp/xjson"
)

// InteractiveListRow represents an individual row item within an interactive list section.
type InteractiveListRow struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SendInteractiveListOpts contains the configuration and content for sending an interactive list message.
type SendInteractiveListOpts struct {
	Header     string                          `json:"header"`
	Message    string                          `json:"message"`
	Footer     string                          `json:"footer"`
	ButtonText string                          `json:"btnTxt"`
	Rows       map[string][]InteractiveListRow `json:"rows"`
}

// SendInteractiveList sends an interactive list message through the WhatsApp API.
//   - It receives a [Header] and [SendInteractiveListOpts] as parameters.
//   - Returns the ID of the sent message or an error if the request fails.
func (api *Client) SendInteractiveList(
	h Header,
	opts SendInteractiveListOpts,
) (id string, err error) {
	// Iterate the section titles in a deterministic (sorted) order, since
	// ranging over a map yields a non-deterministic order between sends.
	titles := make([]string, 0, len(opts.Rows))
	for title := range opts.Rows {
		titles = append(titles, title)
	}
	sort.Strings(titles)
	sections := []xjson.JSON{}
	for _, title := range titles {
		sections = append(sections, xjson.JSON{
			"title": title,
			"rows":  opts.Rows[title],
		})
	}
	interactiveType := "interactive"
	h["type"] = interactiveType
	h[interactiveType] = xjson.JSON{
		"type": "list",
		"header": xjson.JSON{
			"type": "text",
			"text": opts.Header,
		},
		"body": xjson.JSON{
			"text": opts.Message,
		},
		"footer": xjson.JSON{
			"text": opts.Footer,
		},
		"action": xjson.JSON{
			"button":   opts.ButtonText,
			"sections": sections,
		},
	}

	res, err := MessagesEndpointRequest(api, h.Bytes())
	return res.FirstId(), err
}

package whatsapp

import "github.com/atendi9/meta/xhttp/xjson"

// AuthTemplate generates an authentication template [Message] to be sent via WhatsApp.
//   - It constructs the payload using the provided template name, recipient number (to [string]),
//     authentication token [string], and language [Lang]. The resulting message utilizes [xjson.JSON]
//     to define the template components, including a text body and a URL button parameter
//     that both contain the provided token.
func AuthTemplate(
	name string,
	to string,
	token string,
	lang Lang,
) Message {
	h := MessageHeader(to, "template")
	authTemplate := xjson.JSON{
		"name":     name,
		"language": xjson.JSON{"code": lang},
		"components": []xjson.JSON{
			{
				"type": "body",
				"parameters": []xjson.JSON{
					{"type": "text", "text": token},
				},
			},
			{
				"type":     "button",
				"sub_type": "url",
				"index":    0,
				"parameters": []xjson.JSON{
					{"type": "text", "text": token},
				},
			},
		},
	}
	return MessageTemplate(h, authTemplate)
}

package whatsapp

import (
	"regexp"

	"github.com/atendi9/meta/xhttp/xjson"
)

// bsuidPattern matches a Business-Scoped User ID (BSUID) and its parent
// variant. A BSUID is the user's ISO 3166 alpha-2 country code, a dot, and up
// to 128 alphanumeric characters (for example "US.13491208655302741918").
// Parent BSUIDs insert an "ENT" segment between the country code and the
// identifier (for example "US.ENT.11815799212886844830").
var bsuidPattern = regexp.MustCompile(`^[A-Za-z]{2}\.(ENT\.)?[A-Za-z0-9]{1,128}$`)

// IsBSUID reports whether identifier follows the BSUID format. Phone numbers,
// which contain only digits and an optional leading "+", never match, so the
// distinction between a recipient phone number and a BSUID is unambiguous.
func IsBSUID(identifier string) bool {
	return bsuidPattern.MatchString(identifier)
}

// SetRecipient writes identifier to the correct WhatsApp payload field. A
// BSUID is assigned to "recipient"; any other value is treated as a phone
// number and assigned to "to". Callers that already pass phone numbers keep an
// identical payload, so existing integrations require no changes.
func SetRecipient(payload xjson.JSON, identifier string) {
	if IsBSUID(identifier) {
		payload["recipient"] = identifier
		return
	}
	payload["to"] = identifier
}

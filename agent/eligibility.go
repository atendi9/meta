package agent

import (
	"github.com/atendi9/meta/xhttp"
	"github.com/atendi9/meta/xhttp/xjson"
)

type Eligibility struct {
	IsEligible bool `json:"is_eligible"`
}

type Onboard struct {
	Client *Client
	Config Config
}

// Eligibility check whether a WhatsApp Business phone number can use Meta Business Agent.
func (o *Onboard) Eligibility() (bool, error) {
	url := o.Config.URL("/agent_eligibility")
	headers := o.Client.Headers("application/json")
	res, err := o.Client.Get(url, &xhttp.Options{
		Headers: headers,
	})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	var result Eligibility
	if err := xjson.Decode(res.Body, &result); err != nil {
		return false, err
	}
	return result.IsEligible, nil
}

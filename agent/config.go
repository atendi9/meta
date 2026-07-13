package agent

import (
	"fmt"
)

type Config struct {
	EntityID   string
	BaseURL    string
	ApiVersion string
}

func (c *Config) URL(endpoint string) string {
	url := fmt.Sprintf("%s/%s", c.BaseURL, c.EntityID+endpoint)
	return url
}

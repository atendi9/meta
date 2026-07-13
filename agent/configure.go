package agent

type Configurator struct{ Client *Client }

func Configure(c *Client) *Configurator {
	return &Configurator{
		Client: c,
	}
}

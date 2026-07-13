package agent

// Operator groups the runtime "operate" APIs used once an agent is live in
// conversations: thread control, agent events, agent testing and agent
// evaluation. Unlike the configuration APIs, these act on active conversations
// and evaluation jobs rather than on the agent knowledge base.
type Operator struct {
	o *Onboard
}

// Operate exposes the operate APIs for the configured entity.
func (c *Configurator) Operate() *Operator {
	return &Operator{o: c.Client.Onboard}
}

package channels

type baseChannel struct {
	protocol  string
	SendEvent func(event string, payload interface{}) error
}

func (c *baseChannel) SetEventHandler(handler func(event string, payload interface{}) error) {
	c.SendEvent = handler
}

func (c *baseChannel) GetProtocol() string {
	return c.protocol
}

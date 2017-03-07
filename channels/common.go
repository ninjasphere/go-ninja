package channels

type baseChannel struct {
	protocol  string
	SendEvent func(event string, payload ...interface{}) error
}

func (c *baseChannel) SetEventHandler(handler func(event string, payload ...interface{}) error) {
	c.SendEvent = handler
}

func (c *baseChannel) GetProtocol() string {
	return c.protocol
}

type BaseChannel struct {
	Protocol  string
	SendEvent func(event string, payload ...interface{}) error
}

func (c *BaseChannel) SetEventHandler(handler func(event string, payload ...interface{}) error) {
	c.SendEvent = handler
}

func (c *BaseChannel) GetProtocol() string {
	return c.Protocol
}

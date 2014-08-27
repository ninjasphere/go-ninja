package channels

type baseChannel struct {
	SendEvent func(event string, payload interface{}) error
}

func (c *baseChannel) SetEventHandler(handler func(event string, payload interface{}) error) {
	c.SendEvent = handler
}

package channels

type IdentifyDevice interface {
	Identify() error
}

type IdentifyChannel struct {
	baseChannel
	device IdentifyDevice
}

func NewIdentifyChannel(device IdentifyDevice) *IdentifyChannel {
	return &IdentifyChannel{baseChannel{
		protocol: "identify",
	}, device}
}

func (c *IdentifyChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

func (c *IdentifyChannel) Identify() error {
	return c.device.Identify()
}

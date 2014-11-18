package channels

type IdentifyDevice interface {
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

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

func (c *IdentifyChannel) Identify() error {
	return c.device.Identify()
}

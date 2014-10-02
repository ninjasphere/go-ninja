package channels

type IlluminanceDevice interface {
}

type IlluminanceChannel struct {
	baseChannel
	device IlluminanceDevice
}

func NewIlluminanceChannel(device IlluminanceDevice) *IlluminanceChannel {
	return &IlluminanceChannel{baseChannel{
		protocol: "illuminance",
	}, device}
}

func (c *IlluminanceChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

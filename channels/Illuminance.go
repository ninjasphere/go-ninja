package channels

type IlluminanceChannel struct {
	baseChannel
}

func NewIlluminanceChannel() *IlluminanceChannel {
	return &IlluminanceChannel{baseChannel{
		protocol: "illuminance",
	}}
}

func (c *IlluminanceChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

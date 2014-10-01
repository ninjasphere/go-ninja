package channels

type BrightnessDevice interface {
	SetBrightness(state float64) error
}

type BrightnessChannel struct {
	baseChannel
	device BrightnessDevice
}

func NewBrightnessChannel(device BrightnessDevice) *BrightnessChannel {
	return &BrightnessChannel{baseChannel{
		protocol: "brightness",
	}, device}
}

func (c *BrightnessChannel) Set(state float64) error {
	c.device.SetBrightness(state)
	return nil
}

func (c *BrightnessChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

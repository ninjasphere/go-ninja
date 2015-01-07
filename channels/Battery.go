package channels

type BatteryDevice interface {
}

type BatteryChannel struct {
	baseChannel
	device BatteryDevice
}

func NewBatteryChannel(device BatteryDevice) *BatteryChannel {
	return &BatteryChannel{baseChannel{
		protocol: "battery",
	}, device}
}

func (c *BatteryChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

func (c *BatteryChannel) SendWarning() error {
	return c.SendEvent("warning")
}

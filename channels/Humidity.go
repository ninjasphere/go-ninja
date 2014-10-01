package channels

type HumidityDevice interface {
	SetHumidity(state float64) error
}

type HumidityChannel struct {
	baseChannel
	device HumidityDevice
}

func NewHumidityChannel(device HumidityDevice) *HumidityChannel {
	return &HumidityChannel{baseChannel{
		protocol: "humidity",
	}, device}
}

func (c *HumidityChannel) Set(state float64) error {
	c.device.SetHumidity(state)
	return nil
}

func (c *HumidityChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

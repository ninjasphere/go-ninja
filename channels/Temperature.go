package channels

type TemperatureDevice interface {
	SetTemperature(state float64) error
}

type TemperatureChannel struct {
	baseChannel
	device TemperatureDevice
}

func NewTemperatureChannel(device TemperatureDevice) *TemperatureChannel {
	return &TemperatureChannel{baseChannel{
		protocol: "humidity",
	}, device}
}

func (c *TemperatureChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

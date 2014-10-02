package channels

type EnergyDevice interface {
}

type EnergyChannel struct {
	baseChannel
	device EnergyDevice
}

func NewEnergyChannel(device EnergyDevice) *EnergyChannel {
	return &EnergyChannel{baseChannel{
		protocol: "energy",
	}, device}
}

func (c *EnergyChannel) SendState(state float64) error {
	return c.SendEvent("state", state)
}

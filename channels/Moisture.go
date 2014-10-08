package channels

type MoistureDevice interface {
}

type MoistureChannel struct {
  baseChannel
  device MoistureDevice
}

func NewMoistureChannel(device MoistureDevice) *MoistureChannel {
  return &MoistureChannel{baseChannel{
    protocol: "moisture",
  }, device}
}

func (c *MoistureChannel) SendState(state float64) error {
  return c.SendEvent("state", state)
}

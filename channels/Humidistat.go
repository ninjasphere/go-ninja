package channels

type HumidiStatActuator interface {
	SetHumiditySetPoint(float64) error
}

type HumidiStatChannel struct {
	baseChannel
	actuator HumidiStatActuator
}

func NewHumidiStatChannel(actuator HumidiStatActuator) *HumidiStatChannel {
	return &HumidiStatChannel{
		baseChannel: baseChannel{protocol: "humidistat"},
		actuator:    actuator,
	}
}

func (c *HumidiStatChannel) Set(setPoint *float64) error {
	return c.actuator.SetHumiditySetPoint(*setPoint)
}

func (c *HumidiStatChannel) SendState(setPoint float64) error {
	return c.SendEvent("state", setPoint)
}

package channels

type ThermoStatActuator interface {
	SetTemperatureSetPoint(float64) error
}

type ThermoStatChannel struct {
	baseChannel
	actuator ThermoStatActuator
}

func NewThermoStatChannel(actuator ThermoStatActuator) *ThermoStatChannel {
	return &ThermoStatChannel{
		baseChannel: baseChannel{protocol: "thermostat"},
		actuator:    actuator,
	}
}

func (c *ThermoStatChannel) Set(setPoint *float64) error {
	return c.actuator.SetTemperatureSetPoint(*setPoint)
}

func (c *ThermoStatChannel) SendState(setPoint float64) error {
	return c.SendEvent("state", setPoint)
}

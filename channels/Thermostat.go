package channels

type ThermoStatActuator interface {
	SetTemperatureSetPoint(float64) error
}

type ThermoStatState struct {
	Target *float64 `json:"target,omitempty"`
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

func (c *ThermoStatChannel) Set(state *ThermoStatState) error {
	if state != nil && state.Target != nil {
		return c.actuator.SetTemperatureSetPoint(*state.Target)
	} else {
		return nil
	}
}

func (c *ThermoStatChannel) SendState(state *ThermoStatState) error {
	return c.SendEvent("state", state)
}

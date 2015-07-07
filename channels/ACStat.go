package channels

type ACState struct {
	Mode           *string   `json:"mode,omitempty"`
	SupportedModes *[]string `json:"supported-modes,omitempty"`
}

type ACStatActuator interface {
	SetACState(acState *ACState) error
}

type ACStatChannel struct {
	baseChannel
	actuator ACStatActuator
}

func NewACStatChannel(actuator ACStatActuator) *ACStatChannel {
	return &ACStatChannel{
		baseChannel: baseChannel{protocol: "acstat"},
		actuator:    actuator,
	}
}

func (c *ACStatChannel) Set(acState *ACState) error {
	return c.actuator.SetACState(acState)
}

func (c *ACStatChannel) SendState(acState *ACState) error {
	return c.SendEvent("state", acState)
}

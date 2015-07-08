package channels

var (
	MODE_OFF  = "off"
	MODE_COOL = "cool"
	MODE_FAN  = "fan"
	MODE_DRY  = "dry"
	MODE_HEAT = "heat"
	ALL_MODES = []string{MODE_OFF, MODE_COOL, MODE_FAN, MODE_DRY, MODE_HEAT}
)

type ACState struct {
	Mode           *string  `json:"mode,omitempty"`
	SupportedModes []string `json:"supported-modes,omitempty"`
}

type ACStatActuator interface {
	SetACState(acState *ACState) error
}

type ACStatChannel struct {
	baseChannel
	actuator ACStatActuator
}

func NewACState() *ACState {
	return &ACState{
		Mode:           &MODE_OFF,
		SupportedModes: ALL_MODES,
	}
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

package channels

type DemandControl struct {
	Enabled  *bool       `json:"enabled,omitempty"`
	Goal     *float64    `json:"goal,omitempty"`
	Strategy interface{} `json:"strategy,omitempty"`
}

type DemandStatActuator interface {
	SetDemandControl(demandControl *DemandControl) error
}

type DemandStatChannel struct {
	baseChannel
	actuator DemandStatActuator
}

func NewDemandStatChannel(actuator DemandStatActuator) *DemandStatChannel {
	return &DemandStatChannel{
		baseChannel: baseChannel{protocol: "demandstat"},
		actuator:    actuator,
	}
}

func (c *DemandStatChannel) Set(demandControl *DemandControl) error {
	return c.actuator.SetDemandControl(demandControl)
}

func (c *DemandStatChannel) SendState(demandControl *DemandControl) error {
	return c.SendEvent("state", demandControl)
}

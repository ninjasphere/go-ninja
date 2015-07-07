package channels

type OpenLoopState struct {
	Enabled         *bool `json:"enabled,omitempty"`
	ReapplyInterval *int  `json:"reapply-interval,omitempty"`
}

type OpenLoopActuator interface {
	SetOpenLoopState(openLoop *OpenLoopState) error
}

type OpenLoopChannel struct {
	baseChannel
	actuator OpenLoopActuator
}

func NewOpenLoopChannel(actuator OpenLoopActuator) *OpenLoopChannel {
	return &OpenLoopChannel{
		baseChannel: baseChannel{protocol: "openloop"},
		actuator:    actuator,
	}
}

func (c *OpenLoopChannel) Set(openLoop *OpenLoopState) error {
	return c.actuator.SetOpenLoopState(openLoop)
}

func (c *OpenLoopChannel) SendState(openLoop *OpenLoopState) error {
	return c.SendEvent("state", openLoop)
}

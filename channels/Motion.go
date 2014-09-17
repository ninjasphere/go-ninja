package channels

type MotionDevice interface {
}

type MotionState struct {
	OnOff *bool `json:"onoff,omitempty"`
}

type MotionChannel struct {
	baseChannel
	device MotionDevice
}

func NewMotionChannel(device MotionDevice) *MotionChannel {
	return &MotionChannel{baseChannel{}, device}
}

func (c *MotionChannel) SendState(onoff *bool) error {
	return c.SendEvent("state", &MotionState{
		OnOff: onoff,
	})
}

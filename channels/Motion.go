package channels

type MotionChannel struct {
	baseChannel
}

func NewMotionChannel() *MotionChannel {
	return &MotionChannel{baseChannel{
		protocol: "motion",
	}}
}

func (c *MotionChannel) SendMotion() error {
	return c.SendEvent("state", true)
}

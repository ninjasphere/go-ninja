package channels

type PresenceChannel struct {
	baseChannel
}

func NewPresenceChannel() *PresenceChannel {
	return &PresenceChannel{baseChannel{
		protocol: "presence",
	}}
}

func (c *PresenceChannel) SendState(state bool) error {
	return c.SendEvent("state", state)
}

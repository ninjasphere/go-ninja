package channels

type OnOffDevice interface {
	ToggleOnOff() error
	SetOnOff(state bool) error
}

// An OnOffChannel can be added to devices, exposing http://schema.ninjablocks.com/protocol/on-off
type OnOffChannel struct {
	baseChannel
	device OnOffDevice
}

func NewOnOffChannel(device OnOffDevice) *OnOffChannel {
	return &OnOffChannel{baseChannel{
		protocol: "on-off",
	}, device}
}

func (c *OnOffChannel) TurnOn() error {
	return c.device.SetOnOff(true)
}

func (c *OnOffChannel) TurnOff() error {
	return c.device.SetOnOff(false)
}

func (c *OnOffChannel) Toggle() error {
	return c.device.ToggleOnOff()
}

func (c *OnOffChannel) Set(state *bool) error {
	return c.device.SetOnOff(*state)
}

func (c *OnOffChannel) SendState(on bool) error {
	return c.SendEvent("state", on)
}

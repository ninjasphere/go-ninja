package channels

type OnOffDevice interface {
	ToggleOnOff() error
	SetOnOff(state bool) error
}

// An OnOffChannel can be added to devices, exposing http://schemas.ninjablocks.com/protocol/on-off
type OnOffChannel struct {
	baseChannel
	device OnOffDevice
}

func NewOnOffChannel(device OnOffDevice) *OnOffChannel {
	return &OnOffChannel{baseChannel{}, device}
}

func (c *OnOffChannel) TurnOn(_, reply *interface{}) error {
	return c.Set(true, reply)
}

func (c *OnOffChannel) TurnOff(_, reply *interface{}) error {
	return c.Set(false, reply)
}

func (c *OnOffChannel) Toggle(_, reply *interface{}) error {
	return c.device.ToggleOnOff()
}

func (c *OnOffChannel) Set(state bool, reply *interface{}) error {
	return c.device.SetOnOff(state)
}

package channels

import "github.com/ninjasphere/go-ninja/rpc"

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
	return &OnOffChannel{baseChannel{
		protocol: "on-off",
	}, device}
}

func (c *OnOffChannel) TurnOn(message *rpc.Message) error {
	return c.device.SetOnOff(true)
}

func (c *OnOffChannel) TurnOff(message *rpc.Message) error {
	return c.device.SetOnOff(false)
}

func (c *OnOffChannel) Toggle(message *rpc.Message) error {
	return c.device.ToggleOnOff()
}

func (c *OnOffChannel) Set(message *rpc.Message, state *bool) error {
	return c.device.SetOnOff(*state)
}

func (c *OnOffChannel) SendState(on bool) error {
	return c.SendEvent("state", on)
}

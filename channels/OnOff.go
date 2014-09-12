package channels

import "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

type OnOffDevice interface {
	ToggleOnOff() error
	SetOnOff(state bool) error
}

// An OnOffChannel can be added to devices, exposing http://schemas.ninjablocks.com/protocol/on-off
type OnOffChannel struct {
	baseChannel
	device OnOffDevice
}

type OnOffState struct {
	OnOff *bool `json:"onoff,omitempty"`
}

func NewOnOffChannel(device OnOffDevice) *OnOffChannel {
	return &OnOffChannel{baseChannel{}, device}
}

func (c *OnOffChannel) TurnOn(message mqtt.Message, _, reply *interface{}) error {
	return c.device.SetOnOff(true)
}

func (c *OnOffChannel) TurnOff(message mqtt.Message, _, reply *interface{}) error {
	return c.device.SetOnOff(false)
}

func (c *OnOffChannel) Toggle(message mqtt.Message, _, reply *interface{}) error {
	return c.device.ToggleOnOff()
}

func (c *OnOffChannel) Set(message mqtt.Message, state *bool, reply *interface{}) error {
	return c.device.SetOnOff(*state)
}

func (c *OnOffChannel) SendState(on *bool) error {
	return c.SendEvent("state", &OnOffState{
		OnOff: on,
	})
}

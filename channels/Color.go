package channels

import (
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/davecgh/go-spew/spew"
)

type ColorDevice interface {
	SetColor(state *ColorState) error
}

type ColorState struct {
	Mode        string   `json:"mode,omitempty"`
	Hue         *float64 `json:"hue,omitempty"`
	Saturation  *float64 `json:"saturation,omitempty"`
	Temperature *int     `json:"temperature,omitempty"`
	X           *float64 `json:"x,omitempty"`
	Y           *float64 `json:"y,omitempty"`
}

type ColorChannel struct {
	baseChannel
	device ColorDevice
}

func NewColorChannel(device ColorDevice) *ColorChannel {
	return &ColorChannel{baseChannel{}, device}
}

func (c *ColorChannel) Set(message mqtt.Message, state *ColorState, reply *interface{}) error {
	spew.Dump("Setting colour state", state)
	return c.device.SetColor(state)
}

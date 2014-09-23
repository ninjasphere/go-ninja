package channels

import "github.com/ninjasphere/go-ninja/rpc"

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

func (c *ColorChannel) Set(message *rpc.Message, state *ColorState, reply *interface{}) error {
	return c.device.SetColor(state)
}

package channels

import "github.com/ninjasphere/go-ninja/rpc"

type BrightnessDevice interface {
	SetBrightness(state float64) error
}

type BrightnessChannel struct {
	baseChannel
	device BrightnessDevice
}

func NewBrightnessChannel(device BrightnessDevice) *BrightnessChannel {
	return &BrightnessChannel{baseChannel{}, device}
}

func (c *BrightnessChannel) Set(message *rpc.Message, state *float64, reply *interface{}) error {
	c.device.SetBrightness(*state)
	return nil
}

package channels

import "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

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

func (c *BrightnessChannel) Set(message mqtt.Message, state *float64, reply *interface{}) error {
	c.device.SetBrightness(*state)
	return nil
}

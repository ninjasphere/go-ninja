package channels

import "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

type TransitionDevice interface {
	SetTransition(state int) error
}

type TransitionChannel struct {
	baseChannel
	device TransitionDevice
}

func NewTransitionChannel(device TransitionDevice) *TransitionChannel {
	return &TransitionChannel{baseChannel{}, device}
}

func (c *TransitionChannel) Set(message mqtt.Message, state *int, reply *interface{}) error {
	c.device.SetTransition(*state)
	return nil
}

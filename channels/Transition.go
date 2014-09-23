package channels

import "github.com/ninjasphere/go-ninja/rpc"

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

func (c *TransitionChannel) Set(message *rpc.Message, state *int, reply *interface{}) error {
	c.device.SetTransition(*state)
	return nil
}

package channels

import (
	"log"
)

type FanState struct {
	Speed     *float64 `json:"speed,omitempty"`     // the speed of the fan as a percentage of maximum
	Direction *string  `json:"direction,omitempty"` // the direction of the fan: "forward" or "reverse"
}

type FanStatActuator interface {
	SetFanState(fanState *FanState) error
}

type FanStatChannel struct {
	baseChannel
	actuator FanStatActuator
}

func NewFanStatChannel(actuator FanStatActuator) *FanStatChannel {
	return &FanStatChannel{
		baseChannel: baseChannel{protocol: "fanstat"},
		actuator:    actuator,
	}
}

func (c *FanStatChannel) Set(fanState *FanState) error {
	return c.actuator.SetFanState(fanState)
}

func (c *FanStatChannel) SendState(fanState *FanState) error {
	log.Printf("SendState: %+v\n, %p", fanState, c.SendEvent)
	return c.SendEvent("state", fanState)
}

package sunspec

import (
	"github.com/ninjasphere/go-ninja/channels"
)

type Model1 struct {
	Mn  string `json:"Mn,omitempty"`
	Md  string `json:"Md,omitempty"`
	Opt string `json:"Opt,omitempty"`
	Vr  string `json:"Vr,omitempty"`
	SN  string `json:"SN,omitempty"`
	DA  uint16 `json:"DA,omitempty"`
}

type Model1Channel struct {
	channels.BaseChannel
}

func NewModel1Channel() *Model1Channel {
	return &Model1Channel{channels.BaseChannel{
		Protocol: "sunspec/model1",
	}}
}

func (c *Model1Channel) SendState(state *Model1) error {
	return c.SendEvent("state", state)
}

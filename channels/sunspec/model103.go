package sunspec

import (
	"github.com/ninjasphere/go-ninja/channels"
)

type Model103 struct {
	// Model 101, 102, 103

	A       float64 `json:"A"`
	AphA    float64 `json:"AphA"`
	PhVphA  float64 `json:"PhVphA"`
	W       float64 `json:"W"`
	Hz      float64 `json:"Hz"`
	VA      float64 `json:"VA"`
	VAr     float64 `json:"VAr"`
	PF      float64 `json:"PF"`
	WH      float64 `json:"WH"`
	DCA     float64 `json:"DCA"`
	DCV     float64 `json:"DCV"`
	DCW     float64 `json:"DCW"`
	TmpCab  float64 `json:"TmpCab"`
	TmpOt   float64 `json:"TmpOt"`
	TmpSnk  float64 `json:"TmpSnk"`
	TmpTrns float64 `json:"TmpTrns"`
	St      uint16  `json:"St"`
	StVnd   uint16  `json:"StVnd"`
	Evt1    uint32  `json:"Evt1"`
	Evt2    uint32  `json:"Evt2"`
	EvtVnd1 uint32  `json:"EvtVnd1"`
	EvtVnd2 uint32  `json:"EvtVnd2"`
	EvtVnd3 uint32  `json:"EvtVnd3"`
	EvtVnd4 uint32  `json:"EvtVnd4"`

	// Model 102, 103

	AphB   *float64 `json:"AphB,omitempty"`
	PhVphB *float64 `json:"PhVphB,omitempty"`

	// Model 103

	AphC   *float64 `json:"AphC,omitempty"`
	PhVphC *float64 `json:"PhVphC,omitempty"`

	PPVphAB *float64 `json:"PPVphAB,omitempty"`
	PPVphBC *float64 `json:"PPVphBC,omitempty"`
	PPVphCA *float64 `json:"PPVphCA,omitempty"`
}

type Model103Channel struct {
	channels.BaseChannel
}

func NewModel103Channel() *Model103Channel {
	return &Model103Channel{channels.BaseChannel{
		Protocol: "sunspec/model103",
	}}
}

func (c *Model103Channel) SendState(state *Model103) error {
	return c.SendEvent("state", state)
}

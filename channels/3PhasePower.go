package channels

type ThreePhasePower struct {
	ActiveA       float64 `json:"activeA"`
	ActiveB       float64 `json:"activeB"`
	ActiveC       float64 `json:"activeC"`
	ActiveTotal   float64 `json:"activeTotal"`
	ReactiveA     float64 `json:"reactiveA"`
	ReactiveB     float64 `json:"reactiveB"`
	ReactiveC     float64 `json:"reactiveC"`
	ReactiveTotal float64 `json:"reactiveTotal"`
	ApparentA     float64 `json:"apparentA"`
	ApparentB     float64 `json:"apparentB"`
	ApparentC     float64 `json:"apparentC"`
	ApparentTotal float64 `json:"apparentTotal"`
	PfA           float64 `json:"pfA"`
	PfB           float64 `json:"pfB"`
	PfC           float64 `json:"pfC"`
	PfAverage     float64 `json:"pfAverage"`
}

type ThreePhasePowerChannel struct {
	baseChannel
}

func NewThreePhasePowerChannel() *ThreePhasePowerChannel {
	return &ThreePhasePowerChannel{baseChannel{
		protocol: "3-phase-power",
	}}
}

func (c *ThreePhasePowerChannel) SendState(state *ThreePhasePower) error {
	return c.SendEvent("state", state)
}

package channels

type DemandState struct {
	Rated        *float64 `json:"rated,omitempty"`        // rated maximum power, in Watts
	FullLoad     *float64 `json:"fullload,omitempty"`     // historical maximum load (may be higher than rated load)
	Current      *float64 `json:"current,omitempty"`      // average power for current period
	Peak         *float64 `json:"peak,omitempty"`         // peak instantaneous power in averaging period
	Goal         *float64 `json:"goal,omitempty"`         // goal power for averaging period
	Controlled   *float64 `json:"controlled,omitempty"`   // average controlled power
	Uncontrolled *float64 `json:"uncontrolled,omitempty"` // average uncontrolled power
	Period       *int     `json:"period,omitempty"`       // averaging period, in secon
}

type DemandChannel struct {
	baseChannel
}

func NewDemandChannel() *DemandChannel {
	return &DemandChannel{
		baseChannel: baseChannel{protocol: "demand"},
	}
}

func (c *DemandChannel) SendState(demandState *DemandState) error {
	return c.SendEvent("state", demandState)
}

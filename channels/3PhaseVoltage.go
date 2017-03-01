package channels

type ThreePhaseVoltage struct {
	VoltageAN    float64 `json:"voltageAN"`
	VoltageBN    float64 `json:"voltageBN"`
	VoltageCN    float64 `json:"voltageCN"`
	VoltageLNAvg float64 `json:"voltageLNAvg"`
	VoltageAB    float64 `json:"voltageAB"`
	VoltageBC    float64 `json:"voltageBC"`
	VoltageCA    float64 `json:"voltageCA"`
	VoltageLLAvg float64 `json:"voltageLLAvg"`
}

type ThreePhaseVoltageChannel struct {
	baseChannel
}

func NewThreePhaseVoltageChannel() *ThreePhaseVoltageChannel {
	return &ThreePhaseVoltageChannel{baseChannel{
		protocol: "3-phase-voltage",
	}}
}

func (c *ThreePhaseVoltageChannel) SendState(state *ThreePhaseVoltage) error {
	return c.SendEvent("state", state)
}

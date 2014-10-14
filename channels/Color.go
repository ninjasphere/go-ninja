package channels

type ColorDevice interface {
	SetColor(state *ColorState) error
}

type ColorState struct {
	Mode        string   `json:"mode,omitempty"`
	Hue         *float64 `json:"hue,omitempty"`
	Saturation  *float64 `json:"saturation,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	X           *float64 `json:"x,omitempty"`
	Y           *float64 `json:"y,omitempty"`
}

type ColorChannel struct {
	baseChannel
	device ColorDevice
}

func NewColorChannel(device ColorDevice) *ColorChannel {
	return &ColorChannel{baseChannel{
		protocol: "color",
	}, device}
}

func (c *ColorChannel) Set(state *ColorState) error {
	return c.device.SetColor(state)
}

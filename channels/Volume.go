package channels

import "github.com/ninjasphere/go-ninja/rpc"

type VolumeDevice interface {
	SetVolume(volume float64) error
	VolumeUp() error
	VolumeDown() error
	SetMuted(muted bool) error
	ToggleMuted() error
}

type VolumeState struct {
	Volume *float64 `json:"volume,omitempty"`
	Muted  *bool    `json:"muted,omitempty"`
}

type VolumeChannel struct {
	baseChannel
	device VolumeDevice
}

func NewVolumeChannel(device VolumeDevice) *VolumeChannel {
	return &VolumeChannel{baseChannel{
		protocol: "volume",
	}, device}
}

func (c *VolumeChannel) Set(message *rpc.Message, state *float64) error {
	return c.device.SetVolume(*state)
}

func (c *VolumeChannel) VolumeUp(message *rpc.Message) error {
	return c.device.VolumeUp()
}

func (c *VolumeChannel) VolumeDown(message *rpc.Message) error {
	return c.device.VolumeDown()
}

func (c *VolumeChannel) Mute(message *rpc.Message) error {
	return c.device.SetMuted(true)
}

func (c *VolumeChannel) Unmute(message *rpc.Message) error {
	return c.device.SetMuted(false)
}

func (c *VolumeChannel) ToggleMuted(message *rpc.Message) error {
	return c.device.ToggleMuted()
}

func (c *VolumeChannel) SendState(volume *float64, muted *bool) error {
	return c.SendEvent("state", &VolumeState{
		Volume: volume,
		Muted:  muted,
	})
}

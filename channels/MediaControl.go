package channels

import "github.com/ninjasphere/go-ninja/rpc"

type MediaControlEvent int

const (
	MediaControlEventPlaying MediaControlEvent = iota
	MediaControlEventPaused
	MediaControlEventStopped
	MediaControlEventBuffering
	MediaControlEventBusy
	MediaControlEventIdle
	MediaControlEventInactive
)

var mediaControlEventNames = []string{
	"playing",
	"paused",
	"stopped",
	"buffering",
	"busy",
	"idle",
	"inactive",
}

func (e *MediaControlEvent) Name() string {
	return mediaControlEventNames[int(*e)]
}

type MediaControlDevice interface {
	Play() error
	Pause() error
	TogglePlay() error
	Stop() error
	Next() error
	Previous() error
}

// A MediaControlChannel can be added to devices, exposing http://schemas.ninjablocks.com/protocol/media-control
type MediaControlChannel struct {
	baseChannel
	device MediaControlDevice
}

func NewMediaControlChannel(device MediaControlDevice) *MediaControlChannel {
	return &MediaControlChannel{baseChannel{
		protocol: "media-control",
	}, device}
}

func (c *MediaControlChannel) Play(message *rpc.Message) error {
	return c.device.Play()
}

func (c *MediaControlChannel) Pause(message *rpc.Message) error {
	return c.device.Pause()
}

func (c *MediaControlChannel) TogglePlay(message *rpc.Message) error {
	return c.device.TogglePlay()
}

func (c *MediaControlChannel) Stop(message *rpc.Message) error {
	return c.device.Stop()
}

func (c *MediaControlChannel) Next(message *rpc.Message) error {
	return c.device.Next()
}

func (c *MediaControlChannel) Previous(message *rpc.Message) error {
	return c.device.Previous()
}

func (c *MediaControlChannel) SendState(event MediaControlEvent) error {
	return c.SendEvent(event.Name(), nil)
}

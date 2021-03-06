package channels

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

// A MediaControlChannel can be added to devices, exposing http://schema.ninjablocks.com/protocol/media-control
type MediaControlChannel struct {
	baseChannel
	device MediaControlDevice
}

func NewMediaControlChannel(device MediaControlDevice) *MediaControlChannel {
	return &MediaControlChannel{baseChannel{
		protocol: "media-control",
	}, device}
}

func (c *MediaControlChannel) Play() error {
	return c.device.Play()
}

func (c *MediaControlChannel) Pause() error {
	return c.device.Pause()
}

func (c *MediaControlChannel) TogglePlay() error {
	return c.device.TogglePlay()
}

func (c *MediaControlChannel) Stop() error {
	return c.device.Stop()
}

func (c *MediaControlChannel) Next() error {
	return c.device.Next()
}

func (c *MediaControlChannel) Previous() error {
	return c.device.Previous()
}

func (c *MediaControlChannel) SendState(event MediaControlEvent) error {
	return c.SendEvent(event.Name())
}

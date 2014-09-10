package devices

import (
	"fmt"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/logger"
)

type MediaPlayerDevice struct {
	ApplyTogglePlay   func() error
	ApplyPlayPause    func(playing bool) error
	ApplyStop         func() error
	ApplyPlaylistJump func(delta int) error

	log            *logger.Logger
	bus            *ninja.DeviceBus
	controlChannel *channels.MediaControlChannel
	controlState   channels.MediaControlEvent
}

func (d *MediaPlayerDevice) SetControlState(state channels.MediaControlEvent) error {
	if d.controlChannel == nil {
		return fmt.Errorf("'media-control' channel has not been enabled. Call EnableControlChannel() first")
	}
	if state != d.controlState {
		d.controlState = state
		d.controlChannel.SendState(d.controlState)
	}
	return nil
}

func (d *MediaPlayerDevice) TogglePlay() error {

	if d.ApplyTogglePlay != nil {
		return d.ApplyTogglePlay()
	}

	switch d.controlState {
	case channels.MediaControlEventPlaying, channels.MediaControlEventBuffering, channels.MediaControlEventBusy:
		return d.Pause()
	default:
		return d.Play()
	}

}

func (d *MediaPlayerDevice) Play() error {
	return d.ApplyPlayPause(true)
}

func (d *MediaPlayerDevice) Pause() error {
	return d.ApplyPlayPause(false)
}

func (d *MediaPlayerDevice) Stop() error {
	return d.ApplyStop()
}

func (d *MediaPlayerDevice) Next() error {
	if d.ApplyPlaylistJump == nil {
		return fmt.Errorf("'Next' is not yet supported")
	}
	return d.ApplyPlaylistJump(1)
}

func (d *MediaPlayerDevice) Previous() error {
	if d.ApplyPlaylistJump == nil {
		return fmt.Errorf("'Previous' is not yet supported")
	}
	return d.ApplyPlaylistJump(-1)
}

func (d *MediaPlayerDevice) EnableControlChannel(supportedEvents []string) error {

	d.controlChannel = channels.NewMediaControlChannel(d)

	var supportedMethods []string

	if d.ApplyTogglePlay != nil {
		supportedMethods = append(supportedMethods, "togglePlay")
	}

	if d.ApplyPlayPause != nil {
		supportedMethods = append(supportedMethods, "play", "pause")

		if d.ApplyTogglePlay == nil {
			supportedMethods = append(supportedMethods, "togglePlay")
		}
	}

	if d.ApplyPlaylistJump != nil {
		supportedMethods = append(supportedMethods, "next", "previous")
	}

	if d.ApplyStop != nil {
		supportedMethods = append(supportedMethods, "stop")
	}

	err := d.bus.AddChannelWithSupported(d.controlChannel, "control", "media-control", &supportedMethods, &supportedEvents)
	if err != nil {
		return fmt.Errorf("Failed to create media-control channel: %s", err)
	}

	return nil
}

func CreateMediaPlayerDevice(name string, bus *ninja.DeviceBus) (*MediaPlayerDevice, error) {

	player := &MediaPlayerDevice{
		bus: bus,
		log: logger.GetLogger("MediaPlayerDevice - " + name),
	}

	player.log.Infof("Created")

	return player, nil
}

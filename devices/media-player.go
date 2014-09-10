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

	ApplyVolume      func(volume float64) error
	ApplyMuted       func(muted bool) error
	ApplyToggleMuted func() error
	ApplyVolumeDown  func() error
	ApplyVolumeUp    func() error

	log *logger.Logger
	bus *ninja.DeviceBus

	controlChannel *channels.MediaControlChannel
	controlState   channels.MediaControlEvent

	volumeChannel *channels.VolumeChannel
	volumeState   float64
	mutedState    bool
}

func (d *MediaPlayerDevice) UpdateControlState(state channels.MediaControlEvent) error {
	if d.controlChannel == nil {
		return fmt.Errorf("'media-control' channel has not been enabled. Call EnableControlChannel() first")
	}
	if state != d.controlState {
		d.controlState = state
		d.controlChannel.SendState(d.controlState)
	}
	return nil
}

func (d *MediaPlayerDevice) UpdateVolumeState(state float64) error {
	if d.volumeChannel == nil {
		return fmt.Errorf("'volume' channel has not been enabled. Call EnableVolumeChannel() first")
	}

	d.volumeState = state
	if d.ApplyMuted != nil {
		d.volumeChannel.SendState(&state, &d.mutedState)
	} else {
		d.volumeChannel.SendState(&state, nil)
	}

	return nil
}

func (d *MediaPlayerDevice) UpdateMutedState(muted bool) error {
	if d.volumeChannel == nil {
		return fmt.Errorf("'volume' channel has not been enabled. Call EnableVolumeChannel() first")
	}

	d.mutedState = muted
	if d.ApplyVolume != nil {
		d.volumeChannel.SendState(&d.volumeState, &muted)
	} else {
		d.volumeChannel.SendState(nil, &muted)
	}

	return nil
}

func (d *MediaPlayerDevice) SetMuted(muted bool) error {
	return d.ApplyMuted(muted)
}

func (d *MediaPlayerDevice) ToggleMuted() error {

	if d.ApplyToggleMuted != nil {
		return d.ApplyToggleMuted()
	}

	return d.SetMuted(!d.mutedState)
}

func (d *MediaPlayerDevice) SetVolume(volume float64) error {
	return d.ApplyVolume(volume)
}

func (d *MediaPlayerDevice) VolumeUp() error {
	if d.ApplyVolumeUp != nil {
		return d.ApplyVolumeUp()
	}
	vol := d.volumeState + 0.05
	if vol > 1 {
		vol = 1
	}
	return d.ApplyVolume(vol)
}

func (d *MediaPlayerDevice) VolumeDown() error {
	if d.ApplyVolumeUp != nil {
		return d.ApplyVolumeDown()
	}
	vol := d.volumeState - 0.05
	if vol < 0 {
		vol = 0
	}
	return d.ApplyVolume(vol)
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

func (d *MediaPlayerDevice) EnableVolumeChannel() error {

	d.volumeChannel = channels.NewVolumeChannel(d)

	var supportedMethods []string

	if d.ApplyVolume != nil {
		supportedMethods = append(supportedMethods, "set", "volumeUp", "volumeDown")
	}

	if d.ApplyToggleMuted != nil {
		supportedMethods = append(supportedMethods, "toggleMute")
	}

	if d.ApplyMuted != nil {
		supportedMethods = append(supportedMethods, "mute", "unmute")

		if d.ApplyToggleMuted == nil {
			supportedMethods = append(supportedMethods, "toggleMute")
		}
	}

	supportedEvents := []string{"state"}

	err := d.bus.AddChannelWithSupported(d.volumeChannel, "volume", "volume", &supportedMethods, &supportedEvents)
	if err != nil {
		return fmt.Errorf("Failed to create volume channel: %s", err)
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

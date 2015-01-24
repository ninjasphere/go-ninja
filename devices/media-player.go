package devices

import (
	"errors"
	"fmt"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
)

type MediaPlayerDevice struct {
	baseDevice

	ApplyTogglePlay   func() error
	ApplyPlayPause    func(playing bool) error
	ApplyStop         func() error
	ApplyPlaylistJump func(delta int) error

	ApplyVolume      func(state *channels.VolumeState) error
	ApplyVolumeUp    func() error
	ApplyVolumeDown  func() error
	ApplyToggleMuted func() error

	ApplyPlayURL func(url string, queue bool) error

	controlChannel *channels.MediaControlChannel
	controlState   channels.MediaControlEvent

	volumeChannel *channels.VolumeChannel
	volumeState   float64
	mutedState    bool

	mediaChannel *channels.MediaChannel
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

func (d *MediaPlayerDevice) UpdateVolumeState(state *channels.VolumeState) error {
	if d.volumeChannel == nil {
		return fmt.Errorf("'volume' channel has not been enabled. Call EnableVolumeChannel() first")
	}

	d.volumeChannel.SendState(state)

	return nil
}

func (d *MediaPlayerDevice) SetMuted(muted bool) error {
	if d.ApplyVolume == nil {
		return errors.New("method is not supported")
	}
	return d.ApplyVolume(&channels.VolumeState{&d.volumeState, &muted})
}

func (d *MediaPlayerDevice) ToggleMuted() error {

	if d.ApplyToggleMuted != nil {
		return d.ApplyToggleMuted()
	}

	return d.SetMuted(!d.mutedState)
}

func (d *MediaPlayerDevice) SetVolume(volume *channels.VolumeState) error {
	if d.ApplyVolume == nil {
		return errors.New("method is not supported")
	}
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
	return d.ApplyVolume(&channels.VolumeState{Level: &vol})
}

func (d *MediaPlayerDevice) VolumeDown() error {
	if d.ApplyVolumeUp != nil {
		return d.ApplyVolumeDown()
	}
	vol := d.volumeState - 0.05
	if vol < 0 {
		vol = 0
	}
	return d.ApplyVolume(&channels.VolumeState{Level: &vol})
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

func (d *MediaPlayerDevice) PlayURL(url string, queue bool) error {
	return d.ApplyPlayURL(url, queue)
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

	err := d.conn.ExportChannelWithSupported(d, d.controlChannel, "control", &supportedMethods, &supportedEvents)
	if err != nil {
		return fmt.Errorf("Failed to create media-control channel: %s", err)
	}

	return nil
}

func (d *MediaPlayerDevice) EnableVolumeChannel(supportsMute bool) error {

	d.volumeChannel = channels.NewVolumeChannel(d)

	var supportedMethods []string

	if d.ApplyVolume != nil {
		supportedMethods = append(supportedMethods, "set", "volumeUp", "volumeDown")
		if supportsMute {
			supportedMethods = append(supportedMethods, "mute", "unmute")
		}
	} else {
		if d.ApplyVolumeUp != nil {
			supportedMethods = append(supportedMethods, "volumeUp")
		}
		if d.ApplyVolumeDown != nil {
			supportedMethods = append(supportedMethods, "volumeDown")
		}
	}

	if d.ApplyToggleMuted != nil {
		supportedMethods = append(supportedMethods, "toggleMute")
	}

	supportedEvents := []string{"state"}

	err := d.conn.ExportChannelWithSupported(d, d.volumeChannel, "volume", &supportedMethods, &supportedEvents)
	if err != nil {
		return fmt.Errorf("Failed to create volume channel: %s", err)
	}

	return nil
}

func (d *MediaPlayerDevice) UpdateMusicMediaState(item *channels.MusicTrackMediaItem, position *int) error {
	return d.mediaChannel.SendMusicTrackState(item, position)
}

func (d *MediaPlayerDevice) EnableMediaChannel() error {

	d.mediaChannel = channels.NewMediaChannel(d)

	var supportedMethods []string

	if d.ApplyPlayURL != nil {
		supportedMethods = append(supportedMethods, "playUrl")
	}

	supportedEvents := []string{"state"}

	err := d.conn.ExportChannelWithSupported(d, d.mediaChannel, "media", &supportedMethods, &supportedEvents)
	if err != nil {
		return fmt.Errorf("Failed to create media channel: %s", err)
	}

	return nil
}

func CreateMediaPlayerDevice(driver ninja.Driver, info *model.Device, conn *ninja.Connection) (*MediaPlayerDevice, error) {

	d := &MediaPlayerDevice{
		baseDevice: baseDevice{
			conn:   conn,
			driver: driver,
			log:    logger.GetLogger("MediaPlayerDevice - " + *info.Name),
			info:   info,
		},
	}

	err := conn.ExportDevice(d)
	if err != nil {
		d.log.Fatalf("Failed to export device %s: %s", *info.Name, err)
	}

	d.log.Infof("Created")

	return d, nil
}

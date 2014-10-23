package main

import (
	"fmt"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type fakeMediaPlayer struct {
	ninja *devices.MediaPlayerDevice
}

func NewFakeMediaPlayer(driver ninja.Driver, conn *ninja.Connection, id int) (*fakeMediaPlayer, error) {
	name := fmt.Sprintf("Fancy Fake Media Player %d", id)

	ninja, err := devices.CreateMediaPlayerDevice(driver, &model.Device{
		NaturalID:     fmt.Sprintf("player-%d", id),
		NaturalIDType: "fake",
		Name:          &name,
		Signatures: &map[string]string{
			"ninja:manufacturer": "Fake Co.",
			"ninja:productName":  "fakeMediaPlayer",
			"ninja:productType":  "MediaPlayer",
			"ninja:thingType":    "mediaplayer",
		},
	}, conn)

	var fake *fakeMediaPlayer

	if err == nil {
		fake = &fakeMediaPlayer{
			ninja: ninja,
		}
		err = fake.bindMethods()
		if err != nil {
			fake = nil
		}
	} else {
		fake = nil
	}

	return fake, err
}

func (fake *fakeMediaPlayer) bindMethods() error {

	fake.ninja.ApplyPlayPause = fake.applyPlayPause
	fake.ninja.ApplyStop = fake.applyStop
	fake.ninja.ApplyPlaylistJump = fake.applyPlaylistJump
	fake.ninja.ApplyVolume = fake.applyVolume
	fake.ninja.ApplyPlayURL = fake.applyPlayURL

	err := fake.ninja.EnableControlChannel([]string{
		"playing",
		"paused",
		"stopped",
		"idle",
	})
	if err != nil {
		return err
	}

	err = fake.ninja.EnableVolumeChannel(true)
	if err != nil {
		return err
	}

	err = fake.ninja.EnableMediaChannel()
	if err != nil {
		return err
	}

	return nil
}

func (fake *fakeMediaPlayer) applyPlayPause(playing bool) error {

	fake.ninja.Log().Infof("applyPlayPause called, playing: %t", playing)

	if playing {
		return fake.ninja.UpdateControlState(channels.MediaControlEventPlaying)
	} else {
		return fake.ninja.UpdateControlState(channels.MediaControlEventPaused)
	}
}

func (fake *fakeMediaPlayer) applyStop() error {
	fake.ninja.Log().Infof("applyStop called")

	return fake.ninja.UpdateControlState(channels.MediaControlEventStopped)
}

func (fake *fakeMediaPlayer) applyPlaylistJump(delta int) error {
	fake.ninja.Log().Infof("applyPlaylistJump called, delta : %d", delta)
	return nil
}

func (fake *fakeMediaPlayer) applyVolume(state *channels.VolumeState) error {
	fake.ninja.Log().Infof("applyVolume called, volume %v", state)

	return fake.ninja.UpdateVolumeState(state)
}

func (fake *fakeMediaPlayer) applyPlayURL(url string, queue bool) error {
	fake.ninja.Log().Infof("applyPlayURL called, volume %s, %t", url, queue)
	return nil
}

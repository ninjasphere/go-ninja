package channels

import "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

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
	return &MediaControlChannel{baseChannel{}, device}
}

func (c *MediaControlChannel) Play(message mqtt.Message, _, reply *interface{}) error {
	return c.device.Play()
}

func (c *MediaControlChannel) Pause(message mqtt.Message, _, reply *interface{}) error {
	return c.device.Pause()
}

func (c *MediaControlChannel) TogglePlay(message mqtt.Message, _, reply *interface{}) error {
	return c.device.TogglePlay()
}

func (c *MediaControlChannel) Stop(message mqtt.Message, _, reply *interface{}) error {
	return c.device.Stop()
}

func (c *MediaControlChannel) Next(message mqtt.Message, _, reply *interface{}) error {
	return c.device.Next()
}

func (c *MediaControlChannel) Previous(message mqtt.Message, _, reply *interface{}) error {
	return c.device.Previous()
}

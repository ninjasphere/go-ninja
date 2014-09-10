package channels

import (
	"mime"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type MediaDevice interface {
	PlayURL(url string, autoplay bool) error
}

// A MediaChannel can be added to devices, exposing http://schemas.ninjablocks.com/protocol/media
type MediaChannel struct {
	baseChannel
	device MediaDevice
}

type mediaItem struct {
	ID          *string
	ExternalIDs *map[string]string
	ContentType *MediaContentType
	Type        *string
	Title       *string
	Image       *MediaItemImage
	Duration    *int
}

type MediaContentType string

func (m *MediaContentType) IsValid() bool {
	if _, _, err := mime.ParseMediaType(string(*m)); err != nil {
		return false
	}
	return true
}

type MediaItemImage struct {
	URL    string
	Width  *int
	Height *int
}

type GenericMediaItem struct {
	mediaItem
	Subtitle *string
}

type GenericMediaState struct {
	media *GenericMediaItem
}

type MusicTrackMediaItem struct {
	mediaItem
	Artists *[]MediaItemArtist
}

type MusicTrackMediaState struct {
	media    *MusicTrackMediaItem
	position *int
}

type MediaItemArtist struct {
	ID    *string
	IDs   *map[string]string
	Name  string
	Image *MediaItemImage
}

type MediaItemAlbum struct {
	ID     *string
	IDs    *map[string]string
	Name   string
	Image  *MediaItemImage
	Genres *[]string
}

func NewMediaChannel(device MediaDevice) *MediaChannel {
	return &MediaChannel{baseChannel{}, device}
}

func (c *MediaChannel) PlayUrl(message mqtt.Message, url string, reply *interface{}) error {
	return c.device.PlayURL(url, false)
}

func (c *MediaChannel) QueueUrl(message mqtt.Message, url string, reply *interface{}) error {
	return c.device.PlayURL(url, true)
}

func (c *MediaChannel) SendGenericState(state *GenericMediaItem) error {
	return c.SendEvent("state", &GenericMediaState{state})
}

func (c *MediaChannel) SendMusicTrackState(state *MusicTrackMediaItem, position *int) error {
	return c.SendEvent("state", &MusicTrackMediaState{state, position})
}

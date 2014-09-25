package channels

import (
	"mime"

	"github.com/ninjasphere/go-ninja/rpc"
)

type MediaDevice interface {
	PlayURL(url string, autoplay bool) error
}

// A MediaChannel can be added to devices, exposing http://schemas.ninjablocks.com/protocol/media
type MediaChannel struct {
	baseChannel
	device MediaDevice
}

type MediaItem struct {
	ID          *string            `json:"id,omitempty"`
	ExternalIDs *map[string]string `json:"externalIds,omitempty"`
	ContentType *MediaContentType  `json:"contentType,omitempty"`
	Type        *string            `json:"type,omitempty"`
	Title       *string            `json:"title,omitempty"`
	Image       *MediaItemImage    `json:"image,omitempty"`
	Duration    *int               `json:"duration,omitempty"`
}

type MediaContentType string

func (m *MediaContentType) IsValid() bool {
	if _, _, err := mime.ParseMediaType(string(*m)); err != nil {
		return false
	}
	return true
}

type MediaItemImage struct {
	URL    string `json:"url"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
}

type GenericMediaItem struct {
	MediaItem
	Subtitle *string `json:"subtitle,omitempty"`
}

type GenericMediaState struct {
	Media *GenericMediaItem `json:"media,omitempty"`
}

type MusicTrackMediaItem struct {
	ID          *string            `json:"id,omitempty"`
	ExternalIDs *map[string]string `json:"externalIds,omitempty"`
	ContentType *MediaContentType  `json:"contentType,omitempty"`
	Type        *string            `json:"type,omitempty"`
	Title       *string            `json:"title,omitempty"`
	Image       *MediaItemImage    `json:"image,omitempty"`
	Duration    *int               `json:"duration,omitempty"`
	Artists     *[]MediaItemArtist `json:"artists,omitempty"`
	Album       *MediaItemAlbum    `json:"album,omitempty"`
}

type MusicTrackMediaState struct {
	Media    *MusicTrackMediaItem `json:"media,omitempty"`
	Position *int                 `json:"position,omitempty"`
}

type MediaItemArtist struct {
	ID          *string            `json:"id,omitempty"`
	externalIds *map[string]string `json:"externalIds,omitempty"`
	Name        string             `json:"name"`
	Image       *MediaItemImage    `json:"image,omitempty"`
}

type MediaItemAlbum struct {
	ID          *string            `json:"id,omitempty"`
	externalIds *map[string]string `json:"externalIds,omitempty"`
	Name        string             `json:"name"`
	Image       *MediaItemImage    `json:"image,omitempty"`
	Genres      *[]string          `json:"genres,omitempty"`
}

func NewMediaChannel(device MediaDevice) *MediaChannel {
	return &MediaChannel{baseChannel{
		protocol: "media",
	}, device}
}

func (c *MediaChannel) PlayUrl(message *rpc.Message, url *string) error {
	return c.device.PlayURL(*url, false)
}

func (c *MediaChannel) QueueUrl(message *rpc.Message, url *string) error {
	return c.device.PlayURL(*url, true)
}

func (c *MediaChannel) SendGenericState(state *GenericMediaItem) error {
	return c.SendEvent("state", &GenericMediaState{state})
}

func (c *MediaChannel) SendMusicTrackState(state *MusicTrackMediaItem, position *int) error {
	return c.SendEvent("state", &MusicTrackMediaState{state, position})
}

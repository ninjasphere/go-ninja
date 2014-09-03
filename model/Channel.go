package model

type Channel struct {
	ID        string            `json:"id"`
	Name      string            `json:"channel"`
	Protocol  string            `json:"protocol"`
	Supported *ChannelSupported `json:"supported"`
	Device    *Device           `json:"device"`
}

type ChannelSupported struct {
	Methods *[]string `json:"methods"`
	Events  *[]string `json:"events"`
}

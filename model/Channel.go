package model

type Channel struct {
	ServiceAnnouncement
	ID        string      `json:"id" redis:"id"`
	Protocol  string      `json:"protocol" redis:"protocol"`
	Device    *Device     `json:"device" redis:"-"`
	LastState interface{} `json:"lastState" redis:"-"` // used to store the last state of the channel
}

func (c *Channel) GetServiceAnnouncement() *ServiceAnnouncement {
	return &c.ServiceAnnouncement
}

package model

type Channel struct {
	ServiceAnnouncement
	ID        string      `json:"id" redis:"id"`
	Protocol  string      `json:"protocol" redis:"protocol"`
	DeviceID  string      `json:"deviceId" redis:"deviceId"`
	LastState interface{} `json:"lastState" redis:"-"` // used to store the last state of the channel
}

func (c *Channel) GetServiceAnnouncement() *ServiceAnnouncement {
	return &c.ServiceAnnouncement
}

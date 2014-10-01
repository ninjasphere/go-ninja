package model

type Channel struct {
	ServiceAnnouncement
	ID       string  `json:"id" redis:"id"`
	Protocol string  `json:"protocol" redis:"protocol"`
	Device   *Device `json:"device" redis:"-"`
}

func (c *Channel) GetServiceAnnouncement() *ServiceAnnouncement {
	return &c.ServiceAnnouncement
}

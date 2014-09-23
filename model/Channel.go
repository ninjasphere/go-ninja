package model

type Channel struct {
	ServiceAnnouncement
	ID       string  `json:"id"`
	Protocol string  `json:"protocol"`
	Device   *Device `json:"device"`
}

func (c *Channel) GetServiceAnnouncement() *ServiceAnnouncement {
	return &c.ServiceAnnouncement
}

package model

type Module struct {
	ServiceAnnouncement
	ID          string `json:"id" redis:"id"`
	Name        string `json:"name" redis:"name"`
	Version     string `json:"version" redis:"version"`
	Description string `json:"description" redis:"description"`
	Author      string `json:"author" redis:"author"`
	License     string `json:"license" redis:"license"`
}

func (m *Module) GetServiceAnnouncement() *ServiceAnnouncement {
	return &m.ServiceAnnouncement
}

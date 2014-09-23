package model

type Module struct {
	ServiceAnnouncement
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
	Path        string `json:"path"`
}

func (m *Module) GetServiceAnnouncement() *ServiceAnnouncement {
	return &m.ServiceAnnouncement
}

package ninja

import "github.com/ninjasphere/go-ninja/model"

type discoverService struct {
	conn *Connection
}

func (s *discoverService) GetServiceAnnouncement() *model.ServiceAnnouncement {
	return &model.ServiceAnnouncement{
		Schema: resolveSchemaURI("/service/discover"),
	}
}

func (s *discoverService) Services(schema string) (*[]model.ServiceAnnouncement, error) {
	if schema == "" {
		return &s.conn.services, nil
	}

	schema = resolveSchemaURI(schema)
	matching := []model.ServiceAnnouncement{}

	for _, service := range s.conn.services {
		if service.Schema == schema {
			matching = append(matching, service)
		}
	}

	return &matching, nil
}

package model

// Device TODO: The channels and thing might want to be on a struct thats only used in the devicemodel service?
type Device struct {
	ServiceAnnouncement
	ID            string             `json:"id" redis:"id"`
	NaturalID     string             `json:"naturalId" redis:"naturalId"`
	NaturalIDType string             `json:"naturalIdType" redis:"naturalIdType"`
	Name          *string            `json:"name,omitempty" redis:"name"`
	ThingID       *string            `json:"thing,omitempty" redis:"-"`
	Channels      *[]*Channel        `json:"channels,omitempty" redis:"-"`
	Signatures    *map[string]string `json:"signatures,omitempty" redis:"signatures,json"`
}

func (d *Device) GetServiceAnnouncement() *ServiceAnnouncement {
	return &d.ServiceAnnouncement
}

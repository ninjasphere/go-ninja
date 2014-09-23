package model

// Device TODO: The channels and thing might want to be on a struct thats only used in the devicemodel service?
type Device struct {
	ServiceAnnouncement
	ID         string             `json:"id" redis:"id"`
	IDType     string             `json:"idType" redis:"idType"`
	GUID       string             `json:"guid" redis:"guid"`
	Name       *string            `json:"name,omitempty" redis:"name"`
	Thing      *string            `json:"thing,omitempty" redis:"thing"`
	Channels   *[]Channel         `json:"channels,omitempty" redis:"-"`
	Signatures *map[string]string `json:"signatures,omitempty" redis:"signatures,json"`
}

func (d *Device) GetServiceAnnouncement() *ServiceAnnouncement {
	return &d.ServiceAnnouncement
}

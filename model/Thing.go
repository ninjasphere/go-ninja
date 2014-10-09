package model

type Thing struct {
	ID       string  `json:"id" redis:"id"`
	Type     string  `json:"type" redis:"type"`
	Name     string  `json:"name" redis:"name"`
	Device   *Device `json:"device,omitempty" redis:"-"`
	DeviceID *string `json:"deviceId,omitempty" redis:"-"`
	Location *string `json:"location,omitempty" redis:"location"`
}

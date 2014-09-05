package model

type Thing struct {
	Name     string  `json:"name" redis:"name"`
	ID       string  `json:"id" redis:"id"`
	Device   *Device `json:"device,omitEmpty" redis:"-"`
	DeviceID *string `json:"-" redis:"device"`
	Type     string  `json:"type" redis:"type"`
	Location *string `json:"location,omitEmpty" redis:"location"`
}

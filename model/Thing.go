package model

type Thing struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Device Device `json:"device,omitEmpty"`
}

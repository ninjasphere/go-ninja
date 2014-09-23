package model

type ServiceAnnouncement struct {
	Schema           string    `json:"schema"`
	SupportedMethods *[]string `json:"methods"`
	SupportedEvents  *[]string `json:"events"`
}

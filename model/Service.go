package model

type ServiceAnnouncement struct {
	Topic            string    `json:"topic" redis:"topic"`
	Schema           string    `json:"schema" redis:"schema"`
	SupportedMethods *[]string `json:"supportedMethods" redis:"supportedMethods,json"`
	SupportedEvents  *[]string `json:"supportedEvents" redis:"supportedEvents,json"`
}

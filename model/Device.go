package model

type Device struct {
	ID         string             `json:"id" redis:"id"`
	IDType     string             `json:"idType" redis:"idType"`
	Guid       string             `json:"guid" redis:"guid"`
	Name       *string            `json:"name,omitEmpty" redis:"name"`
	Thing      *string            `json:"thing,omitEmpty" redis:"thing"`
	Channels   *[]Channel         `json:"channels,omitEmpty" redis:"-"`
	Signatures *map[string]string `json:"signatures,omitEmpty" redis:"signatures,json"`
}

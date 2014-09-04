package model

type Device struct {
	ID         string             `json:"id"`
	IDType     string             `json:"idType"`
	Guid       string             `json:"guid"`
	Name       *string            `json:"name,omitEmpty"`
	Thing      *string            `json:"thing,omitEmpty"`
	Channels   *[]Channel         `json:"channels,omitEmpty"`
	Signatures *map[string]string `json:"signatures,omitEmpty"`
}

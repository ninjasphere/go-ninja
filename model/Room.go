package model

type Room struct {
	Name string `json:"name" redis:"name"`
	ID   string `json:"id" redis:"id"`
}

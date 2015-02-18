package model

type ConfigurationRequest struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

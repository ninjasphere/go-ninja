package model

import "encoding/json"

type ConfigurationRequest struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

package events

type PairingRequest struct {
	Duration int `json:"duration"`
}

type PairingStarted struct {
	Duration int `json:"duration"`
}

type PairingEnded struct {
	DevicesFound int `json:"devicesFound"`
}

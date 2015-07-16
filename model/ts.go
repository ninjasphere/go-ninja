package model

type TimeSeriesPayload struct {
	Thing          string                `json:"thing"`
	ThingType      string                `json:"thingType"`
	Promoted       bool                  `json:"promoted"`
	Device         string                `json:"device"`
	Channel        string                `json:"channel"`
	Schema         string                `json:"schema"`
	Event          string                `json:"event"`
	Points         []TimeSeriesDatapoint `json:"points"`
	Time           string                `json:"time"`
	TimeZone       string                `json:"timeZone"`
	TimeOffset     int                   `json:"timeOffset"`
	Site           string                `json:"site"`
	ReportingZones map[string]string     `json:"zones"`
	UserOverride   string                `json:"_"`
	NodeOverride   string                `json:"_"`
}

type TimeSeriesDatapoint struct {
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

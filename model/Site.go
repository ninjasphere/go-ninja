package model

type Site struct {
	ID              string      `json:"id,omitempty" redis:"id"`
	Name            *string     `json:"name,omitempty" redis:"name"`
	Type            *string     `json:"type,omitempty" redis:"type"`
	Latitude        *float64    `json:"latitude,omitempty" redis:"latitude"`
	Longitude       *float64    `json:"longitude,omitempty" redis:"longitude"`
	TimeZoneID      *string     `json:"timeZoneId,omitempty" redis:"timeZoneId"`
	TimeZoneName    *string     `json:"timeZoneName,omitempty" redis:"timeZoneName"`
	TimeZoneOffset  *int        `json:"timeZoneOffset,omitempty" redis:"timeZoneOffset"`
	SitePreferences interface{} `json:"site-preferences,omitempty" redis:"site-preferences,json"`
}

//https://maps.googleapis.com/maps/api/timezone/json?location=-33.86,151.20&timestamp=1414645501

/*{
  id: "whatever",
  name: "Home",
  type: "home",
  latitude: -33.86,
  longitude: 151.20,
  timeZoneID: "Australia/Sydney",
  timeZoneName: "Australian Eastern Daylight Time",
  timeZoneOffset: 36000
}*/

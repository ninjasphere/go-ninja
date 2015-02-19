package suit

import (
	"encoding/json"
	"log"
	"testing"
)

func TestLowerCaseJSON(t *testing.T) {

	js, err := json.Marshal(&ConfigurationScreen{
		Title: "Create new security light",
		Sections: []Section{
			Section{
				Contents: []Typed{
					InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "My Security Light",
						Value:       "Front door light",
					},
					OptionGroup{
						Name:           "sensors",
						Title:          "When these devices detect motion",
						MinimumChoices: 1,
						Options: []OptionGroupOption{
							OptionGroupOption{
								Title:    "Front Door Motion",
								Subtitle: "Motion",
								Value:    "fd",
							},
							OptionGroupOption{
								Title:    "Back Door 1",
								Subtitle: "Presence",
								Value:    "bd",
								Selected: true,
							},
						},
					},
					OptionGroup{
						Name:           "lights",
						Title:          "Turn on these lights",
						MinimumChoices: 1,
						Options: []OptionGroupOption{
							OptionGroupOption{
								Title:    "Front Door",
								Subtitle: "Lamp in Hallway",
								Value:    "fd",
							},
							OptionGroupOption{
								Title:    "Front Door Spotlight",
								Subtitle: "Light in Front Step",
								Value:    "fds",
								Selected: true,
							},
						},
					},
					InputTimeRange{
						Name:  "time",
						Title: "When",
						Value: TimeRange{
							From: "10:00",
							To:   "sunset",
						},
					},
					InputText{
						Title:     "Turn off again after",
						After:     "minutes",
						Name:      "timeout",
						InputType: "number",
						Minimum:   i(0),
						Value:     5,
					},
				},
			},
		},
		Actions: []Typed{
			CloseAction{
				Label: "Cancel",
			},
			ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	})

	log.Printf("JSON: %s %s", js, err)
}

func i(i int) *int {
	return &i
}

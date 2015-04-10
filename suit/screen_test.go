package suit

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

var fixture = ConfigurationScreen{
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
					Minimum:   i(1),
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
}

var jsonFixture = `{
  "actions": [
    {
      "label": "Cancel",
      "type": "close"
    },
    {
      "displayClass": "success",
      "displayIcon": "star",
      "label": "Save",
      "name": "save",
      "type": "reply"
    }
  ],
  "sections": [
    {
      "contents": [
        {
          "before": "Name",
          "name": "name",
          "placeholder": "My Security Light",
          "type": "inputText",
          "value": "Front door light"
        },
        {
          "minimumChoices": 1,
          "name": "sensors",
          "options": [
            {
              "subtitle": "Motion",
              "title": "Front Door Motion",
              "value": "fd"
            },
            {
              "selected": true,
              "subtitle": "Presence",
              "title": "Back Door 1",
              "value": "bd"
            }
          ],
          "title": "When these devices detect motion",
          "type": "optionGroup"
        },
        {
          "minimumChoices": 1,
          "name": "lights",
          "options": [
            {
              "subtitle": "Lamp in Hallway",
              "title": "Front Door",
              "value": "fd"
            },
            {
              "selected": true,
              "subtitle": "Light in Front Step",
              "title": "Front Door Spotlight",
              "value": "fds"
            }
          ],
          "title": "Turn on these lights",
          "type": "optionGroup"
        },
        {
          "name": "time",
          "title": "When",
          "type": "inputTimeRange",
          "value": {
            "from": "10:00",
            "to": "sunset"
          }
        },
        {
          "after": "minutes",
          "inputType": "number",
          "name": "timeout",
          "title": "Turn off again after",
          "type": "inputText",
          "value": 5,
          "minimum": 1
        }
      ]
    }
  ],
  "title": "Create new security light"
}`

func TestLowerCaseJSON(t *testing.T) {

	js, err := json.Marshal(&fixture)
	log.Printf("JSON: %s %v", js, err)
}

func i(i int) *int {
	return &i
}

func normalizeJSON(j string) string {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(j), &m); err != nil {
		return j
	} else {
		b, _ := json.MarshalIndent(m, "", "  ")
		return string(b)
	}
}

func roundtripFromJSON(t *testing.T, j string) {
	n := normalizeJSON(j)
	cs := ConfigurationScreen{}
	if err := cs.UnmarshalJSON([]byte(n)); err != nil {
		t.Fatalf("Failed while unmarshalling json: %v: %s", err, j)
	} else {
		if b, err := cs.MarshalJSON(); err != nil {
			t.Fatalf("Failed while marshalling json: %v", err)
		} else {
			nn := normalizeJSON(string(b))
			if n != nn {
				t.Fatalf("Round trip failed: \n%s\n != \n%s\n", n, nn)
			} else {
				log.Printf("n==nn \n%s\n==\n%s\n", n, nn)
			}
		}
	}
}

func roundtrip(t *testing.T, cs *ConfigurationScreen) {
	log.Printf("original form is %+v", cs)
	if buf, err := cs.MarshalJSON(); err != nil {
		t.Fatalf("marshalling failed: %v", err)
	} else {
		log.Printf("marshalled form is %s", normalizeJSON(string(buf)))
		read := ConfigurationScreen{}
		if err := read.UnmarshalJSON(buf); err != nil {
			t.Fatalf("unmarshalling failed: %v", err)
		}
		log.Printf("unmarshalled form is %+v", read)
		if buf2, err := read.MarshalJSON(); err != nil {
			t.Fatalf("remarshalling failed: %v", err)
		} else {
			if bytes.Compare(buf, buf2) != 0 {
				t.Fatalf("inconsistent remarshalling:\n%s\n---\n%s\n", normalizeJSON(string(buf)), normalizeJSON(string(buf2)))
			}
		}
	}
}

func TestRoundTripEmpty(t *testing.T) {
	roundtrip(t, &ConfigurationScreen{})
}

func TestRoundTripFixture(t *testing.T) {
	roundtrip(t, &fixture)
}

func TestRoundTripFromJSON(t *testing.T) {
	roundtripFromJSON(t, jsonFixture)
}

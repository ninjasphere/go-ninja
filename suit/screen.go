package suit

import (
	"encoding/json"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type ConfigurationScreen struct {
	Title       string
	Subtitle    string
	DisplayIcon string
	Sections    []Section
	Actions     []Typed
}

func (o *ConfigurationScreen) MarshalJSON() ([]byte, error) {
	return json.Marshal(walk(*o))
}

type Section struct {
	Title    string
	Subtitle string
	Contents []Typed
	Well     bool
}

type InputText struct {
	Title       string
	Subtitle    string
	Before      string
	After       string
	Placeholder string
	Name        string
	Value       interface{}
	InputType   string
	Minimum     *int
	Maximum     *int
}

func (o InputText) getType() string {
	return "inputText"
}

type StaticText struct {
	Title    string
	Subtitle string
	Before   string
	After    string
	Value    string
}

func (o StaticText) getType() string {
	return "staticText"
}

type InputCheckbox struct {
	Title    string
	Subtitle string
	Before   string
	After    string
	Name     string
	Value    string
	Selected bool
}

func (InputCheckbox) getType() string {
	return "inputCheckbox"
}

type InputTime struct {
	Title    string
	Subtitle string
	Before   string
	After    string
	Name     string
	Value    string
}

func (o InputTime) getType() string {
	return "inputTime"
}

type Separator struct {
}

func (o Separator) getType() string {
	return "separator"
}

type OptionGroup struct {
	Title          string
	Subtitle       string
	Name           string
	MinimumChoices int
	MaximumChoices int
	Options        []OptionGroupOption
}

func (o OptionGroup) getType() string {
	return "optionGroup"
}

type OptionGroupOption struct {
	Title    string
	Subtitle string
	Value    string
	Selected bool
}

type RadioGroup struct {
	Title    string
	Subtitle string
	Name     string
	Value    string
	Options  []RadioGroupOption
}

func (o RadioGroup) getType() string {
	return "radioGroup"
}

type RadioGroupOption struct {
	Title       string
	Value       string
	DisplayIcon string
}

type Alert struct {
	Title        string
	Subtitle     string
	DisplayClass string
	DisplayIcon  string
}

func (o Alert) getType() string {
	return "alert"
}

type ActionList struct {
	Title           string
	Subtitle        string
	Name            string
	Options         []ActionListOption
	PrimaryAction   Typed
	SecondaryAction Typed
}

func (o ActionList) getType() string {
	return "actionList"
}

type ActionListOption struct {
	Title    string
	Subtitle string
	Value    string
}

type InputTimeRange struct {
	Title    string
	Subtitle string
	Name     string
	Value    TimeRange
}

func (o InputTimeRange) getType() string {
	return "inputTimeRange"
}

type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type InputHidden struct {
	Name  string
	Value string
}

func (o InputHidden) getType() string {
	return "inputHidden"
}

type CloseAction struct {
	Label string
}

func (o CloseAction) getType() string {
	return "close"
}

type ReplyAction struct {
	Label        string
	Name         string
	DisplayClass string
	DisplayIcon  string
}

func (o ReplyAction) getType() string {
	return "reply"
}

func (o *ReplyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(walk(*o))
}

type AutomaticAction struct {
	Name  string
	Delay int
}

func (o AutomaticAction) getType() string {
	return "auto"
}

type ProgressBar struct {
	Title        string
	Subtitle     string
	Label        string
	Progress     int /* percentage */
	DisplayClass string
	DisplayIcon  string
}

func (o ProgressBar) getType() string {
	return "progressBar"
}

type Typed interface {
	getType() string
}

func walk(o interface{}) map[string]interface{} {

	m := make(map[string]interface{})

	if t, ok := o.(Typed); ok {
		m["type"] = t.getType()
	}

	val := reflect.ValueOf(o)

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		val := valueField.Interface()

		if val == nil {
			continue
		}

		valueField = reflect.ValueOf(val)

		if valueField.Kind() == reflect.Ptr && !isZero(valueField) {
			valueField = valueField.Elem()
			val = valueField.Interface()
		}

		switch valueField.Kind() {
		case reflect.Struct:
			val = walk(val)
		case reflect.Slice:
			vals := []interface{}{}
			for i := 0; i < valueField.Len(); i++ {
				if valueField.Index(i).Kind() == reflect.Interface || valueField.Index(i).Kind() == reflect.Struct {
					vals = append(vals, walk(valueField.Index(i).Interface()))
				} else {
					vals = append(vals, valueField.Index(i).Interface())
				}
				val = vals
			}
		default:
			if isZero(valueField) {
				val = nil
			}
		}

		if val != nil {
			m[lF(typeField.Name)] = val
		}
	}

	return m
}

func isZero(valueField reflect.Value) bool {
	return valueField.Interface() == reflect.Zero(valueField.Type()).Interface()
}

func lF(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

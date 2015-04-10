package suit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

func (o *ConfigurationScreen) MarshalJSON() ([]byte, error) {
	return json.Marshal(walk(*o))
}

func (o *ConfigurationScreen) UnmarshalJSON(bytes []byte) error {
	aMap := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &aMap); err != nil {
		return err
	}
	return hydrate(aMap, o)
}

func makeTyped(typeName string) (interface{}, error) {
	switch typeName {
	case "actionList":
		return &ActionList{}, nil
	case "alert":
		return &Alert{}, nil
	case "auto":
		return &AutomaticAction{}, nil
	case "close":
		return &CloseAction{}, nil
	case "inputHidden":
		return &InputHidden{}, nil
	case "inputText":
		return &InputText{}, nil
	case "inputTime":
		return &InputTime{}, nil
	case "inputTimeRange":
		return &InputTimeRange{}, nil
	case "optionGroup":
		return &OptionGroup{}, nil
	case "progressBar":
		return &ProgressBar{}, nil
	case "radioGroup":
		return &RadioGroup{}, nil
	case "reply":
		return &ReplyAction{}, nil
	case "separator":
		return &Separator{}, nil
	case "staticText":
		return &StaticText{}, nil
	default:
		return nil, fmt.Errorf("can't make object for type: %s", typeName)
	}
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

func hydrate(s map[string]interface{}, o interface{}) error {
	v := reflect.ValueOf(o).Elem()
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			var mfv reflect.Value
			fv := v.Field(i)
			ft := t.Field(i)

			if sv, ok := s[lF(ft.Name)]; ok && sv != nil {
				mfv = reflect.ValueOf(sv)
				switch ft.Type.Kind() {
				case reflect.Struct:
					if svMap, ok := sv.(map[string]interface{}); !ok {
						return fmt.Errorf("failed to convert value to map")
					} else {
						if err := hydrate(svMap, fv.Addr().Interface()); err != nil {
							return fmt.Errorf("failed to hydrate %+v", ft)
						}
						continue
					}
				case reflect.Slice:
					if mfv.Kind() != reflect.Slice {
						return fmt.Errorf("hydrate: while processing '%s': failed to map '%v' to slice: value=%v", ft.Name, mfv.Kind(), fv)
					} else {
						if mfv.Type() != ft.Type {
							tmp := reflect.MakeSlice(ft.Type, mfv.Len(), mfv.Len())
							for j := 0; j < mfv.Len(); j++ {
								p := mfv.Index(j)
								vp := reflect.Indirect(p)
								vpMap := vp.Interface().(map[string]interface{})
								if ft.Type.Elem().Kind() == reflect.Interface {
									if typeName, ok := vpMap["type"].(string); ok {
										if typed, err := makeTyped(typeName); err != nil {
											return err
										} else {
											if err := hydrate(vpMap, typed); err != nil {
												return err
											}
											tmp.Index(j).Set(reflect.ValueOf(typed).Elem())
										}
									} else {
										return fmt.Errorf("trying to unmarshall interface, but no type available")
									}
								} else {
									if err := hydrate(vpMap, tmp.Index(j).Addr().Interface()); err != nil {
										return err
									}
								}
							}
							mfv = tmp
						}
					}
				case reflect.Ptr:
					nfv := reflect.New(ft.Type.Elem())
					reflect.Indirect(nfv).Set(mfv.Convert(ft.Type.Elem()))
					mfv = nfv
				}
			} else {
				mfv = reflect.Zero(ft.Type)
			}

			fv.Set(mfv.Convert(ft.Type))
		}
		return nil
	default:
		return fmt.Errorf("hydrate: unhandled kind: %v", v.Kind())
	}
}

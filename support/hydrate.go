package support

import (
	"encoding/json"
)

// Replace keys of a map of interface{} values with the result of rehydrating the JSON representation
// of input value into the corresponding value of the models map.
//
// The purpose of this function is to allow strongly typed access to part of a map in cases
// where strong types are only known for some of a map's values.
//
// Following the execution of this method result[k].({sometype}) will be non-nil for each k in models
// where {sometype} is the type of the value models[k] and equal to input[k] otherwise.
func Rehydrate(input interface{}, models map[string]interface{}) (map[string]interface{}, error) {
	var err error
	output := make(map[string]interface{})
	if input != nil {
		for k, v := range input.(map[string]interface{}) {
			output[k] = v
		}
	}

	for k, v := range models {
		var dv interface{} = nil
		var ev []byte
		if input != nil {
			dv = input.(map[string]interface{})[k]
			if dv != nil {
				if ev, err = json.Marshal(dv); err != nil {
					return nil, err
				}
			}
		}
		if len(ev) != 0 {
			if err = json.Unmarshal(ev, v); err != nil {
				return nil, err
			}
		}
		output[k] = v
	}
	return output, nil
}

package schemas

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ninjasphere/go-ninja/config"
)

func main() {
	log.Println("Howdy!")

	schema := `http://schema.ninjablocks.com/state/common#/definitions/humidity`
	json := -10

	message, err := Validate(schema, json)
	if err != nil {
		log.Fatalf("Validation errored: %s", err)
	}
	if message != nil {
		log.Fatalf("Validation failed: %s", *message)
	}

	log.Fatalf("Validation Passed")

}

var validator *shim
var validationEnabled = config.MustBool("validate")

func init() {
	if validationEnabled {
		startProcess()
	}
}

func startProcess() {
	var err error

	validator, err = newCShim("sphere-validator")

	if err != nil {
		log.Fatalf("Failed to start sphere-validator: %s", err)
	}
}

func Validate(schema string, obj interface{}) (*string, error) {
	js, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return ValidateString(schema, string(js))
}

func ValidateString(schema string, json string) (*string, error) {
	if !validationEnabled {
		log.Printf("Skipping validation of %s", schema)
		return nil, nil
	}

	validator.Lock()
	defer validator.Unlock()

	log.Printf("schema-validator: Sending %s %s", schema, json)

	_, err := fmt.Fprintf(validator, "%s %s", schema, json)
	if err != nil {
		return nil, err
	}

	validator.Scan()

	err = validator.Err()
	result := validator.Text()

	if result == "null" {
		return nil, err
	}

	return &result, err
}

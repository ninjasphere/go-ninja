package schemas

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("schemas")

/*
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

}*/

type validatorConn struct {
	conn net.Conn
	sync.Mutex
	io.Reader
	io.Writer
	bufio.Scanner
}

// newCShim starts the shim named file using the provided args.
func newValidatorConn(port int) (*validatorConn, error) {
	c := new(validatorConn)
	var err error

	// TODO: Automatically redial
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	c.conn = conn

	c.Writer = conn
	c.Reader = conn

	c.Mutex = sync.Mutex{}
	c.Scanner = *bufio.NewScanner(c)

	return c, err
}

var validator *validatorConn
var validationEnabled = config.MustBool("validate")

func init() {
	connectToValidator()
}

func connectToValidator() {
	if validator != nil {
		validator.conn.Close()
	}
	var err error
	validator, err = newValidatorConn(8666)
	if err != nil {
		log.Fatalf("Failed to connect to validator server: %s", err)
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
		log.Debugf("Skipping validation of %s", schema)
		return nil, nil
	}

	validator.Lock()
	defer validator.Unlock()

	log.Debugf("Xschema-validator: validating %s %s", schema, json)

	_, err := fmt.Fprintf(validator, "validate %s %s", schema, json)
	if err != nil {
		log.Infof("Validator errored. Will try to reconnect in a few seconds. Message:%s", err)
		time.Sleep(time.Second * 3)
		connectToValidator()
		time.Sleep(time.Second * 5)
		validator.Unlock()
		return ValidateString(schema, json)
	}

	validator.Scan()

	err = validator.Err()
	result := validator.Text()

	if result == "null" {
		return nil, err
	}

	return &result, err
}

func GetServiceMethods(schema string) ([]string, error) {

	validator.Lock()
	defer validator.Unlock()

	log.Debugf("schema-validator: Getting service methods for %s", schema)

	_, err := fmt.Fprintf(validator, "methods %s", schema)
	if err != nil {
		log.Infof("Validator errored. Will try to reconnect in a few seconds. Message:%s", err)
		time.Sleep(time.Second * 3)
		connectToValidator()
		time.Sleep(time.Second * 5)
		validator.Unlock()
		return GetServiceMethods(schema)
	}

	validator.Scan()

	err = validator.Err()
	result := validator.Text()

	return strings.Split(result, ","), nil
}

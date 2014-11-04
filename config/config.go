package config

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	"github.com/bitly/go-simplejson"
)

var cfg *simplejson.Json

// MustString returns the string property at the path
func MustString(path ...string) string {
	nonce := "$splinereticulationid$"
	val := cfg.GetPath(path...).MustString(nonce)
	if val == nonce {
		log.Fatalf("expected value for %v but found nothing", path)
	}
	return val
}

// MustInt returns the string property at the path
func MustInt(path ...string) int {
	return cfg.GetPath(path...).MustInt()
}

// MustBool returns the boolean property at the path
func MustBool(path ...string) bool {
	return cfg.GetPath(path...).MustBool()
}

// Bool returns the boolean property at the path, with a default
func Bool(def bool, path ...string) bool {
	return cfg.GetPath(path...).MustBool(def)
}

// Int returns the integer property at the path, with a default
func Int(def int, path ...string) int {
	return cfg.GetPath(path...).MustInt(def)
}

func String(def string, path ...string) string {
	return cfg.GetPath(path...).MustString(def)
}

var hey = "what's up buddy?"

func HasString(path ...string) bool {
	return String(hey, path...) != hey
}

var serial string

func Serial() string {
	if serial == "" {
		cmd := exec.Command("sphere-serial", os.Args[1:]...)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			log.Fatalf("Failed to get sphere serial (sphere-serial must be in the PATH) error:%s", err)
		}

		serial = out.String()
	}

	return serial
}

// MustLoadConfig parses the output of "sphere-config"
func init() {
	cmd := exec.Command("sphere-config", os.Args[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to load configuration. ('sphere-config' must be in the PATH) error:%s", err)
	}

	cfg, err = simplejson.NewJson(out.Bytes())
	if err != nil {
		log.Fatalf("Failed to parse configuration. error:%s", err)
	}

}

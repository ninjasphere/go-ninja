package config

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/bitly/go-simplejson"
)

var cfg *simplejson.Json

// MustString returns the string property at the path
func MustString(path ...string) string {
	return cfg.GetPath(path...).MustString()
}

// MustInt returns the string property at the path
func MustInt(path ...string) int {
	return cfg.GetPath(path...).MustInt()
}

// MustBool returns the string property at the path
func MustBool(path ...string) bool {
	return cfg.GetPath(path...).MustBool()
}

func Serial() string {

	cmd := exec.Command("sphere-serial")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to get sphere serial (sphere-serial must be in the PATH) error:%s", err)
	}

	return out.String()
}

// MustLoadConfig parses the output of "sphere-config"
func init() {
	cmd := exec.Command("sphere-config")

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

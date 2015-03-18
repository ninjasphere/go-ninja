package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/juju/loggo"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	config map[string]interface{}
	log    = loggo.GetLogger("config") // avoid the wrapper as it uses this module and will cause a loop
)

func init() {
	MustRefresh()

	if Bool(false, "dumpConfig") {
		spew.Dump(GetAll(false))
	}
}

func GetAll(flatten bool) map[string]interface{} {
	if flatten {
		return config
	}
	return unflatten(config)
}

var serial string

func Serial() string {
	if serial == "" {
		cmd := exec.Command("sphere-serial", os.Args[1:]...)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			log.Errorf("Failed to get sphere serial (sphere-serial must be in the PATH) error:%s", err)
			panic(err)
		}

		serial = out.String()
	}

	return serial
}

func IsPaired() bool {
	return /*HasString("sphereNetworkKey") && */ HasString("token") && HasString("userId")
}

func String(def string, path ...string) string {
	val := get(path...)
	if val == nil {
		return def
	}
	return val.(string)
}

// MustString returns the string property at the path
func MustString(path ...string) string {
	return mustGet(path...).(string)
}

// Duration returns the string property at the path, as a time.Duration
func Duration(def time.Duration, path ...string) time.Duration {
	s := String(hey, path...)
	if s == hey {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Infof("Failed to parse duration '%s': %s", s, err)
		return def
	}
	return d
}

// MustDuration returns the string property at the path, as a time.Duration
func MustDuration(path ...string) time.Duration {
	s := MustString(path...)
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Infof("Failed to parse duration '%s': %s", s, err)
	}
	return d
}

// MustStringArray returns the string array property at the path
func MustStringArray(path ...string) []string {
	a := mustGet(path...).([]interface{})
	b := make([]string, len(a))
	for i := range a {
		b[i] = a[i].(string)
	}
	return b
}

// Int returns the integer property at the path, with a default
func Int(def int, path ...string) int {
	val := get(path...)
	if val == nil {
		return def
	}
	return int(val.(float64))
}

// MustInt returns the string property at the path
func MustInt(path ...string) int {
	return int(mustGet(path...).(float64))
}

func MustFloat(path ...string) float64 {
	return mustGet(path...).(float64)
}

// Bool returns the boolean property at the path, with a default
func Bool(def bool, path ...string) bool {
	val := get(path...)
	if val == nil {
		return def
	}
	return val.(bool)
}

// MustBool returns the boolean property at the path
func MustBool(path ...string) bool {
	return mustGet(path...).(bool)
}

var hey = "what's up buddy?"

func HasString(path ...string) bool {
	val, ok := config[strings.Join(path, ".")]
	if !ok {
		return false
	}

	_, ok = val.(string)
	return ok
}

func mustGet(path ...string) interface{} {
	val, ok := config[strings.Join(path, ".")]
	if !ok {
		log.Errorf("expected value for %v but found nothing", path)
		panic(fmt.Errorf("expected value for %v but found nothing", path))
	}
	return val
}

func get(path ...string) interface{} {
	val, ok := config[strings.Join(path, ".")]
	if !ok {
		return nil
	}
	return val
}

func MustRefresh() {

	flat := make(map[string]interface{})

	// cli overrides
	addArgs(flat)

	// load environments (no value args) from cli args
	environments := []string{"default"}
	for name, value := range flat {
		if value == nil {
			environments = append(environments, name)
		} else {
			if ok, boolValue := value.(bool); ok {
				if boolValue {
					environments = append(environments, name)
				}
			}
		}
	}

	log.Infof("Environments: %s", strings.Join(environments, ", "))

	flat["env"] = environments

	userHome := getUserHome()

	// env vars (if starting with "sphere_")
	addEnv(flat)

	// anything that can be parsed as a number, is a number
	parseNumbers(flat)

	installDir := "/opt/ninjablocks"
	if val, ok := flat["installDirectory"]; ok {
		installDir = val.(string)
	}

	if _, err := os.Stat(installDir); err != nil {
		// check for installation in snappy, apply different default path
		snappAppPath := os.Getenv("SNAPP_APP_PATH")
		if snappAppPath != "" {
			installDir = snappAppPath
		}
	}

	flat["installDirectory"] = installDir

	if _, err := os.Stat(installDir); err != nil {
		log.Errorf("Couldn't load sphere install directory. Override with env var sphere_installDirectory. error:%s", err)
		panic(err)
	}

	dataPath := "/data"

	// in snappy, we default to using the snappy data path
	snappDataPath := os.Getenv("SNAPP_APP_DATA_PATH")
	if snappDataPath != "" {
		dataPath = snappDataPath
	}

	// User overrides (json)
	addFile(dataPath+"/config.json", flat)

	// Add site preference overrides
	addFile(dataPath+"/etc/opt/ninja/site-preferences.json", flat)

	// credentials file
	addFile(dataPath+"/etc/opt/ninja/credentials.json", flat)

	// mesh file
	addFile(dataPath+"/etc/opt/ninja/mesh.json", flat)

	// home directory environment(s) config
	for i := len(environments) - 1; i >= 0; i-- {
		addFile(filepath.Join(userHome, ".sphere", environments[i]+".json"), flat)
	}

	// current directory environment(s) config
	for i := len(environments) - 1; i >= 0; i-- {
		addFile(filepath.Join(".", "config", environments[i]+".json"), flat)
	}

	// common environment(s) config
	for i := len(environments) - 1; i >= 0; i-- {
		addFile(filepath.Join(installDir, "config", environments[i]+".json"), flat)
	}

	// common credentials config
	addFile(filepath.Join(installDir, "config", "credentials.json"), flat)

	//log.Debugf("Loaded config: %v", flat)

	config = flat
}

func addEnv(config map[string]interface{}) {
	for _, v := range os.Environ() {
		split := strings.Split(v, "=")

		name, value := split[0], split[1]

		if strings.HasPrefix(name, "sphere_") {
			name = strings.TrimPrefix(name, "sphere_")
			name = strings.Replace(name, "_", ".", 0)

			if _, ok := config[name]; !ok {
				config[name] = value
			}
		}
	}
}

func addArgs(config map[string]interface{}) {

	parser := kingpin.Tokenize(os.Args[1:]).Tokens

	for token, parser := parser.Peek(), parser.Next(); token.Type != kingpin.TokenEOL; token, parser = parser.Peek(), parser.Next() {

		if token.IsFlag() {
			var value interface{}
			name := token.Value

			next := parser.Peek()
			if next.Type == kingpin.TokenArg {
				if next.Value == "false" {
					value = false
				} else if next.Value == "true" {
					value = true
				} else {
					value = next.Value
				}
			} else {
				value = true
				// It's an environment indicator... like --cloud-production
			}

			config[name] = value
		}

	}

}

func addFile(path string, config map[string]interface{}) error {
	//log.Debugf("Loading config file: %s", path)

	file, e := ioutil.ReadFile(path)
	if e != nil {
		return fmt.Errorf("Failed to load file: %s error: %s", path, e)
	}

	content := make(map[string]interface{})
	e = json.Unmarshal(file, &content)
	if e != nil {
		return fmt.Errorf("Failed to read file: %s error: %s", path, e)
	}

	//spew.Dump(path, content)

	flatten(content, nil, config)
	return nil
}

func flatten(input interface{}, lpath []string, flattened map[string]interface{}) {
	if lpath == nil {
		lpath = []string{}
	}

	if reflect.ValueOf(input).Kind() == reflect.Map {
		for rkey, value := range input.(map[string]interface{}) {
			flatten(value, append(lpath, rkey), flattened)
		}
	} else {
		if _, ok := flattened[strings.Join(lpath, ".")]; !ok {
			flattened[strings.Join(lpath, ".")] = input
		}
	}
}

func unflatten(flat map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	obj := out

	for key, val := range flat {

		obj = out
		keys := strings.Split(key, ".")

		for i, k := range keys {
			if i == len(keys)-1 {
				obj[k] = val
			} else {
				next, ok := obj[k]
				if !ok {
					next = make(map[string]interface{})
				}
				obj[k] = next
				obj = next.(map[string]interface{})
			}
		}

	}

	return out
}

func parseNumbers(config map[string]interface{}) {
	for name, val := range config {
		stringVal, ok := val.(string)
		if ok {
			floatVal, err := strconv.ParseFloat(stringVal, 64)
			if err == nil {
				config[name] = floatVal
			}
		}
	}
}

func getUserHome() string {
	usr, err := user.Current()
	if err != nil {
		return "/root"
	}
	return usr.HomeDir
}

func IsSlave() bool {
	masterNodeId := String("", "masterNodeId")
	serial := Serial()
	return masterNodeId != "" && serial != masterNodeId
}

func IsMaster() bool {
	masterNodeId := String("", "masterNodeId")
	serial := Serial()
	return masterNodeId != "" && serial == masterNodeId
}

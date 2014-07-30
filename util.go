package ninja

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	// "log"
	"os"
	"os/exec"

	"github.com/bitly/go-simplejson"
	"github.com/bugsnag/bugsnag-go"
)

func init() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "a39d43b795d60d16b1d6099236f5825e",
	})
}

func GetSerial() (string, error) {

	var cmd *exec.Cmd

	if Exists("/opt/ninjablocks/bin/sphere-serial") {
		cmd = exec.Command("/opt/ninjablocks/bin/sphere-serial")
	} else {
		cmd = exec.Command("sphere-serial")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func GetConfig() (*simplejson.Json, error) {
	var cmd *exec.Cmd
	if Exists("/opt/ninjablocks/bin/sphere-config") {
		cmd = exec.Command("/opt/ninjablocks/bin/sphere-config")
	} else {
		cmd = exec.Command("sphere-config")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(out.Bytes())
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func GetGUID(in string) string {
	h := md5.New()
	h.Write([]byte(in))
	str := hex.EncodeToString(h.Sum(nil))
	return str[:10]
}

func strArrayToJson(in []string) (*simplejson.Json, error) {
	str := "[ "
	for i, item := range in {
		if i < (len(in) - 1) { //commas between elements except for last item
			str += "\"" + item + "\", "
		} else {
			str += "\"" + item + "\" ]"
		}
	}

	out, err := simplejson.NewJson([]byte(str))
	if err != nil {
		return nil, err
	}

	return out, nil
}

func getDriverInfo(filename string) (*simplejson.Json, error) {

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	js, err := simplejson.NewJson(dat)
	if err != nil {
		return nil, err
	}

	js.Del("scripts")
	return js, nil
}

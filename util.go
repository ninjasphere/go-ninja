package ninja

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/bitly/go-simplejson"
)

func GetSerial() string {

	var cmd *exec.Cmd

	if Exists("/opt/ninjablocks/bin/sphere-serial") {
		cmd = exec.Command("/opt/ninjablocks/bin/sphere-serial")
	} else {
		cmd = exec.Command("./sphere-serial")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

func GetConfig() (*simplejson.Json, error) {
	var cmd *exec.Cmd
	if Exists("/opt/ninjablocks/bin/sphere-config") {
		cmd = exec.Command("/opt/ninjablocks/bin/sphere-config")
	} else {
		cmd = exec.Command("./sphere-config")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
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

func strArrayToJson(in []string) *simplejson.Json {
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
		log.Fatalf("Bad JSON in strArrayToJson %+v: %s", in, err)
	}

	return out
}

func getDriverInfo(filename string) (res *simplejson.Json) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Unable to get driver info from %s. error: ", filename, err)
	}
	js, err := simplejson.NewJson(dat)
	if err != nil {
		log.Fatalf("Malformed JSON in driver info: %s, error: %s ", dat, err)
	}

	js.Del("scripts")
	return js
}

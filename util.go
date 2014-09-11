package ninja

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/bitly/go-simplejson"
)

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
			str += "\"" + item + "\""
		}
	}
	str += " ]"

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

func GetNetAddress() (string, error) {
	var ipAddr string

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Errorf("Failed to get interfaces: %s", err)
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Errorf("Failed to get addresses: %s", err)
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if addr.String() != "127.0.0.1/8" && addr.String() != "::1/128" {
					rawAddy := addr.String()
					ipAddr = strings.Split(rawAddy, "/")[0]
				}
			default:
				fmt.Printf("unexpected type %T", v) // %T prints whatever type t has
			}

		}
	}

	return ipAddr, nil

}

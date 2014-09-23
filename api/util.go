package ninja

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/bitly/go-simplejson"
)

func getGUID(in ...string) string {
	h := md5.New()
	for _, s := range in {
		h.Write([]byte(s))
	}
	str := hex.EncodeToString(h.Sum(nil))
	return str[:10]
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

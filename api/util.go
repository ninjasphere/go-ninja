package ninja

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"github.com/ninjasphere/go-ninja/model"
)

func getGUID(in ...string) string {
	h := md5.New()
	for _, s := range in {
		h.Write([]byte(s))
	}
	str := hex.EncodeToString(h.Sum(nil))
	return str[:10]
}

func LoadModuleInfo(filename string) *model.Module {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read module info file '%s': %s", filename, err)
	}

	var info model.Module
	err = json.Unmarshal(data, &info)
	if err != nil {
		log.Fatalf("Failed to parse module info file '%s': %s", filename, err)
	}

	return &info
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

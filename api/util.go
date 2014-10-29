package ninja

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"

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
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

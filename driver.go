package ninja

import (
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
	"github.com/ninjasphere/go-ninja/logger"
)

// DriverBus Context for driver related announcements.
type DriverBus struct {
	id      string
	name    string
	version string
	mqtt    *MQTT.MqttClient
	log     *logger.Logger
}

// AnnounceDevice Announce a new device has been discovered.
func (d *DriverBus) AnnounceDevice(id string, idType string, name string, sigs *simplejson.Json) (*DeviceBus, error) {
	js, err := simplejson.NewJson([]byte(`{
    "params": [
        {
            "guid": "",
            "id": "",
            "idType": "",
            "name": "",
            "signatures": {},
            "driver": {
                "name": "",
                "version": ""
            }
        }
    ],
    "time": "",
    "jsonrpc": "2.0"
}`))

	if err != nil {
		return nil, err
	}

	guid := GetGUID(d.id + id)
	js.Get("params").GetIndex(0).Set("guid", guid)
	js.Get("params").GetIndex(0).Set("id", id) //TODO patch driver to get MAC ID, rather than numberical ID
	js.Get("params").GetIndex(0).Set("idType", idType)
	js.Get("params").GetIndex(0).Set("name", name)
	js.Get("params").GetIndex(0).Set("signatures", sigs)
	js.Get("params").GetIndex(0).Get("driver").Set("name", d.name)
	js.Get("params").GetIndex(0).Get("driver").Set("version", d.version)
	js.Set("time", time.Now().Unix())

	json, err := js.MarshalJSON()
	if err != nil {
		return nil, err
	}

	d.log.Infof("Outging announcement : %s", json)

	receipt := d.mqtt.Publish(MQTT.QoS(1), "$device/"+guid+"/announce", json)
	<-receipt

	deviceBus := NewDeviceBus(id, idType, name, d, js.Get("params").GetIndex(0))

	return deviceBus, nil
}

package main

import (
	"log"
	"time"

	"github.com/bitly/go-simplejson"
)

type ChannelBus struct {
	name     string
	protocol string
	device   *DeviceBus
	channel  <-chan MQTT.Receipt
}

type JsonMessageHandler func(string, *simplejson.Json)

// $device/7f0fa623af/channel/d00f681ad1/core.batching/announce
func (d *DeviceBus) AnnounceChannel(name string, protocol string, methods []string, events []string, serviceCallback JsonMessageHandler) (*ChannelBus, error) {
	deviceguid, _ := d.devicejson.Get("guid").String()
	channelguid := GetGUID(name + protocol)
	js, _ := simplejson.NewJson([]byte(`{
    "params": [
          {
            "channel": "",
            "supported": {
                "methods": [],
                "events": []
            },
            "device": {}
        }
    ],
    "time": "",
    "jsonrpc": "2.0"
}`))

	js.Get("params").GetIndex(0).Set("device", d.devicejson)
	methodsjson := strArrayToJson(methods)
	js.Get("params").GetIndex(0).Get("supported").Set("methods", methodsjson)
	eventsjson := strArrayToJson(events)
	js.Get("params").GetIndex(0).Get("supported").Set("events", eventsjson)
	js.Get("params").GetIndex(0).Set("channel", name)
	js.Set("time", time.Now().Unix())

	json, err := js.MarshalJSON()

	if err != nil {
		log.Fatalf("Couldn't stringify that message %s", err)
	}

	topicBase := "$device/" + deviceguid + "/channel/" + channelguid + "/" + protocol

	pubReceipt := d.driver.mqtt.Publish(MQTT.QoS(0), topicBase+"/announce", json)
	<-pubReceipt

	log.Printf("Subscribing to : %s", topicBase)
	filter, err := MQTT.NewTopicFilter(topicBase, 0)
	if err != nil {
		log.Fatalf("unable to subscribe to %s in announcechannel: %s", topicBase, err)
	}
	_, err = d.driver.mqtt.StartSubscription(func(client *MQTT.MqttClient, message MQTT.Message) {
		json, _ := simplejson.NewJson(message.Payload())
		method, _ := json.Get("method").String()
		params := json.Get("params")
		serviceCallback(method, params)

	}, filter)

	if err != nil {
		log.Fatal(err)
	}

	channelBus := &ChannelBus{
		name:     name,
		protocol: protocol,
		device:   d,
	}

	return channelBus, nil
}

func (cb *ChannelBus) SendEvent(event string, payload *simplejson.Json) error {
	json, err := payload.MarshalJSON()
	if err != nil {
		return err
	}

	receipt := cb.device.driver.mqtt.Publish(MQTT.QoS(0), "$driver/"+cb.device.driver.id+"/device/"+cb.device.id+"/channel/"+cb.name+"/"+cb.protocol+"/event/"+event, json)
	<-receipt

	return nil
}

func (n *NinjaConnection) AnnounceDriver(id string, name string, path string) (*DriverBus, error) {
	js, err := simplejson.NewJson([]byte(`{
    "params": [
    {
      "name": "",
      "file": "",
      "defaultConfig" : {},
      "package": {}
    }],
    "time":"",
    "jsonrpc":"2.0"
  }`))

	if err != nil {
		log.Fatalf("Bad json: %s", err)
	}

	driverinfofile := path + "package.json"
	pkginfo := getDriverInfo(driverinfofile)
	filename, err := pkginfo.Get("main").String()
	if err != nil {
		log.Fatalf("Couldn't retrieve main filename: %s", err)
	}

	mainfile := path + filename
	js.Get("params").GetIndex(0).Set("file", mainfile)
	js.Get("params").GetIndex(0).Set("name", id)
	js.Get("params").GetIndex(0).Set("package", pkginfo)
	js.Get("params").GetIndex(0).Set("defaultConfig", "{}") //TODO fill me out
	js.Set("time", time.Now().Unix())
	json, _ := js.MarshalJSON()

	serial := GetSerial()
	version, err := pkginfo.Get("version").String()
	if err != nil {
		log.Fatalf("No version available for driver %s: %s", id, err)
	}

	receipt := n.mqtt.Publish(MQTT.QoS(1), "$node/"+serial+"/app/"+id+"/event/announce", json)
	<-receipt

	driverBus := &DriverBus{
		id:      id,
		name:    name,
		mqtt:    n.mqtt,
		version: version,
	}

	return driverBus, nil
}

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
		log.Fatalf("Bad driver announce JSON: %s", js)
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
		log.Fatalf("Couldn't stringify: %s", err)
	}

	receipt := d.mqtt.Publish(MQTT.QoS(1), "$device/"+guid+"/announce/", json)
	<-receipt

	deviceBus := &DeviceBus{
		id:         id,
		idType:     idType,
		name:       name,
		driver:     d,
		devicejson: js.Get("params").GetIndex(0),
	}

	return deviceBus, nil
}

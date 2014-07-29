package ninja

import (
	"log"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
)

// DeviceBus Context for device related announcements.
type DeviceBus struct {
	id         string
	idType     string
	name       string
	driver     *DriverBus
	devicejson *simplejson.Json
}

// AnnounceChannel Announce a new channel has been created.
func (d *DeviceBus) AnnounceChannel(name string, protocol string, methods []string, events []string, serviceCallback JsonMessageHandler) (*ChannelBus, error) {

	// $device/7f0fa623af/channel/d00f681ad1/core.batching/announce

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
	methodsjson, err := strArrayToJson(methods)
	if err != nil {
		return nil, err
	}

	js.Get("params").GetIndex(0).Get("supported").Set("methods", methodsjson)

	eventsjson, err := strArrayToJson(events)
	if err != nil {
		return nil, err
	}

	js.Get("params").GetIndex(0).Get("supported").Set("events", eventsjson)
	js.Get("params").GetIndex(0).Set("channel", name)
	js.Set("time", time.Now().Unix())

	json, err := js.MarshalJSON()
	if err != nil {
		return nil, err
	}

	topicBase := "$device/" + deviceguid + "/channel/" + channelguid + "/" + protocol
	pubReceipt := d.driver.mqtt.Publish(MQTT.QoS(0), topicBase+"/announce", json)
	<-pubReceipt
	log.Printf("Subscribing to : %s", topicBase)
	filter, err := MQTT.NewTopicFilter(topicBase, 0)
	if err != nil {
		return nil, err
	}
	_, err = d.driver.mqtt.StartSubscription(func(client *MQTT.MqttClient, message MQTT.Message) {
		json, _ := simplejson.NewJson(message.Payload())
		method, _ := json.Get("method").String()
		params := json.Get("params")
		serviceCallback(method, params)

	}, filter)

	if err != nil {
		return nil, err
	}

	channelBus := &ChannelBus{
		name:     name,
		protocol: protocol,
		device:   d,
	}

	return channelBus, nil
}

package ninja

import (
	"fmt"
	"log"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
	bugsnag "github.com/bugsnag/bugsnag-go"
)

type NinjaConnection struct {
	mqtt *MQTT.MqttClient
}

type DriverBus struct {
	id      string
	name    string
	version string
	mqtt    *MQTT.MqttClient
}

type DeviceBus struct {
	id         string
	idType     string
	name       string
	driver     *DriverBus
	devicejson *simplejson.Json
}

func Connect(clientId string) (*NinjaConnection, error) {

	mqttUrl, err := GetMQTTUrl()
	if err != nil {
		bugsnag.Notify(err)
		return nil, err
	}

	conn := NinjaConnection{}
	opts := MQTT.NewClientOptions().SetBroker(mqttUrl).SetClientId(clientId).SetCleanSession(true).SetTraceLevel(MQTT.Off)
	conn.mqtt = MQTT.NewClient(opts)

	if _, err := conn.mqtt.Start(); err != nil {
		bugsnag.Notify(err)
		return nil, err
	}

	log.Printf("Connected to %s\n", mqttUrl)
	return &conn, nil
}

func GetMQTTUrl() (url string, err error) {

	var host string
	var port int

	cfg, err := GetConfig()
	if err != nil {
		bugsnag.Notify(err)
		return
	}

	mqttConfig := cfg.Get("mqtt")
	if host, err = mqttConfig.Get("host").String(); err != nil {
		bugsnag.Notify(err)
		return
	}

	if port, err = mqttConfig.Get("port").Int(); err != nil {
		bugsnag.Notify(err)
		return
	}
	url = fmt.Sprintf("tcp://%s:%d", host, port)
	return
}

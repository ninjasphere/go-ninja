package ninja

import (
	"fmt"
	"log"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
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
		return nil, err
	}

	conn := NinjaConnection{}
	opts := MQTT.NewClientOptions().SetBroker(mqttUrl).SetClientId(clientId).SetCleanSession(true).SetTraceLevel(MQTT.Off)
	conn.mqtt = MQTT.NewClient(opts)

	if _, err := conn.mqtt.Start(); err != nil {
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
		return "", err
	}

	mqttConfig := cfg.Get("mqtt")
	if host, err = mqttConfig.Get("host").String(); err != nil {
		return "", err
	}

	if port, err = mqttConfig.Get("port").Int(); err != nil {
		return "", err
	}
	url = fmt.Sprintf("tcp://%s:%d", host, port)
	return url, nil
}

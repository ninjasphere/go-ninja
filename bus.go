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

	host, port := GetMQTTAddress()
	mqttServer := fmt.Sprintf("tcp://%s:%d", host, port)
	conn := NinjaConnection{}
	opts := MQTT.NewClientOptions().SetBroker(mqttServer).SetClientId(clientId).SetCleanSession(true).SetTraceLevel(MQTT.Off)
	conn.mqtt = MQTT.NewClient(opts)
	_, err := conn.mqtt.Start()
	if err != nil {
		log.Fatalf("Failed to connect to mqtt server %s - %s", host, err)
	} else {
		log.Printf("Connected to %s\n", host)
	}
	return &conn, nil
}

func GetMQTTAddress() (host string, port int) {

	cfg, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	mqtt := cfg.Get("mqtt")
	if host, err = mqtt.Get("host").String(); err != nil {
		log.Fatal(err)
	}
	if port, err = mqtt.Get("port").Int(); err != nil {
		log.Fatal(err)
	}

	return host, port

}

package ninja

import (
	"fmt"
	"log"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
	bugsnag "github.com/bugsnag/bugsnag-go"
)

func init() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "a39d43b795d60d16b1d6099236f5825e",
	})
}

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
		bugsnag.Notify(err)
		log.Fatalf("Failed to connect to mqtt server %s - %s", host, err)
	} else {
		log.Printf("Connected to %s\n", host)
	}
	return &conn, nil
}

func GetMQTTAddress() (host string, port int) {

	cfg, err := GetConfig()
	if err != nil {
		bugsnag.Notify(err)
		log.Fatal(err)
	}

	mqtt := cfg.Get("mqtt")
	if host, err = mqtt.Get("host").String(); err != nil {
		bugsnag.Notify(err)
		log.Fatal(err)
	}
	if port, err = mqtt.Get("port").Int(); err != nil {
		bugsnag.Notify(err)
		log.Fatal(err)
	}

	return host, port

}

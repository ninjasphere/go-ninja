// +build ignore

package main

import (
	"log"
	"net/rpc"
	"os"
	"os/signal"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/rpc2"
)

// mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test'
// while true; do mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test';  done;

type Thing struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Device Device
}

type Device struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	IDType   string `json:"idType"`
	Guid     string `json:"guid"`
	Channels []Channel
}

type Channel struct {
	Protocol string `json:"protocol"`
	Name     string `json:"channel"`
	ID       string `json:"id"`
}

func main() {

	nconn, err := ninja.Connect("Hey")
	if err != nil {
		log.Fatalf("Couldn't connect to mqtt: %s", err)
	}

	// You need to export the mqtt connection here if you want to test it.
	client, err := rpc2.GetClient("$home/services/ThingModel", nconn.Mqtt)

	if err != nil {
		log.Fatalf("Failed getting rpc2 client %s", err)
	}

	//time.Sleep(time.Second * 3)

	var things []Thing

	err = client.Call("fetchByType", "light", &things)
	//err = client.Call("fetch", "c7ac05e0-9999-4d93-bfe3-a0b4bb5e7e78", &thing)

	if err != nil {
		log.Fatalf("Failed calling fetch method: %s", err)
	}

	log.Printf("Done")
	spew.Dump(things)

	for _, thing := range things {
		onOffClient, err := GetChannelClient(&thing, "on-off", nconn.Mqtt)
		if err != nil {
			log.Fatalf("Failed getting on/off client for thing %s: %s", thing.ID, err)
		}

		if onOffClient != nil {
			log.Printf("Found on-off on thing %s", thing.ID)

			_ = onOffClient.Go("turnOn", nil, nil, nil)

		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Println("Got signal:", s)

}

func GetChannelClient(thing *Thing, protocol string, mqtt *mqtt.MqttClient) (*rpc.Client, error) {

	for _, channel := range thing.Device.Channels {
		if channel.Protocol == protocol {
			topic := "$device/" + thing.Device.Guid + "/channel/" + channel.ID + "/" + protocol
			return rpc2.GetClient(topic, mqtt)
		}
	}

	return nil, nil
}

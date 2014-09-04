// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/rpc3"
	"github.com/ninjasphere/go-ninja/rpc3/json2"
)

type TestService struct {
	SendEvent func(payload interface{}, event string) error
}

func (t *TestService) SayHello(msg mqtt.Message, name *string, reply *string) error {
	*reply = fmt.Sprintf("Hey there %s!", *name)
	return nil
}

// mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test'
// while true; do mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test';  done;

func main() {

	nconn, err := ninja.Connect("Hey")
	if err != nil {
		log.Fatalf("Couldn't connect to mqtt: %s", err)
	}

	service := &TestService{}
	server := rpc.NewServer(nconn.GetMqttClient(), json2.NewCodec())
	sendEvent, err := server.RegisterService(service, "rpc/test")

	if err != nil {
		log.Fatalf("Failed to register service: %s", err)
	}

	client := rpc.NewClient(nconn.GetMqttClient(), json2.NewClientCodec())

	var response string
	call, err := client.Call("rpc/test", "sayHello", "Erriot", &response)
	if err != nil {
		log.Fatalf("Failed to call service: %s", err)
	}

	<-call.Done

	if call.Error != nil {
		log.Fatalf("Error received from service, or caused at reply: %s", call.Error)
	}
	log.Printf("Response: %s", response)

	time.Sleep(time.Second * 3)

	type testEvent struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	sendEvent("state", &testEvent{
		Name: "Elliot",
		Age:  30,
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Println("Got signal:", s)

}

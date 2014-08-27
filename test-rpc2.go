// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/rpc2"
)

type TestService struct {
	SendEvent func(payload interface{}, event string) error
}

func (t *TestService) SayHello(name *string, reply *string) error {
	*reply = fmt.Sprintf("Hey there %s!", *name)
	return nil
}

func (t *TestService) SetEventHandler(handler func(payload interface{}, event string) error) {
	t.SendEvent = handler
}

// mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test'
// while true; do mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test';  done;

func main() {

	nconn, err := ninja.Connect("Hey")
	if err != nil {
		log.Fatalf("Couldn't connect to mqtt: %s", err)
	}

	service := &TestService{}

	// You need to export the mqtt connection here if you want to test it.
	rpc2.ExportService(service, "rpc/test", nconn.Mqtt)

	time.Sleep(time.Second * 3)

	type testEvent struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	service.SendEvent(&testEvent{
		Name: "Elliot",
		Age:  30,
	}, "state")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Println("Got signal:", s)

}

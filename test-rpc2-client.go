// +build ignore

package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-ninja"
	"github.com/ninjasphere/go-ninja/rpc2"
)

// mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test'
// while true; do mosquitto_pub -m '{"id":123, "params": ["Elliot"],"jsonrpc": "2.0","method":"sayHello","time":132123123}' -t 'rpc/test';  done;

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

	type testEvent struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var reply string

	client.Go("IGNOREME", nil, &reply, nil)

	var thing json.RawMessage

	err = client.Call("fetch", "c7ac05e0-9999-4d93-bfe3-a0b4bb5e7e78", &thing)

	if err != nil {
		log.Fatalf("Failed calling fetch method: %s", err)
	}

	log.Printf("Done")
	spew.Dump(thing)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Println("Got signal:", s)

}

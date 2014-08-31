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

	var things json.RawMessage

	err = client.Call("fetchAll", nil, &things)

	if err != nil {
		log.Fatalf("Failed calling fetch method: %s", err)
	}

	log.Printf("Done")
	spew.Dump(things)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	log.Println("Got signal:", s)

}

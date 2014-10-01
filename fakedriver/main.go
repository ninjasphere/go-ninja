package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
)

// mosquitto_pub -t '$node/OSXC02LW6Z2FH01/driver/com.ninjablocks.fakedriver' -m '{"jsonrpc":"2.0","method":"start","params":[{"NumberOfDevices":29}]}'

func main() {

	_, err := NewFakeDriver()

	if err != nil {
		log.Fatalf("Failed to create fake driver: %s", err)
	}

	//spew.Dump(light)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

}

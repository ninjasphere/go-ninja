package bus

import (
	"os"
	"os/signal"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestWhatever(t *testing.T) {

	bus := MustConnect("localhost:1883", "test")
	bus.Subscribe("$device/#", func(topic string, payload []byte) {
		spew.Dump("message!", topic, payload)
	})

	blah := make(chan os.Signal, 1)
	signal.Notify(make(chan os.Signal, 1), os.Interrupt, os.Kill)
	log.Infof("Got signal: %v", <-blah)

}

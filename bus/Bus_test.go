package bus

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-ninja/support"
)

func TestWhatever(t *testing.T) {

	bus := Connect("localhost:1883", "test")
	bus.Subscribe("#", func(topic string, payload []byte) {
		spew.Dump("message!", message)
	})

	support.WaitUntilSignal()
}

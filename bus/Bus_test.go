package bus

import (
	"testing"
	"time"
)

func TestSimplePubSub(t *testing.T) {

	done := make(chan bool)
	topic := "testing/what/ever"
	payload := `{"hello":123}`

	bus := MustConnect("localhost:1883", "BusTest")
	bus.Subscribe("testing/#", func(t string, p []byte) {
		done <- topic == topic && string(p) == payload
	})

	bus.Publish(topic, []byte(payload))

	select {
	case success := <-done:
		if !success {
			t.Failed()
		}
	case <-time.After(time.Second):
		t.Failed()
	}

}

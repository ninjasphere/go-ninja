package bus

import (
	"fmt"
	"net"
	"os"
	"strings"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("bus")

type Bus struct {
	mqtt          *mqtt.ClientConn
	subscriptions []*subscription
}

type subscription struct {
	topic    string
	callback func(topic string, payload []byte)
}

func Connect(host, id string) *Bus {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(2)
	}
	mqtt := mqtt.NewClientConn(conn)
	mqtt.ClientId = id

	err = mqtt.Connect("", "")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	bus := &Bus{
		mqtt:          mqtt,
		subscriptions: make([]*subscription, 0),
	}

	go func() {
		for m := range bus.mqtt.Incoming {
			bus.onIncoming(m)
		}
	}()

	return bus
}

func (b *Bus) onIncoming(message *proto.Publish) {
	for _, sub := range b.subscriptions {
		if matches(sub.topic, message.TopicName) {
			go sub.callback(message.TopicName, []byte(message.Payload.(proto.BytesPayload)))
		}
	}
}

func (b *Bus) Publish(topic string, payload []byte) {

	b.mqtt.Publish(&proto.Publish{
		TopicName: topic,
		Payload:   proto.BytesPayload(payload),
	})

}

func (b *Bus) Subscribe(topic string, callback func(topic string, payload []byte)) (*subscription, error) {

	subscription := &subscription{
		topic:    topic,
		callback: callback,
	}

	err := b.subscribe(subscription)
	if err != nil {
		return nil, err
	}

	b.subscriptions = append(b.subscriptions, subscription)

	return subscription, nil
}

func (b *Bus) subscribe(subscription *subscription) error {
	_ = b.mqtt.Subscribe([]proto.TopicQos{proto.TopicQos{subscription.topic, proto.QosAtMostOnce}})
	//spew.Dump("subscription ack", ack)
	// TODO: Check ack
	return nil
}

func matches(subscription string, topic string) bool {
	parts := strings.Split(topic, "/")
	subParts := strings.Split(subscription, "/")

	i := 0
	for i < len(parts) {
		// topic is longer, no match
		if i >= len(subParts) {
			return false
		}
		// matched up to here, and now the wildcard says "all others will match"
		if subParts[i] == "#" {
			return true
		}
		// text does not match, and there wasn't a + to excuse it
		if parts[i] != subParts[i] && subParts[i] != "+" {
			return false
		}
		i++
	}

	// make finance/stock/ibm/# match finance/stock/ibm
	if i == len(subParts)-1 && subParts[len(subParts)-1] == "#" {
		return true
	}

	if i == len(subParts) {
		return true
	}
	return false
}

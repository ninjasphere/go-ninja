package bus

import (
	"fmt"
	"net"

	proto "github.com/huin/mqtt"
	"github.com/jeffallen/mqtt"
)

type TinyBus struct {
	mqtt          *mqtt.ClientConn
	subscriptions []*subscription
}

func ConnectTinyBus(host, id string) (*TinyBus, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to: %s", err)
	}
	mqtt := mqtt.NewClientConn(conn)
	mqtt.ClientId = id

	err = mqtt.Connect("", "")
	if err != nil {
		return nil, fmt.Errorf("MQTT Failed to connect to: %s", err)
	}

	bus := &TinyBus{
		mqtt:          mqtt,
		subscriptions: make([]*subscription, 0),
	}

	go func() {
		for m := range bus.mqtt.Incoming {
			bus.onIncoming(m)
		}
	}()

	return bus, nil
}

func (b *TinyBus) onIncoming(message *proto.Publish) {
	for _, sub := range b.subscriptions {
		if matches(sub.topic, message.TopicName) {
			go sub.callback(message.TopicName, []byte(message.Payload.(proto.BytesPayload)))
		}
	}
}

func (b *TinyBus) Publish(topic string, payload []byte) {

	b.mqtt.Publish(&proto.Publish{
		TopicName: topic,
		Payload:   proto.BytesPayload(payload),
	})

}

func (b *TinyBus) Subscribe(topic string, callback func(topic string, payload []byte)) (*subscription, error) {

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

func (b *TinyBus) subscribe(subscription *subscription) error {
	_ = b.mqtt.Subscribe([]proto.TopicQos{proto.TopicQos{subscription.topic, proto.QosAtMostOnce}})
	//spew.Dump("subscription ack", ack)
	// TODO: Check ack
	return nil
}

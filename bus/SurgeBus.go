package bus

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/surge/surgemq/message"
	"github.com/surge/surgemq/service"
)

type SurgeBus struct {
	mqtt          *service.Client
	subscriptions []*subscription
}

func ConnectSurgeBus(host, id string) (*SurgeBus, error) {

	svc, err := service.Connect("tcp://"+host, newConnectMessage(id))

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to %s: %s", host, err)
	}

	bus := &SurgeBus{
		mqtt:          svc,
		subscriptions: make([]*subscription, 0),
	}

	return bus, nil
}

func newConnectMessage(cid string) *message.ConnectMessage {
	msg := message.NewConnectMessage()

	msg.SetVersion(4)
	msg.SetCleanSession(true)
	msg.SetClientId([]byte(cid))
	msg.SetKeepAlive(90)

	msg.SetWillQos(1)
	msg.SetWillTopic([]byte("$client/death"))
	msg.SetWillMessage([]byte(cid))

	return msg
}

func (b *SurgeBus) Publish(topic string, payload []byte) {

	msg := message.NewPublishMessage()
	msg.SetTopic([]byte(topic))
	msg.SetPayload(payload)
	msg.SetQoS(0)

	b.mqtt.Publish(msg, func(msg, ack message.Message, err error) {
		spew.Dump("Publish ack", msg, ack, err)
	})

}

func (b *SurgeBus) Subscribe(topic string, callback func(topic string, payload []byte)) (*subscription, error) {

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

func (b *SurgeBus) subscribe(subscription *subscription) error {

	sub := message.NewSubscribeMessage()
	sub.SetPacketId(1)
	sub.AddTopic([]byte(subscription.topic), 0)

	done := make(chan error, 1)

	b.mqtt.Subscribe(sub,
		func(msg, ack message.Message, err error) {
			done <- err
		},
		func(msg *message.PublishMessage) error {
			subscription.callback(string(msg.Topic()), msg.Payload())
			return nil
		})

	return <-done
}

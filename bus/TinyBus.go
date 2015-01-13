package bus

import (
	"fmt"
	"net"
	"sync"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/mqtt"
)

type TinyBus struct {
	connecting    sync.WaitGroup
	mqtt          *mqtt.ClientConn
	subscriptions []*subscription
	host          string
	id            string
}

func ConnectTinyBus(host, id string) (*TinyBus, error) {

	bus := &TinyBus{
		subscriptions: make([]*subscription, 0),
		host:          host,
		id:            id,
	}

	bus.connect()

	return bus, nil
}

type wrappedConn struct {
	net.Conn
	done chan bool
}

func (c wrappedConn) Close() error {
	log.Warningf("Connection closed!")
	c.done <- true
	c.Conn.Close()
	return nil
}

func (b *TinyBus) connect() {

	b.connecting.Add(1)
	defer func() {
		b.connecting.Done()

		b.publish(&proto.Publish{
			Header: proto.Header{
				Retain: true,
			},
			TopicName: fmt.Sprintf("node/%s/module/%s/state/connected", config.Serial(), b.id),
			Payload:   proto.BytesPayload([]byte("true")),
		})
	}()

	var conn wrappedConn
	for {
		tcpConn, err := net.Dial("tcp", b.host)
		if err == nil {
			conn = wrappedConn{
				Conn: tcpConn,
				done: make(chan bool, 1),
			}
			break
		}

		//log.Warningf("Failed to connect to: %s", err)
		time.Sleep(time.Millisecond * 500)
	}

	if b.mqtt != nil {
		log.Infof("Reconnected to mqtt server")
	}

	mqtt := mqtt.NewClientConn(conn)
	mqtt.ClientId = b.id

	err := mqtt.ConnectCustom(&proto.Connect{
		WillFlag:    true,
		WillQos:     0,
		WillRetain:  true,
		WillTopic:   fmt.Sprintf("$node/%s/module/%s/state/connected", config.Serial(), b.id),
		WillMessage: "false",
	})

	if err != nil {
		log.Fatalf("MQTT Failed to connect to: %s", err)
	}

	b.mqtt = mqtt

	for _, s := range b.subscriptions {
		b.subscribe(s)
	}

	go func() {
		for m := range mqtt.Incoming {
			b.onIncoming(m)
		}
	}()

	go func() {
		<-conn.done
		b.connect()
	}()
}

func (b *TinyBus) onIncoming(message *proto.Publish) {
	for _, sub := range b.subscriptions {
		if matches(sub.topic, message.TopicName) {
			go sub.callback(message.TopicName, []byte(message.Payload.(proto.BytesPayload)))
		}
	}
}

func (b *TinyBus) Publish(topic string, payload []byte) {
	b.connecting.Wait()

	b.publish(&proto.Publish{
		TopicName: topic,
		Payload:   proto.BytesPayload(payload),
	})

}

func (b *TinyBus) publish(message *proto.Publish) {
	b.mqtt.Publish(message)
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

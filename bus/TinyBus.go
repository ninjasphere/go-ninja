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
	baseBus
	connecting    sync.WaitGroup
	mqtt          *mqtt.ClientConn
	subscriptions []*Subscription
	host          string
	id            string
}

func ConnectTinyBus(host, id string) (*TinyBus, error) {

	bus := &TinyBus{
		subscriptions: make([]*Subscription, 0),
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

	b.connected()

	for _, s := range b.subscriptions {
		if !s.cancelled {
			b.subscribe(s)
		}
	}

	go func() {
		for m := range mqtt.Incoming {
			b.onIncoming(m)
		}
	}()

	go func() {
		<-conn.done
		b.disconnected()
		if !b.destroyed {
			b.connect()
		}
	}()
}

func (b *TinyBus) onIncoming(message *proto.Publish) {
	for _, sub := range b.subscriptions {
		if !sub.cancelled && matches(sub.topic, message.TopicName) {
			go sub.callback(message.TopicName, []byte(message.Payload.(proto.BytesPayload)))
		}
	}
}

func (b *TinyBus) Destroy() {
	log.Infof("Destroy called")
	b.destroyed = true
	b.mqtt.Disconnect()
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

func (b *TinyBus) Subscribe(topic string, callback func(topic string, payload []byte)) (*Subscription, error) {

	subscription := &Subscription{
		topic:    topic,
		callback: callback,
	}

	subscription.Cancel = func() {
		// TODO: Actually unsubscribe if we were the only one listening
		// TODO: Remove from from b.subscriptions
		subscription.cancelled = true
	}

	err := b.subscribe(subscription)
	if err != nil {
		return nil, err
	}

	b.subscriptions = append(b.subscriptions, subscription)

	return subscription, nil
}

func (b *TinyBus) subscribe(subscription *Subscription) error {
	_ = b.mqtt.Subscribe([]proto.TopicQos{proto.TopicQos{subscription.topic, proto.QosAtMostOnce}})
	//spew.Dump("subscription ack", ack)
	// TODO: Check ack
	return nil
}

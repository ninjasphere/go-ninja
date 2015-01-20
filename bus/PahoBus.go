package bus

// TODO: Locking!

import (
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/org.eclipse.paho.mqtt.golang"
)

type PahoBus struct {
	baseBus
	host          string
	id            string
	mqtt          *mqtt.MqttClient
	subscriptions []*subscription
	reconnectLock sync.Mutex
}

func ConnectPahoBus(host, id string) (*PahoBus, error) {

	bus := &PahoBus{
		host:          host,
		id:            id,
		subscriptions: make([]*subscription, 0),
		reconnectLock: sync.Mutex{},
	}

	return bus, bus.connect()
}

func (b *PahoBus) connect() error {

	log.Infof("Connecting paho bus")

	opts := mqtt.NewClientOptions().AddBroker("tcp://" + b.host).SetClientId(b.id).SetKeepAlive(10).SetCleanSession(true)
	opts = opts.SetOnConnectionLost(func(client *mqtt.MqttClient, reason error) {
		log.Warningf("Lost connection to server: %s", reason)
		b.disconnected()

		if b.destroyed {
			return
		}

		b.reconnectLock.Lock()
		go func() {
			time.Sleep(time.Second)
			b.reconnectLock.Unlock()
		}()

		b.connect()
	})
	b.mqtt = mqtt.NewClient(opts)

	var err = b.start()
	b.connected()
	return err
}

func (b *PahoBus) start() error {

	if _, err := b.mqtt.Start(); err != nil {
		return err
	}

	for _, sub := range b.subscriptions {
		err := b.subscribe(sub)

		return err
	}

	return nil
}

func (b *PahoBus) Destroy() {
	b.destroyed = true
	b.mqtt.Disconnect(0) // Wait 0ms to clean up
}

func (b *PahoBus) Publish(topic string, payload []byte) {
	b.mqtt.Publish(mqtt.QoS(0), topic, payload)
}

func (b *PahoBus) Subscribe(topic string, callback func(topic string, payload []byte)) (*subscription, error) {

	subscription := &subscription{
		topic:    topic,
		callback: callback,
	}

	b.subscriptions = append(b.subscriptions, subscription)

	err := b.subscribe(subscription)

	return subscription, err
}

func (b *PahoBus) subscribe(subscription *subscription) error {

	filter, err := mqtt.NewTopicFilter(subscription.topic, 0)
	if err != nil {
		log.FatalError(err, "Failed to subscribe to "+subscription.topic)
	}

	receipt, err := b.mqtt.StartSubscription(func(_ *mqtt.MqttClient, message mqtt.Message) {
		// XXX: ES: I've seen paho send me things I didn't ask for.
		if matches(subscription.topic, message.Topic()) {
			subscription.callback(message.Topic(), message.Payload())
		} else {
			log.Infof("FAIL! Asked for %s got %s", subscription.topic, message.Topic())
			spew.Dump(filter)
		}
	}, filter)

	if err != nil {
		return err
	}

	<-receipt

	return nil
}

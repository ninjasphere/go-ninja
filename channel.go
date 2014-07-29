package ninja

import (
	"fmt"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	nlog "github.com/ninjasphere/go-ninja/log"

	"github.com/bitly/go-simplejson"
	"github.com/bugsnag/bugsnag-go"
)

type JsonMessageHandler func(string, *simplejson.Json)

// ChannelBus context for channel related bus operations.
type ChannelBus struct {
	name     string
	protocol string
	device   *DeviceBus
	channel  <-chan MQTT.Receipt
	log      *nlog.Logger
}

// NewChannelBus Build a new channel bus for the supplied device
func NewChannelBus(name string, protocol string, d *DeviceBus) *ChannelBus {
	logger := nlog.GetLogger(fmt.Sprintf("channel.%s.%s", name, protocol))
	logger.Infof("Created")
	return &ChannelBus{
		name:     name,
		protocol: protocol,
		device:   d,
		log:      logger,
	}
}

// TODO move this into the actual drivers as we probably don't want their errors all in the one bucket.
func init() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey: "a39d43b795d60d16b1d6099236f5825e",
	})
}

// SendEvent Publish an event on the channel bus.
func (cb *ChannelBus) SendEvent(event string, payload *simplejson.Json) error {
	cb.log.Debugf("sending event %s", event)
	json, err := payload.MarshalJSON()
	if err != nil {
		return err
	}

	receipt := cb.device.driver.mqtt.Publish(MQTT.QoS(0), "$driver/"+cb.device.driver.id+"/device/"+cb.device.id+"/channel/"+cb.name+"/"+cb.protocol+"/event/"+event, json)
	<-receipt

	return nil
}

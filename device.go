package ninja

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/rpc"
	"github.com/ninjasphere/go-ninja/rpc/json2"

	"github.com/bitly/go-simplejson"
)

// DeviceBus Context for device related announcements.
type DeviceBus struct {
	id         string
	idType     string
	name       string
	driver     *DriverBus
	devicejson *simplejson.Json
	log        *logger.Logger
	rpcServer  *rpc.Server
}

// NewDeviceBus Create a new device bus.
func NewDeviceBus(id string, idType string, name string, driver *DriverBus, devicejson *simplejson.Json) *DeviceBus {
	log := logger.GetLogger(fmt.Sprintf("device.%s.%s", id, name))

	return &DeviceBus{
		id:         id,
		idType:     idType,
		name:       name,
		driver:     driver,
		devicejson: devicejson,
		log:        log,
		rpcServer:  rpc.NewServer(driver.mqtt, json2.NewCodec()),
	}
}

type eventingService interface {
	SetEventHandler(func(event string, payload interface{}) error)
}

func (d *DeviceBus) AddChannel(channel interface{}, name string, protocol string) error {
	return d.AddChannelWithSupported(channel, name, protocol, nil, nil)
}

// AddChannel Exports a channel as an RPC service, and announces it
// If the channel implements eventingService, it will be given a function to send events
func (d *DeviceBus) AddChannelWithSupported(channel interface{}, name string, protocol string, supportedMethods *[]string, supportedEvents *[]string) error {

	deviceguid, _ := d.devicejson.Get("guid").String()
	channelguid := GetGUID(name + protocol)

	topic := "$device/" + deviceguid + "/channel/" + channelguid + "/" + protocol

	exportedService, err := d.rpcServer.RegisterService(channel, topic)

	if err != nil {
		return fmt.Errorf("Failed to register channel service on %s : %s", topic, err)
	}

	if supportedMethods == nil {
		supportedMethods = &exportedService.Methods
	}

	if supportedEvents == nil {
		events := []string{}
		supportedEvents = &events
	}

	channelAnnouncement := &model.Channel{
		ID:       channelguid,
		Protocol: protocol,
		Name:     name,
		Supported: &model.ChannelSupported{
			Methods: supportedMethods,
			Events:  supportedEvents,
		},
		Device: &model.Device{},
	}

	js, err := d.devicejson.MarshalJSON()
	json.Unmarshal(js, &channelAnnouncement.Device)

	// send out channel announcement
	err = d.rpcServer.SendNotification(topic+"/announce", channelAnnouncement) // TODO: This should probably be exposed somewhere else

	if err != nil {
		return fmt.Errorf("Failed sending channel announcement: %s", err)
	}

	d.log.Debugf("Added channel: %s (protocol: %s) with methods: %s", name, protocol, strings.Join(exportedService.Methods, ", "))

	switch channel := channel.(type) {
	case eventingService:
		channel.SetEventHandler(func(event string, payload interface{}) error {
			return exportedService.SendEvent(event, payload)
		})
	}

	return nil
}

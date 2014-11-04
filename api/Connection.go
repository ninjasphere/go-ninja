package ninja

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/rpc"
	"github.com/ninjasphere/go-ninja/rpc/json2"
)

var (
	dummyRawCallback = func(params *json.RawMessage, values map[string]string) bool {
		return false
	}
)

// Connection Holds the connection to the Ninja MQTT bus, and provides all the methods needed to communicate with
// the other modules in Sphere.
type Connection struct {
	mqtt      *mqtt.MqttClient
	log       *logger.Logger
	rpc       *rpc.Client
	rpcServer *rpc.Server
}

// Connect Builds a new ninja connection to the MQTT broker, using the given client ID
func Connect(clientID string) (*Connection, error) {

	log := logger.GetLogger(fmt.Sprintf("%s.connection", clientID))

	conn := Connection{log: log}

	mqttURL := fmt.Sprintf("tcp://%s:%d", config.MustString("mqtt", "host"), config.MustInt("mqtt", "port"))

	log.Infof("Connecting to %s using cid:%s", mqttURL, clientID)

	opts := mqtt.NewClientOptions().AddBroker(mqttURL).SetClientId(clientID).SetCleanSession(true)
	conn.mqtt = mqtt.NewClient(opts)

	if _, err := conn.mqtt.Start(); err != nil {
		return nil, err
	}

	log.Infof("Connected")

	conn.rpc = rpc.NewClient(conn.mqtt, json2.NewClientCodec())
	conn.rpcServer = rpc.NewServer(conn.mqtt, json2.NewCodec())

	job, err := CreateStatusJob(&conn, clientID)

	if err != nil {
		return nil, err
	}
	job.Start()

	return &conn, nil
}

// GetMqttClient will be removed in a later version. All communication should happen via methods on Connection
func (c *Connection) GetMqttClient() *mqtt.MqttClient {
	return c.mqtt
}

type rpcMessage struct {
	Params *json.RawMessage `json:"params"`
}

// Subscribe allows you to subscribe to an MQTT topic. Topics can contain variables of the form ":myvar" which will
// be returned in the values map in the callback.
//
// The provided callback must be a function of 0, 1 or 2 parameters which returns
// "true" if it wants to receive more messages.
//
// The first parameter must either of type *json.RawMessage or else a pointer to a go struct type to which
// the expected event payload can be successfully unmarshalled.
//
// The second parameter should be of type map[string]string and will contain one value for each place holder
// specified in the topic string.
func (c *Connection) Subscribe(topic string, callback interface{}) error {

	adapter, err := getAdapter(c.log, callback)
	if err != nil {
		c.log.FatalError(err, fmt.Sprintf("Incompatible callback function provided as callback for topic %s", topic))
		return err
	}

	filter, err := mqtt.NewTopicFilter(GetSubscribeTopic(topic), 0)
	if err != nil {
		c.log.FatalError(err, "Failed to subscribe to "+topic)
	}

	finished := false
	mutex := &sync.Mutex{}

	receipt, err := c.mqtt.StartSubscription(func(_ *mqtt.MqttClient, message mqtt.Message) {
		// We lock so that the callback has a chance to return false,
		// to prevent any more messages arriving on this subscription
		mutex.Lock()

		// TODO: Implement unsubscribing. For now, it will just skip over any subscriptions that have finished
		if finished {
			return
		}

		values, ok := MatchTopicPattern(topic, message.Topic())
		if !ok {
			c.log.Warningf("Failed to read params from topic: %s using template: %s", message.Topic(), topic)
			mutex.Unlock()
		} else {

			msg := &rpcMessage{}
			err := json.Unmarshal(message.Payload(), msg)

			if err != nil {
				c.log.Warningf("Failed to read parameters in rpc call to %s - %v", message.Topic(), err)
				return
			}

			var params json.RawMessage

			json2.ReadRPCParams(msg.Params, &params)
			if err != nil {
				c.log.Warningf("Failed to read parameters in rpc call to %s - %v", message.Topic(), err)
				return
			}

			// The callback needs to be run in a goroutine as blocking this thread prevents any other messages arriving
			go func() {
				if !adapter(&params, *values) {
					// The callback has returned false, indicating that it does not want to receive any more messages.
					finished = true
				}
				mutex.Unlock()
			}()

		}
	}, filter)

	if err != nil {
		return err
	}

	<-receipt
	return nil
}

// GetServiceClient returns an RPC client for the given service.
func (c *Connection) GetServiceClient(serviceTopic string) *ServiceClient {
	return &ServiceClient{c, serviceTopic}
}

// ExportApp Exports an app using the 'app' protocol, and announces it
func (c *Connection) ExportApp(app App) error {

	if app.GetModuleInfo().ID == "" {
		panic("You must provide an ID in the package.json")
	}
	topic := fmt.Sprintf("$node/%s/app/%s", config.Serial(), app.GetModuleInfo().ID)

	announcement := app.GetModuleInfo()

	announcement.ServiceAnnouncement = model.ServiceAnnouncement{
		Schema: "http://schema.ninjablocks.com/service/app",
	}

	_, err := c.exportService(app, topic, announcement)

	if err != nil {
		return err
	}

	if config.Bool(false, "autostart") {
		err := c.GetServiceClient(topic).Call("start", struct{}{}, nil, time.Second*20)
		if err != nil {
			c.log.Fatalf("Failed to autostart app: %s", err)
		}
	}

	return nil
}

// ExportDriver Exports a driver using the 'driver' protocol, and announces it
func (c *Connection) ExportDriver(driver Driver) error {
	topic := fmt.Sprintf("$node/%s/driver/%s", config.Serial(), driver.GetModuleInfo().ID)

	announcement := driver.GetModuleInfo()

	announcement.ServiceAnnouncement = model.ServiceAnnouncement{
		Schema: "http://schema.ninjablocks.com/service/driver",
	}

	_, err := c.exportService(driver, topic, announcement)

	if err != nil {
		return err
	}

	if config.Bool(false, "autostart") {
		err := c.GetServiceClient(topic).Call("start", struct{}{}, nil, time.Second*20)
		if err != nil {
			c.log.Fatalf("Failed to autostart driver: %s", err)
		}
	}

	return nil
}

// ExportDevice Exports a device using the 'device' protocol, and announces it
func (c *Connection) ExportDevice(device Device) error {
	announcement := device.GetDeviceInfo()
	announcement.ID = getGUID(device.GetDeviceInfo().NaturalIDType, device.GetDeviceInfo().NaturalID)

	topic := fmt.Sprintf("$device/%s", announcement.ID)

	announcement.ServiceAnnouncement = model.ServiceAnnouncement{
		Schema: "http://schema.ninjablocks.com/service/device",
	}

	_, err := c.exportService(device, topic, announcement)

	if err != nil {
		return err
	}

	return nil
}

// ExportChannel Exports a device using the given protocol, and announces it
func (c *Connection) ExportChannel(device Device, channel Channel, id string) error {
	return c.ExportChannelWithSupported(device, channel, id, nil, nil)
}

// ExportChannelWithSupported is the same as ExportChannel, but any methods provided must actually be exported by the
// channel, or an error is returned
func (c *Connection) ExportChannelWithSupported(device Device, channel Channel, id string, supportedMethods *[]string, supportedEvents *[]string) error {
	if channel.GetProtocol() == "" {
		return fmt.Errorf("The channel must have a protocol. Channel ID: %s", id)
	}

	announcement := &model.Channel{
		ID:       id,
		Protocol: channel.GetProtocol(),
	}

	topic := fmt.Sprintf("$device/%s/channel/%s", device.GetDeviceInfo().ID, id)

	announcement.ServiceAnnouncement = model.ServiceAnnouncement{
		Schema:           c.resolveProtocolURI(channel.GetProtocol()),
		SupportedMethods: supportedMethods,
		SupportedEvents:  supportedEvents,
	}

	_, err := c.exportService(channel, topic, announcement)

	if err != nil {
		return err
	}

	// <TEMPORARY> - Expose channels using the old topic (with the protocol)
	/*properAnnouncement := announcement.ServiceAnnouncement

	shortProtocol := strings.TrimPrefix(c.resolveProtocolURI(channel.GetProtocol()), protocolSchemaURL.String())
	oldTopic := fmt.Sprintf("$device/%s/channel/%s/%s", device.GetDeviceInfo().ID, id, shortProtocol)

	deprecated := true
	announcement.ServiceAnnouncement = model.ServiceAnnouncement{
		Schema:           c.resolveProtocolURI(channel.GetProtocol()),
		SupportedMethods: supportedMethods,
		SupportedEvents:  supportedEvents,
		Deprecated:       &deprecated,
	}

	_, err = c.exportService(channel, oldTopic, announcement)

	announcement.ServiceAnnouncement = properAnnouncement

	if err != nil {
		return err
	}
	// </TEMPORARY>*/

	return nil
}

type simpleService struct {
	model.ServiceAnnouncement
}

func (s *simpleService) GetServiceAnnouncement() *model.ServiceAnnouncement {
	return &s.ServiceAnnouncement
}

// MustExportService Exports an RPC service, and announces it over TOPIC/event/announce. Must not cause an error or will panic.
func (c *Connection) MustExportService(service interface{}, topic string, announcement *model.ServiceAnnouncement) *rpc.ExportedService {
	exported, err := c.exportService(service, topic, &simpleService{*announcement})
	if err != nil {
		c.log.Fatalf("Failed to export service on topic '%s': %s", topic, err)
	}
	return exported
}

// ExportService Exports an RPC service, and announces it over TOPIC/event/announce
func (c *Connection) ExportService(service interface{}, topic string, announcement *model.ServiceAnnouncement) (*rpc.ExportedService, error) {
	return c.exportService(service, topic, &simpleService{*announcement})
}

type eventingService interface {
	SetEventHandler(func(event string, payload interface{}) error)
}

type serviceAnnouncement interface {
	GetServiceAnnouncement() *model.ServiceAnnouncement
}

// exportService Exports an RPC service, and announces it over TOPIC/event/announce
func (c *Connection) exportService(service interface{}, topic string, announcement serviceAnnouncement) (*rpc.ExportedService, error) {

	exportedService, err := c.rpcServer.RegisterService(service, topic, announcement.GetServiceAnnouncement().Schema)

	if err != nil {
		return nil, fmt.Errorf("Failed to register service on %s : %s", topic, err)
	}

	if announcement.GetServiceAnnouncement().SupportedMethods == nil {
		announcement.GetServiceAnnouncement().SupportedMethods = &exportedService.Methods
	} else {
		// TODO: Check that all strings in announcement.SupportedMethods exist in exportedService.Methods
		if len(*announcement.GetServiceAnnouncement().SupportedMethods) > len(exportedService.Methods) {
			return nil, fmt.Errorf("The number of actual exported methods is less than the number said to be exported. Check the method signatures of the service. topic:%s", topic)
		}
	}

	if announcement.GetServiceAnnouncement().SupportedEvents == nil {
		events := []string{}
		announcement.GetServiceAnnouncement().SupportedEvents = &events
	}

	announcement.GetServiceAnnouncement().Topic = topic

	// send out service announcement
	err = exportedService.SendEvent("announce", announcement)
	if err != nil {
		return nil, fmt.Errorf("Failed sending service announcement: %s", err)
	}

	c.log.Debugf("Exported service on topic: %s (schema: %s) with methods: %s", topic, announcement.GetServiceAnnouncement().Schema, strings.Join(*announcement.GetServiceAnnouncement().SupportedMethods, ", "))

	switch service := service.(type) {
	case eventingService:
		service.SetEventHandler(func(event string, payload interface{}) error {
			return exportedService.SendEvent(event, payload)
		})
	}

	return exportedService, nil
}

// SendNotification Sends a simple json-rpc notification to a topic
func (c *Connection) SendNotification(topic string, params ...interface{}) error {
	return c.rpcServer.SendNotification(topic, params...)
}

// Pull this out into the schema validation package when we have one
var rootSchemaURL, _ = url.Parse("http://schema.ninjablocks.com")
var protocolSchemaURL, _ = url.Parse("http://schema.ninjablocks.com/protocol/")

func (c *Connection) resolveSchemaURI(uri string) string {
	return c.resolveSchemaURIWithBase(rootSchemaURL, uri)
}

func (c *Connection) resolveProtocolURI(uri string) string {
	return c.resolveSchemaURIWithBase(protocolSchemaURL, uri)
}

func (c *Connection) resolveSchemaURIWithBase(base *url.URL, uri string) string {

	u, err := url.Parse(uri)
	if err != nil {
		c.log.Fatalf("Expected URL to parse: %q, got error: %v", uri, err)
	}
	return base.ResolveReference(u).String()
}

// support for reflective callbacks, modelled on approach used in rpc/server.go

type adapter struct {
	log      *logger.Logger
	function reflect.Value
	argCount int
	argType  reflect.Type
}

func (a *adapter) invoke(params *json.RawMessage, values map[string]string) bool {
	// self.log.Debugf("invoke: params=%s, values=%v", string(*params), values)
	var args []reflect.Value = make([]reflect.Value, a.argCount)

	switch a.argCount {
	case 2:
		args[1] = reflect.ValueOf(values)
		fallthrough
	case 1:
		arg := reflect.New(a.argType.Elem())
		err := json.Unmarshal(*params, arg.Interface())
		if err != nil {
			a.log.Errorf("failed to unmarshal %s as %v because %v", string(*params), arg, err)
			return true
		}
		args[0] = arg
	case 0:
	}
	return a.function.Call(args)[0].Interface().(bool)
}

func getAdapter(log *logger.Logger, callback interface{}) (func(params *json.RawMessage, values map[string]string) bool, error) {
	var err error = nil

	value := reflect.ValueOf(callback)
	valueType := value.Type()

	if valueType == reflect.ValueOf(dummyRawCallback).Type() {
		return callback.(func(params *json.RawMessage, values map[string]string) bool), nil
	}

	kind := value.Kind()
	if kind != reflect.Func {
		return nil, fmt.Errorf("%v is if kind %d, not of kind Func", callback, kind)
	}

	numIn := valueType.NumIn()

	var argType reflect.Type = nil
	empty := make(map[string]string)

	switch numIn {
	case 2:
		valuesType := valueType.In(1)
		if reflect.ValueOf(empty).Type() != valuesType {
			return nil, fmt.Errorf("second parameter, if specified must be of type map[string]string, is actually of type %v", valuesType)
		}
		fallthrough
	case 1:
		argType = valueType.In(0)
		argKind := argType.Kind()
		if argKind != reflect.Ptr {
			return nil, fmt.Errorf("type of first parameter %v must be of type Ptr, is actually of kind %d", argType, argKind)
		}
	case 0:
	default:
		return nil, fmt.Errorf("callback %v has too many (%d) parameters", callback, numIn)
	}

	numOut := valueType.NumOut()
	if numOut != 1 {
		return nil, fmt.Errorf("return type of %v has the wrong number (%d) of arguments", value, numOut)
	}

	if valueType.Out(0) != reflect.ValueOf(true).Type() {
		return nil, fmt.Errorf("return type of %v must be of type bool", value)
	}

	if err != nil {
		return nil, err
	}

	tmp := &adapter{
		log:      log,
		function: value,
		argCount: numIn,
		argType:  argType,
	}

	return tmp.invoke, err
}

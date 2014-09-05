package ninja

import (
	"fmt"
	"path"
	"time"

	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/rpc"
	"github.com/ninjasphere/go-ninja/rpc/json2"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
)

// NinjaConnection Connects to the local mqtt broker.
type NinjaConnection struct {
	mqtt      *MQTT.MqttClient
	log       *logger.Logger
	rpcClient *rpc.Client
}

// Connect Builds a new ninja connection which attaches to the local bus.
func Connect(clientID string) (*NinjaConnection, error) {

	log := logger.GetLogger("connection")

	conn := NinjaConnection{log: log}

	mqttURL, err := getMQTTUrl()
	if err != nil {
		return nil, err
	}

	opts := MQTT.NewClientOptions().AddBroker(mqttURL).SetClientId(clientID).SetCleanSession(true)
	conn.mqtt = MQTT.NewClient(opts)

	if _, err := conn.mqtt.Start(); err != nil {
		return nil, err
	}

	conn.rpcClient = rpc.NewClient(conn.mqtt, json2.NewClientCodec())

	log.Infof("Connected to %s\n", mqttURL)
	return &conn, nil
}

func (n *NinjaConnection) GetMqttClient() *MQTT.MqttClient {
	return n.mqtt
}

func (n *NinjaConnection) GetRPCClient() *rpc.Client {
	return n.rpcClient
}

// AnnounceDriver Anounce a driver has connected to the bus.
func (n *NinjaConnection) AnnounceDriver(id string, name string, driverPath string) (*DriverBus, error) {
	js, err := simplejson.NewJson([]byte(`{
    "params": [
    {
      "name": "",
      "file": "",
      "defaultConfig" : {},
      "package": {}
    }],
    "time":"",
    "jsonrpc":"2.0"
  }`))

	if err != nil {
		return nil, err
	}

	driverinfofile := path.Join(driverPath, "package.json")
	pkginfo, err := getDriverInfo(driverinfofile)
	if err != nil {
		return nil, err
	}
	filename, err := pkginfo.Get("main").String()
	if err != nil {
		return nil, err
	}

	mainfile := driverPath + filename
	js.Get("params").GetIndex(0).Set("file", mainfile)
	js.Get("params").GetIndex(0).Set("name", id)
	js.Get("params").GetIndex(0).Set("package", pkginfo)
	js.Get("params").GetIndex(0).Set("defaultConfig", "{}") //TODO fill me out
	js.Set("time", time.Now().Unix())
	json, _ := js.MarshalJSON()

	serial, err := GetSerial()
	if err != nil {
		return nil, err
	}
	version, err := pkginfo.Get("version").String()
	if err != nil {
		return nil, err
	}

	topic := fmt.Sprintf("$node/%s/app/%s/event/announce", serial, id)
	n.log.Infof("Sending message to topic: %s payload: %s", string(json))

	receipt := n.mqtt.Publish(MQTT.QoS(1), topic, json)
	<-receipt

	log := logger.GetLogger(fmt.Sprintf("driverbus.%s", id))

	driverBus := &DriverBus{
		id:      id,
		name:    name,
		mqtt:    n.mqtt,
		version: version,
		log:     log,
	}

	return driverBus, nil
}

// PublishRPCMessage publish an arbitrary message to the ninja bus and deal with the rpc wrapper!
func (n *NinjaConnection) PublishRPCMessage(topic string, params ...*simplejson.Json) error {
	json, err := BuildRPCRequest(params...)
	if err != nil {
		return err
	}
	receipt := n.mqtt.Publish(MQTT.QoS(1), topic, json)
	<-receipt
	return nil
}

// Alias so we can test this thing
var unixTimestampFunc = time.Now().Unix

// BuildRPCRequest Using the supplied params assemble a json RPC message and marshal it.
func BuildRPCRequest(params ...*simplejson.Json) ([]byte, error) {

	jsonmsg, err := simplejson.NewJson([]byte(`{"params": [],"time": "","jsonrpc": "2.0"}`))

	if err != nil {
		return nil, err
	}

	jsonmsg.Set("params", params)
	jsonmsg.Set("time", unixTimestampFunc())

	return jsonmsg.MarshalJSON()
}

// PublishMessage publish an arbitrary message to the ninja bus
func (n *NinjaConnection) PublishMessage(topic string, jsonmsg *simplejson.Json) error {
	json, err := jsonmsg.MarshalJSON()
	if err != nil {
		return err
	}
	receipt := n.mqtt.Publish(MQTT.QoS(1), topic, json)
	<-receipt
	return nil
}

func getMQTTUrl() (url string, err error) {

	var host string
	var port int

	cfg, err := GetConfig()
	if err != nil {
		return "", err
	}

	mqttConfig := cfg.Get("mqtt")
	if host, err = mqttConfig.Get("host").String(); err != nil {
		return "", err
	}

	if port, err = mqttConfig.Get("port").Int(); err != nil {
		return "", err
	}
	url = fmt.Sprintf("tcp://%s:%d", host, port)
	return url, nil
}

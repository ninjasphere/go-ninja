package rpc2

import (
	"encoding/json"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
)

type mqttJsonRpcConnection struct {
	log          *logger.Logger
	topic        string
	replyTopic   string
	mqttConn     *mqtt.MqttClient
	incomingData chan []byte
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     *json.RawMessage `json:"id"`
	Time   int64            `json:"time"`
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func NewMqttJsonRpcConnection(serving bool, mqttConn *mqtt.MqttClient, topic string, log *logger.Logger) (*mqttJsonRpcConnection, error) {

	c := &mqttJsonRpcConnection{
		mqttConn:     mqttConn,
		topic:        topic,
		replyTopic:   topic + "/reply",
		incomingData: make(chan []byte),
		log:          log,
	}

	if !serving {
		c.topic, c.replyTopic = c.replyTopic, c.topic
	}

	filter, err := mqtt.NewTopicFilter(c.topic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := c.mqttConn.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		c.incomingData <- message.Payload()
	}, filter)

	if err != nil {
		return nil, err
	}

	<-receipt

	return c, nil
}

func (c *mqttJsonRpcConnection) SendEvent(event string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Debugf("Sending event: %s payload: %s", event, jsonPayload)

	pubReceipt := c.mqttConn.Publish(mqtt.QoS(0), c.topic+"/event/"+event, jsonPayload)
	<-pubReceipt
	return nil
}

func (c *mqttJsonRpcConnection) Read(p []byte) (n int, err error) {

	var req = &serverRequest{}

	msg := <-c.incomingData

	err = json.Unmarshal(msg, req)
	if err != nil {
		log.Errorf("Failed to parse incoming json-rpc message %s : %s", err, msg)
		return 0, err
	}

	req.Method = "service." + upperFirst(req.Method)

	data, err := json.Marshal(req)

	if err != nil {
		log.Errorf("Failed to re-marshal incoming json-rpc message %s:", err)
		return 0, err
	}

	return copy(p[0:], data), nil
}

func (c *mqttJsonRpcConnection) Write(p []byte) (n int, err error) {
	pubReceipt := c.mqttConn.Publish(mqtt.QoS(0), c.replyTopic, p)
	<-pubReceipt
	return len(p), nil
}

func (c *mqttJsonRpcConnection) Close() error {
	log.Infof("mqttjson: Closing")
	return nil
}

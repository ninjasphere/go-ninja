package rpc2

import (
	"bufio"
	"encoding/json"
	"time"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-ninja/logger"
)

type mqttJsonRpcConnection struct {
	log            *logger.Logger
	incomingTopic  string
	outgoingTopic  string
	mqttConn       *mqtt.MqttClient
	bufferedReader *bufio.Reader
}

type serverRequest struct {
	Params   *json.RawMessage `json:"params,omitEmpty"`
	Method   *string          `json:"method,omitEmpty"`
	Time     *int64           `json:"time,omitEmpty"`
	Version  *string          `json:"jsonrpc,omitEmpty"`
	ID       *json.RawMessage `json:"id,omitEmpty"`
	Result   *json.RawMessage `json:"result,omitEmpty"`
	Response *json.RawMessage `json:"response,omitEmpty"`
	Error    *json.RawMessage `json:"error,omitEmpty"`
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

	fake := &fakeReader{
		incomingData: make(chan []byte),
	}

	c := &mqttJsonRpcConnection{
		mqttConn:       mqttConn,
		incomingTopic:  topic,
		outgoingTopic:  topic + "/reply",
		log:            log,
		bufferedReader: bufio.NewReaderSize(fake, 999999999), // TODO: Fix this
	}

	if !serving {
		c.incomingTopic, c.outgoingTopic = c.outgoingTopic, c.incomingTopic
	}

	filter, err := mqtt.NewTopicFilter(c.incomingTopic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := c.mqttConn.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		fake.incomingData <- message.Payload()
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

	pubReceipt := c.mqttConn.Publish(mqtt.QoS(0), c.incomingTopic+"/event/"+event, jsonPayload)
	<-pubReceipt
	return nil
}

func (c *mqttJsonRpcConnection) Read(p []byte) (n int, err error) {
	return c.bufferedReader.Read(p)
}

type fakeReader struct {
	incomingData chan []byte
}

func (c *fakeReader) Read(p []byte) (n int, err error) {
	msg := <-c.incomingData

	log.Infof("Incoming %s", msg)

	req, err := simplejson.NewJson(msg)

	if err != nil {
		log.Errorf("Failed to parse incoming json-rpc message %s : %s", err, msg)
		return 0, err
	}

	method := req.Get("method").MustString()
	if method != "" {
		req.Set("method", "service."+upperFirst(method))
	}

	req.Set("result", req.Get("response"))

	data, err := req.MarshalJSON()

	if err != nil {
		log.Errorf("Failed to re-marshal incoming json-rpc message %s:", err)
		return 0, err
	}

	return copy(p[0:], data), nil
}

const blank = "[]"
const version = "2.0"

func (c *mqttJsonRpcConnection) Write(p []byte) (n int, err error) {

	var version = "2.0"
	var now = time.Now().Unix() / 10000
	var req = &serverRequest{
		Version: &version,
		Time:    &now,
	}

	err = json.Unmarshal(p, req)
	if err != nil {
		return 0, err
	}

	spew.Dump(req)

	if req.Params != nil && string(*req.Params) == "[null]" {
		blank := json.RawMessage([]byte("[]"))
		req.Params = &blank
	}

	spew.Dump(req)

	payload, err := json.Marshal(req)

	pubReceipt := c.mqttConn.Publish(mqtt.QoS(0), c.outgoingTopic, payload)
	<-pubReceipt
	return len(p), nil
}

func (c *mqttJsonRpcConnection) Close() error {
	log.Infof("mqttjson: Closing")
	return nil
}

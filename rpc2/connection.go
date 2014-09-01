package rpc2

import (
	"bufio"
	"encoding/json"
	"time"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/bitly/go-simplejson"
	"github.com/ninjasphere/go-ninja/logger"
)

type mqttJsonRpcConnection struct {
	log            *logger.Logger
	incomingTopic  string
	outgoingTopic  string
	mqttConn       *mqtt.MqttClient
	bufferedReader *bufio.Reader
}

type rpcRequestResponse struct {
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
		bufferedReader: bufio.NewReaderSize(fake, 512*1024), // TODO: Fix this
	}

	if !serving {
		c.incomingTopic, c.outgoingTopic = c.outgoingTopic, c.incomingTopic
	}

	fake.incomingTopic = c.incomingTopic

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
	incomingData  chan []byte
	incomingTopic string
}

var result = json.RawMessage([]byte("true"))

func (c *fakeReader) Read(p []byte) (n int, err error) {
	msg := <-c.incomingData

	//log.Infof("< Incoming (%s) (unaltered) %s", c.incomingTopic, msg)

	r := &rpcRequestResponse{}

	err = json.Unmarshal(msg, r)

	if err != nil {
		log.Errorf("Failed to parse incoming json-rpc message %s : %s", err, msg)
		return 0, err
	}

	if r.Method != nil {
		method := "service." + upperFirst(*r.Method)
		r.Method = &method
	}

	r.Result = r.Response
	r.Response = nil
	r.Time = nil
	r.Version = nil

	if r.Result == nil && r.Error == nil {
		r.Result = &result
	}

	data, err := json.Marshal(r)

	//log.Infof("< Incoming (%s) (altered)   %s", c.incomingTopic, data)

	if err != nil {
		log.Errorf("Failed to re-marshal incoming json-rpc message %s:", err)
		return 0, err
	}

	return copy(p[0:], data), nil
}

const blank = "[]"
const version = "2.0"

func (c *mqttJsonRpcConnection) Write(p []byte) (n int, err error) {

	//log.Infof("< Outgoing (%s) (unaltered) %s", c.outgoingTopic, p)

	var version = "2.0"
	var now = time.Now().Unix() / 10000

	req, err := simplejson.NewJson(p)

	if err != nil {
		return 0, err
	}

	req.Set("jsonrpc", version)
	req.Set("time", now)
	req.Set("response", req.Get("result"))
	req.Del("result")
	/*error := req.Get("error").MustString()
	log.Infof("outgoing error %s", error)
	if error == "" {
		req.Del("error")
	}*/
	/*
		if req.Params != nil && string(*req.Params) == "[null]" {
			blank := json.RawMessage([]byte("[]"))
			req.Params = &blank
		}*/

	payload, err := req.MarshalJSON()

	//log.Infof("< Outgoing (%s) (altered)   %s", c.outgoingTopic, payload)

	_ /*pubReceipt :*/ = c.mqttConn.Publish(mqtt.QoS(0), c.outgoingTopic, payload)
	//<-pubReceipt
	return len(p), nil
}

func (c *mqttJsonRpcConnection) Close() error {
	log.Infof("mqttjson: Closing")
	return nil
}

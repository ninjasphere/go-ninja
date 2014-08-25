package rpc2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type mqttJsonRpcConnection struct {
	replyTopic string
	mqttConn   *mqtt.MqttClient
	incoming   chan []byte
}

func ExportService(service interface{}, topic string, mqttConn *mqtt.MqttClient) (*rpc.Server, error) {

	srv := rpc.NewServer()
	if err := srv.RegisterName("service", service); err != nil {
		return nil, fmt.Errorf("Couldn't register: %s", err)
	}

	conn := &mqttJsonRpcConnection{
		mqttConn:   mqttConn,
		replyTopic: topic + "/reply",
		incoming:   make(chan []byte),
	}

	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := mqttConn.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		conn.incoming <- message.Payload()
	}, filter)

	<-receipt

	go srv.ServeCodec(jsonrpc.NewServerCodec(conn))

	return srv, nil
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     *json.RawMessage `json:"id"`
	Time   int              `json:"time"`
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func (c *mqttJsonRpcConnection) Read(p []byte) (n int, err error) {

	var req = &serverRequest{}

	err = json.Unmarshal(<-c.incoming, req)
	if err != nil {
		log.Printf("Failed to parse incoming json-rpc message %s:", err)
		return 0, err
	}

	req.Method = "service." + upperFirst(req.Method)

	data, err := json.Marshal(req)

	if err != nil {
		log.Printf("Failed to re-marshal incoming json-rpc message %s:", err)
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
	log.Printf("mqttjson: Closing")
	return nil
}

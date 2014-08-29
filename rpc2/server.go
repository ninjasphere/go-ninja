package rpc2

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/juju/loggo"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("rpc2")

type mqttJsonRpcConnection struct {
	log      *logger.Logger
	topic    string
	mqttConn *mqtt.MqttClient
	incoming chan []byte
}

type eventingService interface {
	SetEventHandler(func(event string, payload interface{}) error)
}

func ExportService(service interface{}, topic string, mqttConn *mqtt.MqttClient) ([]string, error) {

	log := logger.GetLogger(fmt.Sprintf("RPC - %T - %s", service, topic))

	log.Infof("Starting RPC service")

	srv := rpc.NewServer()
	if err := srv.RegisterName("service", service); err != nil {
		return nil, fmt.Errorf("Couldn't register: %s", err)
	}

	conn := &mqttJsonRpcConnection{
		mqttConn: mqttConn,
		topic:    topic,
		incoming: make(chan []byte),
		log:      log,
	}

	conn.log.SetLogLevel(loggo.TRACE)

	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := mqttConn.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		conn.incoming <- message.Payload()
	}, filter)

	<-receipt

	go srv.ServeCodec(jsonrpc.NewServerCodec(conn))

	switch service := service.(type) {
	case eventingService:

		service.SetEventHandler(func(event string, payload interface{}) error {
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			log.Debugf("Sending event: %s payload: %s", event, jsonPayload)

			pubReceipt := mqttConn.Publish(mqtt.QoS(0), topic+"/event/"+event, jsonPayload)
			<-pubReceipt
			return nil
		})

	}

	return suitableMethods(reflect.TypeOf(service)), nil
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

func (c *mqttJsonRpcConnection) Read(p []byte) (n int, err error) {

	var req = &serverRequest{}

	msg := <-c.incoming

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
	pubReceipt := c.mqttConn.Publish(mqtt.QoS(0), c.topic+"/reply", p)
	<-pubReceipt
	return len(p), nil
}

func (c *mqttJsonRpcConnection) Close() error {
	log.Infof("mqttjson: Closing")
	return nil
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
// Adapted from http://golang.org/src/pkg/net/rpc/server.go
func suitableMethods(typ reflect.Type) []string {
	methods := make([]string, 0)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		methods = append(methods, mname)

	}
	return methods
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

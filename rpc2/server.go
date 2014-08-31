package rpc2

import (
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

	conn, err := NewMqttJsonRpcConnection(true, mqttConn, topic, log)

	if err != nil {
		return nil, err
	}

	conn.log.SetLogLevel(loggo.TRACE)

	go srv.ServeCodec(jsonrpc.NewServerCodec(conn))

	switch service := service.(type) {
	case eventingService:

		service.SetEventHandler(func(event string, payload interface{}) error {
			return conn.SendEvent(event, payload)
		})

	}

	return suitableMethods(reflect.TypeOf(service)), nil
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
// Adapted from http://golang.org/src/pkg/net/rpc/server.go
func suitableMethods(typ reflect.Type) []string {
	var methods []string

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

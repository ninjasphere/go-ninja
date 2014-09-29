// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/schemas"
)

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(topic string, message mqtt.Message) (CodecRequest, error)
	SendNotification(c *mqtt.MqttClient, topic string, payload ...interface{}) error
}

// CodecRequest decodes a request and encodes a response using a specific
// serialization scheme.
type CodecRequest interface {
	// Reads the request and returns the RPC method name.
	Method() (string, error)
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(c *mqtt.MqttClient, response interface{})
	// Writes an error produced by the server.
	WriteError(c *mqtt.MqttClient, err error)
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// NewServer returns a new RPC server.
func NewServer(client *mqtt.MqttClient, codec Codec) *Server {
	return &Server{
		client:   client,
		codec:    codec,
		services: new(serviceMap),
	}
}

// Server serves registered RPC services using registered codecs.
type Server struct {
	client   *mqtt.MqttClient
	codec    Codec
	services *serviceMap
}

type ExportedService struct {
	Methods []string
	topic   string
	server  *Server
	schema  string
}

func (s *ExportedService) SendEvent(event string, payload interface{}) error {

	schema := s.schema + "#/events/" + event + "/value"
	message, err := schemas.Validate(schema, payload)

	if err != nil {
		return err
	}

	if message != nil && event == "announce" {
		// Assume this is a channel announcement (which doesn't actually define the announce event in every protocol)

		schema = "http://schema.ninjablocks.com/model/channel#"
		message, err = schemas.Validate(schema, payload)

		if err != nil {
			return err
		}
	}

	if message != nil {
		return fmt.Errorf("Event '%s' failed validation (schema: %s) message: %s", event, schema, *message)
	}

	return s.server.SendNotification(s.topic+"/event/"+event, payload)
}

// RegisterService adds a new service to the server.
//
// The name parameter is optional: if empty it will be inferred from
// the receiver type name.
//
// Methods from the receiver will be extracted if these rules are satisfied:
//
//    - The receiver is exported (begins with an upper case letter) or local
//      (defined in the package registering the service).
//    - The method name is exported.
//    - The method's first argument is *mqtt.Message
//    - If there is a second argument (the RPC params value) it must be exported and a pointer
//    - If there is a return value, it must be first, exported and a pointer
//    - The method's last return value is an error
//
// All other methods are ignored.
func (s *Server) RegisterService(receiver interface{}, topic string, schema string) (service *ExportedService, err error) {

	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := s.client.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		go s.serveRequest(topic, message)
	}, filter)

	if err != nil {
		return nil, err
	}

	<-receipt

	exportedMethods, err := s.services.register(receiver, topic)

	var exportedMethodsLower []string

	for _, m := range exportedMethods {
		exportedMethodsLower = append(exportedMethodsLower, lowerFirst(m))
	}

	return &ExportedService{Methods: exportedMethodsLower, topic: topic, server: s, schema: schema}, err
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// SendNotification sends a one-way notification. Perhaps shouldn't be in Server....
func (s *Server) SendNotification(topic string, params ...interface{}) error {
	return s.codec.SendNotification(s.client, topic, params...)
}

// HasMethod returns true if the given method is registered on a topic
func (s *Server) HasMethod(topic string, method string) bool {
	if _, _, err := s.services.get(topic, method); err == nil {
		return true
	}
	return false
}

type Message struct {
	Payload []byte
	Topic   string
}

// ServeRequest handles an incoming Json-RPC MQTT message
func (s *Server) serveRequest(topic string, message mqtt.Message) {

	log.Debugf("Serving request to %s", topic)

	// Create a new codec request.
	codecReq, err := s.codec.NewRequest(topic, message)

	if err != nil {
		codecReq.WriteError(s.client, err)
		return
	}

	// Get service method to be called.
	method, errMethod := codecReq.Method()
	if errMethod != nil {
		codecReq.WriteError(s.client, errMethod)
		return
	}

	serviceSpec, methodSpec, errGet := s.services.get(topic, method)
	if errGet != nil {
		codecReq.WriteError(s.client, errGet)
		return
	}
	// Decode the args.
	args := reflect.New(methodSpec.argsType)
	if errRead := codecReq.ReadRequest(args.Interface()); errRead != nil {
		codecReq.WriteError(s.client, errRead)
		return
	}
	// Call the service method.

	params := []reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(&Message{
			Payload: message.Payload(),
			Topic:   topic,
		}),
	}

	if methodSpec.argsType != nil {
		params = append(params, args)
	}

	/*var reply reflect.Value
	if methodSpec.replyType != nil {
		reply = reflect.New(methodSpec.replyType)
		params = append(params, reply)
	}*/

	retVals := methodSpec.method.Func.Call(params)
	// Cast the last result to error if needed.
	var errResult error
	errInter := retVals[len(retVals)-1].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	var reply reflect.Value
	if methodSpec.replyType != nil {
		reply = retVals[0]
	}

	// Encode the response.
	if errResult == nil {
		if methodSpec.replyType != nil {
			codecReq.WriteResponse(s.client, reply.Interface())
		} else {
			codecReq.WriteResponse(s.client, nil)
		}
	} else {
		codecReq.WriteError(s.client, errResult)
	}
}

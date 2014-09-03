// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"reflect"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(topic string, message mqtt.Message) CodecRequest
	SendNotification(c *mqtt.MqttClient, topic string, payload interface{}) error
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

type exportedService struct {
	Methods []string
	topic   string
	server  *Server
}

func (s *exportedService) SendEvent(event string, payload interface{}) error {
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
//    - The method has three arguments: *mqtt.Message, *args, *reply.
//    - All three arguments are pointers.
//    - The second and third arguments are exported or local.
//    - The method has return type error.
//
// All other methods are ignored.
func (s *Server) RegisterService(receiver interface{}, topic string) (service *exportedService, err error) {

	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return nil, err
	}

	receipt, err := s.client.StartSubscription(func(client *mqtt.MqttClient, message mqtt.Message) {
		s.serveRequest(topic, message)
	}, filter)

	if err != nil {
		return nil, err
	}

	<-receipt

	exportedMethods, err := s.services.register(receiver, topic)

	return &exportedService{Methods: exportedMethods, topic: topic, server: s}, err
}

// SendNotification sends a one-way notification. Perhaps shouldn't be in Server....
func (s *Server) SendNotification(topic string, payload interface{}) error {
	return s.codec.SendNotification(s.client, topic, payload)
}

// HasMethod returns true if the given method is registered on a topic
func (s *Server) HasMethod(topic string, method string) bool {
	if _, _, err := s.services.get(topic, method); err == nil {
		return true
	}
	return false
}

// ServeRequest handles an incoming Json-RPC MQTT message
func (s *Server) serveRequest(topic string, message mqtt.Message) {

	// Create a new codec request.
	codecReq := s.codec.NewRequest(topic, message)
	// Get service method to be called.
	method, errMethod := codecReq.Method()
	if errMethod != nil {
		codecReq.WriteError(s.client, errMethod)
		return
	}

	log.Infof("TOPIC '%s' METHOD '%s'", topic, method)

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
	reply := reflect.New(methodSpec.replyType)
	errValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(message),
		args,
		reply,
	})
	// Cast the result to error if needed.
	var errResult error
	errInter := errValue[0].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}

	// Encode the response.
	if errResult == nil {
		codecReq.WriteResponse(s.client, reply.Interface())
	} else {
		codecReq.WriteError(s.client, errResult)
	}
}

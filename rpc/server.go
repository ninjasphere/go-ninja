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

	"github.com/ninjasphere/go-ninja/bus"
	"github.com/ninjasphere/go-ninja/schemas"
)

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(topic string, payload []byte) (CodecRequest, error)
	SendNotification(c bus.Bus, topic string, payload ...interface{}) error
}

// CodecRequest decodes a request and encodes a response using a specific
// serialization scheme.
type CodecRequest interface {
	// Reads the request and returns the RPC method name.
	Method() (string, error)
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(c bus.Bus, response interface{})
	// Writes an error produced by the server.
	WriteError(c bus.Bus, err error)
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// NewServer returns a new RPC server.
func NewServer(client bus.Bus, codec Codec) *Server {
	return &Server{
		client:   client,
		codec:    codec,
		services: new(serviceMap),
	}
}

// Server serves registered RPC services using registered codecs.
type Server struct {
	client   bus.Bus
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

	// We ignore announce events, as we don't define them in all the protocols/services
	if event != "announce" {
		schema := s.schema + "#/events/" + event + "/value"
		message, err := schemas.Validate(schema, payload)

		if message != nil {
			return fmt.Errorf("Event '%s' failed validation (schema: %s) message: %s", event, schema, *message)
		}

		if err != nil {
			log.Warningf("Failed to validate event %s on service %s. Error:%s", event, s.schema, err)
		}
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

	_, err = s.client.Subscribe(topic, func(topic string, payload []byte) {
		s.serveRequest(topic, payload)
	})

	if err != nil {
		return nil, err
	}

	methods, err := schemas.GetServiceMethods(schema)
	if err != nil {
		return nil, err
	}

	exportedMethods, err := s.services.register(receiver, topic, methods)

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
func (s *Server) serveRequest(topic string, payload []byte) {

	log.Debugf("Serving request to %s", topic)

	// Create a new codec request.
	codecReq, err := s.codec.NewRequest(topic, payload)

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
	var args reflect.Value
	if methodSpec.argsType != nil {
		if methodSpec.argsType.Kind() == reflect.Ptr {
			args = reflect.New(methodSpec.argsType.Elem())
		} else {
			args = reflect.New(methodSpec.argsType)
		}

		if errRead := codecReq.ReadRequest(args.Interface()); errRead != nil {
			codecReq.WriteError(s.client, errRead)
			return
		}

	}
	// Call the service method.

	params := []reflect.Value{
		serviceSpec.rcvr,
	}

	/*
		TODO: Allow the method to request an rpc.Message
		reflect.ValueOf(&Message{
			Payload: message.Payload(),
			Topic:   topic,
		}),*/

	if methodSpec.argsType != nil {
		if methodSpec.argsType.Kind() == reflect.Ptr {
			params = append(params, args)
		} else {
			params = append(params, args.Elem())
		}
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

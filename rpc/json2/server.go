// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json2

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/rpc"
)

var null = json.RawMessage([]byte("null"))
var Version = "2.0"

var log = logger.GetLogger("mqtt-jsonrpc2")

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a JSON-RPC request received by the server.
type serverRequest struct {

	// A String containing the name of the method to be invoked.
	Method *string `json:"method,omitempty"`

	// A Structured value to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`

	// The request id. MUST be a string, number or null.
	// Our implementation will not do type checking for id.
	// It will be copied as it is.
	ID *json.RawMessage `json:"id,omitempty"`

	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	Time int64 `json:"time"`
}

// serverResponse represents a JSON-RPC response returned by the server.
type serverResponse struct {

	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	// As per spec the member will be omitted if there was an error.
	Result interface{} `json:"result,omitempty"`

	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	// As per spec the member will be omitted if there was no error.
	Error *Error `json:"error,omitempty"`

	// This must be the same id as the request it is responding to.
	ID *json.RawMessage `json:"id"`

	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	Time int64 `json:"time"`
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// NewCodec returns a new JSON Codec.
func NewCodec() *Codec {
	return &Codec{}
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
}

// NewRequest returns a CodecRequest.
func (c *Codec) NewRequest(topic string, msg mqtt.Message) (rpc.CodecRequest, error) {
	return newCodecRequest(topic, msg)
}

// SendNotification sends a JSON-RPC notification
func (c *Codec) SendNotification(client *mqtt.MqttClient, topic string, payload ...interface{}) error {

	notification := &serverRequest{
		Version: Version,
		Time:    makeTimestamp(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshall rpc notification: %s", err)
	}

	rawPayload := json.RawMessage(jsonPayload)

	notification.Params = &rawPayload
	jsonNotification, err := json.Marshal(notification)

	if !strings.HasSuffix(topic, "/module/status") {
		log.Debugf("< Outgoing to %s : %s", topic, jsonNotification)
	}

	client.Publish(mqtt.QoS(0), topic, jsonNotification)

	if err != nil {
		return fmt.Errorf("Failed to write rpc notification to MQTT: %s", err)
	}

	return nil
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

// newCodecRequest returns a new CodecRequest.
func newCodecRequest(topic string, msg mqtt.Message) (rpc.CodecRequest, error) {

	log.Debugf("> Incoming to %s : %s", topic, msg.Payload())

	// Decode the request body and check if RPC method is valid.
	req := new(serverRequest)
	err := json.Unmarshal(msg.Payload(), req)
	if err != nil {
		err = &Error{
			Code:    E_PARSE,
			Message: err.Error(),
			Data:    req,
		}
		log.Infof("Bad incoming json-rpc request to %s error:%s json:%s ", topic, err, msg.Payload())
	} else {
		method := upperFirst(*req.Method)

		req.Method = &method

		if req.Version != Version {
			err = &Error{
				Code:    E_INVALID_REQ,
				Message: "jsonrpc must be " + Version,
				Data:    req,
			}
		}
	}
	return &CodecRequest{request: req, err: err, topic: topic}, err
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	request *serverRequest
	err     error
	topic   string
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return *c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(args interface{}) error {
	if c.err == nil {
		if c.request.Params != nil {

			c.err = ReadRPCParams(c.request.Params, args)

		} else {
			// Ninja allows a null params field. Should work out how to check missing vs. null.
			// Then again. Fuck it. Yolo.
			/*c.err = &Error{
				Code:    E_INVALID_REQ,
				Message: "rpc: method request ill-formed: missing params field",
			}*/
		}
	}
	return c.err
}

func ReadRPCParams(params *json.RawMessage, args interface{}) error {

	var err error

	// Ninja: If we get an array in, try to pass its contents as one argument
	paramsString := string(*params)
	if paramsString == "[]" {
		params = nil
	} else if strings.HasPrefix(paramsString, "[") {
		rawParams := &json.RawMessage{}
		err := json.Unmarshal([]byte(paramsString[1:len(paramsString)-1]), rawParams)

		if err != nil {
			err = &Error{
				Code:    E_INVALID_REQ,
				Message: "Ninja's golang rpc only accepts one param in an array. Use named params instead.",
				Data:    params,
			}
		} else {
			params = rawParams
		}

	}

	if err == nil && params != nil {
		// JSON params structured object. Unmarshal to the args object.
		err = json.Unmarshal(*params, args)
		if err != nil {
			err = &Error{
				Code:    E_INVALID_REQ,
				Message: err.Error(),
				Data:    params,
			}
		}
	}

	return err
}

// WriteResponse encodes the response and writes it to the reply topic
func (c *CodecRequest) WriteResponse(client *mqtt.MqttClient, reply interface{}) {
	if reply == nil {
		reply = json.RawMessage("null")
	}
	res := &serverResponse{
		Version: Version,
		Result:  reply,
		ID:      c.request.ID,
		Time:    makeTimestamp(),
	}
	c.writeServerResponse(client, res)
}

func (c *CodecRequest) WriteError(client *mqtt.MqttClient, err error) {
	jsonErr, ok := err.(*Error)
	if !ok {
		jsonErr = &Error{
			Code:    E_SERVER,
			Message: err.Error(),
		}
	}
	res := &serverResponse{
		Version: Version,
		Error:   jsonErr,
		ID:      c.request.ID,
		Time:    makeTimestamp(),
	}
	c.writeServerResponse(client, res)
}

func (c *CodecRequest) writeServerResponse(client *mqtt.MqttClient, res *serverResponse) {
	// Id is null for notifications and they don't have a response.

	if c.request.ID != nil {

		payload, err := json.Marshal(res)

		log.Debugf("< Outgoing to %s : %s", c.topic+"/reply", payload)

		if err != nil {
			log.Errorf("Failed to marshall rpc response: %s", err)
			return
		}

		client.Publish(mqtt.QoS(0), c.topic+"/reply", payload)

		if err != nil {
			log.Errorf("Failed to write rpc response to MQTT: %s", err)
			return
		}
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type EmptyResponse struct {
}

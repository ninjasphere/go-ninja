// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Note: This isn't actually json-rpc2. It's got the wrong reply property (response vs. result)
// It's "Ninja RPC". Until we fix it. If we fix it.

package json2

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/rpc3"
)

var null = json.RawMessage([]byte("null"))
var Version = "2.0"

var log = logger.GetLogger("mqtt-jsonrpc2")

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// serverRequest represents a JSON-RPC request received by the server.
type serverRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`

	// The request id. MUST be a string, number or null.
	// Our implementation will not do type checking for id.
	// It will be copied as it is.
	Id *json.RawMessage `json:"id"`
}

// serverResponse represents a JSON-RPC response returned by the server.
type serverResponse struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	// As per spec the member will be omitted if there was an error.
	Result interface{} `json:"response,omitempty"` // FIXME: XXX: TODO: This json property name should be 'result'

	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	// As per spec the member will be omitted if there was no error.
	Error *Error `json:"error,omitempty"`

	// This must be the same id as the request it is responding to.
	Id *json.RawMessage `json:"id"`
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
func (c *Codec) NewRequest(topic string, msg mqtt.Message) rpc.CodecRequest {
	return newCodecRequest(topic, msg)
}

// SendNotification sends a JSON-RPC notification
func (c *Codec) SendNotification(client *mqtt.MqttClient, topic string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("Failed to marshall rpc notification: %s", err)
	}

	log.Debugf("< Outgoing to %s : %s", topic, jsonPayload)

	client.Publish(mqtt.QoS(0), topic, jsonPayload)

	if err != nil {
		return fmt.Errorf("Failed to write rpc notification to MQTT: %s", err)
	}

	return nil
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

// newCodecRequest returns a new CodecRequest.
func newCodecRequest(topic string, msg mqtt.Message) rpc.CodecRequest {

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
	}

	req.Method = upperFirst(req.Method)

	if req.Version != Version {
		err = &Error{
			Code:    E_INVALID_REQ,
			Message: "jsonrpc must be " + Version,
			Data:    req,
		}
	}
	return &CodecRequest{request: req, err: err, topic: topic}
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
		return c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
func (c *CodecRequest) ReadRequest(args interface{}) error {
	if c.err == nil {
		if c.request.Params != nil {

			// Ninja: If we get an array in, try to pass its contents as one argument
			paramsString := string(*c.request.Params)
			if strings.HasPrefix(paramsString, "[") {
				rawParams := &json.RawMessage{}
				err := json.Unmarshal([]byte(paramsString[1:len(paramsString)-1]), rawParams)

				if err != nil {
					c.err = &Error{
						Code:    E_INVALID_REQ,
						Message: "Ninja's golang rpc only accepts one param in an array. Use named params instead.",
						Data:    c.request.Params,
					}
				} else {
					c.request.Params = rawParams
				}

			}

			if c.err == nil {
				// JSON params structured object. Unmarshal to the args object.
				err := json.Unmarshal(*c.request.Params, args)
				if err != nil {
					c.err = &Error{
						Code:    E_INVALID_REQ,
						Message: err.Error(),
						Data:    c.request.Params,
					}
				}
			}
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

// WriteResponse encodes the response and writes it to the reply topic
func (c *CodecRequest) WriteResponse(client *mqtt.MqttClient, reply interface{}) {
	res := &serverResponse{
		Version: Version,
		Result:  reply,
		Id:      c.request.Id,
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
		Id:      c.request.Id,
	}
	c.writeServerResponse(client, res)
}

func (c *CodecRequest) writeServerResponse(client *mqtt.MqttClient, res *serverResponse) {
	// Id is null for notifications and they don't have a response.

	if c.request.Id != nil {

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

type EmptyResponse struct {
}

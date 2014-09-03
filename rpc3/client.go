// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"math/rand"
	"sync"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("rpc")

type ClientCodec interface {
	EncodeClientRequest(call *Call) ([]byte, error)
	DecodeIdAndError(msg []byte) (*uint32, error)
	DecodeClientResponse(msg []byte, reply interface{}) error
}

// Call represents an active RPC.
type Call struct {
	Topic         string      // The MQTT topic this call will be sent to
	ServiceMethod string      // The name of the service and method to call.
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call  // Strobes when call is complete.
	Id            uint32      // Used to map responses
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	mutex   sync.Mutex // protects following
	codec   ClientCodec
	mqtt    *mqtt.MqttClient
	pending map[uint32]*Call
}

// NewClient creates a new rpc client using the provided MQTT connection
func NewClient(mqtt *mqtt.MqttClient, codec ClientCodec) *Client {
	client := &Client{
		pending: make(map[uint32]*Call),
		mqtt:    mqtt,
		codec:   codec,
	}
	return client
}

func (client *Client) send(call *Call) (*Call, error) {

	// Register this call, if we are expecting a reply
	if call.Done != nil {
		call.Id = rand.Uint32()
	}

	payload, err := client.codec.EncodeClientRequest(call)
	if err != nil {
		return nil, err
	}

	log.Debugf("< Outgoing to %s : %s", call.Topic, payload)

	pubReceipt := client.mqtt.Publish(mqtt.QoS(0), call.Topic, payload)

	<-pubReceipt

	if call.Done != nil {
		client.pending[call.Id] = call

		filter, err := mqtt.NewTopicFilter(call.Topic+"/reply", 0)
		if err != nil {
			return nil, err
		}

		receipt, err := client.mqtt.StartSubscription(func(mqtt *mqtt.MqttClient, message mqtt.Message) {
			log.Debugf("< Incoming to %s : %s", call.Topic, message.Payload())
			client.handleResponse(message)
		}, filter)

		if err != nil {
			return nil, err
		}

		<-receipt
	}

	return call, nil
}

func (client *Client) handleResponse(message mqtt.Message) {
	id, err := client.codec.DecodeIdAndError(message.Payload())

	if id == nil {
		log.Infof("Failed to decode reply: %s error: %s", message.Payload(), err)
		return
	}

	client.mutex.Lock()
	call := client.pending[*id]
	client.pending[*id] = nil
	client.mutex.Unlock()

	if err != nil {
		if call != nil {
			call.Error = err
			call.done()
		} else {
			log.Debugf("Ignoring error reply to call %d: %s", *id, err)
		}
		return
	}

	if call == nil {
		log.Debugf("Ignoring reply to call %d", *id)
		return
	}

	call.Error = client.codec.DecodeClientResponse(message.Payload(), call.Reply)
	call.done()
}

func (call *Call) done() {

	select {
	case call.Done <- call:
		// ok
	default:
		log.Infof("Discarding Call reply due to insufficient Done chan capacity")
	}

}

// Call invokes the function asynchronously.  It returns the Call structure representing
// the invocation.  If reply is nil, no reply is expected.
func (client *Client) Call(topic string, serviceMethod string, args interface{}, reply interface{}) (*Call, error) {
	call := new(Call)
	call.Topic = topic
	call.ServiceMethod = serviceMethod
	call.Args = args

	if reply == nil {
		// No reply is expected
	} else {
		call.Done = make(chan *Call, 1)
		call.Reply = reply
	}
	return client.send(call)
}

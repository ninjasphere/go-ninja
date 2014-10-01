// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("rpc")

// ClientCodec encodes and decodes the calls and replies (currently, to json)
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
	ID            uint32      // Used to map responses
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	mutex      sync.Mutex // protects following
	codec      ClientCodec
	mqtt       *mqtt.MqttClient
	pending    map[uint32]*Call
	subscribed map[string]bool
}

// NewClient creates a new rpc client using the provided MQTT connection
func NewClient(mqtt *mqtt.MqttClient, codec ClientCodec) *Client {
	client := &Client{
		pending:    make(map[uint32]*Call),
		subscribed: make(map[string]bool),
		mqtt:       mqtt,
		codec:      codec,
	}
	return client
}

func (client *Client) send(call *Call) error {

	payload, err := client.codec.EncodeClientRequest(call)
	if err != nil {
		return err
	}

	replyTopic := call.Topic + "/reply"

	if !client.subscribed[replyTopic] {

		log.Debugf("Subscribing to %s", replyTopic)

		filter, err := mqtt.NewTopicFilter(replyTopic, 0)
		if err != nil {
			return err
		}

		receipt, err := client.mqtt.StartSubscription(func(mqtt *mqtt.MqttClient, message mqtt.Message) {
			log.Debugf("< Incoming to %s : %s", message.Topic(), message.Payload())
			go client.handleResponse(message)
		}, filter)

		if err != nil {
			return err
		}

		client.subscribed[replyTopic] = true

		<-receipt
	}

	log.Debugf("< Outgoing to %s : %s", call.Topic, payload)

	pubReceipt := client.mqtt.Publish(mqtt.QoS(0), call.Topic, payload)

	<-pubReceipt

	client.pending[call.ID] = call

	return nil
}

func (client *Client) handleResponse(message mqtt.Message) {
	id, err := client.codec.DecodeIdAndError(message.Payload())

	if id == nil {
		log.Debugf("Failed to decode reply: %s error: %s", message.Payload(), err)
		return
	}

	client.mutex.Lock()
	call := client.pending[*id]
	delete(client.pending, *id)
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

	if call.Reply != nil {
		call.Error = client.codec.DecodeClientResponse(message.Payload(), call.Reply)
	}

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

// CallWithTimeout invokes a function synchronously.
func (client *Client) CallWithTimeout(topic string, serviceMethod string, args interface{}, reply interface{}, timeout time.Duration) error {
	call := &Call{
		ID:            rand.Uint32(),
		Topic:         topic,
		ServiceMethod: serviceMethod,
		Args:          args,
		Done:          make(chan *Call, 1),
		Reply:         reply,
	}

	err := client.send(call)
	if err != nil {
		return err
	}

	log.Infof("Waiting for reply...")

	select {
	case <-call.Done:
		return call.Error
	case <-time.After(timeout):
		delete(client.pending, call.ID)
		return fmt.Errorf("Call to service %s - (method: %s) timed out after %d seconds", topic, serviceMethod, timeout/time.Second)
	}

}

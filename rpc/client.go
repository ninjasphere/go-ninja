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

	"github.com/ninjasphere/go-ninja/bus"
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
	mqtt       *bus.Bus
	pending    map[uint32]*Call
	subscribed map[string]bool
}

// NewClient creates a new rpc client using the provided MQTT connection
func NewClient(mqtt *bus.Bus, codec ClientCodec) *Client {
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

	if call.Done != nil {
		replyTopic := call.Topic + "/reply"

		if !client.subscribed[replyTopic] {

			log.Debugf("Subscribing to %s", replyTopic)

			_, err := client.mqtt.Subscribe(replyTopic, func(topic string, payload []byte) {
				log.Debugf("< Incoming to %s : %s", topic, payload)
				go client.handleResponse(topic, payload)
			})

			if err != nil {
				return err
			}

			client.subscribed[replyTopic] = true
		}
	}

	log.Debugf("< Outgoing to %s : %s", call.Topic, payload)

	client.mqtt.Publish(call.Topic, payload)

	if call.Done != nil {
		client.pending[call.ID] = call
	}

	return nil
}

func (client *Client) handleResponse(topic string, payload []byte) {
	id, err := client.codec.DecodeIdAndError(payload)

	if id == nil {
		log.Debugf("Failed to decode reply: %s error: %s", payload, err)
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
		call.Error = client.codec.DecodeClientResponse(payload, call.Reply)
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

// Call invokes a function asynchronously.
func (client *Client) Call(topic string, serviceMethod string, args interface{}) error {
	call := &Call{
		ID:            rand.Uint32(),
		Topic:         topic,
		ServiceMethod: serviceMethod,
		Args:          args,
	}

	return client.send(call)
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
	sentTime := time.Now()

	log.Debugf("id:%d -  Waiting for reply", call.ID)

	select {
	case <-call.Done:
		log.Debugf("id:%d - Returned after %s", call.ID, time.Since(sentTime))
		return call.Error
	case <-time.After(timeout):
		delete(client.pending, call.ID)
		return fmt.Errorf("id:%d - Call to service %s - (method: %s) timed out after %d seconds", call.ID, topic, serviceMethod, timeout/time.Second)
	}

}

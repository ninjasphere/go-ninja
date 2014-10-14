package ninja

import (
	"encoding/json"
	"time"
)

type ServiceClient struct {
	conn  *Connection
	topic string
}

// OnEvent builds a simple subscriber which supports pulling apart the topic
//
// 	err := sm.conn.GetServiceClient("$device/:deviceid/channel/:channelid").OnEvent("state", func(params *json.RawMessage) {
//  	..
//	}, true) // true continues to consume messages
//
//
func (c *ServiceClient) OnEvent(event string, callback func(params *json.RawMessage, values map[string]string) bool) error {
	return c.conn.Subscribe(c.topic+"/event/"+event, func(params *json.RawMessage, values map[string]string) bool {
		return callback(params, values)
	})
}

//
// OnUnmarshalledEvent builds a subscriber that attempts to unmarshall the JSON object onto the first argument
// of the callback which must be a function which matches:
//
//    func(ptr *<T>) bool
//    func(ptr *<T>, values map[string]string) bool) bool
//
// where <T> is go struct type which into which the expected JSON event payload can be successfully unmarshalled.
//
// values, if supplied, is a map of parameters as per OnEvent
//
func (c *ServiceClient) OnUnmarshalledEvent(event string, callback interface{}) error {
	return c.conn.SimplySubscribe(c.topic+"/event/"+event, callback)
}

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.conn.rpc.CallWithTimeout(c.topic, method, args, reply, timeout)
}

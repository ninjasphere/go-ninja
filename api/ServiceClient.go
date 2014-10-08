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

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.conn.rpc.CallWithTimeout(c.topic, method, args, reply, timeout)
}

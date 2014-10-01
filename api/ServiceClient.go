package ninja

import (
	"encoding/json"
	"time"
)

type ServiceClient struct {
	conn  *Connection
	topic string
}

func (c *ServiceClient) OnEvent(event string, callback func(params *json.RawMessage) bool) error {
	return c.conn.Subscribe(c.topic+"/event/"+event, func(params *json.RawMessage, values map[string]string) bool {
		return callback(params)
	})
}

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.conn.rpc.CallWithTimeout(c.topic, method, args, reply, timeout)
}

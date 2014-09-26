package ninja

import (
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type ServiceClient struct {
	conn  *Connection
	topic string
}

func (c *ServiceClient) OnEvent(event string, callback func(message mqtt.Message) bool) error {
	return c.conn.Subscribe(c.topic+"/event/"+event, func(message mqtt.Message, values map[string]string) bool {
		return callback(message)
	})
}

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.conn.rpc.CallWithTimeout(c.topic, method, args, reply, timeout)
}

package ninja

import "time"

type ServiceClient struct {
	conn  *Connection
	topic string
}

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	return c.conn.rpc.CallWithTimeout(c.topic, method, args, reply, timeout)
}

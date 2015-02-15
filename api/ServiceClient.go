package ninja

import (
	"fmt"
	"time"

	"github.com/ninjasphere/go-ninja/bus"
)

type ServiceClient struct {
	conn             *Connection
	Topic            string
	SupportedEvents  []string
	SupportedMethods []string
}

//
// OnEvent builds a simple subscriber which supports pulling apart the topic
//
// 	err := sm.conn.GetServiceClient("$device/:deviceid/channel/:channelid")
//                     .OnEvent("state", func(params *YourEventType, topicKeys map[string]string) bool {
//  	..
//	    return true
//	})
//
// YourEventType must either be *json.RawMessage or a pointer to go type to which the raw JSON message can successfully be unmarshalled.
//
// There is one entry in the topicKeys map for each parameter marker in the topic string used to obtain the ServiceClient.
//
// Both the params and topicKeys parameters can be omitted. If the topicKeys parameter is required, the params parameter must also be specified.
//
func (c *ServiceClient) OnEvent(event string, callback interface{}) (*bus.Subscription, error) {
	return c.conn.Subscribe(c.Topic+"/event/"+event, callback)
}

func (c *ServiceClient) Call(method string, args interface{}, reply interface{}, timeout time.Duration) error {
	if timeout > 0 {
		return c.conn.rpc.CallWithTimeout(c.Topic, method, args, reply, timeout)
	}

	if reply != nil {
		return fmt.Errorf("Attempted async call to method %s on service %s with a non-nil reply", method, c.Topic)
	}

	return c.conn.rpc.Call(c.Topic, method, args)
}

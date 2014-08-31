package rpc2

import (
	"net/rpc"
	"net/rpc/jsonrpc"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ninjasphere/go-ninja/logger"
)

func GetClient(topic string, mqttConn *mqtt.MqttClient) (*rpc.Client, error) {

	log := logger.GetLogger("RPC - " + topic)

	conn, err := NewMqttJsonRpcConnection(false, mqttConn, topic, log)

	if err != nil {
		return nil, err
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	return client, nil
}

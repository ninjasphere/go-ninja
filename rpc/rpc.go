package rpc

import (
	"time"

	"github.com/bitly/go-simplejson"
)

// Alias so we can test this thing
var unixTimestampFunc = time.Now().Unix

// BuildRPCRequest Using the supplied params assemble a json RPC message and marshal it.
func BuildRPCRequest(params ...*simplejson.Json) ([]byte, error) {

	jsonmsg, err := simplejson.NewJson([]byte(`{"params": [],"time": "","jsonrpc": "2.0"}`))

	if err != nil {
		return nil, err
	}

	jsonmsg.Set("params", params)
	jsonmsg.Set("time", unixTimestampFunc())

	return jsonmsg.MarshalJSON()
}

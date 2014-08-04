package rpc

import (
	"bytes"
	"log"
	"reflect"
	"testing"

	"github.com/bitly/go-simplejson"
)

func TestBuildRPCRequest(t *testing.T) {

	unixTimestampFunc = func() int64 { return 1407135638 }

	rpc := struct {
		param    *simplejson.Json
		expected []byte
	}{
		param: createJSON(map[string]interface{}{
			"version": 123,
			"message": "hello",
		}),
		expected: bytes.NewBufferString(`{"jsonrpc":"2.0","params":[{"message":"hello","version":123}],"time":1407135638}`).Bytes(),
	}

	if result, err := BuildRPCRequest(rpc.param); err != nil {
		t.Fatalf("Error building rpc request %v", err)
	} else {

		log.Printf("result : %s expected : %s", result, rpc.expected)

		if !reflect.DeepEqual(rpc.expected, result) {
			t.Fatalf("Result doesn't match, expected : %s, result: %s", rpc.expected, result)
		}

	}

}

func createJSON(data map[string]interface{}) *simplejson.Json {

	js, _ := simplejson.NewJson([]byte(`{}`))

	for key, value := range data {
		js.Set(key, value)
	}

	return js
}

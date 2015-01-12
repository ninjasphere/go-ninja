// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2014 Ninja Blocks Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json2

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjasphere/go-ninja/rpc"
)

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// clientRequest represents a JSON-RPC request sent by a client.
type clientRequest struct {

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// Object to pass as request parameter to the method.
	Params interface{} `json:"params"`

	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	ID string `json:"id"`

	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`
}

// clientResponse represents a JSON-RPC response returned to a client.
type clientResponse struct {
	Result  *json.RawMessage `json:"result"`
	Error   *json.RawMessage `json:"error"`
	ID      *json.RawMessage `json:"id"`
	Version string           `json:"jsonrpc"`
}

func NewClientCodec() *ClientCodec {
	return &ClientCodec{}
}

type ClientCodec struct {
}

// EncodeClientRequest encodes parameters for a JSON-RPC client request.
func (c *ClientCodec) EncodeClientRequest(call *rpc.Call) ([]byte, error) {
	req := &clientRequest{
		Version: "2.0",
		Method:  call.ServiceMethod,
		Params:  []interface{}{},
		ID:      fmt.Sprintf("%d", call.ID),
	}

	if call.Args != nil {
		req.Params = call.Args
	}

	return json.Marshal(req)
}

func (c *ClientCodec) DecodeIdAndError(msg []byte) (*uint32, error) {
	res := &clientResponse{}

	if err := json.Unmarshal(msg, res); err != nil {
		return nil, err
	}

	var id uint32
	err := json.Unmarshal(*res.ID, &id)
	if err != nil {

		var sID string
		err = json.Unmarshal(*res.ID, &sID)

		if err == nil {
			var bigID uint64
			bigID, err = strconv.ParseUint(sID, 10, 32)
			id = uint32(bigID)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("Reply id isn't a uint32 or string uint32. Probably not for us '%s'", *res.ID)
	}

	if res.Error != nil {
		jsonErr := &Error{}
		if err := json.Unmarshal(*res.Error, jsonErr); err != nil {
			return &id, &Error{
				Code:    E_SERVER,
				Message: string(*res.Error),
			}
		}
		return &id, jsonErr
	}

	return &id, nil

}

// DecodeClientResponse decodes the response body of a client request into
// the interface reply.
func (c *ClientCodec) DecodeClientResponse(msg []byte, reply interface{}) error {
	var res clientResponse
	if err := json.Unmarshal(msg, &res); err != nil {
		return err
	}
	if res.Result != nil {
		return json.Unmarshal(*res.Result, reply)
	}
	return nil
}

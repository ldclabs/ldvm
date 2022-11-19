// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"encoding/json"
	"fmt"
)

const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	// -32000 to -32599	Server error, Reserved for implementation-defined server-errors.
	CodeServerError = -32000
)

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	data, er := json.Marshal(e)
	if er == nil {
		return string(data)
	}

	message := er.Error()
	if e.Message != "" {
		message = e.Message + ", " + message
	}

	return fmt.Sprintf(`{"code":%d,"message":%q}`, e.Code, message)
}

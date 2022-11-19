// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import "context"

const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = "application/json; charset=utf-8"
)

type Handler interface {
	ServeRPC(context.Context, *Request) *Response
	OnError(context.Context, *Error)
}

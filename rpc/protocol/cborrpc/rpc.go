// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import "context"

const (
	MIMEApplicationCBOR = "application/cbor"
)

type Handler interface {
	ServeRPC(context.Context, *Request) *Response
}

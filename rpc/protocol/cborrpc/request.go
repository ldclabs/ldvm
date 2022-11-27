// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"fmt"
	"io"

	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/util/encoding"
)

// This is a simple implementation of CBOR-RPC.
// Full reference to https://www.jsonrpc.org/specification
type Request struct {
	ID     string          `cbor:"id,omitempty"`
	Method string          `cbor:"method"`
	Params cbor.RawMessage `cbor:"params,omitempty"`

	buf bytes.Buffer `cbor:"-"`
}

// Grow grows the underlying buffer's capacity
func (req *Request) Grow(n int) {
	req.buf.Grow(n)
}

// ReadFrom implements io.ReaderFrom interface.
func (req *Request) ReadFrom(r io.Reader) (int64, error) {
	n, err := req.buf.ReadFrom(r)
	if err != nil {
		return n, err
	}

	if err = encoding.UnmarshalCBOR(req.buf.Bytes(), req); err != nil {
		return n, err
	}

	return n, nil
}

func (req *Request) String() string {
	return fmt.Sprintf(`{"id":%q,"method":%q,"params":%s}`,
		req.ID, req.Method, string(ToJSON(req.Params)))
}

func (req *Request) DecodeParams(params interface{}) error {
	if err := encoding.UnmarshalCBOR(req.Params, params); err != nil {
		return &Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
		}
	}
	return nil
}

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ldclabs/ldvm/util/erring"
)

// This is a simple implementation of JSON-RPC 2.0.
// https://www.jsonrpc.org/specification

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id"` // only support string id
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// ReadFrom decodes the request from the given reader.
// ReadFrom implements io.ReaderFrom interface.
func (req *Request) ReadFrom(r io.Reader) (int64, error) {
	jd := json.NewDecoder(r)

	if err := jd.Decode(req); err != nil {
		return 0, err
	}

	n := jd.InputOffset()
	if jd.More() {
		return n, errors.New("json: unexpected following extraneous data")
	}

	if req.Version != "2.0" || req.ID == "" || req.Method == "" {
		return n, fmt.Errorf("invalid request, %q", req.String())
	}

	return n, nil
}

// String returns the string representation of the request.
func (req *Request) String() string {
	b, err := json.Marshal(req)
	if err != nil {
		b, _ = json.Marshal(erring.RespondError{Err: err.Error()})
	}

	return string(b)
}

// DecodeParams decodes the request parameters into the given value.
func (req *Request) DecodeParams(params interface{}) error {
	if err := json.Unmarshal(req.Params, params); err != nil {
		return &Error{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("invalid parameter(s), %v", err),
		}
	}
	return nil
}

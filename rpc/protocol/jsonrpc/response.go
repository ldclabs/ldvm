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

// refer to JSON-RPC 2.0. https://www.jsonrpc.org/specification
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	// -32000 to -32599	Server error, Reserved for implementation-defined server-errors.
	CodeServerError = -32000
)

// Error is a JSON-RPC error.
type Error = erring.Error

// Response represents a JSON-RPC response.
type Response struct {
	Version string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Error   *Error          `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

// ReadFrom decodes the JSON-RPC response from the given reader.
// ReadFrom implements io.ReaderFrom interface.
func (res *Response) ReadFrom(r io.Reader) (int64, error) {
	jd := json.NewDecoder(r)

	if err := jd.Decode(res); err != nil {
		return 0, err
	}

	n := jd.InputOffset()
	if jd.More() {
		return n, errors.New("json: unexpected following extraneous data")
	}

	if res.Version != "2.0" || res.ID == "" || (res.Error == nil && len(res.Result) == 0) {
		return n, fmt.Errorf("invalid response, %q", res.String())
	}

	return n, nil
}

// String returns the string representation of the response.
func (res *Response) String() string {
	b, err := json.Marshal(res)
	if err != nil {
		b, _ = json.Marshal(erring.RespondError{Err: err.Error()})
	}

	return string(b)
}

// DecodeResult decodes the result into the given value.
func (res *Response) DecodeResult(result interface{}) error {
	if res.Error != nil {
		return res.Error
	}

	if err := json.Unmarshal(res.Result, result); err != nil {
		return &Error{Code: CodeParseError, Message: err.Error()}
	}
	return nil
}

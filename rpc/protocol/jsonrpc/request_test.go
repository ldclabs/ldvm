// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	assert := assert.New(t)

	req := &Request{}
	str := `{"jsonrpc":"2.0","id":1,"method":"getTx"}`
	n, err := req.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `json: cannot unmarshal number`)
	assert.Equal(int64(0), n)

	req = &Request{}
	str = `{"jsonrpc":"2.0","id":"1","method":"getTx"}_`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `json: unexpected following extraneous data`)
	assert.Equal(int64(43), n)

	req = &Request{}
	str = `{"jsonrpc":"1.0","id":"1","method":"getTx"}`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid request`)
	assert.Equal(int64(43), n)

	req = &Request{}
	str = `{"jsonrpc":"2.0","id":"","method":"getTx"}`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid request`)
	assert.Equal(int64(42), n)

	req = &Request{}
	str = `{"jsonrpc":"2.0","id":"1"}`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid request`)
	assert.Equal(int64(26), n)

	req = &Request{}
	str = `{"jsonrpc":"2.0","id":"1","method":"getTx"}`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.NoError(err)
	assert.Equal(int64(43), n)
	assert.Equal(str, req.String())

	var r0 map[string]interface{}
	err = req.DecodeParams(&r0)
	assert.Equal(CodeInvalidParams, err.(*Error).Code)
	assert.ErrorContains(err, `invalid parameter(s), unexpected end of JSON input`)

	req = &Request{}
	str = `{"jsonrpc":"2.0","id":"1","method":"getTx","params":{"tx":"abc"}}`
	n, err = req.ReadFrom(bytes.NewBufferString(str))
	assert.NoError(err)
	assert.Equal(int64(65), n)
	assert.Equal(str, req.String())

	r0 = map[string]interface{}{}
	err = req.DecodeParams(&r0)
	assert.NoError(err)
	assert.Equal("abc", r0["tx"].(string))

	req = &Request{
		Version: "2.0",
		ID:      "123",
		Method:  "getTx",
		Params:  []byte("hello"),
	}
	assert.Contains(req.String(), `{"error":"json: error calling MarshalJSON for type json.RawMessage`)
}

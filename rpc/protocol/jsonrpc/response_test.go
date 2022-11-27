// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package jsonrpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	assert := assert.New(t)

	res := &Response{}
	str := `{"jsonrpc":"2.0","id":1,"result":true}`
	n, err := res.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `son: cannot unmarshal number`)
	assert.Equal(int64(0), n)

	res = &Response{}
	str = `{"jsonrpc":"2.0","id":"1","result":true}_`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `json: unexpected following extraneous data`)
	assert.Equal(int64(40), n)

	res = &Response{}
	str = `{"jsonrpc":"1.0","id":"1","result":true}`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid response`)
	assert.Equal(int64(40), n)

	res = &Response{}
	str = `{"jsonrpc":"2.0","id":"","result":true}`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid response`)
	assert.Equal(int64(39), n)

	res = &Response{}
	str = `{"jsonrpc":"2.0","id":"1"}`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.ErrorContains(err, `invalid response`)
	assert.Equal(int64(26), n)

	res = &Response{}
	str = `{"jsonrpc":"2.0","id":"1","result":true}`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.NoError(err)
	assert.Equal(int64(40), n)
	assert.Equal(str, res.String())

	var r0 map[string]interface{}
	err = res.DecodeResult(&r0)
	assert.Equal(CodeParseError, err.(*Error).Code)
	assert.ErrorContains(err, `json: cannot unmarshal bool`)

	var r1 bool
	err = res.DecodeResult(&r1)
	assert.NoError(err)
	assert.True(r1)

	res = &Response{}
	str = `{"jsonrpc":"2.0","id":"1","error":{"code":400,"message":"bad request"}}`
	n, err = res.ReadFrom(bytes.NewBufferString(str))
	assert.NoError(err)
	assert.Equal(int64(71), n)
	assert.Equal(str, res.String())

	r0 = map[string]interface{}{}
	err = res.DecodeResult(&r0)
	assert.Equal(400, err.(*Error).Code)
	assert.ErrorContains(err, `bad request`)

	res = &Response{
		Version: "2.0",
		ID:      "123",
		Result:  []byte("hello"),
	}
	assert.Contains(res.String(), `{"error":"json: error calling MarshalJSON for type json.RawMessage`)
}

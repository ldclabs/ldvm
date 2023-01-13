// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	assert := assert.New(t)

	res := &Response{}
	data := cborpatch.MustFromJSON(`{"id":1,"result":true}`)
	n, err := res.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `cbor: cannot unmarshal positive integer`)
	assert.Equal(int64(13), n)

	res = &Response{}
	data = cborpatch.MustFromJSON(`{"id":"1","result":true}`)
	data = append(data, byte(254))
	n, err = res.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `extraneous data`)
	assert.Equal(int64(15), n)

	res = &Response{}
	data = cborpatch.MustFromJSON(`{"id":"","result":true}`)
	n, err = res.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `invalid response`)
	assert.Equal(int64(13), n)

	res = &Response{}
	data = cborpatch.MustFromJSON(`{"id":"1"}`)
	n, err = res.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `invalid response`)
	assert.Equal(int64(6), n)

	res = &Response{}
	str := `{"id":"1","result":true}`
	data = cborpatch.MustFromJSON(str)
	n, err = res.ReadFrom(bytes.NewBuffer(data))
	assert.NoError(err)
	assert.Nil(res.Error)
	assert.NotNil(res.Result)
	assert.Equal(int64(14), n)
	assert.Equal(str, res.String())

	var r0 map[string]any
	err = res.DecodeResult(&r0)
	assert.Equal(CodeParseError, err.(*Error).Code)
	assert.ErrorContains(err, `cbor: cannot unmarshal primitives`)

	var r1 bool
	err = res.DecodeResult(&r1)
	assert.NoError(err)
	assert.True(r1)

	res = &Response{}
	str = `{"id":"1","error":{"code":400,"message":"bad request"}}`
	data = cborpatch.MustFromJSON(str)
	n, err = res.ReadFrom(bytes.NewBuffer(data))
	assert.NoError(err)
	assert.NotNil(res.Error)
	assert.Nil(res.Result)
	assert.Equal(int64(41), n)
	assert.Equal(str, res.String())

	r0 = map[string]any{}
	err = res.DecodeResult(&r0)
	assert.Equal(400, err.(*Error).Code)
	assert.ErrorContains(err, `bad request`)

	res = &Response{
		ID:     "123",
		Result: []byte{255, 254, 253, 252},
	}
	assert.Contains(res.String(), `{"error":"cbor: unexpected`)
}

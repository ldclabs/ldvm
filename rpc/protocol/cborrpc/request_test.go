// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"bytes"
	"testing"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	assert := assert.New(t)

	req := &Request{}
	data := cborpatch.MustFromJSON(`{"id":1,"method":"getTx"}`)
	n, err := req.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `cbor: cannot unmarshal positive integer`)
	assert.Equal(int64(18), n)

	req = &Request{}
	data = cborpatch.MustFromJSON(`{"id":"1","method":"getTx"}`)
	data = append(data, byte(254))
	n, err = req.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `extraneous data`)
	assert.Equal(int64(20), n)

	req = &Request{}
	data = cborpatch.MustFromJSON(`{"id":"","method":"getTx"}`)
	n, err = req.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `invalid request`)
	assert.Equal(int64(18), n)

	req = &Request{}
	data = cborpatch.MustFromJSON(`{"id":"1"}`)
	n, err = req.ReadFrom(bytes.NewBuffer(data))
	assert.ErrorContains(err, `invalid request`)
	assert.Equal(int64(6), n)

	req = &Request{}
	str := `{"id":"1","method":"getTx"}`
	data = cborpatch.MustFromJSON(str)
	n, err = req.ReadFrom(bytes.NewBuffer(data))
	assert.NoError(err)
	assert.Equal(int64(19), n)
	assert.Equal(str, req.String())

	var r0 map[string]any
	err = req.DecodeParams(&r0)
	assert.Equal(CodeInvalidParams, err.(*Error).Code)
	assert.ErrorContains(err, `invalid parameter(s), EOF`)

	req = &Request{}
	str = `{"id":"1","method":"getTx","params":{"tx":"abc"}}`
	data = cborpatch.MustFromJSON(str)
	n, err = req.ReadFrom(bytes.NewBuffer(data))
	assert.NoError(err)
	assert.Equal(int64(34), n)
	assert.Equal(str, req.String())

	r0 = map[string]any{}
	err = req.DecodeParams(&r0)
	assert.NoError(err)
	assert.Equal("abc", r0["tx"].(string))

	req = &Request{
		ID:     "123",
		Method: "getTx",
		Params: []byte{255, 254, 253, 252},
	}
	assert.Contains(req.String(), `{"error":"cbor: unexpected`)
}

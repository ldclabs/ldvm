// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cborrpc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/ldclabs/ldvm/util/encoding"
)

func TestRPC(t *testing.T) {
	assert := assert.New(t)

	req := &Request{
		ID:     "123",
		Method: "getTx",
		Params: encoding.MustMarshalCBOR("hello"),
	}

	res := req.InvalidMethod()
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeMethodNotFound, res.Error.Code)
	assert.Contains(res.Error.Message, `method "getTx" not found`)

	res = req.InvalidParams("some error")
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeInvalidParams, res.Error.Code)
	assert.Contains(res.Error.Message, `invalid parameter(s), some error`)

	var er error
	er = &Error{
		Code:    CodeServerError,
		Message: "some server error",
	}
	res = req.Error(er)
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeServerError, res.Error.Code)
	assert.Contains(res.Error.Message, `some server error`)
	assert.True(errors.Is(er, res.Error))

	er = fmt.Errorf("some error, %w", &Error{
		Code:    CodeInternalError,
		Message: "some internal error",
	})
	res = req.Error(er)
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeInternalError, res.Error.Code)
	assert.Contains(res.Error.Message, `some internal error`)
	assert.True(errors.Is(er, res.Error))

	er = fmt.Errorf("some text error")
	res = req.Error(er)
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeServerError, res.Error.Code)
	assert.Contains(res.Error.Message, `some text error`)
	assert.False(errors.Is(er, res.Error))

	res = req.Result(func() {})
	assert.Equal("123", res.ID)
	assert.Nil(res.Result)
	assert.NotNil(res.Error)
	assert.Equal(CodeInternalError, res.Error.Code)
	assert.Contains(res.Error.Message, `internal error, cbor: unsupported type: func()`)

	res = req.Result("hello world")
	assert.Equal("123", res.ID)
	assert.NotNil(res.Result)
	assert.Nil(res.Error)
	assert.Equal(
		`{"id":"123","result":"hello world"}`,
		res.String())
	assert.Equal(
		`{"id":"123","result":"hello world"}`,
		cborpatch.MustToJSON(encoding.MustMarshalCBOR(res)))
}

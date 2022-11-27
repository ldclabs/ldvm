// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package erring

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("empty Error", func(t *testing.T) {
		assert := assert.New(t)
		err := &Error{}
		assert.False(err.HasErrs())
		assert.Nil(err.Unwrap())
		assert.False((err.Is(io.EOF)))
		assert.False(errors.Is(err, io.EOF))
		assert.True((err.Is(err)))
		assert.True(errors.Is(err, err))
		assert.Equal(
			`{"code":0,"message":"","errors":[],"data":null}`,
			err.Error())

		var er error
		assert.False((err.As(&er)))
		assert.False(err.CatchIf(er))
		assert.False(err.HasErrs())
		assert.Nil(err.Unwrap())
		assert.False((err.Is(io.EOF)))

		assert.True(err.CatchIf(io.EOF))
		assert.True(err.HasErrs())
		assert.True((err.Is(io.EOF)))
		assert.True(errors.Is(err, io.EOF))
		assert.True((err.Is(err)))
		assert.True(errors.Is(err, err))
		assert.True((err.As(&er)))
		assert.Equal(io.EOF, er)

		assert.Equal(`EOF`, er.Error())
		assert.Equal(
			`{"code":0,"message":"","errors":["EOF"],"data":null}`,
			err.Error())
	})

	t.Run("some Error", func(t *testing.T) {
		assert := assert.New(t)
		err := &Error{
			Code:    500,
			Message: "Internal server error",
			Data:    map[string]interface{}{"foo": "bar"},
		}

		assert.False(err.HasErrs())
		assert.Nil(err.Unwrap())
		assert.False((err.Is(io.ErrUnexpectedEOF)))
		assert.False(errors.Is(err, io.ErrUnexpectedEOF))
		assert.True((err.Is(err)))
		assert.True(errors.Is(err, err))
		assert.Equal(
			`{"code":500,"message":"Internal server error","errors":[],"data":{"foo":"bar"}}`,
			err.Error())

		assert.True(err.CatchIf(io.ErrUnexpectedEOF))
		assert.True(err.HasErrs())

		err2 := &Error{}
		assert.False((err.Is(err2)))
		assert.False(errors.Is(err, err2))

		assert.NotEqual(err.Error(), err2.Error())
		assert.True((err.As(err2)))
		assert.Equal(err.Error(), err2.Error())
		assert.False(errors.Is(err, err2))

		err2 = &Error{}
		assert.NotEqual(err.Error(), err2.Error())
		assert.True((errors.As(err, &err2)))
		assert.Equal(err.Error(), err2.Error())

		assert.True((err2.Is(io.ErrUnexpectedEOF)))
		assert.True(errors.Is(err2, io.ErrUnexpectedEOF))

		err2 = &Error{}
		assert.True(err2.CatchIf(err))
		assert.True((err2.Is(err)))
		assert.True(errors.Is(err2, err))
		assert.True((err2.Is(io.ErrUnexpectedEOF)))
		assert.True(errors.Is(err2, io.ErrUnexpectedEOF))
		assert.Equal(
			`{"code":0,"message":"","errors":["{\"code\":500,\"message\":\"Internal server error\",\"errors\":[\"unexpected EOF\"],\"data\":{\"foo\":\"bar\"}}"],"data":null}`,
			err2.Error())
	})
}

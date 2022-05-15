// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawObject(t *testing.T) {
	assert := assert.New(t)

	var nilObj *RawObject
	data := []byte("hello")
	assert.ErrorContains(nilObj.Unmarshal(data), "nil pointer")

	obj := make(RawObject, 10)
	ptr := fmt.Sprintf("%p", &obj)
	assert.NoError(obj.Unmarshal(data))
	assert.Equal(data, []byte(obj))
	assert.Equal(ptr, fmt.Sprintf("%p", &obj), "should reuse underlying array")
}

func TestCacher(t *testing.T) {
	assert := assert.New(t)

	cc := NewCacher(100, 1, nil)

	v, ok := cc.GetObject([]byte("k1"))
	assert.Nil(v)
	assert.False(ok)

	ro := RawObject("Hello")
	cc.SetObject([]byte("k1"), ro)

	v2, ok := cc.GetObject([]byte("k1"))
	assert.True(ok)
	assert.Equal(ro, v2.(RawObject))

	_, err := cc.UnmarshalObject([]byte("Hello"))
	assert.ErrorContains(err, "no function to create object")
}

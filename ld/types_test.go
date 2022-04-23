// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint8(t *testing.T) {
	assert := assert.New(t)
	type Case struct {
		val uint8
		str string
	}

	var cases = []Case{
		{val: uint8(0), str: "0"},
		{val: uint8(255), str: "255"},
		{val: uint8(2), str: "2"},
		{val: uint8(1), str: "1"},
	}

	for _, ca := range cases {
		rt := FromUint8(ca.val)
		assert.Equal(rt.Value(), ca.val)
		assert.Equal(rt.String(), ca.str)

		ptr := PtrFromUint8(ca.val)
		switch ca.val {
		case 0:
			assert.Nil(ptr)
		default:
			assert.Equal(ptr.Value(), ca.val)
		}
	}
}

func TestUint64(t *testing.T) {
	assert := assert.New(t)

	type Case struct {
		val uint64
		str string
	}

	var cases = []Case{
		{val: uint64(0), str: "0"},
		{val: uint64(10_000_000_000_000_000_000), str: "10000000000000000000"},
		{val: uint64(10_000_000_000_000_000), str: "10000000000000000"},
		{val: uint64(10_000_000_000_000), str: "10000000000000"},
		{val: uint64(10_000_000_000), str: "10000000000"},
		{val: uint64(10_000_000), str: "10000000"},
		{val: uint64(10_000), str: "10000"},
		{val: uint64(10), str: "10"},
	}

	for _, ca := range cases {
		rt := FromUint64(ca.val)
		assert.Equal(rt.Value(), ca.val)
		assert.Equal(rt.String(), ca.str)

		ptr := PtrFromUint64(ca.val)
		switch ca.val {
		case 0:
			assert.Nil(ptr)
		default:
			assert.Equal(ptr.Value(), ca.val)
		}
	}
}

func FuzzUint64(f *testing.F) {
	for _, seed := range []uint64{
		1,
		18_000_000_000_000_000_000,
		18_000_000_000_000_000,
		18_000_000_000_000,
		18_000_000_000,
		18_000_000,
		18_000,
		18,
		0,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in uint64) {
		assert := assert.New(t)

		rt := FromUint64(in)
		assert.Equal(rt.Value(), in)

		ptr := PtrFromUint64(in)
		switch in {
		case 0:
			assert.Nil(ptr)
		default:
			assert.Equal(ptr.Value(), in)
		}
	})
}

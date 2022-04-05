// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"testing"
)

func TestUint64(t *testing.T) {
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
		if rt.Value() != ca.val {
			t.Fatalf("Expected Equal(%v, %v)", rt.Value(), ca.val)
		}
		if rt.String() != ca.str {
			t.Fatalf("Expected Equal(%v, %v)", rt.String(), ca.str)
		}
		ptr := PtrFromUint64(ca.val)
		if ca.val == 0 && ptr != nil {
			t.Fatalf("Expected Equal(%v, %v)", ptr, nil)
		}
		if ptr.Value() != ca.val {
			t.Fatalf("Expected Equal(%v, %v)", ptr.Value(), ca.val)
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
		rt := FromUint64(in)
		if rt.Value() != in {
			t.Fatalf("Expected Equal(%v, %v)", rt.Value(), in)
		}
		ptr := PtrFromUint64(in)
		if in == 0 && ptr != nil {
			t.Fatalf("Expected Equal(%v, %v)", ptr, nil)
		}
		if ptr.Value() != in {
			t.Fatalf("Expected Equal(%v, %v)", ptr.Value(), in)
		}
	})
}

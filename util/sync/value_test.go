// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	t.Run("Value[string]", func(t *testing.T) {
		assert := assert.New(t)

		var v Value[string]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Equal("", vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		vv, ok = v.Swap("a")
		assert.False(ok)
		assert.Equal("", vv)
		assert.Equal("a", v.MustLoad())

		vv, ok = v.Swap("b")
		assert.True(ok)
		assert.Equal("a", vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal("b", vv)

		assert.False(v.CompareAndSwap("c", "c"))
		assert.True(v.CompareAndSwap("b", "c"))

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal("c", vv)

		v.Store("cc")
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal("cc", vv)
	})

	t.Run("Value[int]", func(t *testing.T) {
		assert := assert.New(t)

		var v Value[int]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Equal(0, vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		vv, ok = v.Swap(1)
		assert.False(ok)
		assert.Equal(0, vv)
		assert.Equal(1, v.MustLoad())

		vv, ok = v.Swap(2)
		assert.True(ok)
		assert.Equal(1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(2, vv)

		assert.False(v.CompareAndSwap(3, 3))
		assert.True(v.CompareAndSwap(2, 3))

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(3, vv)

		v.Store(33)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(33, vv)
	})

	t.Run("Value[[4]byte]", func(t *testing.T) {
		assert := assert.New(t)

		var v Value[[4]byte]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Equal([4]byte{}, vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		v1 := [4]byte{1, 2, 3}
		vv, ok = v.Swap(v1)
		assert.False(ok)
		assert.Equal([4]byte{}, vv)
		assert.Equal(v1, v.MustLoad())

		v2 := [4]byte{2, 3, 4}
		vv, ok = v.Swap(v2)
		assert.True(ok)
		assert.Equal(v1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v2, vv)

		v3 := [4]byte{3, 4, 5}
		assert.False(v.CompareAndSwap(v1, v3))
		assert.True(v.CompareAndSwap(v2, v3))

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v3, vv)

		vv3 := [4]byte{33, 44, 55}
		v.Store(vv3)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(vv3, vv)
	})

	t.Run("Value[[]byte]", func(t *testing.T) {
		assert := assert.New(t)

		var v Value[[]byte]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Nil(vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		v1 := []byte{1, 2, 3}
		vv, ok = v.Swap(v1)
		assert.False(ok)
		assert.Nil(vv)
		assert.Equal(v1, v.MustLoad())

		v2 := []byte{2, 3, 4}
		vv, ok = v.Swap(v2)
		assert.True(ok)
		assert.Equal(v1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v2, vv)

		assert.Panics(func() {
			v3 := []byte{3, 4, 5}
			assert.False(v.CompareAndSwap(v1, v3))
			assert.True(v.CompareAndSwap(v2, v3))

			vv, ok = v.Load()
			assert.True(ok)
			assert.Equal(v3, vv)
		})

		vv3 := []byte{33, 44, 55}
		v.Store(vv3)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(vv3, vv)
	})
}

func TestPointer(t *testing.T) {
	t.Run("Pointer[string]", func(t *testing.T) {
		assert := assert.New(t)

		v := &atomic.Pointer[string]{}

		vv := v.Load()
		assert.Nil(vv)

		xa := "a"
		vv = v.Swap(&xa)
		assert.Nil(vv)
		vv = v.Load()
		assert.NotNil(vv)
		assert.Equal("a", *vv)

		xb := "b"
		vv = v.Swap(&xb)
		assert.NotNil(vv)
		assert.Equal("a", *vv)

		vv = v.Load()
		assert.NotNil(vv)
		assert.Equal("b", *vv)

		xc := "c"
		assert.False(v.CompareAndSwap(&xa, &xc))
		assert.True(v.CompareAndSwap(&xb, &xc))

		vv = v.Load()
		assert.NotNil(vv)
		assert.Equal("c", *vv)

		v.Store(&xa)
		vv = v.Load()
		assert.NotNil(vv)
		assert.Equal("a", *vv)
	})

	t.Run("Value[int]", func(t *testing.T) {
		t.Skip()
		assert := assert.New(t)

		var v Value[int]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Equal(0, vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		vv, ok = v.Swap(1)
		assert.False(ok)
		assert.Equal(0, vv)
		assert.Equal(1, v.MustLoad())

		vv, ok = v.Swap(2)
		assert.True(ok)
		assert.Equal(1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(2, vv)

		assert.False(v.CompareAndSwap(3, 3))
		assert.True(v.CompareAndSwap(2, 3))

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(3, vv)

		v.Store(33)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(33, vv)
	})

	t.Run("Value[[4]byte]", func(t *testing.T) {
		t.Skip()
		assert := assert.New(t)

		var v Value[[4]byte]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Equal([4]byte{}, vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		v1 := [4]byte{1, 2, 3}
		vv, ok = v.Swap(v1)
		assert.False(ok)
		assert.Equal([4]byte{}, vv)
		assert.Equal(v1, v.MustLoad())

		v2 := [4]byte{2, 3, 4}
		vv, ok = v.Swap(v2)
		assert.True(ok)
		assert.Equal(v1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v2, vv)

		v3 := [4]byte{3, 4, 5}
		assert.False(v.CompareAndSwap(v1, v3))
		assert.True(v.CompareAndSwap(v2, v3))

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v3, vv)

		vv3 := [4]byte{33, 44, 55}
		v.Store(vv3)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(vv3, vv)
	})

	t.Run("Value[[]byte]", func(t *testing.T) {
		t.Skip()
		assert := assert.New(t)

		var v Value[[]byte]

		vv, ok := v.Load()
		assert.False(ok)
		assert.Nil(vv)
		assert.Panics(func() {
			v.MustLoad()
		})

		v1 := []byte{1, 2, 3}
		vv, ok = v.Swap(v1)
		assert.False(ok)
		assert.Nil(vv)
		assert.Equal(v1, v.MustLoad())

		v2 := []byte{2, 3, 4}
		vv, ok = v.Swap(v2)
		assert.True(ok)
		assert.Equal(v1, vv)

		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(v2, vv)

		assert.Panics(func() {
			v3 := []byte{3, 4, 5}
			assert.False(v.CompareAndSwap(v1, v3))
			assert.True(v.CompareAndSwap(v2, v3))

			vv, ok = v.Load()
			assert.True(ok)
			assert.Equal(v3, vv)
		})

		vv3 := []byte{33, 44, 55}
		v.Store(vv3)
		vv, ok = v.Load()
		assert.True(ok)
		assert.Equal(vv3, vv)
	})
}

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Pool is no-op under race detector, so all these tests do not work.
//
//go:build !race

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	t.Run("Pool[string]", func(t *testing.T) {
		assert := assert.New(t)

		var p Pool[string]

		pv, ok := p.Get()
		assert.False(ok)
		assert.Equal("", pv)

		p.Put("a")
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal("a", pv)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Equal("", pv)

		p.Put("b")
		p.Put("c")
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal("b", pv)

		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal("c", pv)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Equal("", pv)

		assert.Panics(func() {
			p.MustGet()
		})
		p.New = func() string { return "x" }
		assert.Equal("x", p.MustGet())
	})

	t.Run("Pool[int]", func(t *testing.T) {
		assert := assert.New(t)

		var p Pool[int]

		pv, ok := p.Get()
		assert.False(ok)
		assert.Equal(0, pv)

		p.Put(9)
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(9, pv)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Equal(0, pv)

		p.Put(99)
		p.Put(999)
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(99, pv)

		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(999, pv)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Equal(0, pv)

		assert.Panics(func() {
			p.MustGet()
		})
		p.New = func() int { return 88 }
		assert.Equal(88, p.MustGet())
	})

	t.Run("Pool[[][4]byte]", func(t *testing.T) {
		assert := assert.New(t)

		var p Pool[[][4]byte]

		pv, ok := p.Get()
		assert.False(ok)
		assert.Nil(pv)

		v1 := [][4]byte{{1, 2, 3, 4}}
		p.Put(v1)
		pv, ok = p.Get()
		assert.True(ok)
		require.Equal(t, 1, len(pv))
		assert.Equal(v1[0], pv[0])

		pv, ok = p.Get()
		assert.False(ok)
		assert.Nil(pv)

		v2 := [][4]byte{{255, 2, 3, 4}}
		v3 := [][4]byte{{254, 2, 3, 4}}

		p.Put(v2)
		p.Put(v3)
		pv, ok = p.Get()
		assert.True(ok)
		require.Equal(t, 1, len(pv))
		assert.Equal(v2[0], pv[0])

		pv, ok = p.Get()
		assert.True(ok)
		require.Equal(t, 1, len(pv))
		assert.Equal(v3[0], pv[0])

		pv, ok = p.Get()
		assert.False(ok)
		assert.Nil(pv)

		assert.Panics(func() {
			p.MustGet()
		})
		p.New = func() [][4]byte { return [][4]byte{{1, 1, 1, 1}} }
		pv = p.MustGet()
		assert.Equal([4]byte{1, 1, 1, 1}, pv[0])
	})

	t.Run("Pool[*Str]", func(t *testing.T) {
		assert := assert.New(t)

		type Str struct {
			v string
		}

		var p Pool[*Str]

		pv, ok := p.Get()
		assert.False(ok)
		assert.Nil(pv)

		v1 := &Str{"a"}
		p.Put(v1)
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(v1.v, pv.v)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Nil(pv)

		v2 := &Str{"b"}
		v3 := &Str{"c"}
		p.Put(v2)
		p.Put(v3)
		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(v2.v, pv.v)

		pv, ok = p.Get()
		assert.True(ok)
		assert.Equal(v3.v, pv.v)

		pv, ok = p.Get()
		assert.False(ok)
		assert.Nil(pv)

		assert.Panics(func() {
			p.MustGet()
		})
		p.New = func() *Str { return &Str{"x"} }
		pv = p.MustGet()
		assert.Equal("x", pv.v)
	})
}

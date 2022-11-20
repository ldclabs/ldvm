// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("Map[int, string]", func(t *testing.T) {
		assert := assert.New(t)

		var m Map[int, string]

		mv, ok := m.Load(0)
		assert.False(ok)
		assert.Equal("", mv)

		mv, ok = m.LoadOrStore(0, "a")
		assert.False(ok)
		assert.Equal("a", mv)

		mv, ok = m.LoadOrStore(0, "b")
		assert.True(ok)
		assert.Equal("a", mv)

		mv, ok = m.LoadAndDelete(0)
		assert.True(ok)
		assert.Equal("a", mv)

		mv, ok = m.LoadAndDelete(0)
		assert.False(ok)
		assert.Equal("", mv)

		m.Store(1, "b")
		mv, ok = m.Load(1)
		assert.True(ok)
		assert.Equal("b", mv)

		m.Delete(1)
		mv, ok = m.Load(1)
		assert.False(ok)
		assert.Equal("", mv)

		m.Store(0, "a")
		m.Store(1, "b")
		m.Range(func(k int, v string) bool {
			switch k {
			case 0:
				assert.Equal("a", v)
			case 1:
				assert.Equal("b", v)
			default:
				assert.Failf("range failed", "unexpected key %d", k)
			}
			m.Delete(k)
			return true
		})

		mv, ok = m.Load(0)
		assert.False(ok)
		assert.Equal("", mv)

		mv, ok = m.Load(1)
		assert.False(ok)
		assert.Equal("", mv)
	})

	t.Run("Map[string, *Str]", func(t *testing.T) {
		assert := assert.New(t)

		type Str struct {
			v string
		}

		var m Map[string, *Str]

		mv, ok := m.Load("a")
		assert.False(ok)
		assert.True(mv == nil)

		v1 := &Str{"a"}
		mv, ok = m.LoadOrStore("a", v1)
		assert.False(ok)
		assert.Equal(v1, mv)

		v2 := &Str{"b"}
		mv, ok = m.LoadOrStore("a", v2)
		assert.True(ok)
		assert.Equal(v1, mv)

		mv, ok = m.LoadAndDelete("a")
		assert.True(ok)
		assert.Equal(v1, mv)

		mv, ok = m.LoadAndDelete("a")
		assert.False(ok)
		assert.True(mv == nil)

		m.Store("b", v2)
		mv, ok = m.Load("b")
		assert.True(ok)
		assert.Equal(v2, mv)

		m.Delete("b")
		mv, ok = m.Load("b")
		assert.False(ok)
		assert.True(mv == nil)

		m.Store("a", v1)
		m.Store("b", nil)
		m.Range(func(k string, v *Str) bool {
			switch k {
			case "a":
				assert.Equal(v1, v)
			case "b":
				assert.True(v == nil)
			default:
				assert.Failf("range failed", "unexpected key %s", k)
			}
			m.Delete(k)
			return true
		})

		mv, ok = m.Load("a")
		assert.False(ok)
		assert.True(mv == nil)

		mv, ok = m.Load("b")
		assert.False(ok)
		assert.True(mv == nil)

		mv, ok = m.LoadOrStore("a", nil)
		assert.False(ok)
		assert.True(mv == nil)

		mv, ok = m.LoadOrStore("a", v1)
		assert.True(ok)
		assert.True(mv == nil)
	})
}

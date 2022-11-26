// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ldclabs/ldvm/util/encoding"
)

type MyLog struct{ Value }

func (l *MyLog) Valid() bool {
	return l != nil && l.Is(Vmap)
}

func TestLog(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		assert := assert.New(t)

		v0 := Log{String("hello")}
		ctx := CtxWith(context.Background(), &v0)

		v1 := CtxValue[Log](context.Background())
		assert.True(v1 == nil)

		v1 = CtxValue[Log](ctx)
		require.NotNil(t, v1)
		assert.True(v1.Is(Vstring))
		assert.Equal(v0, *v1)

		*v1 = Log{String("world")}
		assert.Equal(v0, *v1)

		v2 := CtxValue[Log](ctx)
		assert.Equal(v0, *v2)

		data := encoding.MustMarshalJSON(v2)
		assert.Equal(`"world"`, string(data))

		called := false
		DoIfCtxValueValid(ctx, func(v *Log) { called = true })
		assert.False(called)

		v3 := CtxValue[MyLog](ctx)
		assert.True(v3 == nil)
	})

	t.Run("List", func(t *testing.T) {
		assert := assert.New(t)

		list := List{Bool(true), Int64(-1), String("hello")}
		log := Log{list.Value()}
		ctx := CtxWith(context.Background(), &log)

		v1 := CtxValue[Log](ctx)
		assert.Equal(list, v1.List())

		l := append(v1.List(), String("world"))
		assert.NotEqual(list, l)

		v1.Value = l.Value()
		v2 := CtxValue[Log](ctx)
		assert.Equal(l, v2.List())

		data := encoding.MustMarshalJSON(v2.List())
		assert.Equal(`[true,-1,"hello","world"]`, string(data))

		called := false
		DoIfCtxValueValid(ctx, func(v *Log) { called = true })
		assert.False(called)

		v3 := CtxValue[MyLog](ctx)
		assert.True(v3 == nil)
	})

	t.Run("Map", func(t *testing.T) {
		assert := assert.New(t)

		m := Map{"a": Bool(true), "b": BigInt(big.NewInt(123)), "c": String("hello")}
		log := Log{m.Value()}
		ctx := CtxWith(context.Background(), &log)

		v1 := CtxValue[Log](ctx)
		assert.Equal(m, v1.Map())

		v1.Map()["d"] = String("world")
		assert.Equal(m, v1.Map())

		v2 := CtxValue[Log](ctx)
		assert.Equal(m, v2.Map())

		data := encoding.MustMarshalJSON(v2.Map())
		assert.Equal(`{"a":true,"b":123,"c":"hello","d":"world"}`, string(data))

		list := List{Bool(true), Bool(false)}
		v2.Set("list", list.Value())
		data = encoding.MustMarshalJSON(m)
		assert.Equal(`{"a":true,"b":123,"c":"hello","d":"world","list":[true,false]}`, string(data))

		called := false
		DoIfCtxValueValid(ctx, func(v *Log) { called = true })
		assert.True(called)

		v3 := CtxValue[MyLog](ctx)
		assert.True(v3 == nil)

		var mv Map
		ctx = CtxWith(context.Background(), &Log{Value: mv.Value()})
		called = false
		DoIfCtxValueValid(ctx, func(v *Log) { v.Set("key", Bool(true)) })
		assert.True(mv.Has("key"))
		assert.True(mv["key"].Bool())
	})
}

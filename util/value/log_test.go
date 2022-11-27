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

		v0 := Log{Value: String("hello")}
		ctx := CtxWith(context.Background(), &v0)

		v1 := CtxValue[Log](context.Background())
		assert.True(v1 == nil)

		v1 = CtxValue[Log](ctx)
		require.NotNil(t, v1)
		assert.True(v1.Is(Vstring))
		assert.Equal(v0, *v1)

		*v1 = Log{Value: String("world")}
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
		log := Log{Value: list.ToValue()}
		ctx := CtxWith(context.Background(), &log)

		v1 := CtxValue[Log](ctx)
		assert.Equal(list, v1.ToList())

		l := append(v1.ToList(), String("world"))
		assert.NotEqual(list, l)

		v1.Value = l.ToValue()
		v2 := CtxValue[Log](ctx)
		assert.Equal(l, v2.ToList())

		data := encoding.MustMarshalJSON(v2.ToList())
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
		log := Log{Value: m.ToValue()}
		ctx := CtxWith(context.Background(), &log)

		v1 := CtxValue[Log](ctx)
		assert.Equal(m, v1.ToMap())

		v1.ToMap()["d"] = String("world")
		assert.Equal(m, v1.ToMap())

		v2 := CtxValue[Log](ctx)
		assert.Equal(m, v2.ToMap())

		data := encoding.MustMarshalJSON(v2.ToMap())
		assert.Equal(`{"a":true,"b":123,"c":"hello","d":"world"}`, string(data))

		list := List{Bool(true), Bool(false)}
		v2.Set("list", list.ToValue())
		data = encoding.MustMarshalJSON(m)
		assert.Equal(`{"a":true,"b":123,"c":"hello","d":"world","list":[true,false]}`, string(data))

		called := false
		DoIfCtxValueValid(ctx, func(v *Log) { called = true })
		assert.True(called)

		v3 := CtxValue[MyLog](ctx)
		assert.True(v3 == nil)

		var mv Map
		log = Log{Value: mv.ToValue()}
		ctx = CtxWith(context.Background(), &log)
		called = false
		DoIfCtxValueValid(ctx, func(v *Log) { v.Set("key", Bool(true)) })
		assert.True(mv.Has("key"))
		assert.True(mv["key"].ToBool())

		assert.True(log.ToMap().Has("key"))
		assert.True(log.ToMap()["key"].ToBool())
	})
}

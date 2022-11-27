// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ldclabs/ldvm/util/encoding"
	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	assert := assert.New(t)

	var v Value
	assert.Equal(Invalid, v.Kind())
	assert.True(v.Is(Invalid))
	assert.False(v.Is(Vmap))
	assert.True(v.ToList() == nil)
	assert.True(v.ToMap() == nil)

	var pv *Value
	assert.True(pv == nil)
	assert.Equal(Invalid, pv.Kind())
	assert.False(pv.Is(Invalid))
	assert.False(pv.Is(Vmap))
	assert.Nil(pv.ToList())
	assert.True(pv.ToMap() == nil)

	for _, tc := range []struct {
		kind    Kind
		value   Value
		invalid Value
		i       any
		goStr   string
		jsonStr string
	}{
		{
			kind:    Invalid,
			value:   Value{},
			invalid: Bool(false),
			i:       nil,
			goStr:   `<nil>`,
			jsonStr: `null`,
		},
		{
			kind:    Vbool,
			value:   Bool(true),
			invalid: Value{},
			i:       true,
			goStr:   `true`,
			jsonStr: `true`,
		},
		{
			kind:    Vint64,
			value:   Int64(-99),
			invalid: Value{},
			i:       int64(-99),
			goStr:   `-99`,
			jsonStr: `-99`,
		},
		{
			kind:    Vfloat64,
			value:   Float64(999.9),
			invalid: Value{},
			i:       float64(999.9),
			goStr:   `999.9`,
			jsonStr: `999.9`,
		},
		{
			kind:    Vstring,
			value:   String("hello"),
			invalid: Value{},
			i:       "hello",
			goStr:   `"hello"`,
			jsonStr: `"hello"`,
		},
		{
			kind:    VbigInt,
			value:   BigInt(new(big.Int).Mul(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1000))),
			invalid: Value{},
			i:       new(big.Int).Mul(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1000)),
			goStr:   `18446744073709551615000`,
			jsonStr: `18446744073709551615000`,
		},
		{
			kind:    Vtime,
			value:   Time(time.Date(2022, time.October, 24, 0, 0, 0, 0, time.UTC)),
			invalid: Value{},
			i:       time.Date(2022, time.October, 24, 0, 0, 0, 0, time.UTC),
			goStr:   `time.Date(2022, time.October, 24, 0, 0, 0, 0, time.UTC)`,
			jsonStr: `"2022-10-24T00:00:00Z"`,
		},
		{
			kind:    Vlist,
			value:   NewList(0),
			invalid: Value{},
			i:       List{},
			goStr:   `value.List{}`,
			jsonStr: `[]`,
		},
		{
			kind:    Vmap,
			value:   NewMap(0),
			invalid: Value{},
			i:       Map{},
			goStr:   `value.Map{}`,
			jsonStr: `{}`,
		},
	} {
		assert.False(tc.invalid.Is(tc.kind))
		assert.True(tc.value.Is(tc.kind))
		assert.Equal(tc.kind, tc.value.Kind())
		assert.Equal(tc.value, *tc.value.Ptr())
		assert.Equal(tc.i, tc.value.ToAny())
		assert.Equal(tc.goStr, tc.value.GoString())

		assert.Equal(tc.jsonStr, string(encoding.MustMarshalJSON(tc.value)))
		assert.Equal(encoding.MustMarshalCBOR(tc.i), encoding.MustMarshalCBOR(tc.value))

		switch tc.kind {
		case Invalid:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vbool:
			assert.Equal(tc.i.(bool), tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vint64:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(tc.i.(int64), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vfloat64:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(tc.i.(float64), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vstring:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal(tc.i.(string), tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case VbigInt:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(tc.i.(*big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vtime:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(tc.i.(time.Time), tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vlist:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal(tc.i.(List), tc.value.ToList())
			assert.Equal((Map)(nil), tc.value.ToMap())

		case Vmap:
			assert.Equal(false, tc.value.ToBool())
			assert.Equal(int64(0), tc.value.ToInt64())
			assert.Equal(float64(0), tc.value.ToFloat64())
			assert.Equal("", tc.value.ToString())
			assert.Equal(new(big.Int), tc.value.ToBigInt())
			assert.Equal(time.Time{}, tc.value.ToTime())
			assert.Equal((List)(nil), tc.value.ToList())
			assert.Equal(tc.i.(Map), tc.value.ToMap())
		}
	}
}

func TestList(t *testing.T) {
	assert := assert.New(t)

	list := List{Bool(true), Int64(-1), String("hello")}
	ilist := []interface{}{true, -1, "hello"}
	jsonStr := string(encoding.MustMarshalJSON(ilist))

	assert.Equal(`[true,-1,"hello"]`, jsonStr)
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(list)))
	v := list.ToValue()
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v)))

	v.Append(String("world"))
	jsonStr = `[true,-1,"hello","world"]`
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v)))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v.ToList())))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(list)))

	data := encoding.MustMarshalCBOR([]interface{}{
		true, -1, "hello", "world",
	})
	assert.Equal(data, encoding.MustMarshalCBOR(v))
	assert.Equal(data, encoding.MustMarshalCBOR(v.ToList()))
	assert.Equal(data, encoding.MustMarshalCBOR(list))

	var l List
	assert.True(l == nil)
	assert.Equal(`null`, string(encoding.MustMarshalJSON(l)))
	v2 := l.ToValue()
	v2.Append(v.ToList()...)
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v2)))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(l)))

	v = Value{}
	assert.Panics(func() {
		v.Append(String("world"))
	})
}

func TestMap(t *testing.T) {
	assert := assert.New(t)

	bigInt := new(big.Int).Mul(new(big.Int).SetUint64(math.MaxUint64), big.NewInt(1000))
	im := map[string]interface{}{"a": true, "b": bigInt, "c": "hello"}
	m := Map{"a": Bool(true), "b": BigInt(bigInt), "c": String("hello")}

	jsonStr := string(encoding.MustMarshalJSON(im))

	assert.Equal(`{"a":true,"b":18446744073709551615000,"c":"hello"}`, jsonStr)
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(m)))
	v := m.ToValue()
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v)))

	v.Merge(Map{"d": String("world")})
	jsonStr = `{"a":true,"b":18446744073709551615000,"c":"hello","d":"world"}`
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v)))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v.ToMap())))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(m)))

	data := encoding.MustMarshalCBOR(map[string]interface{}{
		"a": true, "b": bigInt, "c": "hello", "d": "world",
	})
	assert.Equal(data, encoding.MustMarshalCBOR(v))
	assert.Equal(data, encoding.MustMarshalCBOR(v.ToMap()))
	assert.Equal(data, encoding.MustMarshalCBOR(m))

	var m2 Map
	assert.True(m2 == nil)
	v2 := m2.ToValue()
	v2.Merge(v.ToMap())
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v)))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(v2)))
	assert.Equal(jsonStr, string(encoding.MustMarshalJSON(m2)))

	assert.False(m.Has("aaa"))
	m["aaa"] = Int64(1)
	assert.True(m.Has("aaa"))
	assert.Equal([]string{"a", "aaa", "b", "c", "d"}, m.Keys())

	m["aa"] = Int64(2)
	assert.Equal([]string{"a", "aa", "aaa", "b", "c", "d"}, m.Keys())
	assert.Equal(List{Bool(true), Int64(2), Int64(1), BigInt(bigInt), String("hello"), String("world")}, m.ToValues())

	v = Value{}
	assert.Panics(func() {
		v.Append(String("world"))
	})
}

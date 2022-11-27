// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ldclabs/ldvm/util/encoding"
)

type Kind int

const (
	// Invalid is used for a Value without data.
	Invalid Kind = iota
	Vlist        // Vlist is a []Value Value.
	Vmap         // Vmap is a map[string]Value Value.
	Vbool
	Vint64
	Vfloat64
	Vstring
	Vtime   // Vtime is a time.Time Value.
	VbigInt // VbigInt is a *big.Int Value.
)

// String returns the name of k.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return kindNames[0]
}

var kindNames = []string{
	Invalid:  "invalid",
	Vlist:    "list",
	Vmap:     "map",
	Vbool:    "bool",
	Vint64:   "int64",
	Vfloat64: "float64",
	Vstring:  "string",
	Vtime:    "time.Time",
	VbigInt:  "*big.Int",
}

// Value represents a value of a set of types.
type Value struct {
	kind Kind
	v    any
}

// List represents a list of Values.
type List []Value

// Map represents a map of string to Value.
type Map map[string]Value

// NewList creates a new List Value with given capacity.
func NewList(cap int) Value {
	l := make(List, 0, cap)
	return Value{kind: Vlist, v: &l}
}

// NewMap creates a new Map Value with given capacity.
func NewMap(cap int) Value {
	return Value{kind: Vmap, v: make(Map, cap)}
}

// Bool creates a bool Value.
func Bool(v bool) Value {
	return Value{kind: Vbool, v: v}
}

// Int creates a int64 Value.
func Int(v int) Value {
	return Value{kind: Vint64, v: int64(v)}
}

// Int64 creates a int64 Value.
func Int64(v int64) Value {
	return Value{kind: Vint64, v: v}
}

// Float64 creates a float64 Value.
func Float64(v float64) Value {
	return Value{kind: Vfloat64, v: v}
}

// String creates a string Value.
func String(v string) Value {
	return Value{kind: Vstring, v: v}
}

// BigInt creates a *big.Int Value.
func BigInt(v *big.Int) Value {
	if v == nil {
		v = new(big.Int)
	}
	return Value{kind: VbigInt, v: v}
}

// Time creates a time.Time Value.
func Time(v time.Time) Value {
	return Value{kind: Vtime, v: v}
}

// Ptr returns the Value as a pointer.
func (v Value) Ptr() *Value {
	return &v
}

// Type returns a type of the Value.
func (v *Value) Kind() Kind {
	if v == nil {
		return Invalid
	}
	return v.kind
}

// Is returns true if the Value is of the given Kind.
func (v *Value) Is(k Kind) bool {
	if v == nil {
		return false
	}
	return v.kind == k
}

// ToList returns the List value.
// If Kind is not List, it returns nil.
func (v *Value) ToList() List {
	if v == nil {
		return nil
	}
	x, _ := v.v.(*List)
	if x == nil {
		return nil
	}
	return *x
}

// ToMap returns the Map value.
// If Kind is not Map, it returns nil.
func (v *Value) ToMap() Map {
	if v == nil {
		return nil
	}
	x, _ := v.v.(Map)
	return x
}

// ToBool returns the bool value.
// If Kind is not Bool, it returns false.
func (v Value) ToBool() bool {
	x, _ := v.v.(bool)
	return x
}

// ToInt64 returns the int64 value.
// If Kind is not Int64, it returns 0.
func (v Value) ToInt64() int64 {
	x, _ := v.v.(int64)
	return x
}

// ToFloat64 returns the float64 value.
// If Kind is not Float64, it returns 0.
func (v Value) ToFloat64() float64 {
	x, _ := v.v.(float64)
	return x
}

// ToString returns the string value.
// If Kind is not String, it returns "".
func (v Value) ToString() string {
	x, _ := v.v.(string)
	return x
}

// ToBigInt returns the *big.Int value.
// If Kind is not BigInt, it returns &big.Int{}.
func (v Value) ToBigInt() *big.Int {
	if x, ok := v.v.(*big.Int); ok {
		return x
	}
	return &big.Int{}
}

// ToArray returns the Array value.
// If Kind is not Array, it returns time.Time{}.
func (v Value) ToTime() time.Time {
	x, _ := v.v.(time.Time)
	return x
}

// Append appends more Values to the List Value.
// If the Value is not a List, it will panic.
func (v *Value) Append(vs ...Value) {
	if !v.Is(Vlist) {
		panic("Value: cannot append to non-list value")
	}

	*(v.v.(*List)) = append(*(v.v.(*List)), vs...)
}

// Merge merges the given Set into the current Value.
// If the Value is not a Set, it will panic.
func (v *Value) Set(key string, value Value) {
	if !v.Is(Vmap) {
		panic("value: cannot set (key, value) on non-map value")
	}

	v.v.(Map)[key] = value
}

// Merge merges the given Set into the current Value.
// If the Value is not a Set, it will panic.
func (v *Value) Merge(m Map) {
	if !v.Is(Vmap) {
		panic("value: cannot merge into non-map value")
	}

	vm := v.v.(Map)
	for k, e := range m {
		vm[k] = e
	}
}

// ToAny returns the Value as any.
func (v Value) ToAny() any {
	if v.kind == Vlist {
		return v.ToList()
	}
	return v.v
}

// GoString returns a string representation of Value's data.
func (v Value) GoString() string {
	return fmt.Sprintf("%#v", v.ToAny())
}

// MarshalJSON returns the JSON encoding of the Value.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ToAny())
}

// MarshalCBOR returns the CBOR encoding of the Value.
func (v Value) MarshalCBOR() ([]byte, error) {
	return encoding.MarshalCBOR(v.ToAny())
}

// ToValue returns then List as a Value.
func (v *List) ToValue() Value {
	if *v == nil {
		*v = make(List, 0)
	}
	return Value{kind: Vlist, v: v}
}

// ToValue returns the Map as a Value.
func (v *Map) ToValue() Value {
	if *v == nil {
		*v = make(Map)
	}
	return Value{kind: Vmap, v: *v}
}

// Has returns true if the given key exists in the Set.
func (v Map) Has(key string) bool {
	_, ok := v[key]
	return ok
}

// Keys returns a list of sorted keys in the Set.
func (v Map) Keys() []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToValues returns a list of Values in the Set, which are sorted by keys.
func (v Map) ToValues() List {
	values := make(List, len(v))
	for i, k := range v.Keys() {
		values[i] = v[k]
	}
	return values
}

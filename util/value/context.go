// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package value

import (
	"context"
	"reflect"
)

// CtxWith returns a new context.Context with the Value stored in it.
func CtxWith[T any](parent context.Context, v *T) context.Context {
	return context.WithValue(parent, reflect.TypeOf((*T)(nil)), v)
}

// CtxValue returns the Value stored in the context.Context, or nil if not stored.
func CtxValue[T any](ctx context.Context) *T {
	v, _ := ctx.Value(reflect.TypeOf((*T)(nil))).(*T)
	return v
}

// DoIfCtxValueValid calls the function fn with the Value stored in the context.Context.
// If the Value is not valid, the function fn will not be called.
func DoIfCtxValueValid[T any, TI isValid[T]](ctx context.Context, fn func(v *T)) {
	if v := CtxValue[T](ctx); TI(v).Valid() {
		fn(v)
	}
}

type isValid[T any] interface {
	*T
	Valid() bool
}

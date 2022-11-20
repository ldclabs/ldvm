// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import "sync/atomic"

// Value is a generic version of atomic.Value
type Value[T any] struct {
	v atomic.Value
}

// CompareAndSwap executes the compare-and-swap operation for the Value[T].
// It will panic if T is not an comparable type.
func (v *Value[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return v.v.CompareAndSwap(old, new)
}

// Load returns the value set by the most recent Store.
// It returns (zero value of T, false) if there has been no call to Store for this Value.
func (v *Value[T]) Load() (val T, ok bool) {
	val, ok = v.v.Load().(T)
	return
}

// Load returns the value set by the most recent Store.
// It will panic  if there has been no call to Store for this Value.
func (v *Value[T]) MustLoad() T {
	val, ok := v.Load()
	if !ok {
		panic("Value is empty")
	}

	return val
}

// Store sets the value of the Value[T] to x.
func (v *Value[T]) Store(val T) {
	v.v.Store(val)
}

// Swap stores new into Value[T] and returns the previous value.
// It returns (zero value of T, false) if the Value[T] is empty.
func (v *Value[T]) Swap(new T) (val T, ok bool) {
	val, ok = v.v.Swap(new).(T)
	return
}

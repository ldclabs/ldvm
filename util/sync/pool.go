// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import (
	ss "sync"
)

// Pool is a generic version of sync.Pool
// More info sees https://pkg.go.dev/sync#Pool
type Pool[T any] struct {
	p   ss.Pool
	New func() T
}

// Get selects an arbitrary item from the pool, removes it from the pool, and returns it to the caller.
// The ok result indicates whether value was found in the pool.
// It returns (zero value of T, false) if there has nothing in the pool.
func (p *Pool[T]) Get() (T, bool) {
	v := p.p.Get()
	t, ok := v.(T)
	return t, ok
}

// Get selects an arbitrary item from the pool, removes it from the pool, and returns it to the caller.
// It will call Pool.New to create value  if there has nothing in the pool.
// It will panic if Pool.New is nil.
func (p *Pool[T]) MustGet() T {
	t, ok := p.Get()
	if !ok {
		if p.New == nil {
			panic("Pool.New is nil")
		}

		t = p.New()
	}
	return t
}

// Put adds v to the pool.
func (p *Pool[T]) Put(v T) {
	p.p.Put(v)
}

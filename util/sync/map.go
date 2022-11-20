// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import (
	ss "sync"
)

// Map is a generic version of sync.Map
// More info sees https://pkg.go.dev/sync#Map
type Map[K any, V any] struct {
	m ss.Map
}

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Load returns the value stored in the map for a key, or zero value of T if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	v, ok := m.m.Load(key)
	if !ok {
		return
	}

	value, ok = v.(V)
	return
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	v, ok := m.m.LoadAndDelete(key)
	if !ok {
		return
	}

	value, loaded = v.(V)
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	if v, ok := m.m.LoadOrStore(key, value); ok {
		actual, loaded = v.(V)
		return
	}

	return value, false
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
// More info sees https://pkg.go.dev/sync#Map.Range
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

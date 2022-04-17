// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"errors"
	"time"

	"github.com/dgraph-io/ristretto"
)

var _ ObjectCacher = (*Cacher)(nil)

type Cacher struct {
	cache *ristretto.Cache
	ttl   time.Duration
	new   func() Unmarshaler
}

type Unmarshaler interface {
	Unmarshal([]byte) error
}

type RawObject []byte

func (r *RawObject) Unmarshal(data []byte) error {
	if r == nil {
		return errors.New("RawObject: Unmarshal on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

func NewCacher(capacity int64, maxSize int64, ttl time.Duration, new func() Unmarshaler) *Cacher {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: capacity * 10,
		MaxCost:     maxSize,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return &Cacher{c, ttl, new}
}

func (c *Cacher) Get(key []byte) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *Cacher) Set(key []byte, value interface{}, cost int64) {
	c.cache.SetWithTTL(key, value, cost, c.ttl)
}

func (c *Cacher) Unmarshal(data []byte) (interface{}, error) {
	if c.new == nil {
		return nil, errors.New("no new function to create object")
	}
	obj := c.new()
	if err := obj.Unmarshal(data); err != nil {
		return nil, err
	}
	return obj, nil
}

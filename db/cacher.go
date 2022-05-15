// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"errors"

	"github.com/mailgun/holster/v4/collections"
)

var (
	_ ObjectCacher = (*Cacher)(nil)
	_ Objecter     = (*RawObject)(nil)
)

type Cacher struct {
	cache *collections.TTLMap
	ttl   int
	new   func() Objecter
}

type Objecter interface {
	Unmarshal([]byte) error
}

type Verifier interface {
	SyntacticVerify() error
}

type RawObject []byte

func (r *RawObject) Unmarshal(data []byte) error {
	if r == nil {
		return errors.New("RawObject.Unmarshal failed: nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

func NewCacher(capacity, ttlsecs int, fn func() Objecter) *Cacher {
	return &Cacher{collections.NewTTLMap(int(capacity)), ttlsecs, fn}
}

func (c *Cacher) GetObject(key []byte) (interface{}, bool) {
	return c.cache.Get(string(key))
}

func (c *Cacher) SetObject(key []byte, value interface{}) error {
	return c.cache.Set(string(key), value, c.ttl)
}

func (c *Cacher) UnmarshalObject(data []byte) (interface{}, error) {
	if c.new == nil {
		return nil, errors.New("Cacher.Unmarshal failed: no function to create object")
	}

	obj := c.new()
	if err := obj.Unmarshal(data); err != nil {
		return nil, err
	}
	if v, ok := obj.(Verifier); ok {
		if err := v.SyntacticVerify(); err != nil {
			return nil, err
		}
	}
	return obj, nil
}

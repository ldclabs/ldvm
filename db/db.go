// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package db

import (
	"sync"

	"github.com/ava-labs/avalanchego/database"
)

var _ database.KeyValueReaderWriterDeleter = (*PrefixDB)(nil)

type ObjectCacher interface {
	GetObject(key []byte) (interface{}, bool)
	SetObject(key []byte, value interface{}) error
	UnmarshalObject(data []byte) (interface{}, error)
}

type PrefixDB struct {
	mu        sync.Mutex
	db        database.Database
	prefixLen int
	keyBuf    []byte
}

func NewPrefixDB(db database.Database, prefix []byte, keyBufSize int) *PrefixDB {
	p := &PrefixDB{
		db:        db,
		prefixLen: len(prefix),
		keyBuf:    make([]byte, keyBufSize),
	}
	copy(p.keyBuf, prefix)
	return p
}

func (p *PrefixDB) With(prefix []byte) *PrefixDB {
	np := &PrefixDB{
		db:        p.db,
		prefixLen: p.prefixLen + len(prefix),
		keyBuf:    make([]byte, len(p.keyBuf)),
	}
	copy(np.keyBuf, p.keyBuf[:p.prefixLen])
	copy(np.keyBuf[p.prefixLen:], prefix)
	return np
}

func (p *PrefixDB) Has(key []byte) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := copy(p.keyBuf[p.prefixLen:], key)
	return p.db.Has(p.keyBuf[:n+p.prefixLen])
}

func (p *PrefixDB) Get(key []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.get(key)
}

func (p *PrefixDB) get(key []byte) ([]byte, error) {
	n := copy(p.keyBuf[p.prefixLen:], key)
	return p.db.Get(p.keyBuf[:n+p.prefixLen])
}

func (p *PrefixDB) LoadObject(key []byte, c ObjectCacher) (interface{}, error) {
	if v, ok := c.GetObject(key); ok {
		return v, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	// try cache again
	if v, ok := c.GetObject(key); ok {
		return v, nil
	}
	data, err := p.get(key)
	if err != nil {
		return nil, err
	}
	v, err := c.UnmarshalObject(data)
	if err != nil {
		return nil, err
	}
	if err = c.SetObject(key, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (p *PrefixDB) Put(key, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := copy(p.keyBuf[p.prefixLen:], key)
	return p.db.Put(p.keyBuf[:n+p.prefixLen], value)
}

func (p *PrefixDB) Delete(key []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := copy(p.keyBuf[p.prefixLen:], key)
	return p.db.Delete(p.keyBuf[:n+p.prefixLen])
}

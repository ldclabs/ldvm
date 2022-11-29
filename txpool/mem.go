// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txpool

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ldclabs/ldvm/ids"
)

// MemStorage is in-memory POS for testing.
type MemStorage struct {
	mu         sync.Mutex
	id         uint64
	expiration time.Duration
	objects    map[string]map[ids.ID32]*entity
}

func NewMemStorage(expiration time.Duration) POS {
	if expiration <= 0 {
		expiration = time.Second * 15
	}

	return &MemStorage{
		expiration: expiration,
		objects:    make(map[string]map[ids.ID32]*entity),
	}
}

type entity struct {
	id       uint64
	expireAt int64 // UnixNano
	obj      *Object
}

func (m *MemStorage) GetObject(ctx context.Context, bucket string, hash ids.ID32) (*Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	objs, ok := m.objects[bucket]
	if !ok {
		return nil, fmt.Errorf("bucket %q not found", bucket)
	}
	ent, ok := objs[hash]
	if !ok {
		return nil, fmt.Errorf("object %q not found", hash.String())
	}

	now := time.Now().UnixNano()
	if ent.expireAt > 0 && ent.expireAt <= now {
		delete(objs, hash)
		return nil, fmt.Errorf("object %q is expired", hash.String())
	}

	obj := &Object{Raw: make([]byte, len(ent.obj.Raw)), Height: ent.obj.Height}
	copy(obj.Raw, ent.obj.Raw)
	return obj, nil
}

func (m *MemStorage) PutObject(ctx context.Context, bucket string, objectRaw []byte) error {
	obj := &Object{
		Raw:    objectRaw,
		Height: -1,
	}
	hash := obj.Hash()

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.objects[bucket]; !ok {
		m.objects[bucket] = make(map[ids.ID32]*entity)
	}

	ent, ok := m.objects[bucket][hash]
	if ok && (ent.expireAt == 0 || ent.expireAt > 0 && ent.expireAt > time.Now().UnixNano()) {
		return fmt.Errorf("object %q already exists on bucket %s", hash.String(), bucket)
	}

	m.id++
	m.objects[bucket][hash] = &entity{
		id:       m.id,
		expireAt: time.Now().Add(m.expiration).UnixNano(),
		obj:      obj,
	}
	return nil
}

func (m *MemStorage) RemoveObject(ctx context.Context, bucket string, hash ids.ID32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	objs, ok := m.objects[bucket]
	if !ok {
		return fmt.Errorf("bucket %q not found", bucket)
	}
	ent, ok := objs[hash]
	if !ok {
		return fmt.Errorf("object %q not found on bucket %s", hash.String(), bucket)
	}
	if ent.expireAt == 0 {
		return fmt.Errorf("object %q is permanent and can not be removed", hash.String())
	}

	delete(objs, hash)
	return nil
}

func (m *MemStorage) BatchAcquire(ctx context.Context, bucket string, hashList ids.IDList[ids.ID32]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	objs, ok := m.objects[bucket]
	if !ok {
		return fmt.Errorf("bucket %q not found", bucket)
	}

	now := time.Now().UnixNano()
	for _, hash := range hashList {
		ent, ok := objs[hash]
		if !ok {
			return fmt.Errorf("object %q not found on bucket %s", hash.String(), bucket)
		}
		if ent.expireAt == 0 {
			return fmt.Errorf("object %q was already accepted", hash.String())
		}
		if ent.expireAt+AcquireRemainingLife <= now {
			return fmt.Errorf("object %q will be expired", hash.String())
		}
	}

	return nil
}

func (m *MemStorage) BatchAccept(ctx context.Context, bucket string, height uint64, hashList ids.IDList[ids.ID32]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	objs, ok := m.objects[bucket]
	if !ok {
		return fmt.Errorf("bucket %q not found", bucket)
	}

	for _, hash := range hashList {
		ent, ok := objs[hash]
		if !ok {
			return fmt.Errorf("object %q not found on bucket %s", hash.String(), bucket)
		}
		if ent.expireAt == 0 {
			return fmt.Errorf("object %q was already accepted", hash.String())
		}
	}

	for _, hash := range hashList {
		ent := objs[hash]
		ent.expireAt = 0
		ent.obj.Height = int64(height)
	}

	return nil
}

func (m *MemStorage) ListUnaccept(ctx context.Context, bucket, token string) (
	hashList ids.IDList[ids.ID32], nextToken string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ents, ok := m.objects[bucket]
	if !ok {
		return
	}

	if token == "" {
		token = "0"
	}

	id, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid token %q", token)
		return
	}

	list := make([]*entity, 0, 100)
	now := time.Now().UnixNano()
	for i := range ents {
		ent := ents[i]
		if ent.expireAt == 0 || now >= ent.expireAt || ent.id <= id {
			continue
		}
		list = append(list, ent)
	}

	sort.SliceStable(list, func(i, j int) bool { return list[i].id < list[j].id })

	l := len(list)
	if l > 100 {
		l = 100
	}
	hashList = make(ids.IDList[ids.ID32], l)
	for i := 0; i < l; i++ {
		hashList[i] = list[i].obj.Hash()
	}

	if len(list) > l {
		nextToken = strconv.FormatUint(list[l-1].id, 10)
	}
	return
}

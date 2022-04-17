// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
)

var (
	_ BlockState = &blockState{}
)

var poolAccountCache = sync.Pool{
	New: func() any {
		v := make(map[ids.ShortID]*Account, 256)
		return &v
	},
}

type blockState struct {
	ctx   *Context
	id    ids.ID
	state StateDB

	vdb            *versiondb.Database
	blockDB        *db.PrefixDB
	heightDB       *db.PrefixDB
	lastAcceptedDB *db.PrefixDB
	accountDB      *db.PrefixDB
	modelDB        *db.PrefixDB
	dataDB         *db.PrefixDB
	prevDataDB     *db.PrefixDB
	nameDB         *db.PrefixDB

	accountCache map[ids.ShortID]*Account
	events       []*Event
}

type BlockState interface {
	VersionDB() *versiondb.Database

	LoadAccount(ids.ShortID) (*Account, error)
	ResolveNameID(name string) (ids.ShortID, error)
	ResolveName(name string) (*ld.DataMeta, error)
	SetName(name string, id ids.ShortID) error
	LoadModel(ids.ShortID) (*ld.ModelMeta, error)
	SaveModel(ids.ShortID, *ld.ModelMeta) error
	LoadData(ids.ShortID) (*ld.DataMeta, error)
	SaveData(ids.ShortID, *ld.DataMeta) error
	SavePrevData(ids.ShortID, *ld.DataMeta) error

	AddEvent(*Event)
	Events() []*Event

	SaveBlock(*Block) error
	Commit() error
}

func newBlockState(ctx *Context, id ids.ID, baseVDB database.Database) *blockState {
	vdb := versiondb.New(baseVDB)
	accountCache := poolAccountCache.Get().(*map[ids.ShortID]*Account)

	pdb := db.NewPrefixDB(vdb, dbPrefix, 512)
	return &blockState{
		ctx:            ctx,
		id:             id,
		state:          ctx.StateDB(),
		vdb:            vdb,
		blockDB:        pdb.With(blockDBPrefix),
		heightDB:       pdb.With(heightDBPrefix),
		lastAcceptedDB: pdb.With(lastAcceptedKey),
		accountDB:      pdb.With(accountDBPrefix),
		modelDB:        pdb.With(modelDBPrefix),
		dataDB:         pdb.With(dataDBPrefix),
		prevDataDB:     pdb.With(prevDataDBPrefix),
		nameDB:         pdb.With(nameDBPrefix),
		accountCache:   *accountCache,
	}
}

func (bs *blockState) VersionDB() *versiondb.Database {
	return bs.vdb
}

func (bs *blockState) LoadAccount(id ids.ShortID) (*Account, error) {
	a := bs.accountCache[id]
	if a != nil {
		return a, nil
	}

	if id == constants.BlackholeAddr {
		return nil, fmt.Errorf("blackhole address should not be used")
	}

	if bs.accountCache[id] == nil {
		data, err := bs.accountDB.Get(id[:])
		switch err {
		case nil:
			a, err = ParseAccount(id, data)
		case database.ErrNotFound:
			err = nil
			a = NewAccount(id)
		}

		if err != nil {
			return nil, err
		}

		a.Init(bs.accountDB)
		bs.accountCache[id] = a
	}

	return bs.accountCache[id], nil
}

func (bs *blockState) SetName(name string, id ids.ShortID) error {
	key := []byte(strings.ToLower(name))
	data, err := bs.nameDB.Get(key)
	switch err {
	case nil:
		if !bytes.Equal(data, id[:]) {
			return fmt.Errorf("name conflict")
		}
	case database.ErrNotFound:
		err = bs.nameDB.Put(key, id[:])
	}

	return err
}

func (bs *blockState) ResolveNameID(name string) (ids.ShortID, error) {
	data, err := bs.nameDB.Get([]byte(strings.ToLower(name)))
	if err != nil {
		return ids.ShortEmpty, err
	}
	return ids.ToShortID(data)
}

func (bs *blockState) ResolveName(name string) (*ld.DataMeta, error) {
	id, err := bs.ResolveNameID(name)
	if err != nil {
		return nil, err
	}
	return bs.LoadData(id)
}

func (bs *blockState) LoadModel(id ids.ShortID) (*ld.ModelMeta, error) {
	data, err := bs.modelDB.Get(id[:])
	if err != nil {
		return nil, err
	}
	mm := &ld.ModelMeta{}
	if err := mm.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := mm.SyntacticVerify(); err != nil {
		return nil, err
	}
	mm.ID = id
	return mm, nil
}

func (bs *blockState) SaveModel(id ids.ShortID, mm *ld.ModelMeta) error {
	if mm == nil {
		return fmt.Errorf("SaveData with nil ModelMeta")
	}
	if err := mm.SyntacticVerify(); err != nil {
		return err
	}
	return bs.modelDB.Put(id[:], mm.Bytes())
}

func (bs *blockState) LoadData(id ids.ShortID) (*ld.DataMeta, error) {
	data, err := bs.dataDB.Get(id[:])
	if err != nil {
		return nil, err
	}
	dm := &ld.DataMeta{}
	if err := dm.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := dm.SyntacticVerify(); err != nil {
		return nil, err
	}
	dm.ID = id
	return dm, nil
}

func (bs *blockState) SaveData(id ids.ShortID, dm *ld.DataMeta) error {
	if dm == nil {
		return fmt.Errorf("SaveData with nil DataMeta")
	}
	if err := dm.SyntacticVerify(); err != nil {
		return err
	}
	return bs.dataDB.Put(id[:], dm.Bytes())
}

func (bs *blockState) SavePrevData(id ids.ShortID, dm *ld.DataMeta) error {
	if dm == nil {
		return fmt.Errorf("SavePrevData with nil DataMeta")
	}
	if err := dm.SyntacticVerify(); err != nil {
		return err
	}

	v := database.PackUInt64(dm.Version)
	key := make([]byte, 20+len(v))
	copy(key, id[:])
	copy(key[20:], v)
	return bs.prevDataDB.Put(key, dm.Bytes())
}

func (bs *blockState) AddEvent(e *Event) {
	if e != nil {
		bs.events = append(bs.events, e)
	}
}

func (bs *blockState) Events() []*Event {
	return bs.events
}

func (bs *blockState) SaveBlock(blk *Block) error {
	for _, a := range bs.accountCache {
		if err := a.Commit(); err != nil {
			return err
		}
	}
	id := blk.ID()
	if err := bs.blockDB.Put(id[:], blk.Bytes()); err != nil {
		return err
	}
	return bs.heightDB.Put(database.PackUInt64(blk.Height()), id[:])
}

// Commit when accept
func (bs *blockState) Commit() error {
	defer bs.free()
	if err := bs.vdb.SetDatabase(bs.state.DB()); err != nil {
		return err
	}
	return bs.vdb.Commit()
}

func (bs *blockState) free() {
	logging.Log.Info("free blockState %s", bs.id)
	for k := range bs.accountCache {
		delete(bs.accountCache, k)
	}
	poolAccountCache.Put(&bs.accountCache)
	bs.accountCache = nil
}

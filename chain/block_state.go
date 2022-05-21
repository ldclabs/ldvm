// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"golang.org/x/net/idna"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

var (
	_ BlockState = &blockState{}
)

type blockState struct {
	ctx               *Context
	height, timestamp uint64
	state             StateDB

	vdb            *versiondb.Database
	blockDB        *db.PrefixDB
	heightDB       *db.PrefixDB
	lastAcceptedDB *db.PrefixDB
	accountDB      *db.PrefixDB
	modelDB        *db.PrefixDB
	dataDB         *db.PrefixDB
	prevDataDB     *db.PrefixDB
	nameDB         *db.PrefixDB

	accountCache accountCache
	events       []*Event
}

type BlockState interface {
	DeriveState() *blockState
	VersionDB() *versiondb.Database

	LoadAccount(util.EthID) (*Account, error)
	ResolveNameID(name string) (util.DataID, error)
	ResolveName(name string) (*ld.DataMeta, error)
	SetName(name string, id util.DataID) error
	LoadModel(util.ModelID) (*ld.ModelMeta, error)
	SaveModel(util.ModelID, *ld.ModelMeta) error
	LoadData(util.DataID) (*ld.DataMeta, error)
	SaveData(util.DataID, *ld.DataMeta) error
	SavePrevData(util.DataID, *ld.DataMeta) error
	DeleteData(util.DataID, *ld.DataMeta) error

	AddEvent(*Event)
	Events() []*Event

	SaveBlock(*Block) error
	Commit() error
}

func newBlockState(ctx *Context, height, timestamp uint64, baseVDB database.Database) *blockState {
	vdb := versiondb.New(baseVDB)
	pdb := db.NewPrefixDB(vdb, dbPrefix, 512)
	return &blockState{
		ctx:            ctx,
		height:         height,
		timestamp:      timestamp,
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
		accountCache:   getAccountCache(),
	}
}

// DeriveState for the given block
func (bs *blockState) DeriveState() *blockState {
	vdb := versiondb.New(bs.vdb)
	pdb := db.NewPrefixDB(vdb, dbPrefix, 512)
	return &blockState{
		ctx:            bs.ctx,
		height:         bs.height,
		timestamp:      bs.timestamp,
		state:          bs.ctx.StateDB(),
		vdb:            vdb,
		blockDB:        pdb.With(blockDBPrefix),
		heightDB:       pdb.With(heightDBPrefix),
		lastAcceptedDB: pdb.With(lastAcceptedKey),
		accountDB:      pdb.With(accountDBPrefix),
		modelDB:        pdb.With(modelDBPrefix),
		dataDB:         pdb.With(dataDBPrefix),
		prevDataDB:     pdb.With(prevDataDBPrefix),
		nameDB:         pdb.With(nameDBPrefix),
		accountCache:   getAccountCache(),
	}
}

func (bs *blockState) VersionDB() *versiondb.Database {
	return bs.vdb
}

func (bs *blockState) LoadAccount(id util.EthID) (*Account, error) {
	a := bs.accountCache[id]
	if a == nil {
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

		pledge := new(big.Int)
		feeCfg := bs.ctx.Chain().Fee(bs.height)
		switch {
		case a.ld.Type == ld.TokenAccount && id != constants.LDCAccount:
			pledge.Set(feeCfg.MinTokenPledge)
		case a.ld.Type == ld.StakeAccount:
			pledge.Set(feeCfg.MinStakePledge)
		}

		a.Init(pledge, bs.height, bs.timestamp)
		bs.accountCache[id] = a
	}

	return bs.accountCache[id], nil
}

func (bs *blockState) SetName(name string, id util.DataID) error {
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

func (bs *blockState) ResolveNameID(name string) (util.DataID, error) {
	dn, err := idna.Registration.ToASCII(name)
	if err != nil {
		return util.DataIDEmpty, fmt.Errorf("invalid name %s, error: %v",
			strconv.Quote(name), err)
	}
	data, err := bs.nameDB.Get([]byte(dn))
	if err != nil {
		return util.DataIDEmpty, err
	}
	id, err := ids.ToShortID(data)
	return util.DataID(id), err
}

func (bs *blockState) ResolveName(name string) (*ld.DataMeta, error) {
	id, err := bs.ResolveNameID(name)
	if err != nil {
		return nil, err
	}
	return bs.LoadData(id)
}

func (bs *blockState) LoadModel(id util.ModelID) (*ld.ModelMeta, error) {
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

func (bs *blockState) SaveModel(id util.ModelID, mm *ld.ModelMeta) error {
	if mm == nil {
		return fmt.Errorf("SaveData with nil ModelMeta")
	}
	if err := mm.SyntacticVerify(); err != nil {
		return err
	}
	if ok, _ := bs.modelDB.Has(id[:]); ok {
		return fmt.Errorf("SaveModel error: model %s exists", util.ModelID(id).String())
	}
	return bs.modelDB.Put(id[:], mm.Bytes())
}

func (bs *blockState) LoadData(id util.DataID) (*ld.DataMeta, error) {
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

func (bs *blockState) SaveData(id util.DataID, dm *ld.DataMeta) error {
	if dm == nil {
		return fmt.Errorf("SaveData with nil DataMeta")
	}
	if err := dm.SyntacticVerify(); err != nil {
		return err
	}
	if dm.Version == 1 {
		if ok, _ := bs.dataDB.Has(id[:]); ok {
			return fmt.Errorf("SaveData error: data %s exists", util.DataID(id).String())
		}
	}
	return bs.dataDB.Put(id[:], dm.Bytes())
}

func (bs *blockState) SavePrevData(id util.DataID, dm *ld.DataMeta) error {
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

func (bs *blockState) DeleteData(id util.DataID, dm *ld.DataMeta) error {
	version := dm.Version
	dm.Version = 0 // mark dropped
	if err := bs.SaveData(id, dm); err != nil {
		return err
	}
	for version > 0 {
		v := database.PackUInt64(version)
		version--
		key := make([]byte, 20+len(v))
		copy(key, id[:])
		copy(key[20:], v)
		bs.prevDataDB.Delete(key)
	}
	return nil
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
		if err := a.SaveTo(bs.accountDB); err != nil {
			return err
		}
	}
	id := blk.ID()
	if ok, _ := bs.blockDB.Has(id[:]); ok {
		return fmt.Errorf("SaveBlock error: block %s at height %d exists", id.String(), blk.Height())
	}
	if err := bs.blockDB.Put(id[:], blk.Bytes()); err != nil {
		return err
	}
	hKey := database.PackUInt64(blk.Height())
	if ok, _ := bs.heightDB.Has(hKey); ok {
		return fmt.Errorf("SaveBlock height error: block %s at height %d exists", id.String(), blk.Height())
	}
	return bs.heightDB.Put(hKey, id[:])
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
	logging.Log.Info("free blockState at height %d", bs.height)
	putAccountCache(bs.accountCache)
	bs.accountCache = nil
}

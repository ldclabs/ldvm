// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
)

var (
	_ StateBlock = &stateBlock{}
)

var poolAccountCache = sync.Pool{
	New: func() any {
		v := make(map[ids.ShortID]*Account, 256)
		return &v
	},
}

type stateBlock struct {
	mu        sync.RWMutex
	state     *stateDB
	committed bool

	vdb        *versiondb.Database
	heightVDB  *prefixdb.Database
	blockVDB   *prefixdb.Database
	accountVDB *prefixdb.Database
	modelVDB   *prefixdb.Database
	dataVDB    *prefixdb.Database

	accountCache map[ids.ShortID]*Account
}

type StateBlock interface {
	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig

	PreferredBlock() *Block
	SetLastAccepted(blk *Block) error
	SaveBlock(*Block) error

	LoadAccount(ids.ShortID) (*Account, error)

	GivebackTxs(txs ...*ld.Transaction)

	Commit() error
}

func newStateBlock(s *stateDB, db database.Database) *stateBlock {
	vdb := versiondb.New(db)
	accountCache := poolAccountCache.Get().(*map[ids.ShortID]*Account)

	return &stateBlock{
		state:        s,
		vdb:          vdb,
		heightVDB:    prefixdb.New(heightDBPrefix, vdb),
		blockVDB:     prefixdb.New(blockDBPrefix, vdb),
		accountVDB:   prefixdb.New(accountDBPrefix, vdb),
		modelVDB:     prefixdb.New(modelDBPrefix, vdb),
		dataVDB:      prefixdb.New(dataDBPrefix, vdb),
		accountCache: *accountCache,
	}
}

func (sb *stateBlock) ChainConfig() *genesis.ChainConfig {
	return sb.state.ChainConfig()
}

func (sb *stateBlock) FeeConfig() *genesis.FeeConfig {
	return sb.state.FeeConfig()
}

func (sb *stateBlock) PreferredBlock() *Block {
	return sb.state.PreferredBlock()
}

func (sb *stateBlock) LoadAccount(id ids.ShortID) (*Account, error) {
	if id == ids.ShortEmpty {
		return nil, fmt.Errorf("should not be empty account")
	}

	sb.mu.RLock()
	a := sb.accountCache[id]
	sb.mu.RUnlock()
	if a != nil {
		return a, nil
	}

	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.accountCache[id] == nil {
		bytes, err := sb.accountVDB.Get(id[:])
		if err == nil {
			a, err = ParseAccount(bytes)
		} else if err == database.ErrNotFound {
			a = NewAccount()
		} else {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
		a.Init(id, sb.accountVDB)
		sb.accountCache[id] = a
	}

	return sb.accountCache[id], nil
}

func (sb *stateBlock) SetLastAccepted(blk *Block) error {
	return sb.state.SetLastAccepted(blk)
}

func (sb *stateBlock) SaveBlock(blk *Block) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	for _, a := range sb.accountCache {
		if err := a.Commit(); err != nil {
			return err
		}
	}
	if err := sb.blockVDB.Put(blk.id[:], blk.Bytes()); err != nil {
		return err
	}
	return sb.heightVDB.Put(database.PackUInt64(blk.Height()), blk.id[:])
}

// Commit when preferred
func (sb *stateBlock) Commit() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.committed {
		return nil
	}
	defer sb.clear()

	if err := sb.vdb.Commit(); err != nil {
		return err
	}

	sb.committed = true
	return nil
}

func (sb *stateBlock) GivebackTxs(txs ...*ld.Transaction) {
	sb.state.AddTxs(txs...)
}

func (sb *stateBlock) clear() {
	for k := range sb.accountCache {
		delete(sb.accountCache, k)
	}
	poolAccountCache.Put(&sb.accountCache)
	sb.accountCache = nil
}

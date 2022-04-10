// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/prefixdb"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
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
	ctx       *snow.Context
	state     *stateDB
	committed bool
	height    uint64

	vdb        *versiondb.Database
	heightVDB  *prefixdb.Database
	blockVDB   *prefixdb.Database
	accountVDB *prefixdb.Database
	modelVDB   *prefixdb.Database
	dataVDB    *prefixdb.Database
	nameVDB    *prefixdb.Database

	accountCache map[ids.ShortID]*Account
	events       []*Event
}

type BlockState interface {
	ChainConfig() *genesis.ChainConfig
	FeeConfig() *genesis.FeeConfig

	PreferredBlock() *Block
	SetLastAccepted(blk *Block) error
	SaveBlock(*Block) error

	TotalSupply() *big.Int
	LoadAccount(ids.ShortID) (*Account, error)

	ResolveNameID(name string) (ids.ShortID, error)
	ResolveName(name string) (*ld.DataMeta, error)
	SetName(name string, id ids.ShortID) error
	LoadModel(ids.ShortID) (*ld.ModelMeta, error)
	SaveModel(ids.ShortID, *ld.ModelMeta) error
	LoadData(ids.ShortID) (*ld.DataMeta, error)
	SaveData(ids.ShortID, *ld.DataMeta) error

	ProposeMintFeeTx(uint64, ids.ID, *big.Int)
	PopBySize(askSize uint64) []*ld.Transaction
	GivebackTxs(txs ...*ld.Transaction)

	AddEvent(*Event)
	Events() []*Event

	Commit() error
	Log() logging.Logger
}

func newBlockState(s *stateDB, db database.Database, height uint64) *blockState {
	vdb := versiondb.New(db)
	accountCache := poolAccountCache.Get().(*map[ids.ShortID]*Account)

	return &blockState{
		ctx:          s.ctx,
		height:       height,
		state:        s,
		vdb:          vdb,
		heightVDB:    prefixdb.New(heightDBPrefix, vdb),
		blockVDB:     prefixdb.New(blockDBPrefix, vdb),
		accountVDB:   prefixdb.New(accountDBPrefix, vdb),
		modelVDB:     prefixdb.New(modelDBPrefix, vdb),
		dataVDB:      prefixdb.New(dataDBPrefix, vdb),
		nameVDB:      prefixdb.New(nameDBPrefix, vdb),
		accountCache: *accountCache,
	}
}

func (bs *blockState) Log() logging.Logger {
	return bs.state.Log()
}

func (bs *blockState) ChainConfig() *genesis.ChainConfig {
	return bs.state.ChainConfig()
}

func (bs *blockState) FeeConfig() *genesis.FeeConfig {
	return bs.state.FeeConfig(bs.height)
}

func (bs *blockState) PreferredBlock() *Block {
	return bs.state.PreferredBlock()
}

func (bs *blockState) TotalSupply() *big.Int {
	s := new(big.Int)
	if acc, err := bs.LoadAccount(constants.GenesisAddr); err == nil {
		max := bs.ChainConfig().MaxTotalSupply
		s.Sub(max, acc.Balance())
	}
	return s
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
		data, err := bs.accountVDB.Get(id[:])
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

		a.Init(bs.accountVDB)
		bs.accountCache[id] = a
	}

	return bs.accountCache[id], nil
}

func (bs *blockState) SetName(name string, id ids.ShortID) error {
	key := []byte(strings.ToLower(name))
	data, err := bs.nameVDB.Get(key)
	switch err {
	case nil:
		if !bytes.Equal(data, id[:]) {
			return fmt.Errorf("name conflict")
		}
	case database.ErrNotFound:
		err = bs.nameVDB.Put(key, id[:])
	}

	return err
}

func (bs *blockState) ResolveNameID(name string) (ids.ShortID, error) {
	data, err := bs.nameVDB.Get([]byte(strings.ToLower(name)))
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
	data, err := bs.modelVDB.Get(id[:])
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
	return bs.modelVDB.Put(id[:], mm.Bytes())
}

func (bs *blockState) LoadData(id ids.ShortID) (*ld.DataMeta, error) {
	data, err := bs.dataVDB.Get(id[:])
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
	return bs.dataVDB.Put(id[:], dm.Bytes())
}

func (bs *blockState) AddEvent(e *Event) {
	if e != nil {
		bs.events = append(bs.events, e)
	}
}

func (bs *blockState) Events() []*Event {
	return bs.events
}

func (bs *blockState) SetLastAccepted(blk *Block) error {
	return bs.state.SetLastAccepted(blk)
}

func (bs *blockState) SaveBlock(blk *Block) error {
	for _, a := range bs.accountCache {
		if err := a.Commit(); err != nil {
			return err
		}
	}
	id := blk.ID()
	if err := bs.blockVDB.Put(id[:], blk.Bytes()); err != nil {
		return err
	}
	return bs.heightVDB.Put(database.PackUInt64(blk.Height()), id[:])
}

// Commit when preferred
func (bs *blockState) Commit() error {
	if bs.committed {
		return nil
	}
	defer bs.clear()

	if err := bs.vdb.Commit(); err != nil {
		return err
	}

	bs.committed = true
	return nil
}

func (bs *blockState) PopBySize(askSize uint64) []*ld.Transaction {
	return bs.state.PopBySize(askSize)
}

func (bs *blockState) mintTxData(blkID ids.ID) []byte {
	data := make([]byte, 20+32)
	copy(data, bs.ctx.NodeID[:])
	copy(data[20:], blkID[:])
	return data
}

func (bs *blockState) ProposeMintFeeTx(blkHeight uint64, blkID ids.ID, blkCost *big.Int) {
	if recipient := bs.state.Config().FeeRecipient.ShortID(); recipient != ids.ShortEmpty {
		mintFeeTx := &ld.Transaction{
			Type:    ld.TypeMintFee,
			ChainID: bs.ChainConfig().ChainID,
			From:    constants.GenesisAddr,
			To:      recipient,
			Nonce:   blkHeight,
			Amount:  blkCost,
			Data:    bs.mintTxData(blkID),
		}
		bs.state.ProposeTx(mintFeeTx)
	}
}

func (bs *blockState) GivebackTxs(txs ...*ld.Transaction) {
	bs.state.AddTxs(txs...)
}

func (bs *blockState) clear() {
	for k := range bs.accountCache {
		delete(bs.accountCache, k)
	}
	poolAccountCache.Put(&bs.accountCache)
	bs.accountCache = nil
}

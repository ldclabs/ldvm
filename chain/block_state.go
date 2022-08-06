// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"

	"github.com/ldclabs/ldvm/chain/transaction"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

var (
	_ BlockState = &blockState{}
)

var poolAccountCache = sync.Pool{
	New: func() any {
		v := make(transaction.AccountCache, 256)
		return &v
	},
}

func getAccountCache() transaction.AccountCache {
	ac := poolAccountCache.Get().(*transaction.AccountCache)
	return *ac
}

func putAccountCache(cc transaction.AccountCache) {
	for k := range cc {
		delete(cc, k)
	}
	poolAccountCache.Put(&cc)
}

type blockState struct {
	height, timestamp uint64
	ctx               *Context
	ls                *ld.State
	bc                BlockChain
	vdb               *versiondb.Database
	blockDB           *db.PrefixDB
	heightDB          *db.PrefixDB
	lastAcceptedDB    *db.PrefixDB
	accountDB         *db.PrefixDB
	ledgerDB          *db.PrefixDB
	modelDB           *db.PrefixDB
	dataDB            *db.PrefixDB
	prevDataDB        *db.PrefixDB
	stateDB           *db.PrefixDB
	nameDB            *db.PrefixDB
	accountCache      transaction.AccountCache
}

type BlockState interface {
	VersionDB() *versiondb.Database
	DeriveState() (BlockState, error)
	LoadStakeAccountByNodeID(ids.NodeID) (util.StakeSymbol, *transaction.Account)
	GetBlockIDAtHeight(uint64) (ids.ID, error)
	SaveBlock(*ld.Block) error
	Commit() error
	Free()

	transaction.BlockState
}

func newBlockState(ctx *Context, height, timestamp uint64, parentState ids.ID, baseDB database.Database) *blockState {
	vdb := versiondb.New(baseDB)
	pdb := db.NewPrefixDB(vdb, dbPrefix, 512)
	bs := &blockState{
		ctx:            ctx,
		height:         height,
		timestamp:      timestamp,
		bc:             ctx.Chain(),
		ls:             ld.NewState(parentState),
		vdb:            vdb,
		blockDB:        pdb.With(blockDBPrefix),
		heightDB:       pdb.With(heightDBPrefix),
		lastAcceptedDB: pdb.With(lastAcceptedDBPrefix),
		accountDB:      pdb.With(accountDBPrefix),
		ledgerDB:       pdb.With(ledgerDBPrefix),
		modelDB:        pdb.With(modelDBPrefix),
		dataDB:         pdb.With(dataDBPrefix),
		prevDataDB:     pdb.With(prevDataDBPrefix),
		stateDB:        pdb.With(stateDBPrefix),
		nameDB:         pdb.With(nameDBPrefix),
		accountCache:   getAccountCache(),
	}

	bs.nameDB.SetHashKey(nameHashKey)
	return bs
}

func (bs *blockState) Height() uint64 {
	return bs.height
}

func (bs *blockState) Timestamp() uint64 {
	return bs.timestamp
}

// DeriveState for the given block
func (bs *blockState) DeriveState() (BlockState, error) {
	vdb := versiondb.New(bs.vdb)
	batch, err := bs.vdb.CommitBatch()
	if err != nil {
		return nil, err
	}
	if err = batch.Replay(vdb); err != nil {
		return nil, err
	}
	pdb := db.NewPrefixDB(vdb, dbPrefix, 512)
	nbs := &blockState{
		ctx:            bs.ctx,
		height:         bs.height,
		timestamp:      bs.timestamp,
		ls:             bs.ls.Clone(),
		bc:             bs.ctx.Chain(),
		vdb:            vdb,
		blockDB:        pdb.With(blockDBPrefix),
		heightDB:       pdb.With(heightDBPrefix),
		lastAcceptedDB: pdb.With(lastAcceptedKey),
		accountDB:      pdb.With(accountDBPrefix),
		ledgerDB:       pdb.With(ledgerDBPrefix),
		modelDB:        pdb.With(modelDBPrefix),
		dataDB:         pdb.With(dataDBPrefix),
		prevDataDB:     pdb.With(prevDataDBPrefix),
		stateDB:        pdb.With(stateDBPrefix),
		nameDB:         pdb.With(nameDBPrefix),
		accountCache:   getAccountCache(),
	}

	nbs.nameDB.SetHashKey(nameHashKey)

	for _, a := range bs.accountCache {
		data, ledger, err := a.Marshal()
		if err == nil {
			id := a.ID()
			if a.AccountChanged(data) {
				nbs.ls.UpdateAccount(id, data)
				err = nbs.accountDB.Put(id[:], data)
			}

			if err == nil && len(ledger) > 0 && a.LedgerChanged(ledger) {
				nbs.ls.UpdateLedger(id, ledger)
				err = nbs.ledgerDB.Put(id[:], ledger)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return nbs, nil
}

func (bs *blockState) VersionDB() *versiondb.Database {
	return bs.vdb
}

func (bs *blockState) LoadAccount(id util.EthID) (*transaction.Account, error) {
	acc := bs.accountCache[id]
	if acc == nil {
		data, err := bs.accountDB.Get(id[:])
		switch err {
		case nil:
			acc, err = transaction.ParseAccount(id, data)
		case database.ErrNotFound:
			err = nil
			acc = transaction.NewAccount(id)
		}

		if err != nil {
			return nil, err
		}

		pledge := new(big.Int)
		feeCfg := bs.ctx.ChainConfig().Fee(bs.height)
		switch {
		case acc.Type() == ld.TokenAccount && id != constants.LDCAccount:
			pledge.Set(feeCfg.MinTokenPledge)

		case acc.Type() == ld.StakeAccount:
			pledge.Set(feeCfg.MinStakePledge)
		}

		acc.Init(pledge, bs.height, bs.timestamp)
		bs.accountCache[id] = acc
	}

	return bs.accountCache[id], nil
}

func (bs *blockState) LoadLedger(acc *transaction.Account) error {
	if acc.Ledger() == nil {
		id := acc.ID()
		data, err := bs.ledgerDB.Get(id[:])
		if err != nil && err != database.ErrNotFound {
			return err
		}
		return acc.InitLedger(data)
	}
	return nil
}

func (bs *blockState) LoadStakeAccountByNodeID(nodeID ids.NodeID) (util.StakeSymbol, *transaction.Account) {
	id := util.EthID(nodeID).ToStakeSymbol()
	acc, err := bs.LoadAccount(util.EthID(id))
	if err != nil || !acc.Valid(ld.StakeAccount) {
		return util.StakeEmpty, nil
	}
	return id, acc
}

func (bs *blockState) LoadMiner(id util.StakeSymbol) (*transaction.Account, error) {
	miner := constants.GenesisAccount
	if id != util.StakeEmpty && id.Valid() {
		acc, err := bs.LoadAccount(util.EthID(id))
		if err == nil && acc.Valid(ld.StakeAccount) {
			return acc, nil
		}
	}
	return bs.LoadAccount(miner)
}

func (bs *blockState) SaveName(ns *service.Name) error {
	if ns.DID == util.DataIDEmpty {
		return fmt.Errorf("blockState.SaveName: data ID is empty")
	}

	name := ns.ASCII()
	key := []byte(name)
	ok, err := bs.nameDB.Has(key)
	switch {
	case ok:
		return fmt.Errorf("name %q conflict", name)
	case err == nil:
		err = bs.nameDB.Put(key, ns.DID[:])
	}

	return err
}

func (bs *blockState) LoadModel(id util.ModelID) (*ld.ModelInfo, error) {
	data, err := bs.modelDB.Get(id[:])
	if err != nil {
		return nil, err
	}
	mi := &ld.ModelInfo{}
	if err := mi.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := mi.SyntacticVerify(); err != nil {
		return nil, err
	}
	mi.ID = id
	return mi, nil
}

func (bs *blockState) SaveModel(mi *ld.ModelInfo) error {
	if mi.ID == util.ModelIDEmpty {
		return fmt.Errorf("blockState.SaveModel: model ID is empty")
	}

	if err := mi.SyntacticVerify(); err != nil {
		return err
	}
	bs.ls.UpdateModel(mi.ID, mi.Bytes())
	return bs.modelDB.Put(mi.ID[:], mi.Bytes())
}

func (bs *blockState) LoadData(id util.DataID) (*ld.DataInfo, error) {
	data, err := bs.dataDB.Get(id[:])
	if err != nil {
		return nil, err
	}
	di := &ld.DataInfo{}
	if err := di.Unmarshal(data); err != nil {
		return nil, err
	}
	if err := di.SyntacticVerify(); err != nil {
		return nil, err
	}
	di.ID = id
	return di, nil
}

func (bs *blockState) SaveData(di *ld.DataInfo) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("blockState.SaveData: data ID is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	bs.ls.UpdateData(di.ID, di.Bytes())
	return bs.dataDB.Put(di.ID[:], di.Bytes())
}

func (bs *blockState) SavePrevData(di *ld.DataInfo) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("blockState.SavePrevData: data ID is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return err
	}

	return bs.prevDataDB.Put(di.ID.VersionKey(di.Version), di.Bytes())
}

func (bs *blockState) DeleteData(di *ld.DataInfo, message []byte) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("blockState.DeleteData: data ID is empty")
	}

	version := di.Version
	if err := di.MarkDeleted(message); err != nil {
		return err
	}
	if err := bs.SaveData(di); err != nil {
		return err
	}
	for version > 0 {
		bs.prevDataDB.Delete(di.ID.VersionKey(version))
		version--
	}
	return nil
}

func (bs *blockState) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	data, err := bs.heightDB.Get(database.PackUInt64(height))
	if err != nil {
		return ids.Empty, err
	}
	return ids.ToID(data)
}

func (bs *blockState) SaveBlock(blk *ld.Block) error {
	for _, a := range bs.accountCache {
		data, ledger, err := a.Marshal()
		if err == nil {
			id := a.ID()
			if a.AccountChanged(data) {
				bs.ls.UpdateAccount(id, data)
				err = bs.accountDB.Put(id[:], data)
			}

			if err == nil && len(ledger) > 0 && a.LedgerChanged(ledger) {
				bs.ls.UpdateLedger(id, ledger)
				err = bs.ledgerDB.Put(id[:], ledger)
			}
		}
		if err != nil {
			return err
		}
	}
	if err := bs.ls.SyntacticVerify(); err != nil {
		return err
	}

	// will update block's state and id
	blk.State = bs.ls.ID
	if err := blk.SyntacticVerify(); err != nil {
		return err
	}

	hKey := database.PackUInt64(blk.Height)
	if ok, _ := bs.heightDB.Has(hKey); ok {
		return fmt.Errorf("SaveBlock height error: block %s at height %d exists", blk.ID, blk.Height)
	}
	if err := bs.blockDB.Put(blk.ID[:], blk.Bytes()); err != nil {
		return err
	}
	return bs.heightDB.Put(hKey, blk.ID[:])
}

// Commit when accept
func (bs *blockState) Commit() error {
	if err := bs.vdb.SetDatabase(bs.bc.DB()); err != nil {
		return err
	}
	return bs.vdb.Commit()
}

func (bs *blockState) Free() {
	logging.Log.Info("free blockState at height %d", bs.height)
	putAccountCache(bs.accountCache)
	bs.accountCache = nil
}

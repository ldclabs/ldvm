// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"math/big"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/versiondb"
	avaids "github.com/ava-labs/avalanchego/ids"
	"go.uber.org/zap"

	"github.com/ldclabs/ldvm/chain/acct"
	"github.com/ldclabs/ldvm/chain/txn"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util/erring"
)

var (
	_ BlockState = &blockState{}
)

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
	accts             acct.ActiveAccounts
}

type BlockState interface {
	VersionDB() *versiondb.Database
	DeriveState() (BlockState, error)
	LoadValidatorAccountByNodeID(avaids.NodeID) (ids.StakeSymbol, *acct.Account)
	GetBlockIDAtHeight(uint64) (ids.ID32, error)
	SaveBlock(*ld.Block) error
	Commit() error
	Free()

	txn.ChainState
}

func newBlockState(ctx *Context, height, timestamp uint64, parentState ids.ID32, baseDB database.Database) *blockState {
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
		accts:          make(acct.ActiveAccounts, 256),
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
		accts:          make(acct.ActiveAccounts, 256),
	}

	nbs.nameDB.SetHashKey(nameHashKey)

	for _, a := range bs.accts {
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

func (bs *blockState) LoadAccount(id ids.Address) (*acct.Account, error) {
	acc := bs.accts[id]
	if acc == nil {
		errp := erring.ErrPrefix("chain.BlockState.LoadAccount: ")
		data, err := bs.accountDB.Get(id[:])
		switch err {
		case nil:
			acc, err = acct.ParseAccount(id, data)
		case database.ErrNotFound:
			err = nil
			acc = acct.NewAccount(id)
		}

		if err != nil {
			return nil, errp.ErrorIf(err)
		}

		feeCfg := bs.ctx.ChainConfig().Fee(bs.height)
		switch {
		case id == ids.LDCAccount || id == ids.GenesisAccount:
			acc.Init(big.NewInt(0), big.NewInt(0), bs.height, bs.timestamp)

		case acc.Type() == ld.TokenAccount:
			acc.Init(big.NewInt(0), feeCfg.MinTokenPledge, bs.height, bs.timestamp)

		case acc.Type() == ld.StakeAccount:
			acc.Init(big.NewInt(0), feeCfg.MinStakePledge, bs.height, bs.timestamp)

		default:
			acc.Init(feeCfg.NonTransferableBalance, big.NewInt(0), bs.height, bs.timestamp)
		}

		bs.accts[id] = acc
	}

	return bs.accts[id], nil
}

func (bs *blockState) LoadLedger(acc *acct.Account) error {
	return acc.LoadLedger(false, func() ([]byte, error) {
		id := acc.ID()
		data, err := bs.ledgerDB.Get(id[:])
		if err != nil && err != database.ErrNotFound {
			return nil, err
		}
		return data, nil
	})
}

func (bs *blockState) LoadValidatorAccountByNodeID(nodeID avaids.NodeID) (ids.StakeSymbol, *acct.Account) {
	id := ids.Address(nodeID).ToStakeSymbol()
	acc, err := bs.LoadAccount(ids.Address(id))
	if err != nil || !acc.ValidValidator() {
		return ids.StakeSymbol{}, nil
	}
	return id, acc
}

func (bs *blockState) LoadBuilder(id ids.StakeSymbol) (*acct.Account, error) {
	if id != ids.EmptyStake && id.Valid() {
		acc, err := bs.LoadAccount(ids.Address(id))
		if err == nil && acc.ValidValidator() {
			return acc, nil
		}
	}
	return bs.LoadAccount(ids.GenesisAccount)
}

func (bs *blockState) SaveName(ns *service.Name) error {
	errp := erring.ErrPrefix("chain.BlockState.SaveName: ")
	if ns.DataID == ids.EmptyDataID {
		return errp.Errorf("data ID is empty")
	}

	name := ns.ASCII()
	key := []byte(name)
	ok, err := bs.nameDB.Has(key)
	switch {
	case ok:
		return errp.Errorf("name %q is conflict", name)

	case err != nil:
		return errp.ErrorIf(err)

	default:
		return errp.ErrorIf(bs.nameDB.Put(key, ns.DataID[:]))
	}
}

func (bs *blockState) DeleteName(ns *service.Name) error {
	errp := erring.ErrPrefix("chain.BlockState.DeleteName: ")
	if ns.DataID == ids.EmptyDataID {
		return errp.Errorf("data ID is empty")
	}

	name := ns.ASCII()
	key := []byte(name)
	ok, err := bs.nameDB.Has(key)
	switch {
	case ok:
		return errp.ErrorIf(bs.nameDB.Delete(key))

	case err != nil:
		return errp.ErrorIf(err)

	default:
		return errp.Errorf("name %q is not exist", name)
	}
}

func (bs *blockState) LoadModel(id ids.ModelID) (*ld.ModelInfo, error) {
	errp := erring.ErrPrefix("chain.BlockState.LoadModel: ")
	data, err := bs.modelDB.Get(id[:])
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	mi := &ld.ModelInfo{}
	if err := mi.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := mi.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	mi.ID = id
	return mi, nil
}

func (bs *blockState) SaveModel(mi *ld.ModelInfo) error {
	errp := erring.ErrPrefix("chain.BlockState.SaveModel: ")
	if mi.ID == ids.EmptyModelID {
		return errp.Errorf("model id is empty")
	}

	if err := mi.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	bs.ls.UpdateModel(mi.ID, mi.Bytes())
	return errp.ErrorIf(bs.modelDB.Put(mi.ID[:], mi.Bytes()))
}

func (bs *blockState) LoadData(id ids.DataID) (*ld.DataInfo, error) {
	errp := erring.ErrPrefix("chain.BlockState.LoadData: ")
	data, err := bs.dataDB.Get(id[:])
	if err != nil {
		return nil, errp.ErrorIf(err)
	}

	di := &ld.DataInfo{}
	if err := di.Unmarshal(data); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := di.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	di.ID = id
	return di, nil
}

func (bs *blockState) SaveData(di *ld.DataInfo) error {
	errp := erring.ErrPrefix("chain.BlockState.SaveData: ")
	if di.ID == ids.EmptyDataID {
		return errp.Errorf("data id is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}
	bs.ls.UpdateData(di.ID, di.Bytes())
	return errp.ErrorIf(bs.dataDB.Put(di.ID[:], di.Bytes()))
}

func (bs *blockState) SavePrevData(di *ld.DataInfo) error {
	errp := erring.ErrPrefix("chain.BlockState.SavePrevData: ")
	if di.ID == ids.EmptyDataID {
		return errp.Errorf("data id is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	return errp.ErrorIf(bs.prevDataDB.Put(di.ID.VersionKey(di.Version), di.Bytes()))
}

func (bs *blockState) DeleteData(di *ld.DataInfo, message []byte) error {
	errp := erring.ErrPrefix("chain.BlockState.DeleteData: ")
	if di.ID == ids.EmptyDataID {
		return errp.Errorf("data id is empty")
	}

	version := di.Version
	if err := di.MarkDeleted(message); err != nil {
		return errp.ErrorIf(err)
	}
	if err := bs.SaveData(di); err != nil {
		return errp.ErrorIf(err)
	}
	for version > 0 {
		bs.prevDataDB.Delete(di.ID.VersionKey(version))
		version--
	}
	return nil
}

func (bs *blockState) GetBlockIDAtHeight(height uint64) (ids.ID32, error) {
	errp := erring.ErrPrefix("chain.BlockState.GetBlockIDAtHeight: ")
	data, err := bs.heightDB.Get(database.PackUInt64(height))
	if err != nil {
		return ids.ID32{}, errp.ErrorIf(err)
	}
	return ids.ID32FromBytes(data)
}

func (bs *blockState) SaveBlock(blk *ld.Block) error {
	errp := erring.ErrPrefix("chain.BlockState.SaveBlock: ")
	for _, a := range bs.accts {
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
			return errp.ErrorIf(err)
		}
	}
	if err := bs.ls.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	// will update block's state and id
	blk.State = bs.ls.ID
	if err := blk.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	hKey := database.PackUInt64(blk.Height)
	if ok, _ := bs.heightDB.Has(hKey); ok {
		return errp.Errorf("block %s at height %d exists", blk.ID, blk.Height)
	}

	if err := bs.blockDB.Put(blk.ID[:], blk.Bytes()); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(bs.heightDB.Put(hKey, blk.ID[:]))
}

// Commit when accept
func (bs *blockState) Commit() error {
	errp := erring.ErrPrefix("chain.BlockState.Commit: ")
	if err := bs.vdb.SetDatabase(bs.bc.DB()); err != nil {
		return errp.ErrorIf(err)
	}
	return errp.ErrorIf(bs.vdb.Commit())
}

func (bs *blockState) Free() {
	logging.Log.Info("blockState.Free", zap.Uint64("height", bs.height))
}

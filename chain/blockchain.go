// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"golang.org/x/crypto/sha3"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

var _ BlockChain = &blockChain{}

var (
	// Should never happen
	errPreferredBlock = fmt.Errorf("BlockChain is not bootstrapped, no preferred block")
)

var (
	dbPrefix             = []byte("LDVM")
	lastAcceptedDBPrefix = []byte{'K'}
	heightDBPrefix       = []byte{'H'}
	blockDBPrefix        = []byte{'B'}
	accountDBPrefix      = []byte{'A'}
	ledgerDBPrefix       = []byte{'L'}
	modelDBPrefix        = []byte{'M'}
	dataDBPrefix         = []byte{'D'}
	prevDataDBPrefix     = []byte{'P'}
	stateDBPrefix        = []byte{'S'}
	nameDBPrefix         = []byte{'N'} // inverted index

	lastAcceptedKey = []byte("last_accepted_key")
)

// BlockChain defines methods to manage state with Blocks and LastAcceptedIDs.
type BlockChain interface {
	// global context
	Context() *Context
	DB() database.Database

	// global state
	HealthCheck() (interface{}, error)
	Bootstrap() error
	IsBootstrapped() bool
	SetState(snow.State) error
	TotalSupply() *big.Int

	// blocks state
	BuildBlock() (*Block, error)
	ParseBlock([]byte) (*Block, error)
	GetBlockIDAtHeight(uint64) (ids.ID, error)
	GetBlockAtHeight(uint64) (*Block, error)
	GetBlock(ids.ID) (*Block, error)
	LastAcceptedBlock() *Block
	SetLastAccepted(*Block) error
	PreferredBlock() *Block
	SetPreference(ids.ID) error
	AddVerifiedBlock(*Block)
	GetVerifiedBlock(id ids.ID) *Block

	// txs state
	SubmitTx(...*ld.Transaction) error
	AddRemoteTxs(tx ...*ld.Transaction) error
	AddLocalTxs(txs ...*ld.Transaction)
	SetTxsHeight(uint64, ...ids.ID)
	GetTxHeight(ids.ID) int64

	LoadAccount(util.EthID) (*ld.Account, error)
	LoadModel(util.ModelID) (*ld.ModelInfo, error)
	LoadData(util.DataID) (*ld.DataInfo, error)
	LoadPrevData(util.DataID, uint64) (*ld.DataInfo, error)
	ResolveName(name string) (*service.Name, error)
	LoadRawData(rawType string, key []byte) ([]byte, error)
}

type blockChain struct {
	ctx          *Context
	config       *config.Config
	genesis      *genesis.Genesis
	genesisBlock *Block

	db             database.Database
	blockDB        *db.PrefixDB
	heightDB       *db.PrefixDB
	lastAcceptedDB *db.PrefixDB
	accountDB      *db.PrefixDB
	ledgerDB       *db.PrefixDB
	modelDB        *db.PrefixDB
	dataDB         *db.PrefixDB
	prevDataDB     *db.PrefixDB
	stateDB        *db.PrefixDB
	nameDB         *db.PrefixDB

	preferred         *atomicBlock
	lastAcceptedBlock *atomicBlock
	state             *atomicState

	verifiedBlocks *sync.Map
	recentBlocks   *db.Cacher
	recentNames    *db.Cacher
	recentHeights  *db.Cacher
	recentData     *db.Cacher

	bb     *BlockBuilder
	txPool TxPool // Proposed transactions that haven't been put into a block yet
}

func NewChain(
	ctx *snow.Context,
	cfg *config.Config,
	gs *genesis.Genesis,
	baseDB database.Database,
	toEngine chan<- common.Message,
	gossipTx func(*ld.Transaction),
) *blockChain {
	pdb := db.NewPrefixDB(baseDB, dbPrefix, 512)
	s := &blockChain{
		config:            cfg,
		genesis:           gs,
		db:                baseDB,
		txPool:            NewTxPool(),
		preferred:         new(atomicBlock),
		lastAcceptedBlock: new(atomicBlock),
		state:             new(atomicState),
		verifiedBlocks:    new(sync.Map),
		blockDB:           pdb.With(blockDBPrefix),
		heightDB:          pdb.With(heightDBPrefix),
		lastAcceptedDB:    pdb.With(lastAcceptedDBPrefix),
		accountDB:         pdb.With(accountDBPrefix),
		ledgerDB:          pdb.With(ledgerDBPrefix),
		modelDB:           pdb.With(modelDBPrefix),
		dataDB:            pdb.With(dataDBPrefix),
		prevDataDB:        pdb.With(prevDataDBPrefix),
		stateDB:           pdb.With(stateDBPrefix),
		nameDB:            pdb.With(nameDBPrefix),
	}

	s.nameDB.SetHashKey(nameHashKey)

	txPool := NewTxPool()
	builder := NewBlockBuilder(ctx.NodeID, txPool, toEngine)

	txPool.gossipTx = gossipTx
	txPool.signalTxsReady = builder.SignalTxsReady

	s.ctx = NewContext(ctx, s, cfg, gs)
	s.bb = builder
	s.txPool = txPool
	s.preferred.StoreV(emptyBlock)
	s.lastAcceptedBlock.StoreV(emptyBlock)
	s.state.StoreV(0)

	// this data will not change, so we can cache it
	s.recentBlocks = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(Block)
	})
	s.recentHeights = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(db.RawObject)
	})
	s.recentNames = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(db.RawObject)
	})
	s.recentData = db.NewCacher(1_000, 60*10, func() db.Objecter {
		return new(ld.DataInfo)
	})
	return s
}

func (bc *blockChain) DB() database.Database {
	return bc.db
}

func (bc *blockChain) Context() *Context {
	return bc.ctx
}

func (bc *blockChain) Bootstrap() error {
	txs, err := bc.genesis.ToTxs()
	if err != nil {
		logging.Log.Error("stateDB.Bootstrap error: %v", err)
		return err
	}

	genesisBlock, err := NewGenesisBlock(bc.ctx, txs)
	if err != nil {
		logging.Log.Error("Bootstrap newGenesisBlock error: %v", err)
		return err
	}
	if genesisBlock.Parent() != ids.Empty ||
		genesisBlock.ID() == ids.Empty ||
		genesisBlock.Height() != 0 ||
		genesisBlock.Timestamp2() != 0 {
		return fmt.Errorf("Bootstrap invalid genesis block")
	}

	bc.genesisBlock = genesisBlock
	lastAcceptedID, err := database.GetID(bc.lastAcceptedDB, lastAcceptedKey)
	// create genesis block
	if err == database.ErrNotFound {
		logging.Log.Info("Bootstrap Create Genesis Block: %s", genesisBlock.ID())
		data, _ := genesisBlock.MarshalJSON()
		logging.Log.Info("genesisBlock:\n%s", string(data))
		logging.Log.Info("Bootstrap commit Genesis Block")
		bc.preferred.StoreV(genesisBlock)
		if err := genesisBlock.Accept(); err != nil {
			logging.Log.Error("Accept genesis block: %v", err)
			return fmt.Errorf("Accept genesis block error: %v", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("load last_accepted error: %v", err)
	}

	// verify genesis data
	genesisID, err := bc.GetBlockIDAtHeight(0)
	if err != nil {
		return fmt.Errorf("load genesis id error: %v", err)
	}
	// not the one on blockchain, means that the genesis data changed
	if genesisID != genesisBlock.ID() {
		return fmt.Errorf("invalid genesis data, expected genesis id %s", genesisID)
	}

	// genesis block is the last accepted block.
	if lastAcceptedID == genesisBlock.ID() {
		logging.Log.Info("Bootstrap finished at the genesis block %s", lastAcceptedID)
		genesisBlock.InitState(genesisBlock, bc.db)
		bc.preferred.StoreV(genesisBlock)
		bc.lastAcceptedBlock.StoreV(genesisBlock)
		return nil
	}

	// load the last accepted block
	lastAcceptedBlock, err := bc.GetBlock(lastAcceptedID)
	if err != nil {
		return fmt.Errorf("load last accepted block error: %v", err)
	}

	parent, err := bc.GetBlock(lastAcceptedBlock.Parent())
	if err != nil {
		return fmt.Errorf("load last accepted block' parent error: %v", err)
	}

	lastAcceptedBlock.InitState(parent, bc.db)
	lastAcceptedBlock.SetStatus(choices.Accepted)
	bc.preferred.StoreV(lastAcceptedBlock)
	bc.lastAcceptedBlock.StoreV(lastAcceptedBlock)

	// load latest fee config from chain.
	var di *ld.DataInfo
	feeConfigID := bc.genesis.Chain.FeeConfigID
	di, err = bc.LoadData(feeConfigID)
	if err != nil {
		return fmt.Errorf("load last fee config error: %v", err)
	}
	cfg, err := bc.genesis.Chain.AppendFeeConfig(di.Data)
	if err != nil {
		return fmt.Errorf("unmarshal fee config error: %v", err)
	}

	for di.Version > 1 && cfg.StartHeight >= lastAcceptedBlock.ld.Height {
		di, err = bc.LoadPrevData(feeConfigID, di.Version-1)
		if err != nil {
			return fmt.Errorf("load previous fee config error: %v", err)
		}
		cfg, err = bc.genesis.Chain.AppendFeeConfig(di.Data)
		if err != nil {
			return fmt.Errorf("unmarshal fee config error: %v", err)
		}
	}

	logging.Log.Info("Bootstrap load %d versions fee configs", len(bc.genesis.Chain.FeeConfigs))
	logging.Log.Info("Bootstrap finished at the block %s, %d", lastAcceptedBlock.ID(), lastAcceptedBlock.ld.Height)
	return nil
}

func (bc *blockChain) HealthCheck() (interface{}, error) {
	id, err := database.GetID(bc.lastAcceptedDB, lastAcceptedKey)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"LastAccepted": id.String(),
	}, nil
}

func (bc *blockChain) TotalSupply() *big.Int {
	t := new(big.Int)
	if acc, err := bc.LoadAccount(constants.LDCAccount); err == nil {
		t.Sub(acc.MaxTotalSupply, acc.Balance)
	}
	return t
}

func (bc *blockChain) SetState(state snow.State) error {
	switch state {
	case snow.Bootstrapping:
		bc.state.StoreV(state)
		return nil
	case snow.NormalOp:
		if bc.preferred.LoadV() == emptyBlock {
			return fmt.Errorf("Verify bootstrap failed")
		}
		bc.state.StoreV(state)
		return nil
	default:
		return snow.ErrUnknownState
	}
}

func (bc *blockChain) IsBootstrapped() bool {
	return bc.state.LoadV() == snow.NormalOp
}

// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (bc *blockChain) LastAcceptedBlock() *Block {
	return bc.lastAcceptedBlock.LoadV()
}

func (bc *blockChain) AddVerifiedBlock(blk *Block) {
	bc.verifiedBlocks.Store(blk.ID(), blk)
}

func (bc *blockChain) GetVerifiedBlock(id ids.ID) *Block {
	if v, ok := bc.verifiedBlocks.Load(id); ok {
		return v.(*Block)
	}
	return nil
}

func (bc *blockChain) SetLastAccepted(blk *Block) error {
	if parent := bc.lastAcceptedBlock.LoadV(); parent.ID() != blk.Parent() {
		return fmt.Errorf("stateDB.SetLastAccepted invalid parent, expected %s:%d, got %s:%d",
			parent.ID(), parent.Height(), blk.Parent(), blk.Height())
	}

	id := blk.ID()
	height := blk.Height()
	preferred := bc.preferred.LoadV()
	if preferred.ID() != id {
		logging.Log.Debug("Accepting block in non-canonical chain, expected %s:%d, got %s:%d",
			preferred.ID(), preferred.Height(), id, height)

		switch {
		case preferred.Height() <= height:
			if err := bc.setPreference(preferred, blk); err != nil {
				return err
			}

		default:
			canonical, err := preferred.State().GetBlockIDAtHeight(height)
			if err != nil {
				return err
			}
			if canonical != id {
				if err := bc.setPreference(preferred, blk); err != nil {
					return err
				}
			}
		}
	}

	if err := database.PutID(bc.lastAcceptedDB, lastAcceptedKey, id); err != nil {
		return err
	}

	bc.lastAcceptedBlock.StoreV(blk)
	bc.recentBlocks.SetObject(id[:], blk)

	go func() {
		bc.verifiedBlocks.Range(func(key, value any) bool {
			if b, ok := value.(*Block); ok {
				if b.Height() < height {
					b.Free()
					bc.verifiedBlocks.Delete(key)
				}
			}
			return true
		})
	}()
	return nil
}

func (bc *blockChain) PreferredBlock() *Block {
	return bc.preferred.LoadV()
}

// SetPreference persists the VM of the currently preferred block into database.
// This should always be a block that has no children known to consensus.
func (bc *blockChain) SetPreference(id ids.ID) error {
	preferred := bc.preferred.LoadV()
	if preferred.ID() == id {
		return nil
	}

	blk := bc.GetVerifiedBlock(id)
	if blk == nil {
		return fmt.Errorf("SetPreference block %s not verified", id)
	}

	return bc.setPreference(preferred, blk)
}
func (bc *blockChain) setPreference(preferred, blk *Block) error {
	if blk.Parent() != preferred.ID() {
		if err := bc.reorg(preferred, blk); err != nil {
			return err
		}
	}

	bc.preferred.StoreV(blk)
	bc.bb.HandlePreferenceBlock()
	logging.Log.Info("SetPreference OK %s", blk.ID())
	return nil
}

// reorg takes two blocks, an old chain and a new chain and will reconstruct the blocks.
func (bc *blockChain) reorg(oldBlock, newBlock *Block) error {
	accepted := bc.lastAcceptedBlock.LoadV()
	newChain, err := newBlock.AncestorBlocks(accepted.Height())
	if err != nil {
		return err
	}
	if newChain[0].ID() != accepted.ID() {
		return fmt.Errorf("reorg: new chain does not start with the last accepted block")
	}

	set := make(map[uint64]ids.ID, len(newChain))
	for _, blk := range newChain {
		set[blk.Height()] = blk.ID()
	}

	// new chain is longer
	if set[oldBlock.Height()] == oldBlock.ID() {
		return nil
	}

	oldChain, err := oldBlock.AncestorBlocks(accepted.Height())
	if err != nil {
		return err
	}

	for len(oldChain) > 0 {
		blk := oldChain[0]
		oldChain = oldChain[1:]
		if set[blk.Height()] == blk.ID() {
			continue
		}
		blk.Reject()
	}
	oldBlock.Reject()
	return nil
}

func (bc *blockChain) BuildBlock() (*Block, error) {
	if !bc.IsBootstrapped() {
		return nil, fmt.Errorf("stateDB.BuildBlock: state not bootstrapped")
	}

	blk, err := bc.bb.Build(bc.ctx)
	if err != nil {
		return nil, err
	}
	id := blk.ID()
	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) ParseBlock(data []byte) (*Block, error) {
	id := ids.ID(util.HashFromData(data))
	blk, err := bc.GetBlock(id)
	if err == nil {
		return blk, nil
	}

	blk = new(Block)
	if err := blk.Unmarshal(data); err != nil {
		return nil, err
	}

	if id != blk.ID() {
		return nil, fmt.Errorf("blockChain.ParseBlock: invalid block id at %d, expected %s, got %s",
			blk.Height(), id, blk.ID())
	}

	parent, err := bc.GetBlock(blk.Parent())
	if err != nil {
		return nil, err
	}

	blk.SetContext(bc.ctx)
	blk.InitState(parent, parent.State().VersionDB())

	txIDs := blk.TxIDs()
	bc.txPool.ClearTxs(txIDs...)
	bc.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (bc *blockChain) GetBlock(id ids.ID) (*Block, error) {
	if bc.genesisBlock.ID() == id {
		return bc.genesisBlock, nil
	}

	last := bc.lastAcceptedBlock.LoadV()
	if last.ID() == id {
		return last, nil
	}

	if blk := bc.GetVerifiedBlock(id); blk != nil {
		return blk, nil
	}

	obj, err := bc.blockDB.LoadObject(id[:], bc.recentBlocks)
	if err != nil {
		return nil, err
	}

	blk := obj.(*Block)
	blk.SetContext(bc.ctx)

	if blk.Status() == choices.Unknown {
		if blk.Height() > last.Height() {
			blk.SetStatus(choices.Processing)
		} else {
			id, err := bc.GetBlockIDAtHeight(blk.Height())
			switch err {
			case nil:
				if id == blk.ID() {
					blk.SetStatus(choices.Accepted)
				} else {
					blk.SetStatus(choices.Rejected)
				}
			case database.ErrNotFound:
				blk.SetStatus(choices.Processing)
			default:
				return nil, err
			}
		}
	}
	return blk, nil
}

func (bc *blockChain) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	obj, err := bc.heightDB.LoadObject(database.PackUInt64(height), bc.recentHeights)
	if err != nil {
		return ids.Empty, err
	}

	data := obj.(*db.RawObject)
	return ids.ToID(*data)
}

func (bc *blockChain) GetBlockAtHeight(height uint64) (*Block, error) {
	id, err := bc.GetBlockIDAtHeight(height)
	if err != nil {
		return nil, err
	}
	return bc.GetBlock(id)
}

// SubmitTx processes a transaction from API server
func (bc *blockChain) SubmitTx(txs ...*ld.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	for _, tx := range txs {
		if err := tx.SyntacticVerify(); err != nil {
			return err
		}
	}

	blk := bc.preferred.LoadV()
	if err := blk.TryBuildTxs(txs...); err != nil {
		return err
	}

	return bc.AddRemoteTxs(txs...)
}

func (bc *blockChain) AddRemoteTxs(txs ...*ld.Transaction) error {
	var err error
	tx := txs[0]
	if len(txs) > 1 {
		tx, err = ld.NewBatchTx(txs...)
		if err != nil {
			return err
		}
	}

	if tx.Type == ld.TypeTest {
		return fmt.Errorf("TestTx should be in a batch transactions")
	}
	bc.txPool.AddRemote(tx)
	return nil
}

func (bc *blockChain) AddLocalTxs(txs ...*ld.Transaction) {
	bc.txPool.AddLocal(txs...)
}

func (bc *blockChain) SetTxsHeight(height uint64, txIDs ...ids.ID) {
	bc.txPool.SetTxsHeight(height, txIDs...)
}

func (bc *blockChain) GetTxHeight(id ids.ID) int64 {
	return bc.txPool.GetHeight(id)
}

func (bc *blockChain) ResolveName(name string) (*service.Name, error) {
	dn, err := service.NewDN(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name %q, error: %v", name, err)
	}
	obj, err := bc.nameDB.LoadObject([]byte(dn.ASCII()), bc.recentNames)
	if err != nil {
		return nil, err
	}

	data := obj.(*db.RawObject)
	id, err := ids.ToShortID(*data)
	if err != nil {
		return nil, err
	}
	di, err := bc.LoadData(util.DataID(id))
	if err != nil {
		return nil, err
	}
	ns := &service.Name{}
	if err := ns.Unmarshal(di.Data); err != nil {
		return nil, err
	}
	if err := ns.SyntacticVerify(); err != nil {
		return nil, err
	}
	ns.DID = di.ID
	return ns, nil
}

func (bc *blockChain) LoadAccount(id util.EthID) (*ld.Account, error) {
	blk := bc.LastAcceptedBlock()
	acc, err := blk.State().LoadAccount(id)
	if err != nil {
		return nil, err
	}

	rt := &ld.Account{}
	if err = ld.Copy(rt, acc.LD()); err != nil {
		return nil, err
	}
	rt.Height = blk.Height()
	rt.Timestamp = blk.Timestamp2()
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadModel(id util.ModelID) (*ld.ModelInfo, error) {
	blk := bc.LastAcceptedBlock()
	mi, err := blk.State().LoadModel(id)
	if err != nil {
		return nil, err
	}

	rt := &ld.ModelInfo{}
	if err = ld.Copy(rt, mi); err != nil {
		return nil, err
	}
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadData(id util.DataID) (*ld.DataInfo, error) {
	blk := bc.LastAcceptedBlock()
	di, err := blk.State().LoadData(id)
	if err != nil {
		return nil, err
	}

	rt := &ld.DataInfo{}
	if err = ld.Copy(rt, di); err != nil {
		return nil, err
	}
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadPrevData(id util.DataID, version uint64) (*ld.DataInfo, error) {
	if version == 0 {
		return nil, fmt.Errorf("data not found")
	}

	obj, err := bc.prevDataDB.LoadObject(id.VersionKey(version), bc.recentData)
	if err != nil {
		return nil, err
	}
	rt := obj.(*ld.DataInfo)
	rt.ID = id
	return rt, nil
}

func (bc *blockChain) LoadRawData(rawType string, key []byte) ([]byte, error) {
	var pdb *db.PrefixDB
	switch rawType {
	case "block":
		pdb = bc.blockDB
	case "state":
		pdb = bc.stateDB
	case "account":
		pdb = bc.accountDB
	case "ledger":
		pdb = bc.ledgerDB
	case "model":
		pdb = bc.modelDB
	case "data":
		pdb = bc.dataDB
	case "prevdata":
		pdb = bc.prevDataDB
	case "name":
		pdb = bc.nameDB
	default:
		return nil, fmt.Errorf("unknown type %q", rawType)
	}

	return pdb.Get(key)
}

type atomicBlock atomic.Value

func (a *atomicBlock) LoadV() *Block {
	return (*atomic.Value)(a).Load().(*Block)
}

func (a *atomicBlock) StoreV(v *Block) {
	(*atomic.Value)(a).Store(v)
}

type atomicState atomic.Value

func (a *atomicState) LoadV() snow.State {
	v := (*atomic.Value)(a).Load().(*snow.State)
	return *v
}

func (a *atomicState) StoreV(v snow.State) {
	(*atomic.Value)(a).Store(&v)
}

func nameHashKey(key []byte) []byte {
	k := sha3.Sum256(key)
	return k[:]
}

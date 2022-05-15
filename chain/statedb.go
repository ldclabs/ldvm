// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"golang.org/x/net/idna"

	"github.com/ldclabs/ldvm/config"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/db"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/logging"
	"github.com/ldclabs/ldvm/util"
)

var _ StateDB = &stateDB{}

var (
	// Should never happen
	errPreferredBlock = fmt.Errorf("stateDB is not bootstrapped, no preferred block")
)

var (
	dbPrefix             = []byte("LDVM")
	lastAcceptedDBPrefix = []byte{'K'}
	heightDBPrefix       = []byte{'H'}
	blockDBPrefix        = []byte{'B'}
	accountDBPrefix      = []byte{'A'}
	modelDBPrefix        = []byte{'M'}
	dataDBPrefix         = []byte{'D'}
	prevDataDBPrefix     = []byte{'P'}
	nameDBPrefix         = []byte{'N'} // inverted index

	lastAcceptedKey = []byte("last_accepted_key")
)

// StateDB defines methods to manage state with Blocks and LastAcceptedIDs.
type StateDB interface {
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
	GetBlock(ids.ID) (*Block, error)
	LastAcceptedBlock() *ld.Block
	SetLastAccepted(*Block) error
	PreferredBlock() *Block
	SetPreference(ids.ID) error
	AddVerifiedBlock(*Block)

	// txs state
	SubmitTx(...*ld.Transaction) error
	AddTxs(isNew bool, txs ...*ld.Transaction)
	AddRecentTx(Transaction, choices.Status)
	GetTx(ids.ID) Transaction
	RemoveTx(ids.ID)

	LoadAccount(util.EthID) (*ld.Account, error)
	ResolveName(name string) (*ld.DataMeta, error)
	LoadModel(util.ModelID) (*ld.ModelMeta, error)
	LoadData(util.DataID) (*ld.DataMeta, error)
	LoadPrevData(util.DataID, uint64) (*ld.DataMeta, error)

	// events
	QueryEvents() []*Event
}

type atomicBlock atomic.Value

func (a *atomicBlock) LoadV() *Block {
	return (*atomic.Value)(a).Load().(*Block)
}

func (a *atomicBlock) StoreV(v *Block) {
	(*atomic.Value)(a).Store(v)
}

type atomicLDBlock atomic.Value

func (a *atomicLDBlock) LoadV() *ld.Block {
	return (*atomic.Value)(a).Load().(*ld.Block)
}

func (a *atomicLDBlock) StoreV(v *ld.Block) {
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

type stateDB struct {
	ctx          *Context
	config       *config.Config
	genesis      *genesis.Genesis
	genesisBlock *Block

	db             database.Database
	blockDB        *db.PrefixDB
	heightDB       *db.PrefixDB
	lastAcceptedDB *db.PrefixDB
	accountDB      *db.PrefixDB
	modelDB        *db.PrefixDB
	dataDB         *db.PrefixDB
	prevDataDB     *db.PrefixDB
	nameDB         *db.PrefixDB

	preferred         *atomicBlock
	lastAcceptedBlock *atomicLDBlock
	state             *atomicState

	verifiedBlocks *sync.Map
	eventsCache    *EventsCache
	recentBlocks   *db.Cacher
	recentModels   *db.Cacher
	recentData     *db.Cacher
	recentAccounts *db.Cacher
	recentNames    *db.Cacher
	recentHeights  *db.Cacher
	recentTxs      *db.Cacher

	bb          *BlockBuilder
	txPool      TxPool // Proposed transactions that haven't been put into a block yet
	gossipTx    func(tx *ld.Transaction)
	notifyBuild func()
}

func NewState(
	ctx *snow.Context,
	cfg *config.Config,
	gs *genesis.Genesis,
	baseDB database.Database,
	notifyBuild func(),
	gossipTx func(tx *ld.Transaction),
) *stateDB {
	pdb := db.NewPrefixDB(baseDB, dbPrefix, 512)
	s := &stateDB{
		config:            cfg,
		genesis:           gs,
		db:                baseDB,
		notifyBuild:       notifyBuild,
		gossipTx:          gossipTx,
		txPool:            NewTxPool(),
		preferred:         new(atomicBlock),
		lastAcceptedBlock: new(atomicLDBlock),
		state:             new(atomicState),
		eventsCache:       NewEventsCache(cfg.EventCacheSize),
		verifiedBlocks:    new(sync.Map),
		blockDB:           pdb.With(blockDBPrefix),
		heightDB:          pdb.With(heightDBPrefix),
		lastAcceptedDB:    pdb.With(lastAcceptedKey),
		accountDB:         pdb.With(accountDBPrefix),
		modelDB:           pdb.With(modelDBPrefix),
		dataDB:            pdb.With(dataDBPrefix),
		prevDataDB:        pdb.With(prevDataDBPrefix),
		nameDB:            pdb.With(nameDBPrefix),
	}
	s.ctx = NewContext(ctx, s, cfg, gs)
	s.bb = NewBlockBuilder(ctx.NodeID, s.txPool, s.notifyBuild)

	s.preferred.StoreV(emptyBlock)
	s.lastAcceptedBlock.StoreV(emptyBlock.ld)
	s.state.StoreV(0)

	s.recentBlocks = db.NewCacher(10_000, 60*10, func() db.Objecter {
		return new(Block)
	})
	s.recentHeights = db.NewCacher(10_000, 60*10, func() db.Objecter {
		return new(db.RawObject)
	})
	s.recentModels = db.NewCacher(10_000, 60*20, func() db.Objecter {
		return new(ld.ModelMeta)
	})
	s.recentData = db.NewCacher(100_000, 60*20, func() db.Objecter {
		return new(ld.DataMeta)
	})
	s.recentAccounts = db.NewCacher(100_000, 60*20, func() db.Objecter {
		return new(ld.Account)
	})
	s.recentNames = db.NewCacher(100_000, 60*20, func() db.Objecter {
		return new(db.RawObject)
	})
	s.recentTxs = db.NewCacher(100_000, 60*20, nil)
	return s
}

func (s *stateDB) DB() database.Database {
	return s.db
}

func (s *stateDB) Context() *Context {
	return s.ctx
}

func (s *stateDB) Bootstrap() error {
	genesisLdBlock, err := s.genesis.ToBlock()
	if err != nil {
		logging.Log.Error("Bootstrap genesis.ToBlock error: %v", err)
		return err
	}

	genesisBlock, err := NewGenesisBlock(genesisLdBlock, s.ctx)
	if err != nil {
		logging.Log.Error("Bootstrap newGenesisBlock error: %v", err)
		return err
	}
	if genesisBlock.Parent() != ids.Empty ||
		genesisBlock.ID() == ids.Empty ||
		genesisBlock.Height() != 0 ||
		genesisBlock.Timestamp().Unix() != 0 {
		return fmt.Errorf("Bootstrap invalid genesis block")
	}

	s.genesisBlock = genesisBlock
	lastAcceptedID, err := database.GetID(s.lastAcceptedDB, lastAcceptedKey)
	// create genesis block
	if err == database.ErrNotFound {
		logging.Log.Info("Bootstrap Create Genesis Block: %s", genesisBlock.ID())
		genesisBlock.InitState(s.db, false)
		data, _ := genesisBlock.MarshalJSON()
		logging.Log.Info("genesisBlock:\n%s", string(data))
		if err := genesisBlock.VerifyGenesis(); err != nil {
			logging.Log.Error("VerifyGenesis block error: %v", err)
			return fmt.Errorf("VerifyGenesis block error: %v", err)
		}

		logging.Log.Info("Bootstrap commit Genesis Block")
		if err := genesisBlock.Accept(); err != nil {
			logging.Log.Error("Accept genesis block: %v", err)
			return fmt.Errorf("Accept genesis block error: %v", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("load last_accepted failed: %v", err)
	}

	// verify genesis data
	genesisID, err := s.GetBlockIDAtHeight(0)
	if err != nil {
		return fmt.Errorf("load genesis id failed: %v", err)
	}
	// not the one on blockchain, means that the genesis data changed
	if genesisID != genesisBlock.ID() {
		return fmt.Errorf("invalid genesis data, expected genesis id %s", genesisID)
	}

	// genesis block is the last accepted block.
	if lastAcceptedID == genesisBlock.ID() {
		logging.Log.Info("Bootstrap finished at the genesis block %s", lastAcceptedID)
		genesisBlock.InitState(s.db, true)
		s.preferred.StoreV(genesisBlock)
		s.lastAcceptedBlock.StoreV(genesisBlock.ld)
		return nil
	}

	// load the last accepted block
	lastAcceptedBlock, err := s.GetBlock(lastAcceptedID)
	if err != nil {
		return fmt.Errorf("load last accepted block failed: %v", err)
	}

	lastAcceptedBlock.InitState(s.db, true)
	s.preferred.StoreV(lastAcceptedBlock)
	s.lastAcceptedBlock.StoreV(lastAcceptedBlock.ld)

	// load latest fee config from chain.
	var dm *ld.DataMeta
	feeConfigID := s.genesis.Chain.FeeConfigID
	dm, err = s.LoadData(feeConfigID)
	if err != nil {
		return fmt.Errorf("load last fee config failed: %v", err)
	}
	cfg, err := s.genesis.Chain.AppendFeeConfig(dm.Data)
	if err != nil {
		return fmt.Errorf("unmarshal fee config failed: %v", err)
	}

	for dm.Version > 1 && cfg.StartHeight >= lastAcceptedBlock.ld.Height {
		dm, err = s.LoadPrevData(feeConfigID, dm.Version-1)
		if err != nil {
			return fmt.Errorf("load previous fee config failed: %v", err)
		}
		cfg, err = s.genesis.Chain.AppendFeeConfig(dm.Data)
		if err != nil {
			return fmt.Errorf("unmarshal fee config failed: %v", err)
		}
	}

	logging.Log.Info("Bootstrap load %d versions fee configs", len(s.genesis.Chain.FeeConfigs))
	logging.Log.Info("Bootstrap finished at the block %s, %d", lastAcceptedBlock.ID(), lastAcceptedBlock.ld.Height)
	return nil
}

func (s *stateDB) HealthCheck() (interface{}, error) {
	id, err := database.GetID(s.lastAcceptedDB, lastAcceptedKey)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"LastAccepted": id.String(),
	}, nil
}

func (s *stateDB) TotalSupply() *big.Int {
	t := new(big.Int)
	if acc, err := s.LoadAccount(constants.LDCAccount); err == nil {
		t.Sub(acc.MaxTotalSupply, acc.Balance)
	}
	return t
}

func (s *stateDB) SetState(state snow.State) error {
	switch state {
	case snow.Bootstrapping:
		s.state.StoreV(state)
		return nil
	case snow.NormalOp:
		if s.preferred.LoadV() == emptyBlock {
			return fmt.Errorf("Verify bootstrap failed")
		}
		s.state.StoreV(state)
		return nil
	default:
		return snow.ErrUnknownState
	}
}

func (s *stateDB) IsBootstrapped() bool {
	return s.state.LoadV() == snow.NormalOp
}

// LastAccepted returns the ID of the last accepted block.
// If no blocks have been accepted by consensus yet, it is assumed there is
// a definitionally accepted block, the Genesis block, that will be
// returned.
func (s *stateDB) LastAcceptedBlock() *ld.Block {
	return s.lastAcceptedBlock.LoadV()
}

func (s *stateDB) AddVerifiedBlock(blk *Block) {
	s.verifiedBlocks.Store(blk.ID(), blk)
}

func (s *stateDB) SetLastAccepted(blk *Block) error {
	id := blk.ID()
	if err := database.PutID(s.lastAcceptedDB, lastAcceptedKey, id); err != nil {
		return err
	}

	s.lastAcceptedBlock.StoreV(blk.ld)
	height := blk.Height()
	if p := s.preferred.LoadV(); height > 0 && p.ID() != blk.ID() {
		switch {
		case p.Height() < height:
			s.preferred.StoreV(blk)
		case p.Height() == height:
			p.Reject()
			s.preferred.StoreV(blk)
		default:
			ancestors, err := p.AncestorBlocks(height)
			switch {
			case err != nil || len(ancestors) == 0:
				p.Reject()
				s.preferred.StoreV(blk)
			case ancestors[len(ancestors)-1].ID() != id:
				p.Reject()
				for _, v := range ancestors {
					v.Reject()
				}
				s.preferred.StoreV(blk)
			}
		}
	}

	s.recentBlocks.SetObject(id[:], blk)
	go func() {
		s.verifiedBlocks.Range(func(key, value any) bool {
			if b, ok := value.(*Block); ok {
				if b.Height() < height {
					s.verifiedBlocks.Delete(key)
				}
			}
			return true
		})
		s.eventsCache.Add(blk.State().Events()...)
	}()
	return nil
}

func (s *stateDB) PreferredBlock() *Block {
	return s.preferred.LoadV()
}

// SetPreference persists the VM of the currently preferred block into database.
// This should always be a block that has no children known to consensus.
func (s *stateDB) SetPreference(id ids.ID) error {
	pid := s.preferred.LoadV().ID()
	if pid == id {
		return nil
	}

	v, ok := s.verifiedBlocks.Load(id)
	if !ok {
		return fmt.Errorf("SetPreference block %s not verified", id)
	}
	blk := v.(*Block)

	if blk.Parent() != pid {
		if lid := s.lastAcceptedBlock.LoadV().ID; blk.Parent() != lid {
			return fmt.Errorf("SetPreference block %s parent error: expected %s, got %s",
				id, lid, blk.Parent())
		}
	}

	s.preferred.StoreV(blk)
	logging.Log.Info("SetPreference OK %s", id)
	return nil
}

func (s *stateDB) BuildBlock() (*Block, error) {
	if !s.bb.NeedBuild() {
		return nil, fmt.Errorf("wait to build block")
	}

	blk, err := s.bb.Build(s.ctx, s.preferred.LoadV())
	if err != nil {
		return nil, err
	}
	id := blk.ID()
	s.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (s *stateDB) ParseBlock(data []byte) (*Block, error) {
	blk := new(Block)
	if err := blk.Unmarshal(data); err != nil {
		return nil, err
	}
	id := blk.ID()
	if blk2, err := s.GetBlock(id); err == nil {
		return blk2, nil
	}

	if blk.Context() == nil {
		blk.SetContext(s.ctx)
	}
	s.recentBlocks.SetObject(id[:], blk)
	return blk, nil
}

func (s *stateDB) GetBlock(id ids.ID) (*Block, error) {
	if s.genesisBlock.ID() == id {
		return s.genesisBlock, nil
	}

	if blk, ok := s.verifiedBlocks.Load(id); ok {
		return blk.(*Block), nil
	}

	obj, err := s.blockDB.LoadObject(id[:], s.recentBlocks)
	if err != nil {
		return nil, err
	}
	blk := obj.(*Block)

	if blk.Context() == nil {
		blk.SetContext(s.ctx)
	}
	if blk.Status() != choices.Accepted && blk.Height() <= s.lastAcceptedBlock.LoadV().Height {
		blk.SetStatus(choices.Accepted)
	}
	return blk, nil
}

func (s *stateDB) GetBlockIDAtHeight(height uint64) (ids.ID, error) {
	obj, err := s.heightDB.LoadObject(database.PackUInt64(height), s.recentHeights)
	if err != nil {
		return ids.Empty, err
	}

	data := obj.(*db.RawObject)
	return ids.ToID(*data)
}

// SubmitTx processes a transaction from API server
func (s *stateDB) SubmitTx(txs ...*ld.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	var err error
	tx := txs[0]
	if len(txs) > 1 {
		tx, err = ld.NewBatchTx(txs...)
		if err != nil {
			return err
		}
	}

	if tx.Type == ld.TypeTest {
		return fmt.Errorf("TestTx should be in a batch transactions.")
	}
	blk := s.preferred.LoadV()
	if err := blk.TryVerifyTxs(txs...); err != nil {
		return err
	}

	s.AddTxs(true, tx)
	go s.notifyBuild()
	return nil
}

func (s *stateDB) AddTxs(isNew bool, txs ...*ld.Transaction) {
	if isNew {
		now := uint64(time.Now().Unix())
		for i := range txs {
			txs[i].AddedTime = now
		}
	}
	s.txPool.Add(txs...)
}

func (s *stateDB) AddRecentTx(tx Transaction, status choices.Status) {
	id := tx.ID()
	tx.SetStatus(status)
	s.recentTxs.SetObject(id[:], tx)
}

func (s *stateDB) GetTx(id ids.ID) Transaction {
	if tx := s.txPool.Get(id); tx != nil {
		return tx
	}
	if tx, ok := s.recentTxs.GetObject(id[:]); ok {
		return tx.(Transaction)
	}
	return nil
}

func (s *stateDB) RemoveTx(id ids.ID) {
	s.txPool.Remove(id)
}

func (s *stateDB) LoadAccount(id util.EthID) (*ld.Account, error) {
	obj, err := s.accountDB.LoadObject(id[:], s.recentAccounts)
	if err != nil {
		return nil, err
	}
	blk := s.LastAcceptedBlock()
	rt := obj.(*ld.Account)
	rt.Height = blk.Height
	rt.Timestamp = blk.Timestamp
	rt.ID = id
	return rt, nil
}

func (s *stateDB) ResolveName(name string) (*ld.DataMeta, error) {
	dn, err := idna.Registration.ToASCII(name)
	if err != nil {
		return nil, fmt.Errorf("invalid name %s, error: %v",
			strconv.Quote(name), err)
	}
	obj, err := s.nameDB.LoadObject([]byte(dn), s.recentNames)
	if err != nil {
		return nil, err
	}

	data := obj.(*db.RawObject)
	id, err := ids.ToShortID(*data)
	if err != nil {
		return nil, err
	}
	return s.LoadData(util.DataID(id))
}

func (s *stateDB) LoadModel(id util.ModelID) (*ld.ModelMeta, error) {
	obj, err := s.modelDB.LoadObject(id[:], s.recentModels)
	if err != nil {
		return nil, err
	}
	rt := obj.(*ld.ModelMeta)
	rt.ID = id
	return rt, nil
}

func (s *stateDB) LoadData(id util.DataID) (*ld.DataMeta, error) {
	obj, err := s.dataDB.LoadObject(id[:], s.recentData)
	if err != nil {
		return nil, err
	}
	rt := obj.(*ld.DataMeta)
	rt.ID = id
	return rt, nil
}

func (s *stateDB) LoadPrevData(id util.DataID, version uint64) (*ld.DataMeta, error) {
	if version == 0 {
		return nil, fmt.Errorf("data not found")
	}

	v := database.PackUInt64(version)
	key := make([]byte, 20+len(v))
	copy(key, id[:])
	copy(key[20:], v)

	obj, err := s.prevDataDB.LoadObject(key, s.recentData)
	if err != nil {
		return nil, err
	}
	rt := obj.(*ld.DataMeta)
	rt.ID = id
	return rt, nil
}

func (s *stateDB) QueryEvents() []*Event {
	return s.eventsCache.Query()
}

// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type MBCtx struct {
	ChainConfig *genesis.ChainConfig
	Price       uint64
	MinerID     util.StakeSymbol
}

func NewMBCtx() *MBCtx {
	cfg, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		panic(err)
	}
	return &MBCtx{
		ChainConfig: &cfg.Chain,
		Price:       1000,
		MinerID:     ld.MustNewStake("#LDC"),
	}
}

func (m *MBCtx) Chain() *genesis.ChainConfig {
	return m.ChainConfig
}

func (m *MBCtx) FeeConfig() *genesis.FeeConfig {
	return m.ChainConfig.FeeConfig
}

func (m *MBCtx) GasPrice() *big.Int {
	return new(big.Int).SetUint64(m.Price)
}

func (m *MBCtx) Miner() util.StakeSymbol {
	return m.MinerID
}

type MBS struct {
	Height, Timestamp uint64
	Fee               *genesis.FeeConfig
	AC                AccountCache
	NC                map[string]util.DataID
	MC                map[util.ModelID]*ld.ModelMeta
	DC                map[util.DataID]*ld.DataMeta
	PDC               map[util.DataID]*ld.DataMeta
}

func NewMBS() *MBS {
	cfg, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		panic(err)
	}
	return &MBS{
		Fee: cfg.Chain.FeeConfig,
		AC:  make(AccountCache),
		NC:  make(map[string]util.DataID),
		MC:  make(map[util.ModelID]*ld.ModelMeta),
		DC:  make(map[util.DataID]*ld.DataMeta),
		PDC: make(map[util.DataID]*ld.DataMeta),
	}
}

func (m *MBS) LoadAccount(id util.EthID) (*Account, error) {
	acc := m.AC[id]
	if acc == nil {
		acc = NewAccount(id)
		pledge := new(big.Int)
		switch {
		case acc.Type() == ld.TokenAccount && id != constants.LDCAccount:
			pledge.Set(m.Fee.MinTokenPledge)
		case acc.Type() == ld.StakeAccount:
			pledge.Set(m.Fee.MinStakePledge)
		}

		acc.Init(pledge, m.Height, m.Timestamp)
		m.AC[id] = acc
	}

	return m.AC[id], nil
}

func (m *MBS) LoadMiner(id util.StakeSymbol) (*Account, error) {
	miner := constants.GenesisAccount
	if id != util.StakeEmpty && id.Valid() {
		miner = util.EthID(id)
	}
	return m.LoadAccount(miner)
}

func (m *MBS) ResolveNameID(name string) (util.DataID, error) {
	id, ok := m.NC[name]
	if !ok {
		return util.DataIDEmpty, fmt.Errorf("MBS.ResolveNameID: %s not found", strconv.Quote(name))
	}
	return id, nil
}

func (m *MBS) ResolveName(name string) (*ld.DataMeta, error) {
	id, err := m.ResolveNameID(name)
	if err != nil {
		return nil, err
	}
	return m.LoadData(id)
}

func (m *MBS) SetName(name string, id util.DataID) error {
	m.NC[name] = id
	return nil
}

func (m *MBS) LoadModel(id util.ModelID) (*ld.ModelMeta, error) {
	mm, ok := m.MC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadModel: %s not found", id)
	}
	return mm, nil
}

func (m *MBS) SaveModel(id util.ModelID, mm *ld.ModelMeta) error {
	m.MC[id] = mm
	return nil
}

func (m *MBS) LoadData(id util.DataID) (*ld.DataMeta, error) {
	dm, ok := m.DC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadData: %s not found", id)
	}
	return dm, nil
}

func (m *MBS) SaveData(id util.DataID, dm *ld.DataMeta) error {
	m.DC[id] = dm
	return nil
}

func (m *MBS) SavePrevData(id util.DataID, dm *ld.DataMeta) error {
	m.PDC[id] = dm
	return nil
}

func (m *MBS) DeleteData(id util.DataID, dm *ld.DataMeta) error {
	dm.Version = 0
	m.DC[id] = dm
	delete(m.PDC, id)
	return nil
}

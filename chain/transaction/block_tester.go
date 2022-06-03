// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type MockBCtx struct {
	ChainConfig       *genesis.ChainConfig
	Height, Timestamp uint64
	Price             uint64
	MinerID           util.StakeSymbol
}

func NewMockBCtx() *MockBCtx {
	cfg, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		panic(err)
	}
	return &MockBCtx{
		ChainConfig: &cfg.Chain,
		Height:      1,
		Timestamp:   1000,
		Price:       1000,
		MinerID:     ld.MustNewStake("#LDC"),
	}
}

func (m *MockBCtx) Chain() *genesis.ChainConfig {
	return m.ChainConfig
}

func (m *MockBCtx) FeeConfig() *genesis.FeeConfig {
	return m.ChainConfig.FeeConfig
}

func (m *MockBCtx) GasPrice() *big.Int {
	return new(big.Int).SetUint64(m.Price)
}

func (m *MockBCtx) Miner() util.StakeSymbol {
	return m.MinerID
}

type MockBS struct {
	Height, Timestamp uint64
	Fee               *genesis.FeeConfig
	AC                AccountCache
	NC                map[string]util.DataID
	MC                map[util.ModelID]*ld.ModelMeta
	DC                map[util.DataID]*ld.DataMeta
	PDC               map[util.DataID]*ld.DataMeta
}

func NewMockBS(m *MockBCtx) *MockBS {
	return &MockBS{
		Height:    m.Height,
		Timestamp: m.Timestamp,
		Fee:       m.ChainConfig.FeeConfig,
		AC:        make(AccountCache),
		NC:        make(map[string]util.DataID),
		MC:        make(map[util.ModelID]*ld.ModelMeta),
		DC:        make(map[util.DataID]*ld.DataMeta),
		PDC:       make(map[util.DataID]*ld.DataMeta),
	}
}

func (m *MockBS) LoadAccount(id util.EthID) (*Account, error) {
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

func (m *MockBS) LoadMiner(id util.StakeSymbol) (*Account, error) {
	miner := constants.GenesisAccount
	if id != util.StakeEmpty && id.Valid() {
		miner = util.EthID(id)
	}
	return m.LoadAccount(miner)
}

func (m *MockBS) ResolveNameID(name string) (util.DataID, error) {
	id, ok := m.NC[name]
	if !ok {
		return util.DataIDEmpty, fmt.Errorf("MBS.ResolveNameID: %s not found", strconv.Quote(name))
	}
	return id, nil
}

func (m *MockBS) ResolveName(name string) (*ld.DataMeta, error) {
	id, err := m.ResolveNameID(name)
	if err != nil {
		return nil, err
	}
	return m.LoadData(id)
}

func (m *MockBS) SetName(name string, id util.DataID) error {
	m.NC[name] = id
	return nil
}

func (m *MockBS) LoadModel(id util.ModelID) (*ld.ModelMeta, error) {
	mm, ok := m.MC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadModel: %s not found", id)
	}
	return mm, nil
}

func (m *MockBS) SaveModel(id util.ModelID, mm *ld.ModelMeta) error {
	m.MC[id] = mm
	return nil
}

func (m *MockBS) LoadData(id util.DataID) (*ld.DataMeta, error) {
	dm, ok := m.DC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadData: %s not found", id)
	}
	return dm, nil
}

func (m *MockBS) SaveData(id util.DataID, dm *ld.DataMeta) error {
	m.DC[id] = dm
	return nil
}

func (m *MockBS) SavePrevData(id util.DataID, dm *ld.DataMeta) error {
	m.PDC[id] = dm
	return nil
}

func (m *MockBS) DeleteData(id util.DataID, dm *ld.DataMeta, message []byte) error {
	if err := dm.MarkDeleted(message); err != nil {
		return err
	}
	m.DC[id] = dm
	delete(m.PDC, id)
	return nil
}

func (m *MockBS) VerifyState() error {
	for k, v := range m.AC {
		data, err := v.Marshal()
		if err != nil {
			return err
		}
		acc, err := ParseAccount(k, data)
		if err != nil {
			return err
		}
		data2, err := acc.Marshal()
		if err != nil {
			return err
		}
		if !bytes.Equal(data, data2) {
			return fmt.Errorf("Account %s is invalid", k)
		}
	}

	for k, v := range m.MC {
		data, err := v.Marshal()
		if err != nil {
			return err
		}
		mm := &ld.ModelMeta{}
		if err := mm.Unmarshal(data); err != nil {
			return err
		}
		if err := mm.SyntacticVerify(); err != nil {
			return err
		}
		if !bytes.Equal(data, mm.Bytes()) {
			return fmt.Errorf("ModelMeta %s is invalid", k)
		}
	}

	for k, v := range m.DC {
		data, err := v.Marshal()
		if err != nil {
			return err
		}
		dm := &ld.DataMeta{}
		if err := dm.Unmarshal(data); err != nil {
			return err
		}
		if err := dm.SyntacticVerify(); err != nil {
			return err
		}
		if !bytes.Equal(data, dm.Bytes()) {
			return fmt.Errorf("DataMeta %s is invalid", k)
		}
	}

	for k, v := range m.PDC {
		data, err := v.Marshal()
		if err != nil {
			return err
		}
		dm := &ld.DataMeta{}
		if err := dm.Unmarshal(data); err != nil {
			return err
		}
		if err := dm.SyntacticVerify(); err != nil {
			return err
		}
		if !bytes.Equal(data, dm.Bytes()) {
			return fmt.Errorf("DataMeta %s is invalid", k)
		}
	}
	return nil
}

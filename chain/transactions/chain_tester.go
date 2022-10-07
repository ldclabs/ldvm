// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transactions

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type MockChainContext struct {
	cfg               *genesis.ChainConfig
	height, timestamp uint64
	Price             uint64
	MinerID           util.StakeSymbol
}

func NewMockChainContext() *MockChainContext {
	ge, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		panic(err)
	}
	_, err = ge.ToTxs()
	if err != nil {
		panic(err)
	}
	return &MockChainContext{
		cfg:       &ge.Chain,
		height:    1,
		timestamp: 1000,
		Price:     1000,
		MinerID:   ld.MustNewStake("#LDC"),
	}
}

func (m *MockChainContext) ChainConfig() *genesis.ChainConfig {
	return m.cfg
}

func (m *MockChainContext) FeeConfig() *genesis.FeeConfig {
	return m.cfg.FeeConfig
}

func (m *MockChainContext) GasPrice() *big.Int {
	return new(big.Int).SetUint64(m.Price)
}

func (m *MockChainContext) Miner() util.StakeSymbol {
	return m.MinerID
}

func (m *MockChainContext) MockChainState() *MockChainState {
	return &MockChainState{
		ctx: m,
		Fee: m.cfg.FeeConfig,
		AC:  make(AccountCache),
		NC:  make(map[string]util.DataID),
		MC:  make(map[util.ModelID][]byte),
		DC:  make(map[util.DataID][]byte),
		PDC: make(map[util.DataID][]byte),
		ac:  make(map[util.Address][]byte),
		al:  make(map[util.Address][]byte),
	}
}

type MockChainState struct {
	ctx *MockChainContext
	Fee *genesis.FeeConfig
	AC  AccountCache
	NC  map[string]util.DataID
	MC  map[util.ModelID][]byte
	DC  map[util.DataID][]byte
	PDC map[util.DataID][]byte
	ac  map[util.Address][]byte
	al  map[util.Address][]byte
}

func (m *MockChainState) Height() uint64 {
	return m.ctx.height
}

func (m *MockChainState) Timestamp() uint64 {
	return m.ctx.timestamp
}

func (m *MockChainState) LoadAccount(id util.Address) (*Account, error) {
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
		acc.Init(pledge, m.ctx.height, m.ctx.timestamp)
		m.AC[id] = acc
	}
	return m.AC[id], nil
}

func (m *MockChainState) LoadLedger(acc *Account) error {
	if acc.Ledger() == nil {
		return acc.InitLedger(m.al[acc.ID()])
	}
	return nil
}

func (m *MockChainState) MustAccount(id util.Address) *Account {
	acc, err := m.LoadAccount(id)
	if err != nil {
		panic(err)
	}
	return acc
}

func (m *MockChainState) CommitAccounts() {
	for id, acc := range m.AC {
		data, ledger, err := acc.Marshal()
		if err != nil {
			panic(err)
		}
		m.ac[id] = data
		if len(ledger) > 0 {
			m.al[id] = ledger
		}
	}
}

func (m *MockChainState) CheckoutAccounts() {
	for id, data := range m.ac {
		ac, err := ParseAccount(id, data)
		if err != nil {
			panic(err)
		}
		if acc, ok := m.AC[id]; ok {
			acc.ld = ac.ld
			pledge := new(big.Int)
			switch {
			case acc.pledge.Sign() > 0:
				pledge.Set(acc.pledge)
			case acc.Type() == ld.TokenAccount && id != constants.LDCAccount:
				pledge.Set(m.Fee.MinTokenPledge)
			case acc.Type() == ld.StakeAccount:
				pledge.Set(m.Fee.MinStakePledge)
			}
			acc.Init(pledge, m.ctx.height, m.ctx.timestamp)

			if al, ok := m.al[id]; ok {
				if err = acc.InitLedger(al); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (m *MockChainState) LoadMiner(id util.StakeSymbol) (*Account, error) {
	miner := constants.GenesisAccount
	if id != util.StakeEmpty && id.Valid() {
		miner = util.Address(id)
	}
	return m.LoadAccount(miner)
}

func (m *MockChainState) LoadDataByName(name string) (*ld.DataInfo, error) {
	id, ok := m.NC[name]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadDataByName: %q not found", name)
	}
	return m.LoadData(id)
}

func (m *MockChainState) SaveName(ns *service.Name) error {
	if ns.DataID == util.DataIDEmpty {
		return fmt.Errorf("MBS.SaveName: name ID is empty")
	}

	name := ns.ASCII()
	_, ok := m.NC[name]
	switch {
	case ok:
		return fmt.Errorf("name %q conflict", name)
	default:
		m.NC[name] = ns.DataID
	}
	return nil
}

func (m *MockChainState) LoadModel(id util.ModelID) (*ld.ModelInfo, error) {
	data, ok := m.MC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadModel: %s not found", id)
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

func (m *MockChainState) SaveModel(mi *ld.ModelInfo) error {
	if mi.ID == util.ModelIDEmpty {
		return fmt.Errorf("MBS.SaveModel: model ID is empty")
	}

	if err := mi.SyntacticVerify(); err != nil {
		return err
	}
	m.MC[mi.ID] = mi.Bytes()
	return nil
}

func (m *MockChainState) LoadData(id util.DataID) (*ld.DataInfo, error) {
	data, ok := m.DC[id]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadData: %s not found", id)
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

func (m *MockChainState) SaveData(di *ld.DataInfo) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("MBS.SaveData: data ID is empty")
	}
	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.DC[di.ID] = di.Bytes()
	return nil
}

func (m *MockChainState) SavePrevData(di *ld.DataInfo) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("MBS.SavePrevData: data ID is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.PDC[di.ID] = di.Bytes()
	return nil
}

func (m *MockChainState) DeleteData(di *ld.DataInfo, message []byte) error {
	if di.ID == util.DataIDEmpty {
		return fmt.Errorf("MBS.DeleteData: data ID is empty")
	}

	if err := di.MarkDeleted(message); err != nil {
		return err
	}
	if err := m.SaveData(di); err != nil {
		return err
	}
	delete(m.PDC, di.ID)
	return nil
}

func (m *MockChainState) VerifyState() error {
	for k, v := range m.AC {
		data, ledger, err := v.Marshal()
		if err != nil {
			return err
		}
		acc, err := ParseAccount(k, data)
		if err != nil {
			return err
		}
		if len(ledger) > 0 {
			if err = acc.InitLedger(ledger); err != nil {
				return err
			}
		}
		data2, ledger2, err := acc.Marshal()
		if err != nil {
			return err
		}

		if !bytes.Equal(data, data2) || !bytes.Equal(ledger, ledger2) {
			return fmt.Errorf("Account %s is invalid", k)
		}
	}
	return nil
}

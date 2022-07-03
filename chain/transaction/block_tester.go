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
	height, timestamp uint64
	Price             uint64
	MinerID           util.StakeSymbol
}

func NewMockBCtx() *MockBCtx {
	ge, err := genesis.FromJSON([]byte(genesis.LocalGenesisConfigJSON))
	if err != nil {
		panic(err)
	}
	_, err = ge.ToTxs()
	if err != nil {
		panic(err)
	}
	return &MockBCtx{
		ChainConfig: &ge.Chain,
		height:      1,
		timestamp:   1000,
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

func (m *MockBCtx) MockBS() *MockBS {
	return &MockBS{
		ctx: m,
		Fee: m.ChainConfig.FeeConfig,
		AC:  make(AccountCache),
		NC:  make(map[string]util.DataID),
		MC:  make(map[util.ModelID][]byte),
		DC:  make(map[util.DataID][]byte),
		PDC: make(map[util.DataID][]byte),
		ac:  make(map[util.EthID][]byte),
		al:  make(map[util.EthID][]byte),
	}
}

type MockBS struct {
	ctx *MockBCtx
	Fee *genesis.FeeConfig
	AC  AccountCache
	NC  map[string]util.DataID
	MC  map[util.ModelID][]byte
	DC  map[util.DataID][]byte
	PDC map[util.DataID][]byte
	ac  map[util.EthID][]byte
	al  map[util.EthID][]byte
}

func (m *MockBS) Height() uint64 {
	return m.ctx.height
}

func (m *MockBS) Timestamp() uint64 {
	return m.ctx.timestamp
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
		acc.Init(pledge, m.ctx.height, m.ctx.timestamp)
		m.AC[id] = acc
	}
	return m.AC[id], nil
}

func (m *MockBS) LoadLedger(acc *Account) error {
	if acc.Ledger() == nil {
		return acc.InitLedger(m.al[acc.ID()])
	}
	return nil
}

func (m *MockBS) MustAccount(id util.EthID) *Account {
	acc, err := m.LoadAccount(id)
	if err != nil {
		panic(err)
	}
	return acc
}

func (m *MockBS) CommitAccounts() {
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

func (m *MockBS) CheckoutAccounts() {
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

func (m *MockBS) ResolveName(name string) (*ld.DataInfo, error) {
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

func (m *MockBS) LoadModel(id util.ModelID) (*ld.ModelInfo, error) {
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
	return mi, nil
}

func (m *MockBS) SaveModel(id util.ModelID, mi *ld.ModelInfo) error {
	if err := mi.SyntacticVerify(); err != nil {
		return err
	}
	m.MC[id] = mi.Bytes()
	return nil
}

func (m *MockBS) LoadData(id util.DataID) (*ld.DataInfo, error) {
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
	return di, nil
}

func (m *MockBS) SaveData(id util.DataID, di *ld.DataInfo) error {
	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.DC[id] = di.Bytes()
	return nil
}

func (m *MockBS) SavePrevData(id util.DataID, di *ld.DataInfo) error {
	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.PDC[id] = di.Bytes()
	return nil
}

func (m *MockBS) DeleteData(id util.DataID, di *ld.DataInfo, message []byte) error {
	if err := di.MarkDeleted(message); err != nil {
		return err
	}
	if err := m.SaveData(id, di); err != nil {
		return err
	}
	delete(m.PDC, id)
	return nil
}

func (m *MockBS) VerifyState() error {
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

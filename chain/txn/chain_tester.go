//go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txn

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ldclabs/ldvm/chain/acct"
	"github.com/ldclabs/ldvm/genesis"
	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/signer"
	"github.com/ldclabs/ldvm/unit"
)

type MockChainContext struct {
	cfg               *genesis.ChainConfig
	height, timestamp uint64
	Price             uint64
	BuilderID         ids.Address
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
		BuilderID: ids.Address(ld.MustNewStake("#LDC")),
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

func (m *MockChainContext) Builder() ids.Address {
	return m.BuilderID
}

func (m *MockChainContext) MockChainState() *MockChainState {
	cs := &MockChainState{
		ctx: m,
		Fee: m.cfg.FeeConfig,
		AC:  make(acct.ActiveAccounts),
		NC:  make(map[string]ids.DataID),
		MC:  make(map[ids.ModelID][]byte),
		DC:  make(map[ids.DataID][]byte),
		PDC: make(map[ids.DataID][]byte),
		ac:  make(map[ids.Address][]byte),
		al:  make(map[ids.Address][]byte),
	}
	builder := cs.MustAccount(ids.Address(m.BuilderID))
	if err := cs.LoadLedger(builder); err != nil {
		panic(err)
	}
	builder.LD().Balance.Set(m.cfg.FeeConfig.MinStakePledge)

	if err := builder.CreateStake(
		signer.Signer1.Key().Address(),
		m.cfg.FeeConfig.MinStakePledge,
		&ld.TxAccounter{
			Threshold: ld.Uint16Ptr(1),
			Keepers:   &signer.Keys{signer.Signer1.Key()},
		},
		&ld.StakeConfig{
			LockTime:    0,
			WithdrawFee: 1,
			MinAmount:   new(big.Int).SetUint64(unit.LDC),
			MaxAmount:   new(big.Int).SetUint64(unit.LDC * 1000),
		},
	); err != nil {
		panic(err)
	}
	builder.Init(big.NewInt(0), m.cfg.FeeConfig.MinStakePledge, cs.Height(), cs.Timestamp())
	return cs
}

type MockChainState struct {
	ctx *MockChainContext
	Fee *genesis.FeeConfig
	AC  acct.ActiveAccounts
	NC  map[string]ids.DataID
	MC  map[ids.ModelID][]byte
	DC  map[ids.DataID][]byte
	PDC map[ids.DataID][]byte
	ac  map[ids.Address][]byte
	al  map[ids.Address][]byte
}

func (m *MockChainState) Height() uint64 {
	return m.ctx.height
}

func (m *MockChainState) Timestamp() uint64 {
	return m.ctx.timestamp
}

func (m *MockChainState) LoadAccount(id ids.Address) (*acct.Account, error) {
	acc := m.AC[id]
	if acc == nil {
		acc = acct.NewAccount(id)

		switch {
		case id == ids.LDCAccount || id == ids.GenesisAccount:
			acc.Init(big.NewInt(0), big.NewInt(0), m.ctx.height, m.ctx.timestamp)
		case acc.Type() == ld.TokenAccount:
			acc.Init(big.NewInt(0), m.Fee.MinTokenPledge, m.ctx.height, m.ctx.timestamp)
		case acc.Type() == ld.StakeAccount:
			acc.Init(big.NewInt(0), m.Fee.MinStakePledge, m.ctx.height, m.ctx.timestamp)
		default:
			acc.Init(m.Fee.NonTransferableBalance, big.NewInt(0), m.ctx.height, m.ctx.timestamp)
		}

		m.AC[id] = acc
	}
	return m.AC[id], nil
}

func (m *MockChainState) LoadLedger(acc *acct.Account) error {
	return acc.LoadLedger(false, func() ([]byte, error) {
		return m.al[acc.ID()], nil
	})
}

func (m *MockChainState) MustAccount(id ids.Address) *acct.Account {
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
		if acc, ok := m.AC[id]; ok {
			ac, err := acct.ParseAccount(id, data)
			if err != nil {
				panic(err)
			}
			*acc.LD() = (*ac.LD())
			switch {
			case id == ids.LDCAccount || id == ids.GenesisAccount:
				acc.Init(big.NewInt(0), big.NewInt(0), m.ctx.height, m.ctx.timestamp)
			case acc.Type() == ld.TokenAccount:
				acc.Init(big.NewInt(0), m.Fee.MinTokenPledge, m.ctx.height, m.ctx.timestamp)
			case acc.Type() == ld.StakeAccount:
				acc.Init(big.NewInt(0), m.Fee.MinStakePledge, m.ctx.height, m.ctx.timestamp)
			default:
				acc.Init(m.Fee.NonTransferableBalance, big.NewInt(0), m.ctx.height, m.ctx.timestamp)
			}

			if _, ok := m.al[id]; ok {
				if err = acc.LoadLedger(true, func() ([]byte, error) {
					return m.al[acc.ID()], nil
				}); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (m *MockChainState) LoadDataByName(name string) (*ld.DataInfo, error) {
	id, ok := m.NC[name]
	if !ok {
		return nil, fmt.Errorf("MBS.LoadDataByName: %q not found", name)
	}
	return m.LoadData(id)
}

func (m *MockChainState) SaveName(ns *service.Name) error {
	if ns.DataID == ids.EmptyDataID {
		return fmt.Errorf("MBS.SaveName: name ID is empty")
	}

	name := ns.ASCII()
	_, ok := m.NC[name]
	switch {
	case ok:
		return fmt.Errorf("name %q is conflict", name)
	default:
		m.NC[name] = ns.DataID
	}
	return nil
}

func (m *MockChainState) DeleteName(ns *service.Name) error {
	if ns.DataID == ids.EmptyDataID {
		return fmt.Errorf("MBS.DeleteName: name ID is empty")
	}

	name := ns.ASCII()
	_, ok := m.NC[name]
	switch {
	case ok:
		delete(m.NC, name)
		return nil
	default:
		return fmt.Errorf("MBS.DeleteName: name %q is not exist", name)
	}
}

func (m *MockChainState) LoadModel(id ids.ModelID) (*ld.ModelInfo, error) {
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
	if mi.ID == ids.EmptyModelID {
		return fmt.Errorf("MBS.SaveModel: model ID is empty")
	}

	if err := mi.SyntacticVerify(); err != nil {
		return err
	}
	m.MC[mi.ID] = mi.Bytes()
	return nil
}

func (m *MockChainState) LoadData(id ids.DataID) (*ld.DataInfo, error) {
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
	if di.ID == ids.EmptyDataID {
		return fmt.Errorf("MBS.SaveData: data ID is empty")
	}
	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.DC[di.ID] = di.Bytes()
	return nil
}

func (m *MockChainState) SavePrevData(di *ld.DataInfo) error {
	if di.ID == ids.EmptyDataID {
		return fmt.Errorf("MBS.SavePrevData: data ID is empty")
	}

	if err := di.SyntacticVerify(); err != nil {
		return err
	}
	m.PDC[di.ID] = di.Bytes()
	return nil
}

func (m *MockChainState) DeleteData(di *ld.DataInfo, message []byte) error {
	if di.ID == ids.EmptyDataID {
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
		acc, err := acct.ParseAccount(k, data)
		if err != nil {
			return err
		}
		if len(ledger) > 0 {
			if err = acc.LoadLedger(false, func() ([]byte, error) { return ledger, nil }); err != nil {
				return err
			}
		}
		data2, ledger2, err := acc.Marshal()
		if err != nil {
			return err
		}

		if !bytes.Equal(data, data2) || !bytes.Equal(ledger, ledger2) {
			return fmt.Errorf("account %s is invalid", k)
		}
	}
	return nil
}

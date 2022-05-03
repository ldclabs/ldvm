// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type Genesis struct {
	Chain *ChainConfig               `json:"chain"`
	Alloc map[util.EthID]*Allocation `json:"alloc"`
}

type Allocation struct {
	Balance   *big.Int     `json:"balance"`
	Threshold uint8        `json:"threshold"`
	Keepers   []util.EthID `json:"keepers"`
}

type ChainConfig struct {
	ChainID          uint64       `json:"chainID"`
	MaxTotalSupply   *big.Int     `json:"maxTotalSupply"`
	Message          string       `json:"message"`
	FeeConfigID      util.DataID  `json:"feeConfigID"`
	NameServiceID    util.ModelID `json:"nameAppID"`
	ProfileServiceID util.ModelID `json:"profileAppID"`
	FeeConfig        *FeeConfig   `json:"feeConfig"`
	FeeConfigs       []*FeeConfig `json:"feeConfigs"`
}

func (c *ChainConfig) IsNameService(id ids.ShortID) bool {
	return c.NameServiceID == util.ModelID(id)
}

func (c *ChainConfig) Fee(height uint64) *FeeConfig {
	for i, cfg := range c.FeeConfigs {
		if cfg.StartHeight <= height {
			return c.FeeConfigs[i]
		}
	}

	return c.FeeConfig
}

func (c *ChainConfig) AddFeeConfig(data []byte) (*FeeConfig, error) {
	fee := &FeeConfig{}
	if err := json.Unmarshal(data, fee); err != nil {
		return nil, err
	}
	c.FeeConfigs = append(c.FeeConfigs, fee)
	return fee, nil
}

type FeeConfig struct {
	StartHeight     uint64   `json:"startHeight"`
	ThresholdGas    uint64   `json:"thresholdGas"`
	MinGasPrice     uint64   `json:"minGasPrice"`
	MaxGasPrice     uint64   `json:"maxGasPrice"`
	MaxTxGas        uint64   `json:"maxTxGas"`
	MaxBlockTxsSize uint64   `json:"maxBlockTxsSize"`
	GasRebateRate   uint64   `json:"gasRebateRate"`
	MinTokenPledge  *big.Int `json:"minTokenPledge"`
	MinStakePledge  *big.Int `json:"minStakePledge"`
}

func FromJSON(data []byte) (*Genesis, error) {
	g := new(Genesis)
	if err := json.Unmarshal(data, g); err != nil {
		return nil, err
	}
	g.Chain.FeeConfigs = []*FeeConfig{g.Chain.FeeConfig}
	return g, nil
}

func (g *Genesis) ToBlock() (*ld.Block, error) {
	genesisAccount, ok := g.Alloc[util.EthID(constants.GenesisAccount)]
	if !ok {
		return nil, fmt.Errorf("genesis account not found")
	}

	txs := make([]*ld.Transaction, 0)
	// The first transaction is issued by the Genesis account, to create native token.
	// It has included ChainID, MaxTotalSupply and Genesis Message.
	minter := &ld.TxAccounter{
		Amount:  g.Chain.MaxTotalSupply,
		Name:    "Linked Data Chain",
		Message: g.Chain.Message,
	}
	tx := &ld.Transaction{
		Type:    ld.TypeCreateTokenAccount,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		To:      ids.ShortID(constants.LDCAccount),
		Amount:  g.Chain.MaxTotalSupply,
		Data:    minter.Bytes(),
	}
	txs = append(txs, tx)

	// Alloc Txs
	nonce := uint64(0)
	list := make([]ids.ShortID, 0, len(g.Alloc))
	for id := range g.Alloc {
		list = append(list, ids.ShortID(id))
	}
	ids.SortShortIDs(list)
	for _, id := range list {
		v := g.Alloc[util.EthID(id)]
		tx := &ld.Transaction{
			Type:    ld.TypeTransfer,
			ChainID: g.Chain.ChainID,
			Nonce:   nonce,
			From:    constants.LDCAccount,
			To:      id,
			Amount:  v.Balance,
		}
		nonce++
		txs = append(txs, tx)

		if le := len(v.Keepers); le > 0 {
			update := &ld.TxUpdater{
				Threshold: v.Threshold,
				Keepers:   util.EthIDsToShort(v.Keepers...),
			}

			tx := &ld.Transaction{
				Type:    ld.TypeUpdateAccountKeepers,
				ChainID: g.Chain.ChainID,
				From:    id,
				Data:    update.Bytes(),
			}
			txs = append(txs, tx)
		}
	}

	// config data tx
	cfg, err := json.Marshal(g.Chain.FeeConfig)
	if err != nil {
		return nil, err
	}
	cfgData := &ld.TxUpdater{
		ID:        ids.ShortID(constants.JsonModelID),
		Version:   1,
		Threshold: genesisAccount.Threshold,
		Keepers:   util.EthIDsToShort(genesisAccount.Keepers...),
		Data:      cfg,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateData,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    cfgData.Bytes(),
	}
	g.Chain.FeeConfigID = util.DataID(tx.ShortID())
	txs = append(txs, tx)

	// name app tx
	name, sch := service.NameSchema()
	nameModel := &ld.ModelMeta{
		Name:      name,
		Threshold: genesisAccount.Threshold,
		Keepers:   util.EthIDsToShort(genesisAccount.Keepers...),
		Data:      sch,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    nameModel.Bytes(),
	}
	g.Chain.NameServiceID = util.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// Profile app tx
	name, sch = service.ProfileSchema()
	profileModel := &ld.ModelMeta{
		Name:      name,
		Threshold: genesisAccount.Threshold,
		Keepers:   util.EthIDsToShort(genesisAccount.Keepers...),
		Data:      sch,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    profileModel.Bytes(),
	}
	g.Chain.ProfileServiceID = util.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// build genesis block
	return &ld.Block{
		GasPrice:      g.Chain.FeeConfig.MinGasPrice,
		GasRebateRate: g.Chain.FeeConfig.GasRebateRate,
		Txs:           txs,
	}, nil
}

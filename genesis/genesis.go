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
	"github.com/ldclabs/ldvm/ld/app"
)

type Genesis struct {
	Chain *ChainConfig `json:"chain"`
	Block struct {
		GasRebateRate uint64 `json:"gasRebateRate"`
		GasPrice      uint64 `json:"gasPrice"`
	} `json:"block"`
	Alloc map[ld.EthID]struct {
		Balance   *big.Int   `json:"balance"`
		Threshold uint8      `json:"threshold"`
		Keepers   []ld.EthID `json:"keepers"`
	} `json:"alloc"`
}

type ChainConfig struct {
	ChainID            uint64       `json:"chainID"`
	MaxTotalSupply     *big.Int     `json:"maxTotalSupply"`
	Message            string       `json:"message"`
	FeeConfigThreshold uint8        `json:"feeConfigThreshold"`
	FeeConfigKeepers   []ld.EthID   `json:"feeConfigKeepers"`
	FeeConfigID        ld.DataID    `json:"feeConfigID"`
	NameAppThreshold   uint8        `json:"nameAppThreshold"`
	NameAppKeepers     []ld.EthID   `json:"nameAppKeepers"`
	NameAppID          ld.ModelID   `json:"nameAppID"`
	ProfileAppID       ld.ModelID   `json:"profileAppID"`
	FeeConfig          *FeeConfig   `json:"feeConfig"`
	FeeConfigs         []*FeeConfig `json:"feeConfigs"`
}

func (c *ChainConfig) IsNameApp(id ids.ShortID) bool {
	return c.NameAppID == ld.ModelID(id)
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

func (c *ChainConfig) CheckChainID(chainID uint64) error {
	if chainID != c.ChainID {
		return fmt.Errorf("invalid ChainID %d, expected %d", chainID, c.ChainID)
	}
	return nil
}

type FeeConfig struct {
	StartHeight     uint64 `json:"startHeight"`
	ThresholdGas    uint64 `json:"thresholdGas"`
	MinGasPrice     uint64 `json:"minGasPrice"`
	MaxGasPrice     uint64 `json:"maxGasPrice"`
	MaxTxGas        uint64 `json:"maxTxGas"`
	MaxBlockTxsSize uint64 `json:"maxBlockTxsSize"`
	MaxBlockMiners  uint64 `json:"maxBlockMiners"`
	GasRebateRate   uint64 `json:"gasRebateRate"`
	MinMinerStake   uint64 `json:"minMinerStake"`
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
	txs := make([]*ld.Transaction, 0)
	// 道生一，一生二，二生三，三生万物
	// The first transaction is issued by the Blackhole account, to the Genesis account.
	// It has included ChainID, MaxTotalSupply and Genesis Message.
	tx := &ld.Transaction{
		Type:    ld.TypeMintFee,
		ChainID: g.Chain.ChainID,
		From:    constants.BlackholeAddr,
		To:      constants.GenesisAddr,
		Amount:  g.Chain.MaxTotalSupply,
		Data:    []byte(g.Chain.Message),
	}
	txs = append(txs, tx)

	// config data tx
	cfg, err := json.Marshal(g.Chain.FeeConfig)
	if err != nil {
		return nil, err
	}
	cfgData := &ld.DataMeta{
		ModelID:   ids.ShortID(constants.JsonModelID),
		Version:   1,
		Threshold: g.Chain.FeeConfigThreshold,
		Keepers:   ld.EthIDsToShort(g.Chain.FeeConfigKeepers...),
		Data:      cfg,
	}

	// Alloc Txs
	if len(g.Alloc) == 0 {
		return nil, fmt.Errorf("genesis allocation empty")
	}

	genesisNonce := uint64(0)
	for k, v := range g.Alloc {
		tx := &ld.Transaction{
			Type:    ld.TypeTransfer,
			ChainID: g.Chain.ChainID,
			Nonce:   genesisNonce,
			From:    constants.GenesisAddr,
			To:      ids.ShortID(k),
			Amount:  v.Balance,
		}
		genesisNonce++
		txs = append(txs, tx)

		if le := len(v.Keepers); le > 0 {
			update := &ld.TxUpdater{
				Threshold: v.Threshold,
				Keepers:   make([]ids.ShortID, len(v.Keepers)),
			}

			for i, id := range v.Keepers {
				update.Keepers[i] = ids.ShortID(id)
			}

			tx := &ld.Transaction{
				Type:    ld.TypeUpdateAccountKeepers,
				ChainID: g.Chain.ChainID,
				Nonce:   uint64(0),
				From:    ids.ShortID(k),
				Data:    update.Bytes(),
			}
			txs = append(txs, tx)
		}
	}

	// Config tx
	if err := cfgData.SyntacticVerify(); err != nil {
		return nil, err
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateData,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAddr,
		Data:    cfgData.Bytes(),
	}
	if err := tx.SyntacticVerify(); err != nil {
		return nil, err
	}

	g.Chain.FeeConfigID = ld.DataID(tx.ShortID())
	txs = append(txs, tx)

	// name app tx
	name, sch := app.NameSchema()
	nameModel := &ld.ModelMeta{
		Name:      name,
		Threshold: g.Chain.NameAppThreshold,
		Keepers:   ld.EthIDsToShort(g.Chain.NameAppKeepers...),
		Data:      sch,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAddr,
		Data:    nameModel.Bytes(),
	}
	if err := tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.NameAppID = ld.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// Profile app tx
	name, sch = app.ProfileSchema()
	profileModel := &ld.ModelMeta{
		Name: name,
		Data: sch,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAddr,
		Data:    profileModel.Bytes(),
	}
	if err := tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.ProfileAppID = ld.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// build genesis block
	blk := &ld.Block{
		Parent:        ids.Empty,
		Height:        0,
		Timestamp:     0,
		Gas:           0,
		GasPrice:      g.Block.GasPrice,
		GasRebateRate: g.Block.GasRebateRate,
		Txs:           txs,
	}
	if err := blk.SyntacticVerify(); err != nil {
		return nil, err
	}
	return blk, nil
}

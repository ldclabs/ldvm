// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
)

type Genesis struct {
	Chain ChainConfig                `json:"chain"`
	Alloc map[util.EthID]*Allocation `json:"alloc"`
}

type Allocation struct {
	Balance   *big.Int    `json:"balance"`
	Threshold uint8       `json:"threshold"`
	Keepers   util.EthIDs `json:"keepers"`
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

func (c *ChainConfig) IsNameService(id util.ModelID) bool {
	return c.NameServiceID == id
}

func (c *ChainConfig) Fee(height uint64) *FeeConfig {
	// the first one is the latest.
	for i, cfg := range c.FeeConfigs {
		if cfg.StartHeight <= height {
			return c.FeeConfigs[i]
		}
	}
	return c.FeeConfig
}

func (c *ChainConfig) AppendFeeConfig(data []byte) (*FeeConfig, error) {
	fee := new(FeeConfig)
	if err := json.Unmarshal(data, fee); err != nil {
		return nil, err
	}
	if err := fee.SyntacticVerify(); err != nil {
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

func (cfg *FeeConfig) SyntacticVerify() error {
	if cfg == nil {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: nil pointer")
	}
	if cfg.ThresholdGas <= 500 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid thresholdGas")
	}
	if cfg.MinGasPrice <= 500 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid minGasPrice")
	}
	if cfg.MaxGasPrice <= cfg.MinGasPrice {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid maxGasPrice")
	}
	if cfg.MaxTxGas <= 1000000 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid maxTxGas")
	}
	if cfg.MaxBlockTxsSize <= 1000000 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid maxBlockTxsSize")
	}
	if cfg.GasRebateRate > 1000 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid gasRebateRate")
	}
	if cfg.MinTokenPledge.Cmp(new(big.Int).SetUint64(constants.LDC)) < 0 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid minTokenPledge")
	}
	if cfg.MinStakePledge.Cmp(new(big.Int).SetUint64(constants.LDC)) < 0 {
		return fmt.Errorf("FeeConfig.SyntacticVerify failed: invalid minStakePledge")
	}
	return nil
}

func FromJSON(data []byte) (*Genesis, error) {
	g := new(Genesis)
	if err := json.Unmarshal(data, g); err != nil {
		return nil, err
	}
	if err := g.Chain.FeeConfig.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.FeeConfigs = []*FeeConfig{}
	return g, nil
}

func (g *Genesis) ToBlock() (*ld.Block, error) {
	genesisAccount, ok := g.Alloc[constants.GenesisAccount]
	if !ok {
		return nil, fmt.Errorf("genesis account not found")
	}

	txs := make([]*ld.Transaction, 0)
	// The first transaction is issued by the Genesis account, to create native token.
	// It has included ChainID, MaxTotalSupply and Genesis Message.
	minter := &ld.TxAccounter{
		Amount: g.Chain.MaxTotalSupply,
		Name:   "Linked Data Chain",
		Data:   []byte(strconv.Quote(g.Chain.Message)),
	}
	tx := &ld.Transaction{
		Type:    ld.TypeCreateToken,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		To:      &constants.LDCAccount,
		Amount:  g.Chain.MaxTotalSupply,
		Data:    ld.MustMarshal(minter),
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
		to := util.EthID(id)
		tx := &ld.Transaction{
			Type:    ld.TypeTransfer,
			ChainID: g.Chain.ChainID,
			Nonce:   nonce,
			From:    constants.LDCAccount,
			To:      &to,
			Amount:  v.Balance,
		}
		nonce++
		txs = append(txs, tx)

		if le := len(v.Keepers); le > 0 {
			update := &ld.TxAccounter{
				Threshold: v.Threshold,
				Keepers:   v.Keepers,
			}

			tx := &ld.Transaction{
				Type:    ld.TypeUpdateAccountKeepers,
				ChainID: g.Chain.ChainID,
				From:    util.EthID(id),
				Data:    ld.MustMarshal(update),
			}
			txs = append(txs, tx)
		}
	}

	// config data tx
	cfg, err := json.Marshal(g.Chain.FeeConfig)
	if err != nil {
		return nil, err
	}
	mid := util.ModelID(constants.JSONModelID)
	cfgData := &ld.TxUpdater{
		ModelID:   &mid,
		Version:   1,
		Threshold: genesisAccount.Threshold,
		Keepers:   genesisAccount.Keepers,
		Data:      cfg,
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateData,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(cfgData),
	}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.FeeConfigID = util.DataID(tx.ShortID())
	txs = append(txs, tx)

	// name app tx
	nm, err := service.NameModel()
	if err != nil {
		return nil, err
	}
	ns := &ld.ModelMeta{
		Name:      nm.Name(),
		Threshold: genesisAccount.Threshold,
		Keepers:   genesisAccount.Keepers,
		Data:      nm.Schema(),
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(ns),
	}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.NameServiceID = util.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// Profile app tx
	pm, err := service.ProfileModel()
	if err != nil {
		return nil, err
	}
	ps := &ld.ModelMeta{
		Name:      pm.Name(),
		Threshold: genesisAccount.Threshold,
		Keepers:   genesisAccount.Keepers,
		Data:      pm.Schema(),
	}
	tx = &ld.Transaction{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(ps),
	}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, err
	}
	g.Chain.ProfileServiceID = util.ModelID(tx.ShortID())
	txs = append(txs, tx)

	// build genesis block
	blk := &ld.Block{
		GasPrice:      g.Chain.FeeConfig.MinGasPrice,
		GasRebateRate: g.Chain.FeeConfig.GasRebateRate,
		Txs:           txs,
	}
	if err = blk.SyntacticVerify(); err != nil {
		return nil, err
	}
	return blk, nil
}

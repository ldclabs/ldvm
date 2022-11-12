// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/ld/service"
	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
)

type Genesis struct {
	Chain ChainConfig                  `json:"chain"`
	Alloc map[util.Address]*Allocation `json:"alloc"`
}

type Allocation struct {
	Balance   *big.Int    `json:"balance"`
	Threshold uint16      `json:"threshold"`
	Keepers   signer.Keys `json:"keepers"`
}

type ChainConfig struct {
	ChainID        uint64     `json:"chainID"`
	MaxTotalSupply *big.Int   `json:"maxTotalSupply"`
	Message        string     `json:"message"`
	FeeConfig      *FeeConfig `json:"feeConfig"`

	// external assignment fields
	FeeConfigs    []*FeeConfig `json:"feeConfigs"`
	FeeConfigID   util.DataID  `json:"feeConfigID"`
	NameServiceID util.ModelID `json:"nameServiceID"`
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
	errp := util.ErrPrefix("ChainConfig.AppendFeeConfig: ")

	if err := util.UnmarshalCBOR(data, fee); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := fee.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	c.FeeConfigs = append(c.FeeConfigs, fee)
	return fee, nil
}

type FeeConfig struct {
	StartHeight            uint64                        `json:"startHeight"`
	MinGasPrice            uint64                        `json:"minGasPrice"`
	MaxGasPrice            uint64                        `json:"maxGasPrice"`
	MaxTxGas               uint64                        `json:"maxTxGas"`
	GasRebateRate          uint64                        `json:"gasRebateRate"`
	MinTokenPledge         *big.Int                      `json:"minTokenPledge"`
	MinStakePledge         *big.Int                      `json:"minStakePledge"`
	NonTransferableBalance *big.Int                      `json:"nonTransferableBalance"`
	Builders               util.IDList[util.StakeSymbol] `json:"builders"`
}

func (cfg *FeeConfig) SyntacticVerify() error {
	errp := util.ErrPrefix("FeeConfig.SyntacticVerify: ")

	switch {
	case cfg == nil:
		return errp.Errorf("nil pointer")

	case cfg.MinGasPrice <= 500:
		return errp.Errorf("invalid minGasPrice")

	case cfg.MaxGasPrice <= cfg.MinGasPrice:
		return errp.Errorf("invalid maxGasPrice")

	case cfg.MaxTxGas <= 1000000:
		return errp.Errorf("invalid maxTxGas")

	case cfg.GasRebateRate > 1000:
		return errp.Errorf("invalid gasRebateRate")

	case cfg.MinTokenPledge == nil || cfg.MinTokenPledge.Cmp(new(big.Int).SetUint64(constants.LDC)) < 0:
		return errp.Errorf("invalid minTokenPledge")

	case cfg.MinStakePledge == nil || cfg.MinStakePledge.Cmp(new(big.Int).SetUint64(constants.LDC)) < 0:
		return errp.Errorf("invalid minStakePledge")

	case cfg.NonTransferableBalance == nil || cfg.NonTransferableBalance.Cmp(new(big.Int).SetUint64(0)) < 0:
		return errp.Errorf("invalid nonTransferableBalance")

	case cfg.Builders == nil:
		return errp.Errorf("nil builders")
	}

	if err := cfg.Builders.Valid(); err != nil {
		return errp.ErrorIf(err)
	}

	return nil
}

func (cfg *FeeConfig) ValidBuilder(builder util.StakeSymbol) error {
	errp := util.ErrPrefix("FeeConfig.ValidBuilder: ")

	if len(cfg.Builders) > 0 && !cfg.Builders.Has(builder) {
		return errp.Errorf("%s is not in the builder list", builder)
	}

	return nil
}

func FromJSON(data []byte) (*Genesis, error) {
	g := new(Genesis)
	errp := util.ErrPrefix("FromJSON: ")

	if err := json.Unmarshal(data, g); err != nil {
		return nil, errp.ErrorIf(err)
	}
	if err := g.Chain.FeeConfig.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}

	g.Chain.FeeConfigs = []*FeeConfig{}
	return g, nil
}

func (g *Genesis) ToTxs() (ld.Txs, error) {
	errp := util.ErrPrefix("Genesis.ToTxs: ")

	genesisAccount, ok := g.Alloc[constants.GenesisAccount]
	if !ok {
		return nil, errp.Errorf("genesis account not found")
	}

	var err error
	genesisNonce := uint64(0)
	txs := make([]*ld.Transaction, 0)
	// The first transaction is issued by the Genesis account, to create native token.
	// It has included ChainID, MaxTotalSupply and Genesis Message.
	token := &ld.TxAccounter{
		Amount: g.Chain.MaxTotalSupply,
		Name:   "Linked Data Chain",
		Data:   []byte(strconv.Quote(g.Chain.Message)),
	}
	tx := &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateToken,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAccount,
		To:      &constants.LDCAccount,
		Data:    ld.MustMarshal(token),
	}}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	genesisNonce++
	txs = append(txs, tx)

	// Alloc Txs
	ldcNonce := uint64(0)
	list := make([]ids.ShortID, 0, len(g.Alloc))
	for id := range g.Alloc {
		list = append(list, ids.ShortID(id))
	}
	ids.SortShortIDs(list)
	for _, id := range list {
		v := g.Alloc[util.Address(id)]
		to := util.Address(id)
		tx := &ld.Transaction{Tx: ld.TxData{
			Type:    ld.TypeTransfer,
			ChainID: g.Chain.ChainID,
			Nonce:   ldcNonce,
			From:    constants.LDCAccount,
			To:      &to,
			Amount:  v.Balance,
		}}
		if err = tx.SyntacticVerify(); err != nil {
			return nil, errp.ErrorIf(err)
		}
		ldcNonce++
		txs = append(txs, tx)

		if le := len(v.Keepers); le > 0 {
			update := &ld.TxAccounter{
				Threshold: &v.Threshold,
				Keepers:   &v.Keepers,
			}

			nonce := uint64(0)
			tx := &ld.Transaction{Tx: ld.TxData{
				Type:    ld.TypeUpdateAccountInfo,
				ChainID: g.Chain.ChainID,
				Nonce:   nonce,
				From:    util.Address(id),
				Data:    ld.MustMarshal(update),
			}}

			if tx.Tx.From == constants.GenesisAccount {
				tx.Tx.Nonce = genesisNonce
				genesisNonce++
			}
			if err = tx.SyntacticVerify(); err != nil {
				return nil, errp.ErrorIf(err)
			}
			txs = append(txs, tx)
		}
	}

	// config data tx
	cfg, err := util.MarshalCBOR(g.Chain.FeeConfig)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	cfgData := &ld.TxUpdater{
		ModelID:   &ld.CBORModelID,
		Version:   1,
		Threshold: &genesisAccount.Threshold,
		Keepers:   &genesisAccount.Keepers,
		Data:      cfg,
	}
	if err = cfgData.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}

	tx = &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateData,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(cfgData),
	}}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	genesisNonce++
	g.Chain.FeeConfigID = util.DataID(tx.ID)
	txs = append(txs, tx)

	// name service tx
	nm, err := service.NameModel()
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	ns := &ld.ModelInfo{
		Name:      nm.Name(),
		Threshold: genesisAccount.Threshold,
		Keepers:   genesisAccount.Keepers,
		Schema:    nm.Schema(),
	}
	if err = ns.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}

	tx = &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(ns),
	}}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	genesisNonce++
	g.Chain.NameServiceID = util.ModelIDFromHash(tx.ID)
	txs = append(txs, tx)

	// Profile service tx
	pm, err := service.ProfileModel()
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	ps := &ld.ModelInfo{
		Name:      pm.Name(),
		Threshold: genesisAccount.Threshold,
		Keepers:   genesisAccount.Keepers,
		Schema:    pm.Schema(),
	}
	if err = ps.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}

	tx = &ld.Transaction{Tx: ld.TxData{
		Type:    ld.TypeCreateModel,
		ChainID: g.Chain.ChainID,
		Nonce:   genesisNonce,
		From:    constants.GenesisAccount,
		Data:    ld.MustMarshal(ps),
	}}
	if err = tx.SyntacticVerify(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	genesisNonce++
	txs = append(txs, tx)
	return txs, nil
}

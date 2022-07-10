// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package transaction

import (
	"encoding/json"
	"math/big"

	"github.com/ldclabs/ldvm/constants"
	"github.com/ldclabs/ldvm/ld"
	"github.com/ldclabs/ldvm/util"
)

type TxCreateToken struct {
	TxBase
	input *ld.TxAccounter
}

func (tx *TxCreateToken) MarshalJSON() ([]byte, error) {
	if tx == nil || tx.ld == nil {
		return []byte("null"), nil
	}

	v := tx.ld.Copy()
	errp := util.ErrPrefix("TxCreateToken.MarshalJSON error: ")
	if tx.input == nil {
		return nil, errp.Errorf("nil tx.input")
	}
	d, err := json.Marshal(tx.input)
	if err != nil {
		return nil, errp.ErrorIf(err)
	}
	v.Data = d
	return errp.ErrorMap(json.Marshal(v))
}

func (tx *TxCreateToken) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("TxCreateToken.SyntacticVerify error: ")

	if err = tx.TxBase.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.ld.To == nil:
		return errp.Errorf("nil to as token account")

	case *tx.ld.To == util.EthIDEmpty:
		return errp.Errorf("invalid to as token account, expected not %s", tx.ld.To)

	case tx.ld.Amount == nil:
		return errp.Errorf("nil amount")

	case tx.ld.Token != nil:
		return errp.Errorf("invalid token, should be nil")

	case len(tx.ld.Data) == 0:
		return errp.Errorf("invalid data")
	}

	if token := util.TokenSymbol(*tx.ld.To); !token.Valid() {
		return errp.Errorf("invalid token %s", token.GoString())
	}

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	switch {
	case tx.input.Threshold == nil || *tx.input.Threshold == 0:
		return errp.Errorf("invalid threshold, expected >= 1")

	case tx.input.Amount == nil || tx.input.Amount.Sign() <= 0:
		return errp.Errorf("invalid amount, expected >= 1")

	case tx.input.Approver != nil && *tx.input.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver, expected not %s", tx.input.Approver)

	case len(tx.input.Name) < 3:
		return errp.Errorf("invalid name %q, expected length >= 3", tx.input.Name)
	}
	return nil
}

// ApplyGenesis skipping signature verification
func (tx *TxCreateToken) ApplyGenesis(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateToken.ApplyGenesis error: ")

	tx.input = &ld.TxAccounter{}
	if err = tx.input.Unmarshal(tx.ld.Data); err != nil {
		return errp.ErrorIf(err)
	}
	if err = tx.input.SyntacticVerify(); err != nil {
		return errp.ErrorIf(err)
	}

	tx.amount = new(big.Int)
	tx.tip = new(big.Int)
	tx.fee = new(big.Int)
	tx.cost = new(big.Int)

	tx.from, err = bs.LoadAccount(tx.ld.From)
	if err != nil {
		return errp.ErrorIf(err)
	}
	if tx.ldc, err = bs.LoadAccount(constants.LDCAccount); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.miner, err = bs.LoadMiner(bctx.Miner()); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to, err = bs.LoadAccount(*tx.ld.To); err != nil {
		return errp.ErrorIf(err)
	}

	if err = tx.to.CreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to.id != constants.LDCAccount {
		pledge := new(big.Int).Set(bctx.FeeConfig().MinTokenPledge)
		tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}

func (tx *TxCreateToken) Apply(bctx BlockContext, bs BlockState) error {
	var err error
	errp := util.ErrPrefix("TxCreateToken.Apply error: ")

	if err = tx.TxBase.verify(bctx, bs); err != nil {
		return errp.ErrorIf(err)
	}

	feeCfg := bctx.FeeConfig()
	if tx.ld.Amount.Cmp(feeCfg.MinTokenPledge) < 0 {
		return errp.Errorf("invalid amount, expected >= %v, got %v",
			feeCfg.MinTokenPledge, tx.ld.Amount)
	}

	if err = tx.to.CreateToken(tx.input); err != nil {
		return errp.ErrorIf(err)
	}
	if tx.to.id != constants.LDCAccount {
		pledge := new(big.Int).Set(bctx.FeeConfig().MinTokenPledge)
		tx.to.Init(pledge, bs.Height(), bs.Timestamp())
	}
	return errp.ErrorIf(tx.TxBase.accept(bctx, bs))
}
